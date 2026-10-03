# Build folkways with folkways: simpler skills, a command repo, design docs moved

## Context

Summer wants to dogfood folkways on itself from a brain repo, `~/projects/folkways-command`
(`git@github.com:sumsolution/folkways-command.git`, no commits yet), with this monorepo as its project repo.
The skills aren't ready for that: every skill opens with a 16-line "Step 0" that reads `folkways.yaml`,
`.local/me.yaml` and a personal overlay on every run.

Decided in this session:

| Topic | Decision |
|---|---|
| Repo split | Deferred. Command repo + this monorepo. Revisit a content/machinery split after dogfooding |
| Coupling | Reduce the accidental kind: content moves out of the CLI package; one config reader |
| CLI language | Go long term, as its own initiative. **This plan adds no TypeScript, only deletes it** |
| Personal overlays | Removed |
| Config in skills | The `session-start` hook (Go) prints the config once per session. Baking values into skills at compile time waits for the Go port |
| Command repo | Team mode, one member (`summer`), CLI installed from local tarballs |
| `docs/design/` | All of it moves to the command repo, spec included. ADR links point at `https://github.com/sumsolution/folkways-command`; cites in `AGENTS.md` and code comments are removed |
| Tracking | New work runs as separate initiatives (efforts) in the command repo, not one growing `00-STATUS.md` |

Non-goals: splitting the workspace, porting the CLI, config baking, a TypeScript YAML parser, solo-mode CLI
support, publishing to npm.

## How the work runs (the framework's subagent rules)

- The main agent coordinates: briefs, review for proportion and correctness, secret/path scan, merges of
  sub-worktrees. **Summer approves every push and every merge.**
- One worktree per agent off the PR's feature branch, scoped so agents touch separate folders.
- Cheap model first (`sonnet`); escalate to `opus` on conflicting evidence or low confidence. A fresh `opus`
  agent reviews each PR before the push is proposed.
- Each brief carries the goal, the non-goals above, "report lines added and removed", and "push back
  if a fix needs far more code than the problem is worth". Briefs forbid serializing env or local paths.
- TDD: the failing test or fixture first. `pnpm run check && pnpm test` before every commit.
- Step back and show the numbers if a PR outgrows this plan or a review round adds machinery.

## PR 1: `build: move framework content to packages/content`

Mechanical; one `sonnet` agent.

- `git mv packages/cli/assets packages/content` (skills, `instructions.md`, `prompts/`, `settings.json`,
  `workflows/`), as private `@folkways/content`. `packages/cli/templates/` stays (it's the CLI's scaffold).
- CLI: `prepack` copies `packages/content/` into `packages/cli/assets/` (gitignored), so the tarball is
  unchanged. One resolver in `packages/cli/src/package.ts`: `assets/` if present, else `../content`.
  `sync.ts` (`assetMap`, `skillNames`, `mergedSettings`) uses it.
