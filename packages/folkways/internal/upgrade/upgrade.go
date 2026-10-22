// Package upgrade moves a team repo to another folkways version (`folkways upgrade`). A port of
// packages/cli/src/upgrade.ts.
package upgrade

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
	"github.com/sumsolution/folkways/packages/folkways/internal/git"
	"github.com/sumsolution/folkways/packages/folkways/internal/manifest"
	"github.com/sumsolution/folkways/packages/folkways/internal/pkgmgr"
	"github.com/sumsolution/folkways/packages/folkways/internal/teamsync"
)

// Installer pins version in the team repo, with its package manager in production.
type Installer func(team, version string) error

// Result is what an upgrade did. Branch is "" when the repo was already at To.
type Result struct {
	Branch string
	From   string
	To     string
}

// PackageInstall pins version with the team repo's own package manager.
func PackageInstall(team, version string) error {
	pm, err := pkgmgr.Detect(team)
	if err != nil {
		return err
	}
	return pkgmgr.Exec(team, pkgmgr.AddDevExactCommand(pm, bundle.PackageName+"@"+version), nil)
}

// semver is the semver.org grammar; the ident parts are the numeric and alphanumeric identifiers.
const semver = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*))*)?` +
	`(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`

// Upgrade installs version on branch chore/folkways-<version>, runs that version's sync, and commits. The
// new version's own `sync` runs (node_modules/.bin/folkways), so its assets and rules apply.
func Upgrade(cwd, versionArg string, install Installer) (Result, error) {
	version := strings.TrimPrefix(versionArg, "v")
	if !regexp.MustCompile(semver).MatchString(version) {
		return Result{}, clierr.New("'%s' is not a version; use an exact one, e.g. 1.2.0", versionArg)
	}
	team, err := git.Run(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return Result{}, clierr.New("run this inside the team repo")
	}
	m, err := manifest.Read(team)
	if err != nil {
		return Result{}, err
	}
	if m == nil {
		return Result{}, clierr.New("no .folkways/manifest.json here; is this a folkways team repo?")
	}
	if clean, err := git.IsClean(team); err != nil {
		return Result{}, err
	} else if !clean {
		return Result{}, clierr.New("working tree is dirty; commit or stash first")
	}
	edits, err := teamsync.FindEdits(team, m)
	if err != nil {
		return Result{}, err
	}
	if len(edits) > 0 {
		return Result{}, clierr.New("%s", teamsync.EditsMessage(edits, "Or run 'folkways sync --force' to discard them, commit, and rerun."))
	}
	branch := "chore/folkways-" + version
	if git.BranchExists(team, branch) {
		return Result{}, clierr.New("branch %s already exists", branch)
	}
	start, err := git.Run(team, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return Result{}, err
	}
	if start == "HEAD" {
		return Result{}, clierr.New("detached HEAD; check out a branch first")
	}
	// Same version: nothing to install or commit. After the safety checks, so errors keep their order.
	if version == m.Version {
		return Result{To: version, From: m.Version}, nil
	}
	// The tree was clean before we started, so every untracked file is ours. Ignored files (node_modules) stay.
	abandon := func() error {
		for _, args := range [][]string{{"checkout", "-q", "-f", start}, {"clean", "-q", "-f", "-d"}, {"branch", "-q", "-D", branch}} {
			if _, err := git.Run(team, args...); err != nil {
				return err
			}
		}
		return nil
	}
	fail := func(what string, cause error) (Result, error) {
		if err := abandon(); err != nil {
			return Result{}, err
		}
		return Result{}, clierr.New("%s failed; back on %s:\n%s", what, start, git.Detail(cause))
	}

	if _, err := git.Run(team, "checkout", "-q", "-b", branch); err != nil {
		return Result{}, err
	}
	err = install(team, version)
	if err == nil {
		err = pkgmgr.Exec(team, []string{filepath.Join(team, "node_modules/.bin/folkways"), "sync"}, nil)
	}
	if err != nil {
		return fail("upgrade to "+version, err)
	}

	if _, err := git.Run(team, "add", "-A"); err != nil {
		return Result{}, err
	}
	if _, err := git.Run(team, "diff", "--cached", "--quiet"); err == nil {
		return Result{To: version, From: m.Version}, abandon()
	}
	if _, err := git.Run(team, "commit", "-q", "-m", "chore(folkways): upgrade to "+version, "-m", "From "+m.Version+"."); err != nil {
		return fail("git commit", err)
	}
	return Result{Branch: branch, From: m.Version, To: version}, nil
}
