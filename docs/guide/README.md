# The folkways guide

Everything a curious engineer needs to understand folkways, in four short pages. Each one takes under
five minutes.

| Page | Read it for |
|---|---|
| [Concepts](concepts.md) | The vocabulary: team repo, knowledge, state, efforts, people, skills, overlays, hooks, folkways vs mores, headless runs |
| [How it works](how-it-works.md) | The mechanics: how files get from the `folkways` binary into your team repo, who owns what, upgrades, the review gate |
| [Getting started](getting-started.md) | Doing it: set up a team repo as a lead, join one as a member |
| [The name](the-name.md) | Sumner, folkways and mores, and the names we didn't pick |

New here? Read Concepts, then Getting started. Evaluating the design? How it works, then the
[ADRs](../adr/README.md).

## Where things stand

folkways is released ([install](../../README.md#install)). The CLI's `init`, `sync`, `upgrade` and `doctor` work, the 16
skills ship, and `sync` wires in the hooks. `review` and `run` aren't built yet. Each page marks what's coming
soon, and the README's [status table](../../README.md#status) has the details.

## Other docs

- [Voice](../voice.md): how these docs sound, and how to keep them that way
- [ADRs](../adr/README.md): the framework's decisions, with the options we didn't take
- [AGENTS.md](../../AGENTS.md): working on the framework itself
