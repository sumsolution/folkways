package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWarnings(t *testing.T) {
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	if w := Warnings(); len(w) != 1 || w[0] != NotOnPath {
		t.Errorf("folkways not on PATH: %q", w)
	}
	if err := os.WriteFile(filepath.Join(bin, "folkways"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if w := Warnings(); len(w) != 0 {
		t.Errorf("folkways on PATH: %q", w)
	}
}
