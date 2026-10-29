# Architecture decision records

Decisions about the folkways framework itself, one file each, `NNNN-<slug>.md`. Each follows the
template in [0001](0001-overlay-enforcement.md): a status table, then **Issue** (what needs deciding and
why now, ending in the question), **Context** (linked facts), **Decision** (the option and why),
**Options considered** (at least two real options; the Trade-off column is one honest sentence), and
**Consequences**. While **proposed**, the title is the question; once accepted it becomes the answer.
Status values: **proposed** → **accepted** or **rejected**; later **superseded by NNNN** or **deprecated**.
Team repos keep their own ADRs (the [`adr` skill](../../packages/content/skills/adr/SKILL.md)); this folder is for the framework.

| # | Title | Status |
|---|---|---|
| [0001](0001-overlay-enforcement.md) | Mores are enforced by mechanism; folkways are layered by overlays | accepted; M2 amended by 0014 |
| [0002](0002-hook-runtime.md) | Hooks are a Go binary, shipped as per-platform optional dependencies | accepted; partly superseded by 0011 |
| [0003](0003-repo-layout.md) | The repo is a pnpm workspace of independently versioned packages | accepted; partly superseded by 0013 |
| [0004](0004-license.md) | folkways is licensed under Apache-2.0 | accepted |
| [0005](0005-closing-efforts.md) | A closed effort is lifted, indexed and removed | accepted |
| [0006](0006-app-repo-files.md) | What files does an app repo carry? | proposed |
| [0007](0007-hook-delivery.md) | How do hooks reach agents in an app repo? | superseded by 0012 |
| [0010](0010-skill-architecture.md) | Partition folkways by session mode and effort ownership, not by SDLC agents | accepted |
| [0011](0011-cli-in-go.md) | The CLI is one Go binary, installed without Node first | accepted; partly superseded by 0012 |
| [0012](0012-cli-only.md) | folkways is only a CLI; every repo commits its skills and hooks | accepted |
| [0013](0013-release-please.md) | Conventional commits decide folkways' version; release-please opens the release PR and tags | accepted |
| [0014](0014-brain-pushes-without-asking.md) | Brain repo sessions push without asking; `folkways.yaml` sets when a push asks | accepted |
| [0015](0015-effort-homes.md) | Single-repo efforts and their reviews live in the app repo; cross-repo efforts and notes stay in the team repo | accepted |
