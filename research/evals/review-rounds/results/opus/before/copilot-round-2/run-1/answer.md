I read the seven findings against `hooks/internal/gitguard/guard.go` and `shell.go` in `projects/folkways`. Six look real and one looks wrong.

I couldn't fix anything. This session only has read and search tools, so I can't edit files, run `go test` or run `git`. I also couldn't confirm the checkout is at 8188289: `projects/folkways` has no `.git`.

## Triage

| # | Verdict | What's wrong | Fix |
|---|---|---|---|
| 1 `cd $HOME/~` | Real | `cdTarget` rejects any target containing `$`, `` ` `` or `~` (`guard.go:142`). The guard then treats the command as "not a cd" and checks the next git command in the old directory. | Expand `~`, `~/…`, `$HOME` and `${HOME}` with `os.UserHomeDir()`. Add tests for both forms. |
| 2 attached `-C`/`-c` | Probably wrong | As far as I remember, git's `handle_options` in `git.c` matches `-C` and `-c` with `strcmp`, so `git -C/path push` fails with "unknown option" and nothing runs. That's from memory, not checked. | Check with `git -C/tmp status`. If git rejects it, reply on the PR instead of changing code. |
| 3 `--git-dir`/`--work-tree` | Real | Lines 190–193 skip these options without updating the target, so `checkMainClone` and `checkPush` look at the wrong repo. | Track git-dir and work-tree and pass them to `gitOut`. For the main-clone check, use the work tree, or the parent of the git dir. Test both guards without `-C`. |
| 4 `-m"msg"` | Real | `shortMsg` (`^-[a-zA-Z]*m$`) only matches when `m` is the last letter, so `-mbad subject` and `-ambad subject` are never read as messages. | Read the short-option bundle left to right. At `m`, the value is the rest of the token, or the next argument if the rest is empty. Stop at other options that take a value (`F`, `C`, `c`, `t`). |
| 5 unquoted heredoc | Real | `stripHeredocs` drops every heredoc body (`shell.go:279`), including unquoted ones where bash runs `$( )` and backticks. | Record whether the delimiter was quoted. For unquoted bodies, scan with a heredoc-specific scanner: only `\`, `$(` and backticks count, and `'`/`"` are plain text. Reusing `substitutions()` would be wrong because an apostrophe in the text would hide a `$(git push)` after it. |
| 6 backslash-newline | Real | `words` turns `\` + newline into a newline inside the word (`shell.go:323`), so the subcommand becomes `"\npush"`. `appendDouble` (`shell.go:351`) does the same inside double quotes. | Drop `\` + newline in both places. Add a multiline push test and a multiline `cd` test. |
| 7 comments | Real, and worse than reported | Besides the false deny, a comment like `# don't` makes `parseSeq` and `stripHeredocs` treat the `'` as an opening quote. Everything up to the next `'` is then hidden, which can let a later-line `git push` through. A `# <<EOF` in a comment can likewise start a fake heredoc and hide lines. | Skip `#` at the start of a word through to the newline, leaving the newline in place. Do this in `parseSeq`, `stripHeredocs` and `skipSubst`. Test an inline comment, a comment with an apostrophe, and a command on the next line. |

## Decisions for you

1. **#1, other unknown `cd` targets** (`cd "$REPO"`): I recommend expanding only the home-directory forms and keeping fail-open for everything else. That fits the contract README ("catches agent mistakes, not a determined bypass"). The alternative is to deny git commands after any `cd` we can't resolve. Copilot's wording leans that way, but it would block legitimate commands.
2. **#3 scope:** should `GIT_DIR=… git reset` and `GIT_WORK_TREE=…` count too? `commandWords` currently throws away assignments before a command. It's the same kind of gap, and I'd include it.
3. **#2:** okay to reply on the PR instead of changing code, once `git -C/tmp status` confirms git rejects it? I'd draft the reply for you to approve before anything is posted.

After you answer, the work is the fixes and tests for #1, #3–#7 on `feat/hooks-go-git-guard` (in a worktree, not your clone), then `go test ./...`, updating the fixture count in `00-STATUS.md`, and a new `folkways review` before any push. That needs a session with edit and shell tools. I can write the fix plan into `efforts/EX-13-git-guard/` so the next session can start from it.
