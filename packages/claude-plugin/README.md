# @folkways/claude-plugin

Builds the folkways Claude Code plugin: the framework's skills, instructions and hooks, for repos that
aren't synced team repos. `pnpm --filter "@folkways/claude-plugin..." run build` cross-compiles
`folkways-hook` with `@folkways/hooks`'s build (needs Go), then writes a complete plugin to `dist/`. The
plugin's own build only reads the hooks build in `packages/hooks/npm/`; `node src/build.ts <out> <hooks-build>`
takes another one.

| In `dist/` | From |
|---|---|
| `skills/` | `packages/content/skills/`, rewritten for the plugin: `folkways:<name>` skill names, `${CLAUDE_PLUGIN_ROOT}` reference paths |
| `instructions.md` | `packages/content/instructions.md`, rewritten the same way; added as context at session start |
| `hooks/hooks.json` | generated from `packages/content/settings.json`, so the plugin runs the same hooks on the same events |
| `hooks/run.sh`, `hooks/bin/` | the runner and `folkways-hook` for darwin and linux, arm64 and x64. The runner sets `FOLKWAYS_SKILL_NAMESPACE=folkways`, so the hooks name skills as `folkways:<name>` |
| `.claude-plugin/plugin.json` | the manifest, at this package's version |
| `.claude-plugin/marketplace.json` | a `folkways-dev` marketplace for dogfooding the build in place; not published |

## Dogfood

```bash
pnpm --filter "@folkways/claude-plugin..." run build
claude plugin marketplace add "$PWD/packages/claude-plugin/dist"   # once; loads in place
claude plugin install folkways@folkways-dev
# or for one session: claude --plugin-dir packages/claude-plugin/dist
```

Rebuild after changing skills, instructions or hooks; the next session picks it up.

## Repo states

`hooks/run.sh` acts on what the project holds (spec 2.3):

| Repo | Plugin does |
|---|---|
| Synced team repo (`.folkways/manifest.json`) | nothing; its own settings run the hooks |
| Configured (`folkways.yaml`) | prints the instructions and runs every hook |
| Unconfigured (no `folkways.yaml`) | prints nothing and runs no hook except `session-start`, whose binary decides whether to offer one-time setup (the `folkways:setup` skill). With no binary for the platform it prints nothing |

## What differs from a team repo

- **Overlays** aren't applied. Team overlays in `overlays/` need `folkways sync` to compile them in, and
  personal overlays no longer exist, so a solo repo has no overlays until the Go CLI.
- **The push ask rule** isn't set: plugins can't add permission rules.
- **Synced team repos** (with `.folkways/manifest.json`) are left alone: their own settings run the hooks.
- **Hooks without a team repo** read no `folkways.yaml` or `me.yaml`. `git-guard` still denies commits that
  aren't Angular `type(scope): subject`; `precompact` writes `.local/precompact.log` into the repo;
  `context-band` and the main-clone and pre-push checks stay off. Open question 3 in
  [00-STATUS](https://github.com/sumsolution/folkways-command/blob/main/state/framework.md).
