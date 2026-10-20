// Package git runs git for the CLI.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run runs git in cwd and returns trimmed stdout. A non-zero exit is an error that carries git's stderr.
func Run(cwd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// IsClean reports whether the work tree at cwd has no changes, staged or not, and no untracked files.
func IsClean(cwd string) (bool, error) {
	out, err := Run(cwd, "status", "--porcelain")
	return out == "", err
}

// BranchExists reports whether the local branch exists.
func BranchExists(cwd, branch string) bool {
	_, err := Run(cwd, "rev-parse", "--quiet", "--verify", "refs/heads/"+branch)
	return err == nil
}
