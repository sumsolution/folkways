# 0004. folkways is licensed under Apache-2.0

| | |
|---|---|
| Status | accepted |
| Date | 2026-09-30 |
| Owner | Summer Jordan |

## Issue

folkways is open source, but every `package.json` said `UNLICENSED`, which grants nothing. Hooks PR 5
publishes the platform packages, and each one needs a `license`. Which license covers the code, and do the
text assets that `sync` copies into team repos (skills, instructions, prompts) need a different one?

## Context

- Team repos are usually private company repos. `sync` copies text assets into them, and teams extend
  them with overlays ([spec §3](https://github.com/sumsolution/folkways-command/blob/main/efforts/v1-design/06-v1-spec.md)). Company open-source policies allow permissive
  licenses and often ban AGPL ([15](https://github.com/sumsolution/folkways-command/blob/main/efforts/license/15-research-license.md), comparison table).
- The closest comparable projects each use one permissive license for code and skills: Anthropic's skills repo and
  OpenAI Codex use Apache-2.0; spec-kit and obra/superpowers use MIT ([15 §2](https://github.com/sumsolution/folkways-command/blob/main/efforts/license/15-research-license.md)).
- The name `folkways` is pending a trademark check.
- Dependencies: `mvdan.cc/sh/v3` (BSD-3-Clause) and `gopkg.in/yaml.v3` (MIT and Apache-2.0, with a `NOTICE`)
  are compiled into the hook binary. The CLI has no runtime dependencies.
- The beta team works at the owner's employer.

## Decision

**Apache-2.0 for the whole repo, code and text assets, in one root `LICENSE`.** It adds an explicit patent
grant and says the license grants no trademark rights, at the cost of a longer text than MIT. Without a
folkways `NOTICE` file, a team repo that receives synced skills owes nothing extra: no headers per file.

## Options considered

| Option | For | Against | Trade-off |
|---|---|---|---|
| **Apache-2.0** | Patent grant; trademark carve-out; allowed by company policies; matches Anthropic's skills repo and OpenAI Codex | Longer text; a `NOTICE`, if ever added, travels with every copy | A longer license for patent and trademark clarity |
| MIT | Shortest; allowed everywhere | No patent grant; silent on trademarks | Simplest, with less protection for adopters |
| Apache-2.0 code, CC-BY-4.0 for skills and docs | Suits prose | Skills are executed like code; Creative Commons advises against CC licenses for software; two licenses to track per file | Precision nobody needs, at a cost every contributor pays |
| MPL-2.0 or GPL/AGPL | Keeps changes open | Copyleft attaches to files copied into company repos; AGPL is banned at some companies | Blocks the adoption the project exists for |

## Consequences

- Root `LICENSE` (the verbatim Apache-2.0 text); `license: "Apache-2.0"` in every `package.json`. `pnpm pack`
  adds the root `LICENSE` to the CLI tarball; `pnpm run smoke` checks it.
- Hooks PR 5: the platform package generator writes `license`, copies `LICENSE`, and adds a
  `THIRD_PARTY_NOTICES` file with the notices of the Go dependencies (including yaml.v3's `NOTICE`) and the Go standard library.
- The unscoped `folkways` placeholder on npm gets a `0.0.2` under Apache-2.0; `0.0.1` stays `UNLICENSED`.
- **Before the repo goes public:** the owner checks the employment agreement and the employer's open-source
  policy ([15 §4](https://github.com/sumsolution/folkways-command/blob/main/efforts/license/15-research-license.md)). Not legal advice.
- Contributions: inbound = outbound (Apache-2.0 s5). Decide DCO sign-off before the first outside PR.
