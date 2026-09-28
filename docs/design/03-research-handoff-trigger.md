# When should an AI coding agent session hand off to a fresh one?

**Researched:** 2026-09-27 (draft by research agent; re-checked by Opus the same day — see below)
**Method:** Primary-source fetch (Claude Code docs, GitHub/VS Code docs, Chroma report, Anthropic model announcement, NoLiMa paper) plus inspection of field names in one local Claude Code transcript. No experiments run.

## Recommendation

**Rule (inference — no vendor publishes a handoff threshold):** use absolute tokens in context, capped by a share of the window for small windows. Each band = the lower of the two numbers.

| Band | Tokens in context | Or % of window (200K models) | Agent behaviour |
|---|---|---|---|
| 1. Look for a stopping point | 150K | 50% (100K) | Finish the current unit; say "good place to hand off after this." |
| 2. No new multi-step work | 200K | 70% (140K) | Do not start a task that could not finish below band 3. Write `00-STATUS.md` at the next checkpoint and propose a fresh session. |
| 3. Hand off now | 300K | 90% (180K) | Stop at the nearest safe point, write `00-STATUS.md`, end the session. |
| Backstop | auto-compact fires | — | `PreCompact` hook = a band was missed. Log it; ensure `00-STATUS.md` is current. |

Why absolute tokens, not % alone: degradation tracks input length, not the fraction of the window used (§1). On a 1M window, "50%" would be 500K, well past where evidence shows loss. Band 2 at 200K matches the user's current habit.

**Mechanism in Claude Code — buildable today with documented features:**

| Part | What |
|---|---|
| Hook events | `UserPromptSubmit` (each user turn) and `PostToolBatch` (after each batch of tool calls, so long autonomous runs also get checked). |
| What it reads | `transcript_path` from the hook's JSON input. Take the last line where `type=="assistant"` and `isSidechain==false`; context = `message.usage.input_tokens + cache_read_input_tokens + cache_creation_input_tokens` (same formula the statusline uses for `used_percentage`). |
| Window size | Not in hook input. Either set it in the script (1M for Opus 4.7+/Sonnet 5/Fable on the Anthropic API, else 200K), or read it from a bridge file (below). |
| What it injects | JSON `{"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"Context at 212K tokens (band 2): do not start new multi-step work; propose a handoff at the next checkpoint."}}`. Same shape for `PostToolBatch`. |
| Avoid repeats | Record the last band announced per `session_id` (e.g. in `scratchpad_dir`); inject only when the band rises. |
| Do not use | `Stop` + `additionalContext`/`decision:block` — it forces the turn to continue, which is the opposite of stopping. |

Known limit: the transcript "may lag" the current turn (hooks docs), so the reading can be one response behind. Acceptable for band thresholds.

