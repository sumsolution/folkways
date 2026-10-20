// Package doctor checks a team repo (`folkways doctor`). Ported in go-cli Phase 1 round 2.
package doctor

import (
	"errors"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
)

// Finding is one problem, with its fix in the message. Level is "error" or "warning".
type Finding struct {
	Level   string
	Message string
}

// Options for Doctor.
type Options struct {
	Home string // where personal skills live (~/.claude/skills, ...)
}

// Doctor checks config, versions, overlays, skills, settings and the hook install. b is the running build.
func Doctor(team string, b bundle.Bundle, opts Options) ([]Finding, error) {
	return nil, errors.New("doctor is not ported to Go yet")
}
