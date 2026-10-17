# 0008. Who owns an app repo's AGENTS.md?

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-17 |
| Owner | Summer Jordan |

## Issue

folkways is moving to app-repo-first: engineers start sessions in their app repos, and the team repo syncs
framework and team content into them. Every agent that works in an app repo reads its `AGENTS.md`, so
the sync wants to put folkways content there. That file already holds the repo's own instructions,
written by the people who own the repo, and some of them are not on the folkways team. The sync design
needs this settled before it is built.

May folkways write an app repo's `AGENTS.md`, and if it may not, where does folkways content go and how
does `AGENTS.md` refer to it?

## Context

- **octo-team-context overwrites the file.** Its outbound sync copies one hub `agents/AGENTS.md` over each
  app repo's `AGENTS.md` and `CLAUDE.md`, and a daily reconcile puts it back if anyone edits it. Repos
  lose their own instructions, and local edits are reverted without notice
  ([20 §3](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L118-L132);
  [octo](https://github.com/summer-jordan/octo-team-context) at `2e9869c`, `sync-outbound.yml`, `reconcile.yml`).
- **Today's rule for team repos already keeps `AGENTS.md` team-owned.** `init` writes it once; `sync`
  never writes it and only warns when the `@.folkways/instructions.md` line is missing
  ([spec §3 ownership table and §5](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/06-v1-spec.md);
  [`sync.ts` L333–336](https://github.com/sumsolution/folkways/blob/ec8f1fc34d355f17c9423efb4dcc4ae3a6168838/packages/cli/src/sync.ts#L333-L336)).
  Round 12 dropped an earlier `<!-- folkways:begin/end -->` managed block in favour of that one import line
  ([01 round 12](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/01-interview-notes.md?plain=1#L416-L420)).
  That rule was made for team repos, where the folkways team owns the whole repo. In an app repo the file's
  owners can include people outside it (inferred).
- **App-repo-first composition A names the fix:** "a folkways-owned folder instead of overwriting
  `AGENTS.md`" ([20 §5A](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L220-L231)).
  Outbound option O1 "fails the 'AGENTS.md is team-owned' rule unless it writes only a managed file"
  ([20 §4.1](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L140)).
- **Two sync tools avoid the overwrite, in different ways**
  ([20 §7e](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L653-L654),
  [what this changes](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L741-L745)):
  [trick77/agents-md-sync](https://github.com/trick77/agents-md-sync) generates `AGENTS.md` and gives the
  repo an addendum folder, `.agents/`; [yuka-repos/agents-md-sync](https://github.com/yuka-repos/agents-md-sync)
  writes only a marked block inside a hand-written `AGENTS.md`, leaves text outside the markers alone, and
  has a check mode that fails on drift.
- **Committed files are the only thing every agent gets.** Summer's answer 2: an app repo must carry
  everything an agent needs, including Copilot in GitHub
  ([20 §0](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L19-L20)).
  The Copilot cloud agent followed `AGENTS.md`, `CLAUDE.md` and `.github/copilot-instructions.md` in a
  probe; code review followed `REVIEW.md`
  ([20 §7b hands-on](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L398-L408)).
- **`@` imports are confirmed for two of the four agents.** Claude Code and Copilot CLI expand `@` imports
  in `AGENTS.md` and `CLAUDE.md`
  ([spec §5](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/06-v1-spec.md);
  [20 §7a](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L324-L328)).
  Whether the Copilot cloud agent and Copilot code review expand them is **not researched or tested**.
- **Claude Code reads `CLAUDE.md`, not `AGENTS.md`;** a team repo's `CLAUDE.md` holds `@AGENTS.md`
  ([spec §5](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/06-v1-spec.md)).
  Datadog uses the same one-line `CLAUDE.md`
  ([post](https://dev.to/datadog-frontend-dev/steering-ai-agents-in-monorepos-with-agentsmd-13g0)).
  octo overwrites both files, so the question applies to `CLAUDE.md` too.

## Decision

**folkways never overwrites or edits the team's text in an app repo's `AGENTS.md`. The repo's team owns
that file. folkways writes only paths it owns.**

1. Framework and team content goes in a folkways-owned folder, `.folkways/`, plus the framework skill
   folders and settings entries the manifest already lists (spec §3 table). `sync`, or a hub, replaces only
   those paths.
2. `AGENTS.md` refers to that content. Adding the reference is the repo team's edit. folkways checks for
   it and warns, as `sync` does today; it doesn't add it.
3. **Not settled here: the form of the reference.** The research doesn't decide between:
   - **(a) an import line**, `@.folkways/instructions.md`, today's mechanism. Works in Claude Code and
     Copilot CLI; unknown in the Copilot cloud agent and code review.
   - **(b) a plain-text pointer** ("read `.folkways/instructions.md` first"). Any agent can follow it, but
     loading depends on the model doing so (unconfirmed for each agent).
   - **(c) a marked block** the repo team lets folkways own inside `AGENTS.md`. The full text is in the
     file every agent reads, at the cost of folkways writing inside the team's file.

   (a) stays the default because it is built and holds for both local tools. A probe of the Copilot cloud
   agent and code review with an import line decides whether (a) holds for GitHub agents, or whether a
   fallback, (b) or (c), is needed for them.

Rationale: an overwrite costs the repo's owners their own instructions every time the sync runs, and in
an app repo those owners may not be on the folkways team, so they can't fix it in the hub. Owning a
folder keeps every folkways write reviewable as a change to files only folkways writes. It also keeps
today's team-repo rule, so `sync` needs no new ownership model. The reference form is left open because
the one agent class Summer named as most important, Copilot in GitHub, is the one we haven't tested with
imports.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Overwrite `AGENTS.md` wholesale from the hub (octo) | Simplest; full text in the file every agent reads, no import needed | Erases the repo's own instructions; the reconcile reverts edits without notice; the repo team can't change its own file | Easy to build, and it takes the file away from the people who own the repo. |
| 2 | A marked managed block inside `AGENTS.md` (yuka-repos; folkways before round 12) | Full text reaches every agent; text outside the markers stays the team's; a check can catch drift | folkways writes inside a team file; edits inside the markers get reverted; the block makes the file longer; diffs land in the team's file | Works in every agent at the cost of a shared file with two writers. |
| 3 | Tool-owned `AGENTS.md` plus a repo-owned addendum folder (trick77) | Clear split; the tool can lay out the whole file | The team loses `AGENTS.md` and edits a side folder instead; an app repo that already has an `AGENTS.md` must move its text | Clean ownership, but the wrong owner for an app repo. |
| **4** | **folkways-owned folder; the team's `AGENTS.md` refers to it (reference form open: 4a import, 4b pointer)** | folkways never writes the team's file; reuses today's `sync` and ownership table; 4a is built and confirmed in both local tools | One line the team must add; 4a unconfirmed in Copilot cloud agent and code review; 4b relies on the model reading the file | Keeps the file the team's, and leaves one untested gap for GitHub agents. |

## Consequences

- **Spec:** the spec §3 ownership table applies to app repos unchanged: `AGENTS.md` is "team, never
  written". What `init` or `setup` does in an app repo with no `AGENTS.md` is open; creating it once, as
  `init` does today, is the obvious carry-over (unconfirmed).
- **Hub designs** (architecture A, or each repo pulling, decided elsewhere) may write only folkways-owned
  paths. Whether a hub PR that adds the reference line counts as folkways writing `AGENTS.md` goes with
  that decision.
- **Test to run:** a throwaway repo whose `AGENTS.md` and `CLAUDE.md` import a file with a marker; assign
  an issue to the Copilot cloud agent and request Copilot review. The result picks 4a, or 4a plus a
  fallback for GitHub agents.
- **`CLAUDE.md`:** the same rule should hold (Claude Code needs `@AGENTS.md` there). This record doesn't
  decide it. If the repo's `CLAUDE.md` doesn't import `AGENTS.md`, Claude Code misses the reference; `sync`
  could warn on that as well.
- **Cost:** adoption needs one edit by each repo team, and a repo that skips it gets no folkways
  instructions (skills in `.claude/skills` still load). A missing line is a warning, not a failure.
- **Revisit** if the probe shows GitHub agents ignore both imports and pointers, which would leave only
  option 2 for them, or if repo teams consistently ask folkways to manage the file for them.
