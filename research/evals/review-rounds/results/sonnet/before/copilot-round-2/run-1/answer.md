I haven't changed anything. This session has no Bash tool, and I also have no edit or write tool. I can read files but not change them, and I can't run `go test`. Below is what I found reading the code, so the fixes can go in quickly once edits are possible.

Two things to settle first:
- **Where the edits go.** The guard code is in `projects/folkways`. Under the project-repo rules, code changes belong in a worktree on the PR branch (`feat/hooks-go-git-guard`) at `<projects_dir>/.worktrees/folkways/<branch>`, not in the clone.
- **Pushing.** Any push needs your approval.

I read all seven findings against `guard.go` and `shell.go`. Each one holds up. The plan for each:

1. **`cd` with `$HOME` or `~`** (`guard.go:142`)
   - `cdTarget` returns not-a-cd when the target has `$`, a backtick or `~`. `checkSeq` then treats the command as non-moving, so the next `git` runs against the old directory.
   - Expand `~`, `~/x`, `$HOME` and `${HOME}` from `os.UserHomeDir`.
   - For any other unknown target (`$FOO`, `-`), add an "unknown" directory state. Deny `push` and the main-clone commands there instead of falling back to the old directory.

2. **Attached `-C` and `-c`** (`guard.go:187`)
   - In `checkGit`, handle `-C<dir>` (update `dir`) and `-c<k=v>` (skip it) when the word has a `-C` or `-c` prefix and is longer than two characters.

3. **`--git-dir` and `--work-tree`** (`guard.go:190`)
   - Track both options in `checkGit`, in the `--opt=value` and `--opt value` forms.
   - Use `--work-tree` as the effective directory. With only `--git-dir`, use its parent, or `--git-dir` itself when it is bare.
   - Resolve both relative to the current `-C` directory.

4. **Attached `-m` and `-am` values** (`guard.go:217`)
   - `commitMessage` only matches `-m` and `-am` as separate words.
   - Add a case for `^-[a-zA-Z]*m(.+)$`, which returns the attached tail. Check that it doesn't misfire on other short-option clusters.

5. **Command substitutions in unquoted heredocs** (`shell.go:279`)
   - `stripHeredocs` currently drops every body. Record whether the delimiter was quoted.
   - For an unquoted delimiter, keep the body's `$(…)` and backtick substitutions, for example by appending them to the output as separate commands. Drop the plain text.
   - `heredocStart` strips the quotes, so it needs to report whether they were there.

6. **Backslash-newline** (`shell.go:323`)
   - `words` writes the newline after a backslash into the word.
   - Drop `\<newline>` in `words` and `appendDouble`. Bash removes it in both places.
   - `parseSeq` already skips the pair, so the two stay consistent. `cdTarget` picks up the fix through `words`.

7. **Comments** (`shell.go:162`)
   - In `parseSeq`, treat an unquoted `#` at a token boundary as a comment and skip to the next newline.
   - A token boundary means start of input, whitespace, or one of `;&|()`. That leaves `a#b` and `$#` alone.
   - `words` needs the same rule so a trailing comment isn't read as arguments.

Each fix gets a test in `guard_test.go` or `shell_test.go`, and I'd add fixtures to the contract set (currently 63). Findings 1 and 3 are the riskiest. They change how directories are tracked, so I'd do them together with the tests first.

If you turn on Bash and edit access, I'll set up the worktree and do this test-first. Do you want the fixes in one commit or one per finding? After that I'd update `00-STATUS.md` and rerun `folkways review` on the new SHA before you push.
