I haven't changed any code. The changes belong in a worktree, and I can't create one because Bash is disabled in this session. `projects/folkways` is your clone under `projects_dir`, so I won't edit it directly. I also couldn't run the Go tests.

I read the three findings against the code at `projects/folkways/hooks/internal/gitguard/`. All three are real, and each fix is small.

| # | Finding | Valid? | Fix | Size |
|---|---|---|---|---|
| 1 (High) | Global options before the subcommand | Yes. `checkGit` (`guard.go:66`) skips only `-C` and `-c`, so `git --no-pager push` sets `sub` to `--no-pager` and passes. | Extend that loop. Flags with an operand: `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--exec-path=…`. Flags with no operand: `--no-pager`, `-p`/`--paginate`, `--bare`, `--no-replace-objects`, `--literal-pathspecs`, and similar. Any `--x=y` form is skipped as one word. Add tests for `push` and `reset` with `--no-pager`. | ~15 lines plus tests |
| 2 (Medium) | `frontmatterSHA` reads whole files | Yes (`review.go:53`). | Open the file and read through a `bufio.Reader`. Return "" unless the first line is `---`. Stop at the closing `---`, with a cap of about 8 KB. The walk already stops after the first match in one root. It doesn't stop across roots, so I'd add a `found` check before each root. | ~20 lines |
| 3 (Medium) | Heredoc bodies split as commands | Yes. `splitCommands` (`shell.go:113`) has no top-level heredoc case, so the body lines are checked as commands. | Add a `<<` case, excluding `<<<`, that calls the existing `skipHeredoc`. One catch: the body starts on the next line, after the rest of the current command. So I'd record a pending heredoc word, then skip the body when the lexer reaches the next `\n`. Add a lexer test for `cat <<EOF > notes.md\ngit push\nEOF`. | ~20 lines plus tests |

Together that's roughly 60 lines of source plus tests. This is round 1, and the findings are in the original code rather than in an earlier fix. That means none of the step-back triggers apply. Fix 1 also stays within the contract's "agent mistakes, not a determined bypass" limit. I'd cover the common global flags and not try to mirror all of Git's option parsing.

Recommendation: fix all three, with one commit per finding.

To go ahead, I need one of these from you:
- **Bash back on:** I create `projects/.worktrees/folkways/feat/hooks-go-git-guard`, make the fixes, run `go test ./...`, and commit. I'd ask before pushing.
- **You make the worktree:** if you'd rather, I can make the edits there and you run the tests.

Once the fixes land, I'll update `00-STATUS.md`. Since the code changes, the pre-push review gate will also need a fresh `folkways review` on the new SHA before you push.
