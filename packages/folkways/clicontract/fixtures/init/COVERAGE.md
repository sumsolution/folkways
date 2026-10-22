# init fixtures: coverage of the TS tests

Every `it` in `packages/cli/test/init.test.ts` and `package-manager.test.ts` (at `165ec4c`, see
[../../README.md](../../README.md#history)). Fixture names are files in this folder (`<name>.json`).
Skipped cases are Go unit tests.

`init` no longer writes `package.json` or runs a package manager (ADR 0012), so the TS tests of package names,
package managers and the install are gone with their fixtures. Folder names npm refused (`中文`, `!`, `...`)
are now fine.

## init.test.ts

| Test | Fixture or skip |
|---|---|
| templateTarget > 5 table rows | skipped: unit (`scaffold_test.go`) |
| init > scaffolds, pins, syncs and commits | `scaffolds-pins-syncs-commits`: the manifest is the pin; no `package.json` (unit: `TestInitWritesNoNodeFiles`) |
| init > writes folkways.yaml with team mode and every hook on | `folkways-yaml-team-mode-hooks-on` |
| init > names the package after the folder | removed with npm |
| init > creates missing parent folders | `creates-missing-parent-folders` |
| init > refuses a non-empty folder | `refuses-non-empty-folder` |
| packageName (17 cases), init names the package (2) | removed with npm |
| init with pnpm (3 cases) | removed with npm |
| init when the install fails (2 cases) | removed with npm. Cleanup on failure stays pinned by the two cases below |
| init when the first commit fails > removes the scaffold and reports the git error concisely | `first-commit-fails-removes-scaffold` |
| init without a git identity > fails with git's message and removes the scaffold | `no-git-identity-fails-removes-scaffold` |

## package-manager.test.ts

Removed with npm: the Go CLI has no package-manager code.
