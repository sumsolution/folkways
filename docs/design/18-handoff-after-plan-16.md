# Handoff: after plan 16, the brain-repo push gate, `doctor`, then release prep

**Written:** 2026-10-02 (end of session 9) · **For:** the next session · **Index:** [00-STATUS](00-STATUS.md)

## Where we are (verified)

| What | State |
|---|---|
| `main` | `b50d214` `feat: folkways.yaml, guided setup and solo brain repos (#25)`; CI green on all 12 jobs at the merged head |
| Open PRs | none from this work (#26 closed as superseded, #27 merged into #25) |
| Plan | [16](16-plan-config-and-solo.md), merged, with decision 6 (below) |
| Tests | CLI 379, plugin 27, evals 26, hooks 22 plus 152 contract cases (all pass against the built binary) |
| Reviews of #25 | four rounds; the last (fresh agent over #25 and #27) clean, its six non-blocking findings fixed |

**Decision 6 (Summer, session 9):** a solo repo is a **brain repo**, like a team repo. Sessions start there;
knowledge, state and efforts build up across projects; it holds no project code. Code changes happen in
worktrees of project repos under `projects_dir`, which is required in solo too. This removed `git-guard`'s
walk-back, the solo exceptions in `commit` and `pr`, and `setup`'s root-collision check. In a repo with no
`folkways.yaml` the one-time offer points to the brain repo; solo `setup` creates one (new or empty folder,
never a code repo). Spec §2.3, §4.2, §7.1, §7.3, §8.5.

## Next

### 1. The brain-repo push gate (hooks, one small PR)

With `review.trigger: pre-push`, `git-guard` denies every direct push of the brain repo itself, team or
solo: no review covers a commit that only changes `state/`, `work/` or `inbox/`. Under decision 6 every
session starts there, so this blocks normal use.

- **Fix:** the push gate skips the repo holding `folkways.yaml` (top level = `cfg.Dir`, symlinks resolved,
  as `sameDir` in `checkMainClone` already does). Knowledge and overlays in a team repo go by PR, where
  people review them.
- **Contract first:** spec §7.1 (`git-guard` row) and §7.3, then the contract README, then fixtures.
  Fixtures: flip `team-push-review-then-commit-denied` to a pass (it pushes `${TEAM}`) and add a solo
  one. It's the only push fixture that pushes the brain repo; the rest push `${WORKTREE}` or don't resolve.
- **Validate:** new fixtures fail on the current binary, pass after. Then the usual checks (below).
- Changeset: `@folkways/hooks` patch; bump nothing else.

### 2. `folkways doctor` (built, not a stub), with the cleanup PR

TDD, one PR. Checks, from 00-STATUS follow-ups:
- non-boolean `hooks:` values (yaml.v3 reads `off`, `no` and `False` as false; quoted `"false"`, `falsey`
  and `0` leave the hook on)
- stale overlay compiles (`staleOverlays`)
- personal overlay lint as errors
- shadowing skills (ADR 0001 M10), `disableAllHooks` (M11)
- `node_modules` left on the new version after a failed `upgrade`
- drift between `npx` and the pinned install
- hook binary present, and pnpm `publicHoistPattern`

Cleanup that can ride along: add `scripts/` to the tsconfig `include`; gitignore `.folkways-sync-*` in
the team template; the `instructions.md` Git line without a solo exception for the review (00-STATUS);
a TruffleHog CI job pinned to a SHA.

### 3. Then, in order

1. **`folkways review` / `run`:** fill the placeholders in `assets/prompts/review.md` and run the
   fresh-process review headless. Solo keeps the fresh-subagent path (no CLI).
2. **Evals:** a trigger eval for the new `setup` description ("set up folkways", an unconfigured repo, a
   code repo); overlay precedence ported to `packages/evals`.
3. **Summer's items:** npm trusted publishers and the `npm` environment; the employer open-source check;
   the trademark check; plan 11 step 6 on real machines.
4. **Publish:** `@folkways/hooks`, then `@folkways/cli` (bump its pinned `folkways.hooks`), then
   `@folkways/claude-plugin`. Then the beta team repo.

## How to work

- **Branch** from a fresh `main`; one PR per step. Squash-merge with the `(#NN)` title. The repo refuses
  merge commits.
- **Validate before each push:**
  - `pnpm run check && pnpm test`
  - `cd packages/hooks && go vet ./... && go test ./... && go build -o bin/folkways-hook .`
  - `FOLKWAYS_HOOK_CMD="$PWD/packages/hooks/bin/folkways-hook" pnpm --filter @folkways/hooks exec vitest run`
  - plugin changes: `pnpm --filter "@folkways/claude-plugin..." run build`, then drive
    `packages/claude-plugin/dist/hooks/run.sh` in scratch repos with `XDG_CONFIG_HOME` in a scratch dir.
- **Smoke runs only in CI** (`pnpm run smoke` fails in the cloud sandbox at "hook binary is linked", on
  `main` too).
- **Review:** a fresh agent before merge, with the contract sections and the intentional behaviors in its
  brief (below). Merging needs a clean review; the session's auto mode blocks merging unreviewed work.
- **CI runs only on PRs into `main`.** A PR into a feature branch gets no CI; validate locally.

## Carried over

- **Intentional, tell reviewers:** no `team.yaml` fallback; the offer is keyed per git top level (so per
  worktree); toggles fail toward on; `run.sh`'s `session-start: false` line test is approximate (it
  disagrees with yaml.v3 on `False`, `off`, `no`, quoted keys, and keys ending in `session-start`).
- **Two sessions on one PR:** in session 9 another session pushed fixes to #25's branch while this one
  opened #26 for the same findings, and #26 was closed as superseded. Fetch the PR head before fixing
  findings, and reply on threads only after checking what's already there.

## Resume prompt

`Resume docs/design/18-handoff-after-plan-16.md: read it and 00-STATUS.md, then start with "Next" step 1 (the brain-repo push gate).`
