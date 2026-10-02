# Handoff: review #25, address findings, then next steps

**Written:** 2026-10-02 (end of session 8) · **For:** the next session · **Index:** [00-STATUS](00-STATUS.md)

## Where we are (verified)

| What | State |
|---|---|
| PR | [#25](https://github.com/sumsolution/folkways/pull/25) `feat: folkways.yaml, guided setup and solo repos (plan 16)`, open, mergeable, no review threads |
| Branch | `ccr-d5e13116-v42x04`, HEAD `2613d61`, pushed, working tree clean |
| CI on `2613d61` | all 12 jobs green, smoke included (linux x64, linux arm64, macOS) |
| Plan | [16](16-plan-config-and-solo.md), W0–W5 done |
| Local checks at HEAD | `pnpm run check`, `pnpm test` (CLI 368, plugin 25, evals 26, hooks 22), `go vet`, `go test`, contract suite 178/178 against the built binary |
| Already reviewed | One fresh-process review, before the hooks fix (`4dcb34c`) and the smoke fix (`2613d61`). 7 findings: 6 fixed, 1 deferred ([lenient `hooks:` parsing](#carried-over)) |
| Review round 2 | Summer's review on #25 (4 threads): solo review gate in `pr` and `commit`, `commit` step 1 in a solo repo, `run.sh` no-binary notice ignoring `session-start: false`, a walk-back test that passed without its guard. All fixed and replied to; [step 1](#1-kick-off-code-review-on-25) is for anything further |
| Round 3 (session 9) | A second fresh review was not clean (`commit` re-reviewed review-only commits in solo). Summer reframed solo: a solo repo is a **brain repo** with no code (plan 16, decision 6), which removed the walk-back and the solo special cases instead of patching them |

The branch carries a merge of `main` (`d41e0fe`): #24 was squash-merged mid-session and a force-push to
rebase was refused, so squash-merge #25.

## Next

### 1. Kick off code review on #25

A fresh process, so the reviewer can't be framed by this session's context. `folkways review` is still a
stub, so run one of these from a fresh session on the branch:

- **Preferred:** `/code-review 25 high --comment`. It posts findings as inline PR comments, so they're
  tracked as threads.
- **Or:** a fresh agent with the brief below, reporting to the session (no PR comments).

> Review `git diff origin/main...HEAD -- packages/ scripts/` in sumsolution/folkways (read `AGENTS.md`).
> Contract: `docs/design/06-v1-spec.md` §2.3, §4.1, §4.2, §7.1, §7.3, §8.5; plan 16. Intentional, don't
> flag: no `team.yaml` fallback; offer keyed per git top level (so per worktree); toggles fail toward on.
> Look for correctness bugs (Go and shell edge cases, the offer file, gating in `main.go`, the solo
> walk-back in `gitguard/guard.go`, symlinked paths), contract mismatches between spec, fixtures, Go,
> `run.sh`, `scripts/smoke.ts` and skill text, contradictory skill instructions (step 0, solo vs team
> paths, `setup`/`start`/`onboard`), and tests that don't test what they claim. Verify each finding.
> Report Conventional Comments, most severe first: label (blocking or not), file:line, failure scenario,
> fix. Say what you checked and found fine.

Focus the review on what the first one didn't see: `walkBack`, `sameDir` and `isFolkwaysRepo` in
`packages/hooks/internal/gitguard/guard.go`, the subfolder check in `sessionstart.go`, the
`env.commits` harness in `packages/hooks/contract/harness.ts`, and `scripts/smoke.ts`.

### 2. Address findings

- **Triage.** Verify each finding before acting. Blocking → fix. Non-blocking → fix if small and in
  scope, else one line in [00-STATUS](00-STATUS.md) follow-ups. Reply on every PR thread, fixed or not,
  and resolve the ones you fixed.
- **Contract first.** If a fix changes behavior, change the spec (`06-v1-spec.md`) first, then the fixture,
  then the code.
- **Parallel fixes.** Split by package, as in plan 16: hooks, `packages/cli/assets` with
  `test/skills.test.ts`, the rest of `packages/cli`, `packages/claude-plugin`, `scripts/`. Agents in
  worktrees start on an old `main`. Every brief must say: first
  `git fetch origin ccr-d5e13116-v42x04 && git reset --hard origin/ccr-d5e13116-v42x04`. Then cherry-pick
  their commits onto the branch.
- **Step back** (instructions, "Step back") if findings land in code the hooks fix (`4dcb34c`) added, or a fix
  needs a new layer of machinery: show the numbers and offer "do less".
- **Validate before each push:**
  - `pnpm run check && pnpm test`
  - `cd packages/hooks && go vet ./... && go test ./... && go build -o bin/folkways-hook .`
  - `FOLKWAYS_HOOK_CMD="$PWD/packages/hooks/bin/folkways-hook" pnpm --filter @folkways/hooks exec vitest run`
  - For a hooks change, run new fixtures against the binary before the fix (`go build` at the previous
    HEAD) and confirm they fail.
  - For a plugin change, `pnpm --filter "@folkways/claude-plugin..." run build`, then drive
    `packages/claude-plugin/dist/hooks/run.sh` in three scratch repos (synced, configured, unconfigured)
    with `XDG_CONFIG_HOME` in a scratch dir.
- **Smoke runs only in CI.** `pnpm run smoke` fails in the cloud sandbox at "hook binary is linked after
  npm install", on `main` too. Not this PR's.
- **Records.** A behavior change updates its changeset in `.changeset/` (`hooks-config-gating.md`,
  `cli-folkways-yaml.md`, `claude-plugin-quiet.md`) and the PR body.
- **Merge.** Summer approves; squash with the `(#25)` title.

### 3. After merge

1. **Status.** Mark plan 16 merged in [00-STATUS](00-STATUS.md) and [16](16-plan-config-and-solo.md).
   Point "Next step" back at the list below.
2. **`folkways doctor`** (built, not a stub). Checks already listed in 00-STATUS follow-ups:
   - non-boolean `hooks:` values
   - stale overlay compiles
   - personal overlay lint as errors
   - shadowing skills (ADR 0001 M10)
   - `disableAllHooks` (M11)
   - `node_modules` left on the new version after a failed `upgrade`
   - drift between `npx` and the pinned install
   - hook binary and pnpm `publicHoistPattern`
   
   TDD, one PR.
3. **A cleanup PR,** which can ride with `doctor`:
   - add `scripts/` to the tsconfig `include` (`scripts/smoke.ts` is untypechecked today)
   - gitignore `.folkways-sync-*` in the team template
   - a TruffleHog CI job pinned to a SHA
4. **`folkways review` / `run`.** Fill the placeholders in `assets/prompts/review.md` and run the
   fresh-process review headless. This would replace step 1's manual review.
5. **Evals.**
   - A trigger eval for the new `setup` description: does "set up folkways" or an unconfigured repo trigger it?
   - Overlay precedence ported to `packages/evals`.
6. **Summer's items:**
   - npm trusted publishers and the `npm` environment
   - the employer open-source check
   - the trademark check
   - plan 11 step 6 on real machines
7. **Publish, then the beta team repo.** Publish `@folkways/hooks`, then `@folkways/cli` (bump its pinned
   `folkways.hooks`), then `@folkways/claude-plugin`.

## Carried over {#carried-over}

- **Lenient toggles:** `hooks:` parsing: yaml.v3 reads `no` as off, but `off` or `"false"` leave the hook
  on. This fails toward on, which is safe. Goes into `doctor`.
- **Offer per worktree:** the offer is keyed per git top level, so each worktree of a repo gets its own
  one-time offer. Accepted.
- **Session 8's own scheduling:** the PR subscription and a fallback check-in (`trig_01UTZ3jvtQpgd9PpXHYs8ATZ`)
  belong to session 8. A new session doesn't inherit them. Subscribe again if it should watch #25.
- **Tried and refused:** a force-push to rebase after #24 merged was refused. Merge `main` into the branch
  instead.

## Resume prompt

`Resume docs/design/17-handoff-pr25-review.md: read it and 00-STATUS.md, then start with "Next" step 1 (review #25).`
