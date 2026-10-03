---
name: session-close
description: Banks what a session learned - proposes where each durable fact, status change and lesson should be saved, then writes it on approval. Use when the user says "close the session", "wrap up", "bank this", "end of day", "what should we save", "before I go" or "anything worth keeping". Not for writing the status doc for a fresh session (use handoff).
---

# session-close

A session produces artifacts (already saved) and understanding: facts found at cost, status that moved,
procedures that worked. Understanding evaporates when the session ends; this skill catches it. It always
proposes before writing, and never commits unless asked.

## Config

- `folkways.yaml`: `repos`, `models`, `knowledge_audit`.
- `.local/me.yaml`: `projects_dir`, `tool`.
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

0. **Decide where to run.** Banking competes with the session for context. Check the context band.
   - **Room left, no compaction:** run here. This is the default.
   - **Tight, not compacted:** do steps 1 and 2 here while memory is live and write the candidate list to
     a scratch file in `.local/`. Hand steps 3 to 5 to a fresh session.
   - **Already compacted, or nearly out:** hand off to a fresh session pointed at this session's
     transcript. After a compaction the live session sees only a summary of its early turns; the
     transcript on disk has the raw record, so a fresh session can see more of the session than the
     session can. Give the user the one-line prompt to paste.
   - **Don't delegate the whole routine to a subagent.** It inherits none of the conversation, so the
     expensive extraction has to happen here anyway. Subagents are for the bounded lookups in step 3.
1. **Scope the session.** Re-read the conversation for facts derived at cost, decisions, corrections to
   earlier beliefs, dead ends worth remembering, preferences the user expressed, and every side quest
   (out-of-scope findings, drafts, tickets you set aside). Run `git status` and `git diff` in each
   repo touched. Name the effort this belongs to and the project repos touched. "Nothing to bank" is a
   valid result and better than manufacturing an entry; say so and go to step 5.
2. **Classify each candidate to exactly one home.** Use [Homes](#homes) and [Don't record](#dont-record).
3. **Reconcile against what exists** before proposing a new file. See [Reconcile](#reconcile). A subagent
   on `models.<tool>.cheap` may run the lookups; the judgment stays here.
4. **Propose, in one table.** `Target | Action | Why`, grouped by home, then a short **considered and
   dropped** list, one line each. That list is where most of the value shows up: it is the user's chance
   to catch a fact you were about to lose. List every side quest found, with where it lives now (or
   propose writing it to `<personal root>/inbox/`). Wait for approval; the user may approve rows
   individually.
5. **Apply, then report.** Write the approved rows, refresh dates, keep provenance. Report files written,
   what was left out and why, and any side quests still unfiled. Then:
   - If the knowledge audit is due (see `state/knowledge-audit.md`, or no tracker and `knowledge/` is
     non-empty), offer to run `knowledge-audit`. Don't start it unasked.
   - Commit nothing unless asked. Landing paths: `knowledge/` changes go on a branch and a PR;
     in a team repo, `state/`, efforts and `people/` go straight to the default branch, rebased, after
     approval; in a solo repo, ask where they land.

## Homes {#homes}

Each candidate gets one home.

| Home | Test | Requirements |
|---|---|---|
| `knowledge/` (by PR) | Expensive to derive, still true in weeks, not already recorded in the source repo | One topic per file. Ends with `Last verified: YYYY-MM-DD (<method>)`; the method is what the audit re-runs |
| `state/` (direct) | The live status of a tracked activity moved | `Last updated:` date. Supersede, don't delete |
| Effort `00-STATUS.md` (direct) | Where the effort stands, what's next, what's blocked | Readable cold; update via `handoff` if it needs more than a line |
| Overlays | A procedure will recur, or a skill behaved wrong for this team | Amend the existing overlay; a new one needs a real second occasion |
| `<personal root>/inbox/` (direct) | Side quest or draft with no effort | Say where you wrote it |
| Tool auto-memory (optional) | A personal preference of the user's, for one tool | Distinct from `knowledge/`: the tool owns it, the team never sees it. Never conflate the two, and never symlink one to the other |

A fact the team needs goes in `knowledge/`, not only in auto-memory.

## Don't record {#dont-record}

- Anything re-derivable from code or git in under a minute.
- Restatements of what a file or doc already says.
- Unconfirmed speculation (record the question instead, in the status doc).
- Narration of the session ("first we tried...").
- Secrets, tokens, or personal data.

## Reconcile

- **Updating beats creating.** Search `knowledge/`, `state/` and the effort for the same topic first; add
  to that entry.
- **A contradiction is a decision, not an overwrite.** Show the old claim and the new evidence side by
  side and let the user choose. In `state/`, mark the old claim superseded in place, with the date and
  the evidence; the record of having been wrong is worth as much as the correction.
- **Never rename** a file or heading without asking; other entries link to them.
- **Paths** are repo-root-relative. A reference into another repo names the repo first
  (`web-app: src/auth.ts`). No absolute or machine-specific paths.

## Headless

No one can approve, so write nothing to `knowledge/`, `state/` or the effort. Run steps 1 to 4 and write
the proposal table, the considered-and-dropped list and the side-quest list to
`<personal root>/inbox/session-close-<YYYY-MM-DD>.md`, or to a draft PR branch if the prompt asked for
one. Mark each judgment call (a home you weren't sure of, a contradiction) so a reviewer can decide.
