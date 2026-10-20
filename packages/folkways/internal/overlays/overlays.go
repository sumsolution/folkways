// Package overlays compiles a team overlay (overlays/<skill>/OVERLAY.md, ADR 0001) into a framework skill.
// sync calls CompileSkillFiles; the body is ported in go-cli Phase 1 round 1.
package overlays

import "errors"

// Compiled is the result of applying an overlay to a skill.
type Compiled struct {
	Files    map[string]string // only the files that change, by path relative to the skill folder
	Warnings []string
}

// CompileSkillFiles applies overlay (the OVERLAY.md text) to a skill's files, keyed by path relative to the
// skill folder.
func CompileSkillFiles(files map[string]string, overlay string) (Compiled, error) {
	return Compiled{}, errors.New("overlays are not ported to Go yet")
}
