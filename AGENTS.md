# folkways: framework development

Instructions for working on the framework itself. Not shipped; team repos get
`packages/content/instructions.md`.

## Layout

A pnpm workspace ([ADR 0003](docs/adr/0003-repo-layout.md)). Each package owns its source, tests and build.

| Path | Holds |
|---|---|
| `packages/content/` | `@folkways/content` (private): the framework's skills, instructions, prompts, settings fragment, workflows and init templates. `content.go` embeds it in the `folkways` binary; the plugin distributes it too |
| `packages/content/templates/` | One-time scaffold for `init`. `dot-x` becomes `.x`; `name.tmpl` becomes `name` |
| `go.mod`, `packages/folkways/` | The Go module (repo root, `github.com/sumsolution/folkways`). `packages/folkways` is `@folkways/hooks` (private); its `package.json` holds the CLI's version. `cmd/folkways` is the CLI (ADR 0011), with `folkways hook <name>`; `cmd/folkways-hook` is the hook binary the plugin bundles (ADR 0002) until the plugin is removed |
| `packages/folkways/contract/` | Black-box hook contract suite: fixtures, rules, and the Go runner (`*_test.go`, run by `go test`) |
| `packages/folkways/clicontract/` | Black-box CLI contract suite (Go, `go test`): fixtures pinning the Go CLI (ADR 0011) |
| `packages/claude-plugin/` | `@folkways/claude-plugin`: builds the Claude Code plugin into `dist/` from `packages/content/` and the hook binaries. The root `.claude-plugin/marketplace.json` installs it from npm |
| `packages/evals/` | `@folkways/evals` (private, never published): behavior-eval runner. Builds a team repo per arm, runs `claude -p`, grades blind |
| `docs/adr/` | Framework decisions |

## Commands

From the repo root:

| Command | Does |
|---|---|
| `pnpm install` | install dev dependencies (pnpm version pinned in `packageManager`) |
| `pnpm test` | every package's tests |
| `pnpm run check` | every package's `tsc --noEmit` |
| `pnpm run build` | build every package (the hook packages to `packages/folkways/npm/` and the plugin, needs Go) |
| `pnpm --filter @folkways/evals test` | one package only (same for `check`, `build`, `@folkways/hooks`) |
| `pnpm --filter "@folkways/claude-plugin..." run build` | build the hooks, then the Claude Code plugin into `packages/claude-plugin/dist/` (needs Go); dogfood with `claude --plugin-dir packages/claude-plugin/dist` |
| `pnpm --filter @folkways/hooks run build` | cross-compile the `@folkways/hooks-<os>-<cpu>` packages into `packages/folkways/npm/` |
| `cd packages/folkways && go build -o bin/folkways ./cmd/folkways` | build the CLI (`./cmd/folkways-hook` for the plugin's hook binary) (Go 1.26+; `bin/` is gitignored; add `-ldflags "-X main.version=x.y.z"` to set `--version`) |
| `go vet ./... && go test ./...` | Go tests (from the root, includes the `packages/content` embed) and the hook contract suite; the suite builds `./cmd/folkways` and runs `folkways hook` |
| `cd packages/folkways && FOLKWAYS_HOOK_CMD="$PWD/bin/folkways-hook" FOLKWAYS_HOOKS=git-guard go test ./contract` | the contract suite against any build (`FOLKWAYS_HOOK_CMD="$PWD/bin/folkways hook"` works too); `FOLKWAYS_HOOK_BUDGET_MS=10` adds the hook latency check |
| `pnpm --filter @folkways/hooks test` | `go vet`, `go test`, then the vitest suite for the build script |
| `cd packages/folkways && go test -count=1 ./clicontract/` | the CLI contract suite against a fresh Go build; set `FOLKWAYS_CLI_CMD` to run it against another build |
| `go run ./packages/folkways/cmd/folkways <command>` | run the CLI from source |
| `node packages/evals/src/cli.ts run packages/evals/suites/<suite> --dry-run` | plan a behavior eval; drop `--dry-run` to run it (model calls on your account; needs Go; see `packages/evals/README.md`) |
| `pnpm changeset` | record a change for the next release of the packages it touches |

## Rules

- **No company-specific content.** Names, portals, project keys, internal URLs and paths belong in team
  repos. Examples use synthetic values (`Example Team`, `EX`, `example.com`).
- **TypeScript** (evals, content tests, the plugin and hook build scripts): Node 24+, ESM. Only erasable syntax
  (no `enum`, `namespace`, parameter properties); import relative files with `.ts` extensions; `import type` for types. `tsconfig.base.json` enforces these.
- **Dependencies:** runtime deps are fine when they earn their place; they install into team repos.
  `pnpm-workspace.yaml` blocks versions younger than 7 days, provenance downgrades, non-registry
  subdependencies and unapproved build scripts. Don't relax these to make an install pass; wait or ask.
- **Package manager:** pnpm for this repo. Team repos need no Node: the `folkways` binary on `PATH` (ADR 0012).
- **CI actions** are pinned to full commit SHAs with a version comment (the repo requires it).
- **TDD:** write a failing test first, then fix. Run `pnpm run check && pnpm test` before every commit.
- **Expected failures** return `clierr.New(...)`; the CLI prints `folkways: <message>` with no stack trace.
- **Skills:** follow [docs/skill-authoring.md](docs/skill-authoring.md) (skeleton, overlay IDs, headless
  rules, trigger eval). `packages/content/test/skills.test.ts` checks the mechanical rules.
- **Commits:** Angular conventional commits, module scope (`cli`, `content`, `templates`, `skills`, `hooks`, `evals`,
  `docs`, `test`, `build`, `research`). Squash PRs.
- **Tests live with their package:** TypeScript in `packages/<pkg>/test/`, Go as `*_test.go`
  beside the code. A package never imports another's source.
- **Versions are per package** (changesets, ADR 0003). A PR that changes a package adds a changeset
  (`pnpm changeset`); docs, research, CI-only and `packages/evals` changes don't need one. No release pipeline
  until the Go release lands (go-cli Phase 2); don't run `changeset tag`.
- **The Claude Code plugin is generated.** Its skills, instructions and `hooks.json` come from `packages/content/`
  at build time; edit them there. A content change that should reach team repos needs a `@folkways/hooks`
  changeset (the binary embeds the content), and one that should reach plugin users a `@folkways/claude-plugin`
  changeset too.
- **Hook renames, removals and event changes** are major bumps of `@folkways/hooks`.
- **CI** runs only the jobs for the packages a change touches; the `ci` job is the required check. A new
  package gets a filter and jobs in `.github/workflows/ci.yml`.
- **Design record, plans, research and status** live in the team repo, [sumsolution/folkways-command](https://github.com/sumsolution/folkways-command)
  (private): open work in `efforts/`, closed work indexed in `state/closed-efforts.md` ([ADR 0005](docs/adr/0005-closing-efforts.md)),
  what's built in `state/framework.md`.
