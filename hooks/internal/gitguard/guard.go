package gitguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sumsolution/folkways/hooks/internal/output"
	"github.com/sumsolution/folkways/hooks/internal/payload"
	"github.com/sumsolution/folkways/hooks/internal/team"
	"mvdan.cc/sh/v3/syntax"
)

var (
	header    = regexp.MustCompile(`^(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(([a-z0-9._-]+)\))?!?: \S`)
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
// A line that does not parse is allowed: git-guard catches mistakes, it is not a shell.
func checkCommand(cmd, payloadCwd string, cfg team.Config) string {
	file, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(cmd), "")
	if err != nil {
		return ""
	}
	cur := payloadCwd
	if cur == "" {
		cur, _ = os.Getwd()
	}
	home, err := os.UserHomeDir()
	w := &walker{cfg: cfg, res: resolver{home: home, homeOK: err == nil && home != ""}}
	w.list(file.Stmts, dirset{dirs: []string{cur}})
	return w.reason
}

func resolve(base, dir string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Join(base, dir)
}

// messageArg returns the value of the commit message word: its literal text, or the body of
// `$(cat <<'EOF' ... EOF)`. ok is false for any other expansion, which cannot be evaluated here.
func (r resolver) messageArg(w *syntax.Word) (string, bool) {
	if s, ok := r.word(w); ok {
		return s, true
	}
	if len(w.Parts) != 1 {
		return "", false
	}
	dq, ok := w.Parts[0].(*syntax.DblQuoted)
	if !ok || len(dq.Parts) != 1 {
		return "", false
	}
	cs, ok := dq.Parts[0].(*syntax.CmdSubst)
	if !ok || len(cs.Stmts) != 1 {
		return "", false
	}
	st := cs.Stmts[0]
	call, ok := st.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 || len(st.Redirs) != 1 {
		return "", false
	}
	if name, ok := r.word(call.Args[0]); !ok || name != "cat" {
		return "", false
	}
	if h := st.Redirs[0]; (h.Op == syntax.Hdoc || h.Op == syntax.DashHdoc) && h.Hdoc != nil {
		return r.heredocBody(h.Hdoc)
	}
	return "", false
}

// commitMessage returns the first message given by -m, -am, -m<msg>, --message <msg> or --message=<msg>.
// ok is false when there is none or it is not statically known.
func commitMessage(r resolver, args []arg) (string, bool) {
	for i, a := range args {
		if !a.ok {
			continue
		}
		v := a.val
		var attached string
		var hasAttached, isMsg bool
		switch {
		case v == "--message":
			isMsg = true
		case strings.HasPrefix(v, "--message="):
			isMsg, attached, hasAttached = true, strings.TrimPrefix(v, "--message="), true
		case strings.HasPrefix(v, "-") && !strings.HasPrefix(v, "--"):
			// a cluster of short flags: everything after the first m is the message
			if k := strings.IndexByte(v, 'm'); k > 0 && letters(v[1:k]) {
				isMsg, attached, hasAttached = true, v[k+1:], k+1 < len(v)
			}
		}
		if !isMsg {
			continue
		}
		if hasAttached {
			return attached, true
		}
		if i+1 >= len(args) {
			return "", false
		}
		return r.messageArg(args[i+1].w)
	}
	return "", false
}

func letters(s string) bool {
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}

// subjectOf returns the first non-empty line of a message.
func subjectOf(msg string) string {
	for _, line := range strings.Split(msg, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

func checkCommit(r resolver, args []arg, cfg team.Config) string {
	msg, ok := commitMessage(r, args)
	if !ok {
		return ""
	}
	subject := subjectOf(msg)
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
