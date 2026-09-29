# Hook runtime: tool contracts, distribution, benchmark

**Researched:** 2026-09-29 (operator session; Sonnet research agents, Opus re-check of the tool docs against raw
page text, empirical runs in Claude Code 2.1.284). Feeds [ADR 0002](../adr/0002-hook-runtime.md).
**Confidence:** high / medium / low per finding. *Inferred* marks an inference.

Sources, pinned:

| Short | Source |
|---|---|
| CC | Claude Code hooks reference, raw markdown `https://code.claude.com/docs/en/hooks.md` (read 2026-09-29), cited as [hooks#anchor](https://code.claude.com/docs/en/hooks) |
| GHR | Copilot hooks reference, [github/docs@f4e8afc `content/copilot/reference/hooks-reference.md`](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md) |
| CL | Copilot CLI changelog, [github/copilot-cli@8dfa600 `changelog.md`](https://github.com/github/copilot-cli/blob/8dfa6009c4a04b3a22a5ca4a7c36a056edd718dd/changelog.md) (latest entry 1.0.89) |
| EXP | Our runs: `claude -p` 2.1.284, Haiku, a Node capture hook on every event in a scratch repo. Commands and payloads are summarized below; rerun with any capture script that logs stdin, `process.cwd()` and `CLAUDE_PROJECT_DIR` |
| BENCH | [bench/hooks/](../../bench/hooks), results in [bench/hooks/results/](../../bench/hooks/results) |

## 1. Spec §9.3 answers

| # | Question | Answer | Evidence | Conf |
|---|---|---|---|---|
| 1 | `PostToolBatch` exists in current Claude Code? | Yes. It fires once per batch of tool calls, after all have resolved, with no matcher. Input: common fields plus `tool_calls[]` of `{tool_name, tool_input, tool_use_id, tool_response}`, where `tool_response` is the serialized tool result: a string, or a content-block array (EXP: a string for Bash). It accepts `hookSpecificOutput.additionalContext`; `decision: "block"` or exit 2 stops the loop | [CC #posttoolbatch](https://code.claude.com/docs/en/hooks#posttoolbatch); EXP: fired after PostToolUse, also when the tool was denied | high |
| 2 | Event order, one Bash call | SessionStart(`startup`) → UserPromptSubmit → PreToolUse → PostToolUse → PostToolBatch → Stop → SessionEnd(`other` under `-p`) | EXP | high |
| 3 | Transcript usage matches `/context`? | Yes. Last main-chain assistant line: 8 input + 28,934 cache read + 179 cache creation = **29,121**. `/context` said "29.1k / 200k (15%)". SessionStart on resume reported `context_tokens: 29155`, which is that sum plus the line's 34 output tokens (+0.12%). The statusline documents the same formula ("input tokens only") | EXP; [statusline #context-window-fields](https://code.claude.com/docs/en/statusline) | high (one session) |
| 3a | Pitfalls for the formula | Each message can appear on several transcript lines (one per content block) with the same usage. Headless `--output-format json` `usage` is **cumulative** (≈2× here); don't use it. The statusline did not run under `-p`, so it wasn't compared directly | EXP | high |
| 4 | Copilot CLI reads hooks from `.claude/settings.json`? | Yes, plus `.claude/settings.local.json`, since CLI 1.0.12. PascalCase event names get snake_case payloads with Claude-style `tool_name: "Bash"`; camelCase names get camelCase payloads | [GHR L38](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L38), [L261-L264](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L261-L264), [L433-L449](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L433-L449); [CL L2005](https://github.com/github/copilot-cli/blob/8dfa6009c4a04b3a22a5ca4a7c36a056edd718dd/changelog.md#L2005) | high (docs); **unverified** at runtime (no Copilot CLI here) |
| 4a | Copilot `SessionEnd` | Supported in both spellings. `reason` is `complete\|error\|abort\|timeout\|user_exit` (no Claude values). Output is ignored. In `-p` it fires once per turn | [GHR L305-L331](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L305-L331); [CL L384](https://github.com/github/copilot-cli/blob/8dfa6009c4a04b3a22a5ca4a7c36a056edd718dd/changelog.md#L384) | high (docs) |
| 4b | Copilot `PreToolUse` | Supported with Claude matcher semantics. `tool_input` is "parsed from JSON string when possible", so it may arrive as a string. camelCase `toolArgs` is a JSON string | [GHR L408-L431](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L408-L431) | high (docs) |
| 4c | Copilot payload gaps | No `PostToolBatch`. `SessionStart.source` is `startup\|resume\|new` (no `compact`). No transcript path on tool, prompt or session events (only Stop, subagent events, PreCompact). No token counts anywhere | [GHR L242-L257](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L242-L257), [L275](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L275), [L656-L683](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L656-L683) | high (docs) |
| 5 | Hook working directory, Claude Code | The **current** directory, which **follows the Bash tool's `cd`**. After `cd sub`, the next hooks ran with `process.cwd()` and payload `cwd` = `sub`, and a relative hook command **stopped running**. `CLAUDE_PROJECT_DIR` stayed at the launch dir and is exported to every hook. The docs recommend it for script paths | EXP; [CC #reference-scripts-by-path](https://code.claude.com/docs/en/hooks#reference-scripts-by-path) | high |
| 5a | Hook working directory, Copilot CLI | The project root, unless the hook sets `cwd` ([GHR L139](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L139)); fixed in 1.0.88 after a regression to the session cwd ([CL L52](https://github.com/github/copilot-cli/blob/8dfa6009c4a04b3a22a5ca4a7c36a056edd718dd/changelog.md#L52)). `CLAUDE_PROJECT_DIR` is documented only for **plugin** hooks ([CL L1987](https://github.com/github/copilot-cli/blob/8dfa6009c4a04b3a22a5ca4a7c36a056edd718dd/changelog.md#L1987)); the reference doesn't mention it | see answer | high (docs); `CLAUDE_PROJECT_DIR` for repo hooks **unverified** |
| 5b | Shell | Claude Code runs shell-form commands with `sh -c` on macOS and Linux (the `shell` field's documented default says `bash`, an inconsistency). EXP: `$0` was `/bin/sh`, and `$VAR`, `&&` and redirects worked. Copilot runs the `bash` field, with `command` as the fallback that Claude-format entries use | [CC #exec-form-and-shell-form](https://code.claude.com/docs/en/hooks#exec-form-and-shell-form); EXP; [GHR L134-L145](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L134-L145) | high |
| 5c | `settings.local.json` hooks | Claude Code merges hooks across settings files, and an identical handler in two files runs once. EXP: both files' hooks ran. Copilot runs all entries from all sources, with no dedup documented | [CC #hook-locations](https://code.claude.com/docs/en/hooks#hook-locations), [#hook-handler-fields](https://code.claude.com/docs/en/hooks#hook-handler-fields); EXP; [GHR L33](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L33) | high |
| 5d | Launch from a subdirectory | Claude Code took the subdirectory as the project. The parent's `.claude/settings.json` was **not** loaded, but the parent's `settings.local.json` was (seen 3 times, unexplained) | EXP | medium |
| 6 | Output JSON, Claude Code | `hookSpecificOutput.{hookEventName, permissionDecision: allow\|deny\|ask\|defer, permissionDecisionReason, additionalContext}`. JSON is read on every exit code; with valid JSON the code is ignored. Exit 2 blocks blockable events and sends stderr to Claude. Other nonzero codes are non-blocking. Fields are capped at 10,000 chars. EXP: a deny blocked the call and the model quoted the reason; exit 2 blocked and showed stderr; exit 1 did not block; `additionalContext` reached the model on UserPromptSubmit, SessionStart, PreToolUse, PostToolUse and PostToolBatch | [CC #pretooluse-decision-control](https://code.claude.com/docs/en/hooks#pretooluse-decision-control), [#exit-code-2-behavior-per-event](https://code.claude.com/docs/en/hooks#exit-code-2-behavior-per-event), [#other-exit-codes](https://code.claude.com/docs/en/hooks#other-exit-codes); EXP | high |
| 6a | Output JSON, Copilot CLI | preToolUse reads **top-level** `permissionDecision` (`allow\|deny\|ask`) and `permissionDecisionReason` (required for deny); no `hookSpecificOutput` is documented. SessionStart reads top-level `additionalContext`. UserPromptSubmit output: the reference says it's dropped for config-file hooks, but changelog 1.0.65 says `additionalContext` is added (conflict). PreCompact output is ignored and can't block | [GHR L685-L693](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L685-L693), [L293-L303](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L293-L303), [L369](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L369); [CL L911](https://github.com/github/copilot-cli/blob/8dfa6009c4a04b3a22a5ca4a7c36a056edd718dd/changelog.md#L911) | high (docs); UPS medium |
| 6b | Exit codes, Copilot CLI | Exit 2 is a warning, except on preToolUse, where it denies. **preToolUse fails closed**: a crash or any nonzero exit denies the call. Timeouts (30 s default) fail open everywhere | [GHR L846-L855](https://github.com/github/docs/blob/f4e8afc6979acd8de6b5da035066281b9a5f4025/content/copilot/reference/hooks-reference.md#L846-L855) | high (docs) |
| 6c | One output for both tools | Claude Code accepted a deny JSON carrying **both** the nested and the top-level fields: it blocked the call and the model quoted the reason. Whether Copilot ignores the nested object is undocumented | EXP (2.1.284) | high for Claude Code; **unverified** for Copilot |
| 7 | `PreCompact`, SessionStart `compact` | Headless `claude -p "/compact" --resume <id>` fired SessionStart(`resume`, with `context_tokens`) → PreCompact(`manual`) → SessionStart(`compact`) → SessionEnd | EXP; [CC #precompact](https://code.claude.com/docs/en/hooks#precompact) | high |
| 8 | Copilot CLI subagent `model:` | Not researched in this workstream | — | open |

**Consequence for spec §7.1** *(inferred from 5, 5a, 5b)*: a relative hook command breaks in Claude Code after any
`cd`. `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/…` works in both: Claude Code sets the variable, and Copilot CLI
falls back to `.`, its project root. Hooks must also find the team repo from `CLAUDE_PROJECT_DIR`, not the cwd.

## 2. Distribution

### 2.1 How others ship native binaries through npm

| Finding | Evidence | Conf |
|---|---|---|
| esbuild, Biome, Turborepo, rollup and Claude Code all use one main package plus one package per platform (`os`/`cpu`, Biome adds `libc`), listed as exact-version `optionalDependencies` | [esbuild package.json L20-L45](https://github.com/evanw/esbuild/blob/f6058f8364fe7ab91ca57a83e02577ed74c9cae4/npm/esbuild/package.json#L20-L45); [Biome package.json L48-L57](https://github.com/biomejs/biome/blob/f89dd4f2d52e4919133429fa4550912510517a9a/packages/@biomejs/biome/package.json#L48-L57); [registry: @anthropic-ai/claude-code](https://registry.npmjs.org/@anthropic-ai/claude-code/latest) | high |
| Biome and Turborepo point `bin` at a JS launcher that resolves the platform package and spawns it, so every call pays for a Node start | [Biome bin/biome L45-L69](https://github.com/biomejs/biome/blob/f89dd4f2d52e4919133429fa4550912510517a9a/packages/@biomejs/biome/bin/biome#L45-L69); [turbo bin/turbo L427-L502](https://github.com/vercel/turborepo/blob/ac801403d9d58fc13331485de7b11b992c2da5d9/packages/turbo/bin/turbo#L427-L502) | high |
| esbuild's postinstall hard-links the binary over `bin/esbuild`, on the grounds that running node again "is way slower". It is skipped under `--ignore-scripts` and Yarn, which leaves the JS launcher. Its fallbacks run a nested `npm install`, then download from a hard-coded `registry.npmjs.org`, which bypasses a company mirror | [node-install.ts L178-L226](https://github.com/evanw/esbuild/blob/f6058f8364fe7ab91ca57a83e02577ed74c9cae4/lib/npm/node-install.ts#L178-L226), [L236-L250](https://github.com/evanw/esbuild/blob/f6058f8364fe7ab91ca57a83e02577ed74c9cae4/lib/npm/node-install.ts#L236-L250); [jfrog-cli#1304](https://github.com/jfrog/jfrog-cli/issues/1304) | high |
| pnpm ≥ 10 doesn't run dependency build scripts unless allowed (`allowBuilds`; `strictDepBuilds` defaults on in v11). A postinstall design silently degrades to the launcher. This repo's own `pnpm-workspace.yaml` uses the same settings | [pnpm v10.0.0 release](https://github.com/pnpm/pnpm/releases/tag/v10.0.0); [pnpm settings/build](https://pnpm.io/settings/build) | high |

### 2.2 `.bin` entries (tested, npm 11.19.0, pnpm 12.4.1, Linux)

| Finding | Evidence | Conf |
|---|---|---|
| npm: `.bin` entries are symlinks. npm links bins of every package in the root `node_modules`, **including a hoisted transitive dependency**. A platform package that is an optional dep of the main package got `node_modules/.bin/<bin>` → the native binary | [bin-links bin-target.js L6-L9](https://github.com/npm/bin-links/blob/7f9f97ae2a41bd1ffea516993971b6291b5b0aca/lib/bin-target.js#L6-L9); BENCH (`--keep`, inspected) | high |
| pnpm: `.bin` entries are `/bin/sh` scripts. For a native target the script ends in `exec "<path>/bin/hook" "$@"` with no Node, after a `readlink`/`sed`/`uname`/`cd -P` preamble that costs ≈4 ms | [pnpm linker L463-L477](https://github.com/pnpm/pnpm/blob/d162953a10e63c7572b2a3f2076bbed1f2fcda89/pnpm11/bins/linker/src/index.ts#L463-L477); [cmd-shim L280-L298](https://github.com/pnpm/cmd-shim/blob/83f9b9449c2adeec26e1cc2d03ef0fe82d2e41ca/src/index.ts#L280-L298); BENCH | high |
| pnpm links bins of **direct** dependencies only: the transitive platform package got no `.bin` entry. With `publicHoistPattern: ["<platform pkgs>"]` in the team repo's `pnpm-workspace.yaml` it did | BENCH (scratch test, then the bench's pnpm team repo) | high |
| Both lockfiles record every platform package, including ones skipped on this machine (tested with an `aix-ppc64` package): npm with its `os`/`cpu`, and pnpm as `optional: true`. `npm ci` from that lockfile reinstalled cleanly. A same-named `bin` in the skipped package caused no conflict | BENCH (`--keep`, inspected) | high on Linux, local tarballs; medium for registry installs |
| npm/cli#4828 (a lockfile made on one platform drops the other platforms' optional packages) was addressed in npm 11.3.0, but #8320 reported the symptom later (npm 10.9.2, 11.4.1). Node 24 ships npm 11 | [npm CHANGELOG L1003](https://github.com/npm/cli/blob/0c3b82a9a612c3f9399d35c28c86708b1f8ea7d4/CHANGELOG.md#L1003); [npm/cli#8320](https://github.com/npm/cli/issues/8320) | medium |
| `--omit=optional` / `--no-optional` leaves the binary out, and the hook command then fails | [npm config `omit`](https://docs.npmjs.com/cli/v11/using-npm/config) | high |

### 2.3 Company mirrors, signing, release

| Finding | Evidence | Conf |
|---|---|---|
| No report found of Artifactory, Nexus or Verdaccio stripping `os`/`cpu` or dropping platform packages. Scoped-package 404s through mirrors have happened ([pnpm#4283](https://github.com/pnpm/pnpm/issues/4283), [verdaccio#4070](https://github.com/verdaccio/verdaccio/issues/4070)). Needs a smoke test on the real mirror | search; see links | low |
| macOS arm64 needs at least an ad-hoc signature. Go's linker adds one for darwin/arm64, including cross-builds from Linux. Deno ad-hoc signs. Node SEA must be unsigned before injection and re-signed after. npm-extracted files get no quarantine xattr (*inferred*; a pnpm store bug propagated it, fixed in [pnpm#11095](https://github.com/pnpm/pnpm/pull/11095)) | [go lib.go L301-L304](https://github.com/golang/go/blob/38f24c5c4659b7b8f468e5a2544412e9532e55e9/src/cmd/link/internal/ld/lib.go#L301-L304); [Deno compile](https://docs.deno.com/runtime/reference/cli/compile/); [Node SEA](https://nodejs.org/docs/latest-v24.x/api/single-executable-applications.html) | high (Go, Node); medium (quarantine) |
| Release pattern: build every target, generate the platform `package.json` files at the root version, publish the platform packages **first**, then the main package with exact `optionalDependencies`, all with provenance from a GitHub-hosted runner | [Biome release.yml L370-L416](https://github.com/biomejs/biome/blob/f89dd4f2d52e4919133429fa4550912510517a9a/.github/workflows/release.yml#L370-L416); [esbuild Makefile L445-L467](https://github.com/evanw/esbuild/blob/f6058f8364fe7ab91ca57a83e02577ed74c9cae4/Makefile#L445-L467) | high |
| Cross-compiling: Go does all four targets from one Linux job (`GOOS`/`GOARCH`, `CGO_ENABLED=0`, a static Linux binary, so no glibc/musl split). Biome (Rust) builds darwin on macOS runners. Bun (`--target`) and Deno (`--target`) cross-compile but embed their runtime | [Biome release.yml L120-L200](https://github.com/biomejs/biome/blob/f89dd4f2d52e4919133429fa4550912510517a9a/.github/workflows/release.yml#L120-L200); [Bun executables](https://bun.com/docs/bundler/executables) | high |
| Node 24: SEA is "Active development" (1.1) and needs `postject` (`--build-sea` arrived in v25.5). `module.enableCompileCache()` is stable since 24.15; `--build-snapshot` since 24.13.1 | [Node SEA](https://nodejs.org/docs/latest-v24.x/api/single-executable-applications.html); [module.enableCompileCache](https://nodejs.org/docs/latest-v24.x/api/module.html#moduleenablecompilecacheoptions); [--build-snapshot](https://nodejs.org/docs/latest-v24.x/api/cli.html#--build-snapshot) | high |

### 2.4 Hook command options

| Option | Works with npm | Works with pnpm | Postinstall | Per-machine step |
|---|---|---|---|---|
| A. `.bin/folkways hook <name>` (Node CLI, today's spec) | yes | yes | no | no |
| B. JS launcher → binary (Biome) | yes | yes | no | no; still pays a Node start |
| C. Postinstall swaps the launcher for the binary (esbuild) | fast path | launcher only, unless the team allows builds | **yes** | no |
| D. Main package + platform optional deps with their own `bin`, committed `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook <name>` | yes (hoisted) | with `publicHoistPattern` | no | no |
| E. As D, but the team repo lists the platform packages directly | yes | yes | no | no; 5 deps to bump together |
| F. `sync`/`onboard` writes the resolved path into gitignored `.claude/settings.local.json` (both tools read it) | yes | yes | no | **yes**; no hooks until it runs, and nothing fails loudly |

## 3. Benchmark

**Method.** Every candidate runs the same program, roughly `git-guard`'s real work ([bench/hooks/README.md](../../bench/hooks/README.md)):
read the stdin JSON, match the Bash command against a regex, parse `team.yaml` with a real YAML parser, check the
Angular header and scope, and write a deny JSON. `run.ts` checks every call path's output against the expected
JSON before timing it. Then it times 50 fresh processes per path, interleaved, as wall clock around
`spawnSync`, from a warm OS file cache. npm and pnpm paths install local tarballs into throwaway team repos in
layout D. Machine: Linux x64, Intel Xeon 2.1 GHz, 4 CPUs, Node 24.21.0, npm 11.19.0, pnpm 12.4.1, Go 1.24.7,
Rust 1.94.1, Bun 1.3.11, Deno 2.9.7. Full table: [results/linux-x64-2026-09-29.md](../../bench/hooks/results/linux-x64-2026-09-29.md).

| Candidate | Call path | p50 ms | p95 ms | Shipped size (MB, per platform) |
|---|---|--:|--:|--:|
| spawn floor (`true`) | direct | 2.4 | 2.9 | — |
| Node, plain ESM + `yaml` | direct / npm `.bin` / pnpm `.bin` | 65.1 / 65.8 / 72.0 | 81.0 / 75.4 / 81.2 | 0.1 + `yaml` |
| Node, bundled | direct / npm / pnpm | 52.4 / 54.7 / 58.3 | 63.2 / 61.9 / 66.8 | 0.3 |
| Node, bundle + compile cache | direct / npm / pnpm | 49.7 / 51.7 / 54.4 | 55.8 / 61.4 / 65.3 | 0.3 |
| Node, V8 startup snapshot | direct | 50.1 | 57.3 | 6.8 (blob) |
| Node SEA (code cache) | direct / npm / pnpm | 46.5 / 49.6 / 52.9 | 53.4 / 56.8 / 59.1 | 126.9 (44.3 packed) |
| Bun, `bun hook.ts` | direct | 81.9 | 97.0 | needs Bun installed |
| Bun `--compile` | direct / npm / pnpm | 53.5 / 54.8 / 57.0 | 61.1 / 61.0 / 66.7 | 99.4 (39.1 packed) |
| Deno `compile` | direct / npm / pnpm | 64.7 / 65.1 / 71.1 | 75.8 / 77.2 / 81.1 | 121.9 (41.3 packed) |
| **Go** | direct / npm `.bin` / pnpm `.bin` | **3.8 / 3.6 / 7.6** | 4.4 / 4.2 / 8.9 | **2.5** (1.1 packed) |
| Go, JS launcher (Biome pattern) | npm / pnpm | 36.3 / 38.1 | 44.1 / 43.1 | 2.5 |
| **Go, proposed command via `sh -c`** | npm / pnpm | **4.2 / 8.0** | 5.0 / 9.6 | — |
| Rust | direct / npm / pnpm | 3.7 / 3.7 / 7.7 | 4.5 / 4.3 / 9.4 | 2.0 (0.8 packed) |
| Rust, JS launcher | npm / pnpm | 34.5 / 37.8 | 42.3 / 45.5 | 2.0 |
| Rust, proposed command via `sh -c` | npm / pnpm | 4.6 / 8.6 | 5.2 / 9.7 | — |

**macOS arm64** (2026-09-29). Machine: Apple M5 Max, 18 CPUs, Node 24.21.0, npm 11.19.0, pnpm 12.4.1, Go 1.27.1.
Rust, Bun and Deno are not installed there, so their candidates were skipped. Full table:
[results/darwin-arm64-2026-09-29.md](../../bench/hooks/results/darwin-arm64-2026-09-29.md).

| Candidate | Call path | p50 ms | p95 ms | Shipped size (MB, per platform) |
|---|---|--:|--:|--:|
| spawn floor (`true`) | direct | 1.2 | 1.3 | — |
| Node, plain ESM + `yaml` | direct / npm / pnpm | 33.6 / 34.5 / 41.3 | 34.9 / 36.1 / 43.7 | 0.1 + `yaml` |
| Node, bundled | direct / npm / pnpm | 28.4 / 29.3 / 35.8 | 29.4 / 31.9 / 37.4 | 0.3 |
| Node, bundle + compile cache | direct / npm / pnpm | 27.0 / 27.9 / 34.8 | 28.2 / 29.2 / 36.0 | 0.3 |
| Node, V8 startup snapshot | direct | 25.3 | 27.1 | 6.9 (blob) |
| Node SEA (code cache) | direct / npm / pnpm | 26.1 / 25.9 / 33.7 | 28.0 / 27.2 / 37.4 | 121.5 (39.7 packed) |
| **Go** | direct / npm `.bin` / pnpm `.bin` | **2.6 / 2.2 / 9.3** | 2.6 / 2.4 / 9.7 | **3.0** (1.3 packed) |
| Go, JS launcher (Biome pattern) | npm / pnpm | 23.7 / 29.6 | 24.8 / 31.0 | 3.0 |
| **Go, proposed command via `sh -c`** | npm / pnpm | **4.5 / 11.3** | 4.6 / 11.6 | — |

Decision rule: Go unless Node lands under about 20 ms p50 through the npm and pnpm `.bin` paths. Node's best is 27.9
(npm) and 34.8 (pnpm), so Go is chosen. The first run after a build is slow (Go 115–285 ms, SEA 0.8–1.4 s), likely
macOS checking a new binary.

Also measured (Linux):

- **Install** of one candidate's package from a local tarball, warm cache: Go/Rust/Node 0.3 s npm, 0.1 s pnpm;
  Node SEA, Bun and Deno 1.7–2.5 s npm, 0.6 s pnpm. Registry downloads scale with packed size (1 MB vs about 40 MB).
- **Build:** Go 0.2 s (cached modules), Rust 18 s (release, LTO), Bun 0.1 s, Deno 1.4 s, SEA 3 s.
- **The full Go `git-guard` prototype** (all 51 git-guard contract cases pass): p50 3.2 ms for a commit check, 4.7 ms for a
  push check (it spawns `git rev-parse`), 2.9 ms for an unrelated command. 2.7 MB stripped.

**Reading** *(inferred)*:

1. Compiled Go or Rust is about 13× faster than the best Node option on this machine: 4 ms against 47–55 ms.
   Every JS runtime, compiled or not, sits at 46–82 ms, because runtime start-up dominates a job this small.
2. Node optimizations recover 13–19 ms from 65 ms and no more. The floor is Node start-up itself.
3. The call path matters as much as the language. A JS launcher in front of a native binary gives back ~31–33 ms,
   and a pnpm shim adds ~4 ms. A committed command through `sh -c` adds ~0.5 ms.
4. Size rules out the compiled-runtime options: 100–127 MB per platform, against 2–2.5 MB.
5. Per call the difference is ~45–65 ms. The hooks run on every Bash call (PreToolUse) and after every tool batch
   (PostToolBatch), so a busy session makes hundreds of calls. That is seconds per session, not minutes. The case for
   Go is responsiveness and headroom (future hooks doing more work), not throughput.
6. macOS arm64 (below) tells the same story: Node 25–41 ms, Go 2–9 ms. The best Node path is 27.9 ms, so Go wins
   there too.

## 4. Unverified

- Anything run in Copilot CLI (not installed here): loading `.claude/settings.json` hooks, payload shapes, whether
  it ignores the nested `hookSpecificOutput` in a combined output, `CLAUDE_PROJECT_DIR` for repo hooks, and
  UserPromptSubmit `additionalContext`.
- macOS: Rust, Bun and Deno (not installed on the test Mac), macOS x64, Gatekeeper on npm-installed binaries.
- Registry installs through a company mirror; npm lockfiles written on macOS and installed on Linux.
- The statusline's number (it doesn't run under `claude -p`); `/context` matched instead.
