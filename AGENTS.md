# folkways: framework development

Instructions for working on the framework itself. Not shipped; team repos get
`packages/content/instructions.md`.

## Layout

A pnpm workspace ([ADR 0003](docs/adr/0003-repo-layout.md)). Each package owns its source, tests and build.

| Path | Holds |
|---|---|
| `packages/cli/` | `@folkways/cli`, published to npm (spec §3). Team repos pin an exact version |
| `packages/cli/src/` | TypeScript source of the `folkways` CLI. `tsc` compiles it to `dist/` on publish |
| `packages/content/` | `@folkways/content` (private): the framework's skills, instructions, prompts, settings fragment and workflows. The CLI and the plugin both distribute it |
| `packages/cli/assets/` | Generated at pack time from `packages/content/`; gitignored. Not edited |
| `packages/cli/templates/` | One-time scaffold for `init`. `dot-x` becomes `.x`; `name.tmpl` becomes `name` |
| `packages/cli/test/` | vitest suites for the CLI |
| `packages/hooks/` | `@folkways/hooks` (private): the `folkways-hook` binary, shipped as `@folkways/hooks-<os>-<cpu>` (ADR 0002) |
| `packages/hooks/contract/` | Black-box hook contract suite: fixtures, harness, rules |
| `packages/claude-plugin/` | `@folkways/claude-plugin`: builds the Claude Code plugin into `dist/` from `packages/content/` and the hook binaries. The root `.claude-plugin/marketplace.json` installs it from npm |
| `packages/evals/` | `@folkways/evals` (private, never published): behavior-eval runner. Builds a team repo per arm, runs `claude -p`, grades blind |
| `scripts/` | Cross-package scripts: `smoke.ts` (packed-tarball smoke test), `release.ts` (builds and checks a tag's tarballs), `pack.ts` (shared by both) |
| `research/` | Benchmarks and evals behind decisions, with their results. Not in the workspace; CI doesn't run it |
| `docs/design/` | Design record. `00-STATUS.md` is the entry point; `06-v1-spec.md` is the spec |
| `docs/adr/` | Framework decisions |

## Commands

From the repo root:

| Command | Does |
|---|---|
| `pnpm install` | install dev dependencies (pnpm version pinned in `packageManager`) |
| `pnpm test` | every package's tests |
| `pnpm run check` | every package's `tsc --noEmit`, then the root `scripts/` |
| `pnpm run build` | build every package (the CLI's `src/` to `dist/`, also on `pnpm pack` / `pnpm publish`; the hook packages to `packages/hooks/npm/`, needs Go) |
| `pnpm --filter @folkways/cli test` | one package only (same for `check`, `build`, `@folkways/hooks`) |
| `pnpm run smoke` | pack the CLI and the four hook packages, install them with npm and pnpm, drive `init` and `sync`, run the contract suite through the installed binary (needs Go) |
| `pnpm --filter "@folkways/claude-plugin..." run build` | build the hooks, then the Claude Code plugin into `packages/claude-plugin/dist/` (needs Go); dogfood with `claude --plugin-dir packages/claude-plugin/dist` |
| `pnpm --filter @folkways/hooks run build` | cross-compile the `@folkways/hooks-<os>-<cpu>` packages into `packages/hooks/npm/` |
| `cd packages/hooks && go build -o bin/folkways-hook .` | build the hook binary (Go 1.26+; `bin/` is gitignored; add `-ldflags "-X main.version=x.y.z"` to set `--version`) |
| `FOLKWAYS_HOOK_CMD="$PWD/packages/hooks/bin/folkways-hook" FOLKWAYS_HOOKS=git-guard pnpm --filter @folkways/hooks test` | `go vet`, `go test`, then the hook contract suite against the built binary |
| `node packages/cli/src/cli.ts <command>` | run the CLI from source (Node 24 strips types) |
| `node packages/evals/src/cli.ts run research/evals/<suite> --dry-run` | plan a behavior eval; drop `--dry-run` to run it (model calls on your account; see `packages/evals/README.md`) |
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
  rules, trigger eval). `packages/content/test/skills.test.ts` checks the mechanical rules.
- **Commits:** Angular conventional commits, module scope (`cli`, `content`, `templates`, `skills`, `hooks`, `evals`,
  `docs`, `test`, `build`, `research`). Squash PRs.
- **Tests live with their package:** TypeScript in `packages/<pkg>/test/` (or `contract/`), Go as `*_test.go`
  beside the code. Cross-package checks go in `scripts/`. A package never imports another's source.
- **Versions are per package** (changesets, ADR 0003). A PR that changes a package adds a changeset
  (`pnpm changeset`); docs, research, CI-only and `packages/evals` changes don't need one. Release: merge the
  `changeset version` PR, then `pnpm changeset tag` and push the tags; each tag publishes its package
  (`.github/workflows/release.yml`, after approval). A CLI release fails until the hooks version it
  pins is on npm: approve the hooks publish first, then re-run the CLI's release.
- **The CLI's `optionalDependencies`** are added at pack time by `.pnpmfile.mjs`, never committed. A new hook
  platform goes in `TARGETS` (`packages/hooks/scripts/build.ts`) and `PLATFORMS` (`.pnpmfile.mjs`); smoke checks they match.
- **The Claude Code plugin is generated.** Its skills, instructions and `hooks.json` come from `packages/content/`
  at build time; edit them there. A content change that should reach team repos needs a `@folkways/cli`
  changeset, and one that should reach plugin users a `@folkways/claude-plugin` changeset too.
- **The CLI pins one hooks version** in `folkways.hooks` of `packages/cli/package.json`. Bump it to ship a
  hooks release to team repos. Hook renames, removals and event changes are major hooks bumps.
- **CI** runs only the jobs for the packages a change touches; the `ci` job is the required check. A new
  package gets a filter and jobs in `.github/workflows/ci.yml`.
- Keep `docs/design/00-STATUS.md` current at milestones.
