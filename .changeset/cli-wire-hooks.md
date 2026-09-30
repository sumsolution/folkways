---
"@folkways/cli": minor
---

Wire the Go hooks into team repos: `settings.json` runs `folkways-hook` for each hook and sets `disableAllHooks: false`, `init` sets pnpm's `publicHoistPattern`, `sync` warns when the binary or that setting is missing, and `folkways hook <name>` runs a hook by hand.
