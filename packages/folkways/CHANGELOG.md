# Changelog

## [0.3.0](https://github.com/sumsolution/folkways/compare/v0.2.0...v0.3.0) (2026-10-28)


### Features

* **hooks:** let brain repo sessions push without asking ([#71](https://github.com/sumsolution/folkways/issues/71)) ([ef2ad0e](https://github.com/sumsolution/folkways/commit/ef2ad0ed0327cdfe622ea27dd1cb9831c20c8f50))

## 0.2.0

The first release. `folkways` is one binary for macOS and Linux (amd64 and arm64), installed with Homebrew
(`brew install sumsolution/tap/folkways`), the install script or `go install`; see the README's Install section.
Team repos need no Node (ADR 0011, ADR 0012).

### Commands
- `init <path>`: scaffold a team repo, sync and commit. `.folkways/manifest.json` records the version: the team's pin.
- `sync [--force]`: copy the framework's instructions, skills and prompts in, compile team overlays, merge
  `.claude/settings.json`. Stops when a framework file was edited.
- `upgrade <version>`: on branch `chore/folkways-<version>`, sync that version and commit. Run it with that version
  installed.
- `doctor`: check config, versions, overlays, skills, settings and hooks.
- `hook <name>`, `version`. `review` and `run` aren't built yet.

### Hooks
Committed in `.claude/settings.json` for Claude Code and Copilot CLI. Each calls `folkways hook <name>` on `PATH`;
without the binary the hooks are off and the session start says how to install.
- `git-guard`: commit format per `git.commit_convention`; with `review.trigger: pre-push`, a review of HEAD before a
  push; no checkout, switch, pull, reset or rebase in your clones under `projects_dir`. The brain repo is exempt
  from the last two.
- `session-start`: points the agent at the `start` skill, prints the config, and warns when your `folkways` differs
  from the repo's pinned version.
- `context-band`: announces 125K, 150K and 200K tokens (Claude Code only; it needs the transcript).
- `precompact`: logs a missed handoff.
- `hooks.<name>: false` in `folkways.yaml` turns one off. Without `folkways.yaml` they do nothing, except that
  `session-start` offers setup once per repo.

### Skills
16 skills (`setup`, `onboard`, `start`, `status`, `handoff`, `session-close`, `research`, `side-quest`, `ticket`,
`adr`, `tests`, `commit`, `review`, `pr`, `qa`, `knowledge-audit`) and the review prompt.

Licensed under Apache-2.0.
