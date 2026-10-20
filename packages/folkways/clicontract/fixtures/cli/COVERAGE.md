# Coverage: packages/cli/test/cli.test.ts

Every `it` → fixture in this folder, or why not.

| `it` | Fixture |
|---|---|
| prints usage | `help` (seed; also pins `--help`, `-h`, exact text) |
| prints the package version | `version` (seed; also pins `--version`) |
| init then sync | `init-then-sync-from-subfolder` |
| hook runs the installed binary with the payload on stdin | skipped: TS-only, the Go CLI runs hooks in-process (hook contract suite covers hooks) |
| hook explains a missing binary | skipped: TS-only. The missing-binary half: no `folkways-hook` binary in Go. The usage half (`hook` with no name): TS exits 1, Go exits 0 because the hook path never fails (Copilot denies the tool call on a nonzero exit) |
| doctor prints each finding and exits 1 on errors | `doctor-prints-findings-and-exits-1` |
| prints expected errors without a stack trace | `expected-error-without-stack-trace` |
| rejects unknown options | `unknown-option` (pins `folkways: Unknown option '--nope'` only, not Node's trailing hint) |
| exits 1 on an unknown command | `unknown-command` (seed; also pins usage on stderr) |

Extras with no TS test, pinning dispatch code in `cli.ts`:

| Fixture | Behavior |
|---|---|
| `not-built-commands` | `review` and `run`: `folkways: '<cmd>' is not built yet`, exit 1 |
| `outside-team-repo` | `sync` and `doctor` outside a git repo: `folkways: run this inside the team repo`, exit 1. `hook <name>` outside a repo is not pinned: TS-only, the Go CLI runs hooks in-process (ADR 0011) |
