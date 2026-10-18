# 0010. Partition folkways by session mode and effort ownership, not by SDLC agents

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-18 |
| Owner | Summer Jordan |
| Reviewers | none required (`adr.reviewers.min: 0`) |
| Review window | until 2026-10-25 |

## Issue

folkways ships 16 skills (14,260 words) and a 1,076-word `instructions.md` that every session loads. The
skills have grown into each other:
- four of them write the same `00-STATUS.md`;
- about 10 trigger phrases are claimed by two skills;
- about 15% of skill text repeats rules stated elsewhere.

When folkways#40 put "close this effort" in `status`, the close prompts reached it in 1 of 12 eval runs.
The agents went to `session-close` or worked on the files directly.

Interactive and headless sessions also need different things, but no session is told which one it is. The
headless rules are copied into 16 skills, and they disagree about pushes and notifications.

How should folkways partition its skills, agents and instructions so each session loads only what it
needs and each request has one obvious skill?

## Relevant context

- Skill audit at folkways `fd86000`: [01](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/01-research-skill-audit.md). It covers the overlaps, the
  repetition (about 2,300 words), the headless conflicts, and what hooks already enforce.
- Eval evidence for the close miss: [evals-next 01](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/evals-next/01-trigger-results-close.md).
- Platform limits ([02](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/02-research-agents-and-modes.md)):
  - A Claude subagent's `skills:` preloads skills but doesn't limit them.
  - Subagents can't ask the user, and workflows can't pause for input partway through.
  - Copilot `handoffs` are ignored by the cloud agent.
- Mode signals per surface ([02](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/02-research-agents-and-modes.md)):
  - `COPILOT_AGENT_SESSION_ID` for the cloud agent, seen in
    [app-repo-first 23](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/app-repo-first/23-probe-copilot-github.md);
  - `GITHUB_ACTIONS` for Actions;
  - `CLAUDE_CODE_SESSION_ATTENDED=0` and `CLAUDE_CODE_ENTRYPOINT=sdk-cli` for `claude -p`. Both are
    undocumented.
- SessionStart hooks can return different context per mode on both tools ([02](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/02-research-agents-and-modes.md)).
- Hiding skills: `disable-model-invocation: true` removes a skill's description from context in both tools.
  In Copilot `-p` the skill then can't be reached at all.
- Platform targets (Summer, 2026-10-18): headless work and CI evals are built for Copilot first (cloud
  agent, Copilot CLI in Actions). Interactive sessions are built for Claude Code first.
- The TypeScript CLI takes only deletions until the Go port ([ADR 0011](0011-cli-in-go.md), [go-cli](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/go-cli/00-STATUS.md)), so
  deterministic steps can't move into the CLI yet.

## Decision

Options 2, 4 and 6 below, decided together because each depends on the others:

1. **Agents only for headless roles (option 2).** We will ship two agents: `reviewer` and
   `knowledge-auditor`.
   - Each is a short file: a role, its limits, and the skill it uses.
   - Interactive work stays in one main session that picks skills.
   - Agent files go in `.claude/agents/`, which Claude Code, VS Code and Copilot CLI read. A copy in
     `.github/agents/` is added if the cloud agent needs it **(unconfirmed: pending the control probe in
     02)**.
2. **The session-start hook states the mode (option 4).** It prints `Mode: interactive` or
   `Mode: headless (<surface>)`, checking Copilot's variables before Claude's.
   - In headless runs it also prints the single headless rule set.
   - `instructions.md` keeps one line for surfaces with no hook: "If the session context doesn't say
     `Mode: interactive`, act as headless."
   - `setup` and `onboard` become user-only (`disable-model-invocation: true`). The hook tells the user
     to run them.
3. **Consolidate around what each skill owns (option 6).**
   - `start`, `status`, `handoff` and folkways#40's Close become one `effort` skill. Its tasks are
     start/resume, update, handoff, roll-up, promote and close. Per-task detail lives in `references/`.
   - `session-close` keeps only the banking of what a session learned.
   - `prompts/review.md` shrinks to its inputs plus "use the `review` skill".
   - Repeated text leaves every skill: the config line, the personal-root line, `## Headless`, and the
     landing, push and worktree rules. The hook and `instructions.md` each state a rule once.
   - Descriptions are rewritten so that no trigger phrase belongs to two skills.

Why this combination:
- The lifecycle split is real, but neither platform can enforce it through agents (02).
- The confusion we measured comes from overlapping descriptions and duplicated duties. Fewer skills that
  each own one thing remove it directly.
- Telling the session its mode removes the repeated headless text and its contradictions.
- The headless roles do benefit from an agent: they run without a person and start from a prompt that
  names them.

The headless rule set (one place):
- Push only to a non-default branch, only in the draft-PR flow.
- @mentions on the draft PR are allowed.
- No destructive actions and no external systems.
- Flag every judgement call with what a reviewer needs.

