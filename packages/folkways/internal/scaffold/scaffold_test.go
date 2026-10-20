package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
)

// The unit cases the init fixtures can't express (fixtures/init/COVERAGE.md): templateTarget, and
// packageName with a folder name longer than NAME_MAX or empty.

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

func TestPackageName(t *testing.T) {
	for in, want := range map[string]string{
		"My Team":                "my-team",
		".hidden":                "hidden",
		"équipe":                 "equipe",
		strings.Repeat("x", 300): strings.Repeat("x", 214),
		"a\U0001F600b":           "a--b", // JS replaces each UTF-16 unit of an astral character
		"Team_A.B~c":             "team_a.b-c",
	} {
		got, err := packageName(in)
		if err != nil || got != want {
			t.Errorf("packageName(%.20q) = %.20q, %v; want %.20q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "...", "node_modules", "favicon.ico", "中文"} {
		if got, err := packageName(in); err == nil {
			t.Errorf("packageName(%q) = %q, want an error", in, got)
		}
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
	if _, err := Init(dest, bundle.Bundle{}, Options{}); err == nil {
		t.Error("Init on an unreadable folder succeeded")
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("folder is gone: %v", err)
	}
}
