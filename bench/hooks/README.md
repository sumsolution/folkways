# Hook runtime benchmark

Evidence for [ADR 0002](../../docs/adr/0002-hook-runtime.md). Results and method:
[08-research-hook-runtime.md](../../docs/design/08-research-hook-runtime.md). These are prototypes, not
production hooks.

## Run it

```
node bench/hooks/run.ts                 # all candidates whose toolchain is installed, 50 runs each
node bench/hooks/run.ts --runs 100      # more runs
node bench/hooks/run.ts --only go,node  # a subset (candidate ids, see the table below)
```

Needs Node 24+ and npm. Optional: pnpm, Go, Rust (cargo), Bun, Deno. A missing toolchain skips its
candidates and says so. Builds go to `bench/hooks/.build/` (gitignored). Results print as Markdown and are
written to `bench/hooks/results/<os>-<arch>-<date>.md` and `.json`. Commit the result files you want to keep.

## The bench hook

Every candidate implements the same program, roughly `git-guard`'s real work:

1. Read the whole of stdin and parse it as JSON (a Claude Code `PreToolUse` payload,
   [fixture/payload.json](fixture/payload.json)).
2. If `tool_name` is `Bash`, match `tool_input.command` against the commit regex below.
3. On a match, read `team.yaml` **from the current directory** and parse it with a real YAML parser
   (the whole file, [fixture/team.yaml](fixture/team.yaml)).
4. Check the first line of the message against the Angular header regex. If it fails, deny with reason
   `Commit message is not Angular "type(scope): subject": "<subject>".`. If `git.scopes` is non-empty and the
   scope is not in it, deny with reason `Scope "<scope>" is not in team.yaml git.scopes (<scopes joined by ", ">).`
5. A deny writes one line of JSON to stdout and exits 0:
   `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"<reason>"}}`.
   Otherwise write nothing and exit 0.

Regexes (same semantics in every language):

| Name | Pattern |
|---|---|
| commit | `\bgit\s+commit\b[^\|;&]*?\s-m\s*(?:"([^"]*)"\|'([^']*)')` |
| header | `^(?:feat\|fix\|docs\|style\|refactor\|perf\|test\|build\|ci\|chore\|revert)(?:\(([a-z0-9._-]+)\))?!?: \S` |

The bench fixture always denies; `run.ts` checks every candidate's stdout against
[fixture/expected.json](fixture/expected.json) before timing it.

## Candidates

| Id | What runs | Source |
|---|---|---|
| `node` | `node hook.mjs`, `yaml` from `node_modules` (TS stripped with `module.stripTypeScriptTypes`) | [js/hook.ts](js/hook.ts) |
| `node-bundle` | `node hook.bundle.mjs`, one file (esbuild) | same |
| `node-cache` | lean entry that calls `module.enableCompileCache()` then imports the bundle | same |
| `node-snapshot` | `node --snapshot-blob hook.blob` (V8 startup snapshot of the bundle) | same |
| `node-sea` | Node single executable application of the bundle, with code cache | same |
| `bun` | `bun hook.ts` | same |
| `bun-compile` | `bun build --compile` binary | same |
| `deno-compile` | `deno compile` binary | same |
| `go` | `go build -ldflags="-s -w"` binary | [go/](go) |
| `rust` | `cargo build --release` binary | [rust/](rust) |

## Call paths

| Path | What it measures |
|---|---|
| `direct` | the executable (or `node <file>`) spawned directly |
| `npm-bin` | `node_modules/.bin/<name>` after `npm install` of a local tarball: a symlink on POSIX |
| `pnpm-bin` | the same after `pnpm install`: a generated `/bin/sh` shim |
| `*-shim` | esbuild/Biome pattern: `.bin` points at a JS launcher in the main package that resolves the platform binary and spawns it |
| `sh -c` | the hook runners start commands through a shell; this row adds that cost for the proposed design |

Timing is wall clock around `spawnSync`, one fresh process per run, candidates interleaved. `floor` is the
cost of spawning `true`; subtract it to compare with other machines. "Cold" here means a fresh process with a
warm OS file cache; the first run after a build is reported separately.
