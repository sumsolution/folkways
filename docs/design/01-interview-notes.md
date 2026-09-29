# Interview notes

Raw record of the alignment interviews, by round. Decisions graduate to `00-STATUS.md`;
this file keeps what was said and why.

## Round 1 — 2026-09-27: scope and audience

- **What it is:** a clone of agent-space that goes further, focused on **multiple users on a
  team**. Built as a **generic template** for a typical domain engineering team of **3–8
  people**. The beta team tests it and we iterate from there.
- **Shared across the team:** `skills/`, `knowledge/`, `state/`.
- **Personal:** `work/`.
- **Sync mechanism:** undecided (git + PRs, direct push, git + a hosted layer).
- **Relationship to the current agent-space:** undecided (successor, sibling, or personal
  overlay on the team base).
- **Name:** to be chosen together.

## Round 2 — 2026-09-27: functionality and tooling

- **In scope for v1:** shared knowledge with trust, team state and visibility, onboarding,
  handoffs between people.
- **The bigger goal, in the user's words:** "build a framework of sorts that encodes my
  methods to make them repeatable." It adds structure around common activities: code review,
  commit/push/PR, research and discovery for new initiatives, and others.
- **Automation:** knowledge-audit and session-close should run automatically, via hooks or
  some other mechanism.
- **Collaboration:** the user imagined `state/` holding team efforts and `work/` holding only
  truly personal efforts. Discussed and resolved in round 3 as a three-tier split.
- **Tooling:** Claude Code (CLI and desktop), Cursor / Copilot, mixed. **The template has to
  be as tool-agnostic as possible** (AGENTS.md plus Claude-specific layers).

## Round 3 — 2026-09-27: methods (code review, git/PR, discovery, way of working)

### Layout: three tiers (accepted)

| Area | Holds | Owner |
|---|---|---|
| `state/` | Live trackers and the team's rolled-up status board | Team, one file per track |
| `efforts/` (new, shared) | Team initiatives, a folder each: `00-STATUS.md`, owner, contributors, numbered artifacts | Named owner, open to contributors |
| personal `work/` | Scratch, drafts, personal efforts | Individual |

Efforts start personal and are **promoted** to `efforts/` when a second person joins or the
effort becomes team-visible. Status roll-ups, standups and handoffs read `efforts/*/00-STATUS.md`.
**Every initiative gets a `00-STATUS.md`,** whatever its size.

### Code review

Confirmed method, from a past PR review:

1. **Review against intent, not only the diff.** Load the ticket, design of record, decision log
   and `knowledge/` entries first.
2. **Verify by running.** Check out at a pinned SHA, run the suite, and test risky claims with
   throwaway tests. When docs and source disagree, source is the citation.
3. **Verdict first,** then a findings table: ID + kind (Decision / Fix / Fix before QA / Question).
4. **Evidence carries provenance.** Env, window and read date for metrics.
5. **Draft locally.** Numbered file in the effort; a rereview is a new file, not an edit.

Additions:

- **Every claim links to authoritative evidence:** docs, code (file/line at a SHA), or other
  primary sources. No unsupported assertions. *(Applies to all agent output, not only reviews.)*
- **Cross-team integrations:** clone the other team's repo locally. Read their docs, and their
  source when needed, to understand the integration.
- **Standards hierarchy:** follow repo patterns, README and AGENTS.md, **but** weigh best
  practices and evolving technology. If a repo pattern is outdated, call it out, respect the
  PR's scope, and **suggest follow-up tickets** for out-of-scope work.
- **Tone:** friendly and collaborative, direct and succinct. Follows
  [Conventional Comments](https://conventionalcomments.org/).
- **Posting is the user's call.** Always provide drafts and offer to post.
- **Team-uniform, individually extensible.** Every teammate's agent gives the same baseline
  review. An individual can layer extra rules or guidance in their local instance.
  → Needs an **extension mechanism** (a shared skill plus a personal overlay).

### Commit / push / PR

- **Commits to the default branch** (usually `main`, sometimes `beta` or others) use
  **Conventional Commits, Angular rules, with a module scope, not a ticket scope.**
  *Enforcing team standards like this is a reason the project exists.* (Note: the existing
  PR title `feat(PROJ-1623): …` is exactly the pattern this rule replaces.)
- **PRs are squash-merged.** The PR title becomes the commit message, so it follows the same
  convention.
- **Push always waits for explicit approval.** The agent asks whether to push.
- **Opening a PR:** ask whether to open it as a draft or ready for review. If that's unclear,
  or the agent is headless, **open it as a draft.**
- **Branches:** `{type}/{TICKET-ID}-{short-descriptive-name}`, e.g.
  `fix/PROJ-1234-handle-api-failures`.
- **Pre-PR review gate:** before any PR opens, the **code-review skill runs in a fresh session
  with no context from the agent that did the work.** A clean reviewer is required.

### Research and discovery

A past epic's arc (current state → gaps → open questions → target architecture → breakdown → Jira
plan → ticket drafts) is a typical **epic-level** initiative shape, not universal. The universal
part is `00-STATUS.md`.

