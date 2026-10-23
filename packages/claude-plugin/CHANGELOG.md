# @folkways/claude-plugin

## 0.2.0

### Minor Changes

- 6e7594a: Skills no longer read config files or attach personal overlays in a step 0: they take their config from the `session-start` hook's context. Solo repos have no overlays for now (`overlays/` is ignored).
- b50d214: Quiet in repos without `folkways.yaml`: no instructions and no hooks, only a one-time offer that points to the brain repo, where sessions start, and to `folkways:setup`, which creates a solo one or sets up a team one.
- 834b43a: New package: the folkways Claude Code plugin. Framework skills (as `folkways:<name>`), the session instructions and the hooks, with `folkways-hook` for darwin and linux, arm64 and x64. Install it with `claude plugin marketplace add sumsolution/folkways` and `claude plugin install folkways@folkways`.

### Patch Changes

- 399696d: `onboard`, `setup`, `commit` and `pr` call `folkways` on `PATH` instead of `node_modules/.bin/folkways` or `npx`.
- ea5cf95: The `adr` skill no longer cites the v1 spec, which moved out of the package repo.
- 90fcf78: The instructions say how a solo repo, which has no CLI, gets the review that must cover the SHA: the `review` skill in a fresh subagent.
