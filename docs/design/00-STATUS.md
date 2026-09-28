# folkways: status

**Started:** 2026-09-27
**Last updated:** 2026-09-28 (session 3: skeleton, `init`, `upgrade`)
**Owner:** Summer Jordan
**Name:** **`folkways`** (pending a trademark check by Summer; see [02-naming.md](02-naming.md)).
**Stage: building v1.** Skeleton, `init` and `upgrade` done (0.1.0, untagged). Hooks and skills next.
**Home:** this repo, [sumsolution/folkways](https://github.com/sumsolution/folkways). Design docs live in `docs/design/`.

## What this is

An open-source, team-oriented framework built from agent-space. It supports 3–8 person domain
engineering teams, encodes Summer's working methods as repeatable skills and automation, and
works across Claude Code and Copilot CLI. A domain engineering team at the author's employer is the beta team.

## Files

| File | Holds |
|---|---|
| [01-interview-notes.md](01-interview-notes.md) | Every decision and its reasoning, rounds 1–11. **The source of record.** |
| [02-naming.md](02-naming.md) | Name research and the `folkways` decision |
| [03-research-handoff-trigger.md](03-research-handoff-trigger.md) | When and how to hand off to a fresh session (Opus-checked) |
| [04-research-copilot-capabilities.md](04-research-copilot-capabilities.md) | Claude Code vs Copilot mechanisms (Opus-checked) |
| [05-research-repo-split.md](05-research-repo-split.md) | Team / personal split options; why the personal repo was dropped |
| [06-v1-spec.md](06-v1-spec.md) | **The v1 spec (draft).** Repos, vendoring, overlays, config, skills, hooks, flows |

## Decided (details in the notes, by round)

- **Scope (1–2):** generic template; The beta team tests it. Open source, owned by Summer. **No
  company-specific content in the framework.** v1 tools: **Claude Code (desktop + CLI) and Copilot CLI
  only** (11). No IDE integrations, no Copilot cloud agent.
- **Architecture (3, 6, 11):** framework repo (versioned) **vendored** into one team repo by
  `folkways upgrade`. Team repo: `team.yaml`, `knowledge/`, `state/`, `efforts/`, `overlays/`, team
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

- [ ] Choose a license (open source; owner's call).
- [ ] Dedicated session: architecture knowledge for the `adr` skill.
- [ ] Trademark check for `folkways` (USPTO/EUIPO); register folkways.dev and a GitHub handle (Summer).

## Built (session 3)

| Part | Where | Notes |
|---|---|---|
| CLI `init`, `upgrade`, `version` | [bin/folkways](../../bin/folkways) | bash 3.2 + git + jq. `review`, `run`, `doctor` are stubs |
| Way-of-working block | [vendor/instructions.md](../../vendor/instructions.md) | Rendered into team `AGENTS.md`. Names skills not built yet |
| Settings fragment | [vendor/settings.hooks.json](../../vendor/settings.hooks.json) | Push ask rule only; no hooks yet |
| Team scaffold | [templates/](../../templates) | `team.yaml`, READMEs, `.gitignore`, `CLAUDE.md` |
| Tests | [tests/run.sh](../../tests/run.sh) | 55 checks: init, upgrade, edit detection, clashes, removals, no-op |

Spec changes: `folkways.version` dropped from `team.yaml` (manifest holds it); `upgrade --from`;
template renaming rule. Manifest also checksums the `AGENTS.md` managed block, so edits there stop an upgrade.

## Next step

1. **Verify before building hooks** (spec §9.3): `PostToolBatch`, transcript `usage` vs statusline,
   Copilot CLI `sessionEnd` / `PreToolUse` from `.claude/settings.json`.
2. **Hooks** (spec §7.1): `git-guard.sh`, `session-start.sh`, `session-end.sh`, `precompact.sh`,
   `context-band.sh`, with tests.
3. Ported skills → new skills → `review` / `run` / `doctor` → tag 0.1.0 → set up the beta team repo.