"Research and discovery" means the **information-gathering process itself**: how to search
codebases, which MCPs and CLIs to use, fallbacks when they're unavailable, and the company
resources for finding information. **Needs its own interview.**

### Other recurring activities to encode

Ticket drafting, ADRs, QA testing, unit test writing.

### The general way of working (needs a deeper interview)

- **Context management and parallelism are constant priorities.**
- **The file system is short-term memory.** Every session tracks work in status docs and is
  **always ready to end and hand off to a fresh session** when context fills.
- **Subagents** research docs and best practices, keeping the main agent's context clean.
  The main agent orchestrates.
- **Side quests don't derail the task.** When something out of scope or context-heavy comes
  up, the main agent writes a ticket draft or a minimal findings summary for later. Or, if
  possible, it has a **subagent use skills plus a summary to write a full ticket draft** for
  review.

## Round 4 — 2026-09-27: general way of working

### Session start

| Case | Procedure |
|---|---|
| **Resuming** an effort | Always the same: read `00-STATUS.md` → check git state → confirm next steps. |
| **New** effort | Ask if there's a ticket or docs to read → read them → suggest next steps (usually research) → run the research skill for preliminary findings → create `00-STATUS.md`. |

### Handoff timing

- The agent should raise it itself.
- Current practice: **~200K tokens** is the "about time to end" mark. Look for a natural
  stopping point. **Don't start a task that will clearly overshoot.**
- Evolving topic. **User wants to research best practice before we fix a rule.**

### Status doc

- **Update:** at milestones · **before risky or long-running operations (also commit then)**
  · at the end of every session, before any session-close work.
- **Shape:** slim. A glossary: *where we are, what's next,* and pointers to topic files.
  No content that belongs in a topic file.

### Parallel work (focus: multiple agents on the *same* effort)

- **Scope first** so agents work on isolated parts of the code.
- **Main agent orchestrates:** reviews the work, merges worktrees into the feature branch,
  resolves diffs, and **asks the user on judgement calls.**
- Not everything should be parallelized. It's a judgement call; iterate as we learn.

### Subagents and model escalation

- Purpose: **throughput**. Parallel research agents (different docs or codebases → summaries
  for the main agent to review), or parallel implementation of isolated parts of a ticket,
  merged by the main agent.
- **Cheap model first, escalate when needed.** E.g. Sonnet does research; if the main agent's
  review finds weak or contested claims, an Opus agent digs deeper or validates.

### Side quests

| Nature | Destination |
|---|---|
| Tightly coupled to the current effort | The current effort folder |
| Broad (dep upgrades, a pervasive outdated pattern) | `work/inbox/` |

Written **when found**, **mentioned when written**, and **all listed again at session close**.

### Ticket drafts

- Standard: the beta team's **ticket skill** (added to `skills/` 2026-09-27). Its writing
  rules: outcomes not implementation, conclusion not path, link evidence, omit empty sections,
  trim pass.
- User's additions: succinct, don't over-explain, **tables where appropriate**, describe
  outcomes rather than how, link useful info, **call out gotchas**, short and direct.
- To reconcile: the ticket skill writes drafts to `.agents/tickets`. Side-quest drafts go to the
  effort or `work/inbox/`.

### Approval required (always)

Jira writes · Confluence edits · posting comments · writes to shared `knowledge/`/`state/` ·
deleting files · running against non-local envs · pushing. **Generally: anything destructive,
anything public, or anything that would appear to come from the user.**

### Headless / unattended runs

- **Allowed:** open **draft** PRs, write findings to local files, possibly notify users
  (a messaging mechanism is TBD).
- **Never:** destructive actions, or changes to `main` or external systems (Datadog,
  Confluence, Jira, etc.).
- **Judgement calls:** best effort, **explicitly flagged for review** with the details a
  reviewer needs.
- Principle: **unmonitored agents create things for review. They don't change real systems.**

### Cross-cutting principle: communication style

