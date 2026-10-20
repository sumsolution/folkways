// Package upgrade moves a team repo to another folkways version (`folkways upgrade`). Ported in go-cli
// Phase 1 round 2.
package upgrade

import "errors"

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
	return errors.New("upgrade is not ported to Go yet")
}

// Upgrade installs version on branch chore/folkways-<version>, runs that version's sync, and commits.
func Upgrade(cwd, version string, install Installer) (Result, error) {
	return Result{}, errors.New("upgrade is not ported to Go yet")
}
