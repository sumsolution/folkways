# The name

## Sumner's folkways

In 1906 the Yale sociologist William Graham Sumner published *Folkways*. His subject was how groups
actually behave: not the rules they write down, but the customary ways of doing things that members pick
up, repeat and pass on. He called those **folkways**.

That's exactly what a team's way of working with agents is. How we review. How we write commits. When
we hand off. Every team has folkways already; they're just scattered across people's heads and private
prompts. This project writes them down so every teammate's agent shares them.

## Folkways and mores

Sumner drew a second line. Some customs a group cares about enough to enforce. He called those
**mores**. Ignore a folkway and people think you're odd. Ignore the mores and the group pushes back.

folkways the framework is built on that split:

| | Folkways | Mores |
|---|---|---|
| What | Customs the team shapes freely | Norms the team enforces |
| Examples | Skills, overlays, templates, `team.yaml` | Ask before push, fresh review, commit format, hands off your main clone |
| Backed by | Instructions and skills | Settings and hooks |
| Changed by | An overlay or a PR | A team decision |

Most of the framework is folkways, on purpose. Mores are few, and each one has a reason we'd defend in a
review. Today one of them is enforced (push asks
first); the rest arrive with the hooks.

We borrow Sumner's words, not his whole worldview: he was also a champion of laissez-faire and social
Darwinism, and this project is about the opposite instinct, a group looking after each other.

## How we chose it

The brief: meaningful and thoughtful, a bit witty and fun, and above all it shows intention.
*Quorum* was the spark. `hivemind` and `cortex` were loved and taken. `crew*` names were too close to
CrewAI.

It turns out agent tooling has already claimed most collective-coordination metaphors. That confirmed
the idea, and it meant the name had to come from further afield. Sociology was far enough.

### The shortlist

| Name | Meaning | Why not |
|---|---|---|
| `kith` | Old English for both "knowledge" and "one's people", as in *kith and kin* | Taken on npm; a similar agent-workspace project exists |
| **`folkways`** | A group's customary ways of doing things | **Chosen** |
| `collegium` | Roman body of people sharing a function; the root of *colleague* | Free on npm; not picked |
| `quorate` | Enough members present for decisions to count | Taken on npm, and an LLM tool already uses it |

### Names we loved and let go

| Name | The idea | What was already there |
|---|---|---|
| `cairn` | Trail markers each hiker adds to for the next | Two AI agent products |
| `stigmergy` | Coordination through traces left in a shared space, like ants | An agent coordination MCP server |
| `waggle` | The bee dance that tells the hive where the flowers are | An agent handoff project |
| `commonplace` | Commonplace books, where you collect what's worth keeping | An agent-operated knowledge framework |
| `brigade` | A kitchen brigade: head chef orchestrates, the pass is the review gate | brigadecore, a 2.4k-star project |
| `mise` | *Mise en place*: everything in its place before service | A popular dev tool |
| `dogwatch` | Short split watches, built for crew rotation and handoffs | An observability DSL, confusing next to it |
| `quire` | Bound pages; sounds like *choir* | A project management product |

The full research, with checks and dates, is in [02-naming.md](../design/02-naming.md). For now
`folkways` is a working name.
