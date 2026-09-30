I've read the three findings next to the code and the git-guard contract. All three are real. Here's what I recommend for each, sized to the contract's aim (`test/hooks/README.md:81`: "catches agent mistakes, not a determined bypass").

One caveat: `projects/folkways` has no `.git` in this checkout, so I couldn't confirm the files I read are at 75477a4. The line numbers Copilot cites do match.

## 1. High: `git --no-pager push` gets past the review gate — fix it

`guard.go:66` only skips `-C` and `-c`. After that it treats the next word as the subcommand, so `--no-pager` is taken for the subcommand and the push check never runs. Agents do add `--no-pager`, so this counts as a mistake the hook should catch, not a deliberate bypass.

- **Fix:** skip any argument starting with `-`. For `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace` and `--config-env`, also skip the value that follows. `-C` keeps changing the target directory as it does now.
- **Size:** about 10 lines, plus two tests: `git --no-pager push` with no review, and `git --no-pager reset` in a main clone.
- **Not doing:** full handling of `--git-dir`/`--work-tree` (changing which repo we check). It's rare, and it drifts toward the bypass cases the contract excludes.

## 3. Medium: heredoc bodies are read as commands — fix it

`splitCommands` (`shell.go:113`) only handles heredocs inside `$(...)` (`skipSubst`, `shell.go:45`). At the top level it splits on every newline, so writing a note with `cat <<EOF > notes.md` that contains `git push` or `git commit -m "wip"` gets the tool call denied. Agents in this repo write docs this way all the time, so this false positive will actually happen.

- **Watch out:** calling `skipHeredoc` directly would also swallow the rest of the `<<EOF` line. In `cat <<EOF > f && git push`, the real push would then go unchecked.
- **Fix:** on `<<` (but not `<<<`), note the delimiter word and keep scanning the line. At the next newline, end the segment and skip the body up to the closing delimiter.
- **Size:** about 15–20 lines, plus two lexer tests: a heredoc body containing `git push` (ignored), and `<<EOF … && git push` on the opening line (still split out).

## 2. Medium: the review scan reads whole files — do the small version

`frontmatterSHA` (`review.go:54`) calls `os.ReadFile` on every file under `efforts/` and `people/<handle>/work/`. Most are small markdown files, but QA screenshots do end up there, so the cost is real but modest. There's also a small bug: `fs.SkipAll` only stops the current root, so the second root gets walked even after a match (`review.go:26-34`).

- **Smaller option (recommended):** stop the root loop once a match is found. Open each file with a buffered reader and give up unless the first line is `---`. Otherwise read only up to the closing `---`. About 15 lines, one test for a file with no frontmatter, and no change to which files count as a review.
- **Bigger option:** add a size cap or skip binary files by extension. It adds rules the contract doesn't state, for little gain. Not worth it.

## Totals

About 45 lines of code and 5 tests. All of it is in code from d76fb42, not in fixes from an earlier round, so there's no reason to step back yet.

**Options:**
1. **Fix all three as above** (recommended).
2. **Do less:** fix 1 and 3 now and park 2 in the effort as a side-quest note. That saves about 15 lines and leaves a slow-but-working scan.

I don't have edit or shell tools in this session, so I haven't changed anything, including `00-STATUS.md`. If you pick option 1, I'll need a session with those tools. The work would go in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`, with `go test ./...`, a commit, a fresh `folkways review` on the new SHA, and then your approval before any push.
