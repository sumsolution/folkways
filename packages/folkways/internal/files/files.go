// Package files holds the file helpers sync, init and doctor share.
package files

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
)

// Ops are the file operations sync stages with. A seam, so tests can inject failures. Copy reads src from
// fsys (the bundle's content) and writes dest on disk.
type Ops struct {
	Copy   func(fsys fs.FS, src, dest string) error
	Rename func(from, to string) error
}

// OS is Ops on the real filesystem.
var OS = Ops{Copy: CopyFS, Rename: os.Rename}

// CopyFS copies src in fsys to dest on disk, creating or truncating dest. dest is 0755 if src has an
// executable bit, else 0644: embedded files carry no permission bits worth keeping.
func CopyFS(fsys fs.FS, src, dest string) error {
	b, err := fs.ReadFile(fsys, src)
	if err != nil {
		return err
	}
	info, err := fs.Stat(fsys, src)
	if err != nil {
		return err
	}
	perm := fs.FileMode(0o644)
	if info.Mode()&0o111 != 0 {
		perm = 0o755
	}
	if err := os.WriteFile(dest, b, perm); err != nil {
		return err
	}
	return os.Chmod(dest, perm) // WriteFile keeps an existing file's mode
}

// SHA256 is the hex sha256 of a file's content.
func SHA256(file string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return SHA256Text(b), nil
}

// SHA256Text is the hex sha256 of content.
func SHA256Text(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// ignored are names List skips.
func ignored(name string) bool { return name == ".gitkeep" || name == ".DS_Store" }

// List returns the files under dir as sorted slash paths relative to it. Empty if dir is missing.
// It lists symlinks to files, as the TS CLI does. Unlike the TS CLI it doesn't descend symlinked folders
// (no loops) and skips dangling links instead of failing; framework content and skill folders have neither.
func List(dir string) ([]string, error) {
	return ListFS(os.DirFS(dir), ".")
}

// ListFS is List over an fs.FS, such as the embedded content, rooted at dir.
func ListFS(fsys fs.FS, dir string) ([]string, error) {
	var out []string
	err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == dir && errors.Is(err, fs.ErrNotExist) {
				return fs.SkipAll
			}
			return err
		}
		if p == dir || ignored(d.Name()) {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			info, err := fs.Stat(fsys, p)
			if err != nil || !info.Mode().IsRegular() {
				return nil
			}
		} else if !d.Type().IsRegular() {
			return nil
		}
		rel := p
		if dir != "." {
			rel = p[len(dir)+1:]
		}
		out = append(out, path.Clean(rel))
		return nil
	})
	sort.Strings(out)
	return out, err
}

// PruneEmptyDirs deletes empty directories under dir, bottom up, and dir itself if it ends up empty.
func PruneEmptyDirs(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := PruneEmptyDirs(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	if entries, err = os.ReadDir(dir); err != nil {
		return err
	}
	if len(entries) == 0 {
		return os.Remove(dir)
	}
	return nil
}
