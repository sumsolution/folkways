# Hook contract suite

Black-box tests for the five hooks in spec §7.1. They run **any** implementation as a subprocess, so the
same suite checks a Node, Go or Rust build. Defined by [ADR 0002](../../docs/adr/0002-hook-runtime.md);
tool behavior behind each rule: [08-research-hook-runtime.md](../../docs/design/08-research-hook-runtime.md).

## Run

```
FOLKWAYS_HOOK_CMD="/abs/path/to/folkways-hook" pnpm test test/hooks
FOLKWAYS_HOOK_CMD="node /abs/path/to/src/cli.ts hook" FOLKWAYS_HOOKS=git-guard,session-end pnpm test test/hooks
```

- The harness runs `/bin/sh -c "$FOLKWAYS_HOOK_CMD <hook>"`, as both tools do, with the payload on stdin.
- Without `FOLKWAYS_HOOK_CMD` the contract tests skip ("set FOLKWAYS_HOOK_CMD to run"). `harness.test.ts`
  and a fixture-count check still run, so `pnpm test` stays green.
- `FOLKWAYS_HOOKS` limits the run to some hooks (a prototype that only has `git-guard`).

## Fixtures

`fixtures/<hook>/<case>.json`:

| Field | Meaning |
|---|---|
| `description` | one line, shown in the test name |
| `env.team` | deep-merged into the base `team.yaml` (scopes `web-app, auth, ci`, `review.trigger: pre-push`, bands 125K/150K/200K) |
| `env.me` | deep-merged into the base `.local/me.yaml` (`handle: alex`, `projects_dir: ${PROJECTS}`); `null` removes it |
| `env.files` | extra files in the team repo (review files, session markers) |
| `env.claudeProjectDir` | default `true`; `false` leaves `CLAUDE_PROJECT_DIR` unset, as Copilot CLI does for repo hooks |
| `env.transcript` | transcript lines: a number is a main-chain assistant line with that many tokens in context, `{ "sidechain": n }` a subagent line, an object is written as is |
| `steps[]` | one or more calls in the same environment: `payload` (stdin), optional `cwd`, optional `transcript` (replaces the file), `expect` |
| `expect.exit` | default `0` |
| `expect.stdout` | `null` = empty; an object = stdout is JSON containing it (deep subset). Strings starting `re:` are case-insensitive regexes |
| `expect.exists` / `absent` / `contains` | files after the step (`contains`: path → regex) |

Placeholders, filled in payloads, files and expectations: `${TEAM}` (team repo, the default cwd and
`CLAUDE_PROJECT_DIR`), `${PROJECTS}`, `${MAIN}` (a main clone under `projects_dir`), `${WORKTREE}` (a worktree
of it on `feat/EX-1-orders`), `${HEAD}` / `${HEAD7}` (the worktree's HEAD), `${OLD_HEAD}` / `${OLD_HEAD7}` (its first commit), `${TRANSCRIPT}`.

## The contract

Rules every implementation must follow. Where spec §7.1 leaves details open, this contract pins them down;
change them here and in the fixtures together.

**All hooks**

- Exit 0, always, including on bad input. Copilot CLI denies the tool call when a `preToolUse` hook exits
  nonzero or crashes ([08 §1](../../docs/design/08-research-hook-runtime.md)), so a bug must fail open.
- Find the team repo from `CLAUDE_PROJECT_DIR`, else the process cwd. Claude Code's hook cwd follows the
  Bash tool's `cd`; `CLAUDE_PROJECT_DIR` doesn't. Copilot CLI runs repo hooks in the project root.
- Accept Claude Code payloads and both Copilot CLI shapes: PascalCase events (snake_case fields, ISO
  `timestamp`, `tool_input` possibly a JSON string) and camelCase events (`sessionId`, `toolName: "bash"`,
  `toolArgs` as a JSON string).
- Print nothing unless there is something to say. Never print `permissionDecision: "allow"`: it would skip
  the user's permission rules.

**Output shapes**

| Case | stdout (one JSON object) |
|---|---|
| Deny a tool call | `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":R},"permissionDecision":"deny","permissionDecisionReason":R}`. The nested form is Claude Code's, the top-level form Copilot CLI's. Claude Code accepted both together (tested, 08). |
| Add context | `{"hookSpecificOutput":{"hookEventName":<event>,"additionalContext":T}}`, plus top-level `"additionalContext":T` for `SessionStart` (Copilot CLI reads only that) |

**`git-guard`** (`PreToolUse`, Bash)

- Target repo: the last `cd <dir>` before the git command, else `git -C <dir>`, else the payload `cwd`.
- `git commit` with `-m`/`-am`/`--message` (double, single quotes, or the `$(cat <<'EOF' … EOF)` heredoc):
  the first line must be Angular `type(scope)!?: subject`. When `git.scopes` is non-empty the scope is
  **required** and must be listed. No message flag (`--no-edit`, `-F`): pass.
- `git push` when the effective `review.trigger` is `pre-push`: deny unless the team repo holds a review
  covering the target repo's HEAD. The effective trigger is the stricter of `team.yaml` and
  `.local/me.yaml` (`on-demand` < `after-commit` < `pre-push`); a person can tighten it, not loosen it.
  A review is either a file under `efforts/**` or `people/<handle>/work/**` with frontmatter
  `sha: <full HEAD>`, or `people/<handle>/inbox/<YYYY-MM-DD>-review-<sha7>.md` with `<sha7>` the first 7
  characters of HEAD.
- `checkout`, `switch`, `pull`, `reset`, `rebase` in a **main clone** under `projects_dir` (its `.git` is a
  directory): deny, and say to use a worktree. Worktrees (`.git` is a file), repos outside `projects_dir`,
  and other commands (`fetch`, `worktree add`) pass. No `.local/me.yaml`: no clone guard.
- Every git command in a chain (`&&`, `;`, `||`) is checked.

**`session-start`** (`SessionStart`): always remind the `start` procedure. `source: compact`: say to update
`00-STATUS.md` now. Any `.local/sessions/*.unclosed`: offer `session-close`.

**`session-end`** (`SessionEnd` / `sessionEnd`): write `.local/sessions/<session id>.unclosed` in the team
repo. No output. No session id: do nothing.

**`precompact`** (`PreCompact`): append a line with the session id and `trigger` to `.local/precompact.log`.
No output; never block compaction.

**`context-band`** (`UserPromptSubmit`, `PostToolBatch`): tokens in context = the last transcript line with
`type: "assistant"`, `isSidechain: false` and non-zero usage: `input_tokens + cache_read_input_tokens +
cache_creation_input_tokens`. Bands from `team.yaml` `handoff.bands`. A threshold counts when context is `>=` it. When the band rises above the last
one announced for this session, inject the band text naming the threshold (`125K`); announce only the
highest band crossed. When context falls below the last band (compaction), lower the stored band so the
next crossing is announced again. State lives in `.local/`. No transcript
(Copilot CLI) or a missing file: do nothing.
