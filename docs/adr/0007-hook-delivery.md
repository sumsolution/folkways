# 0007. How do hooks reach agents in an app repo?

| | |
|---|---|
| Status | superseded by [0012](0012-cli-only.md): no plugin; repos commit their hooks. The cloud-agent exit carries over |
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
  repo. Copilot CLI 1.0.98, logged in, single runs ([probe Q5][q-q5]): one prompt, folder trust; then the
  plugin auto-installed, its skill ran, and its `SessionStart` and `PreToolUse` (matcher `Bash`) hooks ran
  with `CLAUDE_PROJECT_DIR` and `COPILOT_PROJECT_DIR` set to the repo. The hook's working directory was the
  plugin root, and `SessionStart` stdin had `source: "new"`.
- **Copilot CLI fails closed on `PreToolUse`** ([probe Q5][q-q5]). When the hook command failed (exit 127,
  script not found), the log said `(fail-closed)` and the tool call was denied. A failing `SessionStart`
  was only logged. ADR 0002 already expected this from the docs.
- **Headless Copilot CLI runs need saved trust or an explicit install** ([probe Q5][q-q5]). `copilot -p`
  loads a repo-enabled plugin only if the folder was saved in `trustedFolders` ("Yes, and remember this
  folder") or the plugin was installed with `copilot plugin install`. Otherwise the run has no plugin
  skills or hooks, and shows no error. Headless runs never auto-installed the plugin.
- **The Copilot cloud agent runs the same plugin's hooks.** Second probe, private repo
  [folkways-agent-sandbox][q-repo], one run per result, plugin enabled only in `.claude/settings.json`,
  no other hook source in the repo. The agent cloned the plugin from the job's starting commit.
  `SessionStart` ran in 2 of 2 runs ([PR #3][q-3], [PR #8][q-8]). `PreToolUse` (matcher `Bash`) ran in the
  one run that used tools (#8). The first probe saw the same ([research §7b, hands-on][r-hands-on]).
- **A hook can tell it's in the cloud agent.** These were set in both cloud runs:
  `COPILOT_AGENT_SESSION_ID`, `COPILOT_AGENT_ACTION=fix`, `GITHUB_ACTOR=copilot-swe-agent[bot]`. Local
  Copilot CLI set no `COPILOT_AGENT_*` variable ([probe Q5][q-q5]). `CI` and `GITHUB_ACTIONS` don't tell
  the cloud apart, because every Actions job sets them, and `COPILOT_CLI=1` is set both locally and in the
  cloud.
- **The cloud hook runs in a different environment.** Its working directory is the plugin root, not the
  repo. `CLAUDE_PROJECT_DIR` and `COPILOT_PROJECT_DIR` point at the workspace. `SessionStart` stdin has
  `source: "new"`, where Claude Code sends `startup`, `resume`, `clear` or `compact`. The environment also
  holds `COPILOT_SDK_AUTH_TOKEN`, `GITHUB_VERIFICATION_TOKEN` and `COPILOT_AGENT_CALLBACK_URL` (the probe
  recorded names only). The probe didn't check whether `SessionStart` output reaches the agent's context.
- **Other agents in GitHub:** hooks in Copilot code review weren't checked. Claude Code cloud sessions
  don't install repo-enabled plugins ([research §7b, Claude Code outside a local machine][r-cloud]).
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

**Option 1: in an app repo, the folkways plugin is the only source of hooks, and every hook exits at once
in the Copilot cloud agent. Agents running in GitHub get committed instructions and CI checks instead of
hooks.**

- Each hook is defined in one place: the plugin's `hooks/hooks.json`. The app repo commits no hook
  entries; its settings only turn the plugin on. Local Claude Code and Copilot CLI run the hooks.
- The plugin also reaches the cloud agent, and that can't be switched off per surface. So the hook binary
  checks `COPILOT_AGENT_SESSION_ID` first and exits 0 when it is set. That covers every hook,
  `session-start` included.
- In GitHub, what `session-start` adds comes from committed instructions, which the cloud agent already
  reads. What `git-guard` checks is checked again by CI at PR time.
- The hooks stay the Go binary from ADR 0002. This changes how the binary reaches an app repo, not what
  it is. The npm path in ADR 0002 stays where `folkways sync` runs today. Whether it stays there long
  term depends on the hub-or-per-repo sync decision.

Rationale: the plugin is the only option that works in a non-Node repo without downloading a binary, and
it already exists. Summer has accepted the trust prompt it costs. In the cloud agent, `git-guard` isn't
wanted, and nobody has shown that `session-start`'s output reaches the agent. Committed instructions do
reach it. Exiting at once keeps one behavior for every hook, and it keeps hook logic out of a process that
holds live agent credentials. The cost: one environment check, which rests on single runs of each
surface.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| **1** | **Plugin only; every hook exits in the cloud agent; instructions and CI in GitHub (H2 + H7)** | Works in any repo, Node or not; already built; one hook definition; the cloud agent runs no folkways logic | Trust prompt; no hooks for people who decline it or whose enterprise blocks the marketplace; a binary for every platform in each install; headless Copilot CLI runs get no hooks without saved trust or an explicit install; the cloud check rests on single runs | Uses the code we have and works on every surface tested, but whether a session has hooks depends on the plugin being installed and trusted, and it doesn't say when it isn't |
| 1b | As 1, but `session-start` still runs in the cloud agent | Team config reaches the cloud run through the same hook as locally | Its output reaching the agent is unobserved; `source: "new"` and the plugin-root cwd need handling; more code runs beside live credentials | Worth it only if a test shows the output lands and instructions can't carry it |
| 1c | As 1, with no cloud check: every hook runs in the cloud agent | No detection code | `git-guard` runs where Summer says it's unneeded, and a crash or nonzero exit makes Copilot deny the tool call; nothing has tested `git-guard` there | Saves one check and accepts untested blocking in the cloud |
| 2 | npm in each app repo, as today (H1; architecture C) | Proven; committed and reviewable; no trust prompt | A Node manifest and `npm ci` in every Go, Java or Python repo; misses the cloud agent unless setup steps run `npm ci` | Fine for Node teams, and a non-starter for teams without Node |
| 3 | Committed launcher that downloads a pinned, checksummed binary (H3) | One committed definition that every surface with network can run | Cache, offline behavior, a Windows script; a security review of "downloads a binary" that may fail | The widest reach, but it rests on a security review nobody has done yet |
| 4 | Plugin locally, plus committed hooks and setup steps for the cloud agent (H2 + H4) | Hooks run before commands in the cloud agent too | Two definitions of each hook; the plugin's hooks run in the cloud agent as well; a workflow file per repo; Copilot CLI may run hooks twice (unknown 11, untested) | Puts hooks in the cloud agent, which Summer says it doesn't need, at the cost of the single definition |

Committing the binaries (H5) and rewriting hooks as scripts (H6, which reverses ADR 0002) were dismissed
in the research for repo bloat and the reversal.

## Consequences

- **App repos carry no hook entries.** Their `.claude/settings.json` holds the marketplace and
  `enabledPlugins` block. Which other files an app repo carries belongs to ADR 0006.
- **The binary gains a cloud check:** exit 0 when `COPILOT_AGENT_SESSION_ID` is set, before any other
  work. It goes in the binary, not in `run.sh`, so every route to the binary gets it.
- **Hooks resolve the repo from `CLAUDE_PROJECT_DIR`, never the cwd.** In the cloud agent the cwd is the
  plugin root, so a cwd fallback reads the wrong folder. ADR 0002 already prefers the variable. Any code
  that reads `SessionStart`'s `source` must accept values other than Claude Code's four, such as `new`.
- **The plugin's hook code is as sensitive as a workflow file.** In the cloud agent any plugin hook runs
  with `COPILOT_SDK_AUTH_TOKEN`, `GITHUB_VERIFICATION_TOKEN` and `COPILOT_AGENT_CALLBACK_URL` in its
  environment. Changes to the plugin's hooks, and to a team plugin built on it, need the same review as
  CI changes. Hooks must never log their environment.
- **`run.sh`'s synced-repo exit has to change** in any repo where the plugin is the only source. Repos
  that still have synced hook entries must not run the plugin's hooks too.
- **A broken hook blocks every tool in Copilot CLI.** A failing `PreToolUse` command denies the tool call,
  so the binary must never fail on its own errors: it exits 0 and prints a warning, as ADR 0002 already
  requires. `run.sh` exits 0 when the binary is missing. What neither can catch is a hook command that
  can't start, such as a broken install path. In that case Copilot CLI denies every Bash call until the
  plugin is fixed. `doctor` should check that the plugin's hook command runs.
- **Open: the cloud check rests on single runs** of one cloud runtime (`runtime-cca-vendored-e7c2e09`).
  GitHub could rename the variables. If it does, the check stops matching and every hook runs in the
  cloud, as in option 1c.
- **Open: hooks in Copilot code review** weren't checked.
- **Open: how a hook binary reaches clones with no `node_modules` outside the plugin's reach** (Claude Code
  cloud sessions, a tool without plugins). This ADR doesn't solve that ([research §0][r-0]). H3 and H4 stay
  candidates if a cloud hook is ever needed.
- **Open: running a hook twice (unknown 11)** was dropped, not tested ([research, new unknowns][r-unknowns]).
  It matters wherever a synced settings file and the plugin overlap, for example during migration.
- **Risk: weaker enforcement.** People who decline the trust prompt, enterprises that set
  `strictKnownMarketplaces`, and plugin-update bugs all leave a session without `git-guard`. So does a
  headless `copilot -p` run in a folder whose trust wasn't saved, unless the plugin was installed with
  `copilot plugin install`, and it shows no error. `onboard` should tell Copilot CLI users to pick
  "Yes, and remember this folder", and scripts or CI that run Copilot headless should install the plugin
  explicitly. All of this makes ADR 0001's mores in an app repo depend on the plugin. CI is the
  backstop at PR time. `doctor` should report a missing or stale plugin.
- **Risk: platforms.** The plugin ships no Windows binary, so a Windows engineer gets no hooks.
- **Follow-on:** the cloud check in the binary; CI checks matching `git-guard`'s commit and branch rules
  (M1, M5) for app repos; a repeat cloud-agent and Copilot CLI run with the check in place.
- **Revisit** if a later Copilot CLI sets `COPILOT_AGENT_*` locally, if plugin hook updates keep
  breaking, or if a team needs `session-start` in the cloud agent (option 1b).

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
[q-repo]: https://github.com/sumsolution/folkways-agent-sandbox
[q-3]: https://github.com/sumsolution/folkways-agent-sandbox/pull/3
[q-8]: https://github.com/sumsolution/folkways-agent-sandbox/pull/8
[q-q5]: https://github.com/sumsolution/folkways-command/blob/e7781d5/efforts/app-repo-first/23-probe-copilot-github.md#q5-copilot-cli-with-a-login-2026-10-17-session-15
