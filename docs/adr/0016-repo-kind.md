# 0016. How does a repo say what kind it is, and how does an app repo find its team repo?

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-30 |
| Owner | Summer Jordan |
| Reviewers | Summer Jordan (one-person team, `adr.reviewers.min: 0`) |
| Review window | until Summer's review of the PR that adds this record |

Absorbs the "sessions start in either repo" ADR (team repo `efforts/app-repo-first`, Direction row 1) and the
content half of draft 0009 ([app-repo-first 22](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/app-repo-first/22-adr-outbound-sync.md)).

Code links are folkways `origin/main` @ `2bd8838` (0.5.0).

## Issue

Sessions can start in an app repo or in the team repo (app-repo-first Direction row 1), and single-repo efforts
and reviews now live in the app repo ([0015][adr-0015]). The code has no idea which kind of repo it is in. It
calls whatever directory the session started in "the team repo" ([team.go:51-70][team-load]), and git-guard's
"brain" means "same git common dir as that directory" ([push.go:130-135][isbrain]). Two bugs follow:

- [folkways#86][i86]: in a session started in an app repo, git-guard treats the app repo as the brain, skips the
  review gate ([guard.go:182-184][checkpush]) and applies `push.brain: never`. An unreviewed push reaches `main`
  with no prompt (probed, [app-repo-first 27](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/app-repo-first/27-finding-gate-in-app-repo-sessions.md)).
- [folkways#81][i81]: `folkways sync` run in an app repo writes brain files there, because it accepts any git
  toplevel ([main.go:203-214][teamroot]).

An app repo also can't see the person's `.local/me.yaml` or the team's rules: both live in the team repo
([lifecycle-skills 06](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/lifecycle-skills/06-research-config.md) findings 6, 7). The path resolver, `efforts.root` and per-repo packs
all need the same answer, so it comes first (team repo `efforts/lifecycle-skills/00-STATUS.md`, Next, order agreed 2026-10-29).

How does a repo declare its kind, what does sync write into an app repo, and how does an app-repo session find the
team repo?

## Relevant context

- Summer chose an explicit `kind` over inferring it from "has `folkways.yaml` with `members`" (2026-10-29,
  team repo `efforts/lifecycle-skills/00-STATUS.md`, Decisions).
- `folkways.yaml` is parsed leniently: unknown keys are ignored ([team.go:72-77][team-load]; 06 finding 4), so a
  new `kind` key is backward compatible.
- `sync` writes `.claude/skills/`, `.claude/settings.json`, `.folkways/instructions.md`, `.folkways/prompts/` and
  the manifest, and refuses to run over edited framework files ([teamsync.go:71-96, 142-181][teamsync]).
  `.folkways/` is framework-owned ([0015][adr-0015] context).
- Every repo, brain or app, commits its skills and hook entries ([0012][adr-0012]). Hooks run wherever those
  entries are committed, so git-guard runs in app-repo sessions.
- Overlays live in the team repo's `overlays/` and are compiled by `sync`. An app repo can only get the team's
  compiled skills from a sync that reads the team repo.
- The main-clone guard skips the session's project dir: "the folkways repo itself is never guarded"
  ([guard.go:206-226][mainclone]). In an app-repo session that exempts the app repo's main clone (inferred, not
  run; 27).
- Session start offers brain setup in any repo without `folkways.yaml` ([sessionstart.go:21-26][ss]), so an
  app-repo session is told folkways isn't set up.
- Worktrees follow `<projects_dir>/.worktrees/<repo>/<branch>` and main clones `<projects_dir>/<repo>`
  (`.folkways/instructions.md`, "Project repos"). From any checkout, `git rev-parse --git-common-dir` gives the
  main clone's `.git`, so `projects_dir` is two levels up from it.
- The brain does not list itself in `repos[]` (this repo's `folkways.yaml`); app repos are listed there.
- Cloud agents and outside contributors work in an app repo with no team-repo clone. The gate must still hold
  there.

## Decision

Option 3: sync writes a repo file (Summer chose it over 3a, 2026-10-30). We will:

1. **Declare kind.** `folkways.yaml` gets `kind: brain`; a file without `kind` is a brain (current repos keep
   working). An app repo carries `.folkways/repo.yaml` with `kind: app`. A repo with neither has kind `none`.
2. **Sync from the team repo into the app repo.** `folkways sync` run in an app repo finds the team repo
   (rule 4), checks that the app repo's `origin` URL is in the team's `repos[]`, and writes the same files as
   today plus `.folkways/repo.yaml`. Not listed, or no team repo found: exit 1, write nothing, and say what to do.
   That is the fix for [#81][i81]. Sync run in a repo with neither file errors as #81 asks.
3. **Copy the team's config into `repo.yaml`.** Sync copies all of the team's `folkways.yaml`, plus `kind`,
   `efforts.root`, the team repo's name and URL, and later the repo's effective packs. Skills read what their
   `## Config` section names, and the app-repo skills together name nearly every key (all but `team` and
   `knowledge_audit` at `2bd8838`), so a partial copy would need a list kept in step with the skills. Session
   start prints `repo.yaml` where it prints `folkways.yaml` today. The file is in the manifest, so local edits
   stop the next sync; changes go in the team repo.
4. **Find the team repo** from `FOLKWAYS_BRAIN` if set, else at `<projects_dir>/<brain name>`, with
   `projects_dir` derived from the checkout layout. Accept it only if its `folkways.yaml` is `kind: brain` and its
   `origin` URL matches `repo.yaml`. Found: read `.local/me.yaml` there for the handle, `projects_dir` and personal
   overrides. Not found: run on `repo.yaml` alone (06's option (c), with (a) as override).
5. **Key git-guard on the pushing repo's kind**, not on the session dir. Kind `brain`: today's brain rules. Kind
   `app` or `none`: the review gate, `push.app`, and reviews looked up in the pushing repo's `efforts.root`
   ([0015][adr-0015]). Rules come from the team repo when found, else from `repo.yaml`; a `none` repo has no `repo.yaml`, so it
   gets rules only from a found team repo. The main-clone guard
   exempts only kind `brain`. That is the fix for [#86][i86].
6. **Session start in an app repo** prints the resolved team repo and personal root, or, when no team repo is
   found, the clone command from `repo.yaml`'s URL (on a local machine) or a note that it runs without one. It
   no longer offers brain setup.

Why: the gate has to hold with no team repo on the machine, so the rules must be in the app repo; rendering them
from the team repo keeps one place to edit them. The file is under `.folkways/`, which sync already owns and
checks for edits, so it adds no new ownership rule. Finding the team repo next to the app repo needs no new
setup step on most machines, and the env var covers the rest.

Non-goals: how sync reaches app repos without a person running it (draft 0009, narrowed to delivery and
parked); the resolver's command and fields; pack membership (06's catalog question); `efforts.root` per repo.

Size: about 300 Go lines plus about 15 fixtures: `team` kind and `repo.yaml` parse about 50, team-repo discovery
about 60, sync in an app repo (registry check, render, error paths) about 100-150, git-guard re-keyed about 40,
session start about 20. (unconfirmed: estimated from 06 "Size" and 0015's git-guard figure, not drafted)

## Options considered

Deep analysis: [06](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/lifecycle-skills/06-research-config.md) ("How an app repo finds the brain", "Failure modes"), [08](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/lifecycle-skills/08-research-cli-steps.md)
(Risks 1), app-repo-first [27](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/app-repo-first/27-finding-gate-in-app-repo-sessions.md).

Kind and app-repo config:

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Infer kind: brain if `folkways.yaml` has `members` (or solo) | No new key or file | A team adopting a repo can't say what it is; a missing field silently flips the kind and the gate; Summer ruled it out | Cheapest, but the gate depends on a guess |
| 2 | App repo commits its own `folkways.yaml` with `kind: app` | One file name everywhere; the app team can edit it | Two places set the rules, so they drift; today's code and #81's acceptance criteria read `folkways.yaml` as "brain or adopting" | Gives app repos control, at the cost of a second source of truth |
| 3 | Sync writes `.folkways/repo.yaml` with kind, team repo and rendered rules (chosen) | Gate works with no team repo; one place to edit rules; edit check already exists | Rules in app repos lag until the next sync; app teams can't tune a rule locally | Keeps app repos self-sufficient, but every rule change needs a sync |
| 3a | …`repo.yaml` holds only kind and team repo; rules read from the team repo at run time | Rules always current; smallest file | Cloud agents and contributors with no team repo get no gate, which is #86 again | Simple, but fails open exactly where the gate matters most |
| 4 | No repo file; env var or a pointer file in `~/.config/folkways/` only | No sync change | Nothing in the clone says it is an app repo; cloud agents have neither | Fixes local sessions, leaves the rest unguarded |

Finding the team repo (rule 4):

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| a | `FOLKWAYS_BRAIN` env var only | Explicit; trivial | Every person sets it per shell and tool; unset means silent fallback | Precise, but easy to forget |
| b | Pointer file written by `onboard` (precedent: `offered` file, [sessionstart.go:91-125][ss]) | Set once per machine | One brain per machine; another setup step; stale when the clone moves | Works for one team per laptop |
| c | Name from `repo.yaml`, resolved next to the app repo; `FOLKWAYS_BRAIN` overrides (chosen) | No setup on a standard layout; verified by `kind` and URL | Fails on a non-standard layout without the override | Uses the layout folkways already prescribes, with an escape hatch |

## Consequences

- Easier: an app-repo session gets the gate, `push.app`, its effort root and the person's overrides without
  being in the team repo. The resolver can read `kind` instead of guessing (lifecycle-skills 08, Risks 1).
- Harder: a rule change in the team repo reaches an app repo only when someone syncs it. With manual sync (the
  dogfood repo) that can lag; draft 0009's delivery closes it.
- A person's stricter `review.trigger` and `push.app` overrides apply only when the team repo is found.
- Copying rules means the app-repo gate and the team's can differ for a while. `session-start` could warn when
  the found team repo's rules differ from `repo.yaml` (not decided).
- Every app repo carries `members` (names, GitHub handles), `repos` and the tracker and docs URLs. In a public app
  repo that publishes them. If a public app repo with real team data appears, add a per-repo setting that leaves
  those keys out (for example `repos[].public: true`).
- An app repo not in `repos[]` can't be synced. The team adds it first, which makes `repos[]` the registry 0009
  assumed.
- Brain-only skills (`onboard`, `setup`, `knowledge-audit`) are still written into app repos until packs exist.
- #86 stays open until the git-guard change ships. The interim from 27 (ask instead of allow when the session dir
  has no `folkways.yaml`) is separate and small.
- Follow-on work, build step 5 in the team repo's lifecycle-skills effort: `kind` and `repo.yaml` in `team`; sync in app repos (#81); git-guard
  re-keyed (#86) together with 0015's effort-root lookup and `skip_paths` default (ticket to draft); session start.
  Then the resolver reads `kind` (step 6).
- Revisit if app teams need to tune rules locally (option 2), or if non-standard layouts make the override common.

## Related decisions

- [0012][adr-0012]: every repo commits its skills and hooks. This decides what else an app repo carries.
- [0014][adr-0014]: `push.brain` and `push.app`. This decides which one applies.
- [0015][adr-0015]: effort homes. Its git-guard lookup needs the kind; its `efforts.root` is rendered here.
- [0006][adr-0006] (proposed): Claude-format files only. `repo.yaml` is folkways config, not tool config.
- Draft 0008 (folkways#43): sync writes only folkways-owned paths. `.folkways/repo.yaml` is one.
- Draft 0009 ([22](https://github.com/sumsolution/folkways-command/blob/ceb6cda/efforts/app-repo-first/22-adr-outbound-sync.md)): narrowed to delivery; this takes its content half.

## Interested parties

- Beta team (when it is set up).

[adr-0006]: 0006-app-repo-files.md
[adr-0012]: 0012-cli-only.md
[adr-0014]: 0014-brain-pushes-without-asking.md
[adr-0015]: 0015-effort-homes.md
[i81]: https://github.com/sumsolution/folkways/issues/81
[i86]: https://github.com/sumsolution/folkways/issues/86
[team-load]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/internal/team/team.go#L51-L77
[isbrain]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/internal/gitguard/push.go#L130-L135
[checkpush]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/internal/gitguard/guard.go#L178-L184
[mainclone]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/internal/gitguard/guard.go#L206-L226
[teamroot]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/cmd/folkways/main.go#L203-L214
[teamsync]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/internal/teamsync/teamsync.go#L71-L181
[ss]: https://github.com/sumsolution/folkways/blob/2bd8838/packages/folkways/internal/sessionstart/sessionstart.go#L21-L125
