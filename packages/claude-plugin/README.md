# @folkways/claude-plugin

Builds the folkways Claude Code plugin: the framework's skills, instructions and hooks, for repos that
aren't synced team repos. `pnpm --filter @folkways/claude-plugin run build` writes a complete plugin to
`dist/` (needs Go, to cross-compile `folkways-hook`).

| In `dist/` | From |
|---|---|
| `skills/` | `packages/cli/assets/skills/`, rewritten for the plugin: `folkways:<name>` skill names, `${CLAUDE_PLUGIN_ROOT}` reference paths |
| `instructions.md` | `packages/cli/assets/instructions.md`, rewritten the same way; added as context at session start |
| `hooks/hooks.json` | generated from `packages/cli/assets/settings.json`, so the plugin runs the same hooks on the same events |
| `hooks/run.sh`, `hooks/bin/` | the runner and `folkways-hook` for darwin and linux, arm64 and x64 |
| `.claude-plugin/plugin.json` | the manifest, at this package's version |
| `.claude-plugin/marketplace.json` | a `folkways-dev` marketplace for dogfooding the build in place; not published |

## Dogfood

```bash
pnpm --filter @folkways/claude-plugin run build
claude plugin marketplace add "$PWD/packages/claude-plugin/dist"   # once; loads in place
claude plugin install folkways@folkways-dev
# or for one session: claude --plugin-dir packages/claude-plugin/dist
```

Rebuild after changing skills, instructions or hooks; the next session picks it up.

## What differs from a team repo

- **Team overlays** in `overlays/` aren't compiled in (that needs `folkways sync`). Personal overlays
  (`.local/me/overlays/`) still attach.
- **The push ask rule** isn't set: plugins can't add permission rules.
- **Synced team repos** (with `.folkways/manifest.json`) are left alone: their own settings run the hooks.
