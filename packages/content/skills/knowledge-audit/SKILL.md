---
name: knowledge-audit
description: Re-checks the team's knowledge/ store against reality and proposes refresh, edit, merge or prune. Use when the user says "audit knowledge", "is our knowledge stale", "check knowledge/", "knowledge audit" or "verify the docs are still true", when session-close finds the audit due, or when the monthly GitHub Action runs it.
---

# knowledge-audit

A knowledge store decays quietly, and a confidently wrong fact is worse than a missing one: nobody
double-checks a fact that looks authoritative. This skill re-checks entries on a cadence and proposes
fixes. It proposes before writing, and pruning is deletion, so it always asks.

It is cheap because every `knowledge/` entry ends with `Last verified: YYYY-MM-DD (<method>)`. The audit
mostly re-runs the stated method. An entry with no method is itself a finding.

## Config

- `folkways.yaml`: `models`, `knowledge_audit`, `repos`.
- `.local/me.yaml`: `projects_dir`, `tool`.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

1. **Inventory.** List every `knowledge/` entry with its `Last verified` date, age, and re-verify interval
   from the tracker (`state/knowledge-audit.md`, see [Tracker](#tracker)). Past its interval = due. No
   interval recorded = assign one now from [Intervals](#intervals). No tracker = first run; everything
   is due.
2. **Verify what's due.** Breadth first, cheapest check first, re-running the entry's stated method.
   Cap the effort: confirm or flag each due entry, don't re-derive it. If it needs a real investigation,
   mark it `needs-investigation` and propose it as its own work.
   Fan out one subagent per due entry, on `models.<tool>.cheap` first; the method note is a
   self-contained instruction and the answer is small. Escalate to `strong` when a verdict is
   contested or the entry backs a decision. Keep the verdicts here, not in the subagents. Each verdict
   is one of:
   - **confirmed:** refresh the date and note the method used.
   - **drifted:** propose the specific correction and say what was read.
   - **unverifiable:** propose an explicit staleness marker. Don't silently trust it.
   - **needs-investigation:** too deep for an audit pass.
3. **Check the whole store.** Cheap, and catches what dates miss:
   - contradictions: entry vs entry, entry vs `state/` (where `state/` is the live truth)
   - duplicates and overlap: propose which entry absorbs which
   - dead references: links, paths, repos or SHAs that no longer resolve
   - missing provenance: no `Last verified` line or no method
   - path-rule violations: absolute or machine paths, `../` escaping a repo
   - prune candidates: duplicates an easier source of truth, or describes finished work nothing
     references
4. **Propose, in one table, worst first.** `Entry | Verdict | Proposal`. Entries that were confirmed
   without change get one summary line, not rows. Wait for approval; rows may be approved individually.
5. **Apply and record.** On approval, apply the edits and refresh dates. Update the tracker with the run
   date, entries checked and their verdicts, deferrals, and the next due date. **Write the tracker even
   when nothing was found**; it is what makes the next run cheap. `knowledge/` changes go on a branch
   and a PR; the tracker goes straight to the default branch, rebased, after approval. Commit nothing
   unless asked.
6. **Pruning is deletion, so always ask,** row by row, even inside an approved batch. Git history keeps a
   deleted entry recoverable; say so.

## Intervals {#intervals}

Interval follows what the entry describes (its volatility), not its length.

| Volatility | Interval | Looks like |
|---|---|---|
| High | 30 days | Package versions, commit SHAs, CI workflow contents, environment URLs, in-flight config |
| Medium | 90 days | Service architecture, route policies, deploy topology, product inventories |
| Low | 180 days | Repo ownership, org structure, terminology, external research |

An entry that mixes tiers takes the fastest interval, or is split into two entries.

## Tracker {#tracker}

`state/knowledge-audit.md` holds one row per entry (path, tier, interval, last verified, verdict) and a
run log. Start from [references/tracker-template.md](references/tracker-template.md)
(`.claude/skills/knowledge-audit/references/tracker-template.md`) when the file doesn't exist. Keep
`Last updated:` current.

## Headless

The monthly GitHub Action runs this with no one to ask.

1. Run steps 1 to 4 against a fresh checkout of the default branch. Project repos are read via
   `git fetch` and `origin/<default>` only.
2. Apply the proposals as edits on a branch `chore/knowledge-audit-<YYYY-MM>`, including the tracker
   update, and open a **draft PR**. The PR body is the proposal table. Prune candidates are listed in the
   body, never deleted in the diff.
3. **@mention the knowledge owners.** The owner of an entry is the `owner` field in the entry if it has
   one, else the author of the most recent `git log` commit that touched it. Say which rule you used
   per entry, and flag entries where the author is a bot or is no longer in the repo.
4. Flag every judgment call (a verdict you weren't sure of, an interval you assigned) in the PR body.
5. Never merge, never delete, never touch the default branch or external systems. If no branch can be
   pushed, write the proposal to `.local/knowledge-audit-<YYYY-MM>.md` and say so.
