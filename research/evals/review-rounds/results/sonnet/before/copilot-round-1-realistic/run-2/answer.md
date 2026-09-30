I read the review and the code it points at, but I haven't changed anything. Bash is disabled in this session, so I can't create the worktree, run `go test`, or check the clone's git state. The rules say code changes go in a worktree at `projects/.worktrees/folkways/<branch>`, not in your clone, so I didn't edit the files under `projects/folkways` directly.

The three findings hold up against the code. Here's the fix I'd make for each.

1. **High: global options bypass the guard** (`guard.go:66`).
   - `checkGit` only skips `-C` and `-c`, so `git --no-pager push` sets `sub` to `--no-pager` and gets through.
   - I'd skip all of Git's global flags. Flags that take a separate operand are `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--exec-path` and `--super-prefix`. Flags without one are `--no-pager`, `-p`/`--paginate`, `--bare`, `--no-replace-objects`, `--literal-pathspecs` and similar. The `--opt=value` forms also need handling.
   - `--git-dir` and `--work-tree` change the target repo, so I'd resolve `--work-tree` the way `-C` is resolved.
   - New tests: `--no-pager push` and `--no-pager reset` in a main clone.

2. **Medium: whole-file reads in the review scan** (`review.go:53`).
   - I'd replace `os.ReadFile` with a `bufio.Reader` that checks the first line for `---`, then reads lines up to the closing `---`, with a size cap. It returns immediately if the file doesn't start with frontmatter.
   - `WalkDir` currently keeps walking the second root after a match, so I'd stop it there too. `fs.SkipAll` only ends the current walk.
   - Review files anywhere under the roots still count, as the contract requires.

3. **Medium: heredoc bodies are split as commands** (`shell.go:113`).
   - `splitCommands` doesn't recognize a top-level `<<WORD`, so `cat <<EOF > notes.md\ngit push\nEOF` is checked as a real push.
   - I'd call the existing `skipHeredoc` from `splitCommands` for `<<` (but not `<<<`), so the body isn't scanned.
   - New lexer test: a top-level heredoc containing `git push`.

Should I proceed once Bash is available, or would you rather set up the worktree yourself and point me at it? I'd also add the Copilot fix commits to `00-STATUS.md`. Nothing gets pushed without your approval.
