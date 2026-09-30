I haven't changed anything. This round meets the repo's "step back" conditions, so I'm checking direction with you before fixing anything:
- Most of the findings land in code that round 1 added.
- Fixing all seven would be the second time in a row we fix every finding.
- A complete fix for #5 would need yet another piece of shell parsing on top of what we have.

## The numbers

- **Review rounds:** 2. Round 1 had 7 findings and we fixed all 7 in 8188289. Round 2 has 7 more (6 High, 1 Medium).
- **Size now:** the `hooks/internal/gitguard/` package is 780 lines of Go (`shell.go` 364, `guard.go` 322, `review.go` 94) plus 380 lines of tests. Fixtures grew from 51 to 63 in round 1.
- **Size at the plan:** I can't give it. The prototype at `bench/hooks/go-git-guard/` isn't in this checkout, and I have no shell here to diff against d76fb42. If you want that number, it's a `git diff --stat d76fb42 8188289` in the worktree.
- **Where the findings are:**

| # | Finding | Code it's in | Would an agent plausibly write this? |
|---|---|---|---|
| 1 | `cd ~/…` / `cd "$HOME/…"` ignored | Round 1: cd tracking (`guard.go:142` gives up on any target containing `$`, `` ` `` or `~`) | Yes, often |
| 2 | Attached `-C/path`, `-ck=v` | Round 1: global options (`guard.go:187-193`) | Rarely |
| 3 | `--git-dir` / `--work-tree` ignored when choosing the target repo | Round 1: global options (`guard.go:190`) | Rarely |
| 4 | `-m"msg"` with the message attached to the flag | Prototype (`guard.go:217`) | Sometimes |
| 5 | `$(git push)` inside a heredoc body | Round 1: heredocs as data (`shell.go:276-283`) | No, that's a deliberate bypass |
| 6 | Backslash-newline before the subcommand | Prototype lexer (`shell.go:323`, which keeps the newline as part of the next word) | Yes, in multi-line commands |
| 7 | `# comment; git push` wrongly denied | Prototype lexer (`shell.go:162`) | Sometimes, and it only blocks too much, never lets something through |

Four of the seven (1, 2, 3, 5) are in code round 1 added. The contract already limits git-guard's job: it catches agent mistakes, not a determined bypass (`test/hooks/README.md:87`). Finding 5 is clearly outside that. Findings 2 and 3 are forms an agent is unlikely to write.

## Options

1. **Do less (recommended).** Fix #1, #4 and #6, which are forms agents actually write. Each is a small, contained change: expand `~` and `$HOME` in `cdTarget`, read the attached `-m`/`-am` value, and drop backslash-newline in `words`. Probably under 40 lines plus a fixture each.
   - I'd also fix #2. It's a few lines in the same option loop, and it also closes the attached `-c` part.
   - Don't fix #3 or #5. Add one line to the contract README naming `--git-dir`/`--work-tree` and heredoc substitutions as out of scope, and answer Copilot on the PR with that reasoning.
   - Park #7 as a side quest, since it only denies too much.
   - Expect round 3 to raise more cases of the same kind. The contract line is what lets us say no to them.
2. **Fix all 7.** Findings 3 and 5 each add machinery: resolving `--git-dir`/`--work-tree` to a target repo, and a second scan of heredoc bodies that has to tell quoted delimiters from unquoted ones. Based on round 1, this likely leads to a round 3 against the new code.
3. **Change direction.** Replace the hand-written lexer with a real shell parser such as `mvdan.cc/sh`. That handles #5, #6 and #7 properly and removes most of `shell.go`. The costs: a second dependency (plan decision 1 allows only `yaml.v3`), a bigger binary, a rewrite partway through PR 2, and #1–#4 would still need their own fixes.

Should I go ahead with option 1? If so, I'll update `00-STATUS.md` with the decision first, then do the fixes test-first.
