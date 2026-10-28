# 0014. Brain repo sessions push without asking; `folkways.yaml` sets when a push asks

| | |
|---|---|
| Status | accepted (Summer Jordan, 2026-10-27) |
| Date | 2026-10-27 |
| Owner | Summer Jordan |
| Reviewers | Summer Jordan (one-person team, `adr.reviewers.min: 0`) |
| Review window | until Summer's review of the PR that adds this record, done 2026-10-27 |
| Amends | [0001](0001-overlay-enforcement.md) M2 ("Push needs the user's approval") |

## Issue

Every push waits for the person's approval: `sync` writes the permission rule `ask: Bash(git push:*)` into the
brain repo's `.claude/settings.json` ([settings.json][c-settings]), and the instructions say "Ask before every
push" ([instructions.md:104][c-instr]). In the brain repo most pushes are status docs, research and trackers
(`efforts/`, `state/`, `people/`) that land straight on the default branch. Approving each one is friction
with no review value, and parallel sessions in one brain clone step on each other's working tree. When
should a push ask, and where does a brain session work?

## Context

- The question was open as the `push-approval-mode` effort in the team repo: ask before every push, never,
  or only for the default branch, configured in `folkways.yaml`. Whether to ask is a personal preference as
  much as a team one; where `knowledge/` lands is the team's.
- `knowledge/` is the one brain path that lands by PR ([instructions.md:27][c-instr]); nothing enforces it.
- A settings `ask` rule prompts whatever a `PreToolUse` hook returns, `allow` included
  ([permissions: precedence][cc-perm]), so a static rule can't be loosened per person.
  The architecture review planned a git-guard rule for it (team repo `architecture-review` Next 2).
- A permission rule is all or nothing per command prefix. It can't tell a brain push from an app repo push
  run in the same session, or a push to `main` from a push to a branch. git-guard already parses every
  `git push` and knows its directory ([guard.go][c-guard]).
- Claude Code's `PreToolUse` hooks may return `permissionDecision: "ask"`. Copilot CLI documents `deny`
  only (team repo `knowledge/claude-code-vs-copilot-mechanisms.md`); whether it honors `ask` is unverified.
- git-guard exempted the brain repo from the review gate by comparing git top levels, so a worktree of the
  brain was gated like an app repo.

## Decision

1. **Config.** `folkways.yaml` gets `push.brain` and `push.app`, each `always`, `default-branch` or `never`.
   Defaults: `brain: never` (`always` in a solo repo, whose landing is the person's choice), `app: always`.
   `.local/me.yaml` `push.brain` / `push.app` override the team's values in either direction. The
   session-start hook prints the effective values.
2. **git-guard asks.** On `git push` it reads the target repo: the brain repo or any worktree of it (same
   git common dir) uses `push.brain`, anything else `push.app`. `always` returns `ask`; `default-branch`
   asks when a destination is the remote's default branch or isn't known. A push in an unknown directory
   asks unless both are `never`. A deny wins over an ask.
3. **knowledge/ by PR, enforced and configurable.** `folkways.yaml` `push.knowledge` is `pr` or `direct`. With
   `pr` (the default; `direct` in a solo repo), a brain push to the default branch whose new commits change
   `knowledge/` is denied. Only `folkways.yaml` sets it: `.local/me.yaml` can't. It is a separate key, not
   folded into `push.brain`, because asking and landing are different questions, and a person may change one
   but not the other.
4. **No static rule.** The framework stops writing `ask: Bash(git push:*)`: it would prompt even when
   `.local/me.yaml` says not to. `sync` doesn't remove it from existing settings, since it can't tell the
   old framework rule from one a team keeps on purpose (say, as a backstop for when hooks are off). A repo
   synced before this change deletes the rule by hand to let `push.*` decide.
5. **Brain sessions work in a worktree.** Each session edits in `<projects_dir>/.worktrees/<brain>/session-<slug>`
   off `origin/<default>`, and lands with `git rebase origin/<default>` and `git push origin HEAD:<default>`.
   The brain's worktrees count as the brain, and the review gate finds reviews in any of them. Hooks,
   config and skills load from the main clone, so `start` fast-forwards it when it is clean and on the
   default branch.

## Options considered

| Option | Trade-off |
|---|---|
| **A. Per repo kind in `folkways.yaml`, enforced by git-guard `ask` (chosen)** | Brain and app pushes get different answers in one session, but asking relies on the hook running and on the agent tool honoring `ask` |
| B. One switch that drops the static rule in the brain | Smallest change, but app pushes from a brain session would then go unasked, with no mechanism behind the instruction |
| C. Keep asking for every push | No change, but the friction stays and parallel sessions still share one working tree |

## Consequences

- Brain pushes need no approval by default. A team that wants the old behavior sets `push.brain: always`; a
  person can still set `never` for themselves.
- Every push ask now comes from git-guard. With hooks off (no `folkways` on `PATH`), or when git-guard fails or
  times out, nothing asks for any push and `knowledge/` isn't enforced: the instructions still say to ask and
  to use a PR.
- Repos synced before this change keep the `Bash(git push:*)` rule, so every push there still asks until
  someone deletes it. A team may keep it on purpose as a hooks-off backstop; then `push.*` can't loosen.
- Pushes in an unknown directory (`cd "$X" && git push`) now ask, where before the static rule asked for
  every push anyway.
- Copilot CLI: if it ignores `ask`, its pushes go unasked where the static rule is gone. To probe.
- ADR 0001 M2 said the push ask held even in bypass mode, which came from the static rule. Whether a hook
  `ask` prompts under `bypassPermissions` is unverified. To probe.
- `default-branch` and the `knowledge/` deny need the remote's default branch: `refs/remotes/<remote>/HEAD`,
  else `master` when the remote has it and no `main`, else `main`. When the remote branch can't be read, the `knowledge/` check passes (fail open).
- The brain is found by the project dir (`CLAUDE_PROJECT_DIR`). An app repo synced with its own
  `folkways.yaml` (ADR 0012) would read as a brain; it needs a repo-kind marker when app repo sync lands.
- Skills changed: `start` creates the session worktree; `commit`, `pr`, `status`, `handoff`, `onboard` follow
  `push.*` instead of always asking.

[c-settings]: https://github.com/sumsolution/folkways/blob/be656e1/packages/content/settings.json
[c-instr]: https://github.com/sumsolution/folkways/blob/be656e1/packages/content/instructions.md
[c-guard]: https://github.com/sumsolution/folkways/blob/be656e1/packages/folkways/internal/gitguard/guard.go
[cc-perm]: https://code.claude.com/docs/en/permissions#precedence-among-rules-hooks-and-mods
