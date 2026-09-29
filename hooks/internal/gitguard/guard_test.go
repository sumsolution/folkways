package gitguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/hooks/internal/team"
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
		{`git commit -m "feat(auth): x"`, scoped("web-app", "ci"), "not in team.yaml"},
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

func TestCheckCommandDirectories(t *testing.T) {
	e := newEnv(t)
	wt, main := e.worktree, e.main
	cases := []struct {
		name   string
		cmd    string
		cwd    string
		denied string
	}{
		{"cd then push", "cd " + wt + " && git push", "/", "Push blocked"},
		{"failed cd falls back to cwd", "cd /nonexistent-dir || git push", wt, "Push blocked"},
		{"cd and: only the new dir", "cd " + wt + " && git checkout x", main, ""},
		{"cd and: main clone", "cd " + main + " && git checkout x", wt, "main clone"},
		{"cd semicolon: new dir", "cd " + main + "; git checkout x", wt, "main clone"},
		{"cd semicolon: original dir", "cd " + wt + "; git checkout x", main, "main clone"},
		{"cd or: original dir only", "cd " + main + " || git checkout x", wt, ""},
		{"cd and-or: either dir", "cd " + wt + " && git status || git checkout x", main, "main clone"}, // cd failed: still in the original main clone
		{"cd in pipeline does not persist", "cd " + main + " | git checkout x", wt, ""},
		{"cd in background does not persist", "cd " + main + " & git checkout x", wt, ""},
		{"subshell cd does not leak", "(cd " + main + ") && git checkout x", wt, ""},
		{"subshell push", "(cd " + wt + " && git push)", "/", "Push blocked"},
		{"subshell main clone", "(cd " + main + " && git reset --hard HEAD)", wt, "main clone"},
		{"brace group", "{ cd " + wt + "; git push; }", "/", "Push blocked"},
		{"command substitution", "cd " + wt + " && echo $(git push)", "/", "Push blocked"},
		{"backticks", "cd " + wt + " && echo `git push`", "/", "Push blocked"},
		{"substitution in quotes", "cd " + wt + " && echo \"x $(git push)\"", "/", "Push blocked"},
		{"heredoc body is data", "cat <<EOF > notes.md\ngit push\nEOF", wt, ""},
		{"heredoc then command", "cat <<EOF > n.md\nx\nEOF\ngit push", wt, "Push blocked"},
		{"heredoc same line chain", "cat <<EOF && git push\nx\nEOF", wt, "Push blocked"},
		{"global option --no-pager", "cd " + wt + " && git --no-pager push", "/", "Push blocked"},
		{"global option main clone", "cd " + main + " && git --no-pager reset --hard", wt, "main clone"},
		{"global options with operands", "git -c a=b --git-dir " + main + "/.git --work-tree " + main + " -C " + wt + " push", "/", "Push blocked"},
		{"global options with equals", "git -C " + wt + " --git-dir=.git --namespace=y --exec-path=z --config-env=a=B -p -P --bare --paginate --no-replace-objects --no-optional-locks --literal-pathspecs --glob-pathspecs --noglob-pathspecs --icase-pathspecs --no-lazy-fetch push", "/", "Push blocked"},
		{"-C relative to cd", "cd " + filepath.Dir(wt) + " && git -C app-wt push", "/", "Push blocked"},
	}
	for _, c := range cases {
		got := checkCommand(c.cmd, c.cwd, e.cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s: %q: reason %q, want substring %q", c.name, c.cmd, got, c.denied)
		}
	}
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

func TestCheckCommandShellSyntax(t *testing.T) {
	e := newEnv(t)
	home := filepath.Dir(e.main) // projects/
	t.Setenv("HOME", filepath.Dir(home))
	rel := "projects/app-wt"
	wt, main := e.worktree, e.main
	cases := []struct {
		name   string
		cmd    string
		cwd    string
		denied string
	}{
		{"$HOME", `cd "$HOME/` + rel + `" && git push`, "/", "Push blocked"},
		{"${HOME}", `cd ${HOME}/` + rel + ` && git push`, "/", "Push blocked"},
		{"tilde", "cd ~/" + rel + " && git push", "/", "Push blocked"},
		{"tilde main clone", "cd ~/projects/app && git reset --hard HEAD", wt, "main clone"},
		{"quoted tilde is literal", `cd "~/` + rel + `" && git push`, "/", ""},
		{"unknown cd then and: fail open", "cd $ELSEWHERE && git push", wt, ""},
		{"unknown cd then or: original dir", "cd $ELSEWHERE || git push", wt, "Push blocked"},
		{"unknown cd then semicolon: original dir", "cd $ELSEWHERE; git push", wt, "Push blocked"},
		{"unknown cd then relative cd stays unknown", "cd $ELSEWHERE && cd sub && git push", wt, ""},
		{"unknown git dir option", "git -C $ELSEWHERE push", wt, ""},
		{"unknown subcommand", "git $SUB push", wt, ""},
		{"plain cd goes home", "cd && git checkout x", wt, ""},
		{"cd dash is unknown", "cd - && git push", wt, ""},
		{"attached -C", "git -C" + wt + " push", "/", "Push blocked"},
		{"attached -c", "cd " + main + " && git -ccore.x=y reset --hard", wt, "main clone"},
		{"git-dir and work-tree", "git --git-dir=" + main + "/.git --work-tree=" + main + " reset --hard", wt, "main clone"},
		{"git-dir only, .git suffix", "git --git-dir=" + main + "/.git reset --hard", wt, "main clone"},
		{"git-dir separate operand", "git --git-dir " + wt + "/.git push", "/", "Push blocked"},
		{"git-dir relative to -C", "git -C " + filepath.Dir(wt) + " --git-dir=app-wt/.git push", "/", "Push blocked"},
		{"work-tree wins", "git --git-dir=" + main + "/.git --work-tree=" + wt + " push", "/", "Push blocked"},
		{"if body", "if true; then cd " + wt + "; git push; fi", "/", "Push blocked"},
		{"if condition", "if git push; then :; fi", wt, "Push blocked"},
		{"else body", "if false; then :; else git push; fi", wt, "Push blocked"},
		{"while body", "while true; do git push; done", wt, "Push blocked"},
		{"for body", "for i in 1 2; do git push; done", wt, "Push blocked"},
		{"for items substitution", "for i in $(git push); do :; done", wt, "Push blocked"},
		{"case body", "case x in x) git push;; esac", wt, "Push blocked"},
		{"case word substitution", "case $(git push) in x) :;; esac", wt, "Push blocked"},
		{"function body", "f() { git push; }", wt, "Push blocked"},
		{"time", "time git push", wt, "Push blocked"},
		{"negation", "! git push", wt, "Push blocked"},
		{"pipeline", "echo a | git push", wt, "Push blocked"},
		{"background", "git push &", wt, "Push blocked"},
		{"declare with substitution", "export X=$(git push)", wt, "Push blocked"},
		{"assignment substitution", "X=$(git push) true", wt, "Push blocked"},
		{"[[ ]] substitution", "[[ $(git push) ]]", wt, "Push blocked"},
		{"arithmetic substitution", "echo $(( $(git push) + 1 ))", wt, "Push blocked"},
		{"process substitution", "cat <(git push)", wt, "Push blocked"},
		{"redirect target substitution", "echo > $(git push)", wt, "Push blocked"},
		{"nested substitution", "echo $(echo $(git push))", wt, "Push blocked"},
		{"substitution inherits cd", "cd " + wt + " && echo $(git push)", "/", "Push blocked"},
		{"substitution cd does not leak", "echo $(cd " + main + ") && git checkout x", wt, ""},
		{"unquoted heredoc substitution", "cat <<EOF\n$(git push)\nEOF", wt, "Push blocked"},
		{"quoted heredoc is data", "cat <<'EOF'\n$(git push)\nEOF", wt, ""},
		{"double-quoted delimiter heredoc is data", "cat <<\"EOF\"\n$(git push)\nEOF", wt, ""},
		{"heredoc with dash", "cat <<-EOF\n\t$(git push)\n\tEOF", wt, "Push blocked"},
		{"comment", "echo ok # git push", wt, ""},
		{"line continuation", "git \\\n push", wt, "Push blocked"},
		{"parse error fails open", "git push; ) (", wt, ""},
		{"unterminated quote fails open", "git push 'abc", wt, ""},
		{"absolute git path", "/usr/bin/git push", wt, ""}, // not recognised: documented non-goal
	}
	for _, c := range cases {
		got := checkCommand(c.cmd, c.cwd, e.cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s: %q: reason %q, want substring %q", c.name, c.cmd, got, c.denied)
		}
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
	}
	for _, c := range cases {
		cfg := scoped(c.scopes...)
		got := checkCommand(c.cmd, "/nonexistent", cfg)
		if (c.denied == "") != (got == "") || !strings.Contains(got, c.denied) {
			t.Errorf("%s: reason %q, want substring %q", c.cmd, got, c.denied)
		}
	}
}
