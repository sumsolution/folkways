# 0003. The repo is a pnpm workspace of independently versioned packages

| | |
|---|---|
| Status | accepted |
| Date | 2026-09-29 |
| Owner | Summer Jordan |

## Issue

The repo root is the `@folkways/cli` package. `test/` holds both the CLI's unit tests and the hook contract
suite, which tests a Go binary. The Go hooks ([ADR 0002](0002-hook-runtime.md)) add a module at `hooks/` and
four published platform packages. Runnable research code sits in `docs/design/07-eval/` and `bench/`. More
deliverables are coming: `review`, `run` and `doctor`, a sync GitHub Action (spec §9.7), and later a
Claude Code plugin for solo projects. Spec §2.1 still says the hook source is `src/hooks/*.ts`.

Each new package makes a later move more expensive, so we decide now. How do we lay out the repo, version
its packages, and build, test and publish them so each package can move at its own pace?

## Context

- **Deliverables today and planned.** `@folkways/cli` (TypeScript, published). `folkways-hook` (Go), published
  as `@folkways/hooks-<os>-<cpu>`, pinned exactly by the CLI ([ADR 0002](0002-hook-runtime.md)). Planned: the
  sync Action (spec §9.7), a Claude Code plugin, and later other harnesses, for solo projects as well as
  teams (see "Future direction").
- **Tests by language.** CLI tests are vitest in `test/`. Go unit tests belong beside the code as
  `*_test.go`. The hook contract ([test/hooks/](../../packages/hooks/contract)) is JSON fixtures plus a TypeScript
  black-box runner that works on any build of the binary. It specifies the hooks, not the CLI.
- **Release coupling.** A CLI release pins one hooks version. The hooks and the CLI change at different
  rates: a lexer fix in `git-guard` shouldn't force a CLI release, and a new skill shouldn't rebuild four
  binaries.
- **Committed pins to unpublished packages break installs.** `pnpm install` fails on a version that isn't on
  npm yet, and `minimumReleaseAge` (7 days) blocks fresh ones. So [11](../design/11-plan-hooks-build.md)
  decision 3 adds `optionalDependencies` at pack time.
- **Required checks and path filters.** A workflow skipped by a workflow-level `paths:` filter never
  reports, so a required check on it stays pending. A job skipped by `if:` reports as skipped, which counts
  as passing.
- **Constraints kept** (`AGENTS.md`): pnpm for this repo; the supply-chain settings in `pnpm-workspace.yaml`;
  CI actions pinned to full SHAs; Angular commits with a module scope; squash PRs.

## Decision

**A pnpm workspace. Every deliverable is a package under `packages/`, owning its source, tests, fixtures and
build. Packages version independently with changesets. CI filters jobs by path, and each package releases
on its own tag.**

```
folkways/
├── package.json            private root: pnpm -r check / test / build; ships nothing
├── pnpm-workspace.yaml     packages: [packages/*], plus the supply-chain settings, unchanged
├── pnpm-lock.yaml
├── go.work                 use ./packages/hooks
├── tsconfig.base.json      shared compilerOptions; each package extends it
├── .changeset/             changesets config and pending changesets
├── packages/
│   ├── cli/                @folkways/cli, published
│   │   ├── src/ test/ assets/ templates/
│   │   ├── package.json    files: dist, assets, templates
│   │   ├── tsconfig.json tsconfig.build.json vitest.config.ts
│   │   ├── CHANGELOG.md
│   │   └── README.md       the npm page; links to the repo README
│   └── hooks/              folkways-hook (Go), private in npm terms, versioned and tagged
│       ├── go.mod          module github.com/sumsolution/folkways/packages/hooks
│       ├── main.go internal/<pkg>/        Go unit tests beside the code
│       ├── contract/       README.md (the contract), fixtures/<hook>/*.json, harness.ts, *.test.ts
│       ├── npm/            template for the four platform package.json files
│       ├── scripts/        build-hooks.ts: cross-compile, generate platform packages
│       ├── package.json    @folkways/hooks: build, test (go vet + go test + contract)
│       └── CHANGELOG.md
├── scripts/                cross-package only: smoke.ts, release helpers
├── research/               evidence behind decisions; not built or tested by CI
│   ├── bench/hooks/
│   └── evals/overlay-enforcement/
├── docs/                   prose only: adr/, design/, guide/, skill-authoring.md, voice.md
└── .github/                ci.yml, release.yml, dependabot.yml
```

Rules:

- **A package owns its tests.** TypeScript tests in `packages/<pkg>/test/`, Go tests as `*_test.go` beside
  the code. Tests that span packages (the packed-tarball smoke) live in root `scripts/`. No shared top-level
  `test/`.
