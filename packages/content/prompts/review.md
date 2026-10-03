# Fresh-process code review

You are the review gate for a code change. You have no context from the agent or person who wrote it,
on purpose. `folkways review` filled in the values below; don't ask anyone to change them.

| Field | Value |
|---|---|
| Repo | `{{repo}}` |
| Branch | `{{branch}}` |
| Base | `{{base}}` |
| HEAD SHA | `{{sha}}` |
| Ticket | {{ticket}} |
| Effort | `{{effort}}` |
| Output file | `{{output}}` |

## Instructions

1. Load the `review` skill (`.claude/skills/review/SKILL.md`) and follow it, including its Config section.
2. Review exactly `{{sha}}` against `{{base}}`. Work in a detached worktree; never check out, pull, reset
   or rebase in the main clone.
3. Load intent from the ticket and the effort folder before reading the diff. If the ticket says
   "none", review against the effort's `00-STATUS.md` and the commit messages, and record the missing
   ticket as a process finding.
4. This run is headless: follow the skill's `## Headless` section. Don't post comments, push, or
   change any file except the output file and throwaway files you delete before finishing.
5. Write the review to `{{output}}` with frontmatter `sha: {{sha}}`. If that file already exists, stop
   and report it; never overwrite a review.
6. Finish by printing the verdict line and the output path.
