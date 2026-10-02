package sessionstart

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/hooks/internal/output"
	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

func run(t *testing.T, cfg team.Config) string {
	t.Helper()
	var buf bytes.Buffer
	output.Stdout = &buf
	t.Cleanup(func() { output.Stdout = os.Stdout })
	Run(payload.Payload{}, cfg)
	return buf.String()
}

func TestOfferOnce(t *testing.T) {
	xdg, repo := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("FOLKWAYS_SKILL_NAMESPACE", "folkways")
	cfg := team.Config{Dir: repo}
	if got := run(t, cfg); !strings.Contains(got, "`folkways:setup` skill") {
		t.Fatalf("first run should offer folkways:setup, got %q", got)
	}
	if got := run(t, cfg); got != "" {
		t.Errorf("second run should be silent, got %q", got)
	}
	raw, _ := os.ReadFile(filepath.Join(xdg, "folkways", "offered"))
	if strings.TrimSpace(string(raw)) != repo {
		t.Errorf("offered = %q, want %q", raw, repo)
	}
	if entries, _ := os.ReadDir(repo); len(entries) != 0 {
		t.Errorf("the hook wrote into the repo: %v", entries)
	}
}

func TestOfferSilentWhenFileUnusable(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	if err := os.MkdirAll(filepath.Join(xdg, "folkways", "offered"), 0o755); err != nil { // a directory: unreadable
		t.Fatal(err)
	}
	if got := run(t, team.Config{Dir: t.TempDir()}); got != "" {
		t.Errorf("unusable offered file should mean no offer, got %q", got)
	}
}

func TestConfiguredRemindsStart(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if got := run(t, team.Config{Dir: t.TempDir(), Configured: true}); !strings.Contains(got, "`start` skill") {
		t.Errorf("configured repo should remind start, got %q", got)
	}
}

func TestRunNamesTheStartSkill(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("FOLKWAYS_SKILL_NAMESPACE", "")
	if got := run(t, team.Config{Dir: t.TempDir(), Configured: true}); !strings.Contains(got, "follow the `start` skill") {
		t.Errorf("session-start printed %q", got)
	}
}

func TestNoOfferInASubfolderOfAConfiguredRepo(t *testing.T) {
	xdg, repo := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	if err := os.WriteFile(filepath.Join(repo, "folkways.yaml"), []byte("mode: solo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	sub := filepath.Join(repo, "docs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := run(t, team.Config{Dir: sub}); got != "" {
		t.Errorf("subfolder of a configured repo should not offer, got %q", got)
	}
	if _, err := os.Stat(filepath.Join(xdg, "folkways", "offered")); err == nil {
		t.Error("nothing should be recorded")
	}
}

func TestRunNamesTheStartSkillInANamespace(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("FOLKWAYS_SKILL_NAMESPACE", "folkways")
	if got := run(t, team.Config{Dir: t.TempDir(), Configured: true}); !strings.Contains(got, "follow the `folkways:start` skill") {
		t.Errorf("session-start printed %q", got)
	}
}
