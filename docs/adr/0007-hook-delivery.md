# 0007. How do hooks reach agents in an app repo?

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-17 |
| Owner | Summer Jordan |

## Issue

Engineers are to start sessions in their app repos, not in a brain repo, and Copilot running in GitHub
(the cloud agent, code review) is now a first-class target ([research §0][r-0]). Today hooks reach a repo
one way: `folkways sync` commits `.claude/settings.json` entries that run
`node_modules/.bin/folkways-hook` ([ADR 0002](0002-hook-runtime.md)). Many app repos have no Node, and
agents in GitHub run on a fresh clone. Each tool also reads hooks from more than one file, so one hook can
end up defined twice. How do hooks reach agents in an app repo, and from how many places?

## Context

- **What the hooks do** ([research §7b, hook delivery][r-hooks]): `session-start` adds the `start`
  reminder and the contents of `folkways.yaml` and `.local/me.yaml`; `git-guard` checks commits and
  guards main clones; `context-band` and `precompact` serve long interactive sessions.
- **The plugin already carries the hooks.** `packages/claude-plugin` builds `hooks/hooks.json` from the
  same `settings.json`, with `folkways-hook` for darwin and linux, arm64 and x64, run by `hooks/run.sh`
  ([README][p-readme], [run.sh][p-run]). Today `run.sh` exits early in a synced repo so the committed
  settings run the hooks. Without that check, a synced repo would run each hook twice.
- **A repo can turn the plugin on.** `extraKnownMarketplaces` and `enabledPlugins` in the app repo's
  `.claude/settings.json` register and install it. Tested ([research §7a][r-7a]): in Claude Code 2.1.301,
  after the folder is trusted, the plugin's `SessionStart` hook ran with `CLAUDE_PROJECT_DIR` set to the app
  repo. Copilot CLI 1.0.98 registered the marketplace and installed the plugin with its hooks. With no
  Copilot login, its auto-install at startup, skill run and hook run were **not observed**.
- **Agents in GitHub** ([research §7b, hands-on][r-hands-on]): in one run the Copilot cloud agent ran
  hooks from `.github/hooks/*.json`, `.claude/settings.json` and `.github/copilot/settings.json`, and
  **installed a plugin enabled only in `.claude/settings.json` and ran its hook**. Code review ran no
  observed hooks. Claude Code cloud sessions don't install repo-enabled plugins
  ([research §7b, Claude Code outside a local machine][r-cloud]).
- **Summer's answers** ([research §7b, hook delivery][r-hooks]): `git-guard` isn't needed in cloud
  agents, because CI is enough there; `session-start` is the only cloud hook that matters, and its route
  is to be revisited. App repos are not all Node. A one-time trust prompt is acceptable. Define each hook
  once. A downloaded binary may or may not pass the beta company's security review.
