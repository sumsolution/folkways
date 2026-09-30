# EX-13: git-guard hook in Go

**Repo:** `projects/folkways` (branch `feat/hooks-go-git-guard`, PR #13, draft)
**Plan:** [plan 11, PR 2](../../projects/folkways/docs/design/11-plan-hooks-build.md). Contract: `projects/folkways/test/hooks/README.md`.

## Where we are

- d76fb42: Go module with `git-guard`, moved from the benchmark prototype. Passes the git-guard contract cases.
- 75477a4: merged `main`. PR #13 opened as a draft; Copilot reviewed it.
- Contract README (git-guard): catches agent mistakes, not a determined bypass.

## Next

- Copilot review round 1 on 75477a4 just came in: [copilot-review-round-1.md](copilot-review-round-1.md).
- Then PR 3 (the other hooks).
