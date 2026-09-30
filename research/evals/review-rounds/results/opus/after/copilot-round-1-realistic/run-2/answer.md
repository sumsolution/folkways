I checked all three Copilot findings against the code, and all three are real. My recommendation is to fix them all this round. Each fix is small: about 40–50 lines of Go altogether, plus tests.

I read the code in `projects/folkways`, which is checked out locally. I have no shell in this session, so I couldn't confirm that checkout is at 75477a4.

## 1. High: global options skip the guard — fix

- **Confirmed:** at `hooks/internal/gitguard/guard.go:66`, only `-C` and `-c` are skipped before the subcommand is read. So `git --no-pager push` gets past the review check, and `git --no-pager reset` in a main clone gets past the clone guard.
- **Why fix it:** the contract says the guard catches agent mistakes, not a determined bypass (`test/hooks/README.md:81`). But agents type `--no-pager` all the time, so this is a mistake case, not a deliberate bypass.
- **Fix:** skip every leading `-…` argument. Options that take a separate value (`-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`) also skip that value. About 10 lines.
- **Tests:** `git --no-pager push` with no review is denied; `git --no-pager reset` in `${MAIN}` is denied. No test covers this today.

## 2. Medium: the review scan reads whole files — fix, keeping it small

- **Confirmed:** `frontmatterSHA` calls `os.ReadFile` on every file (`review.go:54`). This matters because the `qa` skill saves screenshots into effort folders.
- **Also confirmed:** `fs.SkipAll` only stops the current walk. The loop still walks `people/<handle>/work` after a match in `efforts/` (`review.go:26`).
- **Fix:** read line by line with `bufio` and stop if the first line isn't `---`. Stop at the closing `---` or after a line cap. Add `if found { break }` to the loop. About 15 lines, and review files anywhere under those folders still count.
- **Smaller option:** only add the `break` and skip files that aren't `.md`. That's fewer lines, but it adds a filename rule the contract doesn't have.

## 3. Medium: heredoc bodies are parsed as commands — fix

- **Confirmed:** `splitCommands` treats every newline as the end of a command (`shell.go:135`). So an agent writing a status doc with `cat <<EOF > 00-STATUS.md` whose body has a line starting `git push` gets a false deny. Heredocs are only skipped inside `$(...)` today (`shell.go:45`).
- **Catch:** reusing `skipHeredoc` in the top-level parser would jump over the rest of the first line. In `cat <<EOF > f && git push` that would hide a real push. The fix needs to note the heredoc word at `<<`, and skip the body only at the next newline. About 15–20 lines.
- **Tests:** a heredoc with `git push` in its body is not flagged as a push; `cat <<EOF > f && git push` still is.

## Next

All the fixes go on `feat/hooks-go-git-guard` in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`, not in your clone. After that: run the tests and the contract suite, then do one commit per finding, e.g. `fix(git-guard): skip global options before the subcommand`. Then a `folkways review` on the new HEAD, and I'll ask you before pushing.

I can't create the worktree, run `go test` or commit from this session because I have no shell here. You can either run it from a session that has one, or I can write the patches as files for you to apply.

Should I go ahead with all three? The smaller option is to fix only #1 now and park #2 and #3 as side-quest notes in the effort. That saves about 35 lines now, but leaves the heredoc false deny and the whole-file reads in place.
