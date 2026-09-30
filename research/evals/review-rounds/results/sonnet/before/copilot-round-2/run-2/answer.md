I haven't changed anything yet. Bash is disabled in this session, so I can't set up the worktree the rules require for code changes, and I can't run `go test` either.

The rules put code changes in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`, and no worktree exists yet. Editing the files in your `projects/folkways` clone would break the "never touch the user's clones" rule. I'd also be making seven parser changes without being able to run the tests.

I read `guard.go` and `shell.go` at the current checkout. All seven findings look real against the code:

| # | Cause | Planned fix |
|---|---|---|
| 1 | `cdTarget` returns false for any target containing `$`, `` ` `` or `~` (`guard.go:142`). The caller then keeps the original directory as the `cd` destination. | Expand `~`, `$HOME` and `${HOME}`. For other unknown targets, see the question below. |
| 2 | `checkGit` only matches `-C` and `-c` as separate words (`guard.go:187`). | Handle `-C<dir>` and `-c<k=v>` as attached forms. |
| 3 | `--git-dir` and `--work-tree` are skipped without changing `dir` (`guard.go:190`). | Derive the effective directory from them, with `-C` applied in order. |
| 4 | `shortMsg` only matches `-m` and `-am` as whole words (`guard.go:19`). | Match `-m<msg>` and `-[a-z]*m<msg>`. |
| 5 | `stripHeredocs` drops every heredoc body (`shell.go:279`). | Return whether the delimiter was quoted. For unquoted heredocs, keep only the `$(…)` and backtick substitutions from the body. |
| 6 | `words` keeps the newline from a backslash-newline, while `parseSeq` treats it as a continuation (`shell.go:323`). | Remove backslash-newline once, outside single quotes, before parsing. |
| 7 | `parseSeq` doesn't recognise `#` comments (`shell.go:162`). | Skip from an unquoted `#` at a word start to the newline. |

Each fix would get a regression fixture in `guard_test.go` or `shell_test.go`.

**Decisions for you**
1. **Bash:** can you re-enable it? Or tell me to edit the clone directly and skip the worktree; you'd then handle the branch state.
2. **Unknown `cd` targets:** for something like `cd "$(mktemp -d)"` or `cd $FOO`, I suggest checking the original directory and leaving the target unchecked. The alternative is to deny `push`, `reset` and the other guarded commands after it, which would block harmless commands. The README says this guard catches agent mistakes, not a determined bypass, so I lean towards the first.

Once I can run commands, I'll create the worktree, write the failing fixtures first, fix, run `go test ./...`, and commit as `fix(gitguard): …`. I'll ask before any push.
