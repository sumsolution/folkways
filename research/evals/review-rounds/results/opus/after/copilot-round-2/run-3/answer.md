I'm stopping to check direction before fixing anything. This round meets the repo's step-back rule: most of the findings are in code that round 1 added, and fixing all of them would be the second time in a row we fix every finding.

## The numbers

- **Review rounds:** round 1 had 7 findings, and all 7 were fixed in 8188289. Round 2 has 7 findings (6 High, 1 Medium).
- **Where round 2 lands:**
  - 5 of the 7 are in round 1 fixes: #1 `cd` tracking, #2 and #3 git global options, #5 heredocs treated as data, and #6 partly (the `cd` targets).
  - 2 are in the original parser from the prototype: #4 commit message parsing and #7 command splitting.
- **Size now:** `hooks/internal/gitguard` is about 780 lines of code (`shell.go` 364, `guard.go` 322, `review.go` 94) plus about 380 lines of tests. Contract fixtures went from 51 to 63.
- **Size at plan time:** I can't give this. The prototype (`bench/hooks/go-git-guard/`) isn't in this checkout, and I have no shell here to check the git history.
- **Scope:** the contract says git-guard "catches agent mistakes, not a determined bypass" (`test/hooks/README.md:87`).

## The findings sorted by that scope

| # | Finding | Would an agent write this by mistake? | Cost |
|---|---|---|---|
| 6 | `\`-newline line continuations | Yes, often | Small: `words` and `parseSeq` in `shell.go` |
| 7 | `#` comments cause false denies | Yes | Small, in `parseSeq` |
| 1 | `cd ~/…` or `cd "$HOME/…"` | Yes | Small: expand only `~` and `$HOME` in `cdTarget` (`guard.go:142`) |
| 4 | `-m"msg"` with the message attached | Sometimes | Small, in `commitMessage` |
| 2 | `-C/path` or `-ck=v` with no space | Rarely | Small, two cases in `checkGit` |
| 3 | `--git-dir` / `--work-tree` | Rarely, and it's deliberate | Medium: needs rules for how the options combine with `-C` |
| 5 | `$(git push)` inside a heredoc body | No, this is a bypass | Medium: a new scan of heredoc bodies |

## Options

- **A (recommended): fix #6, #7, #1, #4 and #2; decline #3 and #5.** In the PR, reply to #3 and #5 citing the README scope line, and add both to that line as examples. This is about 5 small changes with a regression fixture each. Most of the risk of another round comes from #3 and #5, since both add new machinery.
- **B (do less): fix only #6 and #7.** These are the parser bugs agents actually hit. Decline the rest as out of scope.
- **C: fix all 7.** This adds `--git-dir`/`--work-tree` resolution and heredoc-body scanning on top of a hand-written lexer that already needed fixes this round. Expect a round 3 with findings in the new code.
- **D (change direction): replace the hand-written lexer in `shell.go` with `mvdan.cc/sh/v3/syntax`, a full bash parser.** It would fix #5, #6 and #7 and the next round of lexer findings structurally. The cost is a new dependency, a larger binary, and redoing the lexer part of PR 2. We would also need to re-check the 3 ms budget (plan 11, the risks section). It's worth it only if more hooks will parse shell commands.

I haven't edited anything yet. Which option do you want? After you choose, I'll update `00-STATUS.md` with the decision.
