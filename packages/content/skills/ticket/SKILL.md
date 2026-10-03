---
name: ticket
description: Drafts, splits and grooms tracker tickets (bug, story, task, spike, epic) that a developer can pick up cold, and creates them in the tracker on approval. Use when the user says "write a ticket", "create a Jira", "file a bug", "open an issue", "draft a story", "groom this ticket", "split this ticket" or "turn this into tickets", or when work needs a ticket before it starts. Not for a quick note on an out-of-scope finding (use side-quest).
---

# ticket

Write tickets a developer who wasn't in the session can pick up weeks later and finish without asking
what was meant. Templates give a ticket its shape; the writing rules give it its length and altitude,
and the rules win. A ticket that fills three sections well beats one that fills every section.

## Config

- `folkways.yaml`: `tracker` (`type`, `base_url`, `project_key`), `members`, `repos`.
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

1. **Pick the type** from [Issue types](#issue-types). If the user's words and the content disagree
   ("a story" that is really a bug), say so and suggest the better fit.
2. **Check it is one ticket, in both directions.** Propose a split when it holds outcomes that could
   ship separately, more than about five acceptance criteria, or more than about six repro steps.
   Split vertically (each piece delivers something observable), by workflow step, business rule, data
   variation or simple-then-complex; prefer a split that lets the team drop or defer a piece. Propose a
   merge when a ticket has no outcome anyone can verify without a sibling ticket; persistent hedging
   ("confirm whether", "if needed") is the symptom.
3. **Read the template** for the type in [references/templates.md](references/templates.md)
   (`.claude/skills/ticket/references/templates.md`).
4. **Fill only verified sections.** Content comes from the user, the effort's files, code at a SHA,
   or linked docs. A section with nothing verified gets deleted, not padded.
5. **Ask instead of guessing.** Never invent a link, SME, date, customer, environment or repro step.
   Ask for them in one short list. Unknowns that the implementer must resolve become an explicit open
   question in Notes, not a hedge spread through the body.
6. **Trim, then check length** (rules 7 and 8). Run the [ready check](#definition-of-ready).
7. **Write the draft file** per [Drafts](#drafts) and show it. This is local; no approval needed.
8. **Ask before any tracker write.** Show the title, type, parent, labels and fields you will set.
   Create or update only after an explicit yes; "write a ticket" approves the draft, not the tracker
   write. Tracker mechanics are in the reference for `tracker.type`:
   [references/jira.md](references/jira.md) (`.claude/skills/ticket/references/jira.md`),
   [references/github.md](references/github.md) (`.claude/skills/ticket/references/github.md`),
   [references/linear.md](references/linear.md) (`.claude/skills/ticket/references/linear.md`).
   With no tracker tool available, give the user the text to paste.
9. **Verify what the tracker stored.** Read the created ticket back; fix mangled formatting. Add the
   key to the draft file and report the link.

To customize this skill, a team copies
[references/example-overlay.md](references/example-overlay.md)
(`.claude/skills/ticket/references/example-overlay.md`) to `overlays/ticket/OVERLAY.md` and edits it.

## Writing rules

1. **Write for someone who wasn't in the session.** A developer picking it up cold, weeks later.
   Never reference the conversation ("as we discussed", "per the analysis above").
2. **Describe the outcome, not the implementation.** Bugs and stories say what must be true when done,
   not which files, functions or props to change. Too prescriptive: "Add an optional `variant` prop to
   the banner component and thread it through the config hook." Right altitude: "An editor can pick a
   compact banner layout for a campaign without a code change." Surface hard constraints; don't
   dictate the solution. Settling a shared convention (a naming scheme) is fine; that's an agreement.
   Explaining why a step is needed usually prescribes the step. Tasks are the exception: a task may
   name the technical approach, but stops at what needs to happen.
3. **Record the conclusion, not the path.** Final direction only: no dropped options, no "we
   originally planned", no investigation story. Two exceptions, one line each under Notes: a rejected
   option likely to be re-proposed, and a constraint the implementer would otherwise trip on. A
   decision that deserves a durable record is an ADR (the `adr` skill), not a ticket.
4. **State conclusions, link the evidence.** Ask of each sentence: does it tell the implementer what
   to do, or tell a reviewer I'm right? Cut the second kind (commit hashes as proof, "verified
   against", rebuttals). If the implementer needs real analysis, write a short doc in the effort and
   link the section. Most tickets need no doc.
5. **Omit sections rather than filling them.** Template bullets are prompts. No verified content,
   delete the heading. Never leave a literal placeholder.
6. **Don't paste code.** Reference a repo-relative path, optionally a symbol. Paste only when the
   snippet is the subject, under ten lines.
7. **Let length follow content.** No word limit, but no unearned length. Specifications (tables of
   inputs and outputs, enumerated cases, contracts) are exempt from trimming; prose around them isn't.
   Don't invent headings. Use a table wherever the content is a lookup or a set of cases.
8. **Trim before showing the draft.** For each sentence: would a developer do anything differently
   without it? If not, delete it. Delete restatements and empty framing; keep the shorter of two
   duplicate bullets. Check by name for four things a generic trim misses: scope fences ("X is out of
   scope") stay; orienting pointers at existing code stay; field data in the body (blockers, repos,
   priority) moves to tracker fields; prose around links shrinks to a list with a clause of gloss at
   most. Show only the trimmed draft; don't report what was cut.

**Call out gotchas.** A trap the implementer would otherwise hit (a shared cache, a consumer on an
old API version, a feature flag that defaults on) gets one line under Notes, prefixed `Gotcha:`.

## Grooming existing tickets

- Read the ticket, its comments and links first. Rewrite the description freely to meet the rules;
  show the before and after, and write it to the tracker only on approval.
- **Never change the title without asking.** Titles are referenced in branches, commits and
  conversations. A title that no longer matches the content is a finding: report it and offer a new
  one; don't explain the mismatch away in the body.
- Apply the split and merge check (step 2) to what you groom; propose, don't restructure silently.

## Issue types {#issue-types}

| Type | Use for | Template |
|---|---|---|
| Bug | Existing behavior that is wrong | `{#bug-template}` |
| Story | New or changed user-visible behavior | `{#story-template}` |
| Task | Technical work with no user-visible change (upgrades, refactors, config) | `{#task-template}` |
| Spike | Time-boxed research when uncertainty blocks estimating or designing; use sparingly | `{#spike-template}` |
| Epic | A goal made of several tickets | `{#epic-template}` |

The tracker reference maps these to the tracker's own types or labels.

## Titles {#titles}

A specific, scannable statement of the outcome or problem, about ten words. Bugs name the symptom
and where ("Checkout total ignores discount codes on saved carts"), stories and tasks the outcome
("Editors can schedule a banner to end at a set time"). No ticket type, component tags or dates in
the title; those are fields.

## Acceptance criteria {#acceptance-criteria}

Each criterion is observable and testable: someone could write a test from it.

- **Behavior with states and outcomes** (workflows, rules, API contracts): a Given / When / Then
  table. Given is a specific entry point or state ("A call to `POST /widgets` with a valid token",
  not "the API"). When names the real input with real field names. Then is an observable output
  (a response, a message, a rendered state), not internal state. A criterion that describes a code
  change belongs in a task.
- **Simple or non-functional criteria** (a copy change, a limit, a performance budget): a checklist.

## Required fields {#required-fields}

Default: type and title. Set a parent when one exists ([Linking](#linking)). The tracker reference
lists fields the tracker itself requires; a team adds its custom fields here by overlay.

## Estimation {#estimation}

None by default. Don't add points, sizes or time estimates unless an overlay asks for them.

## Definition of ready {#definition-of-ready}

A soft check before showing the draft, not a gate. Mention anything unmet; don't block on it.

- [ ] One outcome, small enough for one iteration (or one PR for a task)
- [ ] Acceptance criteria are testable (bugs: expected result and repro steps)
- [ ] Open questions are listed, each with who can answer
- [ ] Dependencies and parent are linked

The team's definition of done applies to every ticket; link it, don't restate it per ticket.

## Labels {#labels}

None by default. Use only labels that already exist in the tracker; never create one without asking.

## Linking {#linking}

- Set the parent (epic or parent issue) when the user names one or the effort's `00-STATUS.md` does.
  Don't guess a parent; ask.
- Link blockers and related tickets with the tracker's link types, not prose in the body.
- A split creates siblings under the same parent; say which ships first if order matters.

## Drafts {#drafts}

| Context | File |
|---|---|
| Inside an effort | `<effort>/NN-ticket-<slug>.md`, `NN` = the effort's next topic number |
| No effort | `<personal root>/inbox/<YYYY-MM-DD>-ticket-<slug>.md` |

A split writes one file holding every ticket in the set, in the order they should ship. After
creation, add the tracker key under each title.

## Headless

No questions and no tracker writes. Write the draft file per [Drafts](#drafts) and stop. Leave out
anything you would have asked for, and list it under a `## Open questions` heading at the end of the
draft. Flag every judgment call (type, split, parent, anything inferred) in the run output with what
a reviewer needs to decide.
