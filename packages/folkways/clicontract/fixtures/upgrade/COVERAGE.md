# upgrade coverage

Every `it` in `packages/cli/test/upgrade.test.ts` (at `165ec4c`, see [../../README.md](../../README.md#history))
and the fixture that pins it. `upgrade` now syncs the running binary's content and refuses another version
(ADR 0012); it installs nothing.

| TS test | Fixture |
|---|---|
| runs the new version's sync on a branch and commits | `upgrade-commits-on-branch` |
| works from a subfolder | `works-from-subfolder` |
| is a no-op at the same version | `same-version-is-noop` |
| refuses a dirty tree | `dirty-tree` |
| refuses edited framework files before installing | `edited-framework-file` |
| refuses an existing branch | `existing-branch` |
| goes back and cleans up when the install fails | removed with npm |
| removes files a failed sync leaves behind | Go unit test (`TestFailedSyncLeavesNothingBehind`) |
| refuses a detached HEAD | `detached-head` |
| refuses the version %j (7 values) | `invalid-version` (one step per value) |
| accepts v0.2.0 | `upgrade-commits-on-branch` |
| accepts 0.2.0 | `accepts-version-without-v` |
| accepts a prerelease | `prerelease`: parsed, then refused because the binary isn't that version |
| refuses an unsafe manifest | `unsafe-manifest` |
| refuses outside a team repo | `outside-team-repo` |
| upgrade when the commit fails: goes back, cleans up and keeps git's error | `commit-fails` |

The package-manager tests that ran through `upgrade` are removed with npm. A running version other than
`${VERSION}` and the target needs a second build, so `TestOtherVersionNeedsThatBinary` and
`TestUpgradeCommitsOnABranch` are Go unit tests.

## Notes

- The old state sets the manifest `version` to `0.0.1` (a `git -c alias.edit='!sed ...' edit` step, as in the
  sync fixtures), so messages read `0.0.1 -> ${VERSION}` and the commit body `From 0.0.1.`
