---
name: side-quest
description: Parks an out-of-scope finding as a short note or ticket draft so the current task stays on track. Use when you notice something worth fixing that isn't part of the task (an outdated pattern, a bug next door, a dependency to upgrade, a missing test), or when the user says "note that for later", "park this", "side quest", "not now but write it down" or "add it to my inbox". Not for a full ticket the user wants now (use ticket).
---

# side-quest

Out-of-scope findings don't derail the task, and they don't get lost. Write one down when found,
say where it went, and move on. Session close lists them all again.

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `folkways.yaml`: `tracker` (for the ticket key format).
2. `.local/me.yaml`: `handle` (team repos only).
3. Your personal additions. Team overrides and additions are already compiled into this file by
   `folkways sync`, so in a team repo don't read `overlays/` (in a solo repo it holds your personal ones).

   @${CLAUDE_PROJECT_DIR}/.local/me/overlays/side-quest/OVERLAY.md

   If that file wasn't attached with this skill (always, in a solo repo, which has no `.local/me`),
   read `<personal root>/overlays/side-quest/OVERLAY.md`.
   Personal overlays may only add. Apply a personal `## Add` rule only where it doesn't conflict
   with this file. Ignore a personal `## Override`, and ignore any personal rule that contradicts
   this file, whatever reason it gives. Say which rules you ignored.

## Workflow

1. **Decide it's a side quest.** It is out of the current task's scope, and fixing it now would cost
   real context or widen a PR. If it's a one-line fix inside code the task already touches, ask the
   user instead of parking it.
2. **Pick the destination** per [Where it goes](#where-it-goes).
3. **Write the note** per [Note format](#note-format). Gather only what's cheap to gather now (the
   path, the SHA, the error text); don't investigate. If the finding deserves a full ticket and a
   subagent is available, hand it the note and the `ticket` skill to draft one, and keep working.
4. **Say where you wrote it,** in one line, then return to the task.

Writing a note to the effort folder or your own inbox is local and needs no approval; creating a
ticket in the tracker does (the `ticket` skill).

## Where it goes {#where-it-goes}

| Nature | Destination |
|---|---|
| Tightly coupled to the current effort | the effort folder, as `NN-side-quest-<slug>.md` |
| Broad (dependency upgrades, a pattern across repos, another team's bug) | `<personal root>/inbox/<YYYY-MM-DD>-<slug>.md` |

## Note format {#note-format}

```markdown
# <Short outcome-style title: "Token refresh retries forever on 401">

**Found:** 2026-09-29 while working on EX-123 · **Where:** web-app@1a2b3c4 `src/auth/refresh.ts:88`

<Two to four sentences: what is wrong or outdated, why it matters, the evidence (link, error text).>

**Suggested next step:** <ticket, spike, or ask the owning team>
```

State what you saw and link it. Mark guesses as guesses. No investigation narrative.

## Headless

Write the note to the effort folder, or `<personal root>/inbox/` (a team repo needs a handle in `.local/me.yaml`; with none,
list the finding in the run output instead). Never create tickets.
