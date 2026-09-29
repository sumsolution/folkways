# folkways v1 spec

**Status:** draft, 2026-09-28. Proposal items settled. Round 12: npm distribution. Sources: [01](01-interview-notes.md) rounds 1–12,
[03](03-research-handoff-trigger.md), [04](04-research-copilot-capabilities.md), [05](05-research-repo-split.md).

## 1. Scope

| In v1 | Out of v1 |
|---|---|
| Claude Code (desktop + CLI), Copilot CLI | VS Code, JetBrains, any IDE integration; Copilot cloud agent; Cursor |
| One team per person | A person on several teams |
| Agent-side enforcement (skills + hooks) | CI checks in project repos |
| Manual QA (local, dev) | E2E test writing |
| macOS and Linux | Windows (round 12) |

**Hard rule:** the framework repo holds no company-specific content. Company and team values live in the
team repo (`team.yaml`, `overlays/`, team `skills/`). Examples in the framework are synthetic.

## 2. Repos

Two repos: the **framework** (versioned, public, published to npm as `@folkways/cli`) and one **team
repo** per team. Sessions run from the team repo. Code changes happen in project repos, in worktrees.

### 2.1 Framework repo (npm package `@folkways/cli`)

```
folkways/
├── AGENTS.md                # framework's own dev instructions (not shipped)
├── package.json             # bin: folkways; deps: yaml; engines: node >=24
├── CHANGELOG.md
├── src/                     # TypeScript. Compiled to dist/ by tsc on publish
│   ├── cli.ts               #   init, sync, upgrade, hook, review, run, doctor
│   └── hooks/*.ts           #   run as `folkways hook <name>` (§7.1)
├── assets/                  # text files copied into team repos by `sync`
│   ├── instructions.md      #   imported by the team AGENTS.md (§5)
│   ├── skills/<name>/       #   SKILL.md + references/
│   ├── prompts/*.md         #   fixed prompts for headless runs (review, knowledge-audit)
│   ├── settings.json        #   hook entries + permission rules merged into .claude/settings.json
│   └── workflows/*.yml      #   GitHub Actions (knowledge-audit)
├── templates/               # one-time scaffold for `init`: team.yaml, READMEs, .gitignore
│                            #   (`dot-x` → `.x`, `name.tmpl` → `name`, so they stay inert here)
└── test/                    # vitest
```

TypeScript is compiled because Node won't strip types in `node_modules` (verified 2026-09-28:
`ERR_UNSUPPORTED_NODE_MODULES_TYPE_STRIPPING`). Source imports use `.ts` extensions, so `src/` also runs
directly under Node 24 during development.

### 2.2 Team repo

```
team-repo/
├── AGENTS.md                # team text + `@.folkways/instructions.md` (§5)
├── CLAUDE.md                # @AGENTS.md (+ Claude-only extras)
├── package.json             # private; devDependencies: "@folkways/cli": "<exact version>";
│                            #   pnpm teams also pin `packageManager: pnpm@<version>`
├── package-lock.json        # or pnpm-lock.yaml; committed; install after clone and after upgrades
├── team.yaml                # team config (§4.1)
├── knowledge/               # durable facts, `Last verified:` + method          [PR]
├── state/                   # live trackers, one file per track                 [direct push]
├── efforts/<KEY>-<slug>/    # shared efforts, each with 00-STATUS.md            [direct push]
├── overlays/<skill>/        # team overlays: add + override (§3.2)              [PR]
├── people/<handle>/         # per person, committed                             [direct push]
│   ├── overlays/<skill>/    #   personal overlays: add only
│   ├── inbox/               #   side quests, ticket drafts with no effort
│   └── work/<KEY>-<slug>/   #   personal efforts; promote with git mv to efforts/
├── .local/                  # gitignored: me.yaml, session markers, private scratch
├── node_modules/            # gitignored; holds the pinned @folkways/cli
├── .folkways/               # SYNCED, framework-owned (§3)
│   ├── manifest.json        #   version + checksum of every synced file
│   ├── instructions.md  prompts/
├── .claude/
│   ├── settings.json        # team settings + framework hook entries (§3.1)
│   └── skills/<name>/       # synced framework skills + team skills
└── .github/workflows/       # synced knowledge-audit workflow + team workflows
```

