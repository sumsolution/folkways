# folkways: framework development

Instructions for working on the framework itself. Not shipped; team repos get `assets/instructions.md`.

## Layout

| Path | Holds |
|---|---|
| `src/` | TypeScript source of the `folkways` CLI. `tsc` compiles it to `dist/` on publish |
| `hooks/` | Go source of `folkways-hook` (module `github.com/sumsolution/folkways/hooks`). Not in the npm package |
| `assets/` | Text files `sync` copies into team repos (spec §2.1) |
| `templates/` | One-time scaffold for `init`. `dot-x` becomes `.x`; `name.tmpl` becomes `name` |
| `test/` | vitest suites |
| `docs/design/` | Design record. `00-STATUS.md` is the entry point; `06-v1-spec.md` is the spec |

Published to npm as `@folkways/cli` (spec §3). Team repos pin an exact version.

## Commands

| Command | Does |
|---|---|
| `pnpm install` | install dev dependencies (pnpm version pinned in `packageManager`) |
| `pnpm test` | vitest |
| `pnpm run check` | `tsc --noEmit` |
| `pnpm run build` | compile `src/` to `dist/` (runs on `pnpm pack` / `pnpm publish`) |
| `cd hooks && go vet ./... && go test ./...` | vet and test the Go hooks (Go 1.24+) |
| `cd hooks && go build -o bin/folkways-hook .` | build the hook binary (`bin/` is gitignored; add `-ldflags "-X main.version=x.y.z"` to set `--version`) |
| `FOLKWAYS_HOOK_CMD="$PWD/hooks/bin/folkways-hook" FOLKWAYS_HOOKS=git-guard pnpm test test/hooks` | run the hook contract suite against the built binary |
| `node src/cli.ts <command>` | run the CLI from source (Node 24 strips types) |

## Rules

- **No company-specific content.** Names, portals, project keys, internal URLs and paths belong in team
  repos. Examples use synthetic values (`Example Team`, `EX`, `example.com`).
- **TypeScript:** Node 24+, ESM. Only erasable syntax (no `enum`, `namespace`, parameter properties); import
  relative files with `.ts` extensions; `import type` for types. `tsconfig.json` enforces these.
- **Dependencies:** runtime deps are fine when they earn their place; they install into team repos.
  `pnpm-workspace.yaml` blocks versions younger than 7 days, provenance downgrades, non-registry
  subdependencies and unapproved build scripts. Don't relax these to make an install pass; wait or ask.
- **Package manager:** pnpm for this repo. Team repos use npm or pnpm (`src/package-manager.ts`, spec §3.1).
- **CI actions** are pinned to full commit SHAs with a version comment (the repo requires it).
- **TDD:** write a failing test first, then fix. Run `pnpm run check && pnpm test` before every commit.
- **Expected failures** throw `FolkwaysError`; the CLI prints `folkways: <message>` with no stack trace.
- **Skills:** follow [docs/skill-authoring.md](docs/skill-authoring.md) (skeleton, overlay IDs, headless
  rules, trigger eval). `test/skills.test.ts` checks the mechanical rules.
- **Commits:** Angular conventional commits, module scope (`cli`, `assets`, `templates`, `skills`, `hooks`,
  `docs`, `test`, `build`). Squash PRs.
- **Releases:** bump `version` in `package.json`, add a `CHANGELOG.md` entry, tag `v<version>`, publish.
- Keep `docs/design/00-STATUS.md` current at milestones.
