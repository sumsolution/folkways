package gitguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

func scoped(scopes ...string) team.Config {
	var c team.Config
	c.Team.Git.Scopes = scopes
	return c
}

func TestCheckCommit(t *testing.T) {
	cases := []struct {
		cmd    string
		cfg    team.Config
		denied string // substring of the reason; "" = pass
	}{
		{`git commit -m "feat: add x"`, team.Config{}, ""},
		{`git commit -m "add x"`, team.Config{}, "not Angular"},
		{`git commit -am 'fix(api)!: y'`, team.Config{}, ""},
		{`git commit --message="oops"`, team.Config{}, "not Angular"},
		{`git commit --no-edit`, team.Config{}, ""},
		{`git commit -F msg.txt`, team.Config{}, ""},
		{`git commit -m "feat: x"`, scoped("web-app"), "needs a scope"},
		{`git commit -m "feat(auth): x"`, scoped("web-app", "ci"), "not in folkways.yaml"},
		{`git commit -m "feat(ci): x"`, scoped("web-app", "ci"), ""},
		{"git commit -m \"$(cat <<'EOF'\nfeat(ci): x\n\nbody\nEOF\n)\"", scoped("ci"), ""},
		{"git commit -m \"$(cat <<'EOF'\nnope\nEOF\n)\"", scoped("ci"), "not Angular"},
		{`git commit -m "$(date)"`, scoped("ci"), ""},
		{`cd x && git status && git commit -m "bad"`, team.Config{}, "not Angular"},
		{`FOO=1 git commit -m "bad"`, team.Config{}, "not Angular"},
	}
	for _, c := range cases {
		got := checkCommand(c.cmd, "/nonexistent", c.cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s: reason %q, want substring %q", c.cmd, got, c.denied)
		}
	}
}

// repo makes a git repo with one commit in dir and returns dir.
func repo(t *testing.T, dir string) string {
	t.Helper()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "init"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Skipf("git unavailable: %v %s", err, out)
		}
	}
	return dir
}

// env is a team repo with review.trigger pre-push, a main clone under projects_dir and a worktree of it.
type env struct {
	cfg            team.Config
	main, worktree string
}

func newEnv(t *testing.T) env {
	t.Helper()
	root, _ := filepath.EvalSymlinks(t.TempDir())
	e := env{main: repo(t, mkdir(t, root, "projects", "app"))}
	e.worktree = filepath.Join(root, "projects", "app-wt")
	if out, err := exec.Command("git", "-C", e.main, "worktree", "add", "-q", "-b", "feat/x", e.worktree).CombinedOutput(); err != nil {
		t.Skipf("git worktree: %v %s", err, out)
	}
	e.cfg.Dir = mkdir(t, root, "team")
	e.cfg.Team.Review.Trigger = "pre-push"
	e.cfg.Me.Handle = "alex"
	e.cfg.Me.ProjectsDir = filepath.Join(root, "projects")
	return e
}

