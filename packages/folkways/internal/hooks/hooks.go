// Package hooks names the framework's hooks and checks that the committed hook entries can run.
package hooks

import "os/exec"

// Names are the hooks the binary runs; packages/content/settings.json wires each one.
var Names = []string{"context-band", "git-guard", "precompact", "session-start"}

// NotOnPath is the warning when the hooks' guard won't find the binary.
const NotOnPath = "folkways isn't on PATH, so its hooks are off; install it: https://github.com/sumsolution/folkways#install"

// Warnings is NotOnPath when `folkways` isn't on PATH: committed hook entries call it there, behind a guard
// that exits 0 without it (ADR 0012).
func Warnings() []string {
	if _, err := exec.LookPath("folkways"); err != nil {
		return []string{NotOnPath}
	}
	return nil
}
