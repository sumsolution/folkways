# people

One folder per member, named by the `handle` in `team.yaml`. The `onboard` skill creates it.

| Path | Holds |
|---|---|
| `overlays/<skill>/OVERLAY.md` | Personal additions to a skill (add only) |
| `inbox/` | Side quests and ticket drafts with no effort |
| `work/<KEY>-<slug>/` | Personal efforts; promote with `git mv` to `efforts/` |

Pushed directly, rebased. Anything private goes in `.local/`, which is not committed.
