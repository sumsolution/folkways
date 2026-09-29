# Skill authoring: frontmatter, triggering, portability, evals, and the ticket and ADR domains

**Researched:** 2026-09-29 (web search + doc fetch; three parallel research passes).
**Confidence tags:** high (stated in a primary doc or spec) / medium (secondary source, issue report, or fetch summary) / low (single weak source or unverified). "(inferred)" marks a conclusion the sources do not state. Fetch tools returned summaries, so wording was not always visible; quotes are kept under 15 words.

## Summary

| # | Topic | Headline finding |
|---|-------|------------------|
| 1 | Frontmatter | Spec needs only `name` + `description` (max 1,024). Claude Code adds many fields and ignores unknown ones; claude.ai upload rejects them. Copilot CLI unknown-field behaviour is undocumented |
| 2 | Descriptions | Say what + when, key use case first, no workflow summary. Listing budgets differ per tool: Claude Code 1% of context and 1,536 chars per entry, Codex 2% or 8,000 chars, Copilot undocumented |
| 3 | Disclosure | SKILL.md under 500 lines and about 5,000 tokens; references one level deep; TOC threshold disputed (100 vs 300 lines) |
| 4 | Freedom | Match specificity to fragility; explain why; give a default, not a menu |
| 5 | Portability | `.claude/skills/` is the one project path Claude Code and Copilot CLI both read. Substitutions, tool names, `allowed-tools` and `disable-model-invocation` are not portable |
| 6 | Scripts | Bundle when logic is reinvented each run; non-interactive, `--help`, structured output; expect permission prompts in both tools |
| 7 | Evals | Description optimizer: about 20 queries, near-miss negatives, 60/40 split, 3 runs. Output evals with vs without skill. Pressure-test against a baseline |
| 8 | Anti-patterns | Vague descriptions, deep references, menus, time-sensitive text, all-caps commands without reasons, scripts that punt errors |
| 9 | Tickets | Outcome-first title, context, scope, testable criteria; splitting patterns; soft readiness; tracker specifics are adapter config |
| 10 | ADRs | Small MADR-style record, immutable once accepted, supersede rather than edit, only for significant decisions, guard against anti-patterns |
| 11 | SE techniques | Iron laws, rationalization tables, red flags, verification gates, confidence filters, labelled review comments |

Source shorthand used in links: CC = Claude Code skills doc, BP = Claude platform best-practices doc, SPEC = agentskills.io specification, SC = anthropics skill-creator.

## 1. Frontmatter and limits

