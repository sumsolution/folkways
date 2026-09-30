I haven't changed any code. This session has no shell or edit tools (Bash is disabled, and I only have Read, Glob and Grep). I also can't run `go test` or create the worktree that the repo rules require for code changes. Below is what I found reading the code against the seven findings, so the fixes can go in quickly once I have the tools back.

The code is in `projects/folkways/hooks/internal/gitguard/`. I checked each finding against `guard.go` and `shell.go` and think all seven are valid.

| # | Finding | What the code does now | Proposed fix |
|---|---|---|---|
| 1 | `cd` with `$HOME` or `~` | `cdTarget` (`guard.go:142`) returns "unknown" for any `$`, backtick or `~`. `checkSeq` then treats the `cd` as not moving, so the original dir is checked. | Expand `~`, `~/x`, `$HOME` and `${HOME}` from `os.UserHomeDir`. For any other unknown target, check both the old dir and a "unknown" outcome. The simplest version is to deny `push` and main-clone commands after a `cd` whose target can't be resolved. |
| 2 | Attached `-C` and `-c` | `checkGit` only matches `-C` as a separate word. | Handle `-C<dir>` and `-c<k=v>` by prefix match before the default case. |
| 3 | `--git-dir` and `--work-tree` | They are skipped without changing `dir`. | Resolve `--work-tree` first and then `--git-dir`, which for a `.git` path means its parent. Apply `-C` first and resolve relative paths against the result. Both `--opt=v` and `--opt v` forms need handling. |
| 4 | `-m"msg"` and `-am"msg"` | `shortMsg` (`guard.go:19`) only matches a bare `-m` or `-am`. | Change the regex to `^-[a-zA-Z]*m(.*)$`. A non-empty capture is the message. Otherwise use the next word. |
| 5 | Unquoted heredoc bodies | `stripHeredocs` drops every body. | For an unquoted delimiter, keep the body text so `substitutions` sees `$(…)` and backticks. For a quoted delimiter, keep dropping it. This means recording whether the delimiter was quoted in `heredocStart`. |
| 6 | Backslash-newline | `parseSeq` skips `\` plus the newline, but `words` writes the newline into the word. | Handle it in one place. The simplest option is to remove `\`+newline (outside single quotes) before parsing, in `checkScript`. `cdTarget` then gets it for free. |
| 7 | `#` comments | Nothing recognizes them. | In `parseSeq`, at a token boundary, skip from an unquoted `#` to the newline. Also strip comments in `words` so the text handed to `checkSegment` is clean. |

Findings 2, 3 and 4 are small and independent. Findings 5, 6 and 7 all touch the `shell.go` lexer, so I'd do them together to avoid conflicting edits. For every finding I'd add regression cases to `guard_test.go` and `shell_test.go`, plus fixtures (currently 63). The tests should come first, as the `tests` skill asks. Finding 1 is the only one that needs a judgement call: expanding `$HOME` versus failing closed on an unknown target. I'd do both, expanding the common forms and denying on anything else unresolvable. That is stricter than today, which fails open, and the contract README already says the guard catches agent mistakes rather than a determined bypass.

To go ahead, I need a session with shell and edit access. I'd then:
1. Create a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard` off the PR branch, leaving your clone untouched.
2. Write the failing tests, then the fixes.
3. Commit as `fix(git-guard): …`.
4. Update `00-STATUS.md` with the round 2 results.

I would ask before pushing.
