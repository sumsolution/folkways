---
name: setup
description: Guided setup of a folkways brain repo, solo or team - asks which, then interviews for folkways.yaml (tracker, docs, commit format, review trigger, hooks). Use when folkways isn't set up here (no folkways.yaml), when `folkways init` has just run, or when the user says "set up folkways", "configure folkways", "fill in folkways.yaml" or "import our existing templates". Not for one person's machine in a team repo (use onboard).
---

# setup

Set up a brain repo: the repo sessions start in, where knowledge, state and efforts build up across
every project. It holds no project code; code changes happen in worktrees of the project repos. Solo:
a new repo for one person. Team: a scaffolded team repo made into this team's. Everything
repo-specific lives in `folkways.yaml`, `overlays/` and team skills, never in framework files.

## Config

- `folkways.yaml`: missing in a new repo (the plugin case); after `folkways init` the scaffold's example
  values show every field, and this skill replaces them.
- `.local/me.yaml`: `handle` (the lead may not have one yet; carry on without it).

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

0. **Solo or team?** Ask first; it decides the rest. Solo: one person's brain repo, no handle and no
   `people/`. Team: the repo `folkways init` scaffolded; if this isn't one, say to run
   `npx @folkways/cli init <dir>` and start `setup` there. An existing `folkways.yaml` already has a
   `mode`; confirm it instead of asking. Then follow [Solo](#solo) or the steps below.
1. **Interview in small clusters,** one topic per turn, per [Interview](#interview). Offer the
   scaffold's value as the default. Prefer links and exact names over descriptions.
2. **Verify what you can** as you go: repo URLs clone (`git ls-remote`), the tracker project key exists,
   MCP servers named in `research.sources` are connected. Mark anything unverified.
3. **Show `folkways.yaml`** in full with changes highlighted; write on approval.
4. **Import existing material,** if the team has any (ticket templates, ADR examples, review checklists,
   old agent instructions):
   - a change to a framework skill's behavior → `overlays/<skill>/OVERLAY.md` (`## Add`, or
     `## Override {#id}` for a section the framework marks overridable); templates and examples as files
     next to it;
   - a procedure the framework doesn't cover → a team skill in `.claude/skills/<name>/`, with a name
     that doesn't clash with a framework skill;
   - team conventions for every session → the team section of `AGENTS.md`.
   Propose each move in one table (Source | Destination | Why); write on approval. Never edit
   `.folkways/` or framework skills; `folkways sync` will refuse to run over edits.
5. **Seed `state/`** with `state/knowledge-audit.md` (no runs yet) and any tracker the team names.
6. **Update the READMEs** only where the team's answers change them.
7. **Commit on approval.** A new repo has no remote yet, so this first commit goes on the default
   branch. Then tell the lead the remaining manual steps: push to a new GitHub repo, add the audit
   workflow's secret, protect the default branch (linear history), and have each member run `onboard`.
   From then on, `folkways.yaml`, `overlays/` and team skills change by PR. On an existing repo with a
   remote, commit on a branch and offer a PR instead.

## Solo {#solo}

1. **Pick the folder.** The brain repo is never a code repo. If the current folder is empty, offer to
   use it; otherwise ask where the new one goes (for example `~/brain`) and create it on approval.
   `git init` it if it isn't a repo. Every later step works in that folder, not the session's.
2. **Interview** the solo rows of [Interview](#interview): tracker and docs, the project repos, research
   sources, commit format, review trigger, context bands, which hooks to turn off. Skip clusters the
   user doesn't use.
3. **Ask for `projects_dir`,** the absolute folder that holds the project clones (`~` is not reliably
   expanded). In a solo repo too, `projects_dir` is required: worktrees and the clone guard depend on it.
4. **Show the files** in full and write them on approval, in the brain repo folder from step 1 (the
   session's folder only if that's the one picked): `folkways.yaml` with `mode: solo`,
   `.local/me.yaml` with `projects_dir` (and `tool` if not `claude`), and a `.gitignore` holding `.local/`.
   Create nothing else; folders appear on first use.
5. **Commit `folkways.yaml` and `.gitignore` on approval** with `git -C <brain repo>`. Tell the user to
   start sessions from the brain repo from now on; the plugin runs per `folkways.yaml` from the next
   session there. Solo needs no `onboard`.

## Interview {#interview}

| Cluster | `folkways.yaml` fields |
|---|---|
| Mode (first question); team and members (team only) | `mode`, `team`, `members` (handle, GitHub user, name) |
| Tracker and docs (solo and team) | `tracker`, `docs` |
| Repos (solo and team) | `repos` (name, URL, default branch, owner: team or other) |
| Research sources, in search order (solo and team) | `research.sources` |
| Commit format (solo and team): Angular with scopes, Angular, a custom pattern, or none | `git.commit_convention` (`angular`, `pattern`, `none`), `git.scopes`, `git.commit_pattern` (regex on the subject line) |
| Branches and review (solo and team) | `git.branch_pattern`, `review.trigger` (`on-demand` turns the review gate off) |
| ADRs | `adr.targets`, `adr.repo_path`, `adr.confluence`, `adr.reviewers` |
| QA, notifications, models | `qa.allowed_envs`, `notifications`, `models` |
| Context bands (solo and team) | `handoff.bands` |
| Hooks to turn off (solo and team) | `hooks:` booleans: `git-guard`, `context-band`, `precompact`, `session-start`; all `true` by default |
| Automation | `knowledge_audit` |

A solo interview asks only the rows marked "solo and team". Skip a cluster nobody uses, and leave the
field out rather than inventing a value.

## Headless

Don't run setup headless: it needs a person's answers. If started headless, write a checklist of the
`folkways.yaml` fields still holding scaffold values (or, with no file, the questions) to the run
output, and change nothing.
