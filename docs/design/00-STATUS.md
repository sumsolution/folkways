# folkways: status

**Started:** 2026-09-27
**Last updated:** 2026-09-28 (session 3: TypeScript CLI, npm distribution, round 12)
**Owner:** Summer Jordan
**Name:** **`folkways`** (pending a trademark check by Summer; see [02-naming.md](02-naming.md)).
**Stage: building v1.** `@folkways/cli` with `init`, `sync`, `upgrade` done (0.1.0, unpublished). Hooks and skills next.
**Home:** this repo, [sumsolution/folkways](https://github.com/sumsolution/folkways). Design docs live in `docs/design/`.

## What this is

An open-source, team-oriented framework built from agent-space. It supports 3–8 person domain
engineering teams, encodes Summer's working methods as repeatable skills and automation, and
works across Claude Code and Copilot CLI. A domain engineering team at the author's employer is the beta team.

## Files

| File | Holds |
|---|---|
| [01-interview-notes.md](01-interview-notes.md) | Every decision and its reasoning, rounds 1–12. **The source of record.** |
| [02-naming.md](02-naming.md) | Name research and the `folkways` decision |
| [03-research-handoff-trigger.md](03-research-handoff-trigger.md) | When and how to hand off to a fresh session (Opus-checked) |
| [04-research-copilot-capabilities.md](04-research-copilot-capabilities.md) | Claude Code vs Copilot mechanisms (Opus-checked) |
| [05-research-repo-split.md](05-research-repo-split.md) | Team / personal split options; why the personal repo was dropped |
| [06-v1-spec.md](06-v1-spec.md) | **The v1 spec (draft).** Repos, npm distribution and sync, overlays, config, skills, hooks, flows |

## Decided (details in the notes, by round)

- **Scope (1–2):** generic template; The beta team tests it. Open source, owned by Summer. **No
  company-specific content in the framework.** v1 tools: **Claude Code (desktop + CLI) and Copilot CLI
  only** (11). No IDE integrations, no Copilot cloud agent.
- **Architecture (3, 6, 11):** framework published as npm package `@folkways/cli`, pinned in one team repo; `folkways sync`
  copies its text files in (12). Team repo: `team.yaml`, `knowledge/`, `state/`, `efforts/`, `overlays/`, team
  skills, committed `people/<handle>/` (`overlays/`, `inbox/`, `work/`), gitignored `.local/` (`me.yaml`).
  **No separate personal repo** (11). Knowledge, skills, overlays via PR; state, efforts, people via direct
  rebased push. Promotion = `git mv` to `efforts/`. Every effort has a `00-STATUS.md`. Agent-space stays as-is.
- **Overlays (10, 11):** team level can add and override; personal level can only add. Skills read overlay
  files. **Overlays never touch framework files;** upgrades can't clobber customizations.
- **Way of working (4):** session start and resume procedures; slim status doc; subagents for
  throughput, cheap model first, escalate to a stronger one; side quests go to the effort or `work/inbox/`;
  approval required for anything destructive or public. Headless runs create things for review only.
- **Handoff (10, 11):** 125K / 150K / 200K token bands, **tokens only**, via a transcript-reading hook. `PreCompact` is the backstop.
- **Git and review (3, 7):** Angular conventional commits with a module scope; squash PRs; branch
  `{type}/{TICKET}-{slug}`; push only on approval; draft PR when headless. Fresh-process review gate,
  trigger configurable (`on-demand` / `after-commit` / `pre-push`). Conventional Comments. Enforced
  agent-side only.
- **Research (5, 7):** source order, tool fallbacks (ask when a tool is unavailable), `NN-<topic>.md`
  with confidence scores; Opus re-check triggers. Agents never touch the user's main clone.
- **Tests and QA (8):** TDD; coverage is a signal; manual QA in local and dev only.
- **ADRs (9):** Confluence and in-repo both supported; suggest one at the moment of decision.
- **Tickets (4, 10):** the beta team's ticket-writing rules, generalized. Drafts go to the effort, falling back to `work/inbox/`.
- **Notifications (10):** GitHub @mention, falling back to Slack.
- **Everywhere:** succinct, direct, plain language. Every claim links to evidence. Every ported skill gets improved.

## Open questions

1. When findings are lifted to `knowledge/`: at session close or immediately? Decide with beta evidence.
2. Verify before building (spec §9.3): `PostToolBatch` in the installed Claude Code; transcript usage vs
   statusline; Copilot CLI `sessionEnd` / `PreToolUse` from `.claude/settings.json`; Copilot CLI subagent `model:`.

## Pending

- [ ] Choose a license (open source; owner's call). `package.json` says `UNLICENSED` until then.
- [x] npm scope `@folkways` (Summer, 2026-09-28). Maybe a dedicated `folkways` GitHub org later.
- [ ] Later: publish a GitHub Action that runs `folkways sync` on dependency-update PRs (spec §9.7).
- [ ] Dedicated session: architecture knowledge for the `adr` skill.
- [ ] Trademark check for `folkways` (USPTO/EUIPO); register folkways.dev and a GitHub handle (Summer).

## Built (session 3)

| Part | Where | Notes |
|---|---|---|
| CLI `init`, `sync`, `upgrade`, `version` | [src/](../../src) | TypeScript, npm package `@folkways/cli`, Node 24+. `hook`, `review`, `run`, `doctor` are stubs |
| Framework instructions | [assets/instructions.md](../../assets/instructions.md) | Team `AGENTS.md` imports it via `@.folkways/instructions.md`. Names skills not built yet |
| Settings fragment | [assets/settings.json](../../assets/settings.json) | Push ask rule only; no hooks yet |
| Team scaffold | [templates/](../../templates) | `team.yaml`, READMEs, `.gitignore`, `CLAUDE.md`, `AGENTS.md` |
| Tests | [test/](../../test) | vitest, 38 tests. Packed tarball checked end to end via `npm exec` |

Round 12 (see [01](01-interview-notes.md)): TypeScript instead of bash; npm distribution, pinned per team
repo; `@` import instead of a managed block. Spec §2, §3, §5, §7, §8 updated.

## Next step

**Branch:** `feat/cli-init-upgrade` (3 commits, not pushed, no PR yet). Push and open a PR, or keep building on it.

1. **Verify before building hooks** (spec §9.3): `PostToolBatch`, transcript `usage` vs statusline,
   Copilot CLI `sessionEnd` / `PreToolUse` from `.claude/settings.json`, hook `cwd` in both tools (hook
   commands are relative: `node_modules/.bin/folkways hook <name>`), and whether Copilot CLI loads the
   instructions twice (it reads both `AGENTS.md` and `CLAUDE.md`, which imports it).
2. **Hooks** (spec §7.1): `git-guard`, `session-start`, `session-end`, `precompact`, `context-band`
   as `folkways hook <name>`, with tests. Add `yaml` as a dependency when the first one reads `team.yaml`.
3. Ported skills → new skills → `review` / `run` / `doctor` → CI for the framework repo → publish 0.1.0 →
   set up the beta team repo.
