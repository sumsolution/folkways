# Example Team

The team's shared agent workspace, built on [folkways](https://github.com/sumsolution/folkways).
Open Claude Code or Copilot CLI here. New to the repo? Run the `onboard` skill.

| Path | Holds | Changes by |
|---|---|---|
| `team.yaml` | Team config | PR |
| `knowledge/` | Durable facts, each with `Last verified:` | PR |
| `state/` | Live trackers, one file per track | direct push |
| `efforts/<KEY>-<slug>/` | Shared efforts, each with `00-STATUS.md` | direct push |
| `overlays/<skill>/` | Team changes to framework skills | PR |
| `people/<handle>/` | Personal overlays, inbox, work | direct push |
| `.claude/skills/` | Framework skills (vendored) and team skills | PR |
| `.folkways/` | Vendored framework. Don't edit; `folkways upgrade` replaces it | PR |
| `.local/` | Machine-local config (`me.yaml`). Not committed | — |

Direct pushes are rebased and linear. Upgrade the framework with
`.folkways/bin/folkways upgrade <version>`.
