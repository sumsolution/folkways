# upgrade coverage

Every `it` in `packages/cli/test/upgrade.test.ts` and the fixture that pins it.

| TS test | Fixture |
|---|---|
| runs the new version's sync on a branch and commits | `upgrade-commits-on-branch` |
| works from a subfolder | `works-from-subfolder` |
| is a no-op at the same version | `same-version-is-noop` |
| refuses a dirty tree | `dirty-tree` |
| refuses edited framework files before installing | `edited-framework-file` |
| refuses an existing branch | `existing-branch` |
| goes back and cleans up when the install fails | `install-fails` |
| removes files a failed sync leaves behind | skipped: needs a stub install override (`node_modules/.bin/folkways` replaced after the stub writes it); Go unit test |
| refuses a detached HEAD | `detached-head` |
| refuses the version %j (7 values) | `invalid-version` (one step per value) |
| accepts v0.2.0 | `upgrade-commits-on-branch` |
| accepts 0.2.0 | `accepts-version-without-v` |
| accepts a prerelease | `prerelease` |
| refuses an unsafe manifest | `unsafe-manifest` |
| refuses outside a team repo | `outside-team-repo` |
| upgrade when the commit fails: goes back, cleans up and keeps git's error | `commit-fails` |

## package-manager.test.ts (through `upgrade`)

| TS test | Fixture |
|---|---|
| detectPackageManager: defaults to npm | `upgrade-commits-on-branch` (no lockfile, `npm install ...`) |
| detectPackageManager: reads the packageManager field first | `package-manager-field-first` |
| detectPackageManager: falls back to pnpm-lock.yaml | `package-manager-pnpm-lockfile` |
| detectPackageManager: falls back to package-lock.json | `package-manager-npm-lockfile` |
| detectPackageManager: rejects an unsupported packageManager | `package-manager-unsupported` |
| commands: adds an exact dev dependency | `package-manager-field-first` (`pnpm add ...`), `package-manager-npm-lockfile` (`npm install ...`) |
| fromUserAgent (4 cases), commands: installs from the lockfile | not reachable through `upgrade` (user agent is `init` only; plain install is `init`); see the init fixtures |

## Notes

- The old state sets the `package.json` pin and the manifest `version` to `0.0.1` (a `git -c alias.edit='!sed ...' edit` step, as in the sync fixtures), so messages read `0.0.1 -> ${VERSION}` and the commit body `From 0.0.1.`
- `prerelease` and `install-fails` make the install fail with `STUB_NPM_FAIL`; the stub would accept any version.
- `same-version-is-noop` starts from `init` with install, so the only npm call is init's own `npm install`: the upgrade makes none. Go returns before any check, so the TS lockfile caveat (the stub's `package-lock.json` making the upgrade a change) no longer applies.
