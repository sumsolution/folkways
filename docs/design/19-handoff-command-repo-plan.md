# Handoff: build folkways with folkways (content package, simpler skills, command repo)

**Written:** 2026-10-03 (session 10) · **For:** the next session · **Index:** [00-STATUS](00-STATUS.md)

This file is untracked on `main`. Commit it with PR 2's doc changes, or move it in step 4.

## Decided this session (Summer)

| Topic | Decision |
|---|---|
| Dogfooding | Build folkways from a brain repo, `~/projects/folkways-command` (`git@github.com:sumsolution/folkways-command.git`, no commits yet). Team mode, one member (`summer`), CLI from local tarballs |
| Repo split | Deferred until after dogfooding. If it happens, the seam is content vs machinery, not one repo per package |
| Coupling | Content moves out of the CLI package (PR 1). One config reader |
| CLI language | Go long term, as its own effort (`go-cli`). Until then **no new TypeScript in the CLI, only deletions** |
| Personal overlays | Removed (PR 2) |
| Config in skills | The `session-start` hook prints `folkways.yaml` and `.local/me.yaml` once per session. Baking values into skills at compile time waits for the Go port |
| `docs/design/` | All of it moves to the command repo, spec included. ADR links point at `https://github.com/sumsolution/folkways-command`; cites in `AGENTS.md` and code comments are removed |
| Tracking | New work runs as separate efforts in the command repo, not one growing `00-STATUS.md` |

Non-goals: splitting the workspace, porting the CLI, config baking, a TypeScript YAML parser, solo-mode CLI
support, publishing to npm.

## The plan

Five steps: PR 1 content package, PR 2 skills, step 3 command repo, step 4 design record and efforts, PR 3
removes `docs/design/` here. The full approved plan, with per-agent scopes and verification, is at
`~/.claude/plans/i-am-considering-a-iterative-metcalfe.md`. Read it before acting.

## Where we are

| What | State |
|---|---|
| `main` | `bde2df8` `build: move framework content to packages/content (#32)`. Check and tests pass on it (content 179, cli 236, plugin 27, evals 27, hooks 22) |
| PR 1 | **Merged** as [sumsolution/folkways#32](https://github.com/sumsolution/folkways/pull/32). Worktree and branch removed |
| PR 2, step 3, step 4, PR 3 | Not started |

What PR 1 left that PR 2 builds on:

- Content is `packages/content/` (private `@folkways/content`, ignored by changesets). Commit scope `content`
  replaces `assets`.
- `assetsDir()` in `packages/cli/src/package.ts` prefers the sibling `packages/content`, else the packed
  `assets/` (gitignored, written by `prepack`).
- The skills test is `packages/content/test/skills.test.ts`; its `sync` cases are
  `packages/cli/test/content-sync.test.ts`.
- A content change runs the content, CLI, evals, plugin and smoke jobs in CI.

## Next

1. **PR 2, `feat(skills): drop personal overlays and step 0 file reads`,** per the plan, on a branch from a
   fresh `main`. Paths in the plan that say `packages/cli/assets` are now `packages/content`.
2. **Coordinator first (contract):** spec §3.2, §4.2, §7.1, `packages/hooks/contract/README.md`, the ADR 0001
   amendment, changesets (`hooks`, `cli`, `claude-plugin` minor). Commit this file with those doc changes.
3. **Three agents, one worktree each off the PR branch:** hooks (session-start prints config; fixtures
   first), content (16 skills, `instructions.md`, authoring guide), CLI deletions (personal overlay code,
   `.local/me` link, tests, templates).
4. **Integrate, validate, fresh Opus review, scan.** Ask Summer before the push and before the squash merge.
   Then check and test on `main`; remove worktrees and branches.
5. **Step 3** (command repo), then **step 4** and **PR 3** (design record), per the plan.

## How to work

- Summer approves every push and every merge. Squash with the `(#NN)` title. Never rebase or force-push.
- One worktree per agent, separate folders. Sonnet first; a fresh Opus agent reviews each PR.
- Briefs carry the goal, the non-goals, "report lines added and removed", and "push back if a fix needs
  far more code than the problem is worth". They forbid serializing env or local paths.
- Step back and show the numbers if a PR outgrows the plan.

## Accepted gaps

- Solo repos lose overlays until the Go port brings compile to the plugin.
- The command repo installs from gitignored tarballs in `vendor/`, so a fresh clone can't install until
  0.1.0 is on npm.
- ADR links to the command repo resolve only once it is pushed and readable.

## Resume prompt

`Resume docs/design/19-handoff-command-repo-plan.md: read it and ~/.claude/plans/i-am-considering-a-iterative-metcalfe.md, then start with "Next" step 1 (PR 2, the skills simplification).`
