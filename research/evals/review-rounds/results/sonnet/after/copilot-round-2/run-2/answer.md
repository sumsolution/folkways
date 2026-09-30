I haven't changed any code. This matches two of the "Step back" triggers in `.folkways/instructions.md`, so I'm checking direction with you first.

**Triggers met**
- **Findings land in code the previous round added.** Round 1 added the `cd` tracking, the global-option skipping, the heredoc handling and the subshell parsing. Round 2 findings 1, 2, 3, 5 and 6 are gaps in exactly those pieces. Finding 4 (`-m"…"`) and finding 7 (comments) are in older parsing code.
- **Second time recommending "fix everything".** Round 1 had 7 findings and we fixed all 7. Round 2 has 7 more (6 High, 1 Medium).

**Numbers**

| | Plan / round 1 | Now |
|---|---|---|
| git-guard fixtures | 51 (original contract) | 63 |
| `gitguard` package | not measured | about 1,160 lines across 6 files (`shell.go` 364, `guard.go` 322, tests 380) |
| Review rounds | | 2 (14 findings) |

I don't have the line count from the plan or from round 1, so I can't show growth there.

**The pattern in the findings**
All seven are the same kind of gap. We are re-implementing bash parsing (`$HOME` and `~`, attached options, `--git-dir`, heredoc substitutions, line continuations, comments) inside a hook. Each round of fixes adds parser surface, and Copilot then finds the next unhandled form. The contract README already says git-guard "catches agent mistakes, not a determined bypass." Most of these findings, such as `git -C/path push` and `cd "$HOME/…"`, are ordinary forms an agent could plausibly write, so they aren't adversarial.

**Options**
1. **Do less (my recommendation).** Fix the cheap, low-risk findings: 2 (attached `-C`/`-c`), 4 (attached `-m`), 6 (line continuations) and 7 (comments). For 1, 3 and 5, deny conservatively instead of resolving:
   - an unresolvable `cd` target, `--git-dir` or `--work-tree`: check the command against the strictest case, or deny with a message to use a plain path;
   - unquoted heredoc bodies containing `$(` or a backtick: inspect them, or deny.

   This adds little code and no new resolver. The cost is occasional false denies on unusual forms.
2. **Fix all seven properly.** This needs `$HOME` and `~` expansion, a git-dir and work-tree resolver, and a heredoc substitution scanner. It is roughly a third new layer of parser machinery, and I'd expect a round 3 of findings in it.
3. **Change direction.** Stop parsing shell. Check the effective repository at the git level (for example, run `git rev-parse` in the candidate directories) or move some of these checks to a different hook point. This is a bigger redesign, and it would need a plan update.

Which do you want? I'll start on option 1 once you confirm. Before I touch code, I'll also note the decision in `00-STATUS.md`.
