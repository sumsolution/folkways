package gitguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sumsolution/folkways/packages/folkways/internal/team"
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

// A top-level `NAME=literal` is tracked, so a path held in a variable keeps the directory known. Anything
// that could change the name in a way the walker can't follow makes it unknown, and so the directory.
func TestCheckCommandShellVariables(t *testing.T) {
	e := newEnv(t)
	wt, main := e.worktree, e.main
	set := "W=" + wt + "; "
	checkAll(t, e, []dirCase{
		{"-C semicolon", set + `git -C "$W" push`, "/", "Push blocked"},
		{"-C and", "W=" + wt + ` && git -C "$W" push`, "/", "Push blocked"},
		{"-C unquoted", set + "git -C $W push", "/", "Push blocked"},
		{"-C braces", set + "git -C ${W} push", "/", "Push blocked"},
		{"cd", "W=" + wt + ` && cd "$W" && git push`, "/", "Push blocked"},
		{"main clone", "M=" + main + `; git -C "$M" checkout x`, wt, "main clone"},
		{"chained", "A=" + filepath.Dir(wt) + `; W=$A/app-wt; git -C "$W" push`, "/", "Push blocked"},
		{"same statement", "A=" + filepath.Dir(wt) + ` W=$A/app-wt; git -C "$W" push`, "/", "Push blocked"},
		{"reassigned", "W=/elsewhere; " + set + `git -C "$W" push`, "/", "Push blocked"},
		{"set again after removal", set + "W=$(pwd); " + set + `git -C "$W" push`, "/", "Push blocked"},
		{"later in an and chain", set + `true && git -C "$W" push`, "/", "Push blocked"},
		{"first of an and chain", "W=" + wt + ` && true; git -C "$W" push`, "/", "Push blocked"},
		{"allowlisted command", set + `echo hi; git status; git -C "$W" push`, "/", "Push blocked"},
		{"subshell sees it", set + `(git -C "$W" push)`, "/", "Push blocked"},
		{"set in subshell", "(" + set + `git -C "$W" push)`, "/", "Push blocked"},
		{"subshell does not leak", set + `(W=/elsewhere); git -C "$W" push`, "/", "Push blocked"},
		{"substitution sees it", set + `echo $(git -C "$W" push)`, "/", "Push blocked"},
		{"substitution does not leak", set + `echo $(W=/elsewhere); git -C "$W" push`, "/", "Push blocked"},

		// Removed: the name is unknown from here on.
		{"unresolvable value", set + `W=$(pwd); git -C "$W" push`, "/", ""},
		{"append", set + `W+=/x; git -C "$W" push`, "/", ""},
		{"index", set + `W[0]=/x; git -C "$W" push`, "/", ""},
		{"array", set + `W=(/x); git -C "$W" push`, "/", ""},
		{"set in if", set + `if true; then W=/x; fi; git -C "$W" push`, "/", ""},
		{"set in brace group in pipeline", set + `{ W=/x; } | true; git -C "$W" push`, "/", ""},
		{"set in pipeline", set + `W=/x | true; git -C "$W" push`, "/", ""},
		{"set in background", set + `W=/x & git -C "$W" push`, "/", ""},
		{"set in ||", set + `false || W=/x; git -C "$W" push`, "/", ""},
		{"prefix assignment", set + `W=/x true; git -C "$W" push`, "/", ""},
		{"export", set + `export W=/x; git -C "$W" push`, "/", ""},
		{"declare", set + `declare W=/x; git -C "$W" push`, "/", ""},
		{"local", set + `local W=/x; git -C "$W" push`, "/", ""},
		{"readonly", set + `readonly W=/x; git -C "$W" push`, "/", ""},
		{"typeset", set + `typeset W=/x; git -C "$W" push`, "/", ""},
		{"for", set + `for W in /x; do :; done; git -C "$W" push`, "/", ""},
		{"arithmetic", set + `((W=1)); git -C "$W" push`, "/", ""},
		{"arithmetic expansion", set + `: $((W=1)); git -C "$W" push`, "/", ""},
		{"increment", set + `((W++)); git -C "$W" push`, "/", ""},
		{"default assign", set + `: ${W:=/x}; git -C "$W" push`, "/", ""},
		{"default assign plain", set + `: ${W=/x}; git -C "$W" push`, "/", ""},
		{"read", set + `read W < f; git -C "$W" push`, "/", ""},
		{"read -r", set + `read -r W; git -C "$W" push`, "/", ""},
		{"mapfile", set + `mapfile W < f; git -C "$W" push`, "/", ""},
		{"readarray", set + `readarray -t W < f; git -C "$W" push`, "/", ""},
		{"getopts", set + `getopts ab W; git -C "$W" push`, "/", ""},
		{"printf -v", set + `printf -v W /x; git -C "$W" push`, "/", ""},
		{"unset", set + `unset W; git -C "$W" push`, "/", ""},
		{"command read", set + `command read W; git -C "$W" push`, "/", ""},
		{"builtin read", set + `builtin read W; git -C "$W" push`, "/", ""},
		{"cd in an and chain", `false && cd ` + wt + `; git push`, "/", ""},
		{"PWD after cd", "PWD=" + wt + `; cd /; git -C "$PWD" push`, "/", ""},
		{"coproc", set + `coproc W { :; }; git -C "$W" push`, "/", ""},
		{"quoted let", set + `let "W=5"; git -C "$W" push`, "/", ""},

		// Stop tracking: nothing is known for the rest of the line.
		{"function", set + `f() { :; }; git -C "$W" push`, "/", ""},
		{"set after function", `f() { :; }; ` + set + `git -C "$W" push`, "/", ""},
		{"trap", set + `trap 'W=/x' DEBUG; git -C "$W" push`, "/", ""},
		{"nameref", set + `declare -n R=W; git -C "$W" push`, "/", ""},
		{"eval", set + `eval "W=/x"; git -C "$W" push`, "/", ""},
		{"set after eval", `eval :; ` + set + `git -C "$W" push`, "/", ""},
		{"source", set + `source ./env.sh; git -C "$W" push`, "/", ""},
		{"dot", set + `. ./env.sh; git -C "$W" push`, "/", ""},
		{"unresolved command name", set + `$CMD 'W=/x'; git -C "$W" push`, "/", ""},
		{"printf -v joined", set + `printf -vW /x; git -C "$W" push`, "/", ""},
		{"read into an index", set + `read 'W[0]'; git -C "$W" push`, "/", ""},
		{"command -p eval", set + `command -p eval W=/x; git -C "$W" push`, "/", ""},
		{"command -- eval", set + `command -- eval W=/x; git -C "$W" push`, "/", ""},
		{"exec fd variable", set + `exec {W}>/dev/null; git -C "$W" push`, "/", ""},
		{"wait -p", set + `wait -p W; git -C "$W" push`, "/", ""},
		{"any other command", set + `./build.sh; git -C "$W" push`, "/", ""},

		// Set back later in an and chain: the step that restores it may not run.
		{"restored in an and chain", set + `true && W=/b && false && ` + set + `git -C "$W" push`, "/", ""},
		{"cd back in an and chain", "cd " + wt + " && true && cd /b && false && cd " + wt + "; git push", "/", ""},
		{"fd variable redirect", set + `echo hi {W}>/dev/null; git -C "$W" push`, "/", ""},
		{"fd variable on true", set + `true {W}>/dev/null; git -C "$W" push`, "/", ""},
		{"arithmetic index", set + `echo $((W[0]=5)); git -C "$W" push`, "/", ""},
		{"underscore", "_=" + wt + `; echo hi; git -C "$_" push`, "/", ""},
	})
}