- **Plugin hooks are the fragile part on update** ([research §7e][r-plugins]): reported Claude Code bugs
  leave bundled hooks on an old path after `autoUpdate` (#52218) or stop them in a running session after a
  mid-session update (#90135, open). Plugin skills update reliably. None of this was retested.
- **An enterprise can block the marketplace.** `strictKnownMarketplaces` restricts installs to
  allowlisted marketplaces ([research §7c][r-7c]).

## Decision

**Option 1: in an app repo, hooks reach local agents (Claude Code, Copilot CLI) from the folkways plugin
only. Agents running in GitHub get instructions and CI checks instead of hooks.**

- Each hook is defined in one place: the plugin's `hooks/hooks.json`. The app repo commits no hook
  entries; its settings only turn the plugin on.
- The hooks stay the Go binary from ADR 0002. This changes how the binary reaches an app repo, not what
  it is. The npm path in ADR 0002 stays where `folkways sync` runs today. Whether it stays there long
  term depends on the hub-or-per-repo sync decision.
- In GitHub, what `session-start` adds comes from committed instructions, and what `git-guard` checks is
  checked again by CI at PR time.

Rationale: the plugin is the only option that works in a non-Node repo without downloading a binary, and
it already exists. Summer has accepted the trust prompt it costs. For agents in GitHub, CI covers what
`git-guard` covers, just later, and no hook binary has to reach them.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| **1** | **Plugin only locally; instructions and CI in GitHub (H2 + H7)** | Works in any repo, Node or not; already built; one hook definition; nothing to install in the cloud | Trust prompt; no hooks for people who decline it or whose enterprise blocks the marketplace; a binary for every platform in each install; Copilot CLI hook run not yet seen; cloud checks happen at PR time, not before the command | Uses the code we have and covers every local tool, but each repo's enforcement depends on the plugin installing and updating correctly |
| 2 | npm in each app repo, as today (H1; architecture C) | Proven; committed and reviewable; no trust prompt | A Node manifest and `npm ci` in every Go, Java or Python repo; misses the cloud agent unless setup steps run `npm ci` | Fine for Node teams, and a non-starter for teams without Node |
| 3 | Committed launcher that downloads a pinned, checksummed binary (H3) | One committed definition that every surface with network can run, the cloud agent included | Cache, offline behavior, a Windows script; a security review of "downloads a binary" that may fail | The widest reach, but it rests on a security review nobody has done yet |
| 4 | Plugin locally, plus committed hooks and setup steps for the cloud agent (H2 + H4) | Hooks run before commands in the cloud agent too | Two definitions of each hook; a workflow file per repo, used only from the default branch; Copilot CLI may run hooks twice (unknown 11, untested) | Puts hooks in the cloud agent, which Summer says it doesn't need, at the cost of the single definition |

Committing the binaries (H5) and rewriting hooks as scripts (H6, which reverses ADR 0002) were dismissed
in the research for repo bloat and the reversal.

## Consequences

- **App repos carry no hook entries.** Their `.claude/settings.json` holds the marketplace and
  `enabledPlugins` block. Which other files an app repo carries belongs to ADR 0006.
- **`run.sh`'s synced-repo exit has to change** in any repo where the plugin is the only source. Repos
  that still have synced hook entries must not run the plugin's hooks too.
- **Open: the plugin's hooks will probably run in the Copilot cloud agent anyway** (unconfirmed). One run
  showed it installing a plugin enabled in `.claude/settings.json` and running its hook. So "no hooks in
  GitHub" needs either a hook that detects the cloud run and stays quiet, or acceptance that they run.
  Neither has been tested.
- **Open: how a hook binary reaches clones with no `node_modules` outside the plugin's reach** (Claude Code
  cloud sessions, a tool without plugins). This ADR doesn't solve that ([research §0][r-0]). H3 and H4 stay
  candidates if a cloud hook is ever needed.
- **Open: how `session-start`'s content reaches a cloud run.** Summer is to revisit this. One candidate is
  committed instructions plus a skill that reads `folkways.yaml` (unconfirmed).
- **Open: Copilot CLI with a login.** Auto-install, the skill run and the hook run (research §7a rows 2, 4,
  5) are untested. Until they pass, Copilot CLI hook parity is an assumption.
- **Open: running a hook twice (unknown 11)** was dropped, not tested ([research, new unknowns][r-unknowns]).
  It matters wherever a synced settings file and the plugin overlap, for example during migration.
- **Risk: weaker enforcement.** People who decline the trust prompt, enterprises that set
  `strictKnownMarketplaces`, and plugin-update bugs all leave a session without `git-guard`. That makes
  ADR 0001's mores in an app repo depend on the plugin. CI is the backstop at PR time. `doctor` should
  report a missing or stale plugin.
- **Risk: platforms.** The plugin ships no Windows binary, so a Windows engineer gets no hooks.
- **Follow-on:** CI checks matching `git-guard`'s commit and branch rules (M1, M5) for app repos; a
  Copilot CLI test with a login; a cloud-agent run of the plugin hooks.
- **Revisit** if the Copilot CLI test fails, if plugin hook updates keep breaking, or if a team needs
  hooks in the cloud agent.

[r-0]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#0-summers-answers-2026-10-16-and-what-they-imply
[r-7a]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#7a-test-results-2026-10-16
[r-hands-on]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#hands-on-results-2026-10-16-the-cloud-agent-does-more-than-its-docs-say
[r-cloud]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#claude-code-outside-a-local-machine
[r-hooks]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#hook-delivery-options-and-summers-answers-2026-10-16
[r-unknowns]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#new-unknowns-add-to-7
[r-7c]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#7c-the--claims-checked-session-13-2026-10-16
[r-plugins]: https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md#claude-code-team-plugins-in-practice
[p-readme]: ../../packages/claude-plugin/README.md
[p-run]: ../../packages/claude-plugin/src/run.sh
