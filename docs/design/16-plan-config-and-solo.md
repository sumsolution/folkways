# Plan: `folkways.yaml`, guided setup and solo repos

**Written:** 2026-10-02 (session 8) · **Status:** merged in [#25](https://github.com/sumsolution/folkways/pull/25) (with decision 6 from #27) · **Closes:** open question 3 in
[00-STATUS](00-STATUS.md); the solo-config question in [ADR 0003](../adr/0003-repo-layout.md) ("Future direction")

## Goal

The plugin is installed per user, so it runs in every repo a person opens. Today it runs in full wherever
there's no synced manifest: it injects the team way of working, sends every session to `onboard`, denies
commits that aren't Angular, and writes `.local/precompact.log` into the repo.

Done when:

- a repo with no `folkways.yaml` gets one offer to set up, then silence: no hooks, no instructions, no
  writes into the repo;
- a guided `setup` writes `folkways.yaml` for a solo or a team repo, including commit format and which
  features are off;
- the hooks do only what `folkways.yaml` enables;
- solo repos have a defined layout and the skills work in them.

## Decisions (Summer, 2026-10-02)

1. **No config, no behavior.** Without `folkways.yaml` the plugin offers setup once per repo, then stays
   silent. Same rule for solo and team.
2. **`team.yaml` becomes `folkways.yaml`.** Nobody uses folkways yet, so there's no fallback and no
   migration: `team.yaml` is no longer read.
3. **Config is committed in the repo.**
4. **Solo layout is the team layout with one layer removed:** what lives in `people/<handle>/` in a
   team repo lives at the repo root in a solo repo.
5. **Built with the framework's orchestrator pattern** (instructions, "Subagents"): parallel agents in
   their own worktrees on separate code; the main agent briefs, reviews, merges and asks on judgement calls.
6. **A solo repo is a brain repo** (Summer, 2026-10-02, after review round 2): like a team repo, it's where
   sessions start and context builds up across projects, and it holds no project code. This replaced a
   solo repo being "usually a code repo", and with it `git-guard`'s walk-back over review-only commits,
   the `commit` skill's solo exception and `setup`'s root-collision check. The setup offer in a code repo
   points to the brain repo.

## The shared contract (written first, in spec §2.3, §4.1 and §7.1)

Every brief points here. Agents don't change it; they ask the orchestrator.

### Repo states

| State | Detected by | Plugin | Hooks (`folkways-hook`) |
|---|---|---|---|
| Synced team repo | `.folkways/manifest.json` | exits; the repo's own settings run the hooks | per `folkways.yaml` |
| Configured | `folkways.yaml` at the project root | instructions + hooks | per `folkways.yaml` |
| Unconfigured | neither | no instructions | silent, except `session-start`'s one offer |

### `folkways.yaml` additions (spec §4.1)

```yaml
mode: team                     # team | solo
git:
  commit_convention: angular   # angular | pattern | none
  commit_pattern: ""           # regex on the subject line, when commit_convention: pattern
  scopes: []
hooks:                         # all on by default; false turns one off
  git-guard: true
  context-band: true
  precompact: true
  session-start: true
```

Booleans, not `on`/`off`: YAML 1.1 parsers read those as booleans and YAML 1.2 parsers as strings.

`review.trigger: on-demand` already turns the review gate off; no new key for it.

### The one-time offer (spec §7.1, `session-start`)

In an unconfigured repo, `session-start` prints one line:

> folkways isn't set up in this repo. folkways runs from a brain repo: start sessions there, and code changes happen in worktrees of repos like this one. No brain repo yet? Run the `folkways:setup` skill to create one. This won't be shown again here.

…and appends the repo's top-level path to `${XDG_CONFIG_HOME:-~/.config}/folkways/offered`. A path
already in that file gets nothing. The hook writes outside the repo, never in it. If the file can't be
written, no offer (silence beats repeating).

### Solo layout (spec §2.3)

```
solo-repo/
├── folkways.yaml          # mode: solo
├── knowledge/  state/     # as in a team repo, created on first use
├── overlays/<skill>/      # personal overlays: add only (no compile in the plugin)
├── inbox/                 # was people/<handle>/inbox/
├── work/<KEY>-<slug>/     # was people/<handle>/work/; no efforts/, nothing to promote to
└── .local/                # gitignored: me.yaml (no handle needed), hook state
```

Skills name one **personal root**: `people/<handle>/` in a team repo, the repo root in a solo repo.
Every path that today starts `people/<handle>/` is written relative to it. The review gate's inbox lookup
uses `<personal root>/inbox/`.

## Workstreams

Sizes are estimates; an agent at 2x its estimate stops and reports (instructions, "Step back").

| # | Owner | Touches only | Model | Estimate | Depends on |
|---|---|---|---|---|---|
| W0 | orchestrator | spec, rename across live files | — | ~60 files, mechanical | — |
| W1 | agent | `packages/hooks/**` | cheap | ~300 Go, ~25 fixtures | W0 |
| W2 | agent | `packages/cli/assets/**`, `packages/cli/test/skills.test.ts` | cheap | ~250 lines of skill text | W0 |
| W3 | agent | `packages/cli/src/**`, `templates/**`, other `packages/cli/test/**` | cheap | ~80 | W0 |
| W4 | agent | `packages/claude-plugin/**` | cheap | ~60 | W0 |
| W5 | orchestrator | `docs/**` (not spec), changesets | — | small | W1–W4 merged |

W1–W4 run in parallel, each in its own worktree off the W0 commit. File ownership doesn't overlap, so
merges are clean by construction. Escalate an agent to the strong model on a failing contract it can't
explain or a second review round in its own code.

### W0: contract and rename (serial, before fan-out)

- Spec §2.2 and §4.1 for `folkways.yaml`; new §2.3 solo layout; §7.1 offer and toggles; §8 flows.
- `team.yaml` → `folkways.yaml` across live files: packages, `docs/guide/`, `docs/skill-authoring.md`,
  `README.md`, `AGENTS.md`, spec, ADRs (as an update note). `git mv templates/team.yaml`. Not in
  `research/` or dated design records (01–15): they're history.
- `pnpm run check && pnpm test`, `go test`, contract suite green. One commit: `refactor: rename team.yaml to folkways.yaml`.

### W1: hooks

- Contract first: fixtures for an unconfigured repo (every hook silent; `session-start` offers once, then
  not; offer file unwritable), `commit_convention: pattern | none`, `hooks.<name>: off`, `mode: solo`
  inbox review lookup.
- `internal/team` → reads `folkways.yaml`, reports `Configured`; each hook returns early when its toggle is off.
- **Non-goals:** new hooks; changing deny wording beyond what convention needs; any CLI or plugin change.

### W2: skills and instructions

- `setup` becomes the guided setup: solo or team first; then tracker, docs, commit format, review trigger,
  context bands, which hooks are off. Works in an empty repo (plugin) and after `folkways init`. Writes
  `folkways.yaml`; in solo mode creates nothing else until first use. (Superseded by decision 6: solo
  setup creates the brain repo and also writes `.local/me.yaml` and `.gitignore`.)
- `start`, `onboard`: no `folkways.yaml` → say so and offer `setup`; never assume the team layout.
  `onboard` in solo mode: `me.yaml` without `handle`.
- Every skill path under `people/<handle>/` restated against the personal root (`start`, `status`,
  `handoff`, `side-quest`, `session-close`, `review`, `ticket`, `research`, …). `instructions.md` "Where
  things go" gets a solo column.
- **Non-goals:** new skills; rewording skills beyond these paths; overlay compile for solo.

### W3: CLI

- `init` writes `folkways.yaml` (`mode: team`, the `hooks:` block). `sync` and its warnings read `folkways.yaml`.
- **Non-goals:** `init --solo` (solo goes through the plugin's `setup`); `doctor`.

### W4: plugin

- `run.sh`: no `folkways.yaml` → skip `instructions` and every hook but `session-start`; still exit 0.
- `PREFACE` mentions solo; build tests cover the three repo states with a fake hook binary.
- **Non-goals:** team overlays in the plugin; changes to the hook binary.

### W5: records

- `00-STATUS`: close open question 3, update Built, Next step. ADR 0003: update note closing solo config.
  Guide: solo section. Changesets: `@folkways/hooks` minor, `@folkways/cli` minor (bumps the pinned
  hooks version), `@folkways/claude-plugin` minor.

## Integration (orchestrator)

1. Review each worktree's diff for proportion and correctness against its brief, then merge into
   `ccr-d5e13116-v42x04` in order W1, W3, W2, W4.
2. `pnpm run check && pnpm test`; build the hook binary and run the contract suite against it; build the
   plugin and drive `run.sh` in three scratch repos (synced, configured, unconfigured), checking the
   offer appears once and nothing is written into the unconfigured repo; `pnpm run smoke`.
3. Fresh-process review of the whole branch; fix, re-run step 2.
4. W5, then one PR, squash-merged.

## Resolved before W0

- **Root collisions in solo repos** (Summer, 2026-10-02): solo folders stay at the root. Superseded by
  decision 6: the brain repo is new or empty, so there's nothing to collide with.
