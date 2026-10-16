# Engineers work in app repos; the team repo syncs: options

**Researched:** 2026-10-16 (session 11). Main agent, no subagents. Tool facts come from primary docs where
fetched (marked ✓); the rest come from search summaries (marked ~) because this container's network
blocks most sites. **Status: options for Summer to review. Nothing decided.** Summer's answers to §8 and
the marketplace tests were added the same day (§0, §7a). Session 13 checked the `~` claims against
primary sources (§7c). Session 14 ran locally with an open network, opened the pages left blocked
(§7c, no `~` left there), and read the three sources the seylox sequel cites (§7d).

**Question:** engineers should keep starting sessions in their app repos, not in a brain repo. The team
still has a central agent repo that holds shared knowledge and understands every team repo, and acts as
the sync between them ([octo-team-context](https://github.com/summer-jordan/octo-team-context) is the
example). What shapes could that take, and what would each one change in folkways?

## 0. Summer's answers (2026-10-16) and what they imply

| # | Answer | Implication for the options |
|---|---|---|
| 1 | Concerns about the brain repo: an unconventional workflow; cloud environments can't reach several repos at once; outside contributors clone only the app repo and lack context; **most impactful: Copilot agents running in GitHub would lack the brain repo's skills and context** | Every agent that touches the app repo must find what it needs **in that repo's clone**. Anything that lives only on a person's machine or in another repo fails concerns 2–4 |
| 2 | App repos should contain **all the context, skills and workflows** an agent needs to operate in that repo. First-class: Claude Code, Copilot CLI, **and Copilot running in GitHub** (code review, cloud coding agent) | **Scope change:** spec §1 excluded the Copilot cloud agent. The cloud agent and code review don't install plugins (to verify), so plugin-only delivery (O4, O5, architectures B and D) can't be the whole answer. Files committed to the app repo are required; a plugin can only add to them |
| 3 | Copilot CLI parity is required; test it now | Tested, §7a |
| 4 | Where efforts and reviews live: not decided; think it through carefully | Stays open (§4.4). Answer 2 favors the app repo or the tracker for anything an agent needs |
| 5 | Bots may open PRs into app repos; to discuss | O2 (hub opens PRs) and I1 stay on the table |
| 6 | Per-repo context generation and architecture synthesis: **first phase** | Inbound (I1/I2) and synthesis are in phase 1, not later |

What the answers rule out or narrow, as a reading of them rather than a decision:

- **D (per machine)** fails answer 2: nothing lands in the repo.
- **B (marketplace only)** fails answer 2 for Copilot in GitHub and for any tool without plugins. B
  survives as an add-on: the repo carries committed files for every agent, and a plugin adds what
  committed files can't (hook binaries, for example) for local Claude Code and Copilot CLI.
- **A (CI hub writing files into the repo)** and **C (package per repo)** both put files in the repo. They
  differ in who writes them: a hub bot by PR (A) or each repo's own `folkways sync` (C).
- New open question: **how hooks reach agents in a clone with no `node_modules`** (non-Node repos, the
  cloud agent). Today the hook binary comes from npm. Candidates: a plugin (local only), the cloud
  agent's setup steps, or a committed launcher that downloads a pinned binary. Not researched yet.

## 1. What changes, and what stays

Today (spec §2, plan 16): sessions start in the brain repo; code changes happen in worktrees of app repos
under `projects_dir`; knowledge, state, efforts, reviews and people live in the brain repo; `git-guard`
keeps the main clone of every app repo untouched.

The new premise flips the starting point. The engineer opens Claude Code or Copilot CLI **in the app
repo**, as they do without folkways. The team repo becomes a hub that the app repos sync with.

What can carry over unchanged, whatever we pick:

| Carries over | Why it doesn't depend on where sessions start |
|---|---|
| Skills as methods (commit, review, tests, ticket, adr, research, qa, pr) | They describe how to work, not where files live. Paths inside them change |
| Mores vs folkways (ADR 0001) and team overlays | Still the right split. Only the delivery changes |
| Go hooks (ADR 0002) | `git-guard` commit checks and `context-band` work in any repo |
| Overlay compile, `doctor`, the eval runner | Same jobs, new inputs |
| npm via the company mirror (round 12) | The constraint that made us pick npm still holds |

What breaks or needs a new answer: the brain-repo push exemption, the main-clone guard (the engineer now
works in the main clone), where reviews, efforts and personal notes live, the setup offer text, and the
idea of a "personal root".

## 2. The three flows

Any design has to answer three questions. They can be answered separately, which is why section 4 lists
options per flow before section 5 combines them.

| Flow | Direction | Carries |
|---|---|---|
| **Outbound** | team repo → app repos | Framework (skills, instructions, hooks, settings), team conventions and overlays, a map of the team's repos |
| **Inbound** | app repos → team repo | Per-repo context (stack, deps, dev setup), decisions, learnings from sessions, effort status, reviews |
| **Lateral** | app repo session ↔ team knowledge | "How does service B handle auth?" asked from repo A. Read, at session time |

## 3. The example: octo-team-context

Read at its single commit, `2e9869c`. A GitHub Actions hub, Python stdlib scripts, one PAT.

**How it works**

```
                 ┌──────────── hub repo ─────────────┐
                 │ agents/AGENTS.md                  │
                 │ workflows/*.yml   (to push out)   │
                 │ repo-sync-config.json (registry)  │
                 │ repos/<name>/{DEPS,STACK,DEV}.md  │
                 │ architecture/*.md (AI synthesis)  │
                 └───┬───────────────────────▲───────┘
   Outbound Sync:    │ Contents API PUT,     │ Inbound Pull: on repository_dispatch
   on push to agents/│ direct to default     │ (+ daily reconcile), reads
   or workflows/     │ branch, 9 files       │ .team-context/this-repo/*
                 ┌───▼───────────────────────┴───────┐
                 │ app repo                          │
                 │ AGENTS.md CLAUDE.md GEMINI.md     │  ← same file, 6 names
                 │ .cursorrules .windsurfrules ...   │
                 │ .github/workflows/ (3 from hub)   │
                 │ .team-context/this-repo/          │  ← AI-written on each push to main
                 └───────────────────────────────────┘
```

- **Outbound:** one `agents/AGENTS.md` copied to six filenames plus three workflows, written straight to
  each repo's default branch through the Contents API. Skips files whose blob SHA already matches.
- **Inbound:** on every push to `main`, a workflow in the app repo decides (one cheap model call) whether
  context is stale, runs a 10-step agentic loop (`list_files`, `read_file`, `search`) to write `STACK.md`
  and `DEV.md`, and builds `DEPS.md` deterministically from manifests. A second workflow fires
  `repository_dispatch` at the hub, which pulls the files into `repos/<name>/`.
- **Synthesis:** daily, the hub summarizes repos by type into `architecture/groups/*.md`, then a system
  overview, plus a deterministic repo map and Mermaid dependency graph. These go **by PR**.
- **Ops:** dry-run on every workflow, issues opened on failure, a daily reconcile for drift, SHA caches to
  skip unchanged inputs.

**What it gets right, for us**

| Strength | Relevance |
|---|---|
| Engineers install nothing; it works in any tool that reads one of the six files | Matches "keep working in app repos" exactly |
| Per-repo context lands in one findable place, and the architecture docs are reviewed by PR | This is the "understands all repos" part we want |
| Deterministic where it can be (deps, diagram), AI where it must be | Same instinct as our mores/folkways split |
| Registry in one file (`repo-sync-config.json`) | We already have one: `repos:` in `folkways.yaml` |

**What doesn't fit folkways as written** (observed in the code)

| Issue | Where | Why it matters to us |
|---|---|---|
| Overwrites `AGENTS.md` and `CLAUDE.md` wholesale | `sync-outbound.yml`, `reconcile.yml` | Repos lose repo-specific instructions; the daily reconcile silently reverts local edits. Our rule since round 12: `AGENTS.md` is team-owned |
| Writes straight to the default branch, no PR | `sync-outbound.yml` | Bypasses review and breaks on protected branches (only the context workflow falls back to a PR) |
| One PAT with `repo`, `workflow` and secrets write, copied into every app repo | `setup-downstream-secrets.yml` | Any repo's workflow can write to every other repo and the hub. Large blast radius; security teams will push back |
| App-repo workflows fetch scripts from the hub's `main` at run time, unpinned | `generate-repo-context.yml`, `update-readme.yml` | A hub change runs in every repo's CI with that PAT. No pinning, no review in the app repo |
| Actions pinned to tags, not SHAs | all workflows | Our repo requires SHA pins |
| A model call and a bot commit on most pushes to `main` | `generate-repo-context.yml` | Cost and history noise in every app repo |
| No skills, hooks or enforcement, only an instruction file | — | Covers about a tenth of what folkways ships |

**Takeaway:** the shape (registry, outbound standards, inbound per-repo context, hub-side synthesis) is
sound and close to what you described. The mechanics (wholesale overwrite, direct writes, a shared
write-everywhere token) are the parts to change.

## 4. Options per flow

### 4.1 Outbound: how team and framework content reaches app repos

| | Mechanism | App-repo footprint | Review | Tools covered | Notes |
|---|---|---|---|---|---|
| **O1. Hub pushes files** | Hub CI writes into each repo (octo) | Synced files + workflows | None (direct write) | Any | Simplest; fails the "AGENTS.md is team-owned" rule unless it writes only a managed file |
| **O2. Hub opens PRs** | Hub CI opens a PR per repo when content changes ([repo-file-sync-action](https://github.com/BetaHuhn/repo-file-sync-action) ✓, [team-ai-sync](https://dev.to/paladini/sync-ai-agent-rules-across-repositories-with-github-actions-5b9h) ✓) | Synced files in a folkways-owned folder | App-repo owners merge | Any | Needs a GitHub App or token with PR write on team repos. Upgrades visible in each repo, as `sync` does today |
| **O3. App repo pins and pulls** | Each app repo has `@folkways/cli` pinned and runs `folkways sync`, pulling team content too | `package.json` + lockfile + synced files | Upgrade PR per repo | Any | Reuses today's CLI. Odd in Go, Java or Python repos (a Node manifest in each). N repos to bump |
| **O4. Team repo is a plugin marketplace** | Team repo carries `.claude-plugin/marketplace.json`; its team plugin = framework + compiled overlays + team skills + hooks. App repo's `.claude/settings.json` adds `extraKnownMarketplaces` and `enabledPlugins` ✓ | About 10 lines of settings | Team repo PRs; plugin auto-updates or pins a ref ✓ | Claude Code ✓; Copilot CLI partly (below) | Nothing else copied into app repos. Teammates are prompted to trust the folder and the marketplace ✓ |
| **O5. Per machine, no app-repo footprint** | Each engineer installs the plugin; `session-start` checks whether the current repo is listed in the team's `folkways.yaml` and loads team context | None | Team repo PRs | Claude Code, Copilot CLI | Only folkways users get anything; the repo itself carries no standards for other tools or newcomers |
| **O6. Import or add-dir** | App repo `AGENTS.md` imports from a local clone of the team repo, or the engineer runs with `--add-dir <team repo>` | One import line, or none | Team repo PRs | Claude Code ✓; Copilot CLI `--add-dir` ✓ ([05](05-research-repo-split.md)) | Import paths are machine-specific and trigger an approval dialog ✓. `--add-dir` loads skills but not `CLAUDE.md` unless an env var is set ✓, and settings-based dirs don't load skills, by design ✓ (§7c) |

**O4 in detail, since it is new to us.** Claude Code reads plugin marketplaces from a repo's own
settings ✓:

```json
// app-repo/.claude/settings.json
{
  "extraKnownMarketplaces": {
    "example-team": { "source": { "source": "github", "repo": "example/example-team-agents" } }
  },
  "enabledPlugins": { "folkways-team@example-team": true }
}
```

- The marketplace registers after the engineer accepts the workspace trust dialog; in `-p` runs only in a
  folder already trusted ✓. A plugin listed by relative path inside the marketplace installs with it; one
  pointing elsewhere needs a manual install ✓. So the team plugin should sit inside the team repo.
- A private marketplace needs each user to have read access, cloned with their git credentials ✓. The
  team repo is in the company org, so the "can't clone outside the org" constraint (round 12) doesn't bite.
- Managed settings can force the same thing org-wide, and `strictKnownMarketplaces` can allowlist it ✓.
  Useful if a security team wants control.
- The plugin build we already have (`packages/claude-plugin`) does most of this. The new step is a
  **team** plugin: framework plugin + compiled overlays + team skills, built in the team repo's CI.
- **Copilot CLI:** reads the same `extraKnownMarketplaces` block from the app repo's
  `.claude/settings.json` (tested, 1.0.98) and from `.github/copilot/settings.json` (its changelog), and
  installs a Claude-format plugin unchanged. See §7a. Enterprise-wide registration is in `.github-private`
  `copilot/managed-settings.json` ✓ (§7c)
  ([changelog](https://github.blog/changelog/2026-07-01-enterprise-managed-settings-json-is-generally-available/) ✓;
  at GA it named VS Code and Copilot CLI only; the current docs add the cloud agent).
- **Copilot in GitHub** (cloud agent, code review): the cloud agent is documented to install plugins
  from `.github/copilot/settings.json`; code review isn't. Claude Code cloud sessions don't install
  repo-enabled plugins. See §7b.

### 4.2 Inbound: how app-repo knowledge reaches the team repo

| | Mechanism | Secrets in app repos | Cost | Freshness | Notes |
|---|---|---|---|---|---|
| **I1. App-repo CI generates and notifies** | octo: workflow in each repo writes context, dispatches to hub | A token that can dispatch to the hub | Model call per push to `main` | Minutes | Bot commits in every app repo |
| **I2. Hub scans** | Hub CI, on a schedule or dispatch, reads each repo with a **read-only** token or app and writes `repos/<name>/` | None | One run per changed repo | Hours, or minutes with a webhook | No workflows or secrets in app repos. Hub-side output goes by PR |
| **I3. Sessions propose** | `session-close` in an app repo writes learnings, decisions or status as a PR (or direct push for state) to the team repo | None (uses the engineer's own git) | None extra | At session end | Captures what CI can't see: why, gotchas, decisions. Needs the team repo reachable from the session (local clone or `gh`) |
| **I4. Live query, no copy** | Agent reads other repos when asked: GitHub MCP, `gh`, Sourcegraph MCP ~, DeepWiki-style wiki ~ | None | Per question | Always current | Nothing to keep in sync; slower, and needs tool access per person |

I2 and I3 are complementary: I2 covers what the code says, I3 covers what people learned. octo has only
the first.

### 4.3 Lateral: how an app-repo session understands the other repos

| | Mechanism | Notes |
|---|---|---|
| **L1. A synced map** | Outbound includes a short `team-map.md` (repos, owners, how they connect) plus pointers into the team repo | Cheap context at start; must stay small (memory docs: under 200 lines per file ✓) |
| **L2. A skill that reads the team repo** | e.g. `research` or a new `team` skill reads the team repo by `gh api` or a local clone, on demand | Progressive disclosure: loads detail only when asked |
| **L3. Team repo as an added dir** | `--add-dir <team clone>` | Full read, but per-person setup and the add-dir caveats above |
| **L4. MCP server** | Team repo served as an MCP server (knowledge search, repo map) | Most work to build and host; tool-agnostic |

### 4.4 Where efforts, reviews and personal notes live

Today they live in the brain repo, and the review file lives apart from the code so committing it never
moves HEAD (spec §7.3).

| | Home | Trade-off |
|---|---|---|
| **S1. Still the team repo** | Session writes to a local clone of the team repo (path in `.local/` or user config) | One home, roll-ups unchanged. Needs file access outside the app repo (`additionalDirectories`, works ✓) and a second repo to push |
| **S2. The app repo** | e.g. `.folkways/efforts/<KEY>/` or `docs/efforts/` in the app repo; inbound rolls status up | Lives with the code. Commits status into app repos (other contributors see it); cross-repo efforts have no single home |
| **S3. The tracker** | Jira or GitHub issue is the status; team repo rolls up from the tracker | No new files; depends on the tracker API and on people updating it |
| **S4. Local only** | Gitignored in the app repo; status reaches people through PR descriptions | Zero noise; no handoff to another person or machine |

Reviews have the same choice. In S2 the review file in the app repo would move HEAD when committed, so it
would go in the PR description or stay gitignored instead.

## 5. Composed architectures

Not every combination makes sense. These four are coherent end to end. Each lists what it would change in
folkways.

### A. Hub with CI sync (octo, hardened)

**O2 + I1 or I2 + L1 + S1 or S3.** The team repo runs octo-style workflows; folkways generates them and
fixes the mechanics: PRs instead of direct writes, a folkways-owned folder instead of overwriting
`AGENTS.md`, a GitHub App instead of a shared PAT, SHA-pinned actions, scripts versioned with the hub.

- **Engineer:** nothing to install. Opens any tool in the app repo; reads synced instructions and skills.
- **Team repo:** registry (`folkways.yaml` `repos:`), knowledge, `repos/<name>/`, synthesis, efforts.
- **folkways changes:** `sync` grows a "push to N repos" mode or becomes a GitHub Action; per-repo context
  generators; synthesis prompts. Hooks run only where the CLI is installed, so mores fall back to
  instructions in repos without it.
- **Risks:** a GitHub App with write on every team repo; PR churn in app repos; enforcement is weaker.

### B. Team repo as plugin marketplace

**O4 + I2 + I3 + L1/L2 + S1.** The team repo is both the knowledge store and a marketplace serving one
team plugin. App repos carry only the settings block. Inbound is hub-side scanning plus session-close PRs.

- **Engineer:** opens Claude Code in the app repo, accepts trust once, gets every skill and hook. Team
  knowledge via the map and a skill that reads the team repo.
- **Team repo:** `folkways.yaml`, knowledge, overlays, `repos/<name>/`, the built team plugin, efforts.
- **folkways changes:** a team-plugin build (framework plugin + overlays + team skills) in team-repo CI;
  `sync` writes settings blocks instead of copying skills; hooks ship in the plugin (already done for
  solo); `git-guard` rules re-scoped to the app repo; `session-close` learns to open a team-repo PR.
- **Risks:** Copilot CLI repo-level registration unverified; Claude Code's plugin cache and trust prompts
  are new failure points; no hooks for people who decline the trust dialog.

### C. Package in every app repo

**O3 + I3 + L2 + S1 or S2.** Today's CLI, pointed at app repos: each repo pins `@folkways/cli` and a team
content source (a second package, or a git ref of the team repo) and runs `folkways sync`.

- **Engineer:** `npm ci` in the app repo, as for any dev dependency.
- **folkways changes:** least new code; `sync` learns a team-content source; the brain-repo bits go.
- **Risks:** a Node manifest in non-Node repos; N upgrade PRs per release; team content versioned twice.

### D. Per-machine overlay

**O5 + I3 + L2/L3 + S1.** Nothing in app repos. Each engineer installs the folkways plugin and clones the
team repo; `session-start` recognizes a team repo and loads team context and skills.

- **Engineer:** installs once, works anywhere.
- **folkways changes:** close to today's plugin; drop the brain-repo framing; add repo recognition.
- **Risks:** the app repos carry no standards, so teammates without the plugin, other tools and CI see
  nothing. It's personal tooling more than team tooling.

### Comparison

| | A. CI hub | B. Marketplace | C. Package per repo | D. Per machine |
|---|---|---|---|---|
| Engineer setup | None | Accept trust once | `npm ci` | Install plugin, clone team repo |
| App-repo footprint | Synced folder + workflows | ~10 lines of settings | `package.json`, lockfile, synced files | None |
| Hooks and mores enforced | Only where CLI installed | Yes (Claude Code) | Yes | Yes, for that person |
| Copilot CLI | Instructions yes, hooks partial | Unverified at repo level | Yes (as today) | Yes |
| Other tools (Cursor, Codex) | Instructions | No | Instructions | No |
| Team understands all repos | Strong (I1/I2 + synthesis) | Strong (I2 + I3) | Weak (I3 only) | Weak |
| Security surface | App with write on all repos | Read access for users; read-only scanner | None new | None new |
| Upgrade path | Hub PR fans out | Team repo PR; plugin updates | N repo PRs | Each person updates |
| Reuse of today's code | Medium | High (plugin build exists) | Highest | High |
| Closest to the octo example | Yes | Partly | No | No |

Mixes are possible: B's outbound with A's inbound is a natural pair, as is B with an octo-style
fallback (O1/O2) that writes only a short instructions file for tools that don't take plugins.

## 6. What every option has to settle in folkways

| Today | Question it raises |
|---|---|
| `git-guard` denies checkout, switch, pull, reset, rebase in a main clone | The engineer now works in the main clone by choice. Keep the guard only for subagents (worktrees for parallel work), or drop it? |
| Brain-repo push exemption (spec §7.3) | Which repo is exempt now: the team repo, when a session pushes to it from an app repo? |
| Review file in the brain repo | Where does a review covering HEAD live without moving HEAD (§4.4)? |
| Personal root `people/<handle>/` | Still in the team repo (S1), or gone? |
| `projects_dir` and worktree flow (§8.3) | Still needed for parallel subagents; no longer the main flow |
| Setup offer: "folkways runs from a brain repo" | Becomes "this repo belongs to team X; connect it" or similar |
| `folkways.yaml` `repos:` | Becomes the sync registry (octo's `repo-sync-config.json`) |
| Knowledge audit Action | Gains per-repo context and synthesis, or stays as is |
| Solo mode | Does a solo user have a hub at all, or just the plugin in each repo? |

## 7. Unknowns to test before choosing

| # | Unknown | How to test | Matters for |
|---|---|---|---|
| 1 | Does Copilot CLI auto-register a marketplace from repo settings (`.claude/settings.json` or a `.github` file)? | Fresh clone with the block, run `copilot`, list plugins | B |
| 2 | Do Claude Code plugin hooks get `CLAUDE_PROJECT_DIR` and behave as our repo hooks do? | Already true for the solo plugin; confirm in an app repo | B, D |
| 3 | How intrusive is the trust prompt and the marketplace prompt on first open? | Fresh machine, record the flow | B |
| 4 | Will the beta company's security team accept a GitHub App with PR write on team repos? A read-only scanner? | Ask | A, B (I2) |
| 5 | Cost and quality of per-repo context generation on real repos | Run octo's generator, or a `claude -p` equivalent, on 3 beta repos | A, B |
| 6 | Can the beta team's enterprise policy block repo-level marketplaces (`strictKnownMarketplaces`)? | Check their managed settings | B |

## 7a. Test results (2026-10-16)

Local, offline: a marketplace repo `example-team-agents` (one plugin with a `team-hello` skill and a
`SessionStart` hook that logs `$CLAUDE_PROJECT_DIR`) and an app repo `web-app` whose `.claude/settings.json`
registers it through a `git` source (`file://` URL) and sets `enabledPlugins`. Isolated HOME and config
dirs. Claude Code 2.1.301, Copilot CLI 1.0.98 (npm `@github/copilot`).

| # | Check | Claude Code 2.1.301 | Copilot CLI 1.0.98 |
|---|---|---|---|
| 1 | Marketplace registered from the app repo's `.claude/settings.json` | Yes, at session start, **only once the folder is trusted** (`hasTrustDialogAccepted`). Before trust: ignored, no message (matches docs ✓) | **Yes**, on `copilot plugin marketplace list`, no trust step and no session needed |
| 2 | Plugin from `enabledPlugins` installed automatically | **Yes** (`claude -p` with `CLAUDE_CODE_SYNC_PLUGIN_INSTALL=1`; `init` lists `example-team@example-team`) | **Not observed**: session start needs a Copilot login, and this container blocks `api.githubcopilot.com`. Its changelog says "Support enabledPlugins in config for automatic plugin installation at startup" |
| 3 | Claude-format plugin (`.claude-plugin/plugin.json`, `skills/`, `hooks/hooks.json`) accepted | Yes | **Yes**: `copilot plugin install example-team@example-team` → "Installed 1 skill", hooks copied |
| 4 | Plugin skill usable | Yes: model ran `example-team:team-hello` and replied `EXAMPLE-TEAM-SKILL-OK` | Not run (no login) |
| 5 | Plugin `SessionStart` hook runs with `CLAUDE_PROJECT_DIR` = the app repo | **Yes** (logged `project=…/web-app`) | Not run. Changelog: plugin hooks get `CLAUDE_PROJECT_DIR`, `CLAUDE_PLUGIN_ROOT`, `COPILOT_PLUGIN_ROOT`, `CLAUDE_PLUGIN_DATA` |

Also from the Copilot CLI 1.0.98 changelog (`changelog.json` in its package): it reads `.claude/settings.json`
and `.claude/settings.local.json` as repo config; `.github/copilot/settings.json` is its own repo config;
it auto-loads skills from `.github/skills`, `.claude/skills` and `.agents/skills`; it expands `@` imports in
`AGENTS.md`, `CLAUDE.md` and Copilot instruction files; managed `enabledPlugins` and
`extraKnownMarketplaces` win over local ones.

**Left for a session with a Copilot login:** rows 2, 4 and 5 for Copilot CLI. **Left for full internet:**
what Copilot in GitHub (cloud agent, code review) reads from a repo: instructions, skills, hooks, MCP,
setup steps.

## 7b. What agents running in GitHub read from a repo (session 12, 2026-10-16)

**Source and confidence.** This container blocks docs.github.com too, so the GitHub docs were read from
their source: a sparse clone of [github/docs](https://github.com/github/docs) at `9f6e816` (2026-10-16),
`content/copilot/` and `data/reusables/copilot/`. These are the Markdown files the published pages render
from, so they're marked ✓; paths below are under `content/copilot/` unless noted. Claude Code pages were
fetched from code.claude.com (✓). Nothing here has been tested hands-on yet.

### Copilot cloud agent and Copilot code review

| Repo content | Cloud agent | Code review | Source |
|---|---|---|---|
| `.github/copilot-instructions.md` | ✓ read | ✓ read | `reference/custom-instructions-support.md` |
| `.github/instructions/**/*.instructions.md` (`applyTo` globs; `excludeAgent: "code-review"` or `"cloud-agent"` drops one of the two) | ✓ | ✓ | `data/reusables/copilot/custom-instructions-path.md` |
| `AGENTS.md` (nearest in the tree wins), `CLAUDE.md`, `GEMINI.md` | ✓ all three | ✓ all three, **plus `REVIEW.md`** | `ci-support-agents-all`, `ci-support-ccr-only` reusables. Review guidance: `copilot-instructions.md` for review criteria, root `AGENTS.md` for project context |
| Skills: `.github/skills`, `.claude/skills`, `.agents/skills` | ✓ | ✓ "when relevant"; review-focused names (e.g. `code-review`) are more likely picked. The how-to names only `.github/skills` for review; the concept page lists all three | `concepts/agents/about-agent-skills.md`, `concepts/agent-capabilities/.../code-review.md`, `how-tos/use-copilot-agents/use-code-review.md` |
| Hooks: `.github/hooks/*.json` | ✓ the **only** hook source. Linux only (`bash`/`command`; `powershell` ignored); `ask` is treated as `deny`; `notification` and `permissionRequest` don't fire; `sessionStart` fires once per job | Not documented | `reference/hooks-reference.md` |
| Hooks in `.claude/settings.json` or `.github/copilot/settings.json` | ✗ Copilot CLI only. "Cloud agent does not load `settings.json`" | ✗ | `reference/hooks-reference.md` |
| Plugins: `enabledPlugins` + `extraKnownMarketplaces` in `.github/copilot/settings.json` | ✓ "you install plugins declaratively" there. **Conflict:** the hooks reference says the sandbox ships no installed plugins and doesn't load `settings.json`. Needs a test | Not documented | `concepts/agents/about-plugins.md`, `reference/copilot-cli-reference/cli-config-dir-reference.md` (repository settings) |
| Plugins via `.claude/settings.json` | Not documented (only `.github/copilot/settings.json` is named) | — | — |
| MCP servers | Repository Copilot settings on github.com (JSON in the UI, not a committed file), shared with code review; secrets must be Agents secrets named `COPILOT_MCP_*`; also custom-agent frontmatter | Same settings; GitHub and Playwright MCP on by default | `how-tos/copilot-on-github/customize-copilot/configure-mcp-servers.md` |
| Custom agents: `.github/agents/*.md` (org: `.github` or `.github-private` repo) | ✓ | — | `reference/customization-cheat-sheet.md` |
| `.github/workflows/copilot-setup-steps.yml` | ✓ one job named `copilot-setup-steps`; runs as Actions **before** the agent; only on the default branch; max 59 min; a failing step skips the rest and the agent starts anyway | Reuses it, or `.github/workflows/copilot-code-review.yml` if present | `how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/customize-the-agent-environment.md` |
| Network | Firewall: GitHub and Copilot hosts only by default. **The firewall doesn't apply to setup steps** (or MCP server processes) | Own firewall settings | `how-tos/copilot-on-github/customize-copilot/customize-the-firewall.md` |
| Which ref is read | The working branch | **The PR's head branch** for instructions and skills, so a PR can change them and see the effect | `data/reusables/copilot/code-review/custom-instructions-branch.md` |

Size limits: none stated for instruction files in the current docs. The only number found is the 30,000
character cap on a custom agent's prompt.

### Org and enterprise level

| Mechanism | Reaches | Source |
|---|---|---|
| Organization custom instructions (org settings UI) | Copilot Chat, code review and cloud agent **on github.com only**; not Copilot CLI or IDEs. All sets are sent; precedence for conflicts: personal > repository > organization | `data/reusables/copilot/custom-instructions-org-support.md`, `custom-instructions-chat-precedence.md` |
| Enterprise `managed-settings.json` (`.github-private`) | Cloud agent honors `model`, `enabledPlugins`, `extraKnownMarketplaces`, `strictKnownMarketplaces`. It ignores `permissions.*`, MCP allow and deny lists, `sandbox` and `telemetry` (CLI-only) | `reference/enterprise-administrators/enterprise-managed-settings.md` (supported-keys table) |
| Org or enterprise custom agents | `/agents/*.md` in the org's `.github` / `.github-private` repo | cheat sheet |

### Claude Code outside a local machine

| Surface | Repo files | Repo-enabled plugins | Source |
|---|---|---|---|
| Claude Code cloud sessions (claude.ai/code) | `CLAUDE.md`, `.claude/rules/`, `.claude/skills/`, `agents/`, `commands/`; `.claude/settings.json` hooks and permissions and `.mcp.json` in a one-repo session | **✗ Not installed.** "A cloud session doesn't install the plugins a repository turns on under `enabledPlugins`". Only org **server-managed** settings deliver plugins there | code.claude.com `cloud-environments` (what carries over), `settings` (settings in cloud sessions), `plugins/install` |
| Claude Code GitHub Action | Whatever `actions/checkout` puts on the runner (e.g. `.claude/skills/`) | Installed only through the action's `plugin_marketplaces` and `plugins` inputs. A private marketplace in another repo needs a PAT or app token (the default token reads only its own repo) | code.claude.com `github-actions`, `plugins/org` |

### What this changes in sections 0, 4.1 and 5 (a reading, not a decision)

- **§0 assumed "the cloud agent and code review don't install plugins".** Half wrong: the cloud agent does
  (documented, via `.github/copilot/settings.json`); code review isn't documented to. **Claude Code cloud
  sessions don't** install repo-enabled plugins. So the plugin path covers local Claude Code, Copilot CLI
  and possibly the cloud agent, and misses Claude Code cloud sessions (unless the org uses server-managed
  settings) and Copilot code review.
- **Committed files reach every agent:** instructions (`AGENTS.md` / `CLAUDE.md`, plus
  `.github/copilot-instructions.md` and `REVIEW.md` for review) and skills under `.claude/skills` or
  `.agents/skills` (read by Claude Code, Copilot CLI, the cloud agent and code review). That is the common
  denominator answer 2 asks for.
- **Hooks are per tool:** Claude Code reads `.claude/settings.json`; the Copilot cloud agent reads only
  `.github/hooks/*.json`; Copilot CLI reads both. Full coverage needs both files, each pointing at the same
  binary. A `.github/hooks` file that's also read by Copilot CLI would make hooks run twice there, unless
  the binary detects that.
- **The hook binary in the cloud agent** (the §0 open question): `copilot-setup-steps.yml` runs before
  the agent, outside the firewall, so it can `npm ci` or download a pinned binary. That's a second
  workflow file per repo, and it only takes effect from the default branch. The Claude Code cloud-session
  analogue is the environment's setup script, which is per environment, not per repo.

### Hands-on results (2026-10-16): the cloud agent does more than its docs say

Throwaway private repo `sumsolution/folkways-copilot-probe`, fixture commit `e30a343`. An issue was assigned
to the Copilot cloud agent (it opened PR 2 and wrote its raw outputs to `PROBE-RESULTS.md`), and Copilot
review was requested on a separate PR (PR 3). Each probe wrote a unique marker.

| Probe | Cloud agent | Code review |
|---|---|---|
| `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md` | ✓ all three followed | Not shown (the markers asked for PR-description lines, which a review doesn't write) |
| `REVIEW.md` | ✗ not injected | ✓ followed |
| Skills in `.claude/skills`, `.github/skills`, `.agents/skills`, plus a `code-review` skill | ✓ all listed (self-reported; not invoked) | ✓ the `code-review` skill was used |
| `.github/hooks/*.json`, native camelCase | ✓ sessionStart and preToolUse | — |
| `.github/hooks/*.json`, version 1 with PascalCase events and a `Bash` matcher | ✓ matcher honored | — |
| `.github/hooks/*.json` in **Claude Code format** (no `version`) | ✓ | — |
| Hooks in **`.claude/settings.json`** | **✓, contrary to the docs** | — |
| Hooks in `.github/copilot/settings.json` | **✓, contrary to the docs** | — |
| Plugin enabled in `.github/copilot/settings.json`, from the repo's own **private** marketplace | ✓ installed; its hook ran | — |
| Plugin enabled **only in `.claude/settings.json`** | **✓ installed; its hook ran** | — |
| Plugin skills | ✓ listed | — |
| Plugin hook environment | `CLAUDE_PLUGIN_ROOT`, `COPILOT_PLUGIN_ROOT`, `PLUGIN_ROOT` set to the plugin dir (`/tmp/copilot-plugins/<repo>_<sha>/plugins/<name>`); `CLAUDE_PROJECT_DIR`, `COPILOT_PROJECT_DIR` set to the workspace | — |
| `copilot-setup-steps.yml` | ✓ ran; a tool it put on `PATH` (`$GITHUB_PATH`) was usable by the agent; files it wrote to `/tmp` persisted into the agent's run | Own Actions run ("Running Copilot Code Review") |
| Network | `registry.npmjs.org` and `github.com` both returned 200 from the agent (default firewall) | — |

Caveats: one run each, observed through the agent's own report; skill results are its self-reported list.

**What this changes:** a Claude-format setup (`CLAUDE.md`, `.claude/skills`, `.claude/settings.json`
hooks and `enabledPlugins` from a private marketplace) works **unchanged** in the Copilot cloud agent. Unknowns
7, 8 and 9 below are answered yes; 10 is answered for skills and `REVIEW.md` (hooks in review unobserved).
Copilot CLI double-running hooks (11) is still open. The remaining gap among cloud surfaces is Claude Code
cloud sessions, which skip repo-enabled plugins.

**Security note:** the agent was asked to print its environment with tokens redacted; the redaction missed
`GITHUB_COPILOT_API_TOKEN`, and PR 2 committed it. Lesson for any agent-facing probe or hook: never ask an
agent to dump `env`. Agents in the sandbox can read a live Copilot token.

### Hook delivery: options and Summer's answers (2026-10-16)

What today's hooks do, and what a cloud run needs: `session-start` adds a "follow the `start` skill"
reminder and the contents of `folkways.yaml` and `.local/me.yaml`; `git-guard` checks commits and guards
main clones; `context-band` and `precompact` serve long interactive sessions.

| | Delivery | Reaches | Costs |
|---|---|---|---|
| H1 | npm, as today (`node_modules/.bin`) | Where `npm ci` ran | A Node manifest in non-Node repos |
| H2 | Team plugin carries `hooks.json` and binaries | Local Claude Code, Copilot CLI, maybe the Copilot cloud agent; Claude Code cloud sessions only through server-managed settings | Trust prompt; every platform's binary in each install; no hooks for people who decline |
| H3 | Committed launcher downloads a pinned, checksummed binary | Any surface with network to the host | Cache, offline behavior, Windows script, security review of "downloads a binary" |
| H4 | Install in `copilot-setup-steps.yml` / the Claude environment setup script | Cloud surfaces | A workflow per repo; the Claude script is per environment, not per repo |
| H5 | Commit the binaries | Everywhere | Repo bloat, unreviewable diffs |
| H6 | Rewrite hooks as scripts | Everywhere hooks run | Reverses ADR 0002 |
| H7 | Tiered: hooks locally; instructions and CI in the cloud | Everywhere, at different strengths | Cloud enforcement happens at PR time, not before the command |

Summer's answers:

1. **`git-guard` isn't needed in cloud agents; CI is enough there.** Summer thinks cloud agents can go
   without hooks. `session-start` is the only one that matters, and its path is to be revisited.
2. **A downloaded binary (H3):** possibly acceptable to the beta company's security team, possibly not.
3. **Claude Code server-managed settings:** likely available for the beta org; Summer can confirm.
4. **App repos are not all Node.** folkways targets a wide variety of teams, some with no Node anywhere.
5. **The one-time trust prompt is acceptable** for local hooks (plugin path). A fresh clone gets hooks
   through the repo-enabled plugin after the trust dialog (§7a); a committed launcher (H3) stays a fallback.
6. **Define each hook once** (one source per tool, e.g. the plugin only). Unknown 11 is not tested.

Reading (not a decision): with 1, the cloud question reduces to how `session-start`'s content reaches a
cloud run. Committed instructions plus a skill that reads `folkways.yaml` are one candidate. The `.github/hooks`
file, the double-run question and H3–H5 matter only if a cloud hook is kept. With 4, H1 alone can't be the
local path for every team; H2, or a non-npm install, has to cover non-Node repos.

### New unknowns (add to §7)

| # | Unknown | How to test |
|---|---|---|
| 7 | Does the cloud agent actually install a plugin from `.github/copilot/settings.json`, given the hooks reference's conflicting text? From a **private** marketplace (which token clones it)? Do the plugin's hooks run? | Throwaway repo with Copilot cloud agent enabled; assign an issue |
| 8 | Does the cloud agent read `enabledPlugins` from `.claude/settings.json` too, as Copilot CLI does? | Same repo, block only in `.claude/settings.json` |
| 9 | Does a Claude-format `hooks.json` (PascalCase `PreToolUse`) placed in `.github/hooks/` work in the cloud agent? The reference accepts PascalCase events with snake_case fields | Same repo |
| 10 | Does code review run hooks or plugins at all? | Open a PR in the same repo, request Copilot review |
| 11 | ~~Does Copilot CLI run a hook defined in two sources twice?~~ Dropped: Summer chose to define each hook once | — |

## 7c. The `~` claims, checked (session 13, 2026-10-16)

**Source and confidence.** This container blocks the same sites as session 12 (docs.github.com, github.blog,
dev.to, agentpatterns.ai, and the GitHub web UI and API for issues). So: GitHub docs from a sparse clone of
`github/docs` at `dae83f7` (2026-10-16), paths under `content/copilot/`; Claude Code pages as Markdown from
code.claude.com; projects from their own repos or npm. ✓ = primary source read. Session 14 (local, open
network) opened the pages that were still ~: the github.blog changelogs, the two Claude Code issues (with
`gh`), AgentPatterns, the dev.to posts, the Ruler page and Arvore's story page. Those rows say "session 14".

| Claim (where in 20) | Verdict | What the primary source says |
|---|---|---|
| Copilot CLI installs plugins from marketplaces and reads `.claude-plugin/marketplace.json` (§4.1 O4) | ✓ confirmed | Manifest lookup order: `marketplace.json`, `.plugin/`, `.github/plugin/`, `.claude-plugin/marketplace.json`; plugin manifest also accepts `.claude-plugin/plugin.json` (`reference/copilot-cli-reference/cli-plugin-reference.md`, "File locations") |
| Copilot CLI registers marketplaces from repo settings (§4.1 O4, §7a row 1) | ✓ confirmed | `.github/copilot/settings.json` supports `enabledPlugins` ("declarative plugin auto-install") and `extraKnownMarketplaces`; "The CLI also reads `.claude/settings.json` and `.claude/settings.local.json` for the shared cross-tool subset … `enabledPlugins`, `extraKnownMarketplaces`, and `hooks`" (`cli-config-dir-reference.md`, "Configuration file settings") |
| (new) A repo-enabled plugin is scoped to that repo | ✓ | "auto-installs and activates in the declaring repository, but stays disabled globally"; leaving the repo deactivates it (same page) |
| (new) Trust in Copilot CLI | ✓ partial | Only `model`, `effortLevel`, `contextTier` are documented as waiting for a trusted folder. Plugin auto-update runs "in a trusted working directory". Nothing says `enabledPlugins` waits for trust, which fits §7a (registered with no trust step) |
| (new) Pinning and updates in Copilot CLI (open item 7) | ✓ | A marketplace entry's plugin `source` takes `ref` and `sha` (full 40-char SHA, "immune to force-pushes"); `marketplace add` takes `owner/repo#ref`. Only first-party plugins auto-update by default. **Repo-level `autoUpdate` is "accepted and ignored"**; only user or managed settings can opt a marketplace in. Auto-update is skipped in CI (`cli-plugin-reference.md`) |
| Enterprise `managed-settings.json` in `.github-private` (§4.1 O4, §7b) | ✓ confirmed | Server-managed via `copilot/managed-settings.json` in the enterprise's `.github-private` repo; "applies to all clients, including Copilot cloud agent", only for users licensed by that enterprise; team overrides via `overridable`. MDM and file-based delivery reach local clients only (`how-tos/administer-copilot/manage-for-enterprise/use-managed-settings/deploy-managed-settings.md`). Supported-keys table as in §7b. Managed `extraKnownMarketplaces` sources: `github` (`repo`, `ref`, `path`), `git`, `directory`; a managed `autoUpdate` can't be overridden |
| Managed settings are GA (§4.1 O4) | ✓ confirmed, changelog opened (session 14) | No preview marker on the reference or how-to pages; the MCP registry reusable calls managed settings "the more secure, generally available method". **GA changelog (2026-07-01):** Enterprise Cloud; `copilot/managed-settings.json` in a chosen org's `.github-private`, default branch; the older `.github/copilot/settings.json` path still works. Keys at GA: `extraKnownMarketplaces`, `enabledPlugins`, `strictKnownMarketplaces`, `disableBypassPermissionsMode`, `model`. **Enforced in VS Code and Copilot CLI** at GA, other clients "planned"; the current docs add the cloud agent, so that came later. Fetched at sign-in, refreshed hourly; Business or Enterprise seats from that enterprise only. **Plugins preview changelog (2026-05-06):** the earlier step, Copilot CLI only, from `.github-private/.github/copilot/settings.json`; also enterprise-wide hooks and MCP config |
| `strictKnownMarketplaces` can block repo marketplaces (§7 unknown 6) | ✓ | "Restricts plugin installation to only the marketplaces explicitly defined by the enterprise. An empty array means complete lockdown." Honored by CLI, VS Code, the app, cloud agent, JetBrains. Entry types include `github`, `git`, `hostPattern`, `pathPattern` |
| Org custom instructions: surfaces, precedence, limit (§7b) | ✓ confirmed | Copilot Chat, code review and cloud agent **on github.com only** (so not Copilot CLI or IDEs); personal > repository > organization, all sent. Set in org settings UI by org owners on Business or Enterprise. **No size limit stated** (`custom-instructions-org-support.md`, `add-organization-instructions.md`). GA changelog opened (session 14): GA 2026-04-02 (first shipped April 2025); same three surfaces, Business and Enterprise; it states no size limit and no precedence |
| `--add-dir` loads skills; settings-based dirs don't (§4.1 O6) | ✓ confirmed, and it's **by design** | `--add-dir` / `/add-dir` load `.claude/skills/` (live reload), `.claude/commands/`, `.claude/agents/`, and **only `enabledPlugins` and `extraKnownMarketplaces`** from its settings; `CLAUDE.md` and rules only with `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1`; **its `AGENTS.md` never loads**. "Directories listed in `permissions.additionalDirectories` … grant file access only" (code.claude.com `permissions`, `skills`, `memory`). So issue 30064 describes documented behavior. **Issues opened (session 14, `gh`):** 30064 closed 2026-08-19 as completed; a maintainer reproduced it on 2.1.233 and called it intended, with the docs since updated (`additionalDirectories` = file access only, not a config root). 44224 (`--add-dir` not loading skills, 2.1.89) was flagged as a duplicate of 30064 and closed as stale, not planned, 2026-05-17; the current docs say `--add-dir` does load skills |
| repo-file-sync-action: auth, PR mode, templating (§4.1 O2) | ✓ | Push model: runs in the source repo, opens a PR per target (`SKIP_PR` pushes direct). Auth: a classic PAT with full `repo` scope (`GITHUB_TOKEN` won't work), or a GitHub App installation token with Contents RW, Metadata R, Pull requests RW, Workflows RW to sync workflows. Nunjucks templates per file; `replace: false`; `deleteOrphaned` for directories; labels, reviewers, team reviewers. Last release v1.21.1, 2024-05-05; `node20` runtime (README at `8b92be3`) |
| netlify/file-sync-action (§4.1 O2) | ✓, and it's a **different model** | Pull model: runs in each **target** repo on a schedule, fetches mapped files from a source repo, opens a PR per changed file (branch `file-sync-<sha>`) using the target's own token. Only two inputs (`files`, `token`). `action.yml` still declares `node12`; last non-bot commit 2023-05 (`3f00664`) |
| seylox meta-repo pattern (§4.1 O2 sources) | ✓, and there's a **sequel** | Read both posts from the blog's repo (`seylox/seylox.github.io` at `a07c868`). Below |
| RepoSwarm | ✓ README | Now `reposwarm/reposwarm` (was `loki-bedlam/repo-swarm`), Apache-2.0, Python, at `34dcfaa` (2026-03-26). Runs on Temporal in Docker, clones each repo, picks a type-aware prompt, writes `<repo>.arch.md` into a **central results-hub repo**; re-analyzes only repos with new commits; cache in DynamoDB or files; Anthropic API, Bedrock or LiteLLM. A hub-side scanner (our I2) with heavier infra |
| Arvore Repo Hub | ✓ npm README | `@arvoretech/hub` 0.29.3 (2026-08-29), MIT, source `arvoreeducacao/rhm`. A `hub.yaml` lists repos, MCPs and skills; `hub generate --editor …` writes Cursor, Kiro, Claude Code or `AGENTS.md` config. The hub **is the workspace**: repos are cloned side by side inside it and you open the AI there. That is the brain-repo model, not app-repo-first. **Our-story page (session 14):** 9 repos (NestJS, Elixir, Next.js, React, libraries, infra); engineering cut from 30 to 10 people. The repos sit in the hub folder, excluded by the hub's `.gitignore` and included for the AI by `.cursorignore`, so git sees separate repos and the AI one workspace. Reported gains are self-reported (features "from 3 to 6 months down to weeks"); no limits or downsides stated |
| AgentPatterns central repo (§4.1, Sources) | ✓ opened (session 14) | A canonical `agent-standards` repo fanned out to project repos. Lists five delivery routes: symlinks, `@` imports, generate-then-fan-out by hook or CI (called the most robust cross-tool), packages with semver, and plugins (best for a single vendor). Its example: submodule pinned to a tag, pre-commit copy, CI diff that fails on drift. **When not to:** single-team orgs; conventions still changing weekly ("standards should stabilize before they centralize"); tacit or contested conventions; catalog bloat; no CI diff, so the hub turns advisory. Evidence it cites: 22% of 673 skills across 41 repos fail structural checks; an ETH Zurich study where LLM-generated context files changed task success by −0.5 to −2% (not significant) and raised cost 20–23%. Real teams named (Nx, Datadog) without outcome data. Treats Copilot org instructions, managed settings and marketplaces as complements to the repo, not replacements |
| dev.to "context problem nobody solved" (Sources) | ✓ opened (session 14) | Gabriel Menezes; promotional for his tool **ivar**. Problem: agents bind to one working dir; cross-repo changes break LSP, search, git. Rejects parent dirs, submodule meta-repos, path hacks, MCP bridges, agent swarms. ivar builds a per-session "View Dir" of symlinked worktrees: promoted repos writable on one feature branch, the rest read-only by `chmod`; a `HALL.md` is synthesized into the root `CLAUDE.md`/`AGENTS.md`. Sessions start in the View Dir. No team data |
| team-ai-sync (§4.1 O2) | ✓ opened (session 14) | Fernando Paladini. **Push model by PR**, like repo-file-sync-action: a workflow in the source repo reads a manifest and opens or updates one PR per target; never merges, never bypasses protection. Fine-grained PAT (Contents RW, PRs RW, Metadata R) or a GitHub App, recommended for org-wide use. Per-target file mappings, no templating, `deleteOrphans` off by default. Also GitLab and Bitbucket. Demo only, no team data; notes it was drafted with AI help |
| dev.to Copilot CLI plugins and marketplaces (§4.1 O4) | ✓ opened, **partly outdated** | Marcel Lupo, 2026-05-27. Commands and manifest lookup match the docs (§7c). It says plugins install per user only, with no repo-level auto-enable; the current docs and our §7a test contradict that (`enabledPlugins` in `.github/copilot/settings.json` or `.claude/settings.json`). Its claims that plugin contents are cached at install and that plugin skills don't override project ones are unsourced |
| Ruler and rulesync page (Sources) | ✓ opened (session 14) | Both are **single-repo** generators: one source folder (`.ruler/`, `.rulesync/`) → each tool's native files. Ruler concatenates Markdown into one file per agent, 30+ agent ids, also MCP, skills, subagents; it gitignores its output by default, so clones, cloud agents and reviewers miss it unless `--no-gitignore`. rulesync writes native scoped files (`.claude/rules/`, `.cursor/rules/`), also hooks and permissions, and has `generate --check` for CI; it shipped four majors in three days (Sept 2026), so pin it. Neither syncs across repos |

**seylox, in more detail** (because it's the closest real-team account, and the October post answers §8 Q1's
"evidence on the concerns" in part):

- Team of five, a dozen repos, two product lines, Claude Code. A no-code **meta-repo cloned on every
  machine**, beside the code repos in a workspace folder. A bootstrap script writes a generated, uncommitted
  workspace-root `AGENTS.md` of `@` includes, chosen by which product lines are on disk. So it's a
  per-machine overlay (our D/L3) plus per-repo `AGENTS.md` files that keep build and test detail.
- Layers: router `AGENTS.md`, `workspace.yml` registry, `repos.yaml` per product line, org rules always
  loaded, per-domain conventions and workflows loaded by routing (ticket prefix, then keywords, else ask).
- Workflows are plain markdown with trigger phrases, not Claude Code skills, "because the repo is for every
  agent, not one tool". Active-work trackers are "two thirds of all commits"; the archive feeds conventions.
- Costs they report: about 64,000 lines of markdown, three quarters archived; prose overviews drift; a cited
  write-up warns an unmaintained meta-repo "becomes a source of confidently wrong behavior at agent speed".
  No atomic cross-repo commit.
- Their claim on value: the shared clone (many people, one agent memory), not the map itself. Two other
  groups in their company copied the pattern; one stalled after a day.
- Onboarding: a new hire "cloned one repository, ran the script, started the agent".
- Not addressed: cloud agents, Copilot, outside contributors. Every session needs the meta-repo clone, which
  is the gap Summer's concerns 2–4 (§0) point at.
- New sources it cites: the [meta-repo pattern write-up](https://devnewsletter.com/p/meta-repo-pattern/),
  [codified context](https://arxiv.org/abs/2602.20478) (hot vs cold memory), and
  [woditschka/agentic-coding-reference](https://github.com/woditschka/agentic-coding-reference) with the
  SoCraTes Austria session "Agents Forget. Repos Remember." Read in session 14; see §7d.

**What this changes (a reading, not a decision):**

- Copilot CLI's repo-level plugin path is now documented, not just observed: same `.claude/settings.json`
  block as Claude Code, scoped to the repo, no documented trust gate. §5 B's risk "Copilot CLI repo-level
  registration unverified" and the comparison row "Unverified at repo level" are answered for registration
  and install; skill and hook runs with a login (open item 1) are still untested.
- Updates differ by tool: a repo can't turn on Copilot CLI auto-update for a team marketplace. Freshness
  there comes from user settings, managed settings, or pinning a `ref`/`sha` that a PR bumps.
- `strictKnownMarketplaces` is a real blocker for B in an enterprise that sets it; the team marketplace would
  need to be on the enterprise allowlist (unknown 6 becomes a question for the beta admin, not a test).
- The prior art splits three ways: hub writes into repos (octo, repo-file-sync-action, team-ai-sync), repos
  pull from a hub (netlify's action), and the hub as workspace or overlay (Arvore, seylox, ivar). RepoSwarm
  is inbound only. Ruler and rulesync are a fourth kind, inside one repo: one source, every tool's file.
- (Session 14) Managed settings reached the cloud agent after GA, not at it; anything built on that path
  depends on a recent behavior. AgentPatterns' "when not to centralize" list (single team, conventions still
  changing weekly, no CI drift check) is the closest thing found to a published counter-argument.

## 7d. The sources the seylox sequel cites (session 14, 2026-10-16)

**Source and confidence.** Run locally with an open network. All pages opened (✓):
[The Meta-Repo Pattern](https://devnewsletter.com/p/meta-repo-pattern/) (The Dev Newsletter, 2026-03-30, no
named author); [Codified Context](https://arxiv.org/abs/2602.20478) (Aristidis Vasilopoulos, arXiv v1,
2026-02-24), abstract and [full HTML](https://arxiv.org/html/2602.20478v1);
[woditschka/agentic-coding-reference](https://github.com/woditschka/agentic-coding-reference) at `4ccc5fc`
(README, adoption guide, `product-workspace.md`, `human-teams.md`, `cross-tool-strategy.md`, the channel and
workspace ADRs) and the [SoCraTes session page](https://github.com/Softwerkskammer-Linz/SoCraTesAT-2026/tree/main/agents-forget-repos-remember)
(talk by Bernhard Woditschka, SoCraTes 2026 Linz; a short README with links to the same repo).

### What each one is, side by side

| | Meta-repo write-up | Codified context | agentic-coding-reference ("agent-team") |
|---|---|---|---|
| Kind | Survey article, four approaches with named practitioners | Single-author experience report with numbers | An open-source harness: skills, agents, hooks, engines, eval bench |
| Scale | Anyline: 6 SDK repos. Owen Zanzal: **35 repos** cloned into one workspace with a root `CLAUDE.md`. Mars CLI (Dean Sharon): declarative repo list, tags, `mars exec`. Nx: synthetic monorepo graph | **One developer, one repo**: 108k lines of C#, 283 sessions over 70 days part time | Two runnable samples (Go, Spring Boot), a generic one, and a four-repo `product-workspace` sample (partly built) |
| Where sessions start | The meta-repo or workspace root in every case | The one repo | **Single repo:** the app repo itself. **Several repos:** the "umbrella" repo only; members carry a one-line `CLAUDE.md` pointing to it, and "no harness session starts here" |
| How content reaches app repos | It doesn't, by design: the meta-repo sits above them and carries no app code | n/a (one repo) | Three channels, **copy** the default: runtime committed into each app repo by `/materialize`; **manifest** (same, gitignored, from a pinned source); **marketplace**: per-tool plugins from one `.claude-plugin/marketplace.json`, enabled per repo in committed `.claude/settings.json`, plus a small engine sliver a setup skill installs and gitignores |
| Cloud agents | Not covered | Not covered | Not covered as a target. Mentions Copilot's `&` delegation to the cloud agent and that `.github/instructions` reach the cloud agent and code review. Copy channel commits skills and `CLAUDE.md`, which §7b says those agents read |
| Tools | Claude Code, Codex, Cursor, Copilot, Gemini, OpenCode, Nx | Claude Code + MCP | Claude Code, Copilot CLI, OpenCode (Junie retired). One `CLAUDE.md` and one `.claude/skills/` tree serve all three |
| Costs stated | Keeping manifests and docs current; disk; context noise; prompt injection through retrieved docs; no atomic cross-repo commits | About 26k lines of context infra (24% of code lines); about 5 min per session when a spec changes plus a 30–45 min biweekly review (1–2 h a week); two incidents of stale specs causing conflicting code | Model spend: $3.90 for one recorded bug fix, about $5 per feature on its public eval series (README), about $9 in the talk; every upgrade replaces the runtime whole; on the marketplace channel, a setup re-run after each plugin update or engines skew |

### Points worth keeping

**Meta-repo write-up**

- **Visibility gap vs transaction gap:** a meta-repo tells the agent what else should change; it doesn't
  give atomic commits, shared releases or integrated tests. Those need their own process.
- **The unmaintained meta-repo warning** (the source of the seylox quote): a stale root `CLAUDE.md`
  naming renamed services, helper scripts that encode one person's habits, system maps that drift the day
  a new service ships. Agents treat the docs as ground truth, so drift hurts faster than with people.
- Recommends measuring outcomes (time to first useful change, cross-repo task completion, review load,
  rollback rate) and trust separately from throughput.
- No adoption criteria; its bottom line is that a meta-repo beats agents working blind.

**Codified context**

- Three tiers: a 660-line **constitution** always loaded (conventions, build commands, known failure
  modes, a trigger table routing tasks to agents); **19 specialist agents** (over half of each is domain
  knowledge, not behavior); **34 specs** (about 16k lines) fetched on demand through a small MCP server with
  keyword search (1,478 retrieval calls across 218 sessions).
- The hot/cold split is the paper's main idea: keep the always-loaded file short and route to detail.
  That matches L1 (small map) + L2 (skill that reads more) in §4.3 and the memory docs' size guidance.
- Limits it names: single developer, single project, observational only, no failure rates, tied to Claude
  Code and MCP. Teams and multi-repo are future work.

**agentic-coding-reference**

- **Its multi-repo design is the brain-repo model.** The umbrella holds the PRD, cross-module decisions,
  glossary and member list (`[workspace.members]` in `scripts/layout.toml`, with contract paths and
  dependencies); members sit as sibling clones and are checked out per slice. It considered and rejected
  submodules (reviews go blind to a pointer diff), subtree (history and sync burden) and nested checkouts
  (gitlinks, hidden from search, the member's rules file loads too).
- **Two tiers of truth with a routing rule:** whatever another team would need to know goes up to the
  umbrella; a member keeps only its own design, ADRs and any deviation from the shared briefs.
- **Partial checkout is designed in:** an absent member is reported, not failed; a contract change with an
  absent dependent blocks, other changes warn.
- **Teams:** one checkout and one slice per person; claims live in the team's tracker, reached by a
  project-owned `backlog.sh`; the harness adds no shared registry or lock. Shared memory is `docs/` through
  git; working memory (`.scratch/`) is per checkout and gitignored.
- **Distribution is close to our O3/O4 split:** copy (committed, like `folkways sync`) as the default after
  real adoptions showed it was what people wanted; marketplace for teams that prefer a plugin pinned in
  `.claude/settings.json`. Its marketplace ADR keeps engines project-side because plugin caches are
  tool-specific paths.

### Against Summer's four concerns (§0 row 1)

| Concern | Meta-repo write-up | Codified context | agentic-coding-reference |
|---|---|---|---|
| 1. Unconventional workflow | Every approach starts sessions in a workspace above the repos | n/a | Single repo: no, sessions start in the app repo. Multi-repo: yes, umbrella only |
| 2. Cloud environments can't reach several repos | Not discussed | n/a | Not discussed; the umbrella needs sibling clones and per-tool directory grants (Copilot's grant is per user, not committed) |
| 3. Outside contributors clone only the app repo | Not discussed | n/a | Copy channel: they get the full runtime in the clone. Workspace member: they get one line pointing elsewhere |
| 4. Copilot agents in GitHub lack context | Not discussed | n/a | Copy channel commits `CLAUDE.md` and `.claude/skills`, which those agents read (§7b); not tested by the project |

### What this changes (a reading, not a decision)

- None of the three sources solves the cross-repo case without a workspace or umbrella where sessions
  start. The seylox, Arvore, Mars, Zanzal and agent-team umbrella designs all land there. The app-repo-first
  premise has prior art only in the single-repo mode (agent-team's copy channel, Ruler/rulesync, octo).
- agent-team is the nearest relative to folkways found so far: skills, hooks, a doctor, evals, overlays
  ("briefs" with a kernel that no brief can change), npm-free install by copy or plugin, Copilot CLI
  parity. Its choices (copy as default, engines outside the plugin, per-person checkouts with the tracker
  as the coordination layer) are data points for §4.1 and §4.4, not a template.
- The cost evidence is thin and all from single teams or one person: 1–2 hours a week to keep one developer's
  context current (codified context), the seylox team's 64k lines, the meta-repo warning. No source reports
  the cost of fanning content out to many app repos.

## 7e. Wider prior art (session 14, 2026-10-16)

**Source and confidence.** Web search, then each page opened with WebFetch (✓); GitHub repos and issues read
with `gh`; Claude Code docs fetched as Markdown from code.claude.com; GitHub docs from docs.github.com. Bug
reports are what users reported at the version they name, not retested here.

### Other hubs, and the octo upstream

- **`TheJeans/octo-team-context`**: one commit (`2e9869c`, 2026-05-30), "Central information hub for the
  ACME engineering team", 7 forks all made the same day and never pushed since. It reads as a workshop or
  talk template, not a running team's repo. The "From Drift to Sync" talk wasn't found; the nearest match
  is Colleen Lake, "Agentic Drift: keeping pace with your agents", WeAreDevelopers World Congress North
  America 2026 (page ✓, description only: from manual reviews to "systemic governance"; video behind a login).
- **Sync tools found by GitHub search** (all ✓ README):

| Tool | How it runs | What it writes | Notes for us |
|---|---|---|---|
| [trick77/agents-md-sync](https://github.com/trick77/agents-md-sync) (npm) | Runs on one machine over local clones; one commit per target on a tool-owned branch, force-pushed, optional PR | A generated `AGENTS.md` from central fragments; each repo's `.agents/<NAME>.md` is a repo-owned addendum placed **before** the matching central section | The clearest ownership split found: tool owns `AGENTS.md`, repo owns `.agents/`. Preview by default |
| [yuka-repos/agents-md-sync](https://github.com/yuka-repos/agents-md-sync) (PowerShell) | Local script over a workspace | Only a **marked block** inside each `AGENTS.md` pointing to workspace-wide docs by absolute path; text outside the markers untouched; a check mode fails on drift | Refuses to write into public repos (absolute paths would leak). The pointer only works on machines with that workspace |
| [lbb00/ai-rules-sync](https://github.com/lbb00/ai-rules-sync) (npm) | Per machine: **symlinks** native assets (rules, skills, agents, `AGENTS.md`) from git repos into each project | Symlinks, no copies | Instant upstream changes, but nothing lands in the app repo's history, so clones and cloud agents see nothing (fails §0 answer 2) |

### Real-team experience

Searches for retros on central agent repos or org-wide plugin rollouts found **no first-hand account with
outcomes**. What exists:

| Source | Who | What they did | What they hit |
|---|---|---|---|
| [Lessons from building Claude Code: how we use skills](https://claude.dev/blog/lessons-from-building-claude-code-how-we-use-skills/) ✓ | Thariq Shihipar, Anthropic, 2026-06-03 | Checked-in `.claude/skills` for small teams and few repos; an **internal plugin marketplace** as the org scales. Skills start in a sandbox folder, move into the marketplace by the owner's PR once they get traction; no central gatekeeper; a `PreToolUse` hook logs skill usage | Every checked-in skill costs context, which is why they moved to a marketplace. "Hundreds" of skills; no adoption numbers |
| [Steering AI agents in monorepos with AGENTS.md](https://dev.to/datadog-frontend-dev/steering-ai-agents-in-monorepos-with-agentsmd-13g0) ✓ | Simon Boudrias, Datadog, 2025-09-26 | Root `AGENTS.md` as a router to folder-level docs; a dev-experience team owns the scaffolding, product teams own their content; one-line `CLAUDE.md` that imports `AGENTS.md`; gitignored `AGENTS.local.md` | Nested files aren't found unless you work in the right folder; keep docs short; test with a shared set of example prompts. Monorepo only |
| [Structuring Claude Code for multi-repo workspaces](https://karun.me/blog/2026/03/26/structuring-claude-code-for-multi-repo-workspaces/) ✓ | Karun Japhet, 2026-03 (updated 2026-10-15) | **Sessions start in the app repo**; org and team `CLAUDE.md` come from a bootstrap repo that is the **parent folder** of the clones (`mani` clones them), so Claude Code's walk-up reads them | Repo lists go stale (point at the manifest instead); gitignore `dir/` vs `dir/*` breaks exceptions; must work with partial checkouts. Works only on machines with the bootstrap layout: cloud agents and outside contributors get the repo level only |
| [My personal skills repo became internal tooling](https://www.joekarlsson.com/blog/my-personal-claude-code-skills-repo-accidentally-became-internal-tooling/) ✓ | Joe Karlsson, CloudQuery marketing, 2026-04-07 | One repo, cloned; an `/update` skill wraps `git pull`; a `/setup` wizard | Terminal was the main adoption barrier; a model update broke one skill; docs rot; it needs a named owner with time; "feel free to contribute" brought nothing, specific asks did |
| [Beyond AGENTS.md](https://mem0.ai/blog/beyond-agents.md-shared-memory-for-coding-agents-across-services-and-repos) ✓ | Mem0 (vendor), 2026 | Shared memory service across repos | One staged two-repo demo: a static `AGENTS.md` can't tell service B about service A's recent change. No team data |
| [Polyrepo synthesis](https://rajiv.com/blog/2025/11/30/polyrepo-synthesis-synthesis-coding-across-multiple-repositories-with-claude-code-in-visual-studio-code/) ✓ | Rajiv Pant, solo, 2025-11 | Same repo-ecosystem table in each repo's `CLAUDE.md`; VS Code multi-root | Git commands in the wrong repo; logic duplicated into whichever repo was open; stale docs unless updated in the same commit |

### Claude Code team plugins in practice

From the current docs (code.claude.com `plugins/install`, `plugins/loading`, `plugins/org`,
`plugins/host-marketplace`, ✓):

- **Auto-update is off by default for a team marketplace** (on only for official ones and those added from
  claude.ai). Precedence: `autoUpdate` on the `extraKnownMarketplaces` entry in a settings file, then the
  user's `/plugin` toggle, then the default. So, unlike Copilot CLI (§7c), an app repo's committed
  `.claude/settings.json` **can** turn auto-update on.
- When it runs: after the first message, a random delay of up to ten minutes, then a background refresh.
  The running session keeps what it loaded; the next session gets the new version. Old versions are kept 14
  days so running sessions keep working.
- **Private marketplace updates** use the user's git credential helper without prompting: SSH agent keys and
  stored HTTPS credentials work; a helper that needs to prompt fails quietly and the old checkout stays.
- Pinning: `owner/repo#ref` on add, `ref`/`sha` on plugin sources. `strictKnownMarketplaces` entries must
  match `ref` and `path` exactly (an entry without `ref` doesn't cover `ref: "main"`).
- `-p` and CI: `extraKnownMarketplaces` only in folders already trusted; `CLAUDE_CODE_PLUGIN_SEED_DIR` for
  images that can't clone at runtime (seed marketplaces never auto-update).

Reported bugs in `anthropics/claude-code` (`gh`, ✓ read; status as of 2026-10-16):

| Issue | Version or date | Reported | State |
|---|---|---|---|
| [#14061](https://github.com/anthropics/claude-code/issues/14061) | 2025-12 | `/plugin update` refreshes the marketplace but not the plugin cache | Open (labelled duplicate) |
| [#52218](https://github.com/anthropics/claude-code/issues/52218) | 2026-04 | `autoUpdate` updates skills but leaves **bundled hooks** on the old install path | Closed, stale |
| [#55355](https://github.com/anthropics/claude-code/issues/55355) | 2.1.126 | **Project-scope** installs: wrong entry picked across projects, no auto-update | Closed, not planned |
| [#67868](https://github.com/anthropics/claude-code/issues/67868) | 2.1.175 | Third-party marketplace with `autoUpdate: true` not refreshed across restarts | Closed, not planned |
| [#73952](https://github.com/anthropics/claude-code/issues/73952) → [#90135](https://github.com/anthropics/claude-code/issues/90135) | 2026-07/08 | A mid-session plugin update removes the old path and **silently stops hooks** in running sessions | Re-filed, open. The 14-day retention in the docs may postdate it; not retested |
| [#14485](https://github.com/anthropics/claude-code/issues/14485) | 2025-12 | Marketplace clones over SSH trigger SSH-agent approval prompts at startup | Open |
| [#28729](https://github.com/anthropics/claude-code/issues/28729) | 2026-02 | Feature request: a git repo as the source of org skills (instead of uploading in the admin console) | Open |

The pattern across them: skills update reliably, **plugin hooks lag or break on update**. That matters most
for the hook-in-plugin path (H2 in §7b).

### GitHub App vs PAT for a hub with PR write

GitHub docs (✓ [deciding when to build a GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/deciding-when-to-build-a-github-app),
[managing PATs](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens),
[App best practices](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/best-practices-for-creating-a-github-app))
and the `actions/create-github-app-token` README (v3.3.0, ✓):

| | GitHub App | Fine-grained PAT | Classic PAT |
|---|---|---|---|
| Identity | The app (bot) | A user | A user |
| Repo scope | Repos chosen at install; each token can narrow further (`repositories:` input) | Chosen repos of **one** owner | Everything the user can reach |
| Permissions | Fine-grained, narrowable per token (`permission-<name>:`) | Fine-grained | Broad scopes |
| Lifetime | Installation token: **1 hour** | Up to 366 days or none, unless an org caps it | Until revoked; removed after a year unused |
| When the creator leaves | Keeps working | Breaks | Breaks |
| Org controls | Owner installs and approves new permissions | Org can require approval; can block them | Org can block classic PATs |
| GitHub's advice | "For long-lived integrations, you should use a GitHub App" | — | "Less secure" |

- The long-lived secret becomes the app's **private key**, which can act on every installation. GitHub
  advises a key vault with sign-only use; an environment variable is weaker.
- Writing `.github/workflows/*` (for example `copilot-setup-steps.yml`) needs the **Workflows** permission
  on top of Contents and Pull requests, for apps and PATs alike (repo-file-sync-action, §7c).
- What security teams usually accept: **no evidence found**. Still a question for the beta company (§7
  unknown 4).

### Per-repo context generation: tools and costs

| Tool | What it produces | Cost signal |
|---|---|---|
| octo (`generate-repo-context.py` at `2e9869c`, read) | `STACK.md`, `DEV.md` from a 10-step tool loop (file reads cut at 6,000 chars, 2,000 output tokens per step); `DEPS.md` deterministic | Runs on **GitHub Models** with the workflow's `GH_TOKEN`, default `gpt-4.1`, so the bill would land on the org's GitHub Models allowance, not an API key (GitHub Models pricing not checked). Up to 11 calls per push to `main` |
| RepoSwarm (§7c) | `<repo>.arch.md` per repo in a hub | Temporal + Docker + DynamoDB to run; re-analyzes only changed repos. No cost figures published |
| [Litho / deepwiki-rs](https://github.com/sopaco/deepwiki-rs) (✓ README, 3.1k stars) | C4 architecture docs from code, CI-friendly; succeeded by "Terrain", a "living map" for agents | None stated |
| [DeepWiki MCP](https://docs.devin.ai/work-with-devin/deepwiki-mcp) (✓) | Wiki and Q&A per repo over MCP (`read_wiki_structure`, `read_wiki_contents`, `ask_question`) | Free for **public** repos; private repos need a Devin account and API key, pricing not on the page |
| `claude -p` with a fixed prompt | Whatever the prompt asks | No published figure. Search found only generic estimates (one blog: ~900k input + 120k output tokens for a full docs pass); not a basis for a number |

**Cost stays unknown 5:** a trial on three beta repos is still the only way to get a real figure.

### What this changes (a reading, not a decision)

- The two managed-file designs found in the wild match what §3 said octo lacks: a tool-owned file plus a
  repo-owned addendum folder (trick77), or a marked block inside a hand-written file (yuka-repos).
  Both avoid octo's wholesale overwrite.
- Karun's walk-up layout is the only found setup where sessions start in the app repo **and** pick up team
  context without copying anything. It depends on the parent folder, so it fails the same concerns (2–4) as
  seylox.
- Anthropic's own advice (check skills in for small setups, a marketplace when it grows) lines up with the
  B-as-add-on reading in §0. The bug list says plugin **hooks** are the fragile part of that path, while
  plugin skills are not.
- Every real-team account names an **owner with time** or a named team (Datadog's dev-experience team,
  Karlsson's "AI champion"); none reports a hub running unattended.

## 8. Questions for Summer (answered 2026-10-16, see §0)

1. **The concerns you heard about the brain repo:** which ones? The answer weights the options. For
   example, "a second place to start sessions" rules out D's clone; "nothing new in my repo" favors B or
   D; "the agent writes to my repo without review" rules out O1.
2. **Must app repos carry standards for people and tools without folkways** (other teams' contributors,
   Cursor, CI)? Yes points to A or an A/B mix.
3. **Is Copilot CLI parity a hard v1 requirement** if it means waiting on unknown 1?
4. **Where should efforts and reviews live** now (§4.4)? This one changes the most code.
5. **Who owns writes into app repos:** a bot (PRs from the hub) or only people (sessions)?
6. **Does the octo scope stay in** (per-repo context generation and architecture synthesis), or is that a
   later phase after the outbound side works?

## Sources

Primary, fetched (✓):

- Claude Code: [Manage plugins for your organization](https://code.claude.com/docs/en/plugins/org)
  (repo-level `extraKnownMarketplaces`/`enabledPlugins`, trust, managed allowlists, seeds, auto-update)
- Claude Code: [Create a marketplace](https://code.claude.com/docs/en/plugin-marketplaces) (sources,
  relative paths, validation)
- Claude Code: [Memory](https://code.claude.com/docs/en/memory) (imports, external-import approval,
  `--add-dir` and `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD`, symlinked rules, size guidance)
- [octo-team-context](https://github.com/summer-jordan/octo-team-context) at `2e9869c` (read in full)

Checked in session 13 (§7c): the Copilot docs links below, both sync actions, seylox, RepoSwarm and
Arvore are now ✓ from their source repos or npm. Opened in session 14 (§7c): the changelog posts, the two
Claude Code issues, AgentPatterns, the three dev.to posts, the Ruler page and Arvore's story page, so every
link in the next list is ✓ except the RepoSwarm write-up on robotpaper.ai (its README is ✓), Backstage,
Sourcegraph and DeepWiki, which no session has opened. The §7d sources are listed there.

From search summaries, now opened unless noted:

- Copilot: [enterprise managed-settings.json GA](https://github.blog/changelog/2026-07-01-enterprise-managed-settings-json-is-generally-available/),
  [enterprise-managed plugins preview](https://github.blog/changelog/2026-05-06-enterprise-managed-plugins-in-github-copilot-cli-are-now-in-public-preview/),
  [CLI plugin reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-plugin-reference),
  [Copilot CLI plugins and marketplaces](https://dev.to/pwd9000/github-copilot-cli-plugins-and-marketplaces-extend-your-terminal-agent-16pc),
  [organization custom instructions GA](https://github.blog/changelog/2026-04-02-copilot-organization-custom-instructions-are-generally-available/)
- Claude Code `--add-dir` skills: [skills docs](https://code.claude.com/docs/en/skills),
  [issue 30064](https://github.com/anthropics/claude-code/issues/30064),
  [issue 44224](https://github.com/anthropics/claude-code/issues/44224)
- File sync across repos: [repo-file-sync-action](https://github.com/BetaHuhn/repo-file-sync-action),
  [netlify/file-sync-action](https://github.com/netlify/file-sync-action),
  [team-ai-sync](https://dev.to/paladini/sync-ai-agent-rules-across-repositories-with-github-actions-5b9h)
- Rule generators: [Ruler and rulesync](https://developertoolkit.ai/en/ecosystem/frameworks/rules-sync/)
- Central-repo and meta-repo patterns: [AgentPatterns central repo](https://agentpatterns.ai/workflows/central-repo-shared-agent-standards/)
  (drift; overhead when conventions change weekly),
  [agents meta-repo](https://seylox.github.io/2026/03/05/blog-agents-meta-repo-pattern.html) (`repos.yaml`
  map, per-repo `AGENTS.md` keeps local detail),
  [RepoSwarm](https://robotpaper.ai/reposwarm-give-ai-agents-context-across-all-your-repos/),
  [multi-repo context problem](https://dev.to/mnzs/the-context-problem-nobody-solved-ai-agents-across-multiple-repos-4533),
  [Arvore Repo Hub](https://hub.arvore.com.br/docs/our-story)
- Repo metadata aggregated centrally: Backstage `catalog-info.yaml` discovery
  ([Gitea provider](https://backstage.io/docs/next/integrations/gitea/discovery),
  [Harness GitHub discovery](https://developer.harness.io/docs/internal-developer-portal/plugins/available-plugins/github-catalog-discovery))
- Live cross-repo query: [Sourcegraph MCP](https://sourcegraph.com/mcp),
  [DeepWiki MCP](https://docs.devin.ai/work-with-devin/deepwiki-mcp)
