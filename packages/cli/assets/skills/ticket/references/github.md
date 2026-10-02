# GitHub Issues mechanics

Read when `folkways.yaml` `tracker.type` is `github`. Items marked *(verify)* come from general knowledge,
not a cited source; check them the first time and tell the user if they differ.

## Mapping

| Skill concept | GitHub |
|---|---|
| Repository | where the team's issues live; ask if `folkways.yaml` doesn't say |
| Bug, story, task | organization issue types when the org defines them (defaults: task, bug, feature; a story maps to feature); otherwise a label per type |
| Spike, epic | an issue type if the org has one, else a label; an epic is a parent issue whose sub-issues are the work |
| Parent | sub-issues: up to 100 per parent, up to 8 levels deep |
| Labels | must already exist in the repo; never create one without asking |
| Estimates, custom fields | GitHub Projects fields, not issue fields; set them only if the team uses a project and the overlay says which |

## Repo templates win on shape

If the repo has issue forms or templates in `.github/ISSUE_TEMPLATE/`, keep their headings and
required fields and fill them with this skill's content and rules.

## Create

1. **An MCP server for GitHub** if connected: check the org's issue types, search for duplicates, then
   create with title, body, type and labels; add it as a sub-issue of the parent in a second call.
2. **The GitHub CLI:** write the body to a temporary file, then
   `gh issue create --repo <owner>/<repo> --title "<title>" --body-file <file> --label <label>`.
   Setting the issue type and adding a sub-issue may need `gh api` (REST or GraphQL) rather than a
   flag *(verify with `gh issue create --help` and the REST docs for sub-issues)*.
3. **Neither:** give the user the title, labels, parent and body to paste.

Creating is a tracker write: only after an explicit yes.

## Verify after creating

`gh issue view <number> --repo <owner>/<repo> --json title,body,labels,url` (or the MCP read), and
check the parent shows the sub-issue. Report the number and link.

## Quirks

- The body is GitHub-flavored Markdown; tables and task lists render as written.
- `#123` in a body links to issue 123 in the same repo; write `<owner>/<repo>#123` across repos.
- `Closes #123` in a PR description closes the issue on merge; in an issue body it does nothing.
