# folkways: status

**Started:** 2026-09-27
**Last updated:** 2026-09-29 (session 4: overlays, skills, hook runtime)
**Owner:** Summer Jordan
**Name:** **`folkways`** (pending a trademark check by Summer; see [02-naming.md](02-naming.md)).
**Stage: building v1.** CLI (`init`, `sync`, `upgrade`), 16 skills and overlay compile done (0.1.0, unpublished). Hooks (Go) next.
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
| [07-research-overlay-enforcement.md](07-research-overlay-enforcement.md) | How mores can be enforced and folkways layered; basis of ADR 0001. Eval harness and results in [07-eval/](07-eval) |
| [08-research-hook-runtime.md](08-research-hook-runtime.md) | Spec §9.3 answers, hook runtime options and benchmarks; basis of ADR 0002 |
| [09-research-skill-authoring.md](09-research-skill-authoring.md) | Skill structure, triggering and headless rules; basis of the authoring guide |
| [10-research-skill-at-imports.md](10-research-skill-at-imports.md) | `@` imports inside skills, for personal overlays (step 0) |
| [11-plan-hooks-build.md](11-plan-hooks-build.md) | **Next.** Plan for the Go hooks: decisions to confirm, five PRs, verification |
| [12-plan-repo-layout.md](12-plan-repo-layout.md) | Plan for the pnpm workspace move (ADR 0003), after hooks PR 2 |
| [../adr/](../adr) | Framework decisions, one file each: 0001 overlay enforcement, 0002 hook runtime, 0003 repo layout (all accepted) |
| [../skill-authoring.md](../skill-authoring.md) | Skill skeleton, overlay IDs, headless rules, trigger eval |
| [../guide/](../guide) | User guide: concepts, getting started, how it works, the name. [../voice.md](../voice.md) is the writing voice |
| [../../bench/hooks/](../../bench/hooks) | Hook runtime benchmark (Node, Go, Rust) with Linux and macOS results |

## Decided (details in the notes, by round)

- **Scope (1–2):** generic template; The beta team tests it. Open source, owned by Summer. **No
  company-specific content in the framework.** v1 tools: **Claude Code (desktop + CLI) and Copilot CLI
  only** (11). No IDE integrations, no Copilot cloud agent.
- **Architecture (3, 6, 11):** framework published as npm package `@folkways/cli`, pinned in one team repo; `folkways sync`
  copies its text files in (12). Team repo: `team.yaml`, `knowledge/`, `state/`, `efforts/`, `overlays/`, team
  skills, committed `people/<handle>/` (`overlays/`, `inbox/`, `work/`), gitignored `.local/` (`me.yaml`).
  **No separate personal repo** (11). Knowledge, skills, overlays via PR; state, efforts, people via direct
  rebased push. Promotion = `git mv` to `efforts/`. Every effort has a `00-STATUS.md`. Agent-space stays as-is.
- **Overlays (10, 11; ADR 0001):** mores are enforced by mechanism, folkways are layered by overlays. Team
  overlays can add and override and are compiled into skills by `folkways sync`; personal overlays can only
  add and are attached at runtime by step 0. **Overlays never touch framework files;** upgrades can't
  clobber customizations.
- **Hook runtime (ADR 0002):** hooks are a Go binary, `folkways-hook`, shipped as per-platform optional
  dependencies. It replaces the Node `folkways hook <name>` path (macOS arm64: Go about 2 ms npm / 9 ms pnpm,
  best Node path about 28 ms / 35 ms).
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
2. Spec §9.3 is answered by [08](08-research-hook-runtime.md) except: a Copilot CLI run confirming it ignores
   the nested `hookSpecificOutput` and runs repo hooks from the project root (ADR 0002, "Still open"); Copilot
   CLI subagent `model:` (not researched); whether Copilot CLI loads the instructions twice (it reads both
   `AGENTS.md` and `CLAUDE.md`, which imports it).

## Pending

