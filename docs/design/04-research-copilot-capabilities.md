# Claude Code vs GitHub Copilot: mechanism-by-mechanism capability check

**Researched:** 2026-09-27 (first pass: web search + doc fetch). **Re-checked:** 2026-09-27 against primary docs (see Opus re-check below).
**Confidence tags:** verified (read in primary doc) / high / medium / low.

## Summary

| # | Mechanism | Claude Code | Copilot (per surface) | Shared-file strategy |
|---|-----------|-------------|------------------------|----------------------|
| 1 | Instructions | `CLAUDE.md`; reads `AGENTS.md` natively **only if no `CLAUDE.md`/`CLAUDE.local.md`** exists (default, v2.1.277+) [verified] | VS Code, CLI, cloud agent read `AGENTS.md`; VS Code also reads `CLAUDE.md` [verified] | `AGENTS.md` = content; `CLAUDE.md` = `@AGENTS.md` + Claude-only extras |
| 2 | Skills | `.claude/skills/`, `~/.claude/skills/`, plugins, managed. Not `.agents/` or `.github/` [verified] | Project: `.github/skills/`, `.claude/skills/`, `.agents/skills/` on all surfaces. Personal: `~/.copilot/skills/`, `~/.agents/skills/` (+ `~/.claude/skills/` in VS Code). No enable setting in VS Code [verified] | `.claude/skills/<name>/SKILL.md` (project) is the one path both read |
| 3 | Hooks | `.claude/settings.json`; `PreToolUse` blocks (exit 2); `if: "Bash(git commit *)"` filter [verified] | **CLI reads `.claude/settings.json`** with Claude matcher semantics. **VS Code** reads it only with `chat.useClaudeHooks` (Preview, off by default) and **ignores matchers**. **Cloud agent** reads only `.github/hooks/*.json` [verified] | One `.claude/settings.json` for Claude + CLI + VS Code, if scripts self-filter. Second file (`.github/hooks/`) only if cloud agent must be gated |
| 4 | Headless CLI | `claude -p`, `--output-format`, `--max-turns`, `--max-budget-usd` [high] | `copilot -p`, `--model`, `--agent`, `--allow-all-tools`, `--no-ask-user`, `-s`, `--add-dir` [verified] | Wrapper script dispatches on tool; same prompt file |
| 5 | Subagent model | `.claude/agents/*.md` `model:` honored [verified] | VS Code: agent `model:` honored for subagents, **but capped at the main session's cost tier** [verified]. CLI subagent behavior unverified | Escalation *up* must start from a strong main session or be driven externally via `copilot -p --model` |
| 6 | Scheduled/CI | `anthropics/claude-code-action` on cron [high] | Cloud agent automations; or Actions cron + `copilot -p` [medium] | Shared workflow skeleton, tool-specific step |
| 7 | MCP | `.mcp.json` (`mcpServers`) [high] | Root `.mcp.json` read by VS Code + CLI; `.vscode/mcp.json` VS Code-only [medium] | Root `.mcp.json` |
| 8 | Layering / personal | `@import` (4 hops), user `~/.claude/`, `CLAUDE.local.md` [verified] | `applyTo:` globs; personal instructions exist in VS Code (profile, `~/.copilot/instructions`, `~/.claude/CLAUDE.md` depending on harness) and CLI [verified] | Skill-driven overlay (see §8) works on both; needs outside-workspace read access |

## Opus re-check — 2026-09-27

