package pkgmgr

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func put(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// From package-manager.test.ts.
func TestDetect(t *testing.T) {
	if pm, err := Detect(t.TempDir()); pm != "npm" || err != nil {
		t.Errorf("default = %q, %v; want npm", pm, err)
	}

	team := t.TempDir()
	put(t, team, "package.json", `{"packageManager": "pnpm@12.4.1+sha512.abc"}`)
	put(t, team, "package-lock.json", "{}")
	if pm, err := Detect(team); pm != "pnpm" || err != nil {
		t.Errorf("packageManager field = %q, %v; want pnpm", pm, err)
	}

	for lockfile, want := range map[string]string{"pnpm-lock.yaml": "pnpm", "package-lock.json": "npm"} {
		team := t.TempDir()
		put(t, team, "package.json", "{}")
		put(t, team, lockfile, "")
		if pm, err := Detect(team); pm != want || err != nil {
			t.Errorf("%s = %q, %v; want %s", lockfile, pm, err, want)
		}
	}

	team = t.TempDir()
	put(t, team, "package.json", `{"packageManager": "yarn@4.0.0"}`)
	if _, err := Detect(team); err == nil || err.Error() != "yarn is not supported; use npm or pnpm" {
		t.Errorf("yarn: %v", err)
	}
}

func TestFromUserAgent(t *testing.T) {
	cases := map[string]string{
		"pnpm/12.4.1 npm/? node/v24.1.0 darwin arm64":             "pnpm",
		"npm/11.19.0 node/v24.14.1 darwin arm64 workspaces/false": "npm",
		"yarn/4.0.0 npm/? node/v24.1.0":                           "",
		"":                                                        "",
	}
	for ua, want := range cases {
		if got := FromUserAgent(ua); got != want {
			t.Errorf("FromUserAgent(%q) = %q, want %q", ua, got, want)
		}
	}
}

func TestCommands(t *testing.T) {
	if got := InstallCommand("pnpm"); !slices.Equal(got, []string{"pnpm", "install"}) {
		t.Errorf("InstallCommand = %q", got)
	}
	if got := AddDevExactCommand("npm", "x@1.0.0"); !slices.Equal(got, []string{"npm", "install", "--save-dev", "--save-exact", "x@1.0.0"}) {
		t.Errorf("npm = %q", got)
	}
	if got := AddDevExactCommand("pnpm", "x@1.0.0"); !slices.Equal(got, []string{"pnpm", "add", "--save-dev", "--save-exact", "x@1.0.0"}) {
		t.Errorf("pnpm = %q", got)
	}
}
