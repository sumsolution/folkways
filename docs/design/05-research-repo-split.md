# Team / personal split: options

**Researched:** 2026-09-28. Tool facts: Sonnet subagent against primary docs; two decision-critical
citations re-read by the main agent (marked ✓). Design comparison: main agent.
**Decision (round 11): people/ + .local/ in the team repo; separate personal repo dropped.**
**Question:** does the private personal repo (round 6) earn its complexity, or should personal
files live with the team repo?

## What the personal layer holds today

`me.yaml` (projects dir, model prefs, notifications) · `work/` (scratch, personal efforts) ·
`work/inbox/` (side quests) · `overlays/` (personal skill additions).

## Tool facts that decide it

| # | Fact | Confidence | Source |
|---|---|---|---|
| 1 | **VS Code: `chat.additionalReadAccessFolders` is read-only** and user-scoped. Built-in edit tools stay inside the workspace. So a VS Code session in the team repo **cannot write** to a separate personal repo without terminal commands (approval) or a multi-root workspace. | verified (read-only), medium (no other write path) | [VS Code security](https://code.visualstudio.com/docs/agents/security) |
| 2 | **Claude Code: `permissions.additionalDirectories` in settings grants file access only.** Skills and other config load only from `--add-dir` / `/add-dir`. | verified ✓ | [Claude permissions](https://code.claude.com/docs/en/permissions#additional-directories-grant-file-access-not-configuration) |
| 3 | Claude Code: `~` expansion in `additionalDirectories` is not documented (a sibling setting, `plansDirectory`, has a bug where it isn't expanded). Use absolute paths. | low | [#44606](https://github.com/anthropics/claude-code/issues/44606) |
| 4 | Copilot CLI: `--add-dir` grants read and write. Persisting it via `/add-dir` → `trustedFolders` is documented, but an open issue says grants are lost after the session. | medium | [CLI config](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/configure-copilot-cli), [#2284](https://github.com/github/copilot-cli/issues/2284) |
| 5 | **VS Code: gitignored files are left out of text search, grep and the semantic index**, unless the file is open or has a selection. | verified ✓ | [Workspace context](https://code.visualstudio.com/docs/agents/reference/workspace-context) |
| 6 | VS Code: whether the read-file tool opens a gitignored in-workspace file by explicit path. | unconfirmed; test | — |
| 7 | Claude Code: gitignore is not a permission boundary. Read works on an explicit path; the `respectGitignore` setting covers the `@` picker. Glob/Grep behavior is inconsistent (issues). | high (read), medium (search) | [Settings](https://code.claude.com/docs/en/settings-reference), [permissions](https://code.claude.com/docs/en/permissions) |
| 8 | Cloud agent works in a clone of one repo. Gitignored files are never committed, so they don't exist there. Other repos need an MCP server + PAT. | high (git mechanics); medium (cross-repo) | [Cloud agent access](https://docs.github.com/en/copilot/concepts/agents/cloud-agent/access-management) |
| 9 | A nested git repo inside a gitignored dir is invisible to the outer repo. VS Code's Source Control may still detect it (scan-depth settings; open bugs). | high (git), medium (VS Code) | [gitignore](https://git-scm.com/docs/gitignore) |

**Scope update (Summer, 2026-09-28): v1 targets Claude Code (desktop + CLI) and Copilot CLI only.
IDE integrations (VS Code, JetBrains) and the Copilot cloud agent are out of scope.** Facts 1, 5, 6 and
9 (VS Code part) no longer matter. What's left:

- **Claude Code:** `permissions.additionalDirectories` in user settings gives read and write to an
  outside dir, desktop and CLI alike. Absolute path (fact 3). Skills there don't load, which is fine:
  overlays are read by path.
- **Copilot CLI:** `--add-dir` gives read and write. Persistence is unreliable (fact 4), so it needs a
  shell alias or wrapper.

So a separate repo is **workable** again: two setup steps per person. The choice now rests on
simplicity, promotion and privacy, not on a tool gap.

## Options (for Claude Code + Copilot CLI)

| | A. Separate private repo (current) | B. Committed `people/<handle>/` in team repo | C. Gitignored `.local/` in team repo | D. `~/.folkways/` outside repos |
|---|---|---|---|---|
| Privacy | Private | **Visible to team** | Private (never pushed) | Private |
| Setup per person | Create repo; `additionalDirectories` (Claude); `--add-dir` alias (Copilot CLI); path in config | None | None | Create dir; same two access steps |
| Reads and writes | Yes, after setup | Yes | Yes (explicit paths; search may skip it) | Yes, after setup |
| Backup / multi-machine | Yes | Yes | No, unless a nested private repo | No, unless made a repo |
| Promote an effort | Copy across repos; history lost | `git mv` to `efforts/` | `mv` to `efforts/` | Copy across |
| Sync loops | 2 repos | 1 | 1 | 1 (+ manual) |
| Headless / CI runs | Must be given the extra dir | Just works | Not present in CI (fine) | Must be given the extra dir |
| Repos a newcomer sees | 2 (+ framework, vendored) | 1 | 1 | 1 + a dir |

## Recommendation

**Drop the separate personal repo. Use B for personal files, C for machine config and the rare private item.**
Summer confirmed personal scratch is rarely sensitive, which removes A's main advantage.

```
team-repo/
├── people/<handle>/         # committed, pushed directly (like state/ and efforts/)
│   ├── overlays/            # personal skill additions (add-only)
│   ├── inbox/               # side quests
│   └── work/                # personal efforts; promote with git mv to efforts/
└── .local/                  # gitignored: me.yaml (machine paths, prefs), private scratch
```

Why:

- **No access setup at all.** Everything is inside the working dir, so both tools read and write it
  with normal permissions, and headless runs (review gate, audits) see it without extra flags.
- **One repo to sync.** One clone, one rebase habit.
- **Promotion is a `git mv`,** with history kept.
- **Personal work visible by default helps handoffs,** a v1 goal (round 2).
- Machine-specific values (projects dir, model prefs) belong to a machine, not a person. They fit `.local/`.

Costs:

- **Scratch in `people/` is visible to the team.** The rare sensitive item goes in `.local/`.
- **`.local/` has no backup** unless the person makes it a nested private repo (optional, documented).
- **More commits in the team repo.** `people/` changes are direct pushes, rebased, like `state/` and `efforts/`.
- **Search may skip `.local/`** (gitignored). Skills read it by explicit path, which works (fact 7).

## Other effects of the scope change (for the spec)

- **Hooks:** Copilot CLI reads `.claude/settings.json` with Claude matcher semantics, so matchers can
  be relied on. Scripts still self-filter on command text as a safety net. No `.github/hooks/`.
- **`SessionEnd`:** exists in Claude Code and Copilot CLI (`sessionEnd`), so the unclosed-session marker works on both.
- **Skills:** `.claude/skills/` in the team repo is read by both tools.
- **Model escalation:** Claude Code subagents natively; Copilot CLI via `copilot -p --model` (subagent `model:` unverified).
