# Concepts

The handful of ideas folkways is built from. Mechanics are in [How it works](how-it-works.md).

## The team repo

One git repo per team. Everyone opens Claude Code or Copilot CLI **in the team repo**, so every agent
starts from the same instructions, skills and config. Code changes still happen in your project repos,
in git worktrees; the team repo is where the agent's context lives.

Set and setting matter: the team repo is where every session starts.

`init` creates it with this layout:

| Path | Holds | Lands by |
|---|---|---|
| `folkways.yaml` | Team config: tracker, repos, research sources, git rules, review trigger, models | PR |
| `knowledge/` | Durable facts, each with `Last verified:` and how | PR |
| `state/` | Live trackers, one file per track | direct push |
| `efforts/<KEY>-<slug>/` | Shared pieces of work, each with a `00-STATUS.md` | direct push |
| `people/<handle>/` | One folder per person: `overlays/`, `inbox/`, `work/` | direct push |
| `overlays/<skill>/` | Team changes to framework skills | PR |
| `.folkways/` | Framework files. `sync` owns them; don't edit | PR (upgrades) |
| `.local/` | Machine-local config (`me.yaml`), never committed | — |

Direct pushes are rebased and linear. Every push, direct or not, waits for the person's approval.

## Solo repos

One person can use folkways through the Claude Code plugin, from their own **brain repo**: like a team
repo, it's where sessions start and knowledge, state and efforts build up across every project, and it
holds no code. Code changes happen in worktrees of the project repos. A solo repo is the team layout
with one layer removed: what lives in `people/<handle>/` in a team repo lives at the repo root (`work/`, `inbox/`, `overlays/`). There's no handle, no `efforts/` and nothing to promote.
Skills call that folder the **personal root**.

A repo is set up when it has a `folkways.yaml`. Without one, the plugin stays quiet: the first session
points to the brain repo and offers `setup` once, and after that it says nothing ([spec §2.3](../design/06-v1-spec.md)).

## Knowledge, state, efforts, people

Four kinds of shared memory, split by how long they live and who trusts them:

- **Knowledge** is slow and reviewed: facts the team relies on, changed by PR.
- **State** is fast: what's in flight right now. Pushed directly.
- **Efforts** are units of work. Each has a `00-STATUS.md` that says where we are, what's next and where
  the detail lives. A fresh session reads it and picks up.
- **People** hold personal work. A personal effort in `people/<handle>/work/` becomes shared with one
  `git mv` into `efforts/`. Side quests and drafts land in `people/<handle>/inbox/`.

## Skills

A skill is a folder with a `SKILL.md` that teaches the agent one method: how we review, how we commit,
how we research. Both tools read skills from `.claude/skills/`. Frontmatter is `name` and `description`
only, so the same file works in both.

Framework skills are synced in from the package. Teams can add their own skills next to them. Sixteen
ship: `setup`, `onboard`, `start`, `status`, `handoff`, `session-close`, `side-quest`, `research`,
`ticket`, `adr`, `tests`, `commit`, `pr`, `review`, `qa` and `knowledge-audit`
([spec §6](../design/06-v1-spec.md)).

## Overlays

The way teams and people customize a framework skill without editing it.

| Level | File | Add rules | Replace a section |
|---|---|---|---|
| Team | `overlays/<skill>/OVERLAY.md` | yes | yes, by section ID |
| Person | `people/<handle>/overlays/<skill>/OVERLAY.md` | yes | no |

Framework skills mark replaceable sections with an ID, like `## Commit format {#commit-format}`. An
overlay has `## Add` for extra rules and `## Override {#commit-format}` to replace that section.

Team overlays are **compiled in** by `folkways sync`: an override replaces its section in the synced
skill, and an add goes under `## Team additions`. An unknown ID stops the sync before anything is
written. Rerun `sync` after you change an overlay. Personal overlays are attached at runtime: each
skill's step 0 reads `folkways.yaml` and `.local/me.yaml`, then attaches your overlay through `.local/me`, a
link to `people/<handle>/` that `onboard` creates and `sync` keeps. The framework's precedence text says
personal rules may only add, and `sync` lints them. Reasoning:
[ADR 0001](../adr/0001-overlay-enforcement.md).

So the team gets one shared baseline, and each person can still add their own touches on top.

## Hooks

Small commands the tools run at set moments: before a shell command, after each prompt and tool batch,
at session start, before compaction. They live in `.claude/settings.json`, which **both Claude Code and
Copilot CLI read**.

Each hook runs the pinned framework binary, `folkways-hook`, from `node_modules/`
([spec §7.1](../design/06-v1-spec.md)): `git-guard` (commit format, review before push, no checkouts in
your main clone), `context-band` (announces 125K, 150K and 200K tokens so you hand off in time),
`session-start` (points the agent at the `start` skill) and `precompact` (logs a missed handoff).
`.claude/settings.json` also holds one framework rule: `git push` asks first.

`folkways.yaml` decides what they check. `git.commit_convention` picks Angular, your own pattern or no
commit rule; `hooks:` turns a hook off. Without a `folkways.yaml` the hooks do nothing.

## Folkways and mores

Most of folkways is **folkways**: customs the team shapes and changes freely. Skills, overlays, templates,
`folkways.yaml`. If your team does it differently, write an overlay.

A few things are **mores**: norms the team agrees are worth enforcing, so they're backed by settings and
hooks, not just instructions. They live in config (`folkways.yaml`, framework settings), which changes by
PR, so an overlay can't loosen them.

- Ask before every push. Built: a permission rule, enforced in Claude Code (whether Copilot CLI honors
  `permissions.ask` is unverified).
- Framework files aren't edited. Built: `sync` checks their checksums.
- Team overlay overrides apply. Built: `sync` compiles them in and fails on an unknown ID.
- Personal overlays only add. `sync` warns when one doesn't; `doctor` will report it (coming soon).
- Commits follow the team's format. Built: `git-guard`.
- A fresh-process review covers the commit before a push, with the `pre-push` trigger. Built: `git-guard`
  blocks the push of a project repo; `folkways review`, which writes the review, is coming soon.
- Agents never check out, pull or reset in your main clones. Built: `git-guard`.

The full table is in [ADR 0001](../adr/0001-overlay-enforcement.md). The idea comes from the book the
project is named after; see [The name](the-name.md).

## Headless runs

Sometimes nobody's watching: `folkways run` (coming soon) or a CI job, like the monthly knowledge audit
(the `knowledge-audit` skill ships; the workflow that runs it is coming soon). Headless runs create things **for review only**: local files, draft PRs, an
@mention.
Never `main`, never destructive actions, never writes to external systems. Every judgement call is
flagged in the output, so a person decides.

## Sessions

A session has a shape:

1. **Start.** Read `folkways.yaml` and `.local/me.yaml`. Resume an effort from its `00-STATUS.md`, or start
   a new one from a ticket.
2. **Work.** Subagents for throughput, cheap model first, stronger one for hard calls. Side quests get
   written down, not chased.
3. **Hand off in time.** At 125K tokens look for a stopping point; at 150K start nothing new; at 200K
   write `00-STATUS.md` and end.
4. **Close.** `session-close` lifts what we learned into `knowledge/`, `state/` and the
   status doc.

These rules ship today in `.folkways/instructions.md`, which the team `AGENTS.md` imports.
