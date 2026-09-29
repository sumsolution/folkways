package gitguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/sumsolution/folkways/hooks/internal/output"
	"github.com/sumsolution/folkways/hooks/internal/payload"
	"github.com/sumsolution/folkways/hooks/internal/team"
)

var (
	header    = regexp.MustCompile(`^(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(([a-z0-9._-]+)\))?!?: \S`)
	shortMsg  = regexp.MustCompile(`^-[a-zA-Z]*m$`) // -m, -am
	envAssign = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
	mainGuard = map[string]bool{"checkout": true, "switch": true, "pull": true, "reset": true, "rebase": true}
)

// Run checks the Bash command of a PreToolUse payload and prints a deny decision when a rule is broken.
func Run(p payload.Payload, cfg team.Config) {
	cmd := p.BashCommand()
	if cmd == "" {
		return
	}
	if reason := checkCommand(cmd, p.Cwd, cfg); reason != "" {
		output.Deny(reason)
	}
}

// checkCommand checks every git command of a command line and returns the first deny reason ("" to allow).
// Directories are tracked conservatively: a git command is denied when it is denied in any directory it
// could run in (see dirsFor).
func checkCommand(cmd, payloadCwd string, cfg team.Config) string {
	cur := payloadCwd
	if cur == "" {
		cur, _ = os.Getwd()
	}
	return checkScript(cmd, []string{cur}, cfg)
}

// checkScript checks a shell script (a command line or a substitution body) that starts in one of dirs.
// Top-level heredoc bodies are data and skipped.
func checkScript(script string, dirs []string, cfg team.Config) string {
	elems, _ := parseSeq(stripHeredocs(script), 0, false)
	return checkSeq(elems, dirs, cfg)
}

// checkSeq walks a sequence of commands. It tracks two sets of possible directories:
//   - ok: where the shell is if every command so far succeeded (a `cd` moved),
//   - failed: where it is if some command failed (a failed `cd` did not move).
//
// `a && b` runs b in ok; `a || b` runs b in failed; after `;`, a newline, `|` or `&` a command may run in either.
// A `cd` in a pipeline or in the background, and in a ( ) subshell, does not affect what follows.
func checkSeq(elems []elem, start []string, cfg team.Config) string {
	var ok, failed []string
	for k, e := range elems {
		var dirs []string
		prev := ""
		if k > 0 {
			prev = elems[k-1].op
		}
		switch {
		case k == 0:
			dirs = start
		case prev == "&&":
			dirs = ok
		case prev == "||" && len(failed) > 0:
			dirs = failed
		default:
			dirs = union(ok, failed)
		}
		var moved []string // where a successful cd leaves us; nil when the command is not a persistent cd
		if e.sub != nil {
			if reason := checkSeq(e.sub, dirs, cfg); reason != "" {
				return reason
			}
		} else {
			if reason := checkSegment(e.text, dirs, cfg); reason != "" {
				return reason
			}
			if target, isCD := cdTarget(e.text); isCD && prev != "|" && e.op != "|" && e.op != "&" {
				for _, d := range dirs {
					moved = append(moved, resolve(d, target))
				}
			}
		}
		succ, fail := dirs, dirs
		if moved != nil {
			succ = moved
		}
		switch {
		case k == 0:
			ok, failed = succ, fail
		case prev == "&&":
			ok, failed = succ, union(failed, fail)
		case prev == "||":
			ok, failed = union(ok, succ), fail
		default:
			ok, failed = succ, fail
		}
	}
	return ""
}

func union(a, b []string) []string {
	out := append([]string(nil), a...)
	for _, d := range b {
		if !slices.Contains(out, d) {
			out = append(out, d)
		}
	}
	return out
}

// reserved words that may precede a command: `if git push; then`, `! git push`, `{ git push; }`.
var reserved = map[string]bool{"if": true, "then": true, "else": true, "elif": true, "do": true, "while": true, "until": true, "!": true, "time": true, "{": true, "}": true}

// commandWords returns the words of one command without leading assignments and reserved words.
func commandWords(text string) []string {
	w := words(text)
	for len(w) > 0 && (envAssign.MatchString(w[0]) || reserved[w[0]]) {
		w = w[1:]
	}
	return w
}

// cdTarget returns the directory of a `cd <dir>` whose target is known.
func cdTarget(text string) (string, bool) {
	w := commandWords(text)
	if len(w) < 2 || w[0] != "cd" {
		return "", false
	}
	w = w[1:]
	for len(w) > 0 && (w[0] == "--" || w[0] == "-L" || w[0] == "-P") {
		w = w[1:]
	}
	if len(w) == 0 || w[0] == "-" || strings.ContainsAny(w[0], "$`~") {
		return "", false
	}
	return w[0], true
}

// checkSegment checks one simple command in each of dirs: the commands substituted into it, then the command.
func checkSegment(text string, dirs []string, cfg team.Config) string {
	for _, body := range substitutions(text) {
		if reason := checkScript(body, dirs, cfg); reason != "" {
			return reason
		}
	}
	w := commandWords(text)
	if len(w) >= 2 && w[0] == "git" {
		for _, d := range dirs {
			if reason := checkGit(w[1:], d, cfg); reason != "" {
				return reason
			}
		}
	}
	return ""
}

func resolve(base, dir string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Join(base, dir)
}

