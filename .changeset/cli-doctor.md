---
"@folkways/cli": minor
---

`folkways doctor` checks a team repo and prints each problem with its fix, exiting 1 on errors: `hooks:` values in `folkways.yaml` other than `true`/`false`, the pinned, installed, running and synced CLI versions (a failed `upgrade`, `npx` drift), edited framework files, stale team overlay compiles, personal skills that shadow repo skills, `disableAllHooks`, and the hook binary and pnpm hoisting. Team repos gitignore `.folkways-sync-*/`. The instructions name the solo review path.
