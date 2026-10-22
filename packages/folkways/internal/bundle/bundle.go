// Package bundle is what a folkways build distributes: the framework content and its version.
package bundle

import (
	"io/fs"

	"github.com/sumsolution/folkways/packages/content"
)

// Bundle is the content a command copies from and the version it records. Tests pass a synthetic FS
// (fstest.MapFS) laid out like packages/content: instructions.md, settings.json, skills/, prompts/,
// workflows/, templates/.
type Bundle struct {
	FS      fs.FS
	Version string
}

// Embedded is the content built into this binary, at the binary's version.
func Embedded(version string) Bundle {
	return Bundle{FS: content.FS, Version: version}
}