- [ ] Choose a license (open source; owner's call). `package.json` says `UNLICENSED` until then.
- [x] npm scope `@folkways` (Summer, 2026-09-28). Maybe a dedicated `folkways` GitHub org later.
- [ ] Later: publish a GitHub Action that runs `folkways sync` on dependency-update PRs (spec §9.7).
- [ ] Later: solo projects and a Claude Code plugin; other harnesses after that (ADR 0003, Future direction). Not scoped.
- [ ] Dedicated session: architecture knowledge for the `adr` skill.
- [ ] Trademark check for `folkways` (USPTO/EUIPO); register folkways.dev and a GitHub handle (Summer).

## Built

| Part | Where | Notes |
|---|---|---|
| CLI `init`, `sync`, `upgrade`, `version` | [src/](../../src) | TypeScript, npm package `@folkways/cli`, Node 24+. Sync is staged and atomic and refuses to overwrite. `hook`, `review`, `run`, `doctor` are stubs |
| Skills | [assets/skills/](../../assets/skills) | 16: adr, commit, handoff, knowledge-audit, onboard, pr, qa, research, review, session-close, setup, side-quest, start, status, tests, ticket |
| Overlay compile | [src/overlays.ts](../../src/overlays.ts) | `sync` compiles team overlays into skills (`compileSkillFiles`) and lints personal overlays. Personal overlays attach at step 0 via `@` of `.local/me/...`, with a read fallback |
| `.local/me` link | [src/](../../src) | Symlink to `people/<handle>/`, created by `onboard`, kept by `sync` |
| Framework instructions | [assets/instructions.md](../../assets/instructions.md) | Team `AGENTS.md` imports it via `@.folkways/instructions.md` |
| Settings fragment | [assets/settings.json](../../assets/settings.json) | Push ask rule only; no hooks yet |
| Team scaffold | [templates/](../../templates) | `team.yaml`, READMEs, `.gitignore`, `CLAUDE.md`, `AGENTS.md` |
| Hook contract suite | [test/hooks/](../../test/hooks) | Fixtures and harness for `git-guard`, `context-band`, `session-start`, `session-end`, `precompact`; the Go binary is built against it |
| Hook benchmark | [bench/hooks/](../../bench/hooks) | Node, Go and Rust candidates; Linux and macOS results |
| CI | [.github/workflows/ci.yml](../../.github/workflows/ci.yml) | Linux and macOS; packed-tarball smoke test ([scripts/smoke.ts](../../scripts/smoke.ts)) |
| Tests | [test/](../../test) | vitest: 336 pass, 88 skipped |

## Review follow-ups (PR #1, not blocking)

From the fresh-process review of [sumsolution/folkways#1](https://github.com/sumsolution/folkways/pull/1). Findings 1–4 are fixed there;
5–6 (Windows) are out of scope.

- [x] `sync` with no manifest (adopting an existing repo) overwrites existing `.folkways/` files; check `assetMap` destinations for clashes.
- [x] `sync` isn't atomic: a failure after copying leaves new files with the old manifest. Write the manifest last or stage in a temp dir.
- [x] Hook ownership matches `folkways hook ` anywhere in the command; match the `node_modules/.bin/folkways hook ` prefix. A hook group without a `hooks` array crashes with a TypeError.
- [x] `init .hidden` produces an invalid npm package name.
- [ ] `npx @folkways/cli sync` outside the pinned install syncs whatever version npx resolves; `doctor` should catch the drift.
- [x] CI: smoke test the packed tarball (`pnpm pack` → install → `init --no-install` → `sync`); add `timeout-minutes` and a macOS runner.

## Next step

**Branch:** `main`, all work merged (PRs #4, #5, #6, #8, #9).

1. **Hooks** in Go: follow [11-plan-hooks-build.md](11-plan-hooks-build.md) through PR 2, then the
   **repo layout** move ([12-plan-repo-layout.md](12-plan-repo-layout.md), ADR 0003), then hooks PRs 3–5 on the new layout.
   Hooks PRs: contract amendments → Go module + `git-guard` → the other four hooks → wiring into team
   repos → packaging, smoke and release.
2. `review` / `run` / `doctor`.
3. Publish 0.1.0.
4. Set up the beta team repo.

## Follow-ups (not ticketed)

- [x] Spec §4.2: `review.trigger` becomes a floor a person can only tighten (decided yes). Contract and
  fixtures done (`push-me-cannot-loosen-denied`); the Go implementation lands in plan 11, PR 2.
- [ ] Spec §4.1: add `git.commit_pattern` (decided yes).
- [x] Spec §7.3 and `git-guard`: accept `people/<h>/inbox/<date>-review-<sha7>.md` as a review covering HEAD (keep the fallback). Contract and fixtures done; the Go implementation lands in plan 11, PR 2.
- [ ] `folkways review` / `run`: fill `{{repo}}`, `{{branch}}`, `{{base}}`, `{{sha}}`, `{{ticket}}`, `{{effort}}`, `{{output}}` in [assets/prompts/review.md](../../assets/prompts/review.md).
- [ ] `doctor` checks: stale compiles (`staleOverlays`), personal overlay lint as errors, shadowing skills (ADR 0001 M10), `disableAllHooks` (M11), `node_modules` left on the new version after a failed `upgrade` (`abandon()` restores package files only).
- [ ] [assets/settings.json](../../assets/settings.json): add `"disableAllHooks": false`, carried by `mergeSettings` (M11). Plan 11, PR 4.
- [ ] `scripts/` is not in the `tsconfig.json` `include`; gitignore `.folkways-sync-*` in the team template.
- [ ] Spec §3.1 hook ownership wording (plan 11, PR 4). GitHub ignores `{#id}`, so in-page skill anchors are dead in the web view.
- [ ] Research 10: a live `@` check (new session, marker-file skill) and Copilot CLI.
- [ ] Optional: route compiled-file staging through `fsOps` so rollback tests cover it.

## Security (open)

- [ ] Enable secret scanning and push protection when the repo goes public.
- [ ] Add a CI secret-scan job: TruffleHog pinned to a SHA (not gitleaks-action, which needs an org license).
