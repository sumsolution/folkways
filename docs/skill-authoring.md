# Writing folkways skills

How to write a framework skill in `assets/skills/<name>/`. It is based on
[09-research-skill-authoring.md](design/09-research-skill-authoring.md) (the evidence) and the spec
([06 §3.2, §6, §7.5](design/06-v1-spec.md)). `test/skills.test.ts` checks the mechanical rules.

> **Overlays may change.** This guide uses the overlay model in spec §3.2: the skill reads
> `OVERLAY.md` files itself, as instructions. A pending ADR on overlay enforcement may change how overlays get enforced. If it changes the model, update step 0 in every skill and the
> [Overlays](#overlays-and-section-ids) section here.

## Rules

| Rule | Why | Checked by test |
|---|---|---|
| Frontmatter is `name` and `description` only | The only fields every tool reads the same way; Copilot CLI mishandles some extras ([09 §1](design/09-research-skill-authoring.md#1-frontmatter-and-limits)) | yes |
| `name` is lowercase-hyphenated, at most 64 chars, equals the folder | Spec requirement | yes |
| `description` is at most 500 chars, no `<` or `>` | Spec limit is 1,024; all skills share a listing budget of about 1% of context in Claude Code | yes |
| `SKILL.md` is under 500 lines | Spec and Anthropic guidance; the body loads whole on every use | yes |
| Step 0 reads `team.yaml`, `.local/me.yaml` and both overlays by explicit path | Spec §3.2; overlays only work if the skill reads them | yes |
| Overridable sections carry a unique `{#id}` | Team overlays replace sections by ID | yes (unique) |
| Writes, posts and pushes only after approval | Spec §6; instructions "Approval" list | wording only |
| A `## Headless` section | Spec §7.5; no one can approve in a headless run | yes |
| Relative links resolve inside the skill folder | Only the skill folder is synced to `.claude/skills/<name>/` | yes |
| Every reference file is linked from `SKILL.md` | References one level deep; no orphans | yes |
| No `[placeholder]` brackets outside code | A literal placeholder in output is a defect; use `<angle>` in code spans | yes |
| Examples are synthetic | The repo is public (`AGENTS.md`) | review |

## Anatomy of a skill folder

```
assets/skills/ticket/
├── SKILL.md              # always loaded when the skill triggers; the method
└── references/           # loaded only when a step says to read one
    ├── templates.md      #   one file per concern
    ├── jira.md           #   per-tracker mechanics, chosen by team.yaml
    └── example-overlay.md
```

- `SKILL.md` holds what every run needs: step 0, the workflow, the rules, overridable sections that
  are short, and the headless section.
- `references/` holds what only some runs need: templates, tracker mechanics, long examples. Link each
  from the step that needs it. Keep them one level deep: a reference doesn't send the agent to another
  reference.
- A file two skills need is copied into both, because a skill can't link outside its folder. The test
  keeps copies identical (see `shared reference files` in `test/skills.test.ts`).
- No scripts yet. Add one only when agents keep rewriting the same logic, and make it non-interactive
  with clear errors ([09 §6](design/09-research-skill-authoring.md#6-scripts-in-skills)). Deterministic
  checks belong in the `folkways` CLI or a hook, not a skill script.

## Descriptions

The description is the only part of the skill the agent sees before deciding to load it. Tools list
every skill's name and description in context and drop descriptions when the list gets too long.

- **Shape:** one short clause on what the skill produces, then `Use when ...` with the phrases a user
  would actually type, in quotes. Optionally one `Not for ...` clause naming the neighbouring skill.
- **Key use case first.** Listings truncate from the end.
- **Triggers, not the workflow.** A description that summarizes the steps lets the agent follow the
  summary and skip the body ([09 §2](design/09-research-skill-authoring.md#2-descriptions-and-triggering)).
- **A bit pushy.** Agents under-trigger. Name situations as well as phrases ("before a risky
  operation", "after a compaction").
- **Third person** for the clause ("Writes...", "Reviews..."), never "I" or "you".
- **Boundaries** go in the description when two skills are near each other (`pr` vs `commit`,
  `review` vs `tests`). Check them with the [trigger eval](#testing-a-skill).

```yaml
description: Opens or updates a pull request in the team's convention. Use when the user says "open a PR", "push this", "ship it" or "mark it ready for review". Not for writing the commit itself (use commit).
```

## Body structure

In this order. Omit a section that has nothing to say, except step 0 and Headless.

1. `# <name>`, then two or three sentences: what the skill is for and the standard it holds.
2. `## Step 0: Load config and overlays`: always the same four reads (see the skeleton). List the
   `team.yaml` and `me.yaml` fields this skill uses, so a reader knows what configures it.
3. `## Workflow`: numbered steps. Each step starts with a bold verb phrase. Say where approval is needed
   at the step that needs it.
4. Rule and format sections, overridable ones with `{#id}`.
5. `## Headless`: what changes when no one can answer.

Write for an agent that is capable but new to this team:

- **Say what, and why when it isn't obvious.** One clause of reason beats an all-caps MUST. Skip what
  any competent engineer already knows.
- **Be prescriptive where a mistake is costly or consistency matters** (approval gates, paths, formats
  other tools parse, the review file's frontmatter). Give latitude elsewhere; offer one default, not a
  menu.
- **Tool-neutral verbs.** "Read `team.yaml`", "run `git status`", "ask the user". Never tool names
  (`Read`, `Bash`, `view`): they differ between Claude Code and Copilot CLI.
- **No tool-specific substitutions** (`$ARGUMENTS`, `${CLAUDE_SKILL_DIR}`); Copilot CLI shows them as
  literal text.
- **Paths:** team-repo paths are relative to the team repo root, in code spans. When a step tells the
  agent to read a reference, give the link and the full path:
  `[references/jira.md](references/jira.md) (.claude/skills/ticket/references/jira.md)`. Copilot may
  resolve relative paths against the working directory.
- **Plain, direct, succinct.** Tables for lookups. Synthetic examples: `Example Team`, `EX-123`,
  `example.com`, handle `alex`.

## Overlays and section IDs

Spec §3.2: a skill's behavior = framework skill + team overlay + personal overlay, applied in that order.

| Level | Path | `## Add` | `## Override {#id}` |
|---|---|---|---|
| Team | `overlays/<skill>/OVERLAY.md` (+ data files it references) | applied | replaces the section with that ID |
| Personal | `people/<handle>/overlays/<skill>/OVERLAY.md` | applied | ignored, and the skill says so |

Deciding what gets an ID:

- **Give an ID to anything a reasonable team would do differently:** formats, templates, types, lists
  of values, where output goes, thresholds.
- **Don't give an ID to the method's core:** approval gates, the headless rules, and the rules that make
  the skill worth having. A team that wants a different core writes its own skill.
- **One concern per section.** A team that wants a different title rule shouldn't have to restate the
  template.
- **IDs are an interface.** Renaming or removing one breaks team overlays silently. Treat it like a
  breaking change: note it in `CHANGELOG.md`.
- **IDs are unique within the skill,** across `SKILL.md` and its references. Overridable sections may
  live in reference files (ticket templates do).

## Worked example: designing `ticket` for extension

`ticket` varies more between teams than any other skill: issue types, required fields, acceptance
criteria style, estimation, labels, trackers. The design keeps a small universal core and puts
everything else behind IDs or reference files.

| Part | Where | Overridable |
|---|---|---|
| Writing rules 1 to 8 (outcome over implementation, conclusion over path, trim pass...) | `SKILL.md` | no: they are the method |
| Workflow: one ticket in both directions, verified sections only, ask before any tracker write | `SKILL.md` | no |
| Issue types and which template each uses | `SKILL.md` `{#issue-types}` | yes |
| One template per type | `references/templates.md` `{#bug-template}`, `{#story-template}`... | yes, one at a time |
| Required and custom fields | `{#required-fields}` | yes |
| Acceptance-criteria format (Given/When/Then table or checklist) | `{#acceptance-criteria}` | yes |
| Estimation (none by default) | `{#estimation}` | yes |
| Definition of ready | `{#definition-of-ready}` | yes |
| Labels and components | `{#labels}` | yes |
| Titles | `{#titles}` | yes |
| Linking (epics, parents) | `{#linking}` | yes |
| Where drafts go | `{#drafts}` | yes |
| Tracker mechanics | `references/jira.md`, `github.md`, `linear.md`, chosen by `tracker.type` | teams add rules via `## Add` |

A team overlay then stays small. From
[`assets/skills/ticket/references/example-overlay.md`](../assets/skills/ticket/references/example-overlay.md):

```markdown
## Add
- Every story names the feature flag it ships behind, or says "no flag".

## Override {#estimation}
Estimate stories and tasks in points: 1, 2, 3, 5, 8. Anything above 8 gets split.
```

Lessons that apply to any skill:

- Ship **generic** templates and **synthetic** examples. A team's gold-standard examples go in its
  overlay folder, referenced from `OVERLAY.md`.
- Pick the default a new team can live with (no estimation, checklist or table ACs by type) and let the
  overlay tighten it.
- Put system-specific mechanics (a tracker's API quirks) in a reference chosen by config, not in
  conditionals through the body.

## Approval and headless runs

- **Interactive:** propose, show, then write on an explicit yes. Invoking a skill for an action ("commit
  this") approves that action as shown; a side effect the user didn't ask for needs its own yes. The
  approval list is in `assets/instructions.md`: tracker and wiki writes, posts, `knowledge/` and
  `state/` writes, deletes, non-local environments, pushes.
- **Headless** (`folkways run`, CI): no questions are possible. The skill writes local files or a draft
  PR, never touches `main` or external systems, and flags every judgment call with what a reviewer needs
  to decide. Say exactly what the skill does instead of asking at each approval point.

## Testing a skill

1. **Automated:** `pnpm test` runs `test/skills.test.ts` over every skill.
2. **Trigger eval:** give a fresh subagent the list of all skill names and descriptions and one
   realistic prompt; ask which skill it would load, or none. Use 5 to 10 prompts per skill, half of
   them near-misses that should load a neighbouring skill or none. Run each prompt 3 times. Fix a
   description that misfires and re-run with new prompts, so you aren't tuning to the test set.
   Record results in [09 §12](design/09-research-skill-authoring.md#12-trigger-eval).
3. **Behavior eval** (for a new or heavily changed skill): run the same task in a fresh session with and
   without the skill, one variable changed, at least 3 runs, graded blind against an answer key where
   one exists. Watch where the agent skips or misreads a step; fix the wording there.

## Anti-patterns

- A description that is vague ("Helps with tickets") or summarizes the workflow.
- Explaining what every engineer knows; padding.
- A menu of options with no default.
- Reference chains (a reference that points to another reference).
- Time-sensitive facts ("as of this month", version numbers) in a skill; they belong in `knowledge/`.
- Team or company specifics in the framework; they belong in overlays.
- Steps that need a question with no headless fallback.
- Tool names, tool-specific variables, or absolute paths.

## Skeleton

```markdown
---
name: <name>
description: <What it produces, one clause>. Use when the user says "<phrase>", "<phrase>" or "<phrase>", or <situation>. Not for <neighbour case> (use <other skill>).
---

# <name>

<Two or three sentences: what this is for and the standard it holds.>

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `team.yaml`: <fields this skill uses>.
2. `.local/me.yaml`: `handle`, <other fields>.
3. `overlays/<name>/OVERLAY.md` (team): apply its `## Add` rules; each `## Override {#id}` replaces the
   section with that ID in this skill or its references.
4. `people/<handle>/overlays/<name>/OVERLAY.md` (personal): apply its `## Add` rules. Personal overlays
   can't override; if it has `## Override` sections, ignore them and tell the user.

## Workflow

1. **<Verb phrase>.** <What to do, and why if it isn't obvious.>
2. **<Verb phrase>.** Read [references/<file>.md](references/<file>.md)
   (`.claude/skills/<name>/references/<file>.md`) when <condition>.
3. **Propose, then write on approval.** <What is shown; what is written where.>

## <Format or rule a team may change> {#<id>}

<The default.>

## Headless

<What the skill does instead of asking. Where output goes. How judgment calls are flagged.>
```
