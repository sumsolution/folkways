package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "x"},
	} {
		if _, err := Run(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	if clean, err := IsClean(dir); err != nil || !clean {
		t.Errorf("IsClean = %v, %v; want true", clean, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if clean, err := IsClean(dir); err != nil || clean {
		t.Errorf("IsClean with an untracked file = %v, %v; want false", clean, err)
	}
	if !BranchExists(dir, "main") || BranchExists(dir, "nope") {
		t.Error("BranchExists: want main only")
	}
	_, err := Run(dir, "rev-parse", "--verify", "nope")
	var gitErr *Error
	if !errors.As(err, &gitErr) || !strings.Contains(err.Error(), "git rev-parse --verify nope") || !strings.Contains(gitErr.Stderr, "fatal:") {
		t.Errorf("Run error = %#v; want an *Error with the command and git's stderr", err)
	}
}
