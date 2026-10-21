# 0012. Does folkways need a plugin, or only a CLI?

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-21 |
| Owner | Summer Jordan |
| Reviewers | Summer Jordan |
| Review window | Until the IDE `PATH` check below is done |
| Supersedes | [0007](0007-hook-delivery.md) (proposed), on acceptance |

## Issue

folkways reaches a repo by up to five routes: `sync` commits files, the Claude plugin carries skills and hooks,
npm carries the binary, Homebrew or the install script carries it too, and a draft GitHub App syncs app repos
([go-cli 22, "Where the complexity comes from"][s-22-cx]). The plugin and the committed settings are two sources of
hooks, so the plugin's `run.sh` exits early in a synced repo, and the Phase 2 cutover of the Go CLI had to decide how a
committed hook finds its binary and which copy wins ([go-cli 21][s-21]). Summer has since settled that every repo
must commit what code review and outside contributors need: the full skills, not a minimum
([go-cli 22, answer 1][s-22-ans]). With skills committed everywhere, the plugin's remaining job is the hook binary.
Does folkways still need a plugin, or only a CLI?

## Context

- **Committed hook entries already exist** in synced repos: `packages/content/settings.json` registers
  `SessionStart`, `UserPromptSubmit`, `PreToolUse` (matcher `Bash`), `PostToolBatch` and `PreCompact`, each calling
  `node_modules/.bin/folkways-hook <name>` ([settings.json][c-settings], [`settings.go:15`][c-prefix]). That path works
  only with npm.
- **What the plugin gives today** ([go-cli 22, "What the plugin does today"][s-22-plugin]): skills under a
  `folkways:` namespace with nothing committed, a bundled hook binary for four platforms, install at the folder-trust
  prompt, marketplace updates, and hooks in the Copilot cloud agent ([ADR 0007](0007-hook-delivery.md) context). It
  is used only in repos that aren't synced ([run.sh][c-run]).
