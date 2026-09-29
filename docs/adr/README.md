# Architecture decision records

Decisions about the folkways framework itself, one file each, `NNNN-<slug>.md`. Each follows the
template in [0001](0001-overlay-enforcement.md): a status table, then **Issue** (what needs deciding and
why now, ending in the question), **Context** (linked facts), **Decision** (the option and why),
**Options considered** (at least two real options; the Trade-off column is one honest sentence), and
**Consequences**. While **proposed**, the title is the question; once accepted it becomes the answer.
Status values: **proposed** → **accepted** or **rejected**; later **superseded by NNNN** or **deprecated**.
Team repos keep their own ADRs (spec §6.1 `adr` skill); this folder is for the framework.

| # | Title | Status |
|---|---|---|
| [0001](0001-overlay-enforcement.md) | Which team rules must be enforced by mechanism, and how are the rest layered? | proposed |
| 0002 | Hook runtime (reserved; another session) | reserved |