Optional bridge: the statusline script receives `session_id`, `context_window.used_percentage` and `context_window.context_window_size`; it can write them to a file keyed by `session_id` that the hook reads. Writing a per-session file from the statusline is a documented pattern (the docs' git-cache example); using it to feed hooks is not documented but uses nothing unsupported. The statusline only runs in the interactive UI and is debounced, so treat the bridge as a source for window size, not the primary token count.

**Copilot:** Copilot CLI shows usage via `/context` and has hooks (incl. `preCompact`), but no hook receives token counts and `userPromptSubmitted` cannot inject context. A `postToolUse` hook could inject `additionalContext` after parsing `transcriptPath` — only some events receive that path, so this is **unverified** as buildable. In VS Code, rely on the context-window control in the chat input and the same bands, applied by the user.

## Opus re-check — 2026-09-27

| # | Claim checked | Result | Source |
|---|---|---|---|
| 1a | No hook receives context/token fields | **Confirmed.** Only `SessionStart` on resume/fork gets `context_tokens` (cost of the resumed prompt). | [Hooks reference](https://code.claude.com/docs/en/hooks) |
| 1b | Every hook gets `transcript_path` | **Confirmed**, with a caveat: file is written async and may lag the current turn. | [Hooks reference](https://code.claude.com/docs/en/hooks) |
| 1c | Transcript records per-message usage | **Confirmed locally.** Assistant lines carry `message.usage` with `input_tokens`, `cache_read_input_tokens`, `cache_creation_input_tokens`, `output_tokens`, plus `isSidechain`. (Field names only inspected.) | Local `~/.claude/projects/*/<session>.jsonl` |
| 1d | Hooks can inject text to the model | **Confirmed.** `hookSpecificOutput.additionalContext` on `UserPromptSubmit`, `PostToolUse`, `PostToolBatch`, `SessionStart`, `Stop`; plain stdout also works for `UserPromptSubmit`. Capped at 10,000 chars. | [Hooks reference](https://code.claude.com/docs/en/hooks) |
| 1e | Draft said an automated nudge is "not cleanly available" | **Corrected.** Buildable via transcript parsing (see Recommendation). | as above |
| 1f | Statusline → file → hook bridge | **Feasible, not documented** as a hook input path. | [Statusline docs](https://code.claude.com/docs/en/statusline) |
| 1g | Auto-compact "at 95%" | **Corrected.** Default compacts at the model's context limit; native-1M models compact at ~967K. Tunable via `/autocompact` or `CLAUDE_CODE_AUTO_COMPACT_WINDOW`. | [Model config](https://code.claude.com/docs/en/model-config) |
| 2 | Chroma: "200K window shows serious loss at 50K" | **Corrected — not in the report.** The report shows performance grows "increasingly unreliable as input length grows" across 18 models, with no fixed threshold. | [Chroma, Context Rot](https://www.trychroma.com/research/context-rot) |
| 3 | 1M-window quality | **Partly answered.** Opus 4.6 scores 76% on MRCR v2 8-needle at 1M (Sonnet 4.5: 18.5%) — usable but lower than at shorter lengths. NoLiMa: 11 models fall below half their short-context score by 32K. No Anthropic data for current (4.7+/Sonnet 5) models on coding tasks. | [Anthropic, Opus 4.6](https://www.anthropic.com/news/claude-opus-4-6); [NoLiMa](https://arxiv.org/abs/2502.05167) |
| 4a | Copilot handoff = "Continue in Copilot CLI" on the cited GitHub page | **Corrected.** That page covers monitoring/steering sessions, not handoff. Handoff is documented in VS Code: it "carries the full conversation history" to another harness/agent. Custom agents can define `handoffs` buttons (label, agent, prompt, `send`). | [VS Code sessions](https://code.visualstudio.com/docs/agents/concepts/sessions); [Custom agents](https://code.visualstudio.com/docs/agent-customization/custom-agents) |
| 4b | "No reliable context indicator in VS Code" | **Corrected.** VS Code has a context-window control in the chat input with token breakdown, auto-compaction near the limit, and `/compact`. | [VS Code context](https://code.visualstudio.com/docs/agents/concepts/context) |
| 4c | "No Copilot hook equivalent to PreCompact" | **Corrected.** Copilot CLI hooks include `preCompact`, `sessionEnd`, `agentStop`; none carry token fields. CLI auto-compacts from ~80% and pauses at ~95%. | [Copilot hooks reference](https://docs.github.com/en/copilot/reference/hooks-reference); [CLI context management](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/context-management) |
| 5 | 50% table vs 50/70/90 bands | **Fixed.** One banded rule, above. | — |

## 1. Evidence on long-context quality

- Recall falls as tokens in context rise, for all models, at different rates. Anthropic gives no threshold. — [Anthropic, Effective context engineering](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) — **verified**.
- Chroma (July 2025, 18 models incl. Claude 4): accuracy drops as input grows, even on simple tasks, and worse with distractors; no specific token cutoff. — [Chroma](https://www.trychroma.com/research/context-rot) — **verified**.
- Lost in the Middle: accuracy is U-shaped by position; information in the middle is used worst. — [arXiv:2307.03172](https://arxiv.org/abs/2307.03172) — **high** (abstract-level read).
- NoLiMa: most tested models lose over half their short-context accuracy by 32K when the task needs more than word matching. — [arXiv:2502.05167](https://arxiv.org/abs/2502.05167) — **verified** (older models).
- Newer long-context models hold up far better: Opus 4.6 76% on MRCR v2 8-needle at 1M. — [Anthropic](https://www.anthropic.com/news/claude-opus-4-6) — **verified**. Retrieval benchmarks are not coding sessions; how far this transfers is **low** confidence.
- Band numbers (150K / 200K / 300K) are **inference**: they sit well below 1M, above where older models fail, and match the user's working habit. Revisit if Anthropic publishes guidance or if sessions past 200K show visible errors.

## 2. Compaction vs a clean handoff

- Claude Code compaction summarizes history and replaces earlier turns, then re-injects root `CLAUDE.md`, auto-memory and recently read files. — [Context window docs](https://code.claude.com/docs/en/context-window) — **verified**.
- Compaction is automatic and lossy, and keeps the same session's accumulated assumptions. A handoff with a written `00-STATUS.md` lets the agent choose what carries over and starts from a clean context. Neither Anthropic nor GitHub compares the two. — **medium** (inference).
- Copilot CLI docs advise a new session for unrelated tasks, or after many compactions when context seems lost. — [Copilot CLI context management](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/context-management) — **verified**.

## 3. Detection channels (Claude Code)

| Channel | Gets context usage? | Can tell the model? | Notes |
|---|---|---|---|
| Statusline | Yes: `context_window.used_percentage`, `total_input_tokens`, `context_window_size` (input tokens only) | No — display only | **verified**, [docs](https://code.claude.com/docs/en/statusline) |
| `/context` | Yes, on demand | Manual only | **verified** |
| Hooks + transcript parse | Yes, computed from `message.usage` | Yes, `additionalContext` | **verified** (docs + local field check); script not yet written or tested |
| `PreCompact` | No (only `trigger`, `custom_instructions`) | Can block compaction | Backstop only. **verified** |

## 4. Task-boundary practice

- End at a checkpoint (tests green, PR opened, subtask done); use the token band to decide whether to *start* the next unit, not to interrupt mid-unit. Matches the user's habit. — practitioner consensus — **medium**.

## Open questions

- Script not yet built or tested: confirm `PostToolBatch` fires in the user's Claude Code version and that the last-assistant `usage` matches the statusline's `used_percentage` within a few percent.
- Whether subagent turns ever appear in the main transcript with `isSidechain:false` (current local transcript: all assistant lines `false`, no subagents present).
- Copilot: whether a `postToolUse` hook can reach a transcript path to compute usage — not confirmed.
- No coding-task (not retrieval) degradation data for current 1M-window Claude models.
