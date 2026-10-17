# 0002. Hooks are a Go binary, shipped as per-platform optional dependencies

| | |
|---|---|
| Status | accepted |
| Date | 2026-09-29 |
| Owner | Summer Jordan |

## Issue

Spec §7.1 runs five hooks as `node_modules/.bin/folkways hook <name>`. They fire on every Bash call
(`PreToolUse`) and after every tool batch (`PostToolBatch`), in Claude Code and Copilot CLI, so their start-up
cost is paid constantly. Summer's constraints: the CLI stays TypeScript; hooks must be fast; setup must be
seamless for members on npm or pnpm; no shell scripts; no extra runtime to manage. Which runtime, and which
distribution, so that one committed `.claude/settings.json` works on macOS arm64/x64 and Linux x64/arm64?

## Context

Details and citations: [08-research-hook-runtime.md](https://github.com/sumsolution/folkways-command/blob/main/efforts/hooks-build/08-research-hook-runtime.md). Benchmark:
[bench/hooks/](https://github.com/sumsolution/folkways-command/tree/main/efforts/hooks-build/bench).

- **Speed (Linux x64, 50 runs, p50/p95 ms).** The same git-guard-like hook in each candidate, through each call path:

  | Candidate | direct | npm `.bin` | pnpm `.bin` | Size per platform |
  |---|--:|--:|--:|--:|
  | Node, plain | 65.1 / 81.0 | 65.8 / 75.4 | 72.0 / 81.2 | 0.1 MB + `yaml` |
  | Node, bundle + compile cache | 49.7 / 55.8 | 51.7 / 61.4 | 54.4 / 65.3 | 0.3 MB |
  | Node, snapshot / SEA | 50.1 / 57.3; 46.5 / 53.4 | SEA 49.6 / 56.8 | SEA 52.9 / 59.1 | 6.8 MB / 127 MB |
  | Bun `run` | 81.9 / 97.0 | — | — | needs Bun |
  | Bun `--compile` | 53.5 / 61.1 | 54.8 / 61.0 | 57.0 / 66.7 | 99 MB |
  | Deno `compile` | 64.7 / 75.8 | 65.1 / 77.2 | 71.1 / 81.1 | 122 MB |
  | **Go** | **3.8 / 4.4** | **3.6 / 4.2** | **7.6 / 8.9** | **2.5 MB** |
  | Go behind a JS launcher (Biome pattern) | — | 36.3 / 44.1 | 38.1 / 43.1 | 2.5 MB |
  | Rust | 3.7 / 4.5 | 3.7 / 4.3 | 7.7 / 9.4 | 2.0 MB |

  The proposed command, run through `sh -c` as both tools run it: Go 4.2 (npm) and 8.0 (pnpm) ms. A full Go
  `git-guard` prototype passes all 51 git-guard contract cases at 3.2 ms per commit check.
- **Speed (macOS arm64, Apple M5 Max, 50 runs, p50/p95 ms).** Node is about twice as fast as on the Linux
  box, and Go is still about 4–13× faster (pnpm and npm paths):

  | Candidate | direct | npm `.bin` | pnpm `.bin` | Size per platform |
  |---|--:|--:|--:|--:|
  | Node, plain | 33.6 / 34.9 | 34.5 / 36.1 | 41.3 / 43.7 | 0.1 MB + `yaml` |
  | Node, bundle | 28.4 / 29.4 | 29.3 / 31.9 | 35.8 / 37.4 | 0.3 MB |
  | Node, bundle + compile cache | 27.0 / 28.2 | 27.9 / 29.2 | 34.8 / 36.0 | 0.3 MB |
  | Node, snapshot / SEA | 25.3 / 27.1; 26.1 / 28.0 | SEA 25.9 / 27.2 | SEA 33.7 / 37.4 | 6.9 MB / 121.5 MB |
  | **Go** | **2.6 / 2.6** | **2.2 / 2.4** | **9.3 / 9.7** | **3.0 MB** |
  | Go behind a JS launcher (Biome pattern) | — | 23.7 / 24.8 | 29.6 / 31.0 | 3.0 MB |

  The proposed command through `sh -c`: Go 4.5 (npm) and 11.3 (pnpm) ms. The spawn floor is 1.2 ms. Even the
  best Node path is 27.9 ms, over the 20 ms bar, so Go stays the choice. Rust, Bun and Deno were not measured on
  macOS (not installed). Full table: [bench/hooks/results/](https://github.com/sumsolution/folkways-command/tree/main/efforts/hooks-build/bench/results).
- **Every JS runtime lands at 46–82 ms on Linux** (25–41 ms on macOS), compiled or not; start-up dominates. Node
  optimizations save 13–19 ms of 65 on Linux. Compiled Bun, Deno and Node SEA also cost 100–127 MB per platform.
- **A JS launcher in front of a native binary gives back ~31–33 ms** (esbuild, Biome, Turborepo). esbuild avoids
  that with a postinstall, which pnpm ≥ 10 blocks by default and `--ignore-scripts` skips.
- **npm links the bin of a hoisted transitive dependency; pnpm links only direct ones**, unless the team repo
  sets `publicHoistPattern`. Tested: a main package with platform packages as optional deps gave
  `node_modules/.bin/<bin>` → the native binary with npm (a symlink) and with pnpm plus the pattern (an `exec`
  shim, +4 ms). Both lockfiles kept the non-matching platform package.
- **Hook cwd:** Claude Code runs hooks in the current directory, which follows the Bash tool's `cd`, so a relative
  command stops working. `CLAUDE_PROJECT_DIR` stays at the project root. Copilot CLI runs repo hooks in the
  project root and documents `CLAUDE_PROJECT_DIR` only for plugin hooks.
- **Failure modes differ:** in Claude Code a failing command is a non-blocking error. In Copilot CLI a
  `preToolUse` hook that exits nonzero, crashes or is missing **denies the tool call**.
- **Release:** Go cross-compiles all four targets from one Linux job with no cgo, statically linked, and
  ad-hoc signs darwin/arm64 automatically. Rust needs macOS runners (or a cross linker) for darwin.

## Decision

**Option 4: hooks in Go, one binary `folkways-hook`, shipped as `@folkways/hooks-<os>-<cpu>` optional
dependencies of `@folkways/cli`. The committed command is `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook <name>`.**

- Go gives near-native speed (~2–4 ms direct or through npm, 9 ms through pnpm on macOS; ~13× the best Node option on Linux) in a 2.5–3 MB file, and it is the easiest
  compiled language for TypeScript contributors to read and change. It builds and signs all targets from one
  Linux job. Rust matches the speed but costs more to contribute to and to release. It would win only if
  binary size or the last millisecond mattered.
- Distribution follows esbuild and Biome, minus both of their slow paths: the platform packages carry their
  own `bin`, so there is no JS launcher and no postinstall. npm links the hoisted platform package's bin as a
  symlink. pnpm teams get `publicHoistPattern: ["@folkways/hooks-*"]` in `pnpm-workspace.yaml`, written by
  `init`, so pnpm links it too. The team repo keeps one dependency line, and Dependabot bumps stay one PR.
- The command works on every supported platform. The path contains no platform name: the lockfile holds all four
  platform packages, and each machine installs its own. `CLAUDE_PROJECT_DIR` anchors the command in Claude Code
  after a `cd`. In Copilot CLI the variable is unset for repo hooks, so `.` applies, and that is the project root.
  This needs a shell, so the entry uses shell form, not the exec form (`args`) that the Claude Code docs prefer:
  exec form can't fall back when the variable is unset.
- Accepted after the macOS run: Node's best path is 27.9 ms p50 against Go's 2.2 ms. The rule was Go unless Node
  came in under about 20 ms. Option 1 (Node, lean entry + compile cache) stays the fallback if the release
  matrix proves too costly.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Node: today's `folkways hook <name>`, with a bundled lean entry and compile cache | One language and one package; nothing new to release | 50–55 ms per call on Linux, 28–35 ms on macOS arm64; snapshots and SEA buy at most 3 ms more | The simplest to own, paying 28–55 ms on every hook call forever |
| 2 | Bun (`bun run`) | Fast to write; TypeScript as is | 82 ms here; every member must install and update Bun; no npm pinning | Adds a runtime to every machine and is slower than Node for this job |
| 3 | Bun `--compile` / Deno `compile` / Node SEA | No runtime to install; TypeScript source | 46–71 ms on Linux; 100–127 MB per platform; SEA needs `postject` and re-signing on macOS | Pays compiled-binary costs without compiled-binary speed |
| **4** | **Go binary, per-platform optional deps with their own `bin`, `publicHoistPattern` for pnpm** | ~2–4 ms; 2.5–3 MB; no postinstall or runtime; one committed command; one dependency line | A second language; a 4-package release; pnpm teams need one setting; relies on npm hoisting | Buys a ~13× faster hook with a Go toolchain and a release matrix |
| 5 | Rust, same distribution | Fastest and smallest (3.7 ms, 2.0 MB) | Steepest language for the team; darwin builds need macOS runners; 18 s release builds | Same user experience as Go at a higher contributor and CI cost |
| 6 | Go, team repo lists the platform packages directly | Needs no pnpm setting or reliance on hoisting | Five exact versions to keep in lockstep in every team repo; five Dependabot bumps | More robust linking, clumsier upgrades. The fallback if hoisting breaks |
| 7 | Go behind a JS launcher (Biome/Turbo) | Works with any linker, no settings | 36–38 ms on Linux, 24–30 on macOS, most of the speed gone | Robust, but gives back ~31–33 ms of the ~46 ms gain over Node |
| 8 | Go, with `sync`/`onboard` writing the resolved path into gitignored `.claude/settings.local.json` | Fastest path, no `.bin` | No hooks until each member runs it; nothing fails loudly; per-machine drift | Trades a committed, reviewable config for a hidden per-machine one |

## Consequences

- **Release pipeline:** a new CI workflow cross-compiles `folkways-hook` for darwin-arm64, darwin-x64,
  linux-x64 and linux-arm64 (`CGO_ENABLED=0`). It generates each platform `package.json` (`os`, `cpu`, `bin`,
  version = the `@folkways/hooks` version, per [ADR 0003](0003-repo-layout.md)), publishes the four platform packages first, each with provenance, then
  `@folkways/cli`, whose `optionalDependencies` pin them exactly. A smoke job per OS (macOS arm64, Linux x64
  and arm64) installs the packed tarballs with npm and with pnpm and runs the contract suite
  against `node_modules/.bin/folkways-hook`. Releases that already published cannot be half-undone: the job
  validates everything before the first publish. darwin-x64 is built and shipped but not smoke-tested
  (2026-10-02): Intel Macs end with macOS 26, GitHub's Intel runner image with them, and the beta team is
  unlikely to have one. Its amd64 code is the one linux-x64 tests; only the darwin link is unchecked.
- **Hook command in `settings.json`:** `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook <name>`,
  replacing `node_modules/.bin/folkways hook <name>`. `src/settings.ts` must recognize it (`FRAMEWORK_HOOK_PREFIX` / `isFrameworkHook` today
  matches `folkways hook `). Hooks read `team.yaml` from `CLAUDE_PROJECT_DIR`, else the cwd.
- **`init` / `sync` / `onboard` / `doctor`:** `init` writes `publicHoistPattern: ["@folkways/hooks-*"]` into a pnpm
  team's `pnpm-workspace.yaml`. `sync` warns when that setting is missing or `.bin/folkways-hook` is absent.
  `doctor` runs `folkways-hook --version`, compares it with the manifest, and names the fix
  (`--omit=optional`, a mirror missing the `@folkways/hooks-*` packages, a pnpm setting). `onboard` runs doctor.
- **Hooks must fail open:** they exit 0 on every error. The Go `main` recovers from panics, because a panic exits 2
  and Copilot CLI then denies the tool call.
  A missing binary is loud: Claude Code shows a hook error on every call, and Copilot CLI denies every Bash call.
- **Claude Code plugin** (2026-10-02, [ADR 0003](0003-repo-layout.md) update): one npm package carrying all four
  binaries, no optional dependencies. Its `hooks/run.sh` picks the host's binary and stays quiet in synced
  team repos.
- **CLI:** `folkways hook <name>` in TypeScript goes away, or delegates to the binary for manual runs. The
  contract suite ([packages/hooks/contract/](../../packages/hooks/contract)) stays in TypeScript and tests the Go binary as a black box.
- **Contributor cost:** hook changes need Go 1.26+. Everything else stays TypeScript. The Go code is small (the
  prototype is ~560 lines with a shell-command lexer) and uses two dependencies (`gopkg.in/yaml.v3`, `mvdan.cc/sh/v3` for parsing shell commands). The repo
  gains a Go module (e.g. `hooks/`) and `AGENTS.md` gains its commands.
- **Company mirrors:** members need the mirror to proxy the `@folkways/hooks-*` packages. No known mirror strips
  `os`/`cpu`, but this is untested. The beta team smoke-tests `npm ci` through its mirror before rollout.
- **Still open:** a Copilot CLI run confirming it ignores the nested `hookSpecificOutput` in the combined output
  and runs repo hooks from the project root.
