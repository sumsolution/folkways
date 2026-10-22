# Coverage: sync fixtures

Every `it` in `packages/cli/test/{sync,settings,overlays}.test.ts` (at `165ec4c`, see
[../../README.md](../../README.md#history)), mapped to the fixture(s) in this folder or
to the reason it is skipped. `it.each` tables are one row with all their cases. Fixtures run the real
framework content (`packages/content`), not the synthetic `makePackage` content the TS tests use.
The seed `fresh-scaffold` also pins a plain sync.

Fixtures that edit the manifest (a version bump, a dropped entry) do it with a `git -c alias.edit='!sed ...'` step,
because the harness has no edit-in-place and fixtures can't compute sha256 for a hand-written manifest.

## sync.test.ts

| `it` | Fixture or skip |
|---|---|
| sync to a new version > adds new framework files and records them | `adds-missing-framework-files` (adopting variant), `records-previous-version` (the `old -> new` line). Skipped part: files only a newer version ships (one CLI, one content) |
| sync to a new version > keeps team customizations | `keeps-team-customizations` |
| sync to a new version > removes files and hooks the new version dropped | `removes-dropped-framework-file` (file half: the manifest is edited to list an extra file with the sha256 of a literal file the fixture writes). Skipped: the hooks half, since the real settings fragment is fixed and always has hooks |
| sync to a new version > preserves the executable bit | skipped: no real content file is executable |
| sync stops on edits > modified / deleted / added | `refuses-modified-framework-file`, `refuses-deleted-framework-file`, `refuses-added-skill-file` |
| sync stops on edits > --force discards the edits and reports them | `force-discards-edits` |
| refusals > refuses a team skill with a new framework skill name | `refuses-team-skill-with-framework-name` |
| refusals > rejects invalid settings before writing anything | `rejects-invalid-settings` |
| refusals > warns when dependencies are installed without the hook binary | `warns-folkways-not-on-path`: now a warning that `folkways` isn't on `PATH` (ADR 0012) |
| refusals > does not look for the hook binary before dependencies are installed | removed with npm |
| refusals > pnpm team with %s (10 rows) | removed with npm |
| refusals > skips the pnpm check for a package manager it does not support | removed with npm |
| refusals > warns when folkways.yaml is missing | `warns-folkways-yaml-missing` |
| refusals > warns when AGENTS.md lost the import line | `warns-agents-import-missing` |
| refusals > rejects an unsafe manifest before forced deletion (8 rows) | `unsafe-manifest-skill-name`, `-path-traversal`, `-path-in-other-skill`, `-path-folkways-yaml`, `-overlay-traversal`, `-overlay-nested`, `-overlay-bad-checksum`, `-overlays-not-object` |
| refusals > rejects a manifest that is not JSON | `unsafe-manifest-not-json` |
| adopting > keeps files that already match | `adopt-keeps-matching-files` |
| adopting > refuses to overwrite files that differ, and lists them | `adopt-refuses-differing-files` (uses the prompt file: real content has no workflow) |
| adopting > refuses a folder where a file goes | `adopt-refuses-folder-where-file-goes` |
| adopting > refuses a file where the %s folder goes (2 rows) | `adopt-refuses-file-where-folkways-folder-goes`, `adopt-refuses-file-where-claude-folder-goes` |
| adopting > --force overwrites them and reports them | `adopt-force-overwrites-and-reports` |
| adopting, skills in place > keeps the adopted files when the swap fails | skipped: failure injection (`fsOps.rename`) |
| adopting, skills in place > adopts identical skills | `adopt-identical-skills` |
| adopting, skills in place > refuses differing or extra skill files, and --force replaces them | `adopt-differing-or-extra-skill-files` |
| overlays > writes the compiled skill and records both checksums | `overlay-compiles-skill` (`staleOverlays` itself is a unit call; not pinned) |
| overlays > overrides a section in a reference file | `overlay-overrides-reference-section` |
| overlays > stops before writing anything on an unknown section ID | `overlay-unknown-section-id` |
| overlays > recompiles after the overlay changes, without reporting an edit | `overlay-recompiles-after-change` |
| overlays > reports a compiled skill edited by hand like any framework edit | `overlay-compiled-skill-hand-edit` |
| overlays > adopts a repo whose skills are already compiled | `overlay-adopts-compiled-skills` |
| overlays > warns about an overlay for a skill the framework does not ship | `overlay-for-unknown-skill-warns` |
| after an interrupted sync > refuses and says how to recover | `refuses-after-interrupted-sync` |
| with a manifest > refuses to overwrite a team file at a new framework path | `refuses-team-file-at-new-framework-path` |
| is atomic > leaves the old files when copy %i fails (3 rows) | skipped: failure injection |
| is atomic > leaves the old files when rename %i fails (4 rows) | skipped: failure injection |

## settings.test.ts

| `it` | Fixture or skip |
|---|---|
| mergeSettings > adds framework hooks and ask rules to empty settings | `settings-created-when-absent` |
| mergeSettings > replaces framework hooks and keeps team hooks, even in a shared group | `settings-keeps-team-hooks-in-shared-group` |
| mergeSettings > drops the hooks key when nothing is left | skipped: needs an empty fragment; the real fragment always has hooks (unit) |
| mergeSettings > keeps other keys and does not duplicate ask rules | `settings-keeps-other-keys-and-dedupes-ask` |
| mergeSettings > removes entries in the old `folkways hook` form | `settings-removes-legacy-hook-form` |
| mergeSettings > sets disableAllHooks from the fragment, over the team value | `settings-disable-all-hooks-follows-framework` (first half only; "keeps the team value when the fragment has none" needs an empty fragment) |
| mergeSettings > does not mutate its input | skipped: unit (in-process object identity) |
| isFrameworkHook > owns commands that start with the framework prefix, or the old one | `settings-drops-framework-prefix-hooks` (the `FRAMEWORK_HOOK_PREFIX` constant check is a unit assertion) |
| isFrameworkHook > does not own %j (8 rows) | `settings-keeps-lookalike-commands` (all 8 in one fixture, plus the unguarded `folkways hook x` and a guard joined with `&&`) |
| isFrameworkHook > keeps team hooks that mention folkways hook | `settings-keeps-team-hook-mentioning-folkways-hook` |
| unexpected shapes > keeps a hook group without a hooks array untouched | `settings-keeps-group-without-hooks-array` |
| unexpected shapes > keeps non-array events, non-object groups and non-string commands untouched | `settings-keeps-odd-shapes` |
| unexpected shapes > refuses what it cannot merge (5 rows) | `settings-refuses-hooks-not-object`, `-permissions-not-object`, `-ask-not-array`, `-event-not-array` (uses `SessionStart`: the real fragment has no `Stop`), `-not-an-object` |
| assets/settings.json > wires each hook to its events | skipped: content check, not CLI behavior |
| assets/settings.json > owns every entry and keeps hooks on | skipped: content check, not CLI behavior |

The settings fixtures compare the whole merged `.claude/settings.json` exactly, so they break when
`packages/content/settings.json` changes (regenerate them then).

## overlays.test.ts

| `it` | Fixture or skip |
|---|---|
| parseOverlay > splits Add and Override sections | `overlay-compiles-skill` |
| parseOverlay > reports headings that are neither Add nor Override | `overlay-ignores-other-headings` |
| parseOverlay > joins repeated Add sections | `overlay-joins-repeated-add` |
| parseOverlay > refuses the same override ID twice | `overlay-refuses-duplicate-override` |
| parseOverlay > ignores headings inside a fence, matching fence character and length | `overlay-ignores-headings-in-fences` |
| parseOverlay > treats an empty file as an empty overlay | `overlay-empty-file-changes-nothing` |
| compileSkillFiles > returns nothing changed with an empty overlay | `overlay-empty-file-changes-nothing` |
| compileSkillFiles > overrides a section in SKILL.md and appends team additions there | `overlay-compiles-skill` (the "not contain old text" checks are not expressible: no negative match) |
| compileSkillFiles > overrides a section in a reference file | `overlay-overrides-reference-section` |
| compileSkillFiles > refuses an ID no file has, and an ID in two files | `overlay-unknown-section-id` (first half). Skipped: ID in two files, no real skill has one |
| compileSkillFiles > refuses a section ID the skill uses twice in one file | skipped: needs a synthetic skill (no real skill repeats an ID) |
| compileSkillFiles > says when an override targets a section a parent override already replaced | skipped: needs nested IDs; no real skill has them |
| compileSkillFiles > ignores IDs and headings in non-markdown files | skipped: needs a synthetic skill with a script file |
| compileSkillFiles > warns about sections it does not apply | `overlay-ignores-other-headings` |
