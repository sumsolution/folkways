# 00-STATUS.md template

`00-STATUS.md` is an index, not a report: where the effort stands, what's next, and pointers to the
topic files that hold the content. Anyone, including a fresh agent session, should be able to resume
from it cold. Keep it under about 60 lines; move anything longer into a numbered topic file.

```markdown
# EX-123: Handle expired tokens

**Owner:** alex · **Ticket:** EX-123 (link) · **Last updated:** 2026-09-29
**Repos:** web-app (branch `fix/EX-123-handle-expired-tokens`, worktree `.worktrees/web-app/fix/EX-123-handle-expired-tokens`)

## Where we are
One to three sentences. The last thing that was finished and verified.

## Next
1. The very next action, concrete enough to start without asking.
2. The one after.

## Blocked / waiting
- Waiting on the auth team to confirm the token lifetime (asked 2026-09-28 in #example-team).

## Decisions
- Refresh on the server, not the client. See 02-research-token-refresh.md.

## Files
| File | Holds |
|---|---|
| 01-ticket.md | Ticket text and acceptance criteria |
| 02-research-token-refresh.md | Options, evidence, confidence |
```

Rules:

- **Where we are** and **Next** are required. Drop other sections when empty.
- Link topic files; don't copy their content here.
- Update the `Last updated` date on every write.
- Git state (branch, worktree, last pushed SHA) goes in the header so the next session can check it.
