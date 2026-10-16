# folkways: status

**Started:** 2026-09-27
**Last updated:** 2026-10-02 (session 9: #25 merged with solo brain repos; handoff in [18](18-handoff-after-plan-16.md))
**Owner:** Summer Jordan
**Name:** **`folkways`** (pending a trademark check by Summer; see [02-naming.md](02-naming.md)).
**Stage: building v1.** CLI (`init`, `sync`, `upgrade`), 16 skills and overlay compile done (0.1.0, unpublished). Repo is a pnpm workspace (ADR 0003). Go hooks built, wired into `settings.json`, packaged per platform, with a release workflow. Claude Code plugin built (`packages/claude-plugin`). `folkways.yaml` (was `team.yaml`) with solo and team modes, both brain repos that hold no code; the plugin is quiet without it ([16](16-plan-config-and-solo.md), merged in #25). Nothing published yet.
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
| [07-research-overlay-enforcement.md](07-research-overlay-enforcement.md) | How mores can be enforced and folkways layered; basis of ADR 0001. Eval harness and results in [07-eval/](../../research/evals/overlay-enforcement) |
| [08-research-hook-runtime.md](08-research-hook-runtime.md) | Spec §9.3 answers, hook runtime options and benchmarks; basis of ADR 0002 |
| [09-research-skill-authoring.md](09-research-skill-authoring.md) | Skill structure, triggering and headless rules; basis of the authoring guide |
| [10-research-skill-at-imports.md](10-research-skill-at-imports.md) | `@` imports inside skills, for personal overlays (step 0) |
| [11-plan-hooks-build.md](11-plan-hooks-build.md) | Plan for the Go hooks: five PRs (done) and the checks on real machines (step 6, open) |
| [12-plan-repo-layout.md](12-plan-repo-layout.md) | The pnpm workspace move (ADR 0003): done on `claude/project-architecture-structure-700d4f`, before hooks PR 2 merged |
| [13-retro-git-guard-review-rounds.md](13-retro-git-guard-review-rounds.md) | Retro: git-guard grew through three review rounds; root causes, the skill changes (made) and their eval results |
| [14-research-eval-harness.md](14-research-eval-harness.md) | Eval harness: build or adopt (`claude plugin eval`, promptfoo, Inspect, …). Built `packages/evals`, no new dependencies |
| [15-research-license.md](15-research-license.md) | License options; basis of ADR 0004 |
| [16-plan-config-and-solo.md](16-plan-config-and-solo.md) | `folkways.yaml`, guided setup, solo repos: the plugin outside a team repo (built with parallel agents) |
| [17-handoff-pr25-review.md](17-handoff-pr25-review.md) | Handoff after session 8: reviewing #25, addressing findings, next steps (done) |
| [18-handoff-after-plan-16.md](18-handoff-after-plan-16.md) | Handoff after session 9: the brain-repo push gate, `doctor`, cleanup, then release prep |
| [20-research-app-repo-architecture.md](20-research-app-repo-architecture.md) | **Open (session 11):** engineers work in app repos, the team repo syncs. Options per flow, four composed architectures, the octo-team-context review, unknowns and questions |
| [21-handoff-app-repo-research.md](21-handoff-app-repo-research.md) | Handoff after session 11: verify 20's unverified claims and run the marketplace tests in a container with full internet access |
| [../adr/](../adr) | Framework decisions, one file each: 0001 overlay enforcement, 0002 hook runtime, 0003 repo layout, 0004 license (all accepted) |
| [../skill-authoring.md](../skill-authoring.md) | Skill skeleton, overlay IDs, headless rules, trigger eval |
| [../guide/](../guide) | User guide: concepts, getting started, how it works, the name. [../voice.md](../voice.md) is the writing voice |
| [../../research/](../../research) | Evidence: hook runtime benchmark ([bench/hooks/](../../research/bench/hooks)) and the overlay-enforcement eval ([evals/overlay-enforcement/](../../research/evals/overlay-enforcement)) |

## Decided (details in the notes, by round)

- **Scope (1–2):** generic template; The beta team tests it. Open source, owned by Summer. **No
  company-specific content in the framework.** v1 tools: **Claude Code (desktop + CLI) and Copilot CLI
  only** (11). No IDE integrations, no Copilot cloud agent.
- **Architecture (3, 6, 11):** framework published as npm package `@folkways/cli`, pinned in one team repo; `folkways sync`
  copies its text files in (12). Team repo: `folkways.yaml`, `knowledge/`, `state/`, `efforts/`, `overlays/`, team
  skills, committed `people/<handle>/` (`inbox/`, `work/`), gitignored `.local/` (`me.yaml`).
  **No separate personal repo** (11). Knowledge, skills, overlays via PR; state, efforts, people via direct
  rebased push. Promotion = `git mv` to `efforts/`. Every effort has a `00-STATUS.md`. Agent-space stays as-is.
- **Overlays (10, 11; ADR 0001):** mores are enforced by mechanism, folkways are layered by overlays. Team
  overlays can add and override and are compiled into skills by `folkways sync`. Personal overlays were
  removed (2026-10-03); skills take config from the `session-start` hook's context. **Overlays never touch framework files;** upgrades can't
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
3. ~~The plugin in a repo that isn't a team repo~~ Decided ([16](16-plan-config-and-solo.md)): without
   `folkways.yaml` the plugin and hooks stay silent except a one-time setup offer; `setup` configures a solo
   or team repo; solo is the team layout without `people/<handle>/`. Both are brain repos: sessions start
   there and code changes happen in worktrees of project repos.
4. **Push approval mode** (Summer, 2026-10-03): a config choice between asking before every push (today's
   M2 `ask` rule), never asking, and asking only for a push to a default branch. Open: where it lives
   (`folkways.yaml` as a more, `.local/me.yaml` only tightening it, like `review.trigger`) and how `sync`
   writes the matching permission rule. Becomes an effort in the command repo.

## Pending

- [x] License: Apache-2.0 ([ADR 0004](../adr/0004-license.md), [15](15-research-license.md)).
- [ ] Before going public: check the employment agreement and the employer's open-source policy ([15 §4](15-research-license.md)).
- [ ] Publish `folkways@0.0.2` (the unscoped npm placeholder) under Apache-2.0.
- [ ] Decide DCO sign-off before the first outside PR.
- [x] npm scope `@folkways` (Summer, 2026-09-28). Maybe a dedicated `folkways` GitHub org later.
- [ ] Later: publish a GitHub Action that runs `folkways sync` on dependency-update PRs (spec §9.7).
- [ ] Later: other harnesses (ADR 0003, Future direction). Not scoped. Solo repos are done ([16](16-plan-config-and-solo.md)); team overlays in the plugin are not.
- [ ] Dedicated session: architecture knowledge for the `adr` skill.
- [ ] Trademark check for `folkways` (USPTO/EUIPO); register folkways.dev and a GitHub handle (Summer).

## Built

| Part | Where | Notes |
|---|---|---|
| CLI `init`, `sync`, `upgrade`, `doctor`, `hook`, `version` | [packages/cli/src/](../../packages/cli/src) | TypeScript, npm package `@folkways/cli`, Node 24+. Sync is staged and atomic and refuses to overwrite. `doctor` checks config, versions, overlays, shadowing skills, `disableAllHooks` and the hook install (spec §7.2). `review`, `run` are stubs |
| Skills | [skills/](../../packages/content/skills) | 16: adr, commit, handoff, knowledge-audit, onboard, pr, qa, research, review, session-close, setup, side-quest, start, status, tests, ticket |
| Config and solo repos | [spec §2.3, §4.1](06-v1-spec.md) | `folkways.yaml` with `mode: team \| solo`, `git.commit_convention` (angular, pattern, none), `hooks:` toggles. A solo repo is a brain repo like a team repo: no project code, `projects_dir` required. Unconfigured repos: hooks and plugin silent except the one-time offer pointing to the brain repo (`~/.config/folkways/offered`). Guided `setup` (solo: creates the brain repo); skills use the personal root. The offer is per git top level, so each worktree of a repo gets its own |
| Overlay compile | [src/overlays.ts](../../packages/cli/src/overlays.ts) | `sync` compiles team overlays into skills (`compileSkillFiles`) |
| Config in context | [sessionstart.go](../../packages/hooks/internal/sessionstart/sessionstart.go) | `session-start` prints `folkways.yaml`, `.local/me.yaml`, the personal root and the review trigger; skills have a `## Config` section instead of a step 0 |
| Framework instructions | [instructions.md](../../packages/content/instructions.md) | Team `AGENTS.md` imports it via `@.folkways/instructions.md` |
| Settings fragment | [settings.json](../../packages/content/settings.json) | Push ask rule, `disableAllHooks: false`, one entry per hook and event, run as `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook <name>` |
| Team scaffold | [templates/](../../packages/cli/templates) | `folkways.yaml`, READMEs, `.gitignore`, `CLAUDE.md`, `AGENTS.md` |
| Hook contract suite | [packages/hooks/contract/](../../packages/hooks/contract) | Fixtures and harness for `git-guard`, `context-band`, `session-start`, `precompact`; the Go binary is built against it |
| Go hooks | [packages/hooks/](../../packages/hooks) | `folkways-hook` binary; all four hooks done: `git-guard` (with the trigger floor and inbox review lookup), `context-band` (state in `.local/hooks/context-band/`), `session-start`, `precompact`. `session-end` and the unclosed-session marker were dropped (too noisy: most sessions aren't worth banking). CI runs the full contract suite. Wired into team repos by `sync`; `init` sets pnpm's `publicHoistPattern`; `sync` warns when the binary or that setting is missing; `folkways hook <name>` runs it by hand. Packaged as `@folkways/hooks-<os>-<cpu>` ([build.ts](../../packages/hooks/scripts/build.ts)); the packed CLI pins them via [.pnpmfile.mjs](../../.pnpmfile.mjs). `FOLKWAYS_SKILL_NAMESPACE` makes the hooks name skills `folkways:<name>` (the plugin sets it) |
| Hook benchmark | [research/bench/hooks/](../../research/bench/hooks) | Node, Go and Rust candidates; Linux and macOS results |
| CI | [.github/workflows/ci.yml](../../.github/workflows/ci.yml) | Jobs filtered by package path, one required `ci` check; Linux and macOS. Smoke ([scripts/smoke.ts](../../scripts/smoke.ts)) on linux x64 and arm64 and macOS arm64 (darwin-x64 ships, untested in CI): packed CLI and hook tarballs, npm and pnpm, contract suite through the installed binary |
| Release | [.github/workflows/release.yml](../../.github/workflows/release.yml) | One package per changeset tag. Builds and checks the tarballs ([scripts/release.ts](../../scripts/release.ts)), then publishes after approval in the `npm` environment, with trusted publishing and provenance. Never run yet |
| Claude Code plugin | [packages/claude-plugin/](../../packages/claude-plugin) | Built from `packages/content`: skills rewritten to `folkways:<name>` and `${CLAUDE_PLUGIN_ROOT}` paths, instructions as session-start context, `hooks.json` generated from `settings.json`, `folkways-hook` for all four platforms, read from `@folkways/hooks`'s build (the plugin build never compiles Go), run by `run.sh`, which sets `FOLKWAYS_SKILL_NAMESPACE=folkways`. Quiet in synced team repos. No team overlays, no push ask rule. Released as `@folkways/claude-plugin` on its own tag; the root [marketplace](../../.claude-plugin/marketplace.json) installs it from npm. Dogfood from `dist/` |
| Eval runner | [packages/evals/](../../packages/evals) | Behavior evals: a team repo per arm (git ref) x case x run, `claude -p` isolated, blind LLM judge. Suites in [research/evals/](../../research/evals) ([14](14-research-eval-harness.md)) |
| Workspace | [packages/](../../packages) | pnpm workspace, per-package versions with changesets ([ADR 0003](../adr/0003-repo-layout.md)) |
| Tests | per package | vitest: CLI 412 pass; evals 26 pass (no model calls); hooks 22 pass, 155 contract cases skipped until a binary exists (all 155 pass against it), plus `go test`; claude-plugin 27 pass |

## Review follow-ups (PR #1, not blocking)

From the fresh-process review of [sumsolution/folkways#1](https://github.com/sumsolution/folkways/pull/1). Findings 1–4 are fixed there;
5–6 (Windows) are out of scope.

- [x] `sync` with no manifest (adopting an existing repo) overwrites existing `.folkways/` files; check `assetMap` destinations for clashes.
- [x] `sync` isn't atomic: a failure after copying leaves new files with the old manifest. Write the manifest last or stage in a temp dir.
- [x] Hook ownership matches `folkways hook ` anywhere in the command; match the `node_modules/.bin/folkways hook ` prefix. A hook group without a `hooks` array crashes with a TypeError.
- [x] `init .hidden` produces an invalid npm package name.
- [x] `npx @folkways/cli sync` outside the pinned install syncs whatever version npx resolves; `doctor` warns on the drift.
- [x] CI: smoke test the packed tarball (`pnpm pack` → install → `init --no-install` → `sync`); add `timeout-minutes` and a macOS runner.

## Next step

**Branch:** none open. `main` has everything through #25 (plan 16, solo brain repos).
**Handoff:** [18](18-handoff-after-plan-16.md): the brain-repo push gate and `doctor` with the cleanup (done), then the list below.

1. **Before the first release (Summer):** on npm, add a trusted publisher for `@folkways/cli`,
   `@folkways/claude-plugin` and the four `@folkways/hooks-<os>-<cpu>` packages (repo `sumsolution/folkways`, workflow `release.yml`, environment
   `npm`). A package that doesn't exist yet may need a first manual publish, or npm's pending-publisher
   setup, if npm offers it. In GitHub, create the `npm` environment with Summer as required reviewer.
2. Plan 11 step 6 on real machines: Gatekeeper on an npm-installed darwin binary, Copilot CLI (ADR 0002
   "Still open"), a lockfile made on macOS and installed on Linux.
3. Next evals: overlay precedence ported to `packages/evals`, trigger evals.
4. `review` / `run` (`doctor` done).
5. Publish: `@folkways/hooks@0.1.0`, then `@folkways/cli@0.1.0` and `@folkways/claude-plugin@0.1.0`.
6. Set up the beta team repo.

## Follow-ups (not ticketed)

- [x] Spec §4.2: `review.trigger` becomes a floor a person can only tighten (decided yes). Contract and
  fixtures done (`push-me-cannot-loosen-denied`); the Go implementation lands in plan 11, PR 2.
- [x] Spec §4.1: add `git.commit_pattern` (decided yes). Done in [16](16-plan-config-and-solo.md).
- [x] Spec §7.3 and `git-guard`: accept `people/<h>/inbox/<date>-review-<sha7>.md` as a review covering HEAD (keep the fallback). Contract and fixtures done; the Go implementation lands in plan 11, PR 2.
- [x] With `review.trigger: pre-push`, `git-guard` denied a direct push of the brain repo itself (team or solo). The push gate now skips the repo holding `folkways.yaml` (spec §7.3; fixtures `team-`/`solo-push-brain-repo-passes`).
- [x] `instructions.md` (Git) says `folkways review` must cover the SHA, with no solo exception; solo runs the `review` skill in a fresh subagent (as `pr` and `commit` say).
- [ ] `folkways review` / `run`: fill `{{repo}}`, `{{branch}}`, `{{base}}`, `{{sha}}`, `{{ticket}}`, `{{effort}}`, `{{output}}` in [prompts/review.md](../../packages/content/prompts/review.md).
- [x] `doctor` checks (done; also: an empty or `~` value turns a hook off, and a quoted `"off"` reads as false): non-boolean `hooks:` values in `folkways.yaml` (yaml.v3 reads `off`, `no` and `False` as false, but quoted `"false"`, `falsey` or `0` leave the hook on; `run.sh`'s no-binary line test disagrees with yaml.v3 on some of these), stale compiles (`staleOverlays`), shadowing skills (ADR 0001 M10), `disableAllHooks` (M11), `node_modules` left on the new version after a failed `upgrade` (`abandon()` restores package files only).
- [x] [settings.json](../../packages/content/settings.json): add `"disableAllHooks": false`, carried by `mergeSettings` (M11). Plan 11, PR 4.
- [x] `scripts/` is not in the `tsconfig.json` `include` (root `tsconfig.json`, checked by `pnpm run check` and CI); gitignore `.folkways-sync-*` in the team template.
- [x] Spec §3.1 hook ownership wording (plan 11, PR 4).
- [ ] GitHub ignores `{#id}`, so in-page skill anchors are dead in the web view.
- [ ] Research 10: a live `@` check (new session, marker-file skill) and Copilot CLI.
- [ ] Optional: route compiled-file staging through `fsOps` so rollback tests cover it.
- [x] ADR 0003 says content moves out of `packages/cli/assets/` when a second distribution needs it. Moved to
  `packages/content` (private); the CLI copies it into `assets/` at pack time (ADR 0003, 2026-10-03 update).

## Security (open)

- [ ] Enable secret scanning and push protection when the repo goes public.
- [x] Add a CI secret-scan job: TruffleHog pinned to a SHA (not gitleaks-action, which needs an org license). `secret scan` in ci.yml runs the CLI image pinned by digest (no third-party action), required through `ci`.
