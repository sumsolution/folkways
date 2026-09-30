I haven't changed any code. Two of the step-back triggers in `.folkways/instructions.md` apply, so I'm checking direction with you first.

**Why I stopped**
- **Round 1 code:** findings 1, 2, 3 and 5 land in code that round 1 added: `cd` tracking (`guard.go:132`), git global option parsing (`guard.go:173`) and heredoc stripping (`shell.go:246`). Findings 4, 6 and 7 are in older parsing.
- **Fix everything again:** the default here would be to fix all 7 findings for the second time.
- **Numbers:** I couldn't get diff stats because Bash is disabled in this session.
  - The two files are now `guard.go` at 322 lines and `shell.go` at 364.
  - Fixtures went from 51 (original contract) to 63 after round 1.
  - Review rounds so far: 2.
  - Round 1 gave 7 findings and round 2 gives 7 more. Each fix has widened the hand-written shell lexer, and Copilot finds the next uncovered shell form.

**Where the findings come from**
The guard is a partial reimplementation of bash parsing. The contract README already says it "catches agent mistakes, not a determined bypass". Most of these findings are determined-bypass forms:
- **Rare in agent output:** `-C/path` (2), `-ccore.foo=bar` (2), `--git-dir` with `--work-tree` and no `-C` (3), and `$(git push)` inside a heredoc body (5).
- **Plausible from an agent:**
  - `cd "$HOME/…" && git push` (1).
  - `-m"subject"` (4).
  - A backslash-newline continuation (6).
- **Harmless call wrongly denied:** `echo ok # note; git push` (7). This is a real usability bug.

**Options**
1. **Do less (recommended).** Fix the cheap, realistic ones: 4, 6 and 7, plus the attached `-C`/`-c` forms from 2. Each is a few lines in the existing parsers, plus a test.
   - For 1 and 3, fail safe: when a `cd` target or `--git-dir`/`--work-tree` can't be resolved, check the push against every plausible directory, or deny with a clear message. That avoids new resolution machinery.
   - Answer 5 on the PR as out of scope, and cite the README's stated limit.
   - This costs a third round if Copilot finds more forms, and it leaves 5 as a known gap.
2. **Fix all 7.** This closes every reported hole. It needs `$HOME`/`~` expansion, git-dir resolution, and a heredoc body scanner that tells quoted from unquoted delimiters. That's a new layer on code that already drew findings, and round 3 will likely turn up more.
3. **Change direction.** Stop growing the lexer and use a real shell parser (for example `mvdan.cc/sh`) to walk the AST. That fixes most of these classes at once. It costs a new dependency and a rewrite of `shell.go`, and it would delay PR 3.

I'd go with option 1. Which one do you want? If you pick 1 or 2, I'll add the failing tests first and then fix. I'd also update `00-STATUS.md` with the round 2 outcome before starting.