| # | Finding | Source |
|---|---------|--------|
| 1.1 | Open spec requires `name` and `description`. Optional: `license`, `compatibility` (1-500 chars), `metadata` (string map), `allowed-tools` (space-separated, experimental). [high] | [SPEC](https://agentskills.io/specification#frontmatter) |
| 1.2 | Spec `name`: 1-64 chars, lowercase alphanumerics and hyphens, no leading, trailing or double hyphen, must match the parent directory. [high] | [SPEC](https://agentskills.io/specification#name-field) |
| 1.3 | Spec `description`: 1-1024 chars, non-empty, states what the skill does and when to use it. [high] | [SPEC](https://agentskills.io/specification#description-field) |
| 1.4 | Claude platform docs add: no XML tags in name or description; name may not contain "anthropic" or "claude". Not in the open spec. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#skill-structure) |
| 1.5 | Claude Code makes every field optional: `name` defaults to the directory, `description` is recommended and falls back to the first body line. [high] | [CC](https://code.claude.com/docs/en/skills#frontmatter-reference) |
| 1.6 | Claude Code-only extensions include `when_to_use`, `argument-hint`, `arguments`, `disable-model-invocation`, `user-invocable`, `disallowed-tools`, `model`, `effort`, `context`, `agent`, `background`, `hooks`, `paths`, `shell`. [high] | [CC](https://code.claude.com/docs/en/skills#frontmatter-reference) |
| 1.7 | Claude Code silently ignores unknown fields and accepts `license`, `compatibility`, `metadata` without acting on them. [high] | [CC](https://code.claude.com/docs/en/skills#frontmatter-reference) |
| 1.8 | claude.ai upload, the Skills API and packaging hard-fail on any key outside the six spec fields. [high] | [CC](https://code.claude.com/docs/en/skills#using-skill-frontmatter-outside-claude-code) |
| 1.9 | Claude Code parses frontmatter only if `---` is line 1. Bad YAML loads the skill with no metadata, so it cannot match. `claude plugin validate` catches it. [high] | [CC](https://code.claude.com/docs/en/skills#skill-not-triggering) |
| 1.10 | Copilot CLI docs list `name`, `description` (required), `license`, `allowed-tools`; they set no size limits and name no other fields. [high] | [GitHub add-skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| 1.11 | VS Code Copilot adds `argument-hint`, `user-invocable`, `disable-model-invocation`, experimental `context: fork`; name max 64, description max 1024. [high] | [VS Code skills](https://code.visualstudio.com/docs/copilot/customization/agent-skills) |
| 1.12 | Copilot CLI handling of unknown fields is undocumented. Treat as unverified. [low] | [issue 3095](https://github.com/github/copilot-cli/issues/3095) |
| 1.13 | Copilot CLI bugs: `metadata` as the last field hid a skill (0.0.380); `disable-model-invocation: true` made a skill "not found" even when invoked by name (1.0.79). [medium] | [951](https://github.com/github/copilot-cli/issues/951), [4438](https://github.com/github/copilot-cli/issues/4438) |
| 1.14 | Codex needs `name` and `description`; UI and invocation policy live in a separate `agents/openai.yaml`. Codex ignores `disable-model-invocation`. [medium] | [Codex skills](https://learn.chatgpt.com/docs/build-skills), [issue 516](https://github.com/mattpocock/skills/issues/516) |
| 1.15 | Name gerund form (`processing-pdfs`) is a soft recommendation; noun and verb forms are fine. Stay consistent. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#naming-conventions) |

**Conflicts.**
- Description limit: spec and upload 1,024 vs Claude Code 1,536 (description plus `when_to_use`, truncated not rejected). A 1,100-char description works locally and fails upload. Portable ceiling is 1,024.
- `name`: required by spec and platform, optional in Claude Code. Include it and keep it equal to the directory.

## 2. Descriptions and triggering

| # | Finding | Source |
|---|---------|--------|
| 2.1 | The description is the primary trigger. It states what and when; all "when to use" text belongs there, not in the body. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#writing-effective-descriptions) |
| 2.2 | Voice: platform docs say third person because the text is injected into the system prompt; avoid "I can" and "you can". [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#writing-effective-descriptions) |
| 2.3 | Voice conflict: agentskills.io and skill-creator use imperative "Use when...". Both work; a third-person capability sentence plus an imperative "Use when" sentence satisfies both (inferred). [high] | [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md), [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| 2.4 | Include the words users actually say, and cover cases where the user does not name the domain. [high] | [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md) |
| 2.5 | Claude undertriggers, so skill-creator advises a "little bit pushy" description with explicit contexts. [high] | [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| 2.6 | Undertriggers: broaden the when-context. Overtriggers: state what the skill does not do and the boundary with neighbours. [high] | [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md) |
| 2.7 | Agents consult skills mainly for tasks they cannot do alone; trivial one-step requests may never trigger. Test with substantive queries. [high] | [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| 2.8 | Keep it to a few sentences and put the key use case first, because truncation cuts the tail. [high] | [CC](https://code.claude.com/docs/en/skills#frontmatter-reference) |
| 2.9 | The description must hold triggers, not the workflow. A workflow summary lets the agent follow it and skip the body. [high] | [superpowers writing-skills](https://raw.githubusercontent.com/obra/superpowers/main/skills/writing-skills/SKILL.md) |
| 2.10 | Vague descriptions ("Helps with documents") fail. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#writing-effective-descriptions) |

**Listing budgets** (what all skills' descriptions cost together):

| Tool | Budget | Overflow behaviour | Source |
|------|--------|--------------------|--------|
| Claude Code | 1% of context window; 1,536 chars per entry (description + `when_to_use`) [high] | All names stay; descriptions drop, least-recently-used first. Tunables: `skillListingBudgetFraction`, `skillListingMaxDescChars`, `SLASH_COMMAND_TOOL_CHAR_BUDGET` | [CC](https://code.claude.com/docs/en/skills#skill-descriptions-are-cut-short) |
| Codex | at most 2% of context or 8,000 chars [high] | Full instructions load only when selected | [Codex skills](https://learn.chatgpt.com/docs/build-skills) |
| Copilot CLI | Undocumented [low] | Docs say only that Copilot decides from prompt and description | [GitHub add-skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| Spec | About 100 tokens per skill at startup [high] | n/a | [SPEC](https://agentskills.io/specification#progressive-disclosure) |

- 2.11 `disable-model-invocation: true` removes the description from Claude Code's context (zero listing cost); `user-invocable: false` keeps it. [high] [CC](https://code.claude.com/docs/en/skills#control-who-invokes-a-skill)
- 2.12 Descriptions grow during optimization; recheck the 1,024 cap afterwards. [high] [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md)
- 2.13 superpowers advises under about 500 chars, third person, with symptoms and synonyms. [high] [writing-skills](https://raw.githubusercontent.com/obra/superpowers/main/skills/writing-skills/SKILL.md)
- 2.14 Metadata cost conflict: "~100 tokens" (spec, overview) vs "~100 words" (skill-creator). Immaterial for a cap in chars. [medium]

## 3. Progressive disclosure and length

| # | Finding | Source |
|---|---------|--------|
| 3.1 | Three levels: metadata always loaded, SKILL.md body on trigger, resources on demand (scripts run; only output enters context). [high] | [SPEC](https://agentskills.io/specification#progressive-disclosure), [overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview#how-skills-work) |
| 3.2 | SKILL.md under 500 lines (spec, BP, CC, skill-creator agree) and under about 5,000 tokens (spec, overview). Past that, add a layer of hierarchy. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#progressive-disclosure-patterns), [SPEC](https://agentskills.io/specification#progressive-disclosure) |
| 3.3 | Keep references one level deep from SKILL.md; nested chains lead to partial reads. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-deeply-nested-references) |
| 3.4 | TOC for long reference files: BP says over 100 lines; skill-creator says over 300. Conflict; a TOC above about 100 lines is safe (inferred). [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#structure-longer-reference-files-with-table-of-contents), [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| 3.5 | Say when to load each file ("read X if the API returns non-200"), not a generic "see references/". [high] | [AS best practices](https://agentskills.io/skill-creation/best-practices.md) |
| 3.6 | Gotchas belong in SKILL.md; the agent may not know to open a separate file for them. [high] | [AS best practices](https://agentskills.io/skill-creation/best-practices.md) |
| 3.7 | Split by domain or variant so only the relevant file loads; give files descriptive names. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices) |
| 3.8 | After auto-compaction Claude Code keeps only the first 5,000 tokens of each re-attached skill (25,000 shared). Put standing rules early. [high] | [CC](https://code.claude.com/docs/en/skills#skill-content-lifecycle) |
| 3.9 | Design coherent units: too narrow loads several skills per task; too broad triggers imprecisely. [high] | [AS best practices](https://agentskills.io/skill-creation/best-practices.md) |
| 3.10 | Context engineering: aim for the smallest set of high-signal tokens; load by lightweight path just in time. [high] | [Anthropic engineering](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) |

## 4. Degrees of freedom

- 4.1 High freedom (prose) when many approaches work; medium (parameterised script or pseudocode) with a preferred pattern; low (exact command, do not deviate) for fragile steps such as migrations. [high] [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#set-appropriate-degrees-of-freedom)
- 4.2 Explain why for flexible instructions. skill-creator disfavours all-caps ALWAYS/NEVER in favour of reasoning. [high] [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md)
- 4.3 Tension: BP's Claude A/B example lists stronger "MUST" wording as one fix for an ignored rule. Use it sparingly and only for real hard rules (inferred). [high] [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#develop-skills-iteratively-with-claude)
- 4.4 Give a default and one escape hatch, not a menu of tools. [high] [AS best practices](https://agentskills.io/skill-creation/best-practices.md)
- 4.5 Teach a procedure, not one instance's answer. Use templates for strict output, judgment for flexible output; use checklists and plan-validate-execute for batch or destructive work. [high] [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#common-patterns)
- 4.6 Test on every model you target: smaller models need enough guidance, larger ones may find it over-explained. [high] [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#test-with-all-models-you-plan-to-use)
- 4.7 Ask of each line: would the agent get this wrong without it? [high] [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#concise-is-key)

## 5. Portability between Claude Code and Copilot CLI

| # | Finding | Source |
|---|---------|--------|
| 5.1 | Copilot CLI reads project skills from `.github/skills`, `.claude/skills`, `.agents/skills`; personal from `~/.copilot/skills`, `~/.agents/skills`. [high] | [GitHub skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills) |
| 5.2 | Claude Code reads `.claude/skills`, `~/.claude/skills`, nested dirs, `--add-dir`, plugins. Not `.agents/skills` or `.github/skills`. [high] | [CC](https://code.claude.com/docs/en/skills#where-skills-live) |
| 5.3 | Whether Copilot CLI reads `~/.claude/skills` is unconfirmed (VS Code docs list it; GitHub docs do not). Do not rely on it. [medium] | [VS Code skills](https://code.visualstudio.com/docs/copilot/customization/agent-skills) |
| 5.4 | Codex scans `.agents/skills`, not `.claude/skills`. No single project path covers all three tools. [high] | [Codex skills](https://learn.chatgpt.com/docs/build-skills) |
| 5.5 | Claude Code substitutions (`$ARGUMENTS`, `$N`, `${CLAUDE_SKILL_DIR}`, `${CLAUDE_PROJECT_DIR}`, `!` injection) are extensions. No evidence Copilot CLI substitutes them; assume literal text. [high, other clients inferred] | [CC](https://code.claude.com/docs/en/skills#available-string-substitutions) |
| 5.6 | Copilot CLI may resolve a skill's relative paths against the cwd, notably with symlinks or after a directory change. Claude Code has a similar complaint. [medium] | [copilot-cli 1090](https://github.com/github/copilot-cli/issues/1090), [claude-code 56325](https://github.com/anthropics/claude-code/issues/56325) |
| 5.7 | Spec and BP say to reference files relative to the skill root, with forward slashes. [high] | [SPEC](https://agentskills.io/specification#file-references) |
| 5.8 | Explicit invocation: `/skill-name` in both Claude Code and Copilot CLI; `$name` in Codex CLI. Copilot CLI argument substitution is undocumented. [high; Codex medium] | [add-skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| 5.9 | Tool names differ. Claude Code: Read, Edit, Bash, Grep, Glob, Agent, AskUserQuestion. Copilot CLI: `view`, `edit`, `bash`, `grep`, `glob`, `task`, `ask_user`. [medium, Copilot names from search results] | [CLI command ref](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference) |
| 5.10 | `allowed-tools` syntax differs (`Bash(git diff *)` vs `shell(git:*)`), and Copilot warns against pre-approving shell for untrusted skills. [high] | [add-skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| 5.11 | No official guidance on tool-neutral wording; describing actions in prose is (inferred) from the tool-name divergence and the spec's lack of body rules. [low] | [SPEC](https://agentskills.io/specification#body-content) |
| 5.12 | Copilot CLI subagents exist: `*.agent.md` in `.github/agents/` or `~/.copilot/agents/`; `/agent`, `--agent`, `/fleet`. Agent files list no `model` field. [high; fleet medium] | [custom agents](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/create-custom-agents-for-cli) |
| 5.13 | Skills cannot declare `model` in Copilot CLI; subagent model control is unsettled. Claude Code skills can fork (`context: fork`, `agent:`) but that is Claude-only. [medium] | [issue 3095](https://github.com/github/copilot-cli/issues/3095), [issue 2758](https://github.com/github/copilot-cli/issues/2758) |
| 5.14 | Headless: `copilot -p` runs and exits; `--no-ask-user` disables the ask tool. Docs say nothing about skills in programmatic mode. [high] | [programmatic ref](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference) |
| 5.15 | Headless Claude Code: `/skill-name` works inside `-p`; `--permission-prompts none` removes `AskUserQuestion`; `--bare` skips skill auto-discovery. [high] | [headless](https://code.claude.com/docs/en/headless) |
| 5.16 | Copilot CLI reads AGENTS.md, CLAUDE.md, GEMINI.md and Copilot instruction files, combines them with no precedence. A `@AGENTS.md` import in CLAUDE.md shows as literal text there (inferred). [high; inference medium] | [custom instructions](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions) |

**Conflict with [04](04-research-copilot-capabilities.md):** doc 04 records that VS Code honours agent `model:` for subagents (capped at the main model's tier) and that Copilot CLI is unverified. This pass found no `model` field in CLI agent docs; the two agree that CLI is unverified.

## 6. Scripts in skills

| # | Finding | Source |
|---|---------|--------|
| 6.1 | Bundle a script when the agent reinvents the same logic across runs, a command grows too complex to get right first try, or the operation must be deterministic. [high] | [AS scripts](https://agentskills.io/skill-creation/using-scripts.md) |
| 6.2 | State whether the agent should run the script or read it. Only output enters context. [high] | [overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview#the-skills-architecture) |
| 6.3 | Design for agents: no interactive prompts (hard requirement), `--help`, actionable errors, JSON on stdout and diagnostics on stderr, idempotent, dry-run, exit codes, bounded output (harnesses truncate around 10-30K chars). [high] | [AS scripts](https://agentskills.io/skill-creation/using-scripts.md#designing-scripts-for-agentic-use) |
| 6.4 | Handle errors inside the script instead of leaving the agent to improvise; document every constant. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#solve-dont-defer) |
| 6.5 | Do not assume packages exist; list requirements in the body or `compatibility`. Runtime differs by surface (no network in the API, full network in Claude Code). [high] | [overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview#runtime-environment-constraints) |
| 6.6 | Both tools run scripts through the shell tool and prompt unless allowed. Do not assume pre-approval. [high] | [add-skills](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) |
| 6.7 | Codex advises instructions over scripts unless determinism is needed. [high] | [Codex skills](https://learn.chatgpt.com/docs/build-skills) |
| 6.8 | Treat skills with scripts or external fetches like installed software; audit before use. [high] | [overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview#security-considerations) |
| 6.9 | Self-contained deps: PEP 723 with `uv run`, Deno `npm:` imports, pinned `npx pkg@ver`. [high] | [AS scripts](https://agentskills.io/skill-creation/using-scripts.md#self-contained-scripts) |

## 7. Evaluating and iterating

| # | Finding | Source |
|---|---------|--------|
| 7.1 | Build evals first: run the task without the skill, find gaps, write minimal instructions, iterate. At least three scenarios. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#build-evaluations-first) |
| 7.2 | Description optimizer (skill-creator): about 20 queries as `{query, should_trigger}`, 8-10 each side. Negatives must be near-misses sharing keywords, not obviously unrelated. [high] | [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| 7.3 | Its loop: 60% train / 40% held-out test, each query run 3 times for a trigger rate, propose a new description from failures, up to 5 iterations, pick the best by test score. Needs `claude -p`. [high] | [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| 7.4 | Generic version: pass if trigger rate is at least 0.5 for positives and under 0.5 for negatives; fixed shuffled split; edit only from train failures; do not paste failed-query keywords in; then check 5-10 fresh queries. [high] | [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md#the-optimization-loop) |
| 7.5 | Queries should be realistic: paths, names, typos, backstory. [high] | [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md) |
| 7.6 | Output evals: `evals/evals.json` with prompt, expected output, assertions; run each with and without the skill (or old version) in clean contexts; grade, aggregate pass rate, time, tokens; human review; iterate. [high] | [AS evaluating](https://agentskills.io/skill-creation/evaluating-skills.md) |
| 7.7 | Grading: require concrete evidence for a pass; drop assertions that always pass in both configs; script mechanical checks. [high] | [AS evaluating](https://agentskills.io/skill-creation/evaluating-skills.md#grading-principles) |
| 7.8 | Use a fresh session with and without the skill; the authoring context masks gaps. `claude plugin eval` exists for plugin skills. [high] | [CC](https://code.claude.com/docs/en/skills#evaluate-and-iterate-on-a-skill) |
| 7.9 | Claude A / Claude B: A writes and refines the skill; a fresh B does real tasks; watch B's read order, missed references and unread files; feed specifics back to A. [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#develop-skills-iteratively-with-claude) |
| 7.10 | Pressure-testing (superpowers): run baseline scenarios without the skill, capture verbatim rationalizations, write the minimum, re-test, close new loopholes. [high] | [writing-skills](https://raw.githubusercontent.com/obra/superpowers/main/skills/writing-skills/SKILL.md) |
| 7.11 | Start from real expertise (runbooks, incident reports, review comments), read traces not only outputs, add every correction to a Gotchas section. [high] | [AS best practices](https://agentskills.io/skill-creation/best-practices.md#start-from-real-expertise) |

Conflict: BP says there is no built-in eval runner; the Claude Code doc now documents skill-creator and `claude plugin eval`. Prefer the newer doc. [high]

## 8. Anti-patterns

| Anti-pattern | Source |
|--------------|--------|
| Vague or first/second-person descriptions; vague names (`helper`, `utils`); reserved words in names [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#naming-conventions) |
| Workflow summary inside the description [high] | [writing-skills](https://raw.githubusercontent.com/obra/superpowers/main/skills/writing-skills/SKILL.md) |
| References nested more than one level [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-deeply-nested-references) |
| Time-sensitive text; use a "current method" plus a collapsed "old patterns" section [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-time-sensitive-information) |
| Inconsistent terms for one concept [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#use-consistent-terminology) |
| Menus of options instead of a default [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-offering-too-many-options) |
| Backslash paths; opaque file names (`doc2.md`) [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-windows-style-paths) |
| Explaining what the model already knows; over-comprehensive edge-case lists [high] | [AS best practices](https://agentskills.io/skill-creation/best-practices.md#aim-for-moderate-detail) |
| Generic, LLM-written skills with no domain content [high] | [AS best practices](https://agentskills.io/skill-creation/best-practices.md#start-from-real-expertise) |
| All-caps MUST/NEVER with no reasoning; over-narrow example-specific skills [high] | [SC](https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md) |
| Scripts that prompt, dump huge output, or punt errors; magic constants [high] | [AS scripts](https://agentskills.io/skill-creation/using-scripts.md) |
| MCP tool names without the server prefix [high] | [BP](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#mcp-tool-references) |
| Overfitting the description to test queries; trivial negative queries [high] | [AS optimizing](https://agentskills.io/skill-creation/optimizing-descriptions.md) |
| Overlapping skill scopes: if a human cannot pick the tool, the agent cannot either [high] | [Anthropic engineering](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) |
| Relying on `disable-model-invocation` in Copilot CLI or Codex [medium] | [issue 4438](https://github.com/github/copilot-cli/issues/4438) |

## 9. Ticket-writing best practice

| # | Finding | Source |
|---|---------|--------|
| 9.1 | INVEST: independent, negotiable, valuable, estimable, small, testable. Testable means you understand it well enough to write a test. Split vertically. [high] | [xp123](https://xp123.com/invest-in-good-stories-and-smart-tasks/) |
| 9.2 | Cohn: the "As a... I want... so that" template should serve the team, not constrain it. Card, conversation, confirmation; the card is a trigger for shared understanding. [high] | [Mountain Goat](https://www.mountaingoatsoftware.com/agile/user-stories) |
| 9.3 | Cohn's mistakes: story as complete requirement, oversized stories, too much early detail, forcing every item into the template. [high] | [Mountain Goat](https://www.mountaingoatsoftware.com/agile/user-stories) |
| 9.4 | Linear critique: write plain issues with a concrete task and scannable title, optional context and links, written by the person doing the work. [high] | [Linear](https://linear.app/method/write-issues-not-user-stories) |
| 9.5 | Atlassian: plain language, fits one sprint, has acceptance criteria, user need over implementation. Fetch summary was thin. [medium] | [Atlassian](https://www.atlassian.com/agile/project-management/user-stories) |
| 9.6 | Gherkin: Given is known state, When is the action, Then is observable output. Scenarios are executable specs. [high] | [Cucumber](https://cucumber.io/docs/gherkin/reference/) |
| 9.7 | Gherkin vs checklist: Given/When/Then for stateful behaviour and automatable scenarios, a checklist for simple or non-functional criteria. No primary source; a synthesis (inferred). [low] | (none; the Atlassian criteria page failed to fetch) |
| 9.8 | Lawrence splitting: confirm INVEST, then apply one of nine patterns (workflow steps, CRUD, business rules, data variations, data entry, major effort, simple/complex, defer performance, spike last). Prefer splits that let you drop a piece. [high] | [Humanizing Work](https://www.humanizingwork.com/the-humanizing-work-guide-to-splitting-user-stories/) |
| 9.9 | Bug reports: specific summary of about 10 words, reproducible steps as the key part, expected vs actual apart from speculation, frequency and environment, one issue per report. [high] | [Mozilla](https://bugzilla.mozilla.org/page.cgi?id=bug-writing.html) |
| 9.10 | Spikes: time-boxed, for excessive uncertainty only, end with a decision. Overuse delays value. [high] | [Cohn](https://www.mountaingoatsoftware.com/blog/spikes) |
| 9.11 | Scrum has no formal definition-of-ready checklist; refined items the team can finish in one sprint are "ready for selection". [high] | [Scrum Guide](https://scrumguides.org/scrum-guide.html) |
| 9.12 | DoR critique: a rigid DoR becomes a stage gate that blocks overlapping work; prefer flexible guidelines. [high] | [Cohn](https://www.mountaingoatsoftware.com/agile/the-dangers-of-a-definition-of-ready) |
| 9.13 | Definition of Done describes the Increment's quality bar for the team's work, not per ticket. [high] | [Scrum Guide](https://scrumguides.org/scrum-guide.html) |
| 9.14 | Estimation: Developers size items; Linear offers points or t-shirt sizes. No primary source read on no-estimates. [low] | [Linear](https://linear.app/docs/creating-issues) |
| 9.15 | GitHub: org issue types (defaults task, bug, feature; max 25), sub-issues (100 per parent, 8 levels), issue forms in YAML with required fields. [high] | [sub-issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/adding-sub-issues), [types](https://docs.github.com/en/issues/tracking-your-work-with-issues/configuring-issues/managing-issue-types-in-an-organization), [forms](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms) |
| 9.16 | Jira: three levels: epic, standard types (story, task, bug), subtask; subtasks are children only. [high] | [Atlassian admin](https://support.atlassian.com/jira-cloud-administration/docs/what-are-issue-types/) |
| 9.17 | Linear: one issue type, classification by labels; sub-issues inherit team, priority and project but not labels; suited to work too big for an issue and too small for a project. [high] | [parent/sub-issues](https://linear.app/docs/parent-and-sub-issues) |

## 10. ADR best practice

| # | Finding | Source |
|---|---------|--------|
| 10.1 | Nygard: title (short noun phrase), context, decision ("We will..."), consequences (good, bad, neutral); status proposed, accepted, deprecated, superseded; sequential numbers, never reused; keep in the repo; one to two pages. [high] | [Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) |
| 10.2 | MADR: context and problem, drivers (optional), options, outcome with justification, consequences, confirmation (optional), pros and cons. Front matter has status, date, deciders. Files `NNNN-title-with-dashes.md` in `docs/decisions/`. [medium] | [MADR](https://adr.github.io/madr/) |
| 10.3 | MADR titles represent the solved problem and the chosen solution. [medium] | [MADR](https://adr.github.io/madr/) |
| 10.4 | Title style differs: Nygard wants a noun phrase, Parker Henderson suggests an imperative phrase (`choose-database`), MADR problem plus solution. Conflict; pick one convention per repo. [high] | [Parker Henderson](https://github.com/joelparkerhenderson/architecture-decision-record) |
| 10.5 | When warranted: architecturally significant decisions; skip those small in scope, time, risk and cost, or already documented. One decision per ADR. [high] | [Parker Henderson](https://github.com/joelparkerhenderson/architecture-decision-record) |
| 10.6 | AWS scope: structure, non-functional requirements, dependencies, interfaces, construction techniques. Minimum sections: context, decision, consequences. [high] | [AWS](https://docs.aws.amazon.com/prescriptive-guidance/latest/architectural-decision-records/adr-process.html) |
| 10.7 | Google Cloud: write one when guidance is missing, several options need a comparison, or the solution is undocumented; keep the prior decision and say why it changed. [high] | [Google Cloud](https://docs.cloud.google.com/architecture/architecture-decision-records) |
| 10.8 | Lifecycle (AWS): owner posts as Proposed; silent reading of 10-15 minutes, then discussion; outcome stays Proposed, Rejected (reason recorded) or Accepted (timestamp, stakeholders). [high] | [AWS](https://docs.aws.amazon.com/prescriptive-guidance/latest/architectural-decision-records/adr-process.html) |
| 10.9 | Immutability: accepted or rejected ADRs are not edited; a change is a new ADR and the old one becomes Superseded. Reviewers can cite ADRs in code review. [high] | [AWS](https://docs.aws.amazon.com/prescriptive-guidance/latest/architectural-decision-records/adr-process.html) |
| 10.10 | Thoughtworks Radar moved lightweight ADRs to Adopt and says keep them with the code, not in a wiki. [high] | [Radar](https://www.thoughtworks.com/radar/techniques/lightweight-architecture-decision-records) |
| 10.11 | Y-statement: one sentence with context, concern, decision, rejected alternatives, benefits, drawbacks. [high] | [Zimmermann](https://ozimmer.ch/practices/2020/04/27/ArchitectureDecisionMaking.html) |
| 10.12 | Zimmermann anti-patterns: Fairy Tale, Sales Pitch, Free Lunch Coupon, Dummy Alternative, Sprint, Tunnel Vision, Maze, Blueprint in Disguise, Mega-ADR, Novel, Magic Tricks. [high] | [Zimmermann](https://www.ozimmer.ch/practices/2023/04/03/ADRCreation.html) |
| 10.13 | ecADR done-check: Evidence, Criteria (two or more alternatives), Agreement, Documentation, Realization/Review. [high] | [Zimmermann](https://ozimmer.ch/practices/2020/05/22/ADDefinitionOfDone.html) |
| 10.14 | Nygard's own "when to write" wording and other numbering conventions were not confirmed from the fetched pages. [low] | [Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) |

## 11. Techniques from strong public SE skills

| # | Finding | Source |
|---|---------|--------|
| 11.1 | superpowers TDD: an iron law (failing test before production code), an excuse-vs-reality table, a red-flags list that says start over, an end checklist. [high] | [TDD](https://raw.githubusercontent.com/obra/superpowers/main/skills/test-driven-development/SKILL.md) |
| 11.2 | Systematic debugging: no fixes without root cause; one variable at a time; after three or more failed fixes, question the design. [high] | [debugging](https://raw.githubusercontent.com/obra/superpowers/main/skills/systematic-debugging/SKILL.md) |
| 11.3 | Verification gate: identify the command, run it fresh, read all output and the exit code, then claim. Hedge words ("should", "probably") are red flags. [high] | [verification](https://raw.githubusercontent.com/obra/superpowers/main/skills/verification-before-completion/SKILL.md) |
| 11.4 | Review by a subagent given crafted context (commit range), not session history; severity tiers critical, important, minor. [high] | [requesting-code-review](https://raw.githubusercontent.com/obra/superpowers/main/skills/requesting-code-review/SKILL.md) |
| 11.5 | claude-code code-review plugin: parallel agents, skips draft or trivial PRs, scores issues 0-100 and drops those under 80, filters pre-existing and linter-catchable items. [high] | [code-review](https://github.com/anthropics/claude-code/tree/main/plugins/code-review) |
| 11.6 | Conventional Comments: `<label> [decorations]: <subject>`; labels praise, nitpick, suggestion, issue, todo, question, thought, chore; decorations such as blocking. [high] | [conventionalcomments.org](https://conventionalcomments.org/) |
| 11.7 | Google review standard: approve when the change clearly improves code health; prefix optional points "Nit:"; facts beat preferences. [high] | [standard](https://google.github.io/eng-practices/review/reviewer/standard.html) |
| 11.8 | Google reviewers check design, functionality, complexity, tests, naming, comments, style, docs, and give positive feedback. [high] | [looking-for](https://google.github.io/eng-practices/review/reviewer/looking-for.html) |
| 11.9 | Google small changes: one concern each; refactors separate; tests with the code. Around 100 lines is reasonable, 1,000 too large. [high] | [small-cls](https://google.github.io/eng-practices/review/developer/small-cls.html) |
| 11.10 | Google descriptions: imperative first line that stands alone; body explains context and why; avoid "Fix bug". [high] | [cl-descriptions](https://google.github.io/eng-practices/review/developer/cl-descriptions.html) |
| 11.11 | Conventional Commits 1.0: `type(scope): description`; feat is minor, fix is patch, breaking change (`!` or footer) is major. [high] | [spec](https://www.conventionalcommits.org/en/v1.0.0/) |
| 11.12 | Angular types: build, ci, docs, feat, fix, perf, refactor, test. Imperative lowercase header, no period; body explains why; footers for breaking changes and issue links. [high] | [Angular](https://raw.githubusercontent.com/angular/angular/main/contributing-docs/commit-message-guidelines.md) |
| 11.13 | Recurring pattern (inferred from 11.1-11.5): a one-line law, an excuse table, a red-flag stop signal, a fresh-evidence gate, a threshold (three failures, confidence 80), and isolated reviewer context. [medium] | (synthesis) |

## Open questions

- Copilot CLI: does it warn on or reject unknown frontmatter fields? Does it enforce a listing budget or per-entry cap? Does it read `~/.claude/skills`?
- Copilot CLI: exact tool names (only secondary sources) and whether a subagent honours a `model` field.
- Copilot CLI and Codex: do slash or `$` invocations of skills work in `-p` and `codex exec` modes? Codex exec was not verified.
- Whether `${CLAUDE_SKILL_DIR}` is left literal or stripped by Copilot CLI (assumed literal).
- Description-length effect on trigger rate at 500 chars vs 1,024: no source measures it; needs our own trigger eval (see section 12).
- Primary sources for scenario-vs-checklist criteria, estimation approaches, and Nygard's "when to write".
- Two engineering blog posts (Anthropic) were read through a summariser; anchors not captured.

## Implications for folkways

- **Frontmatter:** `name` and `description` only, `name` equal to the directory. No `license`, `metadata`, `allowed-tools`, `disable-model-invocation`, `when_to_use` or any Claude-only field (1.6-1.13).
- **Description cap:** 500 chars for framework skills, below the spec hard limit of 1,024. Reason: about 17 framework skills plus team skills share one listing budget of about 1% of context in Claude Code, and Codex's is smaller still (inferred from 2.x). Write a third-person capability sentence, then "Use when..." triggers, key use case first, with no workflow steps.
- **Length:** SKILL.md under 500 lines. Put standing rules and gotchas near the top so compaction keeps them (3.2, 3.6, 3.8).
- **References:** one level deep, linked from SKILL.md with a "read this when" note; add a TOC to any reference over about 100 lines (3.3-3.5).
- **Tool-neutral wording:** verbs ("read the file", "search the repo", "run the tests", "delegate if a subagent is available, otherwise do it inline"), not tool names (5.9-5.11).
- **No substitutions:** no `${CLAUDE_SKILL_DIR}`, `$ARGUMENTS` or `!` injection (5.5).
- **Paths:** write links relative to the skill folder, and when an instruction tells the agent to read a file, also give the full path `.claude/skills/<name>/references/...`, because Copilot may resolve against the cwd (5.6, inferred).
- **Location:** ship to `.claude/skills/<name>/`. Consider an `.agents/skills/` copy only if Codex support is wanted (5.1-5.4).
- **Headless:** any skill that may run under `claude -p` or `copilot -p --no-ask-user` must not rely on questions. For each question, define a fallback: state the assumption and continue, or stop and report what is missing (5.14, 5.15).
- **Scripts:** only when logic repeats or must be deterministic; bash or node, non-interactive, `--help`, JSON output, requirements stated in the body, permission prompts expected (6.1-6.6).
- **Degrees of freedom:** low freedom for commit format, ADR numbering and status changes; high freedom for prose sections. Give one default and reasons, not all-caps commands (4.1-4.4).
- **Ticket skill:** outcome-first title, context, scope, testable criteria; story format optional; run an INVEST and splitting check; bug essentials as a fixed template; spikes as question, time-box, decision; soft readiness only; type, parent and estimate as team-config adapter settings for GitHub, Jira and Linear (9.x).
- **ADR skill:** small MADR-style record with status and date; next number computed from existing files and never reused; write only for significant decisions; two or more real alternatives and real downsides; immutable after acceptance, supersede with cross-links; one convention for titles (10.x).
- **Review and commit skills:** labelled comments with a blocking flag, a confidence filter, Conventional Commits with Angular types matching the repo's scopes (11.x).
- **Discipline skills:** where an agent tends to cut corners (verify, debug, TDD), use an iron law, an excuse table and a fresh-evidence gate (11.1-11.3).
- **Process:** write each skill after a baseline run without it; keep 20-query trigger sets (8-10 positive, 8-10 near-miss negative) per skill; check that descriptions stay under the cap after tuning (7.x, 2.12).

## 12. Trigger eval

Results pending; filled in after the skills are written.
