# 0011. The CLI is one Go binary, installed without Node first

| | |
|---|---|
| Status | accepted (Summer Jordan, 2026-10-18); partly superseded by [0012](0012-cli-only.md): no plugin, npm optional, no copy precedence |
| Date | 2026-10-18 |
| Owner | Summer Jordan |
| Reviewers | none required (`adr.reviewers.min: 0`) |
| Review window | until 2026-10-25 |
| Supersedes | in part, [0002](0002-hook-runtime.md): "the CLI stays TypeScript" and the `folkways-hook` command |

## Issue

folkways is two programs in two languages:
- the CLI, TypeScript, 1,403 lines in 14 files;
- `folkways-hook`, Go, 1,195 lines (2,197 with tests).

The CLI needs Node and comes only from npm. The TS CLI has been frozen to deletions since 2026-10-03.

Two things now push on that:
- [architecture-review 06](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/architecture-review/06-adr-deterministic-steps.md) moves about 10
  deterministic skill steps into CLI subcommands. Skills will call `folkways ...` on every surface,
  including the cloud agent, where the only binary is probably the plugin's.
- Summer wants folkways usable without Node, especially for personal use.

What should the CLI be written in, how many binaries should folkways ship, and how do people install
and pin it?

## Context

- Code at folkways `fd86000`:
  - CLI commands: `init`, `sync`, `upgrade`, `hook`, `doctor`, `version`; `review` and `run` are stubs
    ([cli.ts](https://github.com/sumsolution/folkways/blob/fd86000/packages/cli/src/cli.ts)).
  - The hooks module is
    [packages/hooks](https://github.com/sumsolution/folkways/tree/fd86000/packages/hooks). It holds
    the `team` config loader, `payload` and `output`, and has a black-box contract suite of 161
    fixtures.
  - Sizes: [folkways-command go-cli 01 §What exists](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/go-cli/01-plan.md#what-exists).
- ADR 0002 measured hook start-up:
  - Go: 2–4 ms. Every JS runtime: 25–82 ms.
  - Compiled Bun, Deno or Node SEA: 100–127 MB per platform.
  - A JS launcher in front of a native binary adds ~31–33 ms.
- ADR 0002 also found that `npx` and pnpm don't run a transitive package's bin without help (a JS
  launcher, or `publicHoistPattern`).
- Copilot CLI denies the tool call when a `preToolUse` hook exits nonzero, crashes, or is missing (ADR
  0002; [app-repo-first 23](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/app-repo-first/23-probe-copilot-github.md) Q5).
- [ADR 0007](https://github.com/sumsolution/folkways/pull/41) (proposed): hooks come from the plugin's
  bundled binary through `run.sh`. That path already needs no Node.
- Nothing is published yet. [first-release](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/first-release/00-STATUS.md) waits for this port (Summer,
  2026-10-18), so command names and the settings hook command can change at no cost.
- Why the port was chosen on 2026-10-03 isn't recorded. The reasons in this ADR are the current ones
  **(unconfirmed)**.

## Decision

Option 3. We will:
- **Port the CLI to Go.** It joins the hooks module as one `folkways` binary, and hooks run as
  `folkways hook <name>`.
- **Embed the skills, templates and instructions in the binary** with `go:embed`.
- **Make the Node-free install the first-class channel:** GitHub release binaries, a Homebrew tap and an
  install script, built by GoReleaser. npm stays a supported second channel for Node projects.

The port reaches parity first: the TS tests become contract fixtures, and the Go code must pass them.
Then the TS CLI is deleted. Plan: [01](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/go-cli/01-plan.md).

Why one language:
- 06 moves real logic into the CLI, and that logic has to be callable wherever skills run.
- A TS CLI can't run in the plugin without Node, and the TS runtimes that compile to one file cost
  100+ MB per platform.
- Go is already the hook language, its contract-test pattern already works here, and one language
  halves the toolchains a contributor needs.

Why one binary rather than two:
- Skills reach the CLI wherever the hook runs, including the cloud agent through the plugin.
- Four release artifacts, not eight.
- No version skew between hook and CLI.
- Moving from one to two later is a build change (a hook-only build from the same source, using build
  tags). Moving from two to one later would change the hook command in every synced repo.

Three guards on the hook path, because Go runs every linked package's `init` before `main`, and a panic
there escapes `recover` and fails closed in Copilot:
1. No work at package level, enforced by a test.
2. The hook contract suite runs against the release binary.
3. A CI budget for hook p50, under 10 ms.

Neither hook p50 nor the size of the merged binary has been measured. Phase 0 measures both before
the port starts, and the CI budget keeps testing them after.

npm keeps ADR 0002's per-platform packages. `@folkways/cli` gets a small JS launcher so `npx` works,
which costs ~30 ms on CLI commands only, never on hooks. Two details are left to the plan
([01](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/go-cli/01-plan.md)): how a Node-free team repo pins its version, and which copy runs when the plugin's
copy and one on `PATH` both exist.

The contract runner moves to `go test`, for the hooks suite and the new CLI suite. Then the only Node
left in developing folkways is npm packaging.

Non-goals:
- New behavior during the port. Features come after the cutover, including 06's subcommands, `review`,
  `run` and config baking.
- Moving the plugin or ADR 0007's delivery.
- Windows. It isn't a supported platform today.

Size:
- About 1,400 Go lines for the port, plus fixtures (inferred: Go is close to TS line for line in
  file and git code).
- About 440 lines of release scripts replaced by GoReleaser config.
- About 7 stacked PRs: this ADR, skeleton and harness, fixtures, four port PRs, release, deleting TS.
  Some may combine.

## Options considered

The binary trade-offs are in detail in [01 §Phase 0](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/go-cli/01-plan.md#phase-0-foundation-main-agent-serial).

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Keep the TS CLI; add 06's subcommands in TS | No port; the existing tests stay; TS is the larger contributor pool | Needs Node everywhere a skill calls it, including the cloud agent, where it may not exist; two languages forever; npm-only install | Cheapest now, but doesn't meet the Node-free goal or reliably reach skills on every surface |
| 1b | TS compiled to one file (Bun `--compile`, Node SEA) | No Node at install; keeps TS | 100–127 MB per platform (ADR 0002); plugin carries all four; the hook still has to be Go for speed, so still two languages | Node-free in name, at 40× the size, and it doesn't remove Go |
| 2 | Go, two binaries (`folkways`, `folkways-hook`) from one module | The hook surface stays small and separately auditable; smaller plugin if it ships only the hook | 8 artifacts; hook and CLI can drift in version; the plugin must ship both for skills to reach the CLI, which cancels the size win | Buys isolation that a test and a budget can mostly provide |
| **3** | **Go, one binary; Node-free first, npm second** | One language, one artifact per platform; skills reach it wherever hooks run; works without Node; reversible to option 2 by build tags | A larger hook binary, not measured; every CLI package's `init` runs on every hook call; a second release channel (Homebrew tap, install script) to maintain | One small risk on the hook path, guarded, for one program everywhere |
| 4 | Go, one binary, npm only (ADR 0002's channel as is) | One release channel; Dependabot bumps the pin | Personal and non-Node users have to install Node to get a Go binary; `npx` needs the launcher anyway | Simpler release, but misses the stated goal |

## Consequences

- **Easier:**
  - One language, toolchain and test pattern.
  - 06's subcommands can be called from skills on every surface where the plugin runs.
  - Solo and non-Node users install with Homebrew or a script.
  - Content can't drift from the binary, because it's embedded.
- **Harder or costlier:**
  - Contributors need Go for all CLI work, not only hooks.
  - Two release channels.
  - first-release grows: a Homebrew tap repo and a GoReleaser token, as well as the npm trusted
    publishers.
  - Dependabot can't bump a `folkways.yaml` pin, so upgrades go through `folkways upgrade`, or a Renovate
    regex manager for teams that want PRs.
  - The hook binary grows, and the three guards become permanent CI work.
  - Two copies of the binary on one machine (plugin and `PATH`) need a precedence rule (plan).
- **Follow-on work:**
  - Phases 0–3 in [01](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/go-cli/01-plan.md).
  - first-release steps updated for Homebrew and GoReleaser.
  - ADR 0007's `run.sh` and the settings template switch to `folkways hook`.
  - 06's subcommands, one ticket each, after the cutover.
- **Revisit if:**
  - hook p50 goes over the budget, or the binary passes ~15 MB: split out a hook-only build (option 2);
  - the cloud agent can't run the plugin's binary (06 Open 1): skills then can't rely on the CLI there,
    whatever the language;
  - Windows support becomes a goal.

## Related decisions

- [ADR 0002](0002-hook-runtime.md):
  superseded in part. The Go runtime and per-platform npm packages stand. "The CLI stays TypeScript"
  and the `folkways-hook` command do not.
- [ADR 0003](0003-repo-layout.md): the
  module move changes the package layout.
- ADR 0007 ([folkways#41](https://github.com/sumsolution/folkways/pull/41)): constrains this ADR. Hooks
  come from the plugin, which now carries the whole CLI.
- [architecture-review 06](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/architecture-review/06-adr-deterministic-steps.md): depends on this ADR.
- ADR 0010 ([architecture-review 04](https://github.com/sumsolution/folkways-command/blob/46c9211/efforts/architecture-review/04-adr-architecture.md)): its change to
  `sessionstart.go` has to be sequenced with the module move.

## Interested parties

- The beta team, once onboarded: the install changes, and npm becomes optional.