- **The hook contract lives in `packages/hooks/contract/`.** It is the spec of the binary. Its runner stays
  TypeScript and black-box ([ADR 0002](0002-hook-runtime.md)); its vitest config belongs to `packages/hooks`.
  The smoke test imports the harness to run the contract against the installed `.bin/folkways-hook`.
- **Independent versions with changesets.** Each package has its own version, `CHANGELOG.md` and tag
  (`@folkways/cli@0.2.0`, `@folkways/hooks@0.1.0`). `.changeset/config.json` sets
  `privatePackages: { version: true, tag: true }` so `@folkways/hooks` is versioned and tagged even though
  the package itself is never published. A PR that changes a package adds a changeset. `changeset version`
  bumps versions and writes changelogs; `changeset tag` makes the tags.
- **The hooks version and the CLI's pin.** `packages/hooks/package.json` `version` is the hooks version. The
  build embeds it (`-ldflags -X main.version`) and gives it to all four platform packages, which always share
  it. The CLI records the hooks version it pins in `packages/cli/package.json` as
  `"folkways": { "hooks": "<version>" }`. The pack step turns that into exact `optionalDependencies`, so
  nothing unpublished is committed. A hooks release reaches team repos when a CLI release bumps the pin.
- **Hook compatibility.** The CLI writes hook names and events into `settings.json`. Renaming or removing a
  hook, or changing its event, is a major hooks bump and needs a matching CLI change. The contract README
  says so.
- **Root tasks run everything.** `pnpm run check` and `pnpm test` run `pnpm -r`, so they need Go.
  CLI-only work: `pnpm --filter @folkways/cli test`.
- **New deliverables get a package.** For example `packages/action-sync/` for the sync Action, or
  `packages/plugin-claude/` for the plugin.

CI and release:

- **One `ci.yml`, path-filtered jobs.** A `changes` job (`dorny/paths-filter`, pinned SHA) outputs `cli`,
  `hooks` and `shared`. Each package job runs when its filter or `shared` matches. A final `ci` job
  (`if: always()`, needs all) fails if any job failed or was cancelled; it is the only required check.

  | Filter | Paths | Runs |
  |---|---|---|
  | `cli` | `packages/cli/**` | cli check + test, smoke |
  | `hooks` | `packages/hooks/**` | go vet + test, contract, smoke |
  | `shared` | root `package.json`, `pnpm-lock.yaml`, `pnpm-workspace.yaml`, `tsconfig.base.json`, `go.work`, `.github/workflows/**`, `scripts/**` | everything |
  | none | `docs/**`, `research/**`, `*.md` at the root | only `ci` |

  Pushes to `main` run every job.
- **One release per package.** `release.yml` runs on tags `@folkways/cli@*` and `@folkways/hooks@*`, which
  Summer pushes after merging the version PR. The cli job packs `packages/cli` and publishes. It stops
  before publishing if the pinned hooks version isn't on npm. The hooks job cross-compiles, generates and
  validates the four platform packages, then publishes them. Both use npm trusted publishing with
  provenance, a GitHub environment that needs approval, and `id-token: write` on that job only.

## Options considered

**Layout**

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | CLI at the root, `hooks/` beside it | No moves; hooks PR 2 already uses it | The root `package.json` is both the repo and the CLI; hook tests stay in the CLI's `test/`; a third package has no obvious home | Cheapest today, and every new package makes it messier |
| **2** | **pnpm workspace, `packages/*`** | Each package owns its tests and build; room for the Action and plugin; per-package CI and release fit naturally | One restructure PR; paths in docs change; `pnpm publish` runs from `packages/cli` | One disruptive move now for a layout that grows cleanly |
| 3 | Split by language (`ts/`, `go/`) | Toolchains stay apart | The contract (TS) and the binary (Go) it specifies end up in different trees; a package is spread across two roots | Groups files by how they build, not by what they ship |
| 4 | Hooks in a separate repo | Fully independent | Contract changes need two PRs; cross-repo smoke tests; two sets of CI and settings | Buys isolation we can have in one repo at far lower cost |

**Versioning**

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Lockstep: one version for all | One tag, one changelog | Every hooks fix forces a CLI release and the other way round | Simple until the packages' paces diverge, which we expect |
| **2** | **Independent, with changesets** | Per-package versions, changelogs and tags from one tool; handles private packages | A dev dependency; a changeset per PR | A small habit per PR for releases that don't drag each other |
| 3 | Independent, hand-rolled tags | No new tool | Changelogs and bumps by hand; easy to forget one | Saves one dependency and costs discipline on every release |

