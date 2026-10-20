package files

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func put(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	put(t, filepath.Join(dir, "b/x.md"), "x")
	put(t, filepath.Join(dir, "a.md"), "a")
	put(t, filepath.Join(dir, "b/.gitkeep"), "")
	put(t, filepath.Join(dir, ".DS_Store"), "")
	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("a.md", filepath.Join(dir, "link.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("b", filepath.Join(dir, "dirlink")); err != nil {
		t.Fatal(err)
	}
	got, err := List(dir)
	if want := []string{"a.md", "b/x.md", "link.md"}; err != nil || !slices.Equal(got, want) {
		t.Errorf("List = %q, %v; want %q", got, err, want)
	}
	if got, err := List(filepath.Join(dir, "missing")); err != nil || len(got) != 0 {
		t.Errorf("List(missing) = %q, %v; want empty", got, err)
	}
}

func TestSHA256(t *testing.T) {
	const empty = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := SHA256Text(nil); got != empty {
		t.Errorf("SHA256Text(nil) = %s", got)
	}
	f := filepath.Join(t.TempDir(), "f")
	put(t, f, "")
	if got, err := SHA256(f); err != nil || got != empty {
		t.Errorf("SHA256 = %s, %v", got, err)
	}
}

func TestCopyFileKeepsMode(t *testing.T) {
	dir := t.TempDir()
	src, dest := filepath.Join(dir, "src"), filepath.Join(dir, "dest")
	put(t, src, "#!/bin/sh\n")
	if err := os.Chmod(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(src, dest); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dest)
	if err != nil || info.Mode().Perm() != 0o755 {
		t.Errorf("dest mode = %v, %v; want 0755", info.Mode(), err)
	}
}

func TestPruneEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	for _, d := range []string{"a/b/c", "keep/empty"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	put(t, filepath.Join(dir, "keep/f"), "")
	if err := PruneEmptyDirs(dir); err != nil {
		t.Fatal(err)
	}
	got, _ := filepath.Glob(filepath.Join(dir, "*", "*"))
	if want := []string{filepath.Join(dir, "keep/f")}; !slices.Equal(got, want) {
		t.Errorf("left %q, want %q", got, want)
	}
	if err := PruneEmptyDirs(filepath.Join(dir, "missing")); err != nil {
		t.Errorf("missing dir: %v", err)
	}
}
