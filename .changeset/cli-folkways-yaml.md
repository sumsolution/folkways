---
"@folkways/cli": minor
---

`team.yaml` is now `folkways.yaml`, with `mode`, `git.commit_pattern` and a `hooks:` block; `sync` warns when it's missing. The `setup` skill is a guided setup for a solo or team brain repo (where sessions start; it holds no code), `start` and `onboard` offer it when `folkways.yaml` is missing, and skills write personal paths against the personal root (`people/<handle>/`, or the repo root in a solo repo).
