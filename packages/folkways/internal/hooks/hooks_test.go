package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// From sync.test.ts: 'pnpm team with %s'. true: the workspace hoists the hook binary.
func TestHoistsHooks(t *testing.T) {
	cases := map[string]bool{
		"publicHoistPattern:\n  - \"*eslint*\"\n":                                                    false,
		"catalog:\n  x: \"@folkways/hooks-*\"\n":                                                     false,
		"packages: []\npublicHoistPattern:\n  - \"*eslint*\"\n  - '@folkways/hooks-*'\n":             true,
		"publicHoistPattern: [\"@folkways/hooks-*\"]\n":                                              true,
		"publicHoistPattern:\n  - \"*\"\n":                                                           true,
		"shamefullyHoist: true\n":                                                                    true,
		"publicHoistPattern: # hoisted\n  # the hook binary\n  - \"@folkways/hooks-*\" # folkways\n": true,
		"publicHoistPattern:\n  - @folkways/hooks-* # folkways\n":                                    true,
		"publicHoistPattern:\n  - \"*eslint*\" # not \"@folkways/hooks-*\"\n":                        false,
		"publicHoistPattern:\n  - \"a'b\"\n":                                                         false,
		"publicHoistPattern:\n  - \"@folkways/hooks-*'\n":                                            false,
	}
	for yaml, want := range cases {
		if got := HoistsHooks(yaml); got != want {
			t.Errorf("HoistsHooks(%q) = %v, want %v", yaml, got, want)
		}
	}
}

func TestScalar(t *testing.T) {
	for in, want := range map[string]string{`"a#b" # c`: "a#b", `a # c`: "a", ` 'x' `: "x", `"open`: `"open`} {
		if got := scalar(in); got != want {
			t.Errorf("scalar(%q) = %q, want %q", in, got, want)
		}
	}
}

func put(t *testing.T, team, name, content string) {
	t.Helper()
	path := filepath.Join(team, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWarnings(t *testing.T) {
	if w := Warnings(t.TempDir()); len(w) != 0 {
		t.Errorf("npm, no node_modules: %q", w)
	}

	team := t.TempDir()
	put(t, team, "package.json", `{"packageManager": "pnpm@10.0.0"}`)
	if w := Warnings(team); len(w) != 1 || !strings.Contains(w[0], `publicHoistPattern: ["@folkways/hooks-*"]`) {
		t.Errorf("pnpm without workspace: %q", w)
	}
	put(t, team, "pnpm-workspace.yaml", PnpmWorkspace)
	if w := Warnings(team); len(w) != 0 {
		t.Errorf("pnpm with workspace: %q", w)
	}

	team = t.TempDir()
	put(t, team, "package.json", `{"packageManager": "yarn@4.0.0"}`)
	put(t, team, "node_modules/x", "")
	if w := Warnings(team); len(w) != 1 || !strings.HasPrefix(w[0], Bin+" is missing") {
		t.Errorf("yarn, no hook binary: %q", w)
	}
	put(t, team, Bin, "")
	if w := Warnings(team); len(w) != 0 {
		t.Errorf("with hook binary: %q", w)
	}
}