- **What the plugin can't do:** add permission rules, so the `git push` ask rule is missing where it is the source
  ([plugin README][c-readme]); apply overlays; update its hooks reliably (Claude Code #52218, #90135, per ADR 0007).
- **What it costs:** `packages/claude-plugin` (184 lines of build, 43 of `run.sh`, 264 of tests), the
  `FOLKWAYS_SKILL_NAMESPACE` handling in `internal/skill`, the marketplace, the second hook source, and Copilot plugin
  parity probing ([go-cli 22][s-22-plugin]).
- **A missing hook command blocks Copilot CLI.** A failing `preToolUse` command denies the tool call
  ([ADR 0002](0002-hook-runtime.md); probe Q5 in ADR 0007). Claude Code shows a non-blocking error instead.
- **ADR 0011** makes Homebrew, the install script and GitHub release binaries the first channel, npm the second,
  and leaves "which copy runs when the plugin's and `PATH`'s both exist" to the plan
  ([0011](0011-cli-in-go.md), Consequences).
- **`PATH` in a GUI-launched tool, partly checked** (2026-10-21, Summer's Mac, `ps eww` on a session's parent
  processes): `Claude.app` runs with `PATH=/usr/bin:/bin:/usr/sbin:/sbin`, but the Claude Code process it starts
  (2.1.304) has the login-shell `PATH`, including `/opt/homebrew/bin` ([go-cli 22, answer 6][s-22-six]). VS Code,
  JetBrains and Copilot (CLI and IDE) launched from the GUI are not checked.

## Decision

**Option 1: folkways is only a CLI. Every repo (brain, app, solo) commits its skills, instructions and hook entries;
the entries call `folkways hook <name>` on `PATH` behind a guard; each person installs the binary once.**

- The committed command is `command -v folkways >/dev/null 2>&1 || exit 0; folkways hook <name>`. A missing binary
  turns hooks off instead of blocking Copilot CLI. The `SessionStart` entry prints how to install instead of exiting
  silently.
- The binary comes from Homebrew or the install script (npm stays possible for Node repos, not required). Cloud agents
  install it in their setup step.
- The manifest's `version` is the pin; `session-start` warns once when the running binary differs
  ([go-cli 21 §2][s-21]).
- The plugin, its build, the skill namespace and the marketplace are removed.

Rationale: with skills committed everywhere (answer 1), the plugin carries only the hooks, and for that it costs a
second hook source, the precedence question, and the update bugs. One route for content (the binary embeds it, `sync`
writes it) and one hook source make permission rules work in every repo and remove PR 2's shim. The cost is one
install per person, which the guard and the `SessionStart` notice make visible instead of breaking tools.

Non-goals: how app repos get synced (draft ADR 0009), where an app repo's own config lives, Windows binaries. Size:
the hook-command change is about 20 Go lines plus sync and upgrade fixture changes; the plugin removal deletes about
500 lines and the marketplace, in its own PR.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| **1** | **CLI only; committed hooks call `folkways` on `PATH` behind a guard** | One hook source and one content route; permission rules everywhere; no precedence rule; smallest release (npm optional) | One install per person; hooks quietly off without it, apart from the `SessionStart` notice; GUI `PATH` not checked for every tool; no zero-footprint use in a repo you can't commit to | Trades install-on-trust for one install step, and removes a delivery route and its edge cases |
| 2 | Plugin for hooks in app repos, `sync` in brain repos ([ADR 0007](0007-hook-delivery.md) as proposed) | No install step; already built; reaches the Copilot cloud agent | Two hook sources and `run.sh`'s exit; skills twice if a synced repo enables it; no permission rules; plugin hook update bugs; trust prompt or marketplace block leaves no hooks, silently | Zero install for teammates, paid for with a second source and the edge cases it brings |
| 3 | The plugin is everything: skills, hooks and binary; repos commit only the minimum | Smallest repo footprint; one install route | Code review and outside contributors don't get the full skills, which answer 1 requires; a team plugin per team | Ruled out by answer 1 |
| 4 | CLI only, with a committed wrapper that fetches the pinned binary on first run (`gradlew`-style; H3 in ADR 0007) | No manual install; the pin is enforced | Network and checksum work at hook time, a cache, offline behaviour; a new piece of code and a security review of "downloads a binary" | Worth adding if the install step proves a real barrier, not before |
| 5 | npm in each repo, as today | Proven; pinned by `package.json` | A Node manifest in every Go, Java or Python repo | Fine for Node teams only |

Committing the binaries and rewriting hooks as shell scripts were dismissed in ADR 0007's research for repo bloat
and for reversing ADR 0002.

## Consequences

- **One install per person.** `onboard` and `doctor` check that `folkways` is on `PATH`. Without it, `git-guard` and
  the other hooks are off for that person, the same class of risk ADR 0007 accepted for people who decline the trust
  prompt. CI stays the backstop at PR time.
- **Cloud agents need a setup step.** The Copilot cloud agent gets the binary from a committed
  `.github/workflows/copilot-setup-steps.yml` that sync writes (draft ADR 0009 already allows for the Workflows
  permission). Hooks still exit at once when `COPILOT_AGENT_SESSION_ID` is set, as ADR 0007 decided. Claude Code cloud
  sessions need their environment's setup script (inferred, not tested).
- **Removed:** `packages/claude-plugin`, `run.sh` and its synced-repo exit, skill namespacing, the marketplace, the
  `node_modules/.bin/folkways-hook` path, PR 2's planned shim, and ADR 0011's copy-precedence question.
- **Lost:** use in a repo you can't commit to (inferred to be rare), and one-click discovery through the marketplace.
- **ADR 0011:** npm becomes optional; the first release needs GoReleaser, the GitHub binaries, the tap and the install
  script. 0011 gets a note on acceptance, as 0002 did.
- **ADR 0007** is superseded on acceptance. Its cloud-agent exit and `CLAUDE_PROJECT_DIR` rules carry over.
- **Sync migrates** existing entries: today's prefix joins `oldHookPrefixes` ([`settings.go:18`][c-prefix]).
- **Open, before acceptance:** whether VS Code, JetBrains and Copilot (CLI and IDE) launched from the GUI see
  Homebrew's `PATH`. If one doesn't, the guard can also try `/opt/homebrew/bin` and `~/.local/bin` directly, a few
  lines, before this decision changes.
- **Revisit** if the install step stops teammates from getting hooks (then option 4), or if a team needs folkways in
  repos it can't commit to.

## Related decisions

- [0007](0007-hook-delivery.md): superseded on acceptance.
- [0011](0011-cli-in-go.md): npm channel made optional; copy precedence removed.
- [0002](0002-hook-runtime.md): the hook stays the Go binary; its npm delivery is no longer required.
- [0006](0006-app-repo-files.md): app repos still carry Claude-format files only.

[s-21]: https://github.com/sumsolution/folkways-command/blob/e25a7a3/efforts/go-cli/21-decision-hook-command-and-pin.md
[s-22-cx]: https://github.com/sumsolution/folkways-command/blob/e25a7a3/efforts/go-cli/22-step-back-repo-model.md#where-the-complexity-comes-from
[s-22-ans]: https://github.com/sumsolution/folkways-command/blob/e25a7a3/efforts/go-cli/22-step-back-repo-model.md#summers-answers-2026-10-20
[s-22-plugin]: https://github.com/sumsolution/folkways-command/blob/e25a7a3/efforts/go-cli/22-step-back-repo-model.md#what-the-plugin-does-today
[s-22-six]: https://github.com/sumsolution/folkways-command/blob/e25a7a3/efforts/go-cli/22-step-back-repo-model.md#summers-answer-to-6-2026-10-21
[c-settings]: https://github.com/sumsolution/folkways/blob/9bda25a/packages/content/settings.json
[c-prefix]: https://github.com/sumsolution/folkways/blob/9bda25a/packages/folkways/internal/settings/settings.go#L13-L18
[c-run]: https://github.com/sumsolution/folkways/blob/9bda25a/packages/claude-plugin/src/run.sh
[c-readme]: https://github.com/sumsolution/folkways/blob/9bda25a/packages/claude-plugin/README.md
