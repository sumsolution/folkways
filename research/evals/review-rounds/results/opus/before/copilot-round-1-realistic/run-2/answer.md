I checked all three Copilot findings against the code, and each one is a real bug. I didn't change anything, because I can only read files in this session. I read the files as they are in `projects/folkways`, but I can't run git to confirm that the clone is on 75477a4.

## Findings

**1. Global options skip the guard: confirmed.** `guard.go:66` only skips over `-C` and `-c`. So `git --no-pager push` treats `--no-pager` as the subcommand, and the push check never runs. The same goes for `-p`, `-P`, `--git-dir=…` and `--bare`. The contract doesn't have to stop someone deliberately getting around it (`test/hooks/README.md:81`), but `--no-pager` is something agents actually type. No tests use global options.
- **Fix:** before reading the subcommand, skip every leading argument that starts with `-`. For the options that take their value as a separate word (`-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--config-env`), skip that value too.
- **Open point:** `--git-dir`/`--work-tree` also change which repo the command acts on, and nothing tracks that today. I'd leave it out and note it as a limit in the contract, unless you want it covered.

**2. Review scan reads whole files: confirmed. I'd call it Low, but it's cheap to fix.** `frontmatterSHA` (`review.go:54`) reads every file in full, and `hasFrontmatterReview` doesn't stop after a match. `fs.SkipAll` only ends the current walk, and the loop at `review.go:26` still walks the next root.
- **Fix:** read the file line by line and return right away if line 1 isn't `---`. Stop at the closing `---` or after a size limit.
- Also stop looping over the roots once a review is found.
- This also fixes a small bug: `strings.Cut(rest, "\n---")` accepts a `----` line as the end of the frontmatter.

**3. Heredoc bodies get parsed as commands: confirmed, and I'd fix it first.** `splitCommands` (`shell.go:113`) only skips heredocs inside `$(…)`. So `cat <<EOF > notes.md\ngit push\nEOF` gets checked as a real push. This blocks commands that shouldn't be blocked, and agents write files with heredocs all the time, so it will come up more often than #1. No tests cover a heredoc outside `$(…)`.
- **Fix:** calling `skipHeredoc` at `<<` isn't enough, because it would also skip the rest of that line (e.g. `cat <<EOF && git push`). Instead, collect the heredoc words while scanning the line. At the newline, split as normal, then skip each body in order.
- Tests to add: a top-level heredoc containing `git push`, a command after `<<EOF` on the same line, and two heredocs on one line.

## Proposed next steps

1. Create a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard` (not the main clone). Then fix and test in this order: #3, #1, #2. Each fix gets tests first, then a commit such as `fix(hooks): skip top-level heredoc bodies`.
2. Run the Go tests and the git-guard contract suite.
3. Run `folkways review` on the new HEAD, then ask you before pushing to PR #13.
4. Add these results to `00-STATUS.md`.

This session can't edit files or run commands, so steps 1–4 need a session with those tools. Two questions:
- The pasted review says "three of Copilot's findings". Are there more on the PR that I should look at?
- Should `--git-dir`/`--work-tree` repo tracking be part of #1, or noted as a limit?
