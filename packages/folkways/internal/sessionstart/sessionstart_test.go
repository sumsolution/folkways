package sessionstart

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/folkways/internal/output"
	"github.com/sumsolution/folkways/packages/folkways/internal/payload"
	"github.com/sumsolution/folkways/packages/folkways/internal/team"
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

func TestConfigBlock(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("FOLKWAYS_SKILL_NAMESPACE", "")
	write := func(dir, name, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	load := func(dir string) string {
		var out struct {
			AdditionalContext string `json:"additionalContext"`
		}
		if err := json.Unmarshal([]byte(run(t, team.LoadDir(dir))), &out); err != nil {
			t.Fatal(err)
		}
		return out.AdditionalContext
	}

	d := t.TempDir()
	write(d, "folkways.yaml", "review:\n  trigger: after-commit\n")
	write(d, ".local/me.yaml", "handle: alex\n")
	got := load(d)
	for _, want := range []string{
		"for this session.\n\nfolkways config (",
		"- personal root: people/alex\n- review trigger: after-commit\n",
		"folkways.yaml:\n```yaml\nreview:\n  trigger: after-commit\n```\n\n.local/me.yaml:\n```yaml\nhandle: alex\n```",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}

	solo := t.TempDir()
	write(solo, "folkways.yaml", "mode: solo\n")
	got = load(solo)
	for _, want := range []string{"- personal root: .\n", "- review trigger: none\n", ".local/me.yaml: missing. Run the `onboard` skill to create it."} {
		if !strings.Contains(got, want) {
			t.Errorf("solo: missing %q in %q", want, got)
		}
	}

	big := t.TempDir()
	write(big, "folkways.yaml", "# "+strings.Repeat("x", 4096)+"\n")
	write(big, ".local/me.yaml", "handle: ../x\n")
	got = load(big)
	for _, want := range []string{"- personal root: none\n", "folkways.yaml: over 4096 bytes, not shown. Read the file.", "handle: ../x"} {
		if !strings.Contains(got, want) {
			t.Errorf("big: missing %q in %q", want, got)
		}
	}

	unreadable := t.TempDir()
	write(unreadable, "folkways.yaml", "mode: solo\n")
	write(unreadable, ".local/me.yaml/x", "")
	got = load(unreadable)
	if want := ".local/me.yaml: can't be read. Check the file."; !strings.Contains(got, want) {
		t.Errorf("unreadable: missing %q in %q", want, got)
	}
}
