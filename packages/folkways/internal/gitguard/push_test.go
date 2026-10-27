package gitguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// pushRepos is a brain repo cloned from a bare origin, with a session worktree, and an app repo.
type pushRepos struct {
	brain, worktree, app string
}

func commitFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "docs: "+rel)
}

func clonedRepo(t *testing.T, root, name string) string {
	t.Helper()
	origin := filepath.Join(root, name+".git")
	seed := filepath.Join(root, name+"-seed")
	gitIn(t, root, "init", "-q", "--bare", "-b", "main", origin)
	gitIn(t, root, "init", "-q", "-b", "main", seed)
	commitFile(t, seed, "README.md", name+"\n")
	gitIn(t, seed, "push", "-q", origin, "main")
	clone := filepath.Join(root, name)
	gitIn(t, root, "clone", "-q", origin, clone)
	return clone
}

func newPushRepos(t *testing.T) pushRepos {
	t.Helper()
	root := t.TempDir()
	r := pushRepos{brain: clonedRepo(t, root, "brain"), app: clonedRepo(t, root, "app")}
	r.worktree = filepath.Join(root, "worktrees", "brain", "session-x")
	gitIn(t, r.brain, "worktree", "add", "-q", "-b", "session-x", r.worktree, "origin/main")
	return r
}

func (r pushRepos) cfg() team.Config { return team.Config{Dir: r.brain} }

func TestBrainPushesWithoutAskingByDefault(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.worktree, "efforts/x/00-STATUS.md", "status\n")
	for _, cmd := range []string{"git push origin HEAD:main", "git push origin HEAD:refs/heads/main", "git push -u origin session-x"} {
		if got := evaluate(cmd, r.worktree, r.cfg()); got != (verdict{}) {
			t.Errorf("%s: got %+v, want nothing", cmd, got)
		}
	}
	if got := evaluate("git push", r.brain, r.cfg()); got != (verdict{}) {
		t.Errorf("push from the main clone: got %+v, want nothing", got)
	}
}

func TestBrainPushOfKnowledgeToMainIsDenied(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.worktree, "knowledge/fact.md", "fact\n")
	for _, cmd := range []string{"git push origin HEAD:main", "git push origin +HEAD:main", "git push --force-with-lease origin session-x:main"} {
		if got := evaluate(cmd, r.worktree, r.cfg()); !strings.Contains(got.deny, "knowledge/") {
			t.Errorf("%s: got %+v, want a knowledge/ deny", cmd, got)
		}
	}
	if got := evaluate("git push -u origin HEAD:docs/fact", r.worktree, r.cfg()); got.deny != "" {
		t.Errorf("knowledge/ on a PR branch must pass, got %+v", got)
	}
	commitFile(t, r.brain, "knowledge/other.md", "other\n") // the main clone tracks origin/main
	if got := evaluate("git push", r.brain, r.cfg()); !strings.Contains(got.deny, "knowledge/") {
		t.Errorf("plain push from main: got %+v, want a knowledge/ deny", got)
	}
}

func TestBrainPushApprovalFromConfig(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.worktree, "state/now.md", "now\n")
	cfg := r.cfg()
	cfg.Team.Push = map[string]string{"brain": "always"}
	if got := evaluate("git push origin HEAD:session-x", r.worktree, cfg); got.ask == "" || got.deny != "" {
		t.Errorf("always: got %+v, want ask", got)
	}
	cfg.Team.Push = map[string]string{"brain": "default-branch"}
	if got := evaluate("git push origin HEAD:main", r.worktree, cfg); got.ask == "" {
		t.Errorf("default-branch, push to main: got %+v, want ask", got)
	}
	if got := evaluate("git push origin HEAD:session-x", r.worktree, cfg); got != (verdict{}) {
		t.Errorf("default-branch, push to a branch: got %+v, want nothing", got)
	}
}

func TestBrainWorktreeSkipsTheReviewGate(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.worktree, "state/now.md", "now\n")
	cfg := r.cfg()
	cfg.Team.Review.Trigger = "pre-push"
	if got := evaluate("git push origin HEAD:main", r.worktree, cfg); got.deny != "" {
		t.Errorf("a brain worktree is the brain repo: got %+v", got)
	}
}