Non-goals:
- Moving deterministic steps into the CLI or hooks ([06](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/06-adr-deterministic-steps.md)).
- The rule that `knowledge/` lands by PR (a separate ADR).
- Fixing git-guard's `-F` gap.
- Any change to how hooks reach app repos (ADR 0007).

Size:
- `instructions.md` ≤ 600 words.
- SKILL.md total ≤ 10,000 words (−30%).
- 13 model-invocable skills.
- About 6 PRs (implementation plan 05, to be written).
- One Go change in `sessionstart.go`.

## Options considered

Deep analysis: [03](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/03-options-architecture.md).

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Agents per SDLC phase (planner, researcher, implementer, reviewer, shipper) with handoffs | Matches how engineers think about the lifecycle; each agent's prompt could be small | Agents can't limit skills, handoffs don't work headless, and subagents can't ask, so every interactive phase still runs in the main thread; adds 5+ files on top of the skills | The right picture, with no mechanism to enforce it |
| 2 | Agents for headless roles only | Small; matches the roles that run unattended and are named by the caller; one file format reaches three surfaces | The cloud agent may need a second copy; two more files to keep in step with their skills | Uses agents where the platforms support them and nowhere else |
| 3 | No agents | Nothing new to maintain | Headless runs depend on description triggers, the failure we measured | Simplest, but leaves the headless roles to chance |
| 4 | The hook states the mode; interactive-only skills are hidden | Headless rules exist once; skills lose ~450 words; agents know their mode | Relies on undocumented Claude variables for `claude -p`; a hidden skill can't be reached in Copilot `-p` | A small Go change replaces 17 copies of the rules |
| 5 | Keep `## Headless` per skill (today) | No change | The conflicts in 01 stay; nothing tells the agent its mode | Keeps the status quo and its contradictions |
| 6 | Full consolidation (16 → 13 model-invocable skills, plus 2 user-only) | Removes the main overlap cluster and the duplicated duties; one status template | Merging four skills into `effort` risks a large skill and new trigger misses; the most PRs | Fixes the cause, at the cost of the riskiest rewrite |
| 7 | Boilerplate removal plus the `effort` merge only | Most of the word saving at about half the PRs | Leaves the smaller clashes (tests/qa, side-quest/ticket, the review prompt) | Cheaper, but leaves known clashes for later |

## Consequences

- **Easier:**
  - One skill per request for the effort lifecycle.
  - Agents know their mode.
  - Headless rules change in one place.
  - About 1,100 fewer always-loaded words, and about 4,000 fewer words of skill text.
- **Harder or costlier:**
  - `effort` is a large skill, and its description must cover six tasks without colliding with
    `session-close`.
  - Mode detection on Claude rests on two undocumented variables, so a test must pin them.
  - `test/skills.test.ts` has to change before any boilerplate goes, because it requires the config line,
    the personal-root line and `## Headless`.
  - Teams with overlays for `start`, `status` or `handoff` must move them to `effort`. `folkways sync`
    should warn about overlays for removed skills.
- **Not tried:** the `effort` merge has no proof of concept. The eval baseline on main
  ([07](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/07-plan-evals.md)) runs before the merge PRs, and the merge PR has to match or beat it.
- **Follow-on work:**
  - The Tier 0 lint: budgets and shared-trigger checks.
  - The Copilot CLI eval driver and a baseline on main before the merge PRs.
  - The `knowledge/` push-rule ADR.
  - A docs PR for bugs 1, 2 and 4 in [01](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/01-research-skill-audit.md).
- **Held PRs:**
  - folkways#40's Close task moves into `effort`; the PR is closed or rebased onto it.
  - ADR 0007 (folkways#41) stands. The plugin hooks it chooses are how the mode reaches the cloud agent.
  - ADR 0008 (folkways#43) stands, with one check: the headless fallback line has to reach surfaces that
    don't follow the `@` import.
- **Revisit if:**
  - Claude Code or Copilot adds per-agent skill limits or headless handoffs;
  - the `effort` merge regresses neighbouring triggers in the evals;
  - Claude drops or renames the mode variables.

## Related decisions

- [ADR 0005](0005-closing-efforts.md) (closing efforts): its procedure moves from `status` into `effort`, with the procedure itself
  unchanged.
- ADR 0007 ([folkways#41](https://github.com/sumsolution/folkways/pull/41)): depends on it. Plugin hooks
  carry the mode to the cloud agent.
- ADR 0008 ([folkways#43](https://github.com/sumsolution/folkways/pull/43)): constrains where the
  headless fallback line lives.
- Separate ADR (to come): `knowledge/` lands by PR, enforced by git-guard.
- Separate ADR (draft, [06](https://github.com/sumsolution/folkways-command/blob/272523f/efforts/architecture-review/06-adr-deterministic-steps.md)): deterministic skill steps move to the Go CLI or
  hooks.

## Interested parties

- The beta team, once onboarded: overlays for `start`, `status` and `handoff` need moving.

## Open questions

1. Does the cloud agent need a `.github/agents` copy? Probes F and G didn't settle it (02 Open-1). Until
   one does, ship both copies.

