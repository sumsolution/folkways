---
name: setup
description: Configures a new team repo - interviews a team lead to fill in team.yaml, moves existing team material into overlays or team skills, and writes the first state/ files. Use right after `folkways init`, or when the user says "set up the team repo", "configure folkways", "fill in team.yaml", "add our Jira project" or "import our existing templates". Not for setting up one person's machine (use onboard).
---

# setup

Turn a freshly scaffolded team repo into this team's repo. Everything team-specific lives in
`team.yaml`, `overlays/` and team skills, never in framework files.

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `team.yaml`: the scaffold's example values show every field; this skill replaces them.
2. `.local/me.yaml`: `handle` (the lead may not have one yet; carry on without it).
3. `overlays/setup/OVERLAY.md` (team): apply its `## Add` rules; each `## Override {#id}` replaces the
   section with that ID in this skill or its references.
4. `people/<handle>/overlays/setup/OVERLAY.md` (personal): apply its `## Add` rules. Personal overlays
   can't override; if it has `## Override` sections, ignore them and tell the user.

## Workflow

1. **Interview in small clusters,** one topic per turn, per [Interview](#interview). Offer the
   scaffold's value as the default. Prefer links and exact names over descriptions.
2. **Verify what you can** as you go: repo URLs clone (`git ls-remote`), the tracker project key exists,
   MCP servers named in `research.sources` are connected. Mark anything unverified.
3. **Show `team.yaml`** in full with changes highlighted; write on approval.
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
7. **Commit on a branch and offer a PR** (`team.yaml`, `overlays/` and team skills go by PR). Then tell
   the lead the remaining manual steps: push to GitHub, add the audit workflow's secret, protect the
   default branch (linear history), and have each member run `onboard`.

## Interview {#interview}

| Cluster | `team.yaml` fields |
|---|---|
| Team and members | `team`, `members` (handle, GitHub user, name) |
| Tracker and docs | `tracker`, `docs` |
| Repos | `repos` (name, URL, default branch, owner: team or other) |
| Research sources, in search order | `research.sources` |
| Git and review | `git.scopes`, `git.branch_pattern`, `review.trigger` |
| ADRs | `adr.targets`, `adr.repo_path`, `adr.confluence`, `adr.reviewers` |
| QA, notifications, models | `qa.allowed_envs`, `notifications`, `models` |
| Automation | `handoff.bands`, `knowledge_audit` |

Skip a cluster the team doesn't use, and leave the field out rather than inventing a value.

## Headless

Don't run setup headless: it needs a person's answers. If started headless, write a checklist of the
`team.yaml` fields still holding scaffold values to the run output, and change nothing.