func mkdir(t *testing.T, parts ...string) string {
	t.Helper()
	p := filepath.Join(parts...)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// checkAll runs each case in e and reports the ones whose reason does not match.
func checkAll(t *testing.T, e env, cases []dirCase) {
	t.Helper()
	for _, c := range cases {
		got := checkCommand(c.cmd, c.cwd, e.cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s: %q: reason %q, want substring %q", c.name, c.cmd, got, c.denied)
		}
	}
}

type dirCase struct {
	name   string
	cmd    string
	cwd    string
	denied string // substring of the reason; "" = pass
}

// The directory model: one current directory, changed only by a top-level cd in an && or ; chain.
// Anything else that could move it makes it unknown, and an unknown directory skips the push and
// main-clone checks.
func TestCheckCommandDirectories(t *testing.T) {
	e := newEnv(t)
	t.Setenv("HOME", filepath.Dir(filepath.Dir(e.main)))
	wt, main := e.worktree, e.main
	checkAll(t, e, []dirCase{
		{"payload cwd", "git push", wt, "Push blocked"},
		{"cd and", "cd " + wt + " && git push", "/", "Push blocked"},
		{"cd and: only the new dir", "cd " + wt + " && git checkout x", main, ""},
		{"cd and: main clone", "cd " + main + " && git checkout x", wt, "main clone"},
		{"cd semicolon", "cd " + main + "; git checkout x", wt, "main clone"},
		{"cd semicolon: only the new dir", "cd " + wt + "; git checkout x", main, ""},
		{"relative cd", "cd " + filepath.Dir(wt) + " && cd app-wt && git push", "/", "Push blocked"},
		{"tilde", "cd ~/projects/app-wt && git push", "/", "Push blocked"},
		{"$HOME", `cd "$HOME/projects/app-wt" && git push`, "/", "Push blocked"},
		{"${HOME}", "cd ${HOME}/projects/app && git reset --hard", wt, "main clone"},
		{"quoted tilde is literal", `cd "~/projects/app-wt" && git push`, "/", ""},
		{"plain cd goes home", "cd && git checkout x", main, ""},
		{"-C", "git -C " + wt + " push", "/", "Push blocked"},
		{"attached -C", "git -C" + wt + " push", "/", "Push blocked"},
		{"-C relative to cd", "cd " + filepath.Dir(wt) + " && git -C app-wt push", "/", "Push blocked"},
		{"-c and flags", "cd " + main + " && git -c a=b -ccore.x=y --no-pager -P --no-optional-locks reset --hard", wt, "main clone"},
		{"subshell", "(cd " + wt + " && git push)", "/", "Push blocked"},
		{"subshell cd does not leak", "(cd " + main + ") && git checkout x", wt, ""},
		{"brace group", "{ cd " + wt + "; git push; }", "/", "Push blocked"},
		{"substitution inherits cd", "cd " + wt + " && echo $(git push)", "/", "Push blocked"},
		{"substitution cd does not leak", "echo $(cd " + main + ") && git checkout x", wt, ""},

		// Unknown: fail open.
		{"cd variable", "cd $ELSEWHERE && git push", wt, ""},
		{"cd dash", "cd - && git push", wt, ""},
		{"relative cd after unknown", "cd $ELSEWHERE && cd sub && git push", wt, ""},
		{"absolute cd after unknown is known", "cd \"$X\" && cd " + wt + " && git push", "/", "Push blocked"},
		{"absolute -C after unknown is known", "cd $X && git -C " + wt + " push", "/", "Push blocked"},
		{"relative -C after unknown", "cd $X && git -C sub push", wt, ""},
		{"-C variable", "git -C \"$X\" push", wt, ""},
		{"cd in ||", "cd /nonexistent-dir || git push", wt, ""},
		{"cd in if", "if true; then cd " + main + "; fi; git push", wt, ""},
		{"cd in pipeline", "cd " + main + " | git checkout x", wt, ""},
		{"cd in background", "cd " + main + " & git checkout x", wt, ""},
		{"pushd", "pushd " + wt + " && git push", wt, ""},
		{"popd", "popd; git push", wt, ""},
		{"--git-dir", "git --git-dir=" + main + "/.git reset --hard", wt, ""},
		{"--git-dir operand", "git --git-dir " + main + "/.git reset --hard", wt, ""},
		{"--work-tree", "cd " + main + " && git --work-tree=/tmp reset --hard", wt, ""},
		{"GIT_DIR", "GIT_DIR=" + main + "/.git git reset --hard", wt, ""},
		{"unknown subcommand", "git $SUB push", wt, ""},
		{"absolute git path", "/usr/bin/git push", wt, ""}, // not recognised: documented non-goal
	})
}

// Git commands are found wherever the shell would run them; data is not commands.
func TestCheckCommandShellSyntax(t *testing.T) {
	e := newEnv(t)
	wt := e.worktree
	cases := []dirCase{
		{"and-or chain", "git status && git log || git push", wt, "Push blocked"},
		{"pipeline", "echo a | git push", wt, "Push blocked"},
		{"background", "git push &", wt, "Push blocked"},
		{"if condition", "if git push; then :; fi", wt, "Push blocked"},
		{"else body", "if false; then :; else git push; fi", wt, "Push blocked"},
		{"while body", "while true; do git push; done", wt, "Push blocked"},
		{"for items substitution", "for i in $(git push); do :; done", wt, "Push blocked"},
		{"case body", "case x in x) git push;; esac", wt, "Push blocked"},
		{"negation", "! git push", wt, "Push blocked"},
		{"backticks", "echo `git push`", wt, "Push blocked"},
		{"substitution in quotes", "echo \"x $(git push)\"", wt, "Push blocked"},
		{"nested substitution", "echo $(echo $(git push))", wt, "Push blocked"},
		{"assignment substitution", "X=$(git push) true", wt, "Push blocked"},
		{"process substitution", "cat <(git push)", wt, "Push blocked"},
		{"redirect target substitution", "echo > $(git push)", wt, "Push blocked"},
		{"heredoc body is data", "cat <<EOF > notes.md\ngit push\nEOF", wt, ""},
		{"heredoc then command", "cat <<EOF > n.md\nx\nEOF\ngit push", wt, "Push blocked"},
		{"unquoted heredoc substitution", "cat <<EOF\n$(git push)\nEOF", wt, "Push blocked"},
		{"quoted heredoc is data", "cat <<'EOF'\n$(git push)\nEOF", wt, ""},
		{"comment", "echo ok # git push", wt, ""},
		{"line continuation", "git \\\n push", wt, "Push blocked"},
		{"parse error fails open", "git push; ) (", wt, ""},
	}
	checkAll(t, e, cases)
}

func TestCommitMessageHeredocIsNotCommands(t *testing.T) {
	msg := "git commit -m \"$(cat <<'EOF'\nfeat(ci): git push\n\ngit push now\nEOF\n)\""
	var cfg team.Config
	cfg.Team.Review.Trigger = "pre-push"
	if got := checkCommand(msg, t.TempDir(), cfg); got != "" {
		t.Errorf("valid heredoc message denied: %q", got)
	}
	bad := "git commit -m \"$(cat <<'EOF'\nnope\nEOF\n)\""
	if got := checkCommand(bad, t.TempDir(), cfg); !strings.Contains(got, "not Angular") {
		t.Errorf("bad heredoc message not denied: %q", got)
	}
}

func TestCommitMessages(t *testing.T) {
	cases := []struct {
		cmd    string
		scopes []string
		denied string
	}{
		{`git commit -m"bad subject"`, nil, "not Angular"},
		{`git commit -am"bad subject"`, nil, "not Angular"},
		{`git commit -m"feat: ok"`, nil, ""},
		{`git commit -m 'feat: ok'`, nil, ""},
		{`git commit --message=bad`, nil, "not Angular"},
		{`git commit --message "feat: ok"`, nil, ""},
		{`git commit -a -m "bad"`, nil, "not Angular"},
		{`git commit -m "feat(ci): ok"`, []string{"ci"}, ""},
		{`git commit -m "feat: ok"`, []string{"ci"}, "needs a scope"},
		{`git commit -m "feat: $(date)"`, nil, ""},
		{`git commit -m "$MSG"`, nil, ""},
		{`git commit -m`, nil, ""},
		{`git commit -m "$(date)"`, nil, ""},
		{"git commit -m \"$(cat <<'EOF'\nfeat(ci): ok\n\nbody\nEOF\n)\"", []string{"ci"}, ""},
		{"git commit -m \"$(cat <<'EOF'\nnope\nEOF\n)\"", nil, "not Angular"},
		{"git commit -m \"$(cat <<EOF\nnope\nEOF\n)\"", nil, "not Angular"},
		{"git commit -m \"$(cat <<-EOF\n\tfeat: ok\n\tEOF\n)\"", nil, ""},
		{"git commit -m \"$(cat <<'EOF'\nfeat: ok\nEOF\n)\" && git commit -m bad", nil, "not Angular"},
		{`git commit -F msg.txt`, nil, ""},
		{`git commit --no-edit`, nil, ""},
		{`git commit -Skeyname -m "feat: ok"`, nil, ""},
		{`git commit -S -m "feat: ok"`, nil, ""},
		{`git commit -Smain -m "bad"`, nil, "not Angular"},
		{`git commit -Cmain`, nil, ""},
		{`git commit -C main`, nil, ""},
		{`git commit -c HEAD -m "bad"`, nil, "not Angular"},
		{`git commit -Fmsg.txt`, nil, ""},
		{`git commit -F msg.txt -m "bad"`, nil, "not Angular"},
		{`git commit -tmsg -m "feat: ok"`, nil, ""},
		{`git commit -umno -m "feat: ok"`, nil, ""},
		{`git commit -sam"bad"`, nil, "not Angular"},
		{`git commit --author "A <a@example.com>" -m "feat: ok"`, nil, ""},
		{`git commit --file msg.txt`, nil, ""},
		{`git commit -m "bad" -Skey`, nil, "not Angular"},
		{`git commit -- -m`, nil, ""},
	}
	for _, c := range cases {
		cfg := scoped(c.scopes...)
		got := checkCommand(c.cmd, "/nonexistent", cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s: reason %q, want substring %q", c.cmd, got, c.denied)
		}
	}
}

func TestCommitCheckIgnoresDirectory(t *testing.T) {
	e := newEnv(t)
	for _, cmd := range []string{
		`cd "$SOMEWHERE" && git commit -m "bad subject"`,
		`git -C "$SOMEWHERE" commit -m "bad subject"`,
		`git --git-dir="$D" commit -m "bad subject"`,
		`GIT_DIR=$D git commit -m "bad subject"`,
		`cd $X && cd $Y && git commit -am"bad subject"`,
	} {
		if got := checkCommand(cmd, e.worktree, e.cfg); !strings.Contains(got, "not Angular") {
			t.Errorf("%s: reason %q, want the commit message deny", cmd, got)
		}
	}
	if got := checkCommand(`cd "$X" && git commit -m "feat: fine"`, e.worktree, e.cfg); got != "" {
		t.Errorf("valid message denied: %q", got)
	}
	// git commit --no-edit and a $(...) message still pass
	if got := checkCommand(`cd "$X" && git commit -m "$(date)"`, e.worktree, e.cfg); got != "" {
		t.Errorf("unresolvable message denied: %q", got)
	}
}
