---
"@folkways/hooks": minor
---

Node-free (ADR 0012). Committed hook entries call `folkways hook <name>` on `PATH` behind a guard that turns them off when the binary is missing, and `SessionStart` then says how to install; sync migrates the old `node_modules/.bin/folkways-hook` entries. `init` no longer writes `package.json` or installs (`--no-install` and `--package-manager` are gone). The manifest's `version` is the pin: `upgrade <version>` runs with that version installed and syncs in-process, `doctor` warns when the running binary differs or `folkways` isn't on `PATH`, and `session-start` adds the same version warning.