All agent writing, and **especially anything that goes beyond the user** (tickets, ADRs,
messages, PR descriptions, review comments), is **succinct and direct**. Plain terms, common
language, no metaphors, no over-explaining. Technical, not dumbed down. Agents tend to be verbose;
the framework must push against that everywhere.

## Round 5 — 2026-09-27: research and discovery

### Source order (topic-dependent; skip what doesn't apply)

`knowledge/` → local repos → other teams' repos → **the internal API docs portal** → Confluence → Jira →
Datadog → Slack → web.

- **Clone another team's repo** when the search is deep (many files) or the repo will be
  referenced repeatedly.
- **Best practice:** org and team precedent first (Confluence ADRs, team and architecture
  spaces), then the web.

### Company resources (beta team; per-team config in the template)

- **Internal API docs portal:** the main source for docs and API specs. Agents have struggled to connect
  in the past, so it needs configuring.
- **ADRs:** team Confluence spaces and the architecture space.

### Local clones

- User keeps repos in `~/projects`, some nested by team (`~/projects/<team>/`). **Varies per
  person**, so it's per-user config.
- **The framework keeps clones current before they're trusted.** Any agent entering a
  workspace for research reviews the **latest default branch** first.
- **Concurrency:** multiple sessions must not step on each other in shared clones.
  Needs a strategy discussion (worktrees, read-only fetch, lock?).

### Tools and fallbacks

| Need | Preferred | Fallback |
|---|---|---|
| Single-repo lookup / understanding code | **Local clone** | — |
| Search across the org / many repos | `gh search code` | Local clones |
| Jira / Confluence | Atlassian MCP | User pastes, exports, or **links** |
| Metrics, logs, monitors | Datadog MCP | User link or screenshot |
| Docs, API specs | Internal API docs portal | — |
| External docs / best practice | Web (subagent), **after** org precedent | — |

**Unavailable tools → always ask the user.** Never proceed without `gh`, the Atlassian MCP,
the Datadog MCP, or the API docs portal access without asking first.

### Output

- `NN-<topic>.md` in the effort, claims linked to evidence, **with a confidence score.**
- Short summary back to the main agent.

### Opus re-check triggers

Conflicting sources · a claim a design decision depends on · a sampled citation the main
agent fails to verify · a low confidence score.

### Lifting findings to `knowledge/`

**Undecided.** The user leans towards proposing at session close (less disruptive), but would
switch to proposing immediately if knowledge is being missed. **Wants an informed decision,
not a guess** (e.g. measure what session-close misses during the beta).

### Done criteria

No fixed rule. **Enough to make a decision you're confident in, with little risk or
well-understood risks.** The user weighs in heavily here. The agent should present the
decision, the confidence, and the known risks, and let the user call it.

### Architectural realization (user)

- The template needs a **setup workflow/skill** that asks for team-specific values and
  writes the team's base config: Confluence space, Jira project key, team-owned repos,
  the API docs portal, ADR locations, and more.
- Points towards a **central team repo** holding the team's core config and state, plus a way
  for members to **onboard and sync**. To be discussed.

## Round 6 — 2026-09-27: architecture (proposal accepted)

Three layers:

| Layer | Holds | Created by |
|---|---|---|
| **Framework repo** (generic, versioned) | Skills, AGENTS.md, hooks, scripts, setup skill, templates | Us |
| **Team repo** (one per team) | `team.yaml` (Jira key, Confluence space, repos, API docs portal, ADR locations, default branches), `knowledge/`, `state/`, `efforts/`, team-specific `skills/` | Setup skill |
| **Personal repo** (one per person, private) | `me.yaml` (projects dir, model prefs, notifications), `work/`, `work/inbox/`, `overlays/` | Onboarding |

Decisions:

- **Framework and team content are separate repos.** The team pins a framework version and
  upgrades deliberately. No template fork.
- **Sync is split by area:** `knowledge/` and `skills/` change via PR. `state/` and `efforts/` are
  pushed directly, rebased, linear.
- **Personal layer = separate private repo.**
- **Skills read overlays** from the personal repo (e.g. `overlays/code-review.md`). *(Proposed,
  implied by the accepted design; mechanism to spec.)*
- **Tool-agnostic:** AGENTS.md is the source; CLAUDE.md imports it; a link script installs into
  each tool. *(Proposed; to spec.)*
- **Lineage: keep agent-space as-is.** Start fresh. Migrate selectively later.

## Round 7 — 2026-09-27: mechanisms, licensing, name

### Clone freshness and concurrency (accepted)

