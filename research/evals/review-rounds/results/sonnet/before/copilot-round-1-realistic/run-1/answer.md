I've read the review and the code, but I haven't changed anything. This session only has read and search tools (no Bash, Edit or Write), so I can't edit files, run `go test` or commit. All three findings hold up against the code.

**1. High: global options bypass `checkGit` (`guard.go:66`)**
- The loop only consumes `-C` and `-c`. For `git --no-pager push`, `sub` becomes `--no-pager`, so the push and main-clone checks never run.
- Fix: replace the loop with a small option parser.
  - Flags with no operand: `--no-pager`, `-p`, `--paginate`, `-P`, `--bare`, `--no-replace-objects`, `--literal-pathspecs`, `--glob-pathspecs`, `--noglob-pathspecs`, `--icase-pathspecs`, `--no-optional-locks`, `--no-advice`, and similar.
  - Flags that take a separate operand: `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--config-env`, `--super-prefix`, `--attr-source`. The `--flag=value` form takes no extra word.
  - Only `-C` changes `dir`.
  - Treat any other leading `-` word as a flag with no operand, so a flag I didn't list fails toward checking.
- Tests: `git --no-pager push`, `git --no-pager reset` in a main clone, `git -P -C dir push`, and `git --git-dir=x push`.

**2. Medium: whole-file reads in the review scan (`review.go:28`, `:53`)**
- `frontmatterSHA` calls `os.ReadFile` on every file under the roots, including binaries.
- Fix in `frontmatterSHA`:
  - Open the file and wrap it in a `bufio.Reader`.
  - Check that the first line is `---`, and return `""` right away if it isn't.
  - Read lines until the closing `---`, capped at about 64 KB.
  - Unmarshal only that block.
  - Handle CRLF per line.
- Fix in `hasFrontmatterReview`:
  - `fs.SkipAll` only ends the current root's walk, so add `if found { break }` in the roots loop.
  - Skip non-regular files with `d.Type().IsRegular()`.
  - Keep matching anywhere under the roots, as the contract allows.
- Tests: a large binary file with no frontmatter next to a valid review, and a frontmatter block with no closing `---`.

**3. Medium: heredoc bodies parsed as commands (`shell.go:137`)**
- `splitCommands` splits on every newline, so `cat <<EOF > notes.md\ngit push\nEOF` yields a `git push` segment. The `skipHeredoc` helper already exists but is only used inside `$(...)`.
- Fix: add a `<<` case to `splitCommands`, excluding `<<<`.
  - Parse the word with `heredocStart` and queue it.
  - Keep lexing the rest of that line, because `<<EOF > f && git push` still contains a real push.
  - When the next top-level newline arrives, jump past the body to the terminator line instead of flushing there. This needs a queue because several heredocs can share one line.
- Tests in `shell_test.go`:
  - The heredoc with `git push` in the body yields only the `cat` segment.
  - `cat <<EOF && git push\n...\nEOF` still yields both commands.
  - A quoted delimiter (`<<'EOF'`) and `<<-EOF`.
  - `<<<` here-strings are unaffected.

**Before I can implement these:**
- Send me a session with write access, or enable Edit/Bash here.
- `AGENTS.md` says code changes go in a worktree at `<projects_dir>/.worktrees/folkways/feat/hooks-go-git-guard`, not in the clone at `projects/folkways`. I'd set that up first.
- I'd commit as `fix(hooks): …`, probably one commit per finding, and ask before pushing.
- I'd also update `00-STATUS.md` to list the three fixes.

Should I start with the High one?
