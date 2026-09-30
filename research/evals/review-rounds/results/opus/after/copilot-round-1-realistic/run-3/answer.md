I've read the three findings against the code and checked each one. All three are real and small. None of them is in code added by an earlier review round, so this doesn't trigger a step-back. Line numbers are from the working tree in `projects/folkways`. I haven't confirmed that it's at 75477a4 yet; I'll pin the SHA when I set up the worktree.

**1. High: global options bypass the guard.** Confirmed at `hooks/internal/gitguard/guard.go:66`. Only `-C` and `-c` are skipped, so with `git --no-pager push` the guard reads `--no-pager` as the subcommand and lets the push through. The contract excludes deliberate bypasses (`test/hooks/README.md:81`), but agents do type `--no-pager` out of habit, so this is a mistake the guard is meant to catch.
- **Fix:** skip any leading `-…` argument. Also consume the operand of the few global options that take a separate one (`--git-dir`, `--work-tree`, `--namespace`, `--config-env`).
- **Size:** about 10 lines, plus tests for push and reset with `--no-pager`.
- I won't make `--git-dir` change which repo is checked. That would be more machinery than this problem needs.

**2. Medium: the review scan reads whole files.** Confirmed at `hooks/internal/gitguard/review.go:54`. I also found a related bug: `fs.SkipAll` (`review.go:30`) only stops the current walk, so after a match in `efforts/`, `people/<handle>/work/` is still walked.
- **Real risk:** low. Efforts are mostly markdown. But the `qa` skill saves screenshots into efforts, so large binary files will show up there.
- **Fix:** read line by line, stop if the first line isn't `---`, and stop at the closing `---`. Return as soon as a root has a match.
- **Size:** about 15 lines, plus a test for early exit.

**3. Medium: heredoc bodies are parsed as commands.** Confirmed at `hooks/internal/gitguard/shell.go:135`. The splitter treats every newline as a command separator, heredoc bodies included. Agents in this repo often write notes through `cat <<EOF`, and any body line starting with `git push` or `git checkout` gets denied by mistake.
- **Fix:** I won't just call the existing `skipHeredoc` from `splitCommands`. It would also skip the rest of the `<<EOF` line, so `cat <<EOF && git push` would slip through. Instead, remember the pending heredoc delimiter and skip its body at the next top-level newline.
- **Size:** about 15 lines, plus lexer tests for a heredoc containing `git push` and for `<<EOF && git push`.

**Recommendation:** fix all three. That's about 40 lines of code plus tests.
- **Smaller option:** fix 1 and 3 only, and park 2 as a side quest in `people/alex/inbox/`. That saves about 15 lines. The cost is that the hook's read time grows as screenshots pile up in efforts.

To do the work, I'd create a worktree for the existing `feat/hooks-go-git-guard` branch at `projects/.worktrees/folkways/feat/hooks-go-git-guard`. There's no worktree yet, and I won't touch your clone. Then I'd make the fixes, run the Go tests and the git-guard contract suite, and make one commit: `fix(git-guard): handle global options, heredoc bodies and large review scans`. I'll ask before pushing.

Should I go ahead with all three, or with the smaller option?