**Push rules:** `knowledge/`, `overlays/`, team `skills/`, `team.yaml` and framework upgrades go by PR.
`state/`, `efforts/`, `people/` are pushed directly, rebased, linear. Every push still needs the user's
approval.

## 3. Distribution, sync and overlay isolation

Requirement (round 11): **overlays never touch framework files, and an upgrade never clobbers a customization.**

Code and text are delivered differently (round 12):

- **Code** (CLI, hooks) runs from the pinned package in `node_modules/`. Its version is set by the team
  repo's `package.json` and lockfile, so the whole team runs the same version.
- **Text** (instructions, skills, prompts, workflows) must be files in the repo, because both tools read
  skills from `.claude/skills/`. `folkways sync` copies them from the installed package. Upgrade PRs
  therefore show every framework change.

### 3.1 Ownership

| Path | Owner | `sync` behavior |
|---|---|---|
| `.folkways/**` | framework | replaced |
| `.claude/skills/<name>/` listed in manifest | framework | replaced |
| `.github/workflows/folkways-*.yml` | framework | replaced |
| `.claude/settings.json` hook entries whose command contains `folkways hook ` | framework | those entries replaced; all other keys and entries untouched |
| `.claude/settings.json` `permissions.ask` rules from the framework | framework | added if missing; never removed |
| `AGENTS.md` | team | never written after `init`; `sync` warns if the `@.folkways/instructions.md` line is missing |
| Everything else (`overlays/`, `people/`, team skills, `team.yaml`, `package.json`, …) | team / person | never written |

`folkways sync` (run from the team repo; copies the **installed** package version):

1. Checks every manifest file against its checksum, and looks for files added inside framework skill
   folders. **If a framework file was edited, it stops**, lists the files, and says to move the change into
   `overlays/`. `--force` discards the edits.
2. Refuses if a team skill in `.claude/skills/` has the same name as a new framework skill.
3. Replaces the framework files, merges settings, writes the manifest. It does not commit.

`folkways upgrade <version>`: refuses on a dirty tree → branch `chore/folkways-<version>` →
installs `@folkways/cli@<version>` as an exact dev dependency with the team's package manager → runs the
**new** version's `sync` → commits
`chore(folkways): upgrade to <version>`. The user reviews and opens the PR.

Dependabot or Renovate may bump the package instead; then someone installs and runs `folkways sync` on
that branch.

**Package managers** (round 12): the team repo uses **npm or pnpm**. `folkways` detects it from the
`packageManager` field, then the lockfile (`pnpm-lock.yaml` → pnpm), else npm, and uses it for every
install. `init` takes `--package-manager`, defaulting to whichever ran it (`npx` → npm, `pnpm dlx` → pnpm). `doctor` and the `SessionStart` hook warn when the installed version differs from the manifest.
**Later:** a published GitHub Action does this in CI (§9.7).

### 3.2 Overlays (ADR 0001)

Rules come in two kinds ([ADR 0001](../adr/0001-overlay-enforcement.md), research in [07](07-research-overlay-enforcement.md)):

- **Mores** must hold whatever the model or an overlay says (commit header format, push approval,
  review gate, main clone untouched, headless limits). They live in **config** (`team.yaml`, framework
  `settings.json`), change by PR, and are enforced by **mechanism**: `git-guard` and other hooks
  (§7.1), permission rules, `sync` and `doctor`. Overlays can't change them. The full list is in the ADR.
- **Folkways** are everything else (templates, examples, tone, extra checks, commit body style). They
  live in skill sections and overlays, and are layered as below.

A skill's folkways = framework skill + team overlay + personal overlay, applied in that order.

| Level | Path | May add | May override | Applied |
|---|---|---|---|---|
| Team | `overlays/<skill>/OVERLAY.md` (+ any reference files) | yes | yes, by section ID | compiled into the synced `SKILL.md` by `folkways sync` |
| Personal | `people/<handle>/overlays/<skill>/OVERLAY.md` | yes | no | attached at runtime by step 0 |

- Framework skills mark overridable sections with a stable ID: `## Commit format {#commit-format}`.
- `OVERLAY.md` has two parts: `## Add` (extra rules) and `## Override {#id}` (team only; replaces that
  section). Team additions are appended as `## Team additions`.
