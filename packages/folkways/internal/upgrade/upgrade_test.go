package upgrade

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/git"
	"github.com/sumsolution/folkways/packages/folkways/internal/scaffold"
)

// newTeam is a team repo initialised at version, with git isolated from the caller's config.
func newTeam(t *testing.T, version string) string {
	t.Helper()
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
	if _, err := scaffold.Init(team, bundle.Embedded(version)); err != nil {
		t.Fatal(err)
	}
	return team
}

// The unit cases the upgrade fixtures can't express (fixtures/upgrade/COVERAGE.md): a sync that fails, and a
// running version other than the one the harness builds.

func TestFailedSyncLeavesNothingBehind(t *testing.T) {
	team := newTeam(t, "0.0.1")
	before, _ := git.Run(team, "rev-parse", "HEAD")
	broken := bundle.Bundle{FS: fstest.MapFS{}, Version: "0.2.0"} // no content: sync fails
	if _, err := Upgrade(team, "0.2.0", broken); err == nil || !strings.HasPrefix(err.Error(), "upgrade to 0.2.0 failed; back on main:") {
		t.Fatalf("err = %v", err)
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

func TestOtherVersionNeedsThatBinary(t *testing.T) {
	team := newTeam(t, "0.0.1")
	_, err := Upgrade(team, "0.2.0", bundle.Embedded("0.1.0"))
	if err == nil || err.Error() != "this is folkways 0.1.0; install folkways 0.2.0, then rerun 'folkways upgrade 0.2.0'" {
		t.Fatalf("err = %v", err)
	}
	if out, _ := git.Run(team, "branch", "--list", "chore/*"); strings.TrimSpace(out) != "" {
		t.Errorf("branch created: %s", out)
	}
}

func TestUpgradeCommitsOnABranch(t *testing.T) {
	team := newTeam(t, "0.0.1")
	r, err := Upgrade(team, "v0.2.0", bundle.Embedded("0.2.0"))
	if err != nil || r != (Result{Branch: "chore/folkways-0.2.0", From: "0.0.1", To: "0.2.0"}) {
		t.Fatalf("Upgrade = %+v, %v", r, err)
	}
	if log, _ := git.Run(team, "log", "-1", "--format=%s"); log != "chore(folkways): upgrade to 0.2.0" {
		t.Errorf("commit = %q", log)
	}
	raw, _ := os.ReadFile(filepath.Join(team, ".folkways", "manifest.json"))
	if !strings.Contains(string(raw), `"version": "0.2.0"`) {
		t.Errorf("manifest not moved: %s", raw)
	}
}

// Upgrading to the version the manifest already records commits nothing, but runs the safety checks first.
func TestSameVersionIsNoop(t *testing.T) {
	team := newTeam(t, "0.0.1")
	r, err := Upgrade(team, "v0.0.1", bundle.Embedded("0.3.0"))
	if err != nil || r.Branch != "" || r.From != "0.0.1" || r.To != "0.0.1" {
		t.Fatalf("Upgrade = %+v, %v", r, err)
	}
	if out, _ := git.Run(team, "branch", "--list", "chore/*"); strings.TrimSpace(out) != "" {
		t.Errorf("branch created: %s", out)
	}
	if err := os.WriteFile(filepath.Join(team, "stray.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Upgrade(team, "0.0.1", bundle.Embedded("0.0.1")); err == nil || err.Error() != "working tree is dirty; commit or stash first" {
		t.Fatalf("dirty tree: err = %v", err)
	}
}
