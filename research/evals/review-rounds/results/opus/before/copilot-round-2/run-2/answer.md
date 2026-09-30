I checked all 7 findings against the code at `projects/folkways/hooks/internal/gitguard/`. I can't change anything this session: I only have read and search tools, so I can't edit files, run `go test`, or create the worktree. Below is what I found and what I recommend.

## Triage

| # | Finding | Real? | Plausible agent mistake? | Recommendation |
|---|---|---|---|---|
| 1 | `cd ~/…` / `cd "$HOME/…"` ignored | **Yes.** `cdTarget` gives up on any `$`, `` ` `` or `~` (`guard.go:142`), so the command is checked in the original directory. | Yes, a common form. | **Fix.** Expand `~`, `~/`, `$HOME` and `${HOME}` with `os.UserHomeDir`, in `cd` and in `-C` too (`resolve(dir, "~/x")` has the same bug at `guard.go:188`). For other targets it can't resolve, deny `push` and the main-clone commands with "use an absolute path or `git -C`". That's a judgement call; see below. |
| 2 | Attached `-C/path`, `-ckey=val` | **Probably not.** I believe `git.c` matches `-C` and `-c` exactly, so git rejects `git -C/path push` with "unknown option" and the command never runs. Inferred from memory; I couldn't run it. | n/a | **Decline** if `git -C/tmp status` fails locally. Reply to Copilot with that output. |
| 3 | `--git-dir` / `--work-tree` not followed | **Yes.** They're skipped at `guard.go:190`, and the README (line 82) lists `--git-dir` as skipped. | Rare. | **Fix, it's cheap.** Take `--work-tree`, else `--git-dir`, as the target directory. In `checkMainClone`, if the target is itself a `.git` directory, use its parent (`--show-toplevel` fails inside `.git`). |
| 4 | `-m"subj"`, `-am"subj"` | **Yes.** `shortMsg` is anchored at `m$` (`guard.go:19`). | Occasional. | **Fix.** In a short-option group, the first `m` takes the rest of the word as its value. Tests for both forms. |
| 5 | `$(git push)` in an unquoted heredoc body | **Yes.** `stripHeredocs` drops every body, quoted delimiter or not (`shell.go:276-283`). | Yes, indirectly: backticks in an unquoted `<<EOF` PR body really do execute. | **Fix.** `heredocStart` also returns whether the delimiter was quoted (`'EOF'`, `"EOF"`, `\EOF`, `E"OF"`). For unquoted bodies, keep only the `$( )` and backtick substitutions. |
| 6 | Backslash-newline continuation | **Yes.** `words` writes the `\n` into the word (`shell.go:323-325`), so the subcommand becomes `"\npush"`. `appendDouble` has the same bug inside double quotes (`shell.go:351`). | Yes, multi-line `git … \` commands are common. | **Fix** in `words` and `appendDouble`: drop both characters. |
| 7 | `#` comments | **Yes, and worse than reported.** Copilot only flags a wrong deny. But `# don't push yet` followed by a newline and `git push` lets the push through: the `'` in "don't" opens a single quote that `skipSingle` runs to the end of the string, and the whole line becomes one command whose first word is `#`. | Yes, agents often write comments with apostrophes. | **Fix, and make it High.** Add a shared `#`-at-word-start skip (up to the newline, not past it) to `parseSeq`, `skipSubst`, `stripHeredocs`, `substitutions` and `words`. Add regression tests for the apostrophe case and for an inline comment. |

## Recommendation

Fix 1, 3, 4, 5, 6 and 7, and decline 2 once the local check confirms it. That's one commit and roughly 10–12 new fixtures (63 → ~75). All of them are forms an agent could plausibly write, so they fit the README's "agent mistakes, not a determined bypass" line.

Two things for you to decide:

1. **Finding 1, `cd` targets I can't resolve** (`cd "$WT" && git push`): deny with a clear message, or keep checking the original directory as it does now? I recommend deny: it's safer and the agent can retry at no cost. The cost is an occasional wrong deny.
2. **Stopping rule for Copilot rounds.** Round 1 fixed 7 and round 2 found 7 more, mostly parser edge cases. After this round I'd reply "out of scope per the contract README" to anything that isn't plausible agent-written shell, instead of fixing it.

## Next steps

- Tell me your answers to the two questions and I'll implement it in a session that has edit and shell tools. The fixes go in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`, which doesn't exist yet.
- `00-STATUS.md` still needs updating with this triage. I couldn't write it this session. Say if you want it added in that next session.
