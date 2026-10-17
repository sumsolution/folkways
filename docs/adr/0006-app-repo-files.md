# 0006. What files does an app repo carry?

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-17 |
| Owner | Summer Jordan |

## Issue

folkways is moving to app-repo-first: engineers start sessions in the app repo, not in a brain repo. Every
agent that works in the app repo must find its instructions, skills and settings in that repo's clone
([20 §0][r-0]). Four agents count as first class: Claude Code, Copilot CLI, the Copilot cloud agent and
Copilot code review. Each tool documents its own paths, and Copilot documents both its own (`.github/…`) and
Claude's. Before `sync` learns to write into app repos, the file set has to be fixed: one format, or one per
tool? What files does an app repo carry?

## Context

- **What `sync` writes today is Claude-format only:** `.claude/skills/<skill>/`, `.claude/settings.json`,
  and a check that `AGENTS.md` imports the framework instructions
  ([sync.ts L108, L143, L334](https://github.com/sumsolution/folkways/blob/ec8f1fc/packages/cli/src/sync.ts#L108)).
  Nothing under `.github/`.
- **Copilot CLI reads the Claude format.** Its docs: it reads `.claude/settings.json` and
  `.claude/settings.local.json` for "the shared cross-tool subset … `enabledPlugins`,
  `extraKnownMarketplaces`, and `hooks`" ([cli-config-dir-reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference),
  [20 §7c][r-7c]). Its 1.0.98 changelog: it loads skills from `.claude/skills` and expands `@` imports in
  `AGENTS.md` and `CLAUDE.md` ([20 §7a][r-7a]). Tested offline on 1.0.98: it registered a marketplace from
  `.claude/settings.json` and installed a Claude-format plugin unchanged. Not tested: a skill or hook
  running in a logged-in session ([21, open item 1][h-open]).
- **The Copilot cloud agent reads the Claude format.** Docs: `AGENTS.md`, `CLAUDE.md`, and skills in
  `.claude/skills` ([custom-instructions-support](https://docs.github.com/en/copilot/reference/custom-instructions-support),
  [about-agent-skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills), [20 §7b][r-7b]).
  One hands-on run on a throwaway repo: it followed `AGENTS.md` and `CLAUDE.md`, listed skills from
  `.claude/skills`, ran hooks from `.claude/settings.json`, and installed a plugin enabled only in
  `.claude/settings.json` ([20 §7b hands-on][r-hands]). The hooks and plugin results contradict its docs,
  which say the cloud agent "does not load `settings.json`" and reads hooks only from `.github/hooks/*.json`
  ([hooks-reference](https://docs.github.com/en/copilot/reference/hooks-reference)). Skills were
  self-reported, not invoked.
- **Copilot code review reads part of it.** Docs: `AGENTS.md` and `CLAUDE.md`, plus `REVIEW.md`; review
  criteria from `.github/copilot-instructions.md`, project context from root `AGENTS.md`. Skills: the
  concept page lists `.claude/skills`, the how-to names only `.github/skills` ([20 §7b][r-7b]). In the
  hands-on run it used a `code-review` skill; the record doesn't say which directory that skill was in
  (unconfirmed).
- **Claude Code** reads only its own format, locally and in cloud sessions, where repo-enabled plugins are
  not installed ([20 §7b][r-cc]).
- **Prior art:** one `CLAUDE.md` and one `.claude/skills/` tree serve Claude Code, Copilot CLI and OpenCode
  in woditschka/agentic-coding-reference ([20 §7d][r-7d]). Generators that write every tool's files
  (Ruler, rulesync) exist, and Ruler gitignores its output by default, so clones and cloud agents miss it
  ([20 §7c][r-7c]).

## Decision

**An app repo carries Claude-format files only: `CLAUDE.md` / `AGENTS.md`, `.claude/skills/` and
`.claude/settings.json`. folkways writes no parallel Copilot-format copies (`.github/copilot-instructions.md`,
`.github/skills/`, `.github/hooks/`, `.github/copilot/settings.json`).**

Rationale: every first-class agent reads the Claude format for instructions, and three of four are
documented to read it for skills. One copy has nothing to drift and gives the Copilot tools no second source
to load twice. The gaps (code review skills and criteria, cloud-agent settings) are named below with what
would make us add a Copilot-format file. Adding a file later costs one `sync` change; removing mirrors from
many app repos costs more.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| **1** | **Claude format only** | One copy, nothing to keep in step; matches what `sync` writes today; no double loading in Copilot CLI | Relies on Copilot reading another tool's paths; for the cloud agent's settings, on behaviour its docs deny | Least to maintain, with a known gap on code review and undocumented cloud-agent settings. |
| 2 | Claude format plus Copilot mirrors under `.github/` | Every tool reads a path its own docs name | Two copies per file to keep in step; Copilot CLI scans both `.claude/skills` and `.github/skills` and reads hooks from `.claude/settings.json` and `.github/hooks`, so the same skill or hook can load twice (unconfirmed); more files in every app repo | Buys documented paths at the cost of drift and double loading. |
| 3 | Tool-neutral paths: `AGENTS.md` and `.agents/skills/` | Not tied to one vendor; Copilot reads `.agents/skills` | Claude Code reading `.agents/skills` is not established in the research (unconfirmed); settings and hooks have no neutral path, so `.claude/settings.json` stays anyway | Neutral in name, but still needs Claude paths and risks losing Claude Code. |

## Consequences

- `sync` keeps writing the Claude-format set when it targets app repos. No new writers for `.github/`.
- **Revisit** (add a Copilot-format file, not a full mirror) if any of these hold:
  - a logged-in Copilot CLI test shows `.claude/skills` or `.claude/settings.json` hooks not running
    ([21, open item 1][h-open]);
  - Copilot code review is shown to skip `.claude/skills`, or folkways needs review criteria there:
    `REVIEW.md` or `.github/copilot-instructions.md` would be the candidate;
  - the cloud agent stops honouring `.claude/settings.json`, matching its docs.
- **Open, untested:** whether the cloud agent and code review expand the `@.folkways/instructions.md` import
  in `AGENTS.md` (unconfirmed). If they don't, the framework rules don't reach them.
- **Lost:** Copilot's path-scoped `.github/instructions/*.instructions.md` (`applyTo` globs). folkways
  doesn't use them today.
- **Related, decided elsewhere:** how hooks reach each surface (ADR 0007); who owns `AGENTS.md` (ADR 0008);
  whether files arrive by hub PR or each repo's own `sync` ([20 §5][r-5], A vs C).

[r-0]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L15-L36
[r-5]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L215-L282
[r-7a]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L309-L332
[r-7b]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L342-L361
[r-cc]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L371-L396
[r-hands]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L398-L431
[r-7c]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L475-L546
[r-7d]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#L548-L634
[h-open]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/21-handoff-app-repo-research.md#L43-L57
