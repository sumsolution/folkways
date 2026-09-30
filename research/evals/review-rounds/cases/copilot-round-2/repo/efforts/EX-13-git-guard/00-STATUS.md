# EX-13: git-guard hook in Go

**Repo:** `projects/folkways` (branch `feat/hooks-go-git-guard`, PR #13, draft)
**Plan:** [plan 11, PR 2](../../projects/folkways/docs/design/11-plan-hooks-build.md). Contract: `projects/folkways/test/hooks/README.md`.

## Where we are

- d76fb42: Go module with `git-guard`, moved from the benchmark prototype. Passes the 51 original git-guard contract cases.
- Copilot review round 1 on 75477a4: 7 findings (5 High, 2 Medium). Fixed all 7 in 8188289:
  - failed `cd` (`cd /missing || git push`): track every directory a command could run in
  - `( … )` subshells and `{ }` groups
  - git global options before the subcommand (`--no-pager`, `--git-dir`, …)
  - handles as literal directory names in the inbox lookup
  - `$( )`, backticks and process substitutions checked recursively
  - read only the frontmatter when scanning review files
  - heredoc bodies treated as data
- Contract README now says: git-guard catches agent mistakes, not a determined bypass.
- git-guard fixtures: 51 → 63.

## Next

- Copilot round 2 on 8188289 just came in: [copilot-review-round-2.md](copilot-review-round-2.md).
- Then PR 3 (the other hooks).
