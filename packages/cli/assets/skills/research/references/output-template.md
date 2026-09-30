# Research output template

Read when writing the research file. Template for `NN-<topic>.md`.

## Template {#output}

```markdown
# <Topic>: research

**Question:** <what we set out to answer, one or two lines>
**Feeds decision:** <what the answer will be used for>
**Non-goals:** <what the answer and any plan from it deliberately won't cover>
**Researched:** <date>, by <handle> with <tool>.
**Answer:** <two or three lines; the conclusion first>
**Confidence:** high | medium | low. <one line on why>
**Enough to decide?** yes | partly | no. <what is missing>

## Findings

| ID | Finding | Confidence | Evidence |
|---|---|---|---|
| 1 | Tokens expire after 15 minutes | high | `src/auth/token.ts:42` at web-app@a1b2c3d |
| 2 | Refresh is handled by the gateway | medium | [EX-123](https://example.com/browse/EX-123); not confirmed in code |
| 3 | The gateway retries once (inferred) | low | Inferred from the timeout setting in finding 2 |

## Details
<Only what the table can't hold. Group by sub-question. Short quotes, each linked.>

## Plan
<Only when the research recommends building something. Steps, non-goals, and a rough size budget
(lines, files, days) so a reviewer can tell when the work outgrows it. Omit otherwise.>

## Risks and open questions
- <What could make this wrong; what the user should decide or check.>

## Sources
| Source | Searched | Result |
|---|---|---|
| knowledge | `knowledge/auth/` | Nothing on refresh |
| web-app | `git grep -n "expiresIn"` | Findings 1 and 2 |
| confluence | not searched: MCP server unavailable, user asked | Open question |

## SHAs read
| Repo | Ref | SHA | Date |
|---|---|---|---|
| web-app | origin/main | <full SHA> | <date fetched> |

## Escalations
<Where the strong model or a manual re-check was used, and what changed. Omit if none.>
```