**CI**

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | No filters | Simple | Every PR runs Go, the contract and smoke, even for a doc typo | Fine at two packages, wasteful as they grow |
| 2 | One workflow per package with `paths:` | Native, no action | Skipped workflows never report, so required checks stay pending | Breaks branch protection |
| **3** | **Job-level filters + one aggregate `ci` check** | Runs only what changed; one required check | A third-party action (pinned); the aggregate job needs care | A little YAML for fast, correct gating |

## Future direction

folkways grew from Summer's own solo setup. Later it should serve solo projects as well as teams, starting
with a Claude Code plugin, and other harnesses after Claude Code and Copilot CLI. We haven't scoped this
yet. For now we record the seams, so today's work leaves them open:

- **Core vs distribution.** The core is the content (skills, instructions, prompts) and the behaviour (the
  hook binary and its contract). The npm CLI syncing into a team repo is one distribution; a plugin is
  another. Each distribution is its own package, built from the same content, not a copy.
- **Content stays in `packages/cli/assets/` for now** (moved 2026-10-03, see the last update). Only `src/sync.ts` and `src/files.ts` may know that
  path. When a second distribution needs the content, it moves to its own package (`packages/skills/` or
  `packages/content/`), which independent versioning already allows.
- **Harness adapters.** Harness specifics stay in named modules: payload shapes in
  `packages/hooks/internal/payload`, settings generation in `packages/cli/src/settings.ts`. A new harness
  gets a new adapter rather than branches through the core. The contract README lists the payload shapes it
  accepts, per harness.
- **Solo config for hooks.** Hooks find their config in a team repo (`CLAUDE_PROJECT_DIR`, `team.yaml`).
  Solo use needs another source, such as the plugin's directory or the user's home. Still open: one config
  model with the team as an optional layer, or two roots. Until then, new hook code notes any new
  team-repo-only assumption.
- **Still open for a plugin:** how to ship a per-platform binary without npm optional dependencies.

**Update (2026-10-02).** The Claude Code plugin is built, as `packages/claude-plugin/` (not
`packages/plugin-claude/`), published as `@folkways/claude-plugin`. The binary question is answered: the
plugin carries all four `folkways-hook` builds and `hooks/run.sh` picks the host's (about 14 MB). The content
did not move: the plugin build reads `packages/cli/assets/` as data, so two packages now know that path.
Moving it to a content package is a follow-up (00-STATUS). The solo-config question stays open.

**Update (2026-10-02, later).** Solo config is decided ([plan 16](../design/16-plan-config-and-solo.md)): one
config model with no second root. `team.yaml` is renamed `folkways.yaml` (no fallback; nothing was
published) and gains `mode: team | solo`. A solo repo is the team layout with `people/<handle>/` removed;
both read `folkways.yaml` at the project root. A repo without it is unconfigured: the hooks and the plugin
stay silent except for a one-time setup offer, recorded in the user's config dir, never the repo (spec §2.3,
§7.1). The CLI stays team-only; solo repos are set up by the plugin's `setup` skill. Like a team repo, a
solo repo is a brain repo that holds no project code (plan 16, decision 6).

## Consequences

- **The restructure** is one PR of `git mv` plus config, planned in
  [12-plan-repo-layout.md](../design/12-plan-repo-layout.md). It landed before hooks PR 2, which rebases onto it.
  `AGENTS.md`, spec §2.1 (which also fixes the stale `src/hooks/*.ts`) and the paths in
  [11](../design/11-plan-hooks-build.md) change in that PR.
- **Hooks plan 11 changes:** decision 3 reads the pin from `folkways.hooks` instead of the root version, and
  PR 5 builds the per-package release above.
- **Contributors:** root `pnpm test` needs Go 1.26+. CLI-only contributors filter to `@folkways/cli`. Every
  PR that changes a package adds a changeset. Commit scopes gain `research`.
- **Branch protection:** the required check becomes `ci` (Summer, in GitHub settings).
- **Dependabot** gains a `gomod` entry for `/packages/hooks`. The npm entry stays at `/` (one workspace
  lockfile).
- **`doctor`** compares `folkways-hook --version` with the CLI's `folkways.hooks` pin.
- **Follow-ups, not in the restructure:** a JSON Schema for `team.yaml` and `.local/me.yaml`, used by the
  CLI to validate and by Go tests as fixtures, so the TypeScript and Go readers can't drift. CLI commands
  move under `src/commands/` when `review`, `run` and `doctor` land.

**Update (2026-10-03).** The content moved to `packages/content/` (`@folkways/content`, private, never
published), so the CLI and the plugin are both consumers of it. The CLI's `prepack` copies the content into
`packages/cli/assets/` (gitignored), so the packed tarball is unchanged. In the workspace `src/package.ts`
resolves the content as `assets/` if present, else the sibling `packages/content`. The follow-up above is
closed.
