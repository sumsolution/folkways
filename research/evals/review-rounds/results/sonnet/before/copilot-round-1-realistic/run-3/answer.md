I read all three findings against the code. They're all valid, but I haven't changed anything.

The code is in `projects/folkways`, which is your clone under `projects_dir`. The repo rules say code changes go in a worktree on the PR branch, `feat/hooks-go-git-guard`, and never in the clone itself. Creating a worktree needs `git`, and Bash is disabled in this session. I can't create the worktree, run the Go tests, or check what the clone has checked out. Editing the clone directly would break the rule, so I stopped.

## What each fix involves

**1. High, global options (`guard.go:66`).** Confirmed. `checkGit` only skips `-C` and `-c`, so `git --no-pager push` sets `sub` to `--no-pager` and passes.
- Skip flags that take no operand: `--no-pager`, `-p`/`--paginate`, `--bare`, `--no-replace-objects`, `--literal-pathspecs`, `--no-optional-locks`, and similar.
- Skip options that take an operand, either as `--opt=value` or as `--opt value`: `--git-dir`, `--work-tree`, `--namespace`, `--exec-path`, `--super-prefix`, `--config-env`, `--attr-source`.
- Treat `--git-dir` and `--work-tree` like `-C` when working out which directory to check.
- Add table tests for `git --no-pager push` and `git --no-pager reset` in a main clone.

**2. Medium, whole-file reads (`review.go:28`).** Confirmed on both points.
- `frontmatterSHA` calls `os.ReadFile` on every file. I'd switch it to a `bufio.Reader` that checks the first line for `---` and returns straight away if it isn't there. It would stop at the closing `---` and cap the bytes read. I'd also skip non-regular files.
- `fs.SkipAll` only ends the current `WalkDir`. The loop over `roots` then walks the next root anyway. It needs an early return once a match is found.
- Review files stay valid anywhere under the roots, as the contract says.

**3. Medium, heredoc bodies (`shell.go:137`).** Confirmed.
- `splitCommands` splits on every `\n`, so `cat <<EOF > notes.md\ngit push\nEOF` yields a `git push` segment.
- Reusing `skipHeredoc` at `<<` in `splitCommands` would be wrong. It jumps to the terminator and drops the rest of the line, so `cat <<EOF && git push` would hide the push.
- Instead, record the pending delimiter when `<<` is seen. Keep lexing the rest of that line as normal. At the next newline, skip to the terminator line.
- Add a lexer test with a top-level heredoc containing `git push`. Add a second test that `<<EOF && git push` still catches the push.

## Decision for you
Do you want me to apply these in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`? That needs a session where Bash is enabled. If you'd rather, I can write the plan above into `efforts/EX-13-git-guard/` as a topic file and add it to `00-STATUS.md`. I won't commit or push anything without asking you first.
