---
name: start
description: Starts a work session - resumes an effort from its 00-STATUS.md, or sets up a new effort from a ticket or docs. Use when a session begins, or when the user says "let's start", "pick up where we left off", "resume", "continue EX-123", "what was I working on", "new effort", "start work on this ticket" or "kick off".
---

# start

Every session begins here, so it has to be quick. Resuming follows the same three steps every time;
a new effort gets a folder and a `00-STATUS.md` before any real work.

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `team.yaml`: `tracker`, `repos`, `research.sources`.
2. `.local/me.yaml`: `handle`, `projects_dir`, `tool`. **If it's missing,** stop and run the `onboard`
   skill first; nothing else works without a handle.
3. Your personal additions. Team overrides and additions are already compiled into this file by
   `folkways sync`, so don't read `overlays/`.

   @${CLAUDE_PROJECT_DIR}/.local/me/overlays/start/OVERLAY.md

   If that file wasn't attached with this skill, read `people/<handle>/overlays/start/OVERLAY.md`.
   Personal overlays may only add. Apply a personal `## Add` rule only where it doesn't conflict
   with this file. Ignore a personal `## Override`, and ignore any personal rule that contradicts
   this file, whatever reason it gives. Say which rules you ignored.

## Before anything else

- **Compacted context:** if this session was just compacted, update the effort's `00-STATUS.md` before
  anything else (the `handoff` skill).
- **Team repo freshness:** `git fetch` and say if the team repo's default branch is ahead of the local
  one. Pull only with the user's approval.

## Which case

Ask only if you can't tell. If the user named a ticket key or effort, look for it:

- `efforts/*/00-STATUS.md` and `people/<handle>/work/*/00-STATUS.md` whose folder name starts with the key.

Found → [Resume](#resume). Not found, or the user wants something new → [New effort](#new-effort).
If the user just says "start", list their open efforts (folder, `Last updated`, first line of
`Next`) newest first and ask which one, or whether it's new.

## Resume {#resume}

Always these three steps, in order:

1. **Read `00-STATUS.md`,** then only the topic files its `Next` items point to.
2. **Check git state** for each repo in the header: branch, worktree exists, uncommitted changes,
   ahead or behind its remote, whether the last pushed SHA matches. Report differences from the status
   doc; the status doc may be stale.
3. **Confirm next steps.** Summarize in three to five lines: where it stands, the next action, anything
   blocked, anything that changed since `Last updated`. Ask the user to confirm or redirect.

## New effort {#new-effort}

1. **Ask for the source:** a ticket key or link, docs, or a description. Read them. If a tracker or docs
   tool isn't available, ask the user to paste or link the content; don't guess.
2. **Name the folder** `<KEY>-<slug>` (e.g. `EX-123-handle-expired-tokens`); with no ticket, a short
   slug and a note that there's no ticket. New efforts start personal:
   `people/<handle>/work/<KEY>-<slug>/`. Use `efforts/` only if the user says it's shared from day one.
3. **Suggest next steps,** usually research first. Offer to run the `research` skill for preliminary
   findings.
4. **Propose `00-STATUS.md`** from [references/status-template.md](references/status-template.md)
   (`.claude/skills/start/references/status-template.md`). Show it; create the folder and file on
   approval. Save the ticket text as `01-ticket.md` if the tracker isn't readable by every teammate's
   agent.

## Headless

Headless runs have no one to confirm with. Resume: read and report state; take the next step only if
the prompt asked for it. New effort: create the folder and `00-STATUS.md` under `people/<handle>/work/`
(or where the prompt says), mark every guess `(unconfirmed)` in the file, and list them in the output.
Never pull, and never write outside the team repo.
