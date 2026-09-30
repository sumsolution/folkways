# Plan: build the Go hooks

**Written:** 2026-09-29 (end of session 4) · **Status:** not started · **Decision:** [ADR 0002](../adr/0002-hook-runtime.md)

> **Layout change after PR 2 ([ADR 0003](../adr/0003-repo-layout.md), [12](12-plan-repo-layout.md)).** PRs 3–5 use
> `packages/hooks/` (contract in `packages/hooks/contract/`, CLI in `packages/cli/`). Decision 3 reads the pinned
> hooks version from `folkways.hooks` in `packages/cli/package.json`, not the root version. PR 5 builds a
> release per package, on `@folkways/hooks@*` and `@folkways/cli@*` tags. Paths below are updated for the move; PR 2's branch merged `main` to pick up the move.

## Goal

Ship the five hooks of [spec §7.1](06-v1-spec.md) as one Go binary, `folkways-hook`, delivered as the
per-platform optional dependencies `@folkways/hooks-<os>-<cpu>` of `@folkways/cli`, and wired into team
repos by `sync`. Done when the [contract suite](../../packages/hooks/contract) passes all 119 cases against the binary
installed from packed tarballs, with npm and pnpm, on all four platforms.

## Starting point

| Have | Where |
|---|---|
| Contract: 119 fixtures (git-guard 85, context-band 17, session-start 8, session-end 5, precompact 4), skipped until `FOLKWAYS_HOOK_CMD` is set | [packages/hooks/contract/](../../packages/hooks/contract) (`contract.test.ts:7-11`, `harness.ts`) |
| Go `git-guard` prototype that passes the 51 original git-guard cases: payload parsing (Claude, Pascal, camel; string `tool_input`), shell lexer, deny shape, panic recovery | [research/bench/hooks/go-git-guard/](../../research/bench/hooks/go-git-guard) |
| Distribution design: package names, hoisting, command string, release order, failure handling | [ADR 0002](../adr/0002-hook-runtime.md) Decision and Consequences; [08 §2](08-research-hook-runtime.md) |
| Settings merge that owns hook entries by prefix | `packages/cli/src/settings.ts:7-11` (`FRAMEWORK_HOOK_PREFIX`, `isFrameworkHook`), `mergeSettings` |
| `folkways hook` stub | `packages/cli/src/cli.ts:75` (throws "not built yet") |
| CI: `check`, `test`, `smoke` on ubuntu + macOS; npm-only tarball smoke | `.github/workflows/ci.yml`, `scripts/smoke.ts` |

Missing: the four other hooks, `--version`, inbox review lookup, the review-trigger floor, the Go module
in the repo, settings entries, packaging, a release workflow.

## Decisions to confirm at the start (recommendations)

1. **Go module at `packages/hooks/`** (module path `github.com/sumsolution/folkways/packages/hooks`, root `go.work`), one `main` package plus
   an internal package per hook. Go 1.26+, two dependencies (`gopkg.in/yaml.v3`, `mvdan.cc/sh/v3`). Not shipped in the npm
   package (`files` stays `dist`, `assets`, `templates`).
2. **Contract changes first, in their own PR** (they're spec, not code): the review trigger becomes a
   floor a person can only tighten (§4.2, decided), and `people/<h>/inbox/<date>-review-<sha7>.md` counts
   as a review of HEAD (§7.3, decided). `git.commit_pattern` (§4.1, decided) waits: nothing needs it yet.
3. **`optionalDependencies` are not committed in `package.json`.** The release script adds them at pack
   time. Committing exact pins to unpublished packages would break `pnpm install` in this repo and trip
   `minimumReleaseAge`. The smoke test installs the host's platform tarball next to the CLI tarball.
4. **context-band state** at `.local/hooks/context-band/<session_id>.json` (the contract only says `.local/`).
5. **`folkways hook <name>`** in TypeScript delegates to `node_modules/.bin/folkways-hook` (spec §7.2), so
   people can run a hook by hand. No TypeScript hook logic.
6. **Publishing** uses npm trusted publishing (OIDC from GitHub Actions) with provenance, behind a GitHub
   environment that needs Summer's approval. Summer sets up the `@folkways` packages on npm once.

## PRs, in order

Each PR: TDD, `pnpm run check && pnpm test`, Go `go vet ./... && go test ./...`, secret/path scan of the
diff, squash with a `(#N)` title, Summer approves push and merge.

### 1. `test(hooks)`: contract amendments

- Spec §4.2: `review.trigger` in `me.yaml` may only tighten the team's (order: `on-demand` < `after-commit`
  < `pre-push`). Spec §7.3: inbox fallback.
- `test/hooks/README.md` rules to match. Fixtures: change `push-me-overrides-team-passes` to expect a
  deny (team `pre-push`, me `on-demand`), add a tightening case, add
  `push-with-inbox-review-passes` and a stale-inbox deny case.
