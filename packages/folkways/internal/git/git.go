// Package git runs git for the CLI.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Error is a failed git call. Stderr is git's raw stderr, for callers that print it as is.
type Error struct {
	Args   []string
	Stderr string
	Err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("git %s: %v: %s", strings.Join(e.Args, " "), e.Err, strings.TrimSpace(e.Stderr))
}

func (e *Error) Unwrap() error { return e.Err }

// Run runs git in cwd and returns trimmed stdout. A failure is an *Error.
func Run(cwd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", &Error{Args: args, Stderr: stderr.String(), Err: err}
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

// Detail is what to show for a failed call: git's trimmed stderr, else the error's own text.
func Detail(err error) string {
	var ge *Error
	if errors.As(err, &ge) {
		if s := strings.TrimSpace(ge.Stderr); s != "" {
			return s
		}
	}
	return err.Error()
}
