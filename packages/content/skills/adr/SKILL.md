---
name: adr
description: Writes an Architecture Decision Record (context, real options with honest trade-offs, decision, consequences), drafted locally and published on approval. Use when the user says "write an ADR", "record this decision", "architecture decision", "should we document this decision" or "supersede ADR 0007", or when a design decision with real alternatives is being made. Not for exploring a question before a decision (use research) or for work items (use ticket).
---

# adr

Record a significant decision so a teammate who wasn't in the room can understand it, and challenge
it, six months later. The value is the reasoning: the context, the options that were really on the
table, and what the choice costs. Show trade-offs honestly; don't pretend the choice was obvious.

## Config

- `folkways.yaml`: `adr` (`targets`, `repo_path`, `confluence`, `reviewers`), `docs`, `members`, `repos`.
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## When an ADR is warranted

Any of these:

- Architecturally significant: structure, integration patterns, deployment, data storage, frameworks
  and tools.
- Measurably affects a quality attribute: security, availability, performance, cost.
- Hard or costly to reverse.
- A real trade-off with no clear winner.
- Touches dependencies, interfaces or domain boundaries, or another team's system.
- Contested: people disagree and the team commits anyway.

Skip it when the decision is small in scope, risk and cost, or already recorded elsewhere (a
standard, an existing ADR). If none apply, ask whether the user still wants one.

**Suggest one at the moment of decision.** When a design choice with real alternatives is settled in
a session, offer an ADR then, while the options and reasons are fresh; don't wait for session close.

## Workflow

