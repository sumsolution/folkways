# Engineers work in app repos; the team repo syncs: options

**Researched:** 2026-10-16 (session 11). Main agent, no subagents. Tool facts come from primary docs where
fetched (marked ✓); the rest come from search summaries (marked ~) because this container's network
blocks most sites. **Status: options for Summer to review. Nothing decided.** Summer's answers to §8 and
the marketplace tests were added the same day (§0, §7a).

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
| **O2. Hub opens PRs** | Hub CI opens a PR per repo when content changes ([repo-file-sync-action](https://github.com/BetaHuhn/repo-file-sync-action) ~, [team-ai-sync](https://dev.to/paladini/sync-ai-agent-rules-across-repositories-with-github-actions-5b9h) ~) | Synced files in a folkways-owned folder | App-repo owners merge | Any | Needs a GitHub App or token with PR write on team repos. Upgrades visible in each repo, as `sync` does today |
| **O3. App repo pins and pulls** | Each app repo has `@folkways/cli` pinned and runs `folkways sync`, pulling team content too | `package.json` + lockfile + synced files | Upgrade PR per repo | Any | Reuses today's CLI. Odd in Go, Java or Python repos (a Node manifest in each). N repos to bump |
| **O4. Team repo is a plugin marketplace** | Team repo carries `.claude-plugin/marketplace.json`; its team plugin = framework + compiled overlays + team skills + hooks. App repo's `.claude/settings.json` adds `extraKnownMarketplaces` and `enabledPlugins` ✓ | About 10 lines of settings | Team repo PRs; plugin auto-updates or pins a ref ✓ | Claude Code ✓; Copilot CLI partly (below) | Nothing else copied into app repos. Teammates are prompted to trust the folder and the marketplace ✓ |
| **O5. Per machine, no app-repo footprint** | Each engineer installs the plugin; `session-start` checks whether the current repo is listed in the team's `folkways.yaml` and loads team context | None | Team repo PRs | Claude Code, Copilot CLI | Only folkways users get anything; the repo itself carries no standards for other tools or newcomers |
| **O6. Import or add-dir** | App repo `AGENTS.md` imports from a local clone of the team repo, or the engineer runs with `--add-dir <team repo>` | One import line, or none | Team repo PRs | Claude Code ✓; Copilot CLI `--add-dir` ✓ ([05](05-research-repo-split.md)) | Import paths are machine-specific and trigger an approval dialog ✓. `--add-dir` loads skills but not `CLAUDE.md` unless an env var is set ✓, and settings-based dirs don't load skills ~ |

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
  `copilot/managed-settings.json` ~
  ([changelog](https://github.blog/changelog/2026-07-01-enterprise-managed-settings-json-is-generally-available/)).
- **Copilot in GitHub** (cloud agent, code review): no evidence yet that it installs plugins. Assume it
  reads only committed files until verified (§0).

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

From search summaries (~, not opened; verify before relying on them):

- Copilot: [enterprise managed-settings.json GA](https://github.blog/changelog/2026-07-01-enterprise-managed-settings-json-is-generally-available/),
  [enterprise-managed plugins preview](https://github.blog/changelog/2026-05-06-enterprise-managed-plugins-in-github-copilot-cli-are-now-in-public-preview/),
  [CLI plugin reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-plugin-reference),
  [Copilot CLI plugins and marketplaces](https://dev.to/pwd9000/github-copilot-cli-plugins-and-marketplaces-extend-your-terminal-agent-16pc),
  [organization custom instructions GA](https://github.blog/changelog/2026-04-02-copilot-organization-custom-instructions-are-generally-available/)
- Claude Code `--add-dir` skills: [skills docs](https://code.claude.com/docs/en/skills),
  [issue 30064](https://claudeissues.com/issue/30064-bug-permissions-additionaldirectories-does-not-load-skills-from-added-directorie),
  [issue 44224](https://claudeissues.com/issue/44224-bug-add-dir-cli-flag-does-not-load-skills-from-added-directorys-claude-skills-v2)
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
