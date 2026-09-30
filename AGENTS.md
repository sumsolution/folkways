# folkways: framework development

Instructions for working on the framework itself. Not shipped; team repos get
`packages/cli/assets/instructions.md`.

## Layout

A pnpm workspace ([ADR 0003](docs/adr/0003-repo-layout.md)). Each package owns its source, tests and build.

| Path | Holds |
|---|---|
| `packages/cli/` | `@folkways/cli`, published to npm (spec §3). Team repos pin an exact version |
| `packages/cli/src/` | TypeScript source of the `folkways` CLI. `tsc` compiles it to `dist/` on publish |
| `packages/cli/assets/` | Text files `sync` copies into team repos (spec §2.1) |
| `packages/cli/templates/` | One-time scaffold for `init`. `dot-x` becomes `.x`; `name.tmpl` becomes `name` |
| `packages/cli/test/` | vitest suites for the CLI |
| `packages/hooks/` | `@folkways/hooks` (private): the `folkways-hook` binary, shipped as `@folkways/hooks-<os>-<cpu>` (ADR 0002) |
| `packages/hooks/contract/` | Black-box hook contract suite: fixtures, harness, rules |
| `scripts/` | Cross-package scripts: `smoke.ts` (packed-tarball smoke test) |
| `research/` | Benchmarks and evals behind decisions. Not in the workspace; CI doesn't run it |
| `docs/design/` | Design record. `00-STATUS.md` is the entry point; `06-v1-spec.md` is the spec |
| `docs/adr/` | Framework decisions |

## Commands

From the repo root:

| Command | Does |
|---|---|
| `pnpm install` | install dev dependencies (pnpm version pinned in `packageManager`) |
| `pnpm test` | every package's tests |
| `pnpm run check` | every package's `tsc --noEmit` |
| `pnpm run build` | build every package (the CLI's `src/` to `dist/`; also runs on `pnpm pack` / `pnpm publish`) |
| `pnpm --filter @folkways/cli test` | one package only (same for `check`, `build`, `@folkways/hooks`) |
| `pnpm run smoke` | pack the CLI, install the tarball, drive `init` and `sync` |
| `cd packages/hooks && go build -o bin/folkways-hook .` | build the hook binary (Go 1.26+; `bin/` is gitignored; add `-ldflags "-X main.version=x.y.z"` to set `--version`) |
| `FOLKWAYS_HOOK_CMD="$PWD/packages/hooks/bin/folkways-hook" FOLKWAYS_HOOKS=git-guard pnpm --filter @folkways/hooks test` | `go vet`, `go test`, then the hook contract suite against the built binary |
| `node packages/cli/src/cli.ts <command>` | run the CLI from source (Node 24 strips types) |
| `pnpm changeset` | record a change for the next release of the packages it touches |

## Rules

- **No company-specific content.** Names, portals, project keys, internal URLs and paths belong in team
  repos. Examples use synthetic values (`Example Team`, `EX`, `example.com`).
- **TypeScript:** Node 24+, ESM. Only erasable syntax (no `enum`, `namespace`, parameter properties); import
  relative files with `.ts` extensions; `import type` for types. `tsconfig.base.json` enforces these.
- **Dependencies:** runtime deps are fine when they earn their place; they install into team repos.
  `pnpm-workspace.yaml` blocks versions younger than 7 days, provenance downgrades, non-registry
  subdependencies and unapproved build scripts. Don't relax these to make an install pass; wait or ask.
- **Package manager:** pnpm for this repo. Team repos use npm or pnpm (`packages/cli/src/package-manager.ts`, spec §3.1).
- **CI actions** are pinned to full commit SHAs with a version comment (the repo requires it).
- **TDD:** write a failing test first, then fix. Run `pnpm run check && pnpm test` before every commit.
- **Expected failures** throw `FolkwaysError`; the CLI prints `folkways: <message>` with no stack trace.
- **Skills:** follow [docs/skill-authoring.md](docs/skill-authoring.md) (skeleton, overlay IDs, headless
  rules, trigger eval). `packages/cli/test/skills.test.ts` checks the mechanical rules.
- **Commits:** Angular conventional commits, module scope (`cli`, `assets`, `templates`, `skills`, `hooks`,
  `docs`, `test`, `build`, `research`). Squash PRs.
- **Tests live with their package:** TypeScript in `packages/<pkg>/test/` (or `contract/`), Go as `*_test.go`
  beside the code. Cross-package checks go in `scripts/`. A package never imports another's source.
- **Versions are per package** (changesets, ADR 0003). A PR that changes a package adds a changeset
  (`pnpm changeset`); docs, research and CI-only changes don't need one. Release: merge the
  `changeset version` PR, then `pnpm changeset tag` and push the tags; each tag publishes its package.
- **The CLI pins one hooks version** in `folkways.hooks` of `packages/cli/package.json`. Bump it to ship a
  hooks release to team repos. Hook renames, removals and event changes are major hooks bumps.
- **CI** runs only the jobs for the packages a change touches; the `ci` job is the required check. A new
  package gets a filter and jobs in `.github/workflows/ci.yml`.
- Keep `docs/design/00-STATUS.md` current at milestones.
