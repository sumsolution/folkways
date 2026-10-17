# 0005. A closed effort is lifted, indexed and removed

| | |
|---|---|
| Status | accepted |
| Date | 2026-10-17 |
| Owner | Summer Jordan |

## Issue

`start` creates an effort folder with a `00-STATUS.md`, and `status` and `session-close` update it. No skill
says what happens when the work is done. In the folkways team repo, 9 of 17 folders in `efforts/` were
closed, so a reader or an agent listing `efforts/` sees finished work mixed with live work. What happens to
an effort when it closes?

## Context

- The design record moved from this repo to the team repo on 2026-10-17
  ([#38](https://github.com/sumsolution/folkways/pull/38)): research and plans went to the closed effort
  each served, open items became open efforts.
- Efforts hold working notes: plans, research, dated retros. They are not verified facts. `knowledge/` is for
  facts that stay true, each ending in `Last verified: YYYY-MM-DD (<method>)` so `knowledge-audit` can
  re-check it.
- Most outcomes already live elsewhere: decisions in ADRs 0001–0004, rules in
  [skill-authoring.md](../skill-authoring.md) and the skills.
- Links into closed efforts existed in both repos: 53 lines in this repo (ADRs, guide, contract README) and
  12 in the team repo (open efforts and `state/framework.md`).
- One closed effort held a living artifact: the review-rounds eval suite, rerun when skills change.
- Git keeps every deleted file. A link pinned to a commit SHA keeps working after the folder is removed.

## Decision

**When an effort closes:**

1. **Lift** what is still true and expensive to re-derive into `knowledge/`, by PR, with a
   `Last verified` footer. Decisions should already be ADRs; write one if missing.
2. **Move** anything still in use to the open effort or repo that uses it.
3. **Index** it with one line in `state/closed-efforts.md`: name, dates, outcome, links (ADRs, PRs, knowledge
   entries) and the last commit where the folder existed.
4. **Repoint** links to the folder to permalinks at that commit (`blob/<sha>/…`), then **`git rm`** it.

Not everything is lifted. A plan or a dated retro stays reachable through the index and is not copied.

## Options considered

| Option | For | Against | Trade-off |
|---|---|---|---|
| **Lift, index, remove** | `efforts/` shows only live work; facts that matter are verified in `knowledge/`; nothing is lost | Each close repoints links and needs a judgement on what to lift | A few minutes per close for a clean `efforts/` |
| Keep closed folders, add an index | No link changes; no lift judgement | `efforts/` keeps growing with finished work; agents read stale plans as current | Cheapest now, noise forever |
| Archive to `efforts/_closed/` | Files stay browsable on disk | Every link still breaks on the move; stale notes stay greppable and get cited as fact | The cost of a move without the benefit of a removal |

## Consequences

- The 9 closed efforts in the team repo are indexed in its `state/closed-efforts.md` and removed. Links here
  point to the team repo at `14c47bd`.
- The review-rounds eval suite moved to [packages/evals/suites/](../../packages/evals/suites). Its first
  results stay with the closed effort.
- Lifting is per effort and verified, so it can trail the removal: a candidate is listed in the index until
  it is lifted or dropped.
- No skill does this yet. A later change adds a close step to the skills.