1. **Establish the decision.** One decision per ADR. Is it decided (recording it) or open (drafting
   to drive it)? That sets the title ([Naming](#naming)). Does it belong to a ticket or effort (that
   sets where the draft lives)? Does it supersede or relate to existing ADRs? Search the targets for
   them.
2. **Gather content.** Start from what the user gives: a description, an RFC or analysis to distill
   (an RFC explores; an ADR decides). Interview for gaps in small clusters, not all at once:
   - the issue, and why now;
   - relevant context: probe for concrete service, endpoint and field names, with links;
   - **at least two real options**, each with pros, cons and a one-sentence trade-off. If the user
     offers only the preferred option, ask what else was considered and why it was dismissed;
     "do nothing" counts when it is viable;
   - the decision and rationale, including failed attempts or proofs of concept that shaped it;
   - consequences and follow-on work, the bad ones too;
   - reviewers and interested parties ([Reviewers](#reviewers)).
3. **Draft locally** with [references/template.md](references/template.md)
   (`.claude/skills/adr/references/template.md`): `<effort>/NN-adr-<slug>.md`, else
   `<personal root>/inbox/<YYYY-MM-DD>-adr-<slug>.md`. Status `draft`. Mark every section you had to
   guess with `(unconfirmed)` and list them for the user. For the right altitude, see
   [references/example.md](references/example.md) (`.claude/skills/adr/references/example.md`).
4. **Run the [done-check](#done-check)** and fix what fails, or say what's missing.
5. **Confirm before publishing:** the title, the group or domain label, the owner, and the target
   ([Targets](#targets)). Never assume a location. Publish only on an explicit yes, as `proposed`
   for review.
6. **Carry it through review.** Update the status as it moves ([Status values](#status-values)).
   Each status change on a wiki or repo is a write: show it and ask first.

## Writing style

- **Concrete, linked context.** Real service, field and endpoint names, linked so a reviewer can
  verify. Vague context is how ADRs fail.
- **State one decision plainly:** "We will serve all settings from the settings API." Active voice.
- **Trade-off column: one honest sentence per option,** the net judgment, not a restatement of pros.
  For example: "Keeps the status quo." / "Easy way out, but exposes a confusing data model." /
  "Kicks the can on the long-term strategy, but decouples reads now."
- **Brevity is fine.** One or two pages. Link deep analysis above the options table instead of
  bloating it. Options may have variants (2a, 2b).
- Plain words, no marketing adjectives, no scoring that looks more precise than the judgment behind it.

## Naming {#naming}

| State | Title | Example |
|---|---|---|
| Decided | A declarative statement of the decision | "Serve all settings from the settings API, read only" |
| Open | The core question; rename when accepted | "Where should we keep ADRs?" |
| Superseded | Keep the title; mark it superseded and link the new ADR | "SUPERSEDED: Serve all settings from..." |

Repo files are `NNNN-<slug>.md`, the slug from the decided title.

## Anti-patterns

| Pattern | Looks like |
|---|---|
| Fairy tale | Only pros; the chosen option has no cons |
| Dummy alternative | A straw-man option added to make the choice look inevitable |
| Sprint / tunnel vision | One option considered, short-term view only |
| Free lunch | Consequences listed are all harmless |
| Sales pitch | Marketing language instead of evidence |
| Policy in disguise | A commanding rulebook, not a decision with reasons |
| Mega-ADR, novel | Several decisions in one record, or pages of narrative for a simple choice |
| Maze | The content doesn't answer the question in the title |
| Decided alone | No peers engaged before it was written up as settled |

## Done-check

- [ ] **Evidence:** claims link to sources; a risky option was tried (proof of concept, benchmark) or
      the record says it wasn't.
- [ ] **Criteria:** at least two real options, compared on the same concerns.
- [ ] **Agreement:** the right reviewers are named ([Reviewers](#reviewers)).
- [ ] **Documentation:** published to the target and linked from the ticket or effort.
- [ ] **Review:** follow-on work is ticketed or listed, and it says what would make us revisit.

## Targets {#targets}

From `folkways.yaml` `adr.targets` (`repo`, `confluence`, or both). If unset, ask; never guess.

- **Repo:** `<adr.repo_path>/NNNN-<slug>.md` (default path `docs/adr`) in the repo the decision
  concerns; ask which when it spans repos. `NNNN` is the next number after the highest on the fetched
  default branch, zero-padded; numbers are never reused, even for rejected records. Assign it at
  publish time, not in the draft. It lands by a branch and a PR (the `commit` and `pr` skills),
  because docs are team-reviewed; the PR is the review venue.
- **Wiki** (`adr.confluence`: `space`, `parent_page`, `group_label`): publish through an MCP server
  for the wiki if one is connected; otherwise give the user the page content to paste. Put the group
  label on the page.
- **Both:** the repo file is the source; the wiki page links to it or mirrors it. Say which.

## Reviewers {#reviewers}

From `folkways.yaml` `adr.reviewers` (`min`, `outside_team`). Defaults: at least 2, at least 1 from
outside the team, no more than about 8.

- Reviewers reflect the decision's scope: at least the tech lead or a senior engineer, plus the
  domain architect when there is one. An outside reviewer hedges against groupthink.
- Interested parties are listed separately; they are told, not asked to approve.
- Review window: team-scoped decisions hours to days; larger ones one to two weeks unless urgent.
  Write the window in the record. Silence past it counts as agreement.
- Use plain-text names while in review. Real @mentions notify immediately; use them only when the
  user approves sending the review request.

## Status values {#status-values}

`draft` → `proposed` (in review) → `accepted` → `superseded`; or `rejected`, with the reason recorded.

- Accepted and rejected records are immutable. A change of mind is a new ADR that supersedes the
  old one. The only edit to the old record is its status (`superseded by NNNN`) and a link; link
  both ways.
- On acceptance, record the date and who agreed.

## Template

The template is the `{#template}` section of [references/template.md](references/template.md)
(`.claude/skills/adr/references/template.md`). A team replaces it with `## Override {#template}` in
its overlay, and keeps its own gold-standard examples next to `overlays/adr/OVERLAY.md`.

## Architecture knowledge {#architecture-knowledge}

Reserved. Guidance on decision quality (principles, quality-attribute trade-offs, common patterns)
will come from a dedicated session (spec §9.2). Until then, a team can add its architecture
principles here with `## Override {#architecture-knowledge}` in `overlays/adr/OVERLAY.md`, and the
skill checks each option against them.

## Headless

No questions and no publishing. Write the draft file (step 3) with status `draft`, mark every
guessed section `(unconfirmed)`, and list open questions at the end of the draft. Don't assign a
number or pick a target. Flag every judgment call (options you inferred, the decision if not stated,
reviewers) in the run output with what a reviewer needs to decide.
