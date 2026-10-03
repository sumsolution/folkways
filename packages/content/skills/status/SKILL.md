---
name: status
description: Updates an effort's 00-STATUS.md at a milestone, rolls up team status or a standup from all efforts, and promotes a personal effort to the team. Use when the user says "update the status", "we hit a milestone", "standup", "what's the team working on", "status roll-up", "what's blocked", "promote this effort" or "share this effort with the team". Not for handing off to a fresh session (use handoff).
---

# status

Status lives in files, not in heads: every effort has a `00-STATUS.md`, and team views are read from
those files and `state/`. This skill keeps them current and reads them back.

## Config

- `folkways.yaml`: `members`, `tracker`.
- `.local/me.yaml`: `handle`.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

Then pick the task: [Update](#update), [Roll-up](#roll-up), or [Promote](#promote).

## Update

At milestones, and at the end of a session before `session-close`. Handing off to a fresh session, or a
checkpoint before a risky operation, is the `handoff` skill.

1. **Read the current `00-STATUS.md`** and check it against git and the effort's files.
2. **Change what moved:** `Where we are`, `Next`, `Blocked / waiting`, `Decisions`, the file index, the
   `Last updated` date. Keep the shape in
   [references/status-template.md](references/status-template.md)
   (`.claude/skills/status/references/status-template.md`). Content goes in topic files, not here.
3. **Show the diff; write on approval.** Offer to commit it.

## Roll-up {#roll-up}

Read, don't write. Sources: `efforts/*/00-STATUS.md`, `people/*/work/*/00-STATUS.md` (in a solo repo `work/*/00-STATUS.md`), and `state/`.

```markdown
## Team status, 2026-09-29

| Effort | Owner | Where we are | Next | Blocked | Updated |
|---|---|---|---|---|---|
| EX-123-handle-expired-tokens | alex | Fix merged to dev | QA in dev | — | 2026-09-28 |

**Blocked:** <each blocker, owner, since when>
**Stale:** <efforts not updated in 14 days>
```

- Sort blocked first, then by last update.
- A status doc not updated in 14 days is **stale**; say so, don't guess its state.
- Work visible in git or the tracker with no effort folder is a process finding; list it.
- For a **standup**, the same data per person: done since the last working day, next, blocked.
  Owners speak last in a live standup, so lead with the list, not a person.

## Promote {#promote}

Team repos only; a solo repo has no `efforts/`. An effort starts in `people/<handle>/work/` and moves to `efforts/` when a second person joins or the
team needs to see it.

1. Confirm the folder and that no `efforts/<same name>/` exists.
2. `git mv people/<handle>/work/<KEY>-<slug> efforts/<KEY>-<slug>` so history follows it.
3. Update links inside the effort and the owner line in `00-STATUS.md`.
4. Commit and push on approval (`efforts/` goes straight to the default branch, rebased).

## Headless

Roll-ups: write the report to the path the prompt gives, or print it. Updates: write the status doc
but don't commit to the default branch. Never promote an effort headless; list candidates instead.
