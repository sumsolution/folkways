# CLI contract suite

Black-box tests for the `folkways` CLI (`init`, `sync`, `upgrade`, `doctor`, `version`, `help`). They run
**any** implementation as a subprocess. By default they pin the Go CLI
([ADR 0011](../../../docs/adr/0011-cli-in-go.md)); `FOLKWAYS_CLI_CMD` points them at another build. Separate from the hook contract in
[`../contract/`](../contract/README.md); `hook` itself is covered there.

## Run

```
cd packages/folkways && go test -count=1 ./clicontract/
FOLKWAYS_CLI_CMD="$PWD/bin/folkways" go test -count=1 ./clicontract/
FOLKWAYS_CLI_CMD="$PWD/bin/folkways" go test -count=1 ./clicontract/ -run 'TestContract/sync/'
```

- Always `-count=1`: Go's test cache can't see the TS sources, the content or a binary outside the module, so
  a cached pass can be stale.
- Command under test: `FOLKWAYS_CLI_CMD`, with an absolute path in its first word (the CLI runs in a temp dir;
  a relative path is refused), run as `/bin/sh -c "$FOLKWAYS_CLI_CMD <args>"` (each arg
  shell-quoted). Unset: `TestMain` builds `./cmd/folkways` from this checkout into a temp dir, with the version
  from `packages/cli/package.json`, so `go test ./...` pins the Go CLI. The TypeScript CLI can still be run by
  hand: `FOLKWAYS_CLI_CMD="node $PWD/../cli/src/cli.ts"`.
- Each fixture runs in its own temp root, in parallel. The CLI gets a fresh environment: nothing leaks in from
  the caller (no `npm_config_user_agent`, no git config). It has `PATH` (stubs first), `HOME` and
  `XDG_CONFIG_HOME` under the root, the caller's `TMPDIR`, `TZ=UTC`, `LC_ALL=C`, git isolated (`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`)
  with identity `test <test@example.com>` and author/committer date `2026-01-01T00:00:00Z`. The npm and pnpm
  stubs keep installs off the network; nothing blocks other tools on the caller's `PATH`.

## Fixtures

`fixtures/<command>/<case>.json`, one JSON object per file. Test name: `TestContract/<command>/<case>`. Unknown
fields and a second JSON value fail the case, so a typo can't silently skip a check. Field names match
case-insensitively (Go's `encoding/json`); write them as below. A case is a list of steps run in order in the
same root; each step sets up, optionally runs the CLI, then checks.

| Field | Meaning |
|---|---|
| `description` | one line: the behavior pinned |
| `steps[]` | one or more steps |
| `steps[].files` | written before the call: path → content. `null` deletes the path (file or folder). Content starting with `#!` is made executable. Folders are created |
| `steps[].git` | git commands run in `${TEAM}` before the call (after `files`), e.g. `[["add", "-A"], ["commit", "-qm", "edit"]]`. Creates `${TEAM}` if missing. A failing command fails the case |
| `steps[].args` | CLI arguments. Omitted: no CLI call (a setup-only step). `[]`: the CLI with no arguments |
| `steps[].cwd` | process cwd. Default `${TEAM}` when it exists, else `${ROOT}` |
| `steps[].env` | extra environment: name → value; `null` unsets a default (e.g. `GIT_AUTHOR_NAME`) |
| `steps[].stdin` | stdin. Default empty |
| `steps[].expect` | checks. Omitted: the call must exit 0, nothing else is checked |
| `expect.exit` | default `0` |
| `expect.stdout` / `stderr` | exact string, or a Go regex after `re:`. A regex is unanchored: add `^` and `$` for a full match. Omitted: not checked. Write `""` to require empty |
| `expect.files` | path → exact content, `re:` regex, `null` (must not exist) or `true` (must exist) |
| `expect.tree` | the files under `${TEAM}` (not `.git/`), as patterns: each file must match one, each pattern must match a file. `path.Match` globs (`*/README.md`); `dir/**` matches everything under `dir` |
| `expect.unchanged` | `true`: every file under `${TEAM}` (not `.git/`) has the same content and mode as before this step |
| `expect.git[]` | `{ "args": [...], "stdout": "..." }`: git run in `${TEAM}`, trimmed stdout exact or `re:`. Clean tree: `["status", "--porcelain"]` → `""` |

Paths in `files` and `expect.files` are relative to `${TEAM}` unless absolute after substitution
(`${HOME}/.claude/skills/x/SKILL.md`, `${CALLS}`).

Placeholders, filled in every string (args, env, stdin, file paths and contents, expectations; in `re:`
expectations the value is regex-quoted):

| Placeholder | Value |
|---|---|
| `${ROOT}` | the case's temp root (symlinks resolved) |
| `${TEAM}` | `${ROOT}/team`; not created for you: `init` it, or write files or run git there |
| `${HOME}` | the CLI's `HOME`, `${ROOT}/home` (personal skills for doctor's shadow check) |
| `${BIN}` | the stub folder, first on `PATH`. Write a script here to wrap or replace a tool |
| `${CALLS}` | the stub call log (below) |
| `${VERSION}` / `${HOOKS_VERSION}` | `version` and `folkways.hooks` of `packages/cli/package.json`. Build a Go binary with `-ldflags "-X main.version=<version>"` to match. Once `packages/cli` is gone, the source moves to the Go binary's `-X main.version` and its pinned hooks version |
| `${CLI}` | the command under test, for a script that should call it |
| `${GIT}` | the real git, for a wrapper in `${BIN}` |

Don't use these names for anything else in file contents: `${HOME}` in a shell script is replaced too.

### Stubs

`npm` and `pnpm` on `PATH` are stubs (no network). Each appends `<cwd>: <tool> <args>` to `${CALLS}`, so a
fixture asserts the calls with `"files": { "${CALLS}": "${TEAM}: npm install\n" }` (`null`: never called). Then:

