package team

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func cfg(team, me string) Config {
	var c Config
	c.Team.Review.Trigger = team
	c.Me.Review.Trigger = me
	return c
}

func TestReviewTriggerFloor(t *testing.T) {
	cases := []struct{ team, me, want string }{
		{"pre-push", "", "pre-push"},
		{"pre-push", "on-demand", "pre-push"}, // me cannot loosen
		{"pre-push", "after-commit", "pre-push"},
		{"on-demand", "pre-push", "pre-push"}, // me can tighten
		{"on-demand", "after-commit", "after-commit"},
		{"after-commit", "on-demand", "after-commit"},
		{"", "pre-push", "pre-push"},
		{"", "", ""},
		{"pre-push", "bogus", "pre-push"}, // unknown values are ignored
		{"bogus", "after-commit", "after-commit"},
		{"bogus", "bogus2", ""},
	}
	for _, c := range cases {
		if got := cfg(c.team, c.me).ReviewTrigger(); got != c.want {
			t.Errorf("team=%q me=%q: got %q, want %q", c.team, c.me, got, c.want)
		}
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("folkways.yaml", "git:\n  scopes: [a, b]\nreview:\n  trigger: on-demand\n")
	write(".local/me.yaml", "projects_dir: /p\nreview:\n  trigger: pre-push\n")
	c := LoadDir(dir)
	if len(c.Team.Git.Scopes) != 2 || c.Me.ProjectsDir != "/p" || c.ReviewTrigger() != "pre-push" {
		t.Errorf("unexpected config: %+v", c)
	}
	if c := LoadDir(t.TempDir()); c.ReviewTrigger() != "" {
		t.Errorf("empty dir should give the zero config, got %+v", c)
	}
}

func TestSafeHandle(t *testing.T) {
	cases := map[string]string{"alex": "alex", "": "", "..": "", "a/b": "", `a\b`: "", "../x": "", ".": "", "a..b": "", "a b": "", "a\nb": "", "a\tb": ""}
	for in, want := range cases {
		var c Config
		c.Me.Handle = in
		if got := c.SafeHandle(); got != want {
			t.Errorf("SafeHandle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestConfiguredAndToggles(t *testing.T) {
	dir := t.TempDir()
	if LoadDir(dir).Configured {
		t.Error("a dir without folkways.yaml must not be configured")
	}
	body := "mode: solo\nhooks:\n  precompact: false\n  git-guard: true\ngit: {commit_convention: pattern, commit_pattern: \"^x\"}\n"
	if err := os.WriteFile(filepath.Join(dir, "folkways.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c := LoadDir(dir)
	if !c.Configured || !c.Solo() || c.Team.Git.CommitConvention != "pattern" || c.Team.Git.CommitPattern != "^x" {
		t.Errorf("config not read: %+v", c.Team)
	}
	for name, want := range map[string]bool{"precompact": false, "git-guard": true, "context-band": true} {
		if got := c.HookEnabled(name); got != want {
			t.Errorf("HookEnabled(%s) = %v, want %v", name, got, want)
		}
	}
	if c.PersonalRoot() != dir {
		t.Errorf("solo personal root = %q, want the repo root", c.PersonalRoot())
	}
}

func TestPersonalRootTeam(t *testing.T) {
	c := Config{Dir: "/r"}
	if c.PersonalRoot() != "" {
		t.Error("a team repo with no handle has no personal root")
	}
	c.Me.Handle = "alex"
	if c.PersonalRoot() != filepath.Join("/r", "people", "alex") {
		t.Errorf("team personal root = %q", c.PersonalRoot())
	}
	c.Me.Handle = "../x"
	if c.PersonalRoot() != "" {
		t.Error("an unsafe handle has no personal root")
	}
}

func TestPushApproval(t *testing.T) {
	cases := []struct{ kind, team, me, want string }{
		{"brain", "", "", "never"}, // brain repos push without asking by default
		{"app", "", "", "always"},  // app repos ask by default
		{"brain", "always", "", "always"},
		{"brain", "always", "never", "never"},                  // me can loosen
		{"brain", "never", "default-branch", "default-branch"}, // me can tighten
		{"app", "never", "", "never"},
		{"app", "default-branch", "always", "always"},
		{"app", "bogus", "", "always"}, // unknown values are ignored
		{"brain", "", "bogus", "never"},
		{"app", "", "never", "never"}, // me overrides the default too
	}
	for _, c := range cases {
		var cfg Config
		cfg.Team.Push = map[string]string{c.kind: c.team}
		cfg.Me.Push = map[string]string{c.kind: c.me}
		if got := cfg.PushApproval(c.kind); got != c.want {
			t.Errorf("%s team=%q me=%q: got %q, want %q", c.kind, c.team, c.me, got, c.want)
		}
	}
}

func TestKnowledgeByPR(t *testing.T) {
	cases := []struct {
		mode, team, me string
		want           bool
	}{
		{"", "", "", true},      // a team repo lands knowledge/ by PR
		{"solo", "", "", false}, // a solo repo's landing is the person's choice
		{"", "direct", "", false},
		{"solo", "pr", "", true},
		{"", "", "direct", true}, // me.yaml can't change it
		{"", "bogus", "", true},
	}
	for _, c := range cases {
		var cfg Config
		cfg.Team.Mode = c.mode
		cfg.Team.Push = map[string]string{"knowledge": c.team}
		cfg.Me.Push = map[string]string{"knowledge": c.me}
		if got := cfg.KnowledgeByPR(); got != c.want {
			t.Errorf("mode=%q team=%q me=%q: got %v, want %v", c.mode, c.team, c.me, got, c.want)
		}
	}
}

func TestReviewSkipPaths(t *testing.T) {
	cases := []struct {
		yaml string
		want []string
	}{
		{"review:\n  trigger: pre-push\n", []string{"docs/**"}},
		{"review:\n  skip_paths: []\n", []string{}},
		{"review:\n  skip_paths: [\"docs/**\", \"**/*.md\"]\n", []string{"docs/**", "**/*.md"}},
	}
	for _, c := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "folkways.yaml"), []byte(c.yaml), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := LoadDir(dir).ReviewSkipPaths(); !slices.Equal(got, c.want) || (got == nil) != (c.want == nil) {
			t.Errorf("%q: ReviewSkipPaths = %#v, want %#v", c.yaml, got, c.want)
		}
	}
}
