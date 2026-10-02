package gitguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// commitFile writes a file, commits it and returns the new SHA.
func commitFile(t *testing.T, dir, rel, body string) string {
	t.Helper()
	write(t, dir, rel, body)
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "x")
	return gitIn(t, dir, "rev-parse", "HEAD")
}

func soloRepo(t *testing.T) (string, team.Config) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	var cfg team.Config
	cfg.Dir = dir
	cfg.Team.Mode = "solo"
	cfg.Team.Review.Trigger = "pre-push"
	return dir, cfg
}

func TestSoloPushWalksBackOverWorkAndInbox(t *testing.T) {
	dir, cfg := soloRepo(t)
	code := commitFile(t, dir, "src/a.txt", "a")
	commitFile(t, dir, "work/e1/r.md", "---\nsha: "+code+"\n---\n")
	if got := checkPush(dir, cfg); got != "" {
		t.Errorf("review committed on top of the code should pass, got %q", got)
	}
	// Another code commit moves the walk past the review.
	commitFile(t, dir, "src/b.txt", "b")
	if got := checkPush(dir, cfg); got == "" {
		t.Error("a code commit after the review must be denied")
	}
	// Team mode never walks back.
	teamCfg := cfg
	teamCfg.Team.Mode = "team"
	teamCfg.Me.Handle = "alex"
	if got := checkPush(dir, teamCfg); got == "" {
		t.Error("team mode must not walk back")
	}
}

func TestSoloPushWalkLimit(t *testing.T) {
	dir, cfg := soloRepo(t)
	code := commitFile(t, dir, "src/a.txt", "a")
	commitFile(t, dir, "work/e1/r.md", "---\nsha: "+code+"\n---\n")
	for i := 1; i < maxWalkBack; i++ {
		commitFile(t, dir, "inbox/n"+strconv.Itoa(i)+".md", "n")
	}
	// code <- 20 work/inbox commits: the code commit is the 20th step back.
	if got := checkPush(dir, cfg); got != "" {
		t.Errorf("review %d steps back should pass, got %q", maxWalkBack, got)
	}
	commitFile(t, dir, "inbox/extra.md", "n")
	if got := checkPush(dir, cfg); got == "" {
		t.Error("review beyond the walk limit must be denied")
	}
}

func TestWalkBackStopsAtMergeAndRoot(t *testing.T) {
	dir, _ := soloRepo(t)
	root := commitFile(t, dir, "inbox/a.md", "a")
	if got := walkBack(dir, root); len(got) != 1 || got[0] != root {
		t.Errorf("root commit walk = %v", got)
	}
	gitIn(t, dir, "checkout", "-q", "-b", "side")
	side := commitFile(t, dir, "inbox/s.md", "s")
	gitIn(t, dir, "checkout", "-q", "main")
	commitFile(t, dir, "inbox/m.md", "m")
	gitIn(t, dir, "merge", "-q", "--no-ff", "-m", "merge", side)
	merge := gitIn(t, dir, "rev-parse", "HEAD")
	if got := walkBack(dir, merge); len(got) != 1 || got[0] != merge {
		t.Errorf("merge walk = %v", got)
	}
}

func TestSoloWalkOnlyInTheFolkwaysRepo(t *testing.T) {
	dir, cfg := soloRepo(t)
	code := commitFile(t, dir, "src/a.txt", "a")
	commitFile(t, dir, "work/e1/r.md", "---\nsha: "+code+"\n---\n")
	other := t.TempDir()
	gitIn(t, other, "init", "-q", "-b", "main")
	commitFile(t, other, "src/o.txt", "other")
	if got := checkPush(other, cfg); got == "" {
		t.Error("another repo has no review and must be denied")
	}
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
