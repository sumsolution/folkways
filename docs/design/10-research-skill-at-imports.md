# `@` file references in skills: does Claude Code attach them?

**Researched:** 2026-09-29, by reading the JavaScript bundled in Claude Code **2.1.284** (the desktop
app's copy), plus the skills doc. No model calls. Names below are behaviours, not the minified
identifiers, which change between builds.
**Why:** step 0 needed a way to load a person's additions without a tool call per skill run
([ADR 0001](../adr/0001-overlay-enforcement.md), spec §3.2).

## Summary

| Question | Answer | Confidence |
|---|---|---|
| Does `@path` in a project `SKILL.md` attach the file? | Yes, both when the user types `/skill` and when the model calls the Skill tool | high (code) |
| When not? | Skills preloaded into a subagent (agent `skills:`), cron-scheduled runs, account-synced skills, coordinator mode | high (code) |
| Syntax | `@` at the start of a line or after whitespace; the path runs to the next whitespace, trimmed to a word boundary. `` `@path` `` in backticks is **not** matched. `@"path with spaces"` works | high (code) |
| Missing file | Skipped silently; no error, no placeholder | high (code) |
| Path variables | `${CLAUDE_PROJECT_DIR}` and `${CLAUDE_SKILL_DIR}` are replaced before mentions are extracted, so `@${CLAUDE_PROJECT_DIR}/x` is an absolute path | high (code) |
| Symlinks | Read through the normal filesystem, so followed (inferred; not run) | medium |
| After compaction | Not restored. Compaction re-attaches each invoked skill's own text (first 5,000 tokens, 25,000 total), not the files its mentions attached | high (code + [CC doc](https://code.claude.com/docs/en/skills#skill-content-lifecycle)) |
| Copilot CLI | Unknown; not tested. Assume the `@` line arrives as literal text | low |

The skills doc confirms the behaviour indirectly: describing synced skills outside a cloud session, it
contrasts them with local skills, which attach the files their `@` references name
([CC, synced skills](https://code.claude.com/docs/en/skills)).

## How it works

- The skill loader renders the body: it prepends the skill's base directory, substitutes arguments and
  `${CLAUDE_*}` variables, and runs `` !`command` `` lines. It does not touch `@`.
- The command processor then runs the same attachment step that handles `@` in a typed prompt over the
  **rendered skill text**, in "input mentions only" mode, and adds the results as file attachments
  after the skill message. The Skill tool calls the same processor. Forked skills (`context: fork`)
  run the same extraction over the body for the fork.
- Each mention is resolved and read like a prompt `@` mention: permission deny rules apply, and a
  directory attaches a listing. A read error drops that mention.

## What folkways does with it

- **Team overlays are compiled** into the skill at `sync`, because only the skill's own text survives
  compaction.
- **Personal additions are attached** by a line in step 0:
  `@${CLAUDE_PROJECT_DIR}/.local/me/overlays/<skill>/OVERLAY.md`, on its own line, not in backticks.
  `.local/me` is a symlink to `people/<handle>/`, so one path works for everyone.
- **A read fallback sits beside it:** "if that file wasn't attached, read
  `people/<handle>/overlays/<skill>/OVERLAY.md`". It covers compaction, preloaded skills and Copilot
  CLI. After compaction the restored skill text still names the path, so the agent can read it again.

## Open

- Run the `commit` skill with an `@` line in Copilot CLI and record what arrives.
- Behaviour is undocumented beyond the one doc sentence and could change; re-check on Claude Code
  upgrades (a skill with a marker file is enough).