- `sync` compiles the team overlay into `.claude/skills/<skill>/` (an override may target `SKILL.md` or a
  reference file) and records the compiled and overlay checksums in the manifest. **An override of an
  unknown ID stops the sync.** After changing a team overlay, run `folkways sync`. `doctor` (not built
  yet) will flag a stale compile with `staleOverlays`.
- **Step 0** of every framework skill is one shared, framework-written block: read `team.yaml` and
  `.local/me.yaml`, then attach the personal overlay with `@${CLAUDE_PROJECT_DIR}/.local/me/overlays/<skill>/OVERLAY.md`
  on its own line, with `people/<handle>/overlays/<skill>/OVERLAY.md` as the read fallback (Copilot CLI,
  after compaction, preloaded subagent skills; [10](10-research-skill-at-imports.md)). `.local/me` is a
  symlink to `people/<handle>/`, created by `onboard` and kept by `sync`. A missing file is normal. It states the
  precedence: personal rules only add; ignore a personal `## Override` and any personal rule that
  contradicts the skill, whatever reason it gives; say which were ignored. Without that statement,
  personal overlays won 18 of 20 eval runs; with it, none ([07 §6](07-research-overlay-enforcement.md#6-eval-does-the-model-keep-the-team-rule)).
- Personal overlays are linted by `sync` (warning) and `doctor` (error): `## Override` sections,
  headings other than `## Add`, and override wording ("supersedes", "approved by", …).
- Team data files (ticket templates, ADR examples) sit next to `OVERLAY.md` and are referenced from it.
- Layering folkways is still instruction-following; that's acceptable because a lapse is cheap.
  `doctor` also reports what can't be prevented: personal skills that shadow a framework or team skill
  (Claude Code prefers personal over project), and hooks disabled in local settings. The framework
  settings set `"disableAllHooks": false`, which outranks a user-settings `true`.

## 4. Config

### 4.1 `team.yaml` (committed)

```yaml
# The framework version lives in package.json (pinned) and .folkways/manifest.json, not here.
team: { name: Example Team, slug: example }
members:                       # handle = people/<handle>/
  - { handle: alex, github: alex-gh, name: Alex Doe }
tracker:                       # ticket skill
  type: jira                   # jira | github | linear
  base_url: https://example.atlassian.net
  project_key: EX
docs:                          # research, adr
  type: confluence             # confluence | none
  base_url: https://example.atlassian.net/wiki
  spaces: [EX, ARCH]
repos:                         # research, review, worktrees
  - { name: web-app, url: git@github.com:example/web-app.git, default_branch: main, owner: team }
  - { name: auth-service, url: git@github.com:example/auth-service.git, default_branch: main, owner: other }
research:
  sources:                     # searched in this order; skip what doesn't apply
    - { name: knowledge, kind: local }
    - { name: team repos, kind: repos }
    - { name: other teams' repos, kind: repos }
    - { name: API docs portal, kind: url, url: https://docs.example.com, notes: "needs VPN" }
    - { name: confluence, kind: mcp, server: atlassian }
    - { name: jira, kind: mcp, server: atlassian }
    - { name: datadog, kind: mcp, server: datadog }
    - { name: slack, kind: mcp, server: slack }
    - { name: web, kind: web }
git:
  commit_convention: angular   # type(scope): subject
  scopes: [web-app, auth, ci]  # allowed module scopes; empty = any
  branch_pattern: "{type}/{ticket}-{slug}"
  merge: squash
review:
  trigger: pre-push            # on-demand | after-commit | pre-push
adr:
  targets: [confluence, repo]  # either or both
  confluence: { space: ARCH, parent_page: "Team ADRs", group_label: "Web" }
  repo_path: docs/adr
  reviewers: { min: 2, outside_team: 1 }
qa:
  allowed_envs: [local, dev]   # agents may drive a browser here without asking
notifications:
  primary: github              # @mention on the draft PR
  fallback: { slack_channel: "#example-team" }
models:                        # per tool; cheap first, escalate
  claude: { cheap: sonnet, strong: opus }
  copilot: { cheap: <model-id>, strong: <model-id> }
handoff:
  bands: [125000, 150000, 200000]
knowledge_audit:
  cadence: monthly
  tool: claude                 # claude | copilot (which CLI the Action runs)
```

### 4.2 `.local/me.yaml` (gitignored, per machine)

```yaml
handle: alex
projects_dir: /Users/alex/projects     # absolute; `~` is not reliably expanded
tool: claude                           # claude | copilot (default for `folkways run`)
review:
  trigger: pre-push                    # may differ from team (round 7)
models: {}                             # optional override of team.yaml models
notifications: { slack_user: "@alex" }
```

## 5. Instructions

- `AGENTS.md` is the source. `CLAUDE.md` holds `@AGENTS.md` so a `CLAUDE.local.md` can't turn it off ([04 §1](04-research-copilot-capabilities.md)).
- The team `AGENTS.md` imports the framework instructions with one line, `@.folkways/instructions.md`.
  Claude Code and Copilot CLI both expand `@` imports in `AGENTS.md` ([Copilot CLI docs](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions);
  not expanded in `*.instructions.md` or `GEMINI.md`). `AGENTS.md` stays fully team-owned (round 12).
- The **framework instructions** (`assets/instructions.md`, synced to `.folkways/instructions.md`) hold the way of working: session start and resume
  (round 4), status doc rules, subagents and escalation, side quests, approval list, headless rules,
  handoff bands, clone rules (round 7), writing style, evidence rule. It stays short and points to skills
  for detail.
- The team writes everything else in `AGENTS.md`.

## 6. Skills

Frontmatter: `name` and `description` only (portable across both tools). Every skill: step 0 reads config
and attaches personal additions (§3.2); writes only after approval; succinct output; claims link to evidence.

### 6.1 Ported (each one improved, not copied)

| Skill | From | Changes |
|---|---|---|
| `ticket` | beta team's ticket skill | Tracker from `team.yaml`. Team templates and examples → `overlays/ticket/`. Framework ships generic templates + synthetic examples. Drafts go to the effort folder, else `people/<h>/inbox/` (drop `.agents/tickets`). Adds: tables where useful, call out gotchas. |
| `adr` | `adr` | Method + default template (Issue → Context → Decision → Options → Consequences) + synthetic examples. Targets, group label, reviewers from `team.yaml`; gold-standard examples → `overlays/adr/`. Confluence and `docs/adr/NNNN-*.md`. Architecture knowledge section: **later** (dedicated session). |
| `session-close` | `session-close` | Homes: `knowledge/` (by PR), `state/`, effort `00-STATUS.md`, overlays, `people/<h>/inbox/`. Lists all side quests found. Clears the `.local/` unclosed-session marker. Claude auto-memory stays optional. |
| `knowledge-audit` | `knowledge-audit` | Tracker `state/knowledge-audit.md`. Headless mode = open a PR of proposals (§7.4), not a log block. |

`epic-retro` and `doc-eval` are not ported (agreed 2026-09-28).

### 6.2 New

| Skill | Purpose | Output |
|---|---|---|
| `setup` | Interview a team lead and write `team.yaml`, first `state/` and README files | `team.yaml`, PR |
| `onboard` | Set up one person on one machine (§8.2) | `.local/me.yaml`, `people/<h>/`, tool access |
| `start` | Session start: resume (read `00-STATUS.md` → git state → confirm next steps) or new (ticket/docs → research → create `00-STATUS.md`) | `00-STATUS.md` |
| `status` | Update an effort's `00-STATUS.md` at milestones; team roll-up and standup from `efforts/*/00-STATUS.md` and `state/`; promote an effort (`git mv` to `efforts/`) | status docs |
| `research` | Source order from `team.yaml`; clone rules; tool fallback = ask; subagents cheap-first; Opus re-check triggers | `NN-<topic>.md` with confidence and SHAs read |
| `review` | Code review method (round 3): intent first, verify by running, verdict + findings table, Conventional Comments | `NN-review-<sha7>.md` in the effort; with no effort, `people/<h>/inbox/<date>-review-<sha7>.md` |
| `commit` | Angular commit with module scope; branch naming | commit |
| `pr` | Requires a review covering HEAD; squash-title rules; draft vs ready; asks before push | PR (draft if headless) |
| `side-quest` | Write a finding or ticket draft to the effort or `people/<h>/inbox/`, and say so | file |
| `handoff` | Write `00-STATUS.md` for a fresh session; used at band 3 and by `PreCompact` recovery | `00-STATUS.md` |
| `tests` | TDD: failing test first for bugs, AC tests first for features; house style, call out bad tests | tests |
| `qa` | Manual verification in `qa.allowed_envs`; evidence (screenshots, traces, links) | QA notes in the effort |

## 7. Hooks and scripts

All hooks live in `.claude/settings.json`, which both Claude Code and Copilot CLI read ([04 §3](04-research-copilot-capabilities.md)).
Each hook entry runs the pinned package: `node_modules/.bin/folkways hook <name>`. If the package isn't
installed, the command fails; `doctor` and `onboard` make sure dependencies are installed.
Matchers work in both, so scripts still self-filter on command text only as a safety net.

### 7.1 Hooks

| Hook | Event | Does | Copilot CLI |
|---|---|---|---|
| `context-band` | `UserPromptSubmit`, `PostToolBatch` | Reads `transcript_path`; last main-chain assistant `usage` (input + cache read + cache creation); injects the band text via `additionalContext` when the band rises. Tokens only: 125K / 150K / 200K ([03](03-research-handoff-trigger.md)). | No token data; exits 0. Bands apply via instructions + `/context`. |
| `precompact` | `PreCompact` | Logs a missed band to `.local/`. | yes |
| `session-start` | `SessionStart` | If `source=compact`: inject "update `00-STATUS.md` now". If an unclosed-session marker exists: inject "offer session-close". Reminds the `start` procedure. | yes |
| `session-end` | `SessionEnd` | Writes `.local/sessions/<id>.unclosed`. | `sessionEnd` |
| `git-guard` | `PreToolUse` on Bash | `git commit`: checks the message format (Angular, allowed scope). `git push`: requires a review file for HEAD when trigger is `pre-push`. Denies `checkout`, `switch`, `pull`, `reset`, `rebase` in a main clone under `projects_dir` (worktrees only). | yes |

Push approval itself comes from a permission rule: `"ask": ["Bash(git push:*)"]` in `.claude/settings.json`.

**Hook command (proposed, ADR 0002):** `"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook <name>`. It runs a Go
binary from a `@folkways/hooks-<os>-<cpu>` optional dependency of `@folkways/cli`; pnpm teams set
`publicHoistPattern: ["@folkways/hooks-*"]`. The relative form breaks in Claude Code, whose hook cwd follows the Bash
tool's `cd` ([08 §1](08-research-hook-runtime.md)). Hooks read `team.yaml` from `CLAUDE_PROJECT_DIR`, else the cwd, and
always exit 0: Copilot CLI denies the tool call when a `preToolUse` hook exits nonzero. Behavior details and output shapes:
the contract suite, [test/hooks/README.md](../../test/hooks/README.md) (proposed, ADR 0002).

### 7.2 `folkways` CLI (`@folkways/cli`, run with `npx folkways` in a team repo)

| Command | Does |
|---|---|
| `init <path> [--package-manager npm\|pnpm]` | `npx @folkways/cli init <path>` or `pnpm dlx @folkways/cli init <path>`: scaffold a team repo from `templates/`, pin this version in `package.json`, install, sync, first commit |
| `sync [--force]` | §3.1 |
| `upgrade <version>` | §3.1 |
| `hook <name>` | run a hook (§7.1); reads the hook payload on stdin |
| `review [--repo <worktree>] [--effort <path>]` | §7.3 |
| `run <prompt> [--tool claude\|copilot] [--model]` | headless wrapper: `claude -p` or `copilot -p` with the right flags and dirs |
| `doctor` | checks config, tool access, hook install, version drift; prints fixes |

### 7.3 Review gate

- `folkways review` starts a **separate headless process** with the fixed prompt `.folkways/prompts/review.md`,
  filled with: repo, branch, base, HEAD SHA, ticket, effort path. The working agent can't frame it.
- Output: `<effort>/NN-review-<sha7>.md` with frontmatter `sha: <full SHA>`.
- `after-commit`: the `commit` skill runs it after each commit. `pre-push`: `git-guard` blocks the push
  until a review for HEAD exists. `on-demand`: never automatic.

### 7.4 Knowledge audit (GitHub Action)

`.github/workflows/folkways-knowledge-audit.yml`: monthly cron + manual dispatch. Runs the
`knowledge-audit` skill headless with `knowledge_audit.tool`, writes proposals, opens a **PR** on
`chore/knowledge-audit-<YYYY-MM>`, @mentions the knowledge owners. Needs one repo secret (API key or
Copilot token). Never merges.

### 7.5 Headless runs

Headless = `folkways run` or CI. Allowed: local files, draft PRs, @mention / Slack notify. Never:
destructive actions, `main`, external systems. Judgement calls are flagged in the output for review.

## 8. Flows

### 8.1 Team setup (once, by a lead)

1. `npx @folkways/cli init example-team` (or `pnpm dlx …`) → scaffold + install + sync + first commit. No framework clone
   needed (works behind a company npm mirror).
2. Open Claude Code or Copilot CLI in the new repo; run `setup`. It interviews for `team.yaml` fields,
   moves any existing team material into `overlays/` or team `skills/`, writes READMEs.
3. Push to a new GitHub repo; add the audit secret; protect `main` (linear history).

### 8.2 Onboarding (per person, per machine)

1. Clone the team repo; install (`npm ci` or `pnpm install --frozen-lockfile`); open the tool there; run `onboard`.
2. It writes `.local/me.yaml` (handle, projects dir, tool) and creates `people/<handle>/`.
3. Grants access to the projects dir, with approval:
   - Claude Code: add `projects_dir` to `permissions.additionalDirectories` in `~/.claude/settings.json` (absolute path).
   - Copilot CLI: prints an alias, e.g. `alias fw='copilot --add-dir /Users/alex/projects'` (`/add-dir` persistence is unreliable, [05](05-research-repo-split.md) fact 4).
4. Checks MCP servers and CLIs from `research.sources` and `gh`; lists what's missing.
5. Runs `folkways doctor`; commits `people/<handle>/` on approval.

### 8.3 Code change in a project repo

`git fetch` in the main clone → `git worktree add <projects_dir>/.worktrees/<repo>/<branch> -b <branch> origin/<default>`
→ work, `tests`, `commit` → `review` per trigger → `pr`. Parallel subagents each get a worktree off the
feature branch; the main agent merges. The main clone's checkout is never changed.

### 8.4 Upgrade

`npx folkways upgrade 1.1.0` → review the branch diff (framework files only) → PR → merge. Or merge a
Dependabot/Renovate bump after running `folkways sync` on its branch. Teammates reinstall after pulling.

## 9. Open and deferred

1. **When to lift findings to `knowledge/`:** session close vs immediately. Decide from beta evidence.
2. **Architecture knowledge for `adr`:** dedicated session.
3. **Verify before building** (checked 2026-09-29, [08 §1](08-research-hook-runtime.md)):
   - `PostToolBatch`: **exists** in Claude Code 2.1.284 (tested). Payload `tool_calls[]` with `tool_response` (a string, or a content-block array);
     `additionalContext` reaches the model. Copilot CLI has no `PostToolBatch`.
   - Transcript `usage` vs Claude Code's own count: **matches** `/context` (29,121 vs "29.1k", tested). The statusline
     doesn't run under `-p`, so it was not compared.
   - Copilot CLI honors `SessionEnd`/`sessionEnd` and `PreToolUse` from `.claude/settings.json`: **documented, not tested**.
     Differences: `SessionEnd.reason` and `SessionStart.source` values differ (no `compact`), no transcript path on these
     events, `tool_input` may be a JSON string, output fields are top-level, and `preToolUse` fails closed on a nonzero exit.
   - Hook cwd: Claude Code follows the Bash tool's `cd` (tested); Copilot CLI uses the project root. See §7.1.
   - Copilot CLI subagent `model:`: still open.
4. **Personal skills:** out of v1 (install to `~/.claude/skills/` and `~/.agents/skills/` by hand).
5. **Real git hooks** (`commit-msg`, `pre-push`) as a tool-independent backstop: deferred; round 7 chose agent-side only.
6. **Trademark and domain** for `folkways` (Summer). npm scope `@folkways` claimed by Summer (2026-09-28).
   A dedicated `folkways` GitHub org may replace `sumsolution` later.
7. **Published GitHub Action for sync (later):** runs on dependency-update PRs (Dependabot/Renovate),
   detects a `@folkways/cli` version change, installs with the team's package manager, runs `folkways sync`, and pushes the synced
   files to the PR branch. Removes the manual sync step.
