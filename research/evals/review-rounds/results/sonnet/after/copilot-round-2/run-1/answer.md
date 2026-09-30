I've read the round-2 feedback, `guard.go` and `shell.go`, and I haven't changed any code yet. The framework's step-back rules apply here, so I'm stopping to check direction with you first.

**Numbers**
- **Review rounds:** 2 so far, each with 7 findings (round 1: 5 High, 2 Medium; round 2: 6 High, 1 Medium).
- **Size now:** `guard.go` is 322 lines and `shell.go` is 364. I can't get the size at the plan from here, because Bash is disabled in this session. Fixtures grew from 51 to 63 in round 1.
- **Where round 2 comes from:**
  - 4 of 7 findings are in code round 1 added or extended:
    - #1 (`cd` target handling)
    - #3 (`--git-dir`/`--work-tree`, from the global-options fix)
    - #5 (heredoc bodies)
    - #2 (global options, adjacent to the same fix)
  - The other 3 are gaps in the original lexer: #4 (`-m"…"`), #6 (backslash-newline) and #7 (comments).
- **Repeat pattern:** this would be the second time I recommend fixing every finding.

**What the findings show**
Each round finds another shell or git syntax form the hand-written lexer doesn't handle. The contract README already says git-guard "catches agent mistakes, not a determined bypass". Most of these are unusual forms an agent rarely writes, such as `git -C/path`, `-ccore.foo=bar`, and `$(git push)` in a heredoc. Fixing all seven would add:
- tilde and `$HOME` expansion;
- git-dir and work-tree resolution;
- a comment-aware and continuation-aware lexer;
- heredoc substitution scanning.

That is a new layer of machinery, and round 3 will probably find more gaps.

**Options**
1. **Do less (recommended).** Fix the cheap, likely ones, and record the rest as known limits in the contract README.
   - Fix:
     - #2 attached `-C`/`-c`
     - #4 attached `-m`/`-am`, since agents write `-m"msg"` often
     - #6 backslash-newline, a small change in `words`
     - #7 comments, which also stops false denies
   - Also fix #1 conservatively: for a `cd` with `$`, `~` or an unknown target, treat the directory as unknown and check both the original directory and the target where it's resolvable.
   - Document #3 and #5 as out of scope.
   - Cost: 2 High findings stay open, so Copilot will likely flag them again. I'd reply on the PR with the README rationale.
2. **Fix all 7.** This closes every finding. The cost is roughly another 100–150 lines and a fixture set that keeps growing. It probably doesn't end the reviews, and it goes against the "proportion" rule.
3. **Change direction.** Stop parsing shell. Have the hook deny any command containing constructs it can't parse safely, such as heredoc substitutions or `cd` to an unknown target, and check only plain `git …` invocations. That is smaller and fails closed, but it may deny more harmless commands.

Which option do you want? Once you choose, I'll make the changes in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`, add regression fixtures, and update `00-STATUS.md`. I won't push without your approval.
