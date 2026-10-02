---
name: handoff
description: Writes the effort's 00-STATUS.md so a fresh session can resume cold, then gives the user a resume prompt. Use when context is filling up (the 200K band), after a compaction, before a risky or long operation, or when the user says "hand off", "save progress so I can resume", "I need to stop", "set up the next session" or "checkpoint". Not for milestone updates or team roll-ups (use status).
---

# handoff

The file system is the session's memory. A handoff writes down what only this session knows, so the
next session starts from the status doc and git instead of from a summary.

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `folkways.yaml`: `handoff.bands`.
2. `.local/me.yaml`: `handle`, `projects_dir`.
3. Your personal additions. Team overrides and additions are already compiled into this file by
   `folkways sync`, so don't read `overlays/`.

   @${CLAUDE_PROJECT_DIR}/.local/me/overlays/handoff/OVERLAY.md

   If that file wasn't attached with this skill, read `people/<handle>/overlays/handoff/OVERLAY.md`.
   Personal overlays may only add. Apply a personal `## Add` rule only where it doesn't conflict
   with this file. Ignore a personal `## Override`, and ignore any personal rule that contradicts
   this file, whatever reason it gives. Say which rules you ignored.

## When

| Trigger | Do |
|---|---|
| 200K tokens (band 3) | Full handoff, then end the session |
| Just compacted | Update `00-STATUS.md` before any other work; your early turns are now a summary, so check claims against files and git |
| Before a risky or long operation | Update `00-STATUS.md` and commit it, so a failure loses nothing |

## Workflow

1. **Find the effort.** The one this session worked on: `efforts/<KEY>-<slug>/` or
   `people/<handle>/work/<KEY>-<slug>/`. No effort yet → offer to create one (the `start` skill) or put
   the handoff in `people/<handle>/inbox/`.
2. **Collect state from sources, not memory:** `git status`, current branch and HEAD SHA, unpushed
   commits, and open worktrees for each repo touched. Files written this session.
3. **Update `00-STATUS.md`** per [Status doc](#status-doc). Change what moved; keep the rest.
4. **Move content out.** Findings, analysis or drafts that live only in the conversation go into
   numbered topic files (`NN-<topic>.md`) linked from the status doc. The status doc stays an index.
5. **Show the diff** of the status doc and new files. Write on approval.
6. **Offer to commit** the effort folder (team repo `state/`, `efforts/` and `people/` go straight to the
   default branch, rebased). Push only on approval.
7. **Give the resume prompt,** one line the user pastes into a fresh session:
   `Resume people/alex/work/EX-123-handle-expired-tokens: read its 00-STATUS.md and run the start skill.`
8. **At band 3,** stop after this. Offer `session-close`; if context is tight, suggest running it in the
   fresh session (its first step handles that).

## Status doc {#status-doc}

Use [references/status-template.md](references/status-template.md)
(`.claude/skills/handoff/references/status-template.md`). What matters for a handoff:

- **Next** item 1 is concrete enough to start without asking: a command, a file and what to change,
  or a question and who to ask.
- **Where we are** says what is verified, not what was attempted.
- **Git state** in the header: branch, worktree path, last pushed SHA, anything uncommitted.
- **Pending decisions** (under `Decisions`, or `Blocked / waiting` when someone else decides) carry the options already weighed, one line each, so the
  next session doesn't redo them.
- Dead ends worth remembering: one line under Decisions ("Tried X; fails because Y"), only if the next
  session would otherwise try it again.

## Headless

Write `00-STATUS.md` and topic files without asking; that's the point of a headless handoff. Don't
commit to the default branch; leave the files for the draft PR or the next session. List anything you
weren't sure of under `## Blocked / waiting` as `Needs review:`.
