// Package scaffold creates a team repo (`folkways init`). It needs no Node: the content is embedded.
package scaffold

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
	"github.com/sumsolution/folkways/packages/folkways/internal/files"
	"github.com/sumsolution/folkways/packages/folkways/internal/git"
	"github.com/sumsolution/folkways/packages/folkways/internal/teamsync"
)

// templateTarget maps a template path: `dot-x` becomes `.x`, `name.tmpl` becomes `name`.
func templateTarget(rel string) string {
	parts := strings.Split(rel, "/")
	for i, p := range parts {
		if rest, ok := strings.CutPrefix(p, "dot-"); ok {
			parts[i] = "." + rest
		}
	}
	return strings.TrimSuffix(strings.Join(parts, "/"), ".tmpl")
}

// Init scaffolds a team repo at path, syncs b (the manifest records b.Version, the pin) and commits. It
// returns the repo's absolute path. On failure it leaves nothing behind, so a rerun works.
func Init(path string, b bundle.Bundle) (string, error) {
	dest, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(dest)
	existed := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err // e.g. unreadable or a file: never treat it as absent, or cleanup would delete it
	}
	if existed && len(entries) > 0 {
		return "", clierr.New("%s exists and is not empty", dest)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}
	if err := scaffold(dest, b); err != nil {
		if existed {
			entries, _ := os.ReadDir(dest)
			for _, e := range entries {
				os.RemoveAll(filepath.Join(dest, e.Name()))
			}
		} else {
			os.RemoveAll(dest)
		}
		return "", err
	}
	return dest, nil
}

func scaffold(dest string, b bundle.Bundle) error {
	list, err := files.ListFS(b.FS, "templates")
	if err != nil {
		return err
	}
	for _, rel := range list {
		out := filepath.Join(dest, filepath.FromSlash(templateTarget(rel)))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := files.CopyFS(b.FS, "templates/"+rel, out); err != nil {
			return err
		}
	}

	if _, err := git.Run(dest, "init", "-q", "-b", "main"); err != nil {
		return err
	}
	if _, err := teamsync.Sync(dest, b, teamsync.Options{}); err != nil {
		return err
	}
	if _, err := git.Run(dest, "add", "-A"); err != nil {
		return err
	}
	if _, err := git.Run(dest, "commit", "-q", "-m", "chore(folkways): init at "+b.Version); err != nil {
		return clierr.New("git commit failed; removed what init created:\n%s", git.Detail(err))
	}
	return nil
}
