# Plan: move to a pnpm workspace

**Written:** 2026-09-29 (session 5) · **Status:** not started · **Decision:** [ADR 0003](../adr/0003-repo-layout.md)

## Goal

Move the repo to the layout in [ADR 0003](../adr/0003-repo-layout.md): `packages/cli`, `packages/hooks`,
`research/`, independent versions with changesets, path-filtered CI. Done when `pnpm run check`,
`pnpm test` (CLI, Go, contract) and `pnpm run smoke` pass from the root, and `pnpm pack` in `packages/cli`
produces the same file list as before.

## Starting point

| Have | Where |
|---|---|
| CLI package at the root: `package.json`, `src/`, `test/`, `assets/`, `templates/`, `tsconfig*.json`, `vitest.config.ts` | root |
| Go hooks module with `git-guard` (hooks PR 2) | `hooks/`, module `github.com/sumsolution/folkways/hooks` |
| Hook contract suite: fixtures, harness, runner | [test/hooks/](../../test/hooks) |
| Research code | [bench/hooks/](../../bench/hooks), [07-eval/](07-eval) |
| Tarball smoke test | [scripts/smoke.ts](../../scripts/smoke.ts) |
| CI: `check`, `test`, `smoke`, `hooks` jobs, no filters | [.github/workflows/ci.yml](../../.github/workflows/ci.yml) |
| Dependabot: npm and github-actions, no gomod | [.github/dependabot.yml](../../.github/dependabot.yml) |
| One `CHANGELOG.md`, one version (0.1.0, unpublished) | root |

## Preconditions

- Hooks PR 2 (`feat/hooks-go-git-guard`) is merged.
- The hooks agent pauses until this PR merges, then continues PRs 3–5 on the new paths.

## PR A: `build: move to a pnpm workspace`

One PR. Moves use `git mv` so history follows; the move and the config are separate commits inside the PR
to keep review readable (it's squashed on merge).

1. **CLI.** `git mv src test assets templates tsconfig.build.json vitest.config.ts packages/cli/`, and the
   root `package.json` to `packages/cli/package.json`. New private root `package.json`: `packageManager`
   pin, `engines`, and scripts `check` / `test` / `build` running `pnpm -r`, `smoke` running
   `scripts/smoke.ts`. `tsconfig.base.json` gets today's compilerOptions; each package's `tsconfig.json`
   extends it. `pnpm-workspace.yaml` gains `packages: [packages/*]`; its supply-chain settings stay as
   they are. Add `packages/cli/README.md` (short npm page).
2. **Hooks.** `git mv hooks packages/hooks`. Module path becomes
   `github.com/sumsolution/folkways/packages/hooks`; update imports. Root `go.work` with
   `use ./packages/hooks`.
3. **Contract.** `git mv test/hooks packages/hooks/contract` (from its new place under `packages/cli/test/`
   after step 1, or before step 1). `packages/hooks/package.json` (`@folkways/hooks`, private, version
   `0.1.0`): `build` = `go build -o bin/folkways-hook .`, `test` = `go vet ./... && go test ./...` then
   vitest on `contract/` with `FOLKWAYS_HOOK_CMD` set to the built binary. Its own `vitest.config.ts`;
   the CLI's no longer includes the contract.
4. **Research.** `git mv bench research/bench` and `git mv docs/design/07-eval research/evals/overlay-enforcement`.
   Fix relative paths inside `research/bench/hooks/run.ts` and the eval's `run.ts` / `score.ts`. Research
   isn't in the workspace and CI doesn't run it.
5. **Smoke.** `scripts/smoke.ts` packs from `packages/cli`. (Hooks PR 5 adds the platform package.)
6. **Versioning.** Add `@changesets/cli` (dev, root; mind the 7-day rule). `.changeset/config.json`:
   independent, `baseBranch: main`, `privatePackages: { version: true, tag: true }`, `access: public`.
   Move the 0.1.0 entry to `packages/cli/CHANGELOG.md`; start `packages/hooks/CHANGELOG.md`; the root
   `CHANGELOG.md` points to both. Add `"folkways": { "hooks": "0.1.0" }` to `packages/cli/package.json`
   and a test that it equals a published version or the current `packages/hooks` version.
7. **CI.** Rewrite `ci.yml`: `changes` job (`dorny/paths-filter`, pinned SHA, filters from ADR 0003),
   jobs `cli` (check, build, test), `hooks` (setup-go from `packages/hooks/go.mod`, vet, test, contract),
   `smoke` (on `cli || hooks || shared`), and the aggregate `ci` job. Pushes to `main` run every job.
   Dependabot: add `gomod` for `/packages/hooks`. Summer switches the required check to `ci`.
8. **Docs.** `AGENTS.md` layout, commands and releases (add a changeset per changed package; release =
   merge the version PR, push the tags); `research` scope. Spec §2.1 tree (drops the stale
   `src/hooks/*.ts`). [11](11-plan-hooks-build.md) paths and decision 3. [00-STATUS.md](00-STATUS.md)
   Built table. Every link to a moved file: grep `test/hooks|bench/|07-eval|src/|assets/|templates/` in
   `docs/`, `README.md`, ADRs and the READMEs inside moved folders.

## Verify

- `pnpm install --frozen-lockfile`, `pnpm run check`, `pnpm test` (CLI + Go + contract), `pnpm run smoke`.
- `pnpm pack --dry-run` in `packages/cli` lists `dist`, `assets`, `templates`, `package.json`, `README.md`
  and nothing else (compare with the file list on `main`).
- `pnpm changeset status` works; a test changeset bumps only its package.
- CI on the PR: a docs-only commit runs only `ci`; a change in `packages/hooks` runs `hooks` and `smoke`,
  not `cli`.
- A link check over `docs/**/*.md` and `README.md`: every relative link resolves.
- `go build ./...` from the root works through `go.work`.

## Risks

- **Link rot.** Many docs link to moved paths. The link check in Verify catches them.
- **The hooks agent's branch.** Landing between hooks PR 2 and PR 3 keeps the conflict to zero open branches.
- **Aggregate check logic.** A wrong `if:` makes `ci` pass when a job failed. Test it by pushing a failing
  commit to the PR once.
- **Publishing from a subfolder.** `pnpm publish` must run in `packages/cli`; the release workflow (hooks
  PR 5) is written for that.

## Out of scope

The release workflow itself (hooks PR 5), the config JSON Schema, `src/commands/`, extracting content to its
own package, the sync Action and the plugin (ADR 0003, Future direction).
