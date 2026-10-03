# Linear mechanics

Read when `folkways.yaml` `tracker.type` is `linear`. Items marked *(verify)* come from general knowledge,
not a cited source; check them the first time and tell the user if they differ.

## Mapping

| Skill concept | Linear |
|---|---|
| Team | `tracker.project_key` is the team key in issue IDs (e.g. `EX-123`) |
| Bug, story, task, spike | Linear has one issue type; classify with labels (e.g. `Bug`, `Feature`); use the workspace's existing labels |
| Epic | a project, or a parent issue when the work is too large for one issue but too small for a project |
| Parent | parent and sub-issues; sub-issues inherit team, priority and project from the parent, but not labels, so set labels on each |
| Estimates | a per-team setting (points or t-shirt sizes) *(verify in team settings)*; only when `{#estimation}` asks |

Linear's own guidance favors plain issues with a concrete, scannable title over the "As a..." story
form. Keep the persona line only when the team uses it.

## Create

1. **An MCP server for Linear** if connected: look up the team, labels and parent, then create with
   title, description, labels and parent.
2. **No MCP server:** give the user the title, labels, parent and description to paste. (Linear has
   an API but no official CLI *(verify)*.)

Creating is a tracker write: only after an explicit yes.

## Verify after creating

Read the issue back: title, labels, parent, project, and that the Markdown description rendered
(tables and checklists). Report the ID and link.

## Quirks

- Descriptions are Markdown; checklists render as interactive checkboxes.
- An issue can be converted to a project later, so an epic that starts as a parent issue can grow.
