I haven't changed anything. There's no worktree for `feat/hooks-go-git-guard`, and I can't create one or edit files with the tools I have here. I read the code in your clone at `projects/folkways`, and the rules say not to change it there.

**Findings and proposed fixes** (all three hold up against the code):

1. **High, global options (`guard.go:66`).** Confirmed. The loop only skips `-C` and `-c`, so `git --no-pager push` sets `sub` to `--no-pager` and passes. I'd skip the no-operand flags (`--no-pager`, `-p`, `--paginate`, `--bare`, `--no-replace-objects`, `--literal-pathspecs` and similar) and the flags that take an operand (`-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--exec-path`). The `--opt=value` forms would be skipped as a single word. I'd treat `--work-tree` and `--git-dir` like `-C` when picking the directory. I'd add tests for `push` and `reset` with `--no-pager`.
2. **Medium, reading whole files (`review.go:53`).** Confirmed: `os.ReadFile` runs on every file in the walk. I'd open the file and use a `bufio.Reader` to check for `---\n` first, then read lines until the closing `---`, with a size cap on the block. Files with no frontmatter, such as binaries, return after a few bytes. The walk already stops at the first match within a root, but not across roots. I'd return early from the loop once `found` is set.
3. **Medium, heredoc bodies (`shell.go:137`).** Confirmed. `splitCommands` only skips heredocs inside `$(...)`, so a top-level `cat <<EOF > f\ngit push\nEOF` gets split at the newlines. I'd handle `<<` (but not `<<<`) in `splitCommands` with the existing `skipHeredoc`, and add a lexer test with `git push` in the body. One catch: `skipHeredoc` returns the index just after the closing word, but the body starts on the next line. The rest of the opening line, such as `> notes.md`, still has to stay part of the current command. I'd handle that by recording the pending heredoc and skipping the body when the lexer reaches the newline.

This is round 1 and the fixes are small, so none of the step-back triggers apply. I expect about 60–80 lines of Go and tests.

**Before I can start:**
- I need a shell in this session to create the worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`. Alternatively, you can create it yourself.
- I'll commit locally, but I'll ask before any push. Per the review gate, the new SHA also needs a `folkways review` before PR #13 is updated.

Do you want all three fixed, or only the High one and the heredoc one? The file-read fix is the most invasive of the three.
