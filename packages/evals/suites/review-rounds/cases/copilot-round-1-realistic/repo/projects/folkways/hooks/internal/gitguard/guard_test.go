package gitguard

import (
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
