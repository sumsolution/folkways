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

- Always `-count=1`: Go's test cache can't see a binary outside the module, so a cached pass can be stale.
- Command under test: `FOLKWAYS_CLI_CMD`, with an absolute path in its first word (the CLI runs in a temp dir;
  a relative path is refused), run as `/bin/sh -c "$FOLKWAYS_CLI_CMD <args>"` (each arg
  shell-quoted). Unset: `TestMain` builds `./cmd/folkways` from this checkout into a temp dir, with the version
  from `packages/folkways/package.json`, so `go test ./...` pins the Go CLI.
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
| `${VERSION}` | `version` of `packages/folkways/package.json`. Build a Go binary with `-ldflags "-X main.version=<version>"` to match |
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

1. Run the CLI by hand to see the exact output.
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

## History

The fixtures were ported from the TypeScript CLI's vitest suites, removed with the TS CLI after
[`165ec4c`](https://github.com/sumsolution/folkways/tree/165ec4c/packages/cli/test). Each folder's `COVERAGE.md`
maps those tests to fixtures; the test names there refer to that commit.
