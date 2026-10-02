# folkways

Your team's way of working, written down once and shared by every teammate's agent.
Shared skills, instructions and hooks for Claude Code and Copilot CLI, versioned like code.

> **Status: pre-release.** v0.1.0, not on npm yet. `init`, `sync` and `upgrade` work. Skills and prompts sync too.
> Hooks, `review`, `run` and `doctor` are next. [What works today](#status).

## What it is

Small engineering teams (3 to 8 people) using AI coding agents each grow private habits: one person's
agent writes great commits, another's does careful reviews, nobody's does both. folkways turns one team's
way of working into shared, versioned **skills, instructions and hooks**, so every teammate's agent works
the same way. Each person can still add their own touches on top.

It is a small CLI (`@folkways/cli`) plus the text files it keeps in sync. No server and no account:
just files in a git repo your team already knows how to review.

## How it works, in 60 seconds

- **Two repos.** This one (the framework, published to npm as `@folkways/cli`) and one **team repo**
  per team. Everyone runs their agent from the team repo.
- **Pinned, not floating.** The team repo pins an exact `@folkways/cli` version in `package.json`, so
  the whole team runs the same one.
- **Sync copies text in.** `folkways sync` copies framework instructions, skills and prompts
  into the team repo, compiles team overlays into the skills and writes a checksummed manifest. If someone edited a framework file, sync stops
  and names it instead of overwriting it.
- **Overlays, not forks.** Teams change framework skills through `overlays/<skill>/OVERLAY.md` (add or
  override a section), which `sync` compiles in. People add their own in `people/<handle>/overlays/`.
  Upgrades never touch them.
- **Upgrades are PRs.** `folkways upgrade <version>` makes a branch, installs, syncs and commits. You
  read the diff like any other change.
- **A few rules have teeth.** Pushing asks first (a permission rule, built; Copilot CLI support unverified).
  Commit format, the fresh-process review gate and "don't touch my main clone" are enforced by the
  `git-guard` hook (built; the `folkways review` command that writes reviews is coming soon).

More: [How it works](docs/guide/how-it-works.md).

## Quickstart

Once `@folkways/cli` is published (coming soon):

```bash
npx @folkways/cli init my-team        # or: pnpm dlx @folkways/cli init my-team
cd my-team
```

Then open Claude Code or Copilot CLI in `my-team` and run the `setup` skill to fill in
`folkways.yaml`.

Today, from a clone of this repo (Node 24+):

```bash
pnpm install
node packages/cli/src/cli.ts init ../my-team --no-install
```

That scaffolds the team repo, syncs the framework files and makes the first commit. Full walkthrough:
[Getting started](docs/guide/getting-started.md).

To use the skills, instructions and hooks in any repo without a team repo, install the Claude Code
plugin, once `@folkways/claude-plugin` is published:

```bash
claude plugin marketplace add sumsolution/folkways
claude plugin install folkways@folkways
```

From a clone, build it and load it in place: [packages/claude-plugin](packages/claude-plugin/README.md).

## Status

| Piece | State |
|---|---|
| `init`, `sync`, `upgrade`, `version` | built, tested |
| Push approval rule in `.claude/settings.json` | built |
| Team repo scaffold (`folkways.yaml`, folders, READMEs) | built |
| Framework instructions (`.folkways/instructions.md`) | built |
| 16 skills (`setup`, `onboard`, `start`, `review`, `commit`, `pr`, …) and the review prompt | built; synced with team overlays compiled in |
| Hooks (`git-guard`, `context-band`, `session-start`, `precompact`) | built, wired into `.claude/settings.json` by `sync`; ship with the npm release |
| Claude Code plugin (skills, instructions, hooks; no team overlays) | built; npm release coming soon |
| `review`, `run`, `doctor` | coming soon; today they say "not built yet" |
| npm release | coming soon |

The design record is in [docs/design/](docs/design/00-STATUS.md).

## License

[Apache-2.0](LICENSE). Copyright 2026 Summer Jordan. The license grants no rights to the name folkways.

## The guide

- [Concepts](docs/guide/concepts.md): team repo, skills, overlays, hooks, folkways vs mores
- [How it works](docs/guide/how-it-works.md): distribution, sync, ownership, upgrades, the review gate
- [Getting started](docs/guide/getting-started.md): for a team lead and for a member
- [The name](docs/guide/the-name.md): Sumner, and the names we didn't pick

## The name

In 1906 the sociologist William Graham Sumner called a group's customary ways of doing things its
**folkways**, and the few norms it actually enforces its **mores**. That's the design: a team shapes its
folkways (skills, overlays, templates), and a handful of mores (ask before pushing, a fresh review)
are backed by settings and hooks. [More on the name](docs/guide/the-name.md).

## Working on folkways

See [AGENTS.md](AGENTS.md) for layout, commands and rules. Short version: Node 24+, pnpm,
`pnpm run check && pnpm test` before every commit. Issues and ideas are welcome; we build this
together. 🙌