// Git global options that take an operand: `--opt value` or `--opt=value`, and `-C dir`, `-c k=v`.
var gitOptWithArg = map[string]bool{"-C": true, "-c": true, "--git-dir": true, "--work-tree": true, "--namespace": true, "--config-env": true}

// Git global options without an operand (`--exec-path` takes one only as `--exec-path=<path>`).
var gitFlag = map[string]bool{"-p": true, "-P": true, "--paginate": true, "--no-pager": true, "--no-replace-objects": true,
	"--no-optional-locks": true, "--bare": true, "--literal-pathspecs": true, "--glob-pathspecs": true,
	"--noglob-pathspecs": true, "--icase-pathspecs": true, "--no-lazy-fetch": true, "--exec-path": true}

// checkGit checks one `git [global options] <sub> args...` (args excludes the leading "git").
func checkGit(args []string, dir string, cfg team.Config) string {
options:
	for len(args) > 0 {
		name, _, hasEq := strings.Cut(args[0], "=")
		switch {
		case args[0] == "-C" && len(args) >= 2:
			dir = resolve(dir, args[1])
			args = args[2:]
		case gitOptWithArg[name] && hasEq, gitFlag[name]:
			args = args[1:]
		case gitOptWithArg[args[0]] && len(args) >= 2:
			args = args[2:]
		default:
			break options
		}
	}
	if len(args) == 0 {
		return ""
	}
	sub, rest := args[0], args[1:]
	switch {
	case sub == "commit":
		return checkCommit(rest, cfg)
	case sub == "push":
		return checkPush(dir, cfg)
	case mainGuard[sub]:
		return checkMainClone(sub, dir, cfg)
	}
	return ""
}

// commitMessage returns the first -m/-am/--message value; ok is false when there is none.
func commitMessage(args []string) (string, bool) {
	for i, a := range args {
		switch {
		case shortMsg.MatchString(a) || a == "--message":
			if i+1 < len(args) {
				return args[i+1], true
			}
		case strings.HasPrefix(a, "--message="):
			return strings.TrimPrefix(a, "--message="), true
		}
	}
	return "", false
}

// subjectOf returns the first non-empty line of a message, reading through `$(cat <<'EOF' ... EOF)`.
// ok is false for other command substitutions, which cannot be evaluated here.
func subjectOf(msg string) (string, bool) {
	if strings.HasPrefix(msg, "$(") {
		at := strings.Index(msg, "<<")
		if at < 0 {
			return "", false
		}
		word, j := heredocStart(msg, at)
		nl := strings.IndexByte(msg[j:], '\n')
		if word == "" || nl < 0 {
			return "", false
		}
		msg = msg[j+nl+1:]
	}
	for _, line := range strings.Split(msg, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line, true
		}
	}
	return "", true
}

func checkCommit(args []string, cfg team.Config) string {
	msg, ok := commitMessage(args)
	if !ok {
		return ""
	}
	subject, ok := subjectOf(msg)
	if !ok {
		return ""
	}
	m := header.FindStringSubmatch(subject)
	if m == nil {
		return fmt.Sprintf("Commit message is not Angular \"type(scope): subject\": \"%s\".", subject)
	}
	scopes := cfg.Team.Git.Scopes
	if len(scopes) == 0 {
		return ""
	}
	if m[1] == "" {
		return fmt.Sprintf("Commit message needs a scope: \"type(scope): subject\", one of team.yaml git.scopes (%s).", strings.Join(scopes, ", "))
	}
	for _, s := range scopes {
		if s == m[1] {
			return ""
		}
	}
	return fmt.Sprintf("Scope \"%s\" is not in team.yaml git.scopes (%s).", m[1], strings.Join(scopes, ", "))
}

func gitOut(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

func checkPush(dir string, cfg team.Config) string {
	if cfg.ReviewTrigger() != "pre-push" {
		return ""
	}
	head, err := gitOut(dir, "rev-parse", "HEAD")
	if err != nil || head == "" {
		return "" // not a repo we can read: fail open
	}
	if hasReview(cfg.Dir, cfg.SafeHandle(), head) {
		return ""
	}
	return fmt.Sprintf("Push blocked: review.trigger is pre-push and no review of HEAD %s exists. Run the review first (a file under efforts/ or people/<handle>/work/ with frontmatter sha: %s, or people/<handle>/inbox/<date>-review-%s.md).", head[:min(7, len(head))], head, head[:min(7, len(head))])
}

// checkMainClone denies checkout/switch/pull/reset/rebase in a main clone (`.git` is a directory)
// under projects_dir. Worktrees have a `.git` file and pass.
func checkMainClone(sub, dir string, cfg team.Config) string {
	if cfg.Me.ProjectsDir == "" {
		return ""
	}
	projects, err1 := filepath.EvalSymlinks(cfg.Me.ProjectsDir)
	target, err2 := filepath.EvalSymlinks(dir)
	if err1 != nil || err2 != nil || !within(projects, target) {
		return ""
	}
	top, err := gitOut(target, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	if info, err := os.Stat(filepath.Join(top, ".git")); err != nil || !info.IsDir() {
		return ""
	}
	return fmt.Sprintf("git %s in the main clone %s is not allowed. Create a worktree instead (git worktree add) and work there.", sub, top)
}

func within(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