| Item | Result | Source |
|------|--------|--------|
| `chat.useClaudeHooks` exists, reads `.claude/settings.json` + `~/.claude/settings.json` | Confirmed. Off by default; VS Code hooks are **Preview**; **matchers ignored** (every command for the event runs); runtime payloads use VS Code's schema and tool names | [VS Code hooks](https://code.visualstudio.com/docs/copilot/customization/hooks) |
| VS Code Local hook events | Corrected: `SessionStart`, `UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `PreCompact`, `SubagentStart`, `SubagentStop`, `Stop`. **No `SessionEnd`** | [VS Code hooks](https://code.visualstudio.com/docs/copilot/customization/hooks) |
| Copilot CLI hook locations | Corrected: CLI **also reads `.claude/settings.json` / `.claude/settings.local.json`**; PascalCase `PreToolUse` gets Claude matcher semantics and `tool_name: "Bash"`; exit 2 or `permissionDecision: "deny"` blocks. Also `.github/hooks/*.json`, `~/.copilot/hooks/`, `.github/copilot/settings.json`, policy dir | [Hooks reference](https://docs.github.com/en/copilot/reference/hooks-reference) |
| Cloud agent hooks | Confirmed: only `.github/hooks/*.json` (`version: 1`, `bash`/`command` fields) | [Hooks reference](https://docs.github.com/en/copilot/reference/hooks-reference) |
| No native pre-commit/pre-push events | Confirmed for both tools | [Claude hooks](https://code.claude.com/docs/en/hooks), [Copilot hooks](https://docs.github.com/en/copilot/reference/hooks-reference) |
| Copilot subagent `model:` "not honored" (#310138) | Corrected. #310138 was a feature request, **closed in milestone 1.117**. Docs now: explicit `runSubagent` model > agent `model:` > main model. Selections above the main model's cost tier **don't run** | [#310138](https://github.com/microsoft/vscode/issues/310138), [VS Code subagents](https://code.visualstudio.com/docs/copilot/agents/subagents) |
| `copilot -p --model` | Confirmed | [CLI programmatic ref](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference) |
| Skills "experimental in VS Code" | Corrected: no enable setting; only "forked skill context" is experimental | [VS Code skills](https://code.visualstudio.com/docs/copilot/customization/agent-skills) |
| Copilot reads `.claude/skills/` | Confirmed: cloud agent, code review, CLI, VS Code, JetBrains | [GitHub skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills) |
| Personal `~/.claude/skills/` read by Copilot | Partly: VS Code lists it; GitHub docs list only `~/.copilot/skills`, `~/.agents/skills` (CLI likely doesn't) [medium] | [VS Code skills](https://code.visualstudio.com/docs/copilot/customization/agent-skills), [GitHub skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills) |
| Claude Code ignores `.agents/skills/`, `.github/skills/` | Confirmed | [Claude skills](https://code.claude.com/docs/en/skills), [memory](https://code.claude.com/docs/en/memory#agents-md) |
| Claude Code `AGENTS.md` | Corrected: default `claude-md-or-agents-md` — AGENTS.md only if no `CLAUDE.md`/`CLAUDE.local.md`. Setting "Project instructions" (`pluginConfigs."agents-md@builtin".options.instructionFiles`, user/managed settings only). Bedrock/telemetry exclusion ended in v2.1.281 | [Claude memory](https://code.claude.com/docs/en/memory#agents-md) |
| "VS Code has no personal instructions" | Corrected: it has profile instructions (Local), `~/.copilot/instructions` and `~/.claude/CLAUDE.md` (Agent Host) | [VS Code instructions](https://code.visualstudio.com/docs/copilot/customization/custom-instructions) |
| Reading files outside workspace | VS Code built-in tools: workspace only unless `chat.additionalReadAccessFolders`. Copilot CLI: cwd only unless `--add-dir` / `--allow-all-paths`. Claude Code: prompts outside working dir unless `--add-dir` / `additionalDirectories` | [VS Code security](https://code.visualstudio.com/docs/agents/security), [CLI ref](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference), [Claude permissions](https://code.claude.com/docs/en/permissions) |
| Agent Skills spec | Exists at agentskills.io: `name` + `description` required; `allowed-tools` experimental; no discovery dirs defined. Claude-only fields (`when_to_use`, `paths`, …) are extensions | [agentskills.io spec](https://agentskills.io/specification) |
| Sections 6, 7 (CI, MCP) | Still unverified in this pass | — |

## 1. Instructions file

- Claude Code default (`claude-md-or-agents-md`): reads `AGENTS.md` only when no `CLAUDE.md`, `.claude/CLAUDE.md`, or `CLAUDE.local.md` exists in cwd or above. **A personal `CLAUDE.local.md` silently turns AGENTS.md off.** Changeable via "Project instructions" in `/config` (user/managed settings only, ignored in project settings). [verified — [memory#agents-md](https://code.claude.com/docs/en/memory#agents-md)]
- The docs' own recommended share pattern is `CLAUDE.md` = `@AGENTS.md` + Claude extras; the import never double-loads. [verified]
- VS Code reads `AGENTS.md` (`chat.useAgentsMdFile`) and `CLAUDE.md` (`chat.useClaudeMdFile`), so Claude-only extras in `CLAUDE.md` are also visible to Copilot in VS Code. [verified — [VS Code instructions](https://code.visualstudio.com/docs/copilot/customization/custom-instructions)]
- **Conclusion:** `CLAUDE.md` containing `@AGENTS.md` is still right — it's the only setup robust to `CLAUDE.local.md` and older versions. Nested dirs: add a nested `CLAUDE.md` with `@AGENTS.md` where a nested `AGENTS.md` exists.

## 2. Skills

- Claude Code: `~/.claude/skills/`, `.claude/skills/`, nested, `--add-dir`, plugins, managed. Never `.agents/` or `.github/`. [verified — [skills](https://code.claude.com/docs/en/skills)]
- Copilot: project `.github/skills/`, `.claude/skills/`, `.agents/skills/` on cloud agent, code review, CLI, VS Code, JetBrains. Personal: `~/.copilot/skills/`, `~/.agents/skills/`; VS Code also `~/.claude/skills/`. Not experimental in VS Code. [verified — [GitHub](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills), [VS Code](https://code.visualstudio.com/docs/copilot/customization/agent-skills)]
- **Plan:** project skills at `.claude/skills/<name>/SKILL.md`. Personal skills: no single home path covers Claude + Copilot CLI (Claude needs `~/.claude/skills/`, CLI needs `~/.copilot/` or `~/.agents/`) — symlink or install to both. Keep frontmatter to spec fields (`name`, `description`) for portability.

## 3. Hooks

- **Claude Code:** `.claude/settings.json` etc.; `PreToolUse` blocks via exit 2; `matcher: "Bash"` + `if: "Bash(git commit *)"`. No git events. Does not read `.github/hooks/`. [verified — [hooks](https://code.claude.com/docs/en/hooks)]
- **Copilot CLI:** reads `.claude/settings.json` and `.claude/settings.local.json`. PascalCase `PreToolUse` uses Claude matcher semantics; payload `tool_name` is `Bash`; exit 2 or `permissionDecision: "deny"` blocks. Whether Claude's `if` field is honored: not documented [low]. [verified — [hooks reference](https://docs.github.com/en/copilot/reference/hooks-reference)]
- **VS Code (Local harness):** reads `.claude/settings.json` only with `chat.useClaudeHooks` (off by default, Preview). **Ignores matchers** — every command for the event runs; payloads use VS Code tool names. No `SessionEnd`. [verified — [VS Code hooks](https://code.visualstudio.com/docs/copilot/customization/hooks)]
- **Cloud agent:** `.github/hooks/*.json` only (`version: 1`, `bash`/`command`). [verified]
- **Conclusion:** write hooks once in `.claude/settings.json` for Claude Code, Copilot CLI, and VS Code — **provided each hook script filters itself** (read stdin, check tool name ∈ {`Bash`, VS Code terminal tool} and the command text for `git commit`/`git push`, else exit 0). Don't rely on `matcher`/`if` alone. VS Code users must turn on `chat.useClaudeHooks`. Add a thin `.github/hooks/*.json` pointing at the same scripts only if cloud-agent gating matters. Avoid `SessionEnd` for anything VS Code must do.
- **Commit/push gates:** in-agent `PreToolUse` as above, plus real git hooks (`commit-msg`, `pre-push`) as the tool-independent backstop.

## 4. Headless CLI

- `claude -p "<prompt>"`, `--output-format text|json|stream-json`, `--max-turns`, `--max-budget-usd`. [high]
- `copilot -p "<prompt>"` with `--model`, `--agent`, `--allow-all-tools`, `--no-ask-user`, `-s`, `--add-dir`, `--allow-all-paths`. [verified — [CLI programmatic ref](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference)]
- **Plan:** one wrapper script, one prompt file, dispatch on tool.

## 5. Subagents with model selection

- Claude Code: `.claude/agents/<name>.md` `model:` (alias, full ID, `inherit`); honored. [verified — [sub-agents](https://code.claude.com/docs/en/sub-agents)]
- VS Code: order is explicit `runSubagent` model → custom agent `model:` (string or fallback list) → main model. **Selections above the main model's cost tier don't run.** [verified — [subagents](https://code.visualstudio.com/docs/copilot/agents/subagents)]. #310138 (the cited "not honored" issue) was a feature request, closed in 1.117.
- Copilot CLI: whether agent `model:` is honored for subagents — not verified [low].
- **Conclusion:** inside VS Code Copilot, "cheap first, escalate" works only *downward* — run the main session on the strong model and delegate the cheap first pass to a cheap-model subagent; the strong main agent then reviews/escalates itself. A cheap main session cannot escalate up. For scripts/CI, drive escalation externally: `copilot -p --model <cheap>` then `copilot -p --model <strong>`.

## 6. Scheduled/CI runs [not re-checked]

- Claude: `anthropics/claude-code-action` supports cron-triggered workflows. [high]
- Copilot: cloud agent automations (schedule/event, draft PRs), or an Actions cron step calling `copilot -p`. [medium]
- **Plan:** shared workflow skeleton, tool-specific agent step.

## 7. MCP servers [not re-checked]

- Claude Code: root `.mcp.json` (`mcpServers`), user `~/.claude.json`. [high]
- VS Code/Copilot: `.vscode/mcp.json` (`servers`) or portable root `.mcp.json` (`mcpServers`); CLI also `~/.copilot/mcp-config.json`. [medium]
- **Plan:** root `.mcp.json`; `.vscode/mcp.json` only for VS Code-only extras.

## 8. Layering and personal overlays

- Claude Code: `@path` imports (4 hops); external imports need one-time approval; user-scope `~/.claude/CLAUDE.md`, `~/.claude/rules/`. [verified — [memory](https://code.claude.com/docs/en/memory)]
- Copilot: `.instructions.md` with `applyTo:` globs; personal instructions exist in VS Code (profile for Local; `~/.copilot/instructions`, `~/.claude/CLAUDE.md`/`rules` for Agent Host) and CLI. Precedence across files: undocumented [low]. [verified — [VS Code instructions](https://code.visualstudio.com/docs/copilot/customization/custom-instructions)]
- **Skill-driven overlay** (shared SKILL.md says "if `<personal repo>/overlays/<skill>.md` exists, read and apply it, additive only"; path from `~/.folkways/me.yaml` or env var): depends on no instruction layering — only on (a) the skill loading, (b) the agent reading two files outside the workspace (config + overlay). (a) holds on all surfaces. (b) needs setup per tool:
  - VS Code: built-in read tools are workspace-only → set `chat.additionalReadAccessFolders` to the personal repo and `~/.folkways` (or the agent falls back to a terminal `cat`, which needs approval). [verified]
  - Copilot CLI: `--add-dir <personal repo>` (and `~/.folkways`), or `--allow-all-paths`. [verified]
  - Claude Code: read prompt outside working dir unless `--add-dir` / `additionalDirectories` / `Read(...)` allow rule in user settings. [verified]
  - Cloud agent: no personal repo exists there — overlay silently skipped (acceptable; skill must treat "missing" as normal).
- It's instruction-following, not enforcement: the model may skip the read. Keep the overlay check as the first step of the skill body, with the exact path. [high]
- **Conclusion:** yes, it closes the gap on every local surface, with a one-time per-user access setup. Simpler variant: put `~/.folkways/` inside the personal repo path, or have the config live at a fixed path already granted.

## Open questions

- Copilot CLI: is the Claude `if` hook field honored? Is agent `model:` honored for subagents? Does it read `~/.claude/skills/`?
- VS Code terminal tool name in hook payloads (needed for self-filtering scripts) — capture via an audit hook.
- Copilot instruction precedence when several files apply.
- Sections 6–7 not re-checked against primary docs.

## Implications for folkways

- **Instructions:** `AGENTS.md` holds content; every dir with an `AGENTS.md` also gets a `CLAUDE.md` with `@AGENTS.md`. Don't rely on Claude's native AGENTS.md read — a user's `CLAUDE.local.md` disables it.
- **Skills:** `.claude/skills/<name>/SKILL.md`, spec-only frontmatter. Personal skills installed to both `~/.claude/skills/` and `~/.agents/skills/`.
- **Hooks:** one `.claude/settings.json` is enough for Claude Code, Copilot CLI and VS Code (with `chat.useClaudeHooks` on). Hook scripts must self-filter on tool + command because VS Code ignores matchers. Add `.github/hooks/*.json` reusing the same scripts only if cloud-agent gating is required. Real git hooks remain the backstop for commit/push gates.
- **Model escalation:** Claude Code subagents do it natively. VS Code Copilot only escalates *down* from the main model (cost-tier cap), so start the session on the strong model. Scripts/CI escalate via `copilot -p --model`.
- **Personal overlays:** the skill-driven overlay design works on both tools and doesn't need instruction layering. Folkways setup must grant read access to the personal repo per tool (`chat.additionalReadAccessFolders`, `copilot --add-dir`, Claude `additionalDirectories`), and skills must treat a missing overlay as normal.
- **MCP:** root `.mcp.json` (not re-checked this pass).
