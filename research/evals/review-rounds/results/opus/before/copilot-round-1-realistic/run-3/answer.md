I checked all three findings against the code. All three are real bugs and worth fixing. None of them is covered by the current tests.

I read the files in the `projects/folkways` working tree. That checkout has no `.git`, so I can't confirm it matches 75477a4, the commit Copilot reviewed. I also can't edit files or run anything in this session, so nothing is changed yet.

## 1. `git --no-pager push` skips the review check (Copilot: High)

**Real bug; I'd call it Medium, but fix it.** `checkGit` only skips `-C` and `-c` before reading the subcommand (`hooks/internal/gitguard/guard.go:66`). With `git --no-pager push`, the subcommand is read as `--no-pager`, so neither the push check nor the main-clone check runs.

The contract only aims to catch agent mistakes, not deliberate bypasses (`test/hooks/README.md:81`). But agents do add `--no-pager` out of habit, so this is a mistake the hook should catch.

**Fix:**
- Skip any leading argument that starts with `-`.
- Also skip the value after options that take one as a separate word: `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--config-env`, `--super-prefix`.
- `--git-dir` and `--work-tree` also change which repo is the target. Either resolve them the way `-C` is resolved, or leave that out and say so.
- Add tests for `git --no-pager push` and `git --no-pager reset` in a main clone.

## 2. The review scan reads whole files (Medium)

**Real.** `frontmatterSHA` reads each whole file with `os.ReadFile`, even when it has no frontmatter (`hooks/internal/gitguard/review.go:54`). There's also a small inefficiency: `fs.SkipAll` only stops the current walk, so the loop still walks the next root after a match is found (`review.go:26-34`).

**Fix:**
- Open the file with a `bufio.Reader`. If the first line isn't `---` (with or without `\r`), return straight away.
- Otherwise read lines until the closing `---`, with a size limit (e.g. 64 KiB).
- Stop the loop over roots once `found` is true.
- The existing CRLF and no-frontmatter tests in `review_test.go` should keep passing. Add a test for a file with no frontmatter that is larger than the limit.

## 3. Heredoc bodies are checked as commands (Medium)

**Real, and likely to happen in this repo.** `splitCommands` only skips heredocs inside `$(...)`. At the top level it splits on every newline (`hooks/internal/gitguard/shell.go:135`). So `cat <<EOF > notes.md\ngit push\nEOF` treats `git push` as a real push and blocks the command. Agents here often write docs and status files with heredocs that mention git commands, so this will come up.

**Fix:**
- Add the case `skipSubst` already has to `splitCommands`: `<<` but not `<<<` calls `skipHeredoc`.
- Add a test in `shell_test.go` for a top-level heredoc containing `git push`.
- Add a test for `git commit -F - <<EOF … EOF && git push` to check that the push after the heredoc is still split out and checked.

## Plan

1. Make the fixes in a worktree on the PR branch, at `projects/.worktrees/folkways/feat/hooks-go-git-guard`, not in your clone.
2. Commit: `fix(hooks): skip git global options, stream review frontmatter, ignore heredoc bodies`. Or one commit per finding, if you want each one to map to a Copilot comment.
3. Run `go test ./...` and the contract suite with `FOLKWAYS_HOOK_CMD` set.
4. Run `folkways review` on the new commit, since the review gate needs it before the push. Then push, with your approval.
5. Update `00-STATUS.md`.

Should I go ahead? I'll need a session with shell and edit access to do it.
