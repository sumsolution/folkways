# The folkways guide

Everything a curious engineer needs to understand folkways, in four short pages. Each one takes under
five minutes.

| Page | Read it for |
|---|---|
| [Concepts](concepts.md) | The vocabulary: team repo, knowledge, state, efforts, people, skills, overlays, hooks, folkways vs mores, headless runs |
| [How it works](how-it-works.md) | The mechanics: how files get from npm into your team repo, who owns what, upgrades, the review gate |
| [Getting started](getting-started.md) | Doing it: set up a team repo as a lead, join one as a member |
| [The name](the-name.md) | Sumner, folkways and mores, and the names we didn't pick |

New here? Read Concepts, then Getting started. Evaluating the design? How it works, then the
[v1 spec](https://github.com/sumsolution/folkways-command/blob/main/knowledge/design/06-v1-spec.md).

## Where things stand

folkways is at **v0.1.0 and unpublished**. The CLI's `init`, `sync` and `upgrade` work and are tested, and the 16 skills ship.
The hooks are built and `sync` wires them in. There's also a Claude Code plugin with the skills,
instructions and hooks for repos without a team repo ([packages/claude-plugin](../../packages/claude-plugin/README.md)).
Each page marks what's coming soon, and the
[status doc](https://github.com/sumsolution/folkways-command/blob/main/state/framework.md) has the details.

## Other docs

- [Voice](../voice.md): how these docs sound, and how to keep them that way
- [Design record](https://github.com/sumsolution/folkways-command/blob/main/state/framework.md): status, decisions and research
- [AGENTS.md](../../AGENTS.md): working on the framework itself
