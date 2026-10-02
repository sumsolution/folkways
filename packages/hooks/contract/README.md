# Hook contract suite

Black-box tests for the four hooks in spec §7.1. They run **any** implementation as a subprocess, so the
same suite checks a Node, Go or Rust build. Defined by [ADR 0002](../../../docs/adr/0002-hook-runtime.md);
tool behavior behind each rule: [08-research-hook-runtime.md](../../../docs/design/08-research-hook-runtime.md).

## Run

```
FOLKWAYS_HOOK_CMD="/abs/path/to/folkways-hook" pnpm --filter @folkways/hooks test
FOLKWAYS_HOOK_CMD="node /abs/path/to/packages/cli/src/cli.ts hook" FOLKWAYS_HOOKS=git-guard,precompact pnpm --filter @folkways/hooks test
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
| `env.team` | deep-merged into the base `folkways.yaml` (`git.commit_convention: angular`, scopes `web-app, auth, ci`, `review.trigger: pre-push`, bands 125K/150K/200K); `null` writes no `folkways.yaml`: an unconfigured repo |
| `env.me` | deep-merged into the base `.local/me.yaml` (`handle: alex`, `projects_dir: ${PROJECTS}`); `null` removes it |
| `env.files` | extra files in the team repo (review files, state) |
| `env.claudeProjectDir` | default `true`; `false` leaves `CLAUDE_PROJECT_DIR` unset, as Copilot CLI does for repo hooks; a string sets it to that path (e.g. `${TEAM}/sub`) |
| `env.commits` | makes the team repo a git repo (branch `main`, `.local/` excluded): an initial commit of everything built so far, then one commit per entry `{ message?, files }`. `${COMMIT_n}` / `${COMMIT_n_7}` (1-based) are their SHAs, `${TEAM_HEAD}` / `${TEAM_HEAD7}` the last one's |
| `env.offered` | seeds the setup-offer file `${XDG}/folkways/offered` (a list of lines); `"unwritable"` puts a directory there instead |
| `env.transcript` | transcript lines: a number is a main-chain assistant line with that many tokens in context, `{ "sidechain": n }` a subagent line, an object is written as is |
| `steps[]` | one or more calls in the same environment: `payload` (stdin), optional `cwd`, optional `transcript` (replaces the file), `expect` |
| `expect.exit` | default `0` |
| `expect.stdout` | `null` = empty; an object = stdout is JSON containing it (deep subset). Strings starting `re:` are case-insensitive regexes |
| `expect.exists` / `absent` / `contains` | files after the step (`contains`: path → regex) |

Placeholders, filled in payloads, files and expectations: `${TEAM}` (team repo, the default cwd and
`CLAUDE_PROJECT_DIR`), `${PROJECTS}`, `${MAIN}` (a main clone under `projects_dir`), `${WORKTREE}` (a worktree
of it on `feat/EX-1-orders`), `${HEAD}` / `${HEAD7}` (the worktree's HEAD), `${OLD_HEAD}` / `${OLD_HEAD7}` (its first commit), `${TRANSCRIPT}`, `${XDG}` (the hook's `XDG_CONFIG_HOME`, one per case: the setup offer's `offered` file is
`${XDG}/folkways/offered`).

## Versioning

The contract is the interface between `@folkways/hooks` and `@folkways/cli`, which writes hook names and
events into team repos' `settings.json` and pins one hooks version (`folkways.hooks` in its `package.json`,
[ADR 0003](../../../docs/adr/0003-repo-layout.md)). Renaming or removing a hook, or changing its event, is a
major hooks bump and needs a matching CLI change. New rules or fixtures that tighten behaviour are minor.

## The contract

Rules every implementation must follow. Where spec §7.1 leaves details open, this contract pins them down;
change them here and in the fixtures together.

**All hooks**

- Exit 0, always, including on bad input. Copilot CLI denies the tool call when a `preToolUse` hook exits
  nonzero or crashes ([08 §1](../../../docs/design/08-research-hook-runtime.md)), so a bug must fail open.
- Find the team repo from `CLAUDE_PROJECT_DIR`, else the process cwd. Claude Code's hook cwd follows the
  Bash tool's `cd`; `CLAUDE_PROJECT_DIR` doesn't. Copilot CLI runs repo hooks in the project root.
- **Config gates every hook** (spec §7.1). No `folkways.yaml` at the project root: print nothing and write
  nothing in the repo (`session-start` alone makes the one-time offer, below). `hooks.<name>: false` in
  `folkways.yaml`: that hook prints nothing and writes nothing; a missing key is on.
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

- Target repo: the current directory (see the model below), moved by `git -C <dir>`; it starts at the payload `cwd`.
- `git commit` with `-m`/`-am`/`--message` (double, single quotes, or the `$(cat <<'EOF' … EOF)` heredoc):
  the first line is checked per `git.commit_convention`. `angular` (also when the key is missing): Angular
  `type(scope)!?: subject`; when `git.scopes` is non-empty the scope is **required** and must be listed.
  `pattern`: the subject must match `git.commit_pattern` (Go RE2); scopes are ignored; an empty or invalid
  pattern checks nothing. `none`: no check. No message flag (`--no-edit`, `-F`): pass.
- `git push` when the effective `review.trigger` is `pre-push`: deny unless the team repo holds a review
  covering the target repo's HEAD. The effective trigger is the stricter of `folkways.yaml` and
  `.local/me.yaml` (`on-demand` < `after-commit` < `pre-push`); a person can tighten it, not loosen it.
  A review is either a file under `efforts/**` or `<personal root>/work/**` with frontmatter
  `sha: <full HEAD>`, or `<personal root>/inbox/<YYYY-MM-DD>-review-<sha7>.md` with `<sha7>` the first 7
  characters of HEAD. The personal root is `people/<handle>/` (`handle` in `.local/me.yaml`) in a team repo
  and the repo root when `mode: solo` (no handle needed). Only the person's own folder counts; another
  person's `people/<other>/` is ignored, and in a team repo with no handle only `efforts/**` counts.
- `checkout`, `switch`, `pull`, `reset`, `rebase` in a **main clone** under `projects_dir` (its `.git` is a
  directory): deny, and say to use a worktree. The folkways repo itself (top level = project dir, in either mode) is never denied. Worktrees (`.git` is a file), repos outside `projects_dir`,
  and other commands (`fetch`, `worktree add`) pass. No `.local/me.yaml`: no clone guard.
- The command line is parsed as bash (`mvdan.cc/sh/v3/syntax`); a line that does not parse is allowed. Every
  simple command is walked in source order, also inside `&&`, `||`, `;`, pipelines, `( )`, `{ }`, `if`/`while`/
  `for`/`case` and substitutions (`$( )`, backticks, `<( )`, unquoted heredoc bodies). Quoted heredoc bodies and
  comments are data. Words resolve quotes, escapes, `~`, `$HOME` and `${HOME}`; any other expansion is unknown.
- **One current directory.** Only a top-level `cd <resolvable dir>` in an `&&` or `;` chain (also inside `( )`,
  `{ }` or a substitution) moves it, and an absolute one always makes it known. Any other `cd`, `pushd` or `popd`
  (`cd "$X"`, `cd` in `||`, `if`, a pipeline or the background) makes it unknown; a `( )` subshell never
  changes the outer directory. Git options: `-C <resolvable>` moves it; `-c`, `--no-pager`, `-p`, `-P`, ... are
  skipped; any other option before the subcommand, or a `GIT_DIR=`/`GIT_WORK_TREE=` prefix, makes it unknown.
- **Unknown means skip:** no push or main-clone check (fail open). The commit message check always runs.
- A commit message is read from `-m`, `-am`, `-m<msg>`, `--message <msg>`, `--message=<msg>` or
  `-m "$(cat <<'EOF' ... EOF)"`. A short-option cluster is read left to right: an `m` ends it with the
  message; an option with an operand (`-F`, `-c`, `-C`, `-t`, and `-S`/`-u` with an attached one) ends it
  without one, so `-Skeyname -m "..."` and `-Cmain` are not misread. Any other expansion in a message is not
  evaluated (pass).
- git-guard catches agent mistakes, not a determined bypass: `bash -c`, `eval`, aliases and scripts are out of scope.

**`session-start`** (`SessionStart`): in a configured repo, always remind the `start` procedure. `source: compact`: say to update
`00-STATUS.md` now. It doesn't offer `session-close`: most sessions aren't worth banking, and the person runs
it when one is.

In an unconfigured repo it makes the one-time offer instead. Key: the git top level of the project dir, else
the project dir (absolute). When that key holds a `folkways.yaml` (a session in a subfolder of a configured repo): no offer, nothing recorded. Look it up in `${XDG_CONFIG_HOME:-$HOME/.config}/folkways/offered` (one path per
line). Absent: append it (creating directories) and inject the spec §7.1 line, naming the `setup` skill (`folkways:setup`
in the plugin). Present, or the file can't be read or written: print nothing. Never write in the repo.

**`precompact`** (`PreCompact`): append a line with the session id and `trigger` to `.local/precompact.log`.
No output; never block compaction.

**`context-band`** (`UserPromptSubmit`, `PostToolBatch`): tokens in context = the last transcript line with
`type: "assistant"`, `isSidechain: false` and non-zero usage: `input_tokens + cache_read_input_tokens +
cache_creation_input_tokens`. Bands from `folkways.yaml` `handoff.bands`. A threshold counts when context is `>=` it. When the band rises above the last
one announced for this session, inject the band text naming the threshold (`125K`); announce only the
highest band crossed. When context falls below the last band (compaction), lower the stored band so the
next crossing is announced again. State lives in `.local/`. No transcript
(Copilot CLI) or a missing file: do nothing.
