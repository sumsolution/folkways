package upgrade

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/git"
	"github.com/sumsolution/folkways/packages/folkways/internal/scaffold"
)

// The unit case the upgrade fixtures can't express (fixtures/upgrade/COVERAGE.md): a sync that fails after
// leaving files behind. It needs an installer that replaces node_modules/.bin/folkways.
func TestFailedSyncLeavesNothingBehind(t *testing.T) {
	for k, v := range map[string]string{
		"GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1",
		"GIT_AUTHOR_NAME": "test", "GIT_AUTHOR_EMAIL": "test@example.com",
		"GIT_COMMITTER_NAME": "test", "GIT_COMMITTER_EMAIL": "test@example.com",
	} {
		t.Setenv(k, v)
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	team := filepath.Join(t.TempDir(), "team")
	if _, err := scaffold.Init(team, bundle.Embedded("0.0.1"), scaffold.Options{PackageManager: "npm"}); err != nil {
		t.Fatal(err)
	}
	before, _ := git.Run(team, "rev-parse", "HEAD")
	install := func(team, version string) error {
		bin := filepath.Join(team, "node_modules/.bin/folkways")
		if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
			return err
		}
		script := "#!/bin/sh\necho partial > stray.txt\necho 'sync exploded' >&2\nexit 1\n"
		return os.WriteFile(bin, []byte(script), 0o755)
	}
	_, err := Upgrade(team, "0.2.0", install)
	if err == nil || err.Error() != "upgrade to 0.2.0 failed; back on main:\nsync exploded" {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(team, "stray.txt")); statErr == nil {
		t.Error("stray.txt left behind")
	}
	if out, _ := git.Run(team, "status", "--porcelain"); strings.TrimSpace(out) != "" {
		t.Errorf("tree not clean: %s", out)
	}
	if git.BranchExists(team, "chore/folkways-0.2.0") {
		t.Error("branch left behind")
	}
	if after, _ := git.Run(team, "rev-parse", "HEAD"); after != before {
		t.Error("HEAD moved")
	}
}

// Upgrading to the version the manifest already records returns before any git or install work.
func TestSameVersionIsNoop(t *testing.T) {
	for k, v := range map[string]string{
		"GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1",
		"GIT_AUTHOR_NAME": "test", "GIT_AUTHOR_EMAIL": "test@example.com",
		"GIT_COMMITTER_NAME": "test", "GIT_COMMITTER_EMAIL": "test@example.com",
	} {
		t.Setenv(k, v)
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	team := filepath.Join(t.TempDir(), "team")
	if _, err := scaffold.Init(team, bundle.Embedded("0.0.1"), scaffold.Options{PackageManager: "npm"}); err != nil {
		t.Fatal(err)
	}
	install := func(team, version string) error {
		t.Error("installer called")
		return nil
	}
	r, err := Upgrade(team, "v0.0.1", install)
	if err != nil || r.Branch != "" || r.From != "0.0.1" || r.To != "0.0.1" {
		t.Fatalf("Upgrade = %+v, %v", r, err)
	}
	if out, _ := git.Run(team, "branch", "--list", "chore/*"); strings.TrimSpace(out) != "" {
		t.Errorf("branch created: %s", out)
	}
}