| Call | Does |
|---|---|
| `STUB_NPM_FAIL` / `STUB_PNPM_FAIL` set (step `env`) | prints the value to stderr, exits 1 |
| `--version` | prints `10.0.0` (npm) or `9.0.0` (pnpm), so `init --package-manager pnpm` pins `pnpm@9.0.0` |
| `install` | reads the `@folkways/cli` pin in `package.json`; writes the lockfile (`package-lock.json` or `pnpm-lock.yaml`, content `stub lockfile`), `node_modules/@folkways/cli/package.json` at the pinned version, `node_modules/.bin/folkways` (runs `${CLI}`, so `upgrade` syncs with the CLI under test) and `node_modules/.bin/folkways-hook` (exits 0) |
| `install`/`add` `--save-dev --save-exact <pkg>@<v>` | rewrites `<pkg>`'s version in `package.json` in place (it must be there already), then as `install` |
| anything else | exits 0 |

The stubs are this test binary re-run with `FOLKWAYS_CLI_STUB=<tool>` (`stubs_test.go`).

### Writing one

1. Find the behavior in `packages/cli/test` (map below). Run the TS CLI by hand to see the exact output.
2. Build the starting state with steps: `init` (`["init", "${TEAM}", "--no-install"]`, or without
   `--no-install` for `node_modules`), then `files` and `git` to edit and commit.
3. Pin exact stdout/stderr where the text is the contract; `re:` where it embeds content that changes with
   `packages/content` (skill lists, settings fragments).
4. `go test -count=1 ./clicontract/ -run 'TestContract/<command>/<case>'`, then break the expectation once to see it fail.

```json
{
  "description": "sync refuses an edited framework file and names it",
  "steps": [
    { "args": ["init", "${TEAM}", "--no-install"] },
    {
      "files": { ".folkways/instructions.md": "mine\n" },
      "args": ["sync"],
      "expect": { "exit": 1, "stdout": "", "stderr": "re:modified  \\.folkways/instructions\\.md", "unchanged": true }
    }
  ]
}
```

## What the TS tests need

Per file in `packages/cli/test`. "Unit" means the test calls an internal function with inputs the CLI can't be
given; those stay unit tests, rewritten in Go beside the port.

| Test file | Harness features | Can't express yet |
|---|---|---|
| `cli.test.ts` | args, `cwd` (subfolder), stderr exact/`re:`, `files` with `#!` scripts, `stdin`, exit passthrough, `${HOME}` | `hook` tests belong in `../contract/`: the Go CLI runs hooks in-process, not `node_modules/.bin/folkways-hook` (ADR 0011), so the passthrough tests are TS-only |
| `init.test.ts` | `tree`, `files`, `expect.git`, `${CALLS}`, `STUB_NPM_FAIL`, env `null` (no git identity), `${BIN}/git` wrapper with `${GIT}`, folder names via `init "${ROOT}/My Team"` | `templateTarget` cases (unit); `packageName` with a 300-char folder (over `NAME_MAX`) and `''` (unit). An existing **empty** folder needs a setup step that writes then deletes a file |
| `sync.test.ts` | `files`, `git`, `unchanged`, `re:` on refusals, manifest edits, pnpm workspace variants, `--force` | Everything that uses `makePackage`/`addFrameworkContent` gets the real content at `${VERSION}` instead. "Previous version" works by editing the manifest's `version`. **Gap:** a manifest listing files this version dropped needs their sha256, which fixtures can't compute (candidate: a `${SHA256:path}` placeholder). Atomicity tests inject `fsOps` failures (unit). Executable bit: no real content file is executable |
| `upgrade.test.ts` | stubs (`add`/`install` pin and link `${CLI}`), `STUB_NPM_FAIL`, `expect.git` (branch, log, `status`), env `null`. Start "older" by setting the pin and manifest `version` to e.g. `0.0.1` and committing, then `upgrade ${VERSION}` | **Gap:** "a failed sync leaves files behind" needs `node_modules/.bin/folkways` replaced *after* the stub's install writes it (candidate: an env var naming a script the stub installs instead). Tests that the new version brings new skills can't differ in content: there is one CLI |
| `doctor.test.ts` | `init` with stub install, `files` (node_modules versions, settings), `${HOME}` skills, a plugin-style repo (`git init` + `folkways.yaml` only) for the `hookValueFindings` cases | `HOOK_NAMES` vs `assets/settings.json` is a content check, not CLI behavior |
| `overlays.test.ts` | through `sync` with `overlays/<skill>/OVERLAY.md` against real skill section IDs | Mostly unit (`parseOverlay`, `compileSkillFiles` on synthetic skills). Fixtures depend on which `{#id}`s the real skills have |
| `settings.test.ts` | through `sync`: a team `.claude/settings.json` in, merged file out (`re:` or exact) | Exact output embeds the real settings fragment; content changes break it. The `assets/settings.json` wiring tests are content checks |
| `package-manager.test.ts` | `${CALLS}` shows which tool `init`/`upgrade` ran; env `npm_config_user_agent` | none |
| `package.test.ts` | — | Packaging internals (`assetsDir`, `hooksVersion` file layout); not CLI behavior |
| `content-sync.test.ts` | — | **Gap:** "synced file equals `packages/content/...`" needs an expectation that compares with a repo file (candidate: `${REPO}` plus a `same-as:` value) |

Error texts from Node's `parseArgs` (`Unknown option '--x'. To specify a positional argument ...`,
`Unexpected argument 'extra'. This command does not take positional arguments`) are TS-runtime wording. Pin
only the `folkways: Unknown option '--x'` prefix unless the port should copy Node's text.
