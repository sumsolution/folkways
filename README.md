# folkways

Your team's way of working, written down once and shared by every teammate's agent.
Shared skills, instructions and hooks for Claude Code and Copilot CLI, versioned like code.

> **Status: pre-release.** v0.1.0, no release yet ([install from source](#install)). `init`, `sync`, `upgrade` and `doctor` work. Skills, prompts and hooks sync too.
> `review` and `run` are next. [What works today](#status).

## What it is

Small engineering teams (3 to 8 people) using AI coding agents each grow private habits: one person's
agent writes great commits, another's does careful reviews, nobody's does both. folkways turns one team's
way of working into shared, versioned **skills, instructions and hooks**, so every teammate's agent works
the same way. Each person can still add their own touches on top.

It is a small CLI (`folkways`, one binary with no runtime dependencies) plus the text files it keeps in sync. No server and no account:
just files in a git repo your team already knows how to review.

## How it works, in 60 seconds

- **Two repos.** This one (the framework and the `folkways` binary) and one **team repo**
  per team. Everyone runs their agent from the team repo.
- **Pinned, not floating.** The team repo records the version it's synced from in `.folkways/manifest.json`,
  and the session start warns anyone whose `folkways` differs, so the whole team runs the same one.
- **Sync copies text in.** `folkways sync` copies framework instructions, skills and prompts
  into the team repo, compiles team overlays into the skills and writes a checksummed manifest. If someone edited a framework file, sync stops
  and names it instead of overwriting it.
- **Overlays, not forks.** Teams change framework skills through `overlays/<skill>/OVERLAY.md` (add or
  override a section), which `sync` compiles in.
  Upgrades never touch them.
- **Upgrades are PRs.** `folkways upgrade <version>` makes a branch, syncs and commits; run it with
  that version installed. You
  read the diff like any other change.
- **A few rules have teeth.** The `git-guard` hook asks before a push per `push.brain` / `push.app` and keeps
  `knowledge/` off the default branch (built; Copilot CLI support for `ask` unverified). Commit format, the
  fresh-process review gate and "don't touch my main clone" are enforced by the same hook (built; the `folkways review` command that writes reviews is coming soon).

More: [How it works](docs/guide/how-it-works.md).

## Install

Each person installs `folkways` once; the team repo's hooks call it on `PATH` and are off without it. macOS and
Linux, amd64 and arm64.

> **No release yet.** Until the first one, build from a clone (Go 1.26+):
> `go build -o ~/.local/bin/folkways ./packages/folkways/cmd/folkways`. The commands below work from the
> first release on.

Homebrew:

```bash
brew install sumsolution/tap/folkways
```

The install script ([read it first](install.sh)) puts the binary in `~/.local/bin`, after checking it against the
release's `checksums.txt`:

```bash
curl -fsSL https://github.com/sumsolution/folkways/releases/latest/download/install.sh | sh
```

To pick a version or a directory, set the variables after the pipe, where `sh` sees them:

```bash
curl -fsSL https://github.com/sumsolution/folkways/releases/latest/download/install.sh | FOLKWAYS_VERSION=0.2.0 FOLKWAYS_INSTALL_DIR=/usr/local/bin sh
```

With Go 1.26+:

```bash
go install github.com/sumsolution/folkways/packages/folkways/cmd/folkways@latest
```

Or download an archive from [Releases](https://github.com/sumsolution/folkways/releases), compare
`shasum -a 256 <archive>` with its line in `checksums.txt`, and check where it was built with
`gh attestation verify <archive> --repo sumsolution/folkways`.

Check it with `folkways version`. Tools started from the Dock or a launcher use your login shell's `PATH`, so
put the install directory there. Cloud agents install it in their environment's setup script; without it, hooks
there do nothing.

## Quickstart

```bash
folkways init my-team
cd my-team
```

That scaffolds the team repo, syncs the framework files and makes the first commit. Then open Claude Code or
Copilot CLI in `my-team` and run the `setup` skill to fill in `folkways.yaml`. Full walkthrough:
[Getting started](docs/guide/getting-started.md).

The Claude Code plugin won't be published; it's being removed ([ADR 0012](docs/adr/0012-cli-only.md)). Until then,
build it from a clone and load it in place: [packages/claude-plugin](packages/claude-plugin/README.md).

## Status

| Piece | State |
|---|---|
| `init`, `sync`, `upgrade`, `doctor`, `version` | built, tested |
| Push approval rule in `.claude/settings.json` | built |
| Team repo scaffold (`folkways.yaml`, folders, READMEs) | built |
| Framework instructions (`.folkways/instructions.md`) | built |
| 16 skills (`setup`, `onboard`, `start`, `review`, `commit`, `pr`, …) and the review prompt | built; synced with team overlays compiled in |
| Hooks (`git-guard`, `context-band`, `session-start`, `precompact`) | built, wired into `.claude/settings.json` by `sync`; run the `folkways` binary on `PATH` |
| Claude Code plugin | being removed ([ADR 0012](docs/adr/0012-cli-only.md)) |
| `review`, `run` | coming soon; today they say "not built yet" |
| Release: GitHub binaries, install script, Homebrew tap | built; first release coming soon |

The design record is in [docs/design/](https://github.com/sumsolution/folkways-command/blob/main/state/framework.md).

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

See [AGENTS.md](AGENTS.md) for layout, commands and rules. Short version: Go 1.26+, Node 24+, pnpm,
`pnpm run check && pnpm test` before every commit. Issues and ideas are welcome; we build this
together. 🙌
