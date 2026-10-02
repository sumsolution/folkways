package gitguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestMainCloneExemptsTheFolkwaysRepo(t *testing.T) {
	projects := t.TempDir()
	repo := filepath.Join(projects, "folkways-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q", "-b", "main")
	cfg := team.Config{Dir: repo}
	cfg.Me.ProjectsDir = projects
	if got := checkMainClone("pull", repo, cfg); got != "" {
		t.Errorf("the folkways repo must be exempt, got %q", got)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	cfg.Dir = link // symlinked project dir
	if got := checkMainClone("pull", repo, cfg); got != "" {
		t.Errorf("symlinked folkways repo must be exempt, got %q", got)
	}
	cfg.Dir = t.TempDir()
	if got := checkMainClone("pull", repo, cfg); got == "" {
		t.Error("another main clone must still be denied")
	}
}
