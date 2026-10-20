# Architecture decision records

Decisions about the folkways framework itself, one file each, `NNNN-<slug>.md`. Each follows the
template in [0001](0001-overlay-enforcement.md): a status table, then **Issue** (what needs deciding and
why now, ending in the question), **Context** (linked facts), **Decision** (the option and why),
**Options considered** (at least two real options; the Trade-off column is one honest sentence), and
**Consequences**. While **proposed**, the title is the question; once accepted it becomes the answer.
Status values: **proposed** → **accepted** or **rejected**; later **superseded by NNNN** or **deprecated**.
Team repos keep their own ADRs ([spec §6.1](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/06-v1-spec.md) `adr` skill); this folder is for the framework.

| # | Title | Status |
|---|---|---|
| [0001](0001-overlay-enforcement.md) | Mores are enforced by mechanism; folkways are layered by overlays | accepted |
| [0002](0002-hook-runtime.md) | Hooks are a Go binary, shipped as per-platform optional dependencies | accepted |
| [0003](0003-repo-layout.md) | The repo is a pnpm workspace of independently versioned packages | accepted |
| [0004](0004-license.md) | folkways is licensed under Apache-2.0 | accepted |
| [0005](0005-closing-efforts.md) | A closed effort is lifted, indexed and removed | accepted |
| [0006](0006-app-repo-files.md) | What files does an app repo carry? | proposed |
| [0007](0007-hook-delivery.md) | How do hooks reach agents in an app repo? | proposed |
| [0010](0010-skill-architecture.md) | Partition folkways by session mode and effort ownership, not by SDLC agents | accepted |
| [0011](0011-cli-in-go.md) | What is the CLI written in, and how is it installed? | proposed |
