// Package scaffold creates a team repo (`folkways init`). A port of packages/cli/src/init.ts.
package scaffold

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
	"github.com/sumsolution/folkways/packages/folkways/internal/files"
	"github.com/sumsolution/folkways/packages/folkways/internal/git"
	"github.com/sumsolution/folkways/packages/folkways/internal/hooks"
	"github.com/sumsolution/folkways/packages/folkways/internal/pkgmgr"
	"github.com/sumsolution/folkways/packages/folkways/internal/teamsync"
)

// Options are init's flags.
type Options struct {
	Install        bool   // install dependencies to create the lockfile
	PackageManager string // npm or pnpm
}

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

// packageName is an npm package name for a team repo in folder: lowercase, URL-safe, no leading `.`, `_` or
// `-`, at most 214 characters (https://docs.npmjs.com/cli/configuring-npm/package-json#name).
func packageName(folder string) (string, error) {
	var b strings.Builder
	for _, r := range norm.NFKD.String(folder) {
		if r >= 0x300 && r <= 0x36f { // combining marks: é -> e
			continue
		}
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		case r > 0xffff:
			b.WriteString("--") // JS replaces each UTF-16 unit
		default:
			b.WriteByte('-')
		}
	}
	name := strings.TrimLeft(b.String(), "-._")
	name = name[:min(len(name), 214)]
	if name == "" || name == "node_modules" || name == "favicon.ico" {
		return "", clierr.New("can't derive a valid npm package name from '%s'; use another folder name", folder)
	}
	return name, nil
}

// Init scaffolds a team repo at path, pins b.Version, syncs and commits. It returns the repo's absolute path.
// On failure it leaves nothing behind, so a rerun works.
func Init(path string, b bundle.Bundle, opts Options) (string, error) {
	dest, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	name, err := packageName(filepath.Base(dest))
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(dest)
	existed := err == nil
	if existed && len(entries) > 0 {
		return "", clierr.New("%s exists and is not empty", dest)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}
	if err := scaffold(dest, name, b, opts); err != nil {
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

func scaffold(dest, name string, b bundle.Bundle, opts Options) error {
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

	pmField, err := pkgmgr.PackageManagerField(opts.PackageManager)
	if err != nil {
		return err
	}
	manifest := struct {
		Name            string            `json:"name"`
		Private         bool              `json:"private"`
		PackageManager  string            `json:"packageManager,omitempty"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{name, true, pmField, map[string]string{bundle.PackageName: b.Version}}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dest, "package.json"), buf.Bytes(), 0o644); err != nil {
		return err
	}
	if opts.PackageManager == "pnpm" {
		if err := os.WriteFile(filepath.Join(dest, "pnpm-workspace.yaml"), []byte(hooks.PnpmWorkspace), 0o644); err != nil {
			return err
		}
	}

	if _, err := git.Run(dest, "init", "-q", "-b", "main"); err != nil {
		return err
	}
	if opts.Install {
		argv := pkgmgr.InstallCommand(opts.PackageManager)
		if err := pkgmgr.Exec(dest, argv, os.Stdout); err != nil {
			return clierr.New("%s failed; removed what init created:\n%v", strings.Join(argv, " "), err)
		}
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
