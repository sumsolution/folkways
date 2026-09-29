# Way of working

Rules for every agent session in this repo. Skills in `.claude/skills/` hold the detail.

Framework skills: `start`, `status`, `handoff`, `session-close`, `research`, `side-quest`, `ticket`, `adr`,
`tests`, `commit`, `review`, `pr`, `qa`, `knowledge-audit`, `onboard`, `setup`.

## Session start

- Read `team.yaml` and `.local/me.yaml`. If `me.yaml` is missing, run the `onboard` skill first.
- **Resuming an effort:** read its `00-STATUS.md`, check git state, confirm next steps with the user.
- **New effort:** ask for a ticket or docs, read them, suggest next steps (usually `research`), create
  `00-STATUS.md`. The `start` skill does this.

## Where things go

| What | Where | Lands by |
|---|---|---|
| Durable facts | `knowledge/` | PR |
| Live trackers | `state/` | direct push |
| Shared efforts | `efforts/<KEY>-<slug>/` | direct push |
| Personal efforts | `people/<handle>/work/<KEY>-<slug>/` | direct push; promote with `git mv` to `efforts/` |
| Side quests, drafts with no effort | `people/<handle>/inbox/` | direct push |
| Machine-local or private | `.local/` | never committed |

Direct pushes are rebased and linear. Every push needs the user's approval.

## Status docs

- Every effort has a `00-STATUS.md`: where we are, what's next, pointers to topic files. Nothing else.
- Topic files are numbered: `NN-<topic>.md`.
- Update `00-STATUS.md` at milestones, before risky or long operations (commit then too), and at the end
  of every session, before session-close.

## Context and handoff

| Tokens in context | Do |
|---|---|
| 125K | Look for a stopping point. |
| 150K | Start no new multi-step work. Propose a handoff at the next checkpoint. |
| 200K | Write `00-STATUS.md` (`handoff` skill) and end the session. |

Don't start a task that will clearly overshoot the next band. In Claude Code a hook announces each band;
in Copilot CLI, check `/context`. After a compaction, update `00-STATUS.md` before anything else.

## Subagents

- Use them for throughput: parallel research, isolated parts of a change. The main agent orchestrates,
  reviews and merges, and asks the user on judgement calls.
- Cheap model first (`models.<tool>.cheap`). Escalate to `strong` for conflicting sources, claims a decision
  depends on, citations that fail a spot check, or low confidence.
- Scope parallel work so agents touch separate code.

## Side quests

Out-of-scope findings don't derail the task. Write a finding or ticket draft when found (`side-quest`
skill): in the effort if tightly coupled, else in `people/<handle>/inbox/`. Say where you wrote it.
Session close lists them all.

## Project repos

- Never check out, pull, reset or rebase in the user's clones under `projects_dir`. They belong to the user.
- **Research:** `git fetch`, then read `origin/<default>` (`git grep`, `git show`, or a detached worktree).
  Record the SHA you read.
- **Code changes:** a worktree on its own branch at `<projects_dir>/.worktrees/<repo>/<branch>`. Parallel
  subagents each get their own worktree off the feature branch.
- **Missing clone:** clone it into `projects_dir`, then the same rules apply.

## Git

- Commits: Angular conventional commits with a module scope from `git.scopes`, never a ticket:
  `fix(auth): handle expired tokens`.
- Branches follow `git.branch_pattern`, e.g. `fix/EX-123-handle-expired-tokens`.
- PRs are squash-merged, so the PR title follows the commit rules.
- A fresh-process review (`folkways review`) must cover the SHA before a PR opens.
- Ask before every push. Ask whether a PR is draft or ready; open it as a draft if unclear.

## Approval

Ask first for anything destructive, public, or that would appear to come from the user: tracker and wiki
writes, posting comments or messages, writes to `knowledge/` or `state/`, deleting files, running against
non-local environments (QA in `qa.allowed_envs` is the exception), pushing.

## Headless runs

`folkways run` or CI. Allowed: local files, draft PRs, @mentions or Slack notifications. Never: destructive
actions, `main`, external systems. Flag every judgement call in the output with what a reviewer needs.

## Writing

- Succinct and direct. Plain words, no metaphors, no over-explaining. Technical, not dumbed down.
- Every claim links to evidence: docs, code at a SHA (file and line), tickets, dashboards. Say when
  something is inferred.

## Customizing

- Never edit `.folkways/` or framework skills in `.claude/skills/`. `folkways sync` and `upgrade` stop if you do.
- Team changes go in `overlays/<skill>/OVERLAY.md` (add or override). Personal ones go in
  `people/<handle>/overlays/<skill>/OVERLAY.md` (add only).
- Every framework skill reads `team.yaml`, `.local/me.yaml` and both overlay files first.
