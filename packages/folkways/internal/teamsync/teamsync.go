// Package teamsync copies the framework content into a team repo (`folkways sync`). The API is fixed so
// init, upgrade and doctor can be ported against it; the bodies are ported in go-cli Phase 1 round 1.
package teamsync

import (
	"errors"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/manifest"
)

var errNotPorted = errors.New("sync is not ported to Go yet")

// Options are sync's flags.
type Options struct {
	// Force discards edits to framework-owned files instead of stopping.
	Force bool
}

// Result is what a sync did.
type Result struct {
	Version   string
	Previous  string // the manifest's version before the sync; "" without a manifest
	Discarded []string
	Warnings  []string
}

// Sync copies b's files into the team repo at team. It does not commit.
func Sync(team string, b bundle.Bundle, opts Options) (Result, error) {
	return Result{}, errNotPorted
}

// FindEdits lists framework-owned files changed since the last sync, e.g. "modified  .folkways/instructions.md".
func FindEdits(team string, m *manifest.Manifest) ([]string, error) {
	return nil, errNotPorted
}

// StaleOverlays lists team overlays changed, added or removed since the last sync compiled them.
func StaleOverlays(team string, m *manifest.Manifest) ([]string, error) {
	return nil, errNotPorted
}

// EditsMessage is the refusal for edited framework files; fix is its last line.
func EditsMessage(edits []string, fix string) string {
	lines := []string{"framework files were edited:"}
	for _, e := range edits {
		lines = append(lines, "  "+e)
	}
	return strings.Join(append(lines, "Move team changes into overlays/, then rerun.", fix), "\n")
}