- **Agents never check out or pull in the user's main clone.** It belongs to the user.
- **Research:** `git fetch`, then read the default branch without checking it out
  (`git grep origin/<default>`, `git show origin/<default>:path`, or a detached worktree for
  deep reads). **Record the SHA read** in the research output.
- **Code changes:** always a dedicated worktree on its own branch. Parallel subagents each get
  their own worktree off the feature branch; the orchestrator merges.
- **Missing clone:** clone into the projects dir from `me.yaml`, then the same rules apply.

### Automation (accepted)

| Routine | Mechanism |
|---|---|
| session-close | Semi-automatic, since it needs approval. `PreCompact` hook forces a `00-STATUS.md` update · `SessionEnd` hook leaves an "unclosed session" marker · the next `SessionStart` offers session-close. |
| knowledge-audit | **CI in the team repo:** a monthly GitHub Action runs headless and opens a **PR of proposals**. |
| Handoff trigger | Deferred to research (status line context meter + threshold hook is the likely shape). |

### Supported tools (v1): **Claude Code (desktop + CLI) and GitHub Copilot only.**

No Cursor in v1. More later. *(Research needed: Copilot's support for AGENTS.md, skills and hooks,
to see which mechanisms carry over and which fall back to instructions.)*

### Licensing and ownership — hard constraint

**The framework is open source and owned personally by the user. It must contain NO company-specific
content.** Everything company- or team-specific (company and team names, internal portals, Jira project keys,
Datadog org details, team-specific local paths) lives only in a team repo or personal repo, or appears as a
clearly generic example.
→ Existing skills need generalizing before they go in the framework. E.g. the beta team's ticket skill becomes a
generic ticket skill; its team templates move into team config.

### Review gate (accepted, extended)

- A `review` script launches a **separate headless CLI process** with a fixed prompt (branch,
  SHA, pointers to the ticket and effort). The working agent can't frame the review.
