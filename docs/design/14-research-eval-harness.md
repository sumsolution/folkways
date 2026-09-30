# Eval harness: build or adopt

**Researched:** 2026-09-30 (operator session; docs read, `claude plugin eval` and `claude -p` probed in
Claude Code 2.1.285). Feeds the harness in [packages/evals/](../../packages/evals).
**Confidence:** high / medium / low per finding. *Inferred* marks an inference.

## Question

We'll run many behavior evals: every skill change, the retro in [13](13-retro-git-guard-review-rounds.md),
overlay precedence ([07](07-research-overlay-enforcement.md)), later hooks and Copilot CLI. Do we adopt
an eval tool or build a small one? Security comes first with anything third-party.

## What an eval here needs

1. **A real team repo.** The subject must see what a team sees: `CLAUDE.md` → `AGENTS.md` →
   `@.folkways/instructions.md`, and skills in `.claude/skills/`. Most of what we test lives in those files.
2. **A/B arms by version.** Compare the skills at `main` with the skills in the working tree, not only with
   and without a plugin.
3. **Blind grading.** A judge that doesn't know which arm produced an answer.
4. **Isolation.** No host `CLAUDE.md`, user settings, hooks or MCP servers leaking into a run.
5. **Nothing leaves the machine** beyond the model calls the run already makes.
6. **Few new dependencies.** They'd sit next to code that ships into team repos.
7. Later: **Copilot CLI** as a subject.

## Options

| Option | Fit | Security | Verdict |
|---|---|---|---|
| **`claude plugin eval`** (first-party, in Claude Code ≥ 2.1.269) | Good graders (`regex`, `tool_used`, `tool_order`, `file_exists`, `llm` with 3 votes, `baseline`), with/without arms, JSON result, CI exit codes, cost ceiling. **But each run ignores the workspace's `CLAUDE.md` and `.claude/skills`** (probe below), so it can't test a team repo. Arms are plugin vs no plugin, not version A vs B | Best: no new dependency, temp home and config per run, git hooks and credential helpers off, Bash under the OS sandbox, case files hidden from the subject. **Publishes the HTML report to claude.ai by default** outside a Claude session; pass `--no-publish` | Use for **trigger evals** of skills packaged as a plugin. Not for team-repo behavior |
| **skill-creator** `evals/evals.json` | In-conversation iteration on one skill; the docs say neither tool reads the other's cases | First-party | Not a harness |
| **promptfoo** (npm, MIT) | Mature matrix runner, many providers, LLM rubric assertions | Large npm dependency tree (*inferred* from its provider list). OpenAI agreed to acquire it on 2026-03-09. Sends telemetry by default; with no key configured, hosted inference may receive prompts and responses | No: supply chain and data flow for a thin gain |
| **Inspect AI** + **inspect_swe** (Python, UK AISI, MIT) | Strong: `claude_code` agent inside a Docker sandbox, scorers, log viewer | Well-run project; adds a Python toolchain and Docker to a Node/Go repo | Revisit if we need sandboxed Bash-heavy tasks at scale |
| **DeepEval**, **Braintrust**, **LangSmith** | Metrics libraries or hosted platforms | Hosted platforms receive prompts and outputs; DeepEval collects telemetry unless opted out (medium) | No |
| **Own runner** (extend [07-eval/run.ts](../../research/evals/overlay-enforcement/run.ts)) | Exactly the shape we need: team repo from `folkways init`, arms by git ref, blind judge | No new dependency: Node built-ins, the `claude` CLI we already use, `git`. Isolation is ours to get right | **Chosen** |

### Probe: what `claude plugin eval` loads (high)

A case whose `scaffold_script` wrote `CLAUDE.md` (`@AGENTS.md`), `AGENTS.md` (a codeword) and
`.claude/skills/zebra-check/SKILL.md` into the workspace. The subject answered `NONE` for the codeword, and
its init message listed no `zebra-check` skill, only the plugin's and the built-in ones. That is by design
([docs, "How runs are isolated"](https://code.claude.com/docs/en/plugin-evals#how-runs-are-isolated): "Nothing
personal or project-level loads").

## Decision

Build a small runner in `packages/evals` (private, never published): Node built-ins only, unit-tested in CI
without model calls. Suites and results live in `research/evals/<suite>/`, next to the evidence they back.

Borrow `claude plugin eval`'s case format so cases read the same and can move there when they fit:

- `prompt.md`: frontmatter `runs`, `model`, `max_turns`, `timeout_seconds`, `allowed_tools`; the body is the
  prompt.
- `graders/<name>.md`: `type` `regex` (`pattern`, `flags`, `match`, `target`), `tool_used` (`tool`,
  `input_match`, `min`, `max`) or `llm` (body = PASS/FAIL rubric), plus `weight`.

What we add: `repo/` (files laid over the team repo `folkways init` made), a `suite.json` naming the arms
(each a git ref or the working tree), and a blind judge: each answer is graded alone, in random order,
with the arm name replaced by a run id.

### Isolation (what the runner does, and what it can't)

| Risk | Mitigation | Left over |
|---|---|---|
| Host `CLAUDE.md`, `AGENTS.md` | Team repos are built in the OS temp dir, outside any git repo | User memory `~/.claude/CLAUDE.md` still loads (no flag drops it without breaking OAuth). The runner records the init message and **flags runs whose skills differ** from the arm's set |
| User settings, hooks, permissions | `--setting-sources project,local` | Managed policy applies |
| MCP servers | `--strict-mcp-config` with no `--mcp-config` | |
| Writes, shell, network | `--tools` lists only the case's `allowed_tools`; default read-only (`Read`, `Glob`, `Grep`, `Skill`) | A case that grants `Bash` runs as you; use a container or CI |
| Session files | `--no-session-persistence` | |
| Attribution trailers | Kept in the raw output; graders see them | Write graders that ignore `Co-Authored-By:` |

For the cleanest runs, use a container or a cloud session (this one ran in one). `--bare` would drop user
memory too, but it also skips `CLAUDE.md` discovery and needs an API key.

## Sources

- [Test plugins with evals](https://code.claude.com/docs/en/plugin-evals), read 2026-09-30; `claude plugin eval --help`, `claude --help` (2.1.285)
- [OpenAI to acquire Promptfoo](https://openai.com/index/openai-to-acquire-promptfoo/); [Promptfoo is joining OpenAI](https://www.promptfoo.dev/blog/promptfoo-joining-openai/)
- [Inspect SWE: Claude Code](https://meridianlabs-ai.github.io/inspect_swe/claude_code.html)
