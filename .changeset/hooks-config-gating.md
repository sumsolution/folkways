---
"@folkways/hooks": minor
---

The hooks read `folkways.yaml` (was `team.yaml`) and do nothing without it, except `session-start`, which offers setup once per repo (recorded in `${XDG_CONFIG_HOME:-~/.config}/folkways/offered`, never in the repo). `hooks.<name>: false` turns a hook off. `git-guard` checks commits per `git.commit_convention`: `angular` (default), `pattern` (`git.commit_pattern`) or `none`. With `mode: solo` the review lookup reads `work/` and `inbox/` at the repo root and needs no handle.
