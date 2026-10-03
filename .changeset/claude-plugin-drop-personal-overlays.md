---
"@folkways/claude-plugin": minor
---

Skills no longer read config files or attach personal overlays in a step 0: they take their config from the `session-start` hook's context. Solo repos have no overlays for now (`overlays/` is ignored).
