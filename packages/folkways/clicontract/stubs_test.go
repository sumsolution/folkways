package clicontract

import (
	"os"
	"path/filepath"
)

// writeStubs puts the CLI under test on PATH as ${BIN}/folkways, where committed hook entries and doctor
// look for it (ADR 0012). A fixture deletes it to test a missing binary.
func (e *env) writeStubs() {
	script := "#!/bin/sh\nexec " + e.vars["CLI"] + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(e.vars["BIN"], "folkways"), []byte(script), 0o755); err != nil {
		e.t.Fatal(err)
	}
}