// A loop body is walked a second time only when the first pass changed something, so nesting stays linear.
func TestCheckCommandNestedLoops(t *testing.T) {
	e := newEnv(t)
	e.cfg.Team.Review.Trigger = "on-demand" // a push that passes, so every pass runs its checks
	cmd := "git push"
	for range 10 {
		cmd = "for i in a b; do " + cmd + "; done"
	}
	start := time.Now()
	checkAll(t, e, []dirCase{{"nested loops", cmd, e.worktree, ""}})
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("10 nested loops took %v", d)
	}
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

func TestCommitConvention(t *testing.T) {
	cases := []struct {
		convention, pattern, cmd, denied string
	}{
		{"pattern", `^[A-Z]+-\d+: .+`, `git commit -m "EX-12: add orders"`, ""},
		{"pattern", `^[A-Z]+-\d+: .+`, `git commit -m "feat: add orders"`, "git.commit_pattern"},
		{"pattern", `^[A-Z]+-\d+: .+`, `git commit -m "feat: add orders"`, `^[A-Z]+-\d+: .+`},
		{"pattern", `^[a-z]`, `git commit -m "lowercase wins"`, ""}, // not Angular, still passes
		{"pattern", "", `git commit -m "anything"`, ""},             // empty: no check
		{"pattern", "([", `git commit -m "anything"`, ""},           // invalid: fail open
		{"none", "", `git commit -m "anything"`, ""},
		{"angular", "", `git commit -m "anything"`, "not Angular"},
		{"", "", `git commit -m "anything"`, "not Angular"}, // missing key: angular
	}
	for _, c := range cases {
		var cfg team.Config
		cfg.Team.Git.CommitConvention, cfg.Team.Git.CommitPattern = c.convention, c.pattern
		cfg.Team.Git.Scopes = []string{"ci"} // scopes only count for angular
		got := checkCommand(c.cmd, "/nonexistent", cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s %q %s: reason %q, want substring %q", c.convention, c.pattern, c.cmd, got, c.denied)
		}
	}
}
