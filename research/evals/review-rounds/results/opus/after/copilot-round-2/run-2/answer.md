I'm stopping to check direction with you before I fix anything. This round meets the step-back triggers: most of the new findings are in code that round 1 added, and fixing all of them would be the second fix-everything round in a row.

## The numbers

- **Plan for PR 2** ([plan 11 §2](projects/folkways/docs/design/11-plan-hooks-build.md)): move the prototype, add the trigger floor and the inbox lookup, and write unit tests. The prototype passed 51 contract cases.
- **Now:** `gitguard/` has 780 lines of code (`shell.go` 364, `guard.go` 322, `review.go` 94) and 380 lines of tests. Fixtures went from 51 to 63. The prototype isn't in this checkout, so I can't give its size for comparison.
- **Review rounds:** 2. Round 1 had 7 findings and we fixed all 7. Round 2 has 7 more.
- **Where round 2 lands:** 5 of the 7 are in code round 1 added:
  - #1 is in the `cd` tracking (`guard.go:142`).
  - #2 and #3 are in the global-option handling (`guard.go:187-190`).
  - #5 is in the heredoc stripping (`shell.go:279`).
  - #6 is in the lexer's word splitting (`shell.go:323`).
  - #4 (`-m"msg"` with the message attached) and #7 (comments) are gaps in the original lexer.

I read the working tree, not a pinned SHA, because I have no shell here. The cited lines match what Copilot describes.

## Sorting against our own scope

The contract README says git-guard catches agent mistakes, not a determined bypass. Against that bar:

| # | Command form | Would an agent write it? | Fix cost |
|---|---|---|---|
| 1 | `cd "$HOME/…" && git push`, `cd ~/…` | Yes, often | Small: expand `$HOME` and `~` in `cdTarget` |
| 4 | `git commit -m"msg"` | Sometimes | Small: one regex case |
| 6 | `git \` newline `push` | Rarely, but it costs almost nothing | Small: drop backslash-newline in `words` |
| 7 | `echo ok # note; git push` wrongly denied | Comments are common, but the harm is only a false deny | Medium: comment handling in `parseSeq` |
| 2 | `git -C/path push`, `-ckey=val` | Rarely | Small |
| 3 | `--git-dir` / `--work-tree` instead of `-C` | Rarely | Medium: resolving the git target |
| 5 | `$(git push)` inside an unquoted heredoc | No, that's a deliberate bypass | Large: parsing substitutions inside heredoc bodies |

## Options

- **A. Do less (recommended).** Fix #1, #4 and #6, and #2 because it's a few lines. I estimate about 40–60 lines plus about 6 fixtures. Reply on the PR that #3, #5 and #7 are outside the "agent mistakes" scope (#7 only causes false denies), and park them as side quests. Then move on to PR 3. The risk is that Copilot flags similar forms in a third round, and we answer those from the same scope line.
- **B. Fix all 7.** I estimate 150–250 more lines. #3 and #5 need new machinery: resolving the git target, and parsing substitutions inside heredoc bodies. Based on the last two rounds, a third round would most likely find problems in that new code.
- **C. Change direction.** Stop growing the parser, and deny whenever the guard can't be sure: an unknown `cd` target, an unrecognized git global option, or `$(` in a heredoc. That's less code than B and closes the bypasses, but agents will hit more false denies.

I recommend A. It covers the forms agents actually write, it's small, and it keeps the PR close to its plan.

Which option do you want? Once you pick, I'll record the decision and the per-finding sorting in `00-STATUS.md`.
