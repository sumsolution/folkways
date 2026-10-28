# 0015. Single-repo efforts and their reviews live in the app repo; cross-repo efforts and notes stay in the team repo

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-28 |
| Owner | Summer Jordan |
| Reviewers | Summer Jordan (one-person team, `adr.reviewers.min: 0`) |
| Review window | until Summer's review of the PR that adds this record |

## Issue

Sessions can now start in an app repo or in the team repo (app-repo-first direction, row 1). Every effort,
review and inbox path the skills write is still relative to the team repo, and the pre-push review gate
looks for reviews only in team-repo checkouts. A session that starts in an app repo then has to reach a second
repo for every status update and every review. Agents on GitHub (the cloud agent and code review) can't see that
repo at all. Where should efforts, reviews and personal notes live when work starts in an app repo?

## Relevant context

- Skill paths are all relative to the team repo: `efforts/<KEY>-<slug>/`, `<personal root>/work/…`,
  `<personal root>/inbox/…` ([start/SKILL.md:59,71][c-start], [review/SKILL.md:91-92][c-review],
  [handoff/SKILL.md:29-31][c-handoff]).
- The gate (`checkPush`) runs `hasReview` and `reviewedSHAs` against brain checkouts only
  ([guard.go][c-guard], [review.go][c-review-go]).
- Reviews were kept outside the code on purpose: a committed review moves HEAD, so it no longer covers HEAD
  (team repo `efforts/app-repo-first/20` §4.4).
