# ADR template

Read when drafting. `<angle>` text is a prompt: replace it or delete the line. Delete a section only
when it truly has nothing to say (Related Decisions often doesn't); never delete Options Considered
or Consequences.

## Template {#template}

```markdown
# <Decision as a statement, or the open question>

| Property | Value |
|---|---|
| Status | draft |
| Date | <YYYY-MM-DD of the last status change> |
| Owner | <name> |
| Reviewers | <plain-text names during review> |
| Group | <group or domain label> |
| Review window | <until YYYY-MM-DD> |
| Supersedes | <ADR number and link; omit if none> |

## Issue

<One or two paragraphs: what forces a decision now, and what goes wrong if we don't decide. End
with the question in one sentence.>

## Relevant context

- <concrete fact with a link: the service, endpoint, field, metric, constraint>
- <prior decisions, standards or deadlines that limit the options>

## Decision

Option <N>: <name>. <We will ... . Two to five sentences of rationale: why this option beats the
others on the concerns that matter here, including failed attempts or proofs of concept that shaped
it.>

Non-goals: <what this decision deliberately doesn't cover>. Size: <rough budget for the work: lines,
files, days>, so a reviewer can tell when the work outgrows it.

## Options considered

<Link deep analysis here instead of bloating the table.>

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | <name> | <...> | <...> | <one honest sentence: the net judgment> |
| 2 | <name> | <...> | <...> | <...> |

## Consequences

- <what becomes easier, and what becomes harder or costs more>
- <derived requirements and follow-on work, with ticket links>
- <what would make us revisit this decision>

## Related decisions

- <ADR number and link>: <relation: supersedes, depends on, constrains>

## Interested parties

- <teams or people to inform; not reviewers>
```
