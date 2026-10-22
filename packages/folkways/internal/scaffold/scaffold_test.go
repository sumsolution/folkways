package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
)

// The unit cases the init fixtures can't express (fixtures/init/COVERAGE.md): templateTarget, and the
// cases below.

func TestTemplateTarget(t *testing.T) {
	for in, want := range map[string]string{
		"dot-gitignore":                  ".gitignore",
		"dot-claude/settings.json":       ".claude/settings.json",
		"AGENTS.md.tmpl":                 "AGENTS.md",
		"a/dot-b/dot-c.tmpl":             "a/.b/.c",
		"README.md":                      "README.md",
		"sub/notdot-x":                   "sub/notdot-x",
		"dot-dot-x":                      ".dot-x",
		"people/README.md":               "people/README.md",
		"efforts/dot-gitkeep.tmpl.tmpl2": "efforts/.gitkeep.tmpl.tmpl2",
	} {
		if got := templateTarget(in); got != want {
			t.Errorf("templateTarget(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInitWritesNoNodeFiles(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for k, v := range map[string]string{"GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@example.com", "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@example.com"} {
		t.Setenv(k, v)
	}
	dest, err := Init(filepath.Join(t.TempDir(), "Équipe (A)"), bundle.Embedded("1.2.3"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"package.json", "pnpm-workspace.yaml", "node_modules"} {
		if _, err := os.Stat(filepath.Join(dest, f)); err == nil {
			t.Errorf("init wrote %s", f)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dest, ".folkways", "manifest.json"))
	if err != nil || !strings.Contains(string(raw), `"version": "1.2.3"`) {
		t.Errorf("manifest: %s, %v", raw, err)
	}
}

// A folder init can't list is an error, and init must not delete it (TS throws EACCES and keeps it).
func TestInitKeepsUnreadableFolder(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can list any folder")
	}
	dest := filepath.Join(t.TempDir(), "team")
	if err := os.Mkdir(dest, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dest, 0o755) })
	if _, err := Init(dest, bundle.Bundle{}); err == nil {
		t.Error("Init on an unreadable folder succeeded")
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("folder is gone: %v", err)
	}
}
