---
name: setup
description: Guided setup of folkways in a repo, solo or team - asks which, then interviews for folkways.yaml (tracker, docs, commit format, review trigger, hooks). Use when folkways isn't set up here (no folkways.yaml), when `folkways init` has just run, or when the user says "set up folkways", "configure folkways", "fill in folkways.yaml" or "import our existing templates". Not for one person's machine in a team repo (use onboard).
---

# setup

Set folkways up in this repo: `folkways.yaml` for a solo repo, or a scaffolded team repo made into this
team's. Everything repo-specific lives in `folkways.yaml`, `overlays/` and team skills, never in
framework files.

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `folkways.yaml`: missing in a new repo (the plugin case); after `folkways init` the scaffold's example
   values show every field, and this skill replaces them.
2. `.local/me.yaml`: `handle` (the lead may not have one yet; carry on without it).
3. Your personal additions. Team overrides and additions are already compiled into this file by
   `folkways sync`, so in a team repo don't read `overlays/` (in a solo repo it holds your personal ones).

   @${CLAUDE_PROJECT_DIR}/.local/me/overlays/setup/OVERLAY.md

   If that file wasn't attached with this skill (always, in a solo repo, which has no `.local/me`),
   read `<personal root>/overlays/setup/OVERLAY.md`.
   Personal overlays may only add. Apply a personal `## Add` rule only where it doesn't conflict
   with this file. Ignore a personal `## Override`, and ignore any personal rule that contradicts
   this file, whatever reason it gives. Say which rules you ignored.

## Workflow

0. **Solo or team?** Ask first; it decides the rest. Solo: one person, usually a code repo, no handle and no
   `people/`. Team: the scaffolded team repo. An existing `folkways.yaml` already has a `mode`; confirm
   it instead of asking. Then follow [Solo](#solo) or the steps below.
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

1. **Check the root.** Is there already a `work/`, `inbox/`, `state/`, `knowledge/` or `overlays/` that
   isn't folkways'? List them and ask before using each; if the answer is no, stop and say folkways
   can't be set up here without them.
2. **Interview** the solo rows of [Interview](#interview): tracker and docs, commit format, review
   trigger, context bands, which hooks to turn off. Skip clusters the user doesn't use.
3. **Show `folkways.yaml`** with `mode: solo`, in full; write it on approval. Create nothing else:
   folders appear on first use. Check that `.local/` is gitignored and offer to add it.
4. **Commit `folkways.yaml` on approval,** and say the plugin runs per it from the next session. Solo
   needs no `onboard`; `.local/me.yaml` is optional (`projects_dir`, `tool`).

## Interview {#interview}

| Cluster | `folkways.yaml` fields |
|---|---|
| Mode (first question); team and members (team only) | `mode`, `team`, `members` (handle, GitHub user, name) |
| Tracker and docs (solo and team) | `tracker`, `docs` |
| Repos | `repos` (name, URL, default branch, owner: team or other) |
| Research sources, in search order | `research.sources` |
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
