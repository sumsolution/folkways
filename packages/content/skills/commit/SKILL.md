---
name: commit
description: Writes and makes a git commit in the team's convention (Angular type, module scope, no ticket in the scope) and names branches. Use when the user says "commit", "commit this", "write a commit message", "save my work to git", "make a branch" or "what should I call this branch", or before a risky operation that needs a checkpoint.
---

# commit

Commit one logical change with a message that follows the team convention. Squash merges turn the PR
title into the commit on the default branch, so the same rules cover PR titles (the `pr` skill).

## Config

- `folkways.yaml`: `git.commit_convention`, `git.scopes`, `git.branch_pattern`, `review.trigger`.
- `.local/me.yaml`: `projects_dir`, `review.trigger` (may only tighten the team value; the session context gives the effective one).

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

1. **Find the repo and branch.** Project code is changed only in a worktree under
   `<projects_dir>/.worktrees/<repo>/<branch>`, never in the user's own clone. If you are in the
   user's clone, stop and say so. In the team or solo repo, see [Where commits go](#where-commits-go).
2. **Check what is staged.** Read `git status` and `git diff --staged`. Stage files by name; don't stage
   everything blindly. Stop and ask if the diff holds:
   - more than one logical change (propose separate commits),
   - files unrelated to the task, generated files, or anything that looks like a secret or a
     machine-local path,
   - debug leftovers (`console.log`, focused tests, commented-out code).
3. **Run the repo's fast checks** (lint, typecheck, the tests for what changed), as its README or
   `AGENTS.md` says. A failing check is a finding to fix or report, not a reason to skip the check.
4. **Write the message** per [Message format](#message-format). Show it with the file list.
5. **Commit on approval.** A user asking you to commit approves the commit you showed them. If you are
   committing on your own initiative (a checkpoint before a risky operation), ask first.
6. **After the commit:** if the review trigger is `after-commit`, run
   `node_modules/.bin/folkways review --repo <worktree> --effort <effort>` from the team repo root (the
   pinned CLI; never a bare `npx folkways`, which can fetch an unrelated package) and say where the
   review file landed. In a solo repo there's no CLI: run the `review` skill in a fresh subagent given
   only the repo, branch, base and HEAD SHA. Never push; that is the `pr` skill, and every push needs approval.

## Message format {#message-format}

Follow `git.commit_convention` (default `angular`). `pattern`: the subject must match
`git.commit_pattern`; the rules below don't apply. `none`: any clear, imperative subject. For `angular`:

```
<type>(<scope>): <subject>

<body: why the change was needed, and anything surprising. Wrap at 72.>

<footer: BREAKING CHANGE: ..., Refs: EX-123, Closes #45>
```

- **Header:** under 72 characters. Imperative, present tense ("add", not "added"), lowercase first
  letter, no full stop.
- **Scope:** the module or area the change touches, from `git.scopes` when that list isn't empty. Never
  a ticket key: `fix(auth): handle expired tokens`, not `fix(EX-123): handle expired tokens`. Omit the
  scope only when the change is truly repo-wide.
- **Body:** optional for small changes. Explain why, not what; the diff shows what.
- **Footer:** see [Ticket reference](#ticket-reference). A breaking change gets `!` after the scope and
  a `BREAKING CHANGE:` footer that says what callers must do.

## Types {#types}

| Type | For |
|---|---|
| `feat` | A new feature |
| `fix` | A bug fix |
| `perf` | A performance improvement |
| `refactor` | A code change that is neither a fix nor a feature |
| `test` | Adding or correcting tests |
| `docs` | Documentation only |
| `build` | Build system or external dependencies |
| `ci` | CI configuration and scripts |
| `chore` | Maintenance that touches no production code or tests (release bumps, tooling config) |

If the change fits two types, it is probably two commits.

## Ticket reference {#ticket-reference}

The ticket goes in the branch name and, optionally, a `Refs: EX-123` footer. Never in the scope or
subject. Work with no ticket is allowed, but note it to the user as a process finding.

## Branch names {#branch-names}

Follow `git.branch_pattern` (default `{type}/{ticket}-{slug}`): `fix/EX-123-handle-expired-tokens`.

- `type` is the commit type of the main change.
- `slug` is 2 to 5 lowercase words joined by hyphens.
- No ticket: `{type}/{slug}`, and note the missing ticket.

Create the branch in a new worktree from a freshly fetched default branch:
`git fetch origin`, then `git worktree add <projects_dir>/.worktrees/<repo>/<branch> -b <branch> origin/<default>`.

## Where commits go {#where-commits-go}

| Repo | Paths | Branch |
|---|---|---|
| Project repo | any | a feature branch in a worktree; never the default branch |
| Team repo | `state/`, `efforts/`, `people/` | the default branch, rebased on push (keep history linear) |
| Solo repo | `state/`, `work/`, `inbox/`, `knowledge/`, `folkways.yaml` | the user's choice; ask |
| Team repo | `knowledge/`, `overlays/`, team skills, `folkways.yaml`, framework upgrades | a branch and a PR |

Use `git mv` for moves so history follows the file.

## Headless

No one can approve, so: commit only on a non-default branch (in a worktree for project repos), never
on `main` or the default branch, never push outside the draft-PR flow. Skip anything that would need a
question (mixed changes, suspected secrets): leave it unstaged and list it, with the reason, in the run
output for a reviewer.