- The review is written to the effort with the SHA it covers.
- **Trigger is configurable at team and personal level:** `on-demand` · `after-commit` ·
  **`pre-push`** (user's preferred default over re-reviewing on every commit).
- The PR skill requires a review that covers the SHA being pushed or opened.

### Commit / PR convention enforcement

**Agent-side only for v1** (skills + commit-message hook). No project-repo CI checks.

### Name

- Loved **hivemind** and **cortex**, but both are in use in this space.
- Next favourite: **crew** or a variation.
- Checked 2026-09-27 (npm registry, `gh search repos`): `crewmind`, `crewbrain`, `crewbase`,
  `crewlog`, `crewhq` are free on npm, and GitHub only has zero-star repos. `crewbrain` has no
  GitHub repos at all. `crewdeck`, `wardroom`, `muster`, `bosun` are taken on npm.
- Gotcha: **CrewAI** is a well-known multi-agent framework, so a `crew*` name risks confusion in the agent space.

## Round 8 — 2026-09-27: unit tests and QA

### Unit tests

- **TDD is the default.**
  - Bugs: reproduce with a failing test, then fix.
  - Features: start with tests for the ACs, then add edge cases and important business logic.
- **Quality:** standard best practice. Test behaviour, not implementation; don't mock what you own.
  **Observe the repo's house style, but call out poor tests and bad patterns.**
- **An extension point.** The user expects to iterate here; team and personal overlays apply.
- **Coverage is a signal, not a gate.**

### QA testing (v1)

- **Manual verification, locally and in dev once deployed.** E2E test writing comes in later versions.
- **Agents may drive a browser against local and dev only.** Any other env needs approval,
  per round 4. *(So dev is an exception to "non-local needs approval", for QA verification.)*
- **Evidence:** screenshots, traces, Datadog links.

## Round 9 — 2026-09-27: ADRs

- **Split accepted:** the framework `adr` skill holds the method plus a default template (Issue →
  Context → Decision → Options → Consequences) and **synthetic** examples. Team config supplies the
  publish target, group labels, reviewer rules and the team's own gold-standard examples. The company's
  architecture-space content and a colleague's example ADRs stay out of the framework.
- **Add solid architecture knowledge to the base skill:** decision quality, not only format.
  **Needs a dedicated session.**
- **Support both** Confluence-published ADRs **and** in-repo ADRs (`docs/adr/NNNN-*.md`).
  Configured per team.
- **Suggest an ADR at the moment** a design decision with real alternatives is made (not batched to
  session-close).
- Nothing else missing today. **Standing rule for the build: whenever a skill is created or ported,
  take the chance to improve it** (not a straight copy).

## Round 10 — 2026-09-28: decisions after research

- **Handoff bands (tokens in context):** **125K** look for a stopping point · **150K** no new
  multi-step work, propose a handoff at the next checkpoint · **200K** write `00-STATUS.md` and end.
  The user lowered them from the re-check's proposal of 150K / 200K / 300K.
  *(The window-% parts of the bands (50 / 70 / 90%, whichever comes first) were not rediscussed.
  Confirm at spec time.)* Mechanism: a transcript-reading hook on `UserPromptSubmit` +
  `PostToolBatch` that injects the band via `additionalContext`. `PreCompact` is the backstop.
  See [03](03-research-handoff-trigger.md).
- **Ticket drafts:** go in the **effort folder**, falling back to **`work/inbox/`**. Drop
  `.agents/tickets` (remove it from the ported ticket skill).
- **Headless notifications:** **GitHub @mention** on the draft PR, **falling back to a Slack** message.
- **Overlays, two levels:**

  | Level | May add | May override |
  |---|---|---|
  | Team (team repo) | yes | yes |
  | Personal (personal repo) | yes | **no** |

  The team can tailor the framework baseline, and every team member shares the team baseline.

## Round 11 — 2026-09-28: tool scope, repo split, pinning

- **Tool scope narrowed:** v1 targets **Claude Code (desktop + CLI) and Copilot CLI only.** IDE
  integrations (VS Code, JetBrains) are out of scope "and likely will never be included". The Copilot
  cloud agent is out too.
- **Repo split: drop the separate personal repo.** One team repo holds everything team-side:
  committed `people/<handle>/` (`overlays/`, `inbox/`, `work/`), pushed directly like `state/` and
  `efforts/`, plus a gitignored `.local/` for `me.yaml` (machine paths, prefs) and the rare private item.
  Reason: no per-tool access setup, one repo to sync, promotion is a `git mv`. Personal scratch is
  "rarely" sensitive. See [05](05-research-repo-split.md). Replaces the round 6 personal-repo layer.
- **Pinning:** the framework is **vendored** into the team repo by a script (`folkways upgrade <version>`).
  **Overlays must never touch framework files,** so an upgrade can never clobber team or personal
  customizations.
- **Handoff bands: tokens only** (125K / 150K / 200K). The window-% caps are dropped.

## Round 12 — 2026-09-28: language, distribution, instructions import

- **Instructions import:** drop the `<!-- folkways:begin/end -->` managed block. The team `AGENTS.md` gets
  one line, `@.folkways/instructions.md`, and is otherwise fully team-owned. Verified: Copilot CLI expands `@`
  imports in `AGENTS.md` (not in `*.instructions.md` or `GEMINI.md`).
- **Language: TypeScript, not bash.** Reason: maintainability. The team and most of the industry know
  TypeScript or Python better than shell, and bash can't read `team.yaml` without `yq`. Python lost because
  YAML isn't in its standard library and macOS ships 3.9.
- **Native TS considered, then dropped:** Node 24 (current LTS) runs `.ts` directly, but it refuses to strip
  types in `node_modules`, and the user wants `yaml` as a normal installed dependency, not a copied one.
- **Distribution: npm package `@folkways/cli`** (scope claimed by Summer). Reasons: the company's enterprise
  GitHub accounts can't clone repos outside the org, but npm is reachable through the company Artifactory
  mirror; `npx` sets up a repo; upgrades are dependency bumps.
- **Pinned per team repo:** the team repo has a `package.json` with an exact version and a committed lockfile,
  so the whole team runs one version. Hooks run `node_modules/.bin/folkways hook <name>`.
- **Text is still copied into the repo** by `folkways sync` (skills must live in `.claude/skills/`), so
  upgrade PRs show framework changes. Edit detection and overlays are unchanged.
- **Later:** publish a GitHub Action that runs on dependency-update PRs, detects a `@folkways/cli` bump, and
  runs the sync automatically.
- **Package managers:** team repos use **whatever the team uses**: npm or pnpm, detected from the
  `packageManager` field or lockfile. The beta team uses npm today; the user is pushing for pnpm. The
  framework repo itself uses pnpm with supply-chain settings (7-day minimum release age, no provenance
  downgrades, registry-only subdependencies, no unapproved build scripts) and a matching Dependabot cooldown.
- **CI actions stay pinned to commit SHAs** (tags can be moved; SHAs can't). The repo requires SHA pinning.
- **GitHub org:** `sumsolution` for now; a dedicated `folkways` org may come later.