func TestAppPushAsksByDefault(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.app, "main.go", "package main\n")
	if got := evaluate("git push", r.app, r.cfg()); got.ask == "" || got.deny != "" {
		t.Errorf("app push: got %+v, want ask", got)
	}
	cfg := r.cfg()
	cfg.Team.Push = map[string]string{"app": "never"}
	if got := evaluate("git push", r.app, cfg); got != (verdict{}) {
		t.Errorf("app never: got %+v, want nothing", got)
	}
	cfg.Team.Review.Trigger = "pre-push"
	if got := evaluate("git push", r.app, cfg); got.deny == "" {
		t.Errorf("the review gate still denies: got %+v", got)
	}
}

func TestPushInUnknownDirectoryAsks(t *testing.T) {
	r := newPushRepos(t)
	if got := evaluate(`cd "$X" && git push`, r.brain, r.cfg()); got.ask == "" {
		t.Errorf("unknown directory: got %+v, want ask", got)
	}
	cfg := r.cfg()
	cfg.Team.Push = map[string]string{"app": "never"}
	if got := evaluate(`cd "$X" && git push`, r.brain, cfg); got != (verdict{}) {
		t.Errorf("unknown directory, both never: got %+v, want nothing", got)
	}
}

func TestReviewInABrainWorktreeCounts(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.app, "main.go", "package main\n")
	head := gitIn(t, r.app, "rev-parse", "HEAD")
	review := filepath.Join(r.worktree, "efforts", "x", "01-review.md")
	if err := os.MkdirAll(filepath.Dir(review), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(review, []byte("---\nsha: "+head+"\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := r.cfg()
	cfg.Team.Review.Trigger = "pre-push"
	if got := evaluate("git push", r.app, cfg); got.deny != "" {
		t.Errorf("a review written in the session worktree must count, got %+v", got)
	}
}

func TestSoloBrainKeepsKnowledgeAndAsks(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.worktree, "knowledge/fact.md", "fact\n")
	cfg := r.cfg()
	cfg.Team.Mode = "solo"
	got := evaluate("git push origin HEAD:main", r.worktree, cfg)
	if got.deny != "" || got.ask == "" {
		t.Errorf("solo: knowledge lands where the user chooses, and pushes ask by default; got %+v", got)
	}
}

func TestDefaultBranchFallsBackToMaster(t *testing.T) {
	r := newPushRepos(t)
	gitIn(t, r.worktree, "push", "-q", "origin", "HEAD:refs/heads/master")
	gitIn(t, r.worktree, "fetch", "-q", "origin")
	gitIn(t, r.brain, "remote", "set-head", "origin", "-d")
	gitIn(t, r.brain, "update-ref", "-d", "refs/remotes/origin/main") // as if origin had only master
	commitFile(t, r.worktree, "knowledge/fact.md", "fact\n")
	if got := evaluate("git push origin HEAD:master", r.worktree, r.cfg()); !strings.Contains(got.deny, "knowledge/") {
		t.Errorf("no origin/HEAD and only master: got %+v, want a knowledge/ deny", got)
	}
}

func TestWorktreeReviewStaysInThePersonsFolder(t *testing.T) {
	r := newPushRepos(t)
	commitFile(t, r.app, "main.go", "package main\n")
	head := gitIn(t, r.app, "rev-parse", "HEAD")
	inbox := filepath.Join(r.worktree, "people", "ann", "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "2026-10-27-review-"+head[:7]+".md"), []byte("review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := r.cfg()
	cfg.Team.Review.Trigger = "pre-push"
	cfg.Me.Handle = "ann"
	if got := evaluate("git push", r.app, cfg); got.deny != "" {
		t.Errorf("ann's inbox review in the worktree must count, got %+v", got)
	}
	cfg.Me.Handle = "bob"
	if got := evaluate("git push", r.app, cfg); got.deny == "" {
		t.Error("another person's inbox review must not count")
	}
}