- `review.skip_paths` (folkways#78) already lets a review of an ancestor cover HEAD when every path changed
  since then matches a glob. The default is `["docs/**"]` ([team.go:106-111][c-team]). It diffs with
  `--no-renames`, so a rename has to match on both paths.
- A throwaway test with `skip_paths: ["efforts/**/*.md"]` checked 14 cases against folkways@a0fc9ff. All
  behaved as intended. A review committed in the app repo's `efforts/` was not found, so the push was blocked
  (team repo `efforts/app-repo-first/24`, "The gate rule: tested").
- Specs and plans are already settled as numbered topic files in the effort folder, wherever the effort lives
  (team repo `efforts/lifecycle-skills/02` item 9).
- Prior art: spec-kit keeps `specs/<branch-name>/` on the feature branch and merges it with the code
  ([spec-driven.md][spec-kit]). Kiro writes specs into `.kiro/` in the project.
- Copilot code review reads `CLAUDE.md`, `AGENTS.md`, `REVIEW.md` and `.claude/skills` from the repo (team
  repo `efforts/app-repo-first/23`). Files in the app repo are what GitHub's agents can see.
- `.folkways/` is owned by the framework (sync's manifest tracks it), so team artifacts can't go there.
- Closed efforts are indexed in the team repo's `state/closed-efforts.md` ([0005][adr-0005]).

## Decision

Option 2: split by scope. We will keep each artifact next to the thing it describes:

| Artifact | Home | Lands by |
|---|---|---|
| Cross-repo effort | Team repo `efforts/<KEY>-<slug>/` | Direct push, as today |
| Single-repo effort (status, spec, plan, topics) | App repo `efforts/<KEY>-<slug>/`, on the feature branch | That PR, squash-merged with the code; later updates go in with the effort's next PR |
| Review of a single-repo change | The app-repo effort folder, `sha:` frontmatter as today | The same PR |
| Review of cross-repo or effort-less work | Team repo, as today | Direct push |
| Personal notes, inbox, side quests | Team repo `<personal root>/inbox/`. With no team-repo clone available, a gitignored `.local/inbox/` in the app repo, which `session-close` moves | Direct push |

- The app-repo effort root defaults to `efforts/` and is set by `efforts.root` in `folkways.yaml`. The CLI
  resolves it.
- A child effort's `00-STATUS.md` header has a `parent:` line. The cross-repo effort lists its children by
  repo and path. If a single-repo effort grows to span repos, it moves to the team repo and leaves a pointer.
- The gate: `review.skip_paths` defaults to `["docs/**", "efforts/**/*.md"]`. git-guard also looks for
  reviews in the pushing repo's own effort root, not only in brain checkouts. A team that sets `skip_paths`
  replaces the default, so it adds the effort glob itself.

Why: work on one repo is planned and reviewed where the code is. It merges with the code, and GitHub's agents
and human reviewers see it in the PR. Today's gate already has the rule a committed review needs; it only lacks
the lookup in the pushing repo. Status updates going in with the next PR avoid direct pushes to app-repo
`main`, which `push.app: always` ([0014][adr-0014]) and branch protection work against.

Non-goals: how the team repo syncs into app repos (draft 0009); the path resolver's design; the inbound hub
scan; `state/` and `knowledge/`, which stay in the team repo; where sessions start (row 1, its own ADR).

Size: git-guard lookup about 20–40 lines plus tests; the `skip_paths` default one line. Skill path changes
follow the resolver (lifecycle-skills 06) and are sized there. (unconfirmed)

## Options considered

Deep analysis: team repo `efforts/app-repo-first/24`. Its "Options considered for reviews" covers git notes,
PR comments and gitignored files for reviews.

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Everything in the team repo (today) | No code change; one place for roll-ups; reviews never move HEAD | App-repo sessions need a second repo for every update; GitHub's agents can't see plans or reviews; plans aren't in the PR | Keeps the status quo, at the cost of the app-repo-first direction |
| 2 | Split by scope (chosen) | Plans and reviews next to the code and in the PR; one home for cross-repo work | Two places to look; roll-ups can't see branch-only efforts until the hub scan exists; status churn in app-repo history | Fits each artifact to where it's used, but roll-ups get harder |
| 2a | …later updates by the next PR (chosen) | Works with `push.app: always` and branch protection | Status on `main` lags until the next merge | Keeps `main` changing only through PRs |
| 2b | …later updates by direct push to `main` | Status on `main` always current | Needs `push.app` and branch rules to allow docs-only pushes | Fast, but means loosening app-repo branch protection |
| 2c | …one effort per PR, closed at merge | Simplest lifetime | Multi-PR work in one repo has to become a team-repo effort | Smallest change, but splits ongoing work across repos |
| 3 | Everything in the app repo; a lead repo holds cross-repo efforts | One rule; no team repo needed for efforts | A cross-repo initiative lives in one of its repos; personal notes in a shared code repo | Simple, but puts team and personal material into code repos |

## Consequences

- Easier: a session in an app repo plans, reviews and hands off without a second clone. Cloud agents and
  reviewers read the plan from the PR.
- Harder: `status` roll-ups. Until the hub scan (phase 1, I2) exists, they read local clones' `origin/<default>`
  and see only merged efforts.
- Status and review files add churn to app-repo PR diffs and history. Spec-kit accepts the same cost.
- A rebase or squash before a push drops the reviewed commit from HEAD's ancestry, so the review stops
  counting. This is unchanged from today.
- A `.md` symlink under `efforts/` passes the gate. A link is not code, so this is accepted.
- Edits to another effort's Markdown also pass the gate. Accepted, since it's Markdown only.
- Follow-on work:
  - git-guard: search the pushing repo's effort root and add the `skip_paths` default. Ticket to draft.
  - Effort-root config key and resolver (lifecycle-skills 06).
  - The start, review, handoff and session-close skills write to the resolved paths.
  - Closing: `state/closed-efforts.md` indexes app-repo efforts, via the hub scan or `session-close`.
- Revisit if status churn makes app-repo PRs hard to review, or if roll-ups keep missing branch-only efforts
  after the hub scan exists.

## Related decisions

- [0005][adr-0005]: closing efforts. App-repo efforts get indexed in the same file.
- [0006][adr-0006]: app repos carry Claude-format files only. Effort Markdown is not tool config.
- [0014][adr-0014]: `push.app`. The reason for option 2a over 2b.
- Draft 0009 (team repo `efforts/app-repo-first/22`): how the team repo writes into app repos. Separate.

## Interested parties

- Beta team (when it is set up).

[c-start]: https://github.com/sumsolution/folkways/blob/486104c44320f918972b98d0355a8639e7704878/packages/content/skills/start/SKILL.md
[c-review]: https://github.com/sumsolution/folkways/blob/486104c44320f918972b98d0355a8639e7704878/packages/content/skills/review/SKILL.md
[c-handoff]: https://github.com/sumsolution/folkways/blob/486104c44320f918972b98d0355a8639e7704878/packages/content/skills/handoff/SKILL.md
[c-guard]: https://github.com/sumsolution/folkways/blob/a0fc9ff/packages/folkways/internal/gitguard/guard.go
[c-review-go]: https://github.com/sumsolution/folkways/blob/a0fc9ff/packages/folkways/internal/gitguard/review.go
[c-team]: https://github.com/sumsolution/folkways/blob/a0fc9ff/packages/folkways/internal/team/team.go#L106-L111
[spec-kit]: https://github.com/github/spec-kit/blob/222a973f8b8794366680c043fb935a31f71902af/spec-driven.md
[adr-0005]: 0005-closing-efforts.md
[adr-0006]: 0006-app-repo-files.md
[adr-0014]: 0014-brain-pushes-without-asking.md
