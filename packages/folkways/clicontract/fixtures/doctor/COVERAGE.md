# Coverage: packages/cli/test/doctor.test.ts

Every `it` (at `165ec4c`, see [../../README.md](../../README.md#history)) → fixture in this folder, or why not.
The npm version checks are gone: the manifest is the pin (ADR 0012).

| `it` | Fixture |
|---|---|
| finds nothing in a freshly synced, installed team repo | `no-problems` (seed; no `package.json`) |
| stops at a repo with no folkways.yaml | `no-folkways-yaml` |
| checks only folkways.yaml and personal skills in a repo without package.json | removed: a repo without `package.json` gets every check |
| reports a missing manifest | `missing-manifest` |
| versions: 6 npm cases (installed, pinned, exact, npx drift, failed upgrade) | removed with npm |
| versions: reports a sync from another version than the pinned one | `version-differs`: a warning that the running binary isn't the manifest's version |
| reports edited framework files, which sync refuses | `edited-framework-file` |
| reports team overlays changed since the last sync | `stale-overlay` |
| warns about personal skills that shadow framework or team skills (M10) | `personal-skill-shadows` |
| disableAllHooks: reports true in .claude/settings.local.json | `disable-all-hooks-local-true` |
| disableAllHooks: reports project settings that no longer set it to false | `disable-all-hooks-project-unset` |
| disableAllHooks: reports unreadable settings | `settings-unreadable` |
| reports a missing hook binary and pnpm hoisting | `folkways-not-on-path`: a warning when `folkways` isn't on `PATH` |
| checks the hooks: values in folkways.yaml | `hooks-yaml-value` |
| HOOK_NAMES are the hooks assets/settings.json wires | skipped: content check, not CLI behavior |
| hookValueFindings: accepts true and false, with comments | `hooks-accepts-true-false` |
| hookValueFindings: flags YAML 1.1 booleans | `hooks-yaml11-booleans` |
| hookValueFindings: flags values that leave the hook on | `hooks-values-that-leave-hook-on` |
| hookValueFindings: flags empty and null values and quoted YAML 1.1 words | `hooks-empty-null-quoted-words` |
| hookValueFindings: reads CRLF files | `hooks-crlf` |
| hookValueFindings: reports tabs | `hooks-tab-indent` |
| hookValueFindings: reports a nested map under a hook | `hooks-nested-map` |
| hookValueFindings: reads a plain value continued on the next line, and does not guess at tags or aliases | `hooks-continued-value-tags-aliases` |
| hookValueFindings: unquotes and trims keys, and accepts an empty or null hooks map | `hooks-key-quoting-null-maps` |
| hookValueFindings: warns about unknown hook names | `hooks-unknown-name` |
| hookValueFindings: flags an inline hooks map it cannot check | `hooks-inline-map` |

Notes:

- The TS tests inject a fake package at version 1.0.0; fixtures use the real CLI at `${VERSION}` and `0.0.1` or
  `9.9.9` for "another version". "Another synced version" uses a minimal manifest (`skills: []`, `files: {}`).
- Stdout ends with the summary line (`N errors, M warnings.`) in every fixture; the TS tests call `doctor()` and
  don't see it, so the CLI contract pins it here.
- `hookValueFindings` is a unit function in TS; here it runs through `doctor` in an initialised repo whose
  `folkways.yaml` the fixture replaces, so it is the only finding.
