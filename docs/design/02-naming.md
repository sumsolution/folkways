# Naming

**Last updated:** 2026-09-27
**Brief (user):** meaningful and thoughtful, a bit witty and fun, and above all it shows
intention. `crew*` dropped (too close to CrewAI). *Quorum* is the spark.

**Method:** npm registry lookup (`404` = free) plus `gh search repos --sort stars`, top 3
results. Checked 2026-09-27. Domains and trademarks not yet checked.

## Shortlist

| Name | Meaning | npm | GitHub conflicts |
|---|---|---|---|
| `kith` | Old English *cȳþþu*: both **"knowledge"** and **"one's people"** (as in *kith and kin*). Same root as *couth*. | taken | **CKang-J/Kith-space ★103: a local multi-agent workspace with AI teammates. Same space.** |
| `folkways` | Sociology (Sumner, 1906): a group's customary ways of doing things. The methods, encoded. | **free** | None in dev tooling |
| `collegium` | Roman body of people sharing a function; the root of *colleague*. | **free** | None in dev tooling |
| `quorate` | "The meeting is quorate": enough members present for decisions to count. | taken | terry-li-hm/quorate ★1, an LLM deliberation CLI |

## Rejected, and why

| Name | Idea | Conflict |
|---|---|---|
| `cairn` | Trail markers each passer adds to for the next | Two AI agent products (★3093, ★218) |
| `commonplace` | Commonplace books | **zby/commonplace: agent-operated knowledge framework, active 2026-09-26** |
| `stigmergy` | Coordination through traces left in a shared environment | Stigmergy-mcp: agent coordination MCP |
| `waggle` | The bee's dance that tells the hive where resources are | modiqo/waggle ★648: agent handoff references |
| `brigade` | Kitchen brigade: head chef orchestrates, the pass is the review gate | brigadecore ★2419 |
| `quire` | Bound pages; sounds like *choir* | quire.io (PM product), getty/quire |
| `dogwatch` | Short split watches built for crew rotation and handoffs | rapid7/dogwatch is a Datadog DSL, so confusing next to Datadog |
| `mise` | *Mise en place* | jdx/mise, a popular dev tool |
| `atelier`, `plenum`, `tessera`, `moot`, `scriptorium`, `tacit` | — | Crowded or wrong connotation |

Observation: agent-space tooling has already claimed most of the obvious collective-coordination
metaphors (cairn, stigmergy, waggle, commonplace). That confirms the concept, and it means the name
needs to come from further afield.

## Decision — 2026-09-27: **`folkways`** (working name, pending trademark check)

Chosen by the user. Availability, checked 2026-09-27:

| Where | Status |
|---|---|
| npm `folkways` | free (404) |
| PyPI `folkways` | free (404) |
| folkways.dev / .sh / .ai | no NS records, probably unregistered |
| folkways.io | **taken**, Folkways OÜ |
| GitHub `Folkways` org | **taken**, Folkways OÜ |

**Conflict:** [Folkways OÜ](https://folkways.io) sells realtime infrastructure and lists AI agents
among its use cases; it publishes `folkways-js` (last push 2025-11-02). Same broad field.
Smithsonian Folkways (a record label) is a different class, so low risk.

**Action (user):** check USPTO/EUIPO trademarks before going public. Use **folkways.dev** and a
GitHub org or handle other than `Folkways`.
