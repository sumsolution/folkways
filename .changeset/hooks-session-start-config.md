---
"@folkways/hooks": minor
---

`session-start` prints the config once per session and after each compaction: the personal root, the effective review trigger, and `folkways.yaml` and `.local/me.yaml` as they are (a file over 4096 bytes is named, not printed), so skills don't read the files. A handle with whitespace no longer counts as safe.
