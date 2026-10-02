# Jira mechanics

Read when `folkways.yaml` `tracker.type` is `jira`. Items marked *(verify)* come from general knowledge,
not a cited source; check them against the team's instance the first time and tell the user if they
differ.

## Mapping

| Skill concept | Jira |
|---|---|
| Project | `tracker.project_key` (e.g. `EX`) |
| Bug, story, task, epic | issue types of the same name; epic sits above standard types, subtasks below them (children only) |
| Spike | usually a task with a `spike` label, unless the project has a Spike type; check the project's types |
| Parent | the `parent` field links a story, task or bug to its epic, and a subtask to its parent *(verify: older company-managed projects used an "Epic Link" custom field)* |
| Labels | the `labels` field; free text, so check existing values to avoid near-duplicates |
| Estimates, team fields | custom fields with instance-specific IDs (`customfield_<n>`); read the create metadata for the project and type rather than guessing IDs *(verify)* |

## Create

1. **An MCP server for Atlassian** if one is connected: look up the project's issue types and create
   metadata first, then create with type, summary, description, parent and fields in one call.
2. **A CLI** the user has set up (for example Atlassian's `acli` or the community `jira` CLI)
   *(verify the exact commands on the user's machine with `--help`)*.
3. **Neither:** give the user the title, type, parent, labels and description to paste.

Creating is a tracker write: only after an explicit yes.

## Description format

- Jira Cloud stores descriptions as Atlassian Document Format; older APIs and Data Center use wiki
  markup (`h2.`, `||header||`, `{code}`) *(verify which the tool expects)*. Most MCP servers accept
  Markdown and convert it; check the result.
- Wiki markup treats `{...}` as macros, so braces in inline code or JSON can vanish or break the
  render. Describe such code in prose, or put it in a code block.
- Tables with pipes inside cells and nested lists are the usual casualties of conversion.

## Verify after creating

Read the issue back and check: type, parent, labels and custom fields as intended; tables, code and
links rendered; no literal Markdown (`**`, `|---|`) showing. Fix what's wrong with an update (same
approval), then report the key and link.

## Quirks

- Required custom fields vary by project and type; a create that fails on a missing field names it.
  Ask the user for the value; don't pick one.
- Changing an issue's type can drop fields that the new type lacks. Ask before changing a type.