- Plugin: `ASSETS` in `packages/claude-plugin/src/build.ts:13` points at `packages/content`.
- Evals: `packages/evals/src/workspace.ts` archives `packages/content` with `packages/cli` (refs before the
  move keep working through the resolver's fallback order).
- `packages/cli/test/skills.test.ts` moves to `packages/content/test/`.
- CI (`.github/workflows/ci.yml`): a `content` filter; it runs the content tests, cli-test (sync tests read
  real skills), plugin and smoke. `AGENTS.md` layout table and the "plugin is generated" rule; ADR 0003
  gets a dated update closing its follow-up.
- No changeset for CLI or plugin (packed output identical; smoke proves it).

## PR 2: `feat(skills): drop personal overlays and step 0 file reads`

Three agents in parallel, then the coordinator integrates.

**Coordinator first (contract):** spec §3.2, §4.2, §7.1 and `packages/hooks/contract/README.md`; ADR 0001
amendment (personal overlays removed, and why); changesets (`hooks` minor, `cli` minor, `claude-plugin` minor).

**Agent A, hooks (`packages/hooks/`):**
- Fixtures first in `contract/fixtures/session-start/` (team, solo, compact, Copilot shape, missing
  `me.yaml`, oversized file); they must fail on the current binary.
- `internal/sessionstart/sessionstart.go`: in a configured repo, add a config block to the existing
  context: `folkways.yaml` and `.local/me.yaml` verbatim (size-capped), plus two derived lines from
  `internal/team`: `PersonalRoot()` and `ReviewTrigger()`. Reuse `output.Context`; no new parsing.
- Runs on `source=compact` too, which replaces what compaction drops.

**Agent B, content (`packages/content/`, `docs/skill-authoring.md`, `docs/guide/`):**
- Update the mechanical rule in the skills test first, then the 16 skills: "Step 0: Load config and
  overlays" becomes a short `## Config`: the fields the skill uses, and one line: values are in the session
  context from the session-start hook; read `folkways.yaml` / `.local/me.yaml` only if they aren't.
  `start` and `onboard` keep their missing-file handling. `onboard` stops creating `.local/me`.
- `instructions.md`: session-start bullet and the Overlays section (team overlays only).
- Authoring guide skeleton and overlay table; guide pages.

**Agent C, CLI deletions (`packages/cli/src`, `test`, `templates`):**
- `overlays.ts`: remove `lintPersonalOverlay`, `OVERRIDE_WORDING`, `CompileInput.personal`, and
  `compileSkill` (only tests use it; `compileSkillFiles` stays).
- `sync.ts`: remove `lintPersonalOverlays`, `linkPersonalFolder`, `HANDLE`, and the
  `people/<handle>/overlays/` wording in two error messages.
- `doctor.ts`: remove the personal-overlay lint check. Its `hooks:` value check stays until the Go port.
- Templates: `people/README.md`, `overlays/README.md`. Delete the matching tests.

**Known gap, accepted:** solo repos lose overlays until the Go port brings compile to the plugin.
`research/evals/overlay-enforcement/README.md` gets a note that the personal arm is historical.

## Step 3: stand up `folkways-command` (after PR 2 merges)

Coordinator, with Summer; each write and push approved.

1. Build and pack here (`pnpm run build`, then `packCli` / `packHooks` from `scripts/pack.ts`) into
   `folkways-command/vendor/` (gitignored).
2. Scaffold: `init` needs an empty folder, so run `folkways init --no-install --package-manager pnpm` in the
   scratchpad and copy the scaffold in, replacing the placeholder `package.json` and lockfile (ask first).
   Install as `scripts/smoke.ts` does: `@folkways/cli` as `file:vendor/…tgz`, `overrides` for the hook
   packages in `pnpm-workspace.yaml`, plus `publicHoistPattern`. Temporary until 0.1.0 is on npm.
3. Run the `setup` skill: `mode: team`, member `summer`, tracker `github`, `repos: folkways`
   (`main`), `git.scopes` from this repo's `AGENTS.md`, `review.trigger: pre-push`, `adr.repo_path: docs/adr`.
   Then `onboard`: `handle: summer`, `projects_dir: /Users/summer/projects`.
4. `folkways sync`, `folkways doctor`, first commit, push on approval.

## Step 4: move the design record, start initiatives

**In the command repo** (copy from a recorded SHA of this repo; mapping shown to Summer before writing):

| From `docs/design/` | To |
|---|---|
| `01` notes, `02` naming, `06` spec, research `03 04 05 07 08 09 10 14 15` | `knowledge/` |
| Plans `11 12 16`, retro `13`, handoffs `17 18` | closed efforts under `efforts/` |
| `00-STATUS.md` | split: "Built" and "Decided" to `state/`; each open item becomes an effort |

New efforts, each with its own `00-STATUS.md`, created with the `start` skill: **go-cli** (ADR, port,
config baking, compile in the plugin, `review`/`run`), **first-release**, **real-machine-checks**,
**evals-next**, **beta-team-repo**, **repo-split** (parked until after dogfooding).

**PR 3 here, `docs: move the design record to folkways-command`** (one `sonnet` agent):
- Delete `docs/design/`.
- ADRs: `../design/NN-…` links become `https://github.com/sumsolution/folkways-command/blob/main/<new path>`.
- `AGENTS.md`: drop the `docs/design/` row and the "keep 00-STATUS current" rule; reword `spec §` cites.
- Code and config comments: remove `spec §…` and plan cites (16 hits in `packages/`, `scripts/`, `.pnpmfile.mjs`).
- `README.md`, `docs/guide/`, `docs/skill-authoring.md`, `packages/hooks/contract/README.md`, research READMEs
  (41 links): repoint or drop.
- The ADR links resolve only once the command repo is pushed and readable; check them before merge.

## Verification

- Every PR: `pnpm run check && pnpm test`; hooks changes also
  `cd packages/hooks && go vet ./... && go test ./... && go build -o bin/folkways-hook .` and
  `FOLKWAYS_HOOK_CMD="$PWD/packages/hooks/bin/folkways-hook" pnpm --filter @folkways/hooks exec vitest run`;
  plugin: `pnpm --filter "@folkways/claude-plugin..." run build`. Smoke runs in CI.
- PR 1: the packed CLI tarball lists the same `assets/` files as before the move.
- PR 2: new session-start fixtures fail before, pass after; `grep -r "Step 0\|\.local/me/" packages/content`
  is empty; line counts reported (expect a net deletion).
- Step 3: in a new session in `folkways-command`, the session-start context shows the config; run the
  `commit` skill and confirm from the transcript it makes no read of `folkways.yaml` or `me.yaml`;
  `folkways doctor` is clean; `git-guard` refuses a `git switch` in the `folkways` main clone.
- Step 4: no `docs/design` or `spec §` left here (`grep`); every ADR link returns 200.
