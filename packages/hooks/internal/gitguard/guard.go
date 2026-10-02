package gitguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sumsolution/folkways/packages/hooks/internal/output"
	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
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
	dir := payloadCwd
	if dir == "" {
		dir, _ = os.Getwd()
	}
	home, _ := os.UserHomeDir()
	w := &walker{cfg: cfg, res: newResolver(home)}
	w.list(file.Stmts, cwd{dir, dir != ""}, true)
	return w.reason
}

// Commit options that take an operand. A short one in a cluster (`-Skey`, `-Fmsg.txt`) takes the rest of the
// cluster, or the next word when nothing is attached. -S and -u take only an attached operand.
const (
	shortWithOperand   = "FcCt"
	shortOptionalAfter = "Su"
)

// commitMessage returns the first message given by -m, -am, -m<msg>, --message <msg> or --message=<msg>.
// A short-option cluster is read left to right: an m ends it with the message, an option with an operand
// ends it without one. ok is false when there is none or it is not statically known.
func commitMessage(args []arg) (string, bool) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !a.ok {
			continue
		}
		v := a.val
		switch {
		case v == "--":
			return "", false
		case v == "--message":
			if i+1 >= len(args) {
				return "", false
			}
			return args[i+1].val, args[i+1].ok
		case strings.HasPrefix(v, "--message="):
			return strings.TrimPrefix(v, "--message="), true
		case strings.HasPrefix(v, "-") && !strings.HasPrefix(v, "--"):
		cluster:
			for k := 1; k < len(v); k++ {
				c := v[k]
				switch {
				case c == 'm':
					if k+1 < len(v) {
						return v[k+1:], true
					}
					if i+1 >= len(args) {
						return "", false
					}
					return args[i+1].val, args[i+1].ok
				case strings.IndexByte(shortWithOperand, c) >= 0:
					if k+1 == len(v) {
						i++ // the operand is the next word
					}
					break cluster
				case strings.IndexByte(shortOptionalAfter, c) >= 0:
					break cluster
				}
			}
		}
	}
	return "", false
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

// personalHint names the personal root in a deny message.
func personalHint(cfg team.Config) string {
	if cfg.Solo() {
		return "the repo root"
	}
	return "people/<handle>"
}

func checkCommit(args []arg, cfg team.Config) string {
	msg, ok := commitMessage(args)
	if !ok {
		return ""
	}
	subject := subjectOf(msg)
	switch cfg.Team.Git.CommitConvention {
	case "none":
		return ""
	case "pattern":
		return checkPattern(subject, cfg.Team.Git.CommitPattern)
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
		return fmt.Sprintf("Commit message needs a scope: \"type(scope): subject\", one of folkways.yaml git.scopes (%s).", strings.Join(scopes, ", "))
	}
	for _, s := range scopes {
		if s == m[1] {
			return ""
		}
	}
	return fmt.Sprintf("Scope \"%s\" is not in folkways.yaml git.scopes (%s).", m[1], strings.Join(scopes, ", "))
}

// checkPattern checks the subject against git.commit_pattern. An empty or invalid pattern checks nothing.
func checkPattern(subject, pattern string) string {
	re, err := regexp.Compile(pattern)
	if pattern == "" || err != nil || re.MatchString(subject) {
		return ""
	}
	return fmt.Sprintf("Commit subject doesn't match folkways.yaml git.commit_pattern (%s): \"%s\".", pattern, subject)
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
	if hasReview(cfg.Dir, cfg.PersonalRoot(), head) {
		return ""
	}
	if cfg.Solo() && isFolkwaysRepo(dir, cfg) {
		for _, sha := range walkBack(dir, head) {
			if hasReview(cfg.Dir, cfg.PersonalRoot(), sha) {
				return ""
			}
		}
	}
	return fmt.Sprintf("Push blocked: review.trigger is pre-push and no review of HEAD %s exists. Run the review first (a file under efforts/ or %s/work/ with frontmatter sha: %s, or %s/inbox/<date>-review-%s.md).", head[:min(7, len(head))], personalHint(cfg), head, personalHint(cfg), head[:min(7, len(head))])
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
	if sameDir(top, cfg.Dir) {
		return "" // the folkways repo itself is never guarded
	}
	return fmt.Sprintf("git %s in the main clone %s is not allowed. Create a worktree instead (git worktree add) and work there.", sub, top)
}

func within(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// sameDir reports whether two paths name the same directory once symlinks are resolved.
func sameDir(a, b string) bool {
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	return err1 == nil && err2 == nil && ra == rb
}

// isFolkwaysRepo reports whether dir is inside the repo that holds the folkways config (cfg.Dir is its top level).
func isFolkwaysRepo(dir string, cfg team.Config) bool {
	top, err := gitOut(dir, "rev-parse", "--show-toplevel")
	return err == nil && sameDir(top, cfg.Dir)
}

// maxWalkBack is how many commits that change only work/ and inbox/ walkBack skips.
const maxWalkBack = 20

// walkBack returns the commits whose review counts for a push of head in a solo repo (spec 7.3): head, then,
// while a commit changes only work/ and inbox/ files, its first parent, up to maxWalkBack such commits; the
// walk ends with (and includes) the first commit that changes anything else, a merge or the root commit.
func walkBack(dir, head string) []string {
	var shas []string
	cur := head
	for steps := 0; ; steps++ {
		shas = append(shas, cur)
		out, err := gitOut(dir, "rev-list", "--parents", "-n", "1", cur)
		parents := strings.Fields(out)
		if err != nil || len(parents) != 2 || steps >= maxWalkBack {
			return shas // root, merge, error or limit
		}
		names, err := exec.Command("git", "-C", dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "-z", cur).Output()
		if err != nil {
			return shas
		}
		for _, n := range strings.Split(strings.TrimRight(string(names), "\x00"), "\x00") {
			if n != "" && !strings.HasPrefix(n, "work/") && !strings.HasPrefix(n, "inbox/") {
				return shas
			}
		}
		cur = parents[1]
	}
}
