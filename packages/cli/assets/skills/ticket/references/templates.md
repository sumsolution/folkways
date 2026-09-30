# Ticket templates

One template per issue type. Bullets and `<angle>` text are prompts: replace them with verified
content or delete the heading. Headings are the ticket's own section headings; keep their order.
A team replaces one template with `## Override {#<type>-template}` in its overlay.

## Bug {#bug-template}

One problem per bug. Keep what you saw apart from what you think causes it.

```markdown
**Request:** <link to the support request or report; omit if none>

## Expected
When <action> in <environment>, <expected behavior>.

## Observed
<one or two lines; screenshot, log line or error text with its source>

## Steps to reproduce
1. <shortest path from an entry point: URL, command, API call>
2. <...>

## Environment
<env, version or SHA, browser or client, account type; frequency: always / intermittent (n of m)>

## Notes
<only lines that change what the implementer does. Known cause in one line with a link, else
nothing. No speculation, no investigation story.>
```

## Story {#story-template}

```markdown
**Request:** <link; omit if none>

As a <persona> I want <outcome> so that <reason>.
<Or one plain sentence of outcome and why, if the team doesn't use the persona form.>

## Users
<who needs it first; who to tell on release>

## Context
- <link>: <a clause of gloss at most>

## Acceptance criteria
| Given | When | Then |
|---|---|---|
| <specific entry point or state> | <real input, real field names> | <observable output> |

## Out of scope
- <what a reader might assume is included but isn't>

## Notes
<optional: Gotcha lines, one rejected option likely to be re-proposed, open questions>
```

## Task {#task-template}

```markdown
<Two or three sentences: what needs to happen and why now. May name the technical approach.>

## Outcomes
- [ ] <verifiable result: "`web-app` builds on Node 24 in CI", not "upgrade Node">

## Notes
<optional>
```

## Spike {#spike-template}

A question, a time-box and the decision it feeds. Use only when uncertainty blocks estimating or
designing; the output is a decision, not working code.

```markdown
I want to find out <question> because it decides <how we approach Y>.

**Time-box:** <days>. At the end, decide: continue, or proceed with what we know.

## Context
- <at most three links>

## Acceptance criteria
- [ ] <decision recorded: in the ticket, a doc, or an ADR>
- [ ] <follow-up tickets drafted>

## Findings
| Option | Description | Pros | Cons |
|---|---|---|---|
| <filled in during the spike> | | | |
```

## Epic {#epic-template}

Short. The child tickets hold the detail.

```markdown
## Goal
<one or two sentences: the outcome for users or the business>

## Scope
- In: <...>
- Out: <...>

## Success measure
<how we will know it worked: a metric with its source, or an observable state>
```

## Example: a filled story

The right altitude: outcomes and observable results, no files or functions, one gotcha.

```markdown
# Editors can schedule a banner to end at a set time

**Request:** https://support.example.com/requests/4411

As a campaign editor I want a banner to stop showing at a time I choose so that expired offers
don't stay live overnight.

## Users
Campaign editors in Example Team's content tool. Tell the support channel on release.

## Acceptance criteria
| Given | When | Then |
|---|---|---|
| A banner with `endsAt` set to 18:00 UTC | a visitor loads the home page at 18:01 UTC | the banner is not shown |
| The banner editor | an editor sets `endsAt` earlier than `startsAt` | the form shows "End must be after start" and doesn't save |
| A banner with no `endsAt` | any time passes | the banner shows until unpublished (today's behavior) |

## Out of scope
Recurring schedules.

## Notes
Gotcha: the home page is cached for 5 minutes at the CDN, so an ended banner can show for up to
5 minutes unless the cache is purged on expiry.
```
