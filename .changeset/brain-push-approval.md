---
"@folkways/hooks": minor
"@folkways/claude-plugin": minor
---

Brain repo sessions push without asking by default. `folkways.yaml` `push.brain` and `push.app` (`always`, `default-branch`, `never`; defaults `never` and `always`) set when a push asks, and `.local/me.yaml` may only tighten them. git-guard asks per repo, denies a direct push of `knowledge/` to the default branch, and treats the brain's worktrees as the brain. `sync` writes the `git push` ask rule only for `push.brain: always`. Sessions in the brain repo work in their own worktree and land with a rebase and `git push origin HEAD:<default>`. See ADR 0014.
