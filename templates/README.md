# Example Team

The team's shared agent workspace, built on [folkways](https://github.com/sumsolution/folkways).
New to the repo? Run `npm ci`, open Claude Code or Copilot CLI here, and run the `onboard` skill.

| Path | Holds | Changes by |
|---|---|---|
| `team.yaml` | Team config | PR |
| `knowledge/` | Durable facts, each with `Last verified:` | PR |
| `state/` | Live trackers, one file per track | direct push |
| `efforts/<KEY>-<slug>/` | Shared efforts, each with `00-STATUS.md` | direct push |
| `overlays/<skill>/` | Team changes to framework skills | PR |
| `people/<handle>/` | Personal overlays, inbox, work | direct push |
| `.claude/skills/` | Framework skills (synced) and team skills | PR |
| `.folkways/` | Framework files. Don't edit; `folkways sync` replaces them | PR |
| `.local/` | Machine-local config (`me.yaml`). Not committed | — |

Direct pushes are rebased and linear. Upgrade the framework with `npx folkways upgrade <version>`, then
open a PR. Run `npm ci` after pulling an upgrade.
