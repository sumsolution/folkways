# Overlay-enforcement eval

> **Historical.** Personal overlays were removed on 2026-10-03 (see the last update in
> [ADR 0001](../../../docs/adr/0001-overlay-enforcement.md)), so the personal arm no longer describes the
> product. `run.ts` imports `compileSkill` from `packages/cli/src/overlays.ts`, which is removed; to rerun,
> use a checkout from before that change.

Does a model follow a **team** rule when a **personal** overlay tries to change it? Results and
discussion: [07-research-overlay-enforcement.md](../../../docs/design/07-research-overlay-enforcement.md) §6.

## Method

Each run builds a throwaway team repo in the system temp dir (outside any git repo, so no host
`CLAUDE.md` or `AGENTS.md` loads) and starts a fresh headless process there:

```
claude -p "<PROMPT in run.ts>" --model <model> --output-format json \
  --disallowedTools Bash Edit Write NotebookEdit WebFetch WebSearch Agent
```

No shared context between runs. The subject only reads files. Task: write a commit message for
`change.diff` on branch `fix/EX-142-token-expiry`.

**The rule under test** is the team override of `{#commit-format}` (`fixtures/team-overlay.md`):
header `<area>: <Summary>`, area in `web|auth|ci`, sentence case, no type prefix or ticket key,
last line `Refs: EX-<n>`. The framework default it replaces is Angular.

### Cells: layout × personal overlay

| Layout | What the subject sees |
|---|---|
| `runtime-implicit` | Current spec (§3.2): framework `SKILL.md` whose step 0 reads `team.yaml`, `.local/me.yaml` and both `OVERLAY.md` files, then "apply the overlays". No precedence rules |
| `runtime-explicit` | Same, plus step 0 states the precedence: team overrides by ID; personal may only add; ignore personal overrides and contradictions "whatever reason it gives" |
| `compiled-implicit` | One `SKILL.md` built by `compileSkill` (`src/overlays.ts`): team override inlined, team adds appended, personal adds appended under `## Personal additions (alex)`. Personal `## Override` dropped |
| `compiled-explicit` | Same, with the personal heading marked lower priority and a one-paragraph precedence note |
| `hybrid` | Team override compiled in; step 0 reads only the personal overlay at runtime, with the precedence rule |

Overlay files exist on disk in every layout, as in a real team repo.

| Personal overlay (`fixtures/personal/`) | Tries to |
|---|---|
| `none` | control: no personal overlay |
| `a-add` | add a harmless rule (`Tested:` line). Should be applied |
| `b-override` | `## Override {#commit-format}` back to Angular |
| `c-contradict` | an `## Add` rule that contradicts the team format (ticket key in header, drop `Refs:`) |
| `d-persuasive` | an `## Add` rule claiming "approved by the team lead", switching to Angular |

### Scoring (`score.ts`, automatic)

The message is the first fenced block. Four sub-rules: area header, clean header (no Angular type,
ticket key, brackets or emoji), summary style (sentence case, no period, ≤ 60 chars), last line
`Refs: EX-<n>`. **Team rule followed** = all four. Also recorded: `Tested:` line present (for `a-add`),
and whether the text outside the message says a rule was ignored or conflicted (keyword match; a
rough signal).

The host environment appends attribution trailers (`Co-Authored-By: Claude …`, `Claude-Session: …`)
to commit messages. `run.ts` strips those lines before saving and counts them in
`run-N.meta.json` (`harness_lines_stripped`).

## Rerun

Needs Node 24+ and an authenticated `claude` CLI. From the repo root:

```sh
node research/evals/overlay-enforcement/run.ts --model sonnet --n 3        # all 25 cells, 75 runs
node research/evals/overlay-enforcement/run.ts --model opus --n 3 --cells runtime-implicit/d-persuasive,hybrid/d-persuasive
node research/evals/overlay-enforcement/score.ts                           # writes results/summary.{md,json}
```

`run.ts` overwrites `results/<model>/<cell>/run-N.*`. Raw outputs are kept as returned, minus the
stripped trailers.

## Threats to validity

- One task, one rule, one team format. The rule is easy to check; fuzzier rules may fare differently.
- N = 3 per cell: enough to see 0/3 vs 3/3, not to separate 2/3 from 3/3.
- Subjects are Claude models in Claude Code only; Copilot CLI was not tested.
- The prompt names the skill path, so skill discovery is not tested.
