---
name: pr
description: Opens or updates a pull request in the team's convention. Use when the user says "open a PR", "create a pull request", "push this", "ship it", "raise a PR", "update the PR description" or "mark it ready for review". Not for writing the commit itself (use commit) or reviewing someone else's PR (use review).
---

# pr

Get a reviewed branch into a pull request that a teammate can review in one pass. PRs are
squash-merged, so the title becomes the commit on the default branch.

## Config

- `folkways.yaml`: `git` (convention, scopes, merge), `review.trigger`, `notifications`, `members`.
- `.local/me.yaml`: `projects_dir`, `review.trigger` (may only tighten the team value; the session context gives the effective one).
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

1. **Check the branch.** Not the default branch; name follows `git.branch_pattern`; working tree clean.
   Uncommitted changes go through the `commit` skill first.
2. **Check the review gate.** A fresh-process review must cover the exact HEAD SHA: a file
   `NN-review-<sha7>.md` in the effort (or `<YYYY-MM-DD>-review-<sha7>.md` in `<personal root>/inbox/`) whose frontmatter `sha` equals
   `git rev-parse HEAD`. If there is none, or it covers an older SHA, run
   `node_modules/.bin/folkways review --repo <worktree> --effort <effort>` from the team repo root and
   wait for it. This holds for every trigger setting; the trigger only decides when the review runs
   on its own. In a solo repo there's no CLI: run the `review` skill in a fresh subagent given only the
   repo, branch, base and HEAD SHA.
3. **Act on the review.** Blocking findings get fixed (then commit, re-review) or the user explicitly
   defers them; deferred ones go in the PR description. Don't open a PR over an unaddressed blocking
   finding without saying so.
4. **Write the title and description** per [Title](#title) and [Description](#description).
5. **Ask two things together:** may I push, and draft or ready for review? If the answer on draft is
   unclear, open it as a draft.
6. **Push and open on approval.** `git push -u origin <branch>`, then create the PR with the GitHub
   CLI (`gh pr create`) or the tracker's equivalent. If no tool is available, say so and give the
   user the title and description to paste. Never force-push a branch someone else has checked out
   without asking.
7. **Report** the PR link and who to ask for review (from `members` or the repo's CODEOWNERS). Mention
   reviewers only if the user agrees; an @mention notifies immediately.

Updating an existing PR: same gate. After new commits, the review must cover the new HEAD before you
push again.

## Acting on review feedback {#feedback}

For comments from people, review bots or a fresh-process review:

1. **Sort each finding** against the change's intent and non-goals: realistic, false positive, or out of
   scope. Judge severity yourself; don't inherit the reviewer's label or threat model.
2. **Look for a cluster.** When several findings share one cause, that's a design question for the user,
   not a fix list: follow "Step back" in the framework instructions.
3. **Show the user the sort** before changing code: what you'd fix, what you'd decline and why, and the
   smaller option.
4. **Record non-goals** where the next reviewer will see them: the PR description, or the spec or
   contract the change implements.
5. **Reply to each declined finding** with the reason, once. Posting is the user's call.

## Title {#title}

The commit header rules from the `commit` skill: `<type>(<scope>): <subject>`, module scope, no ticket
in the scope, imperative, under 72 characters. It describes the whole change, not the last commit.

## Description {#description}

Short. A reviewer should know what to look at and how to check it in under a minute.

```markdown
## Why
<one or two sentences; link the ticket: EX-123>

## What changed
- <reviewer-relevant change, not a file list>

## How to verify
<commands or steps; what you ran and saw>

## Risk
<rollout, migrations, flags, what could break; "Low: <why>" is fine>

## Review
<verdict and SHA from the fresh-process review; deferred findings, if any>
```

Omit a section that has nothing verified to say. No session narrative, no restating the diff. Link
evidence (dashboards, docs, code at a SHA) instead of describing it.

## Size {#size}

Review friction shows in review rounds and time open, not only in line count. If the branch mixes
outcomes that could ship separately, propose splitting it before opening. Say why in one line.

## Headless

Push only to a non-default branch and open the PR as a **draft**, always. If the review gate isn't met,
run the review; if it still can't be met, open nothing and report why. Put every judgment call in a
`## Flagged for review` section of the description, with what the reviewer needs to decide. Notify per
`notifications.primary` (an @mention on the draft PR), falling back to `notifications.fallback`.
