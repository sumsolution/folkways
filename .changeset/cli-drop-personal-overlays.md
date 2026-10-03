---
"@folkways/cli": minor
---

Personal overlays are removed. `sync` no longer lints `people/<handle>/overlays/` or links `.local/me`, and `doctor` drops the personal-overlay check. Framework skills replace their step 0 file reads with a `## Config` section; the values come from the `session-start` hook. A leftover `.local/me` link is unused and can be deleted.
