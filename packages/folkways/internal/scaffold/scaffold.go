// Package scaffold creates a team repo (`folkways init`). Ported in go-cli Phase 1 round 2.
package scaffold

import (
	"errors"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
)

// Options are init's flags.
type Options struct {
	Install        bool   // install dependencies to create the lockfile
	PackageManager string // npm or pnpm
}

// Init scaffolds a team repo at path, pins b.Version, syncs and commits. It returns the repo's absolute path.
func Init(path string, b bundle.Bundle, opts Options) (string, error) {
	return "", errors.New("init is not ported to Go yet")
}