- ADR 0001 open item on round 7 closes.

### 2. `feat(hooks)`: Go module and `git-guard`

- Move the prototype to `packages/hooks/`: `main.go` dispatches on the hook name, recovers panics (exit 0), and
  handles `--version` (embedded with `-ldflags -X`). Unknown hook: exit 0, silent.
- Add the trigger floor and inbox lookup (`guard.go` `hasReview`, `config.go` `reviewTrigger`).
- Unit tests in Go for the lexer and config; the contract suite is the acceptance test.
- git-guard model (after three review rounds over-built it): one current directory, and anything that makes it
  unknown skips the push and main-clone checks (contract README). About 570 lines of Go.
- CI: add Go to the existing `hooks` job (`actions/setup-go`, pinned SHA): vet, test, build, then
  `FOLKWAYS_HOOK_CMD=<built binary> FOLKWAYS_HOOKS=git-guard pnpm --filter @folkways/hooks test`.
- `AGENTS.md`: Go commands; `packages/hooks` `test` script runs `go vet`, `go test`, then the contract; `gomod` entry in Dependabot; scope `hooks` already exists.
- Leave `research/bench/hooks/go-git-guard/` as it is (it's evidence for ADR 0002).

### 3. `feat(hooks)`: `session-end`, `session-start`, `precompact`, `context-band`

- One PR or one per hook; all four are small. Behaviour per spec §7.1 and the contract README.
- `context-band`: last main-chain assistant line with non-zero usage; bands from `team.yaml`
  `handoff.bands`; per-session state; re-announce after compaction; silent on Copilot (no transcript).
- CI job drops `FOLKWAYS_HOOKS`: all 119 run.

### 4. `feat(cli)`: wire hooks into team repos

- `packages/cli/assets/settings.json`: one entry per hook and event, command
  `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook <name>` (shell form). Also
  `"disableAllHooks": false` (ADR 0001 M11), carried by `mergeSettings`.
- `packages/cli/src/settings.ts`: new prefix; still recognise the old `node_modules/.bin/folkways hook ` prefix so
  `sync` removes stale entries. Update `packages/cli/test/settings.test.ts` and `packages/cli/test/helpers.ts:56`. Spec §3.1 wording
  (a follow-up) fixed here.
- `init`: for pnpm teams write `publicHoistPattern: ["@folkways/hooks-*"]` to `pnpm-workspace.yaml`.
- `sync`: warn when the pnpm setting or `node_modules/.bin/folkways-hook` is missing (ADR 0002).
- `folkways hook <name>`: delegate (decision 5).

### 5. `build`: packaging, smoke and release

- `scripts/build-hooks.ts`: cross-compile the four targets (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`)
  from one Linux job and generate each `npm/hooks-<os>-<cpu>/package.json` (`name`, `version` = root,
  `os`, `cpu`, `bin: {"folkways-hook": ...}`, `license`, `repository`).
- Pack step adds exact `optionalDependencies` to the CLI's packed `package.json` (decision 3).
- `scripts/smoke.ts`: pack the CLI and the host platform package, install both with npm **and** with pnpm
  (with `publicHoistPattern`), assert `.bin/folkways-hook` exists, run the contract suite through it.
  Relax the "exactly one .tgz" assert.
- CI smoke matrix: macOS arm64 and x64, Linux x64 and arm64 (check current GitHub-hosted runner labels
  for macOS Intel and Linux arm64 at the time).
- `.github/workflows/release.yml` on tag `v*`: build, validate all five packages, publish the four platform
  packages then the CLI, each with provenance; environment approval; pinned SHAs;
  `permissions: id-token: write` for that job only.

### 6. Verify on real machines (before 0.1.0)

- Gatekeeper on an npm-installed darwin binary (ad-hoc signed by Go's linker; untested, 08 §4).
- Copilot CLI: ignores the nested `hookSpecificOutput`, runs repo hooks from the project root (ADR 0002
  "Still open"). Record results in 08 and close the ADR item.
- A lockfile made on macOS, installed on Linux (npm and pnpm).

## Risks

- **pnpm hoisting** depends on `publicHoistPattern`. If it proves fragile, ADR 0002 option 6 (team repo lists
  the platform packages) is the fallback.
- **npm lockfile drops optional platform packages** (npm/cli#8320). The smoke test's cross-OS check catches it.
- **`--omit=optional`** or a mirror without the packages: the hook is missing and every Bash call errors.
  `sync` warns now; `doctor` (next milestone) names the fix.
- **Copilot CLI hook timeouts fail open.** Keep `git-guard` fast (3 ms in the prototype) and never block on network.

## Out of scope

`doctor` (version check of `folkways-hook`, M10, M11 detection), `review` / `run`, `git.commit_pattern`,
publishing 0.1.0 itself (the release workflow is built; the tag is Summer's call).
