// Package files holds the file helpers sync, init and doctor share.
package files

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
)

// Ops are the file operations sync stages with. A seam, so tests can inject failures.
type Ops struct {
	Copy   func(src, dest string) error
	Rename func(from, to string) error
}

// OS is Ops on the real filesystem.
var OS = Ops{Copy: CopyFile, Rename: os.Rename}

// CopyFile copies src to dest, creating or truncating dest, with src's permission bits.
func CopyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
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
// It follows symlinks to files, as the TS CLI's statSync does.
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
