# init fixtures: coverage of the TS tests

Every `it` in `packages/cli/test/init.test.ts` and `package-manager.test.ts`. Fixture names are files in this
folder (`<name>.json`). `no-install` and `pnpm` are the seed fixtures. Skipped cases become Go unit tests in the port.

## init.test.ts

| Test | Fixture or skip |
|---|---|
| templateTarget > 5 table rows | skipped: unit (internal function, inputs the CLI can't be given) |
| init > scaffolds, pins, syncs and commits | `scaffolds-pins-syncs-commits` (the TS test's `skills: []` differs: real content has skills) |
| init > writes folkways.yaml with team mode and every hook on | `folkways-yaml-team-mode-hooks-on` |
| init > names the package after the folder | `package-name-my-team` |
| init > creates missing parent folders | `creates-missing-parent-folders` |
| init > refuses a non-empty folder | `refuses-non-empty-folder` |
| packageName > My Team -> my-team | `package-name-my-team` |
| packageName > .hidden -> hidden | `package-name-hidden` |
| packageName > __.team -> team | `package-name-underscores-dot-team` |
| packageName > Team_A.B~c -> team_a.b-c | `package-name-team-a-b-c` |
| packageName > équipe -> equipe | `package-name-equipe` |
| packageName > (x) -> x- | `package-name-parens` |
| packageName > -.-team -> team | `package-name-dash-dot-dash-team` |
| packageName > x*300 -> x*214 | skipped: filesystem limit (300-character folder name exceeds NAME_MAX) |
| packageName > refuses '...' | `package-name-refused-dots` |
| packageName > refuses '._' | `package-name-refused-dot-underscore` |
| packageName > refuses '' | skipped: filesystem limit (empty folder name) |
| packageName > refuses 'node_modules' | `package-name-refused-node-modules` |
| packageName > refuses 'favicon.ico' | `package-name-refused-favicon-ico` |
| packageName > refuses '中文' | `package-name-refused-cjk` |
| packageName > refuses '!' | `package-name-refused-bang` |
| packageName > refuses ' node_modules' | `package-name-refused-leading-space-node-modules` |
| init names the package > after a hidden folder | `package-name-hidden` |
| init names the package > or refuses before creating anything | `package-name-refused-dot-underscore` (checks the folder is absent) |
| init with pnpm > pins pnpm in the packageManager field | `pnpm-pins-package-manager` |
| init with pnpm > leaves packageManager unset for npm | `npm-leaves-package-manager-unset` |
| init with pnpm > hoists the hook binary for pnpm (ADR 0002) and commits the setting | `pnpm-hoists-hook-binary` |
| init when the install fails > removes the folder it created and says why | `install-fails-removes-created-folder` |
| init when the install fails > empties a folder that already existed, and keeps it | `install-fails-empties-existing-folder` |
| init when the first commit fails > removes the scaffold and reports the git error concisely | `first-commit-fails-removes-scaffold` |
| init without a git identity > fails with git's message and removes the scaffold | `no-git-identity-fails-removes-scaffold` |

## package-manager.test.ts

| Test | Fixture or skip |
|---|---|
| detectPackageManager > defaults to npm | `pm-default-npm` (through `init`'s default) |
| detectPackageManager > reads the packageManager field first | covered in fixtures/upgrade (package-manager cases; see its COVERAGE.md) |
| detectPackageManager > falls back to pnpm-lock.yaml | covered in fixtures/upgrade (package-manager cases; see its COVERAGE.md) |
| detectPackageManager > falls back to package-lock.json | covered in fixtures/upgrade (package-manager cases; see its COVERAGE.md) |
| detectPackageManager > rejects an unsupported packageManager | covered in fixtures/upgrade (package-manager cases; see its COVERAGE.md) |
| fromUserAgent > pnpm/12.4.1 ... -> pnpm | `pm-user-agent-pnpm` |
| fromUserAgent > npm/11.19.0 ... -> npm | `pm-user-agent-npm` |
| fromUserAgent > yarn/4.0.0 ... -> undefined | `pm-user-agent-yarn-falls-back-to-npm` |
| fromUserAgent > '' -> undefined | `pm-user-agent-empty-falls-back-to-npm` |
| commands > installs from the lockfile | `pm-install-command`, `pm-install-command-pnpm` |
| commands > adds an exact dev dependency | covered in fixtures/upgrade (package-manager cases; see its COVERAGE.md) |

Extra fixture not from a TS test: `pm-flag-overrides-user-agent` (`--package-manager` wins over the user agent,
from `cli.ts`).
