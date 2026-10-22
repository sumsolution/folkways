---
"@folkways/claude-plugin": patch
---

`onboard`, `setup`, `commit` and `pr` call `folkways` on `PATH` instead of `node_modules/.bin/folkways` or `npx`.
