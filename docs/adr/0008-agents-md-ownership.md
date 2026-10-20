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
- **`@` imports reach three of the four agents.** Claude Code expands `@` imports in `AGENTS.md` and
  `CLAUDE.md`
  ([spec §5](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/06-v1-spec.md)).
  Copilot CLI 1.0.98, logged in, expanded one in `AGENTS.md`: it followed a rule found only in the imported
  file and ran no read tool
  ([probe Q5, row 6](https://github.com/sumsolution/folkways-command/blob/e7781d5/efforts/app-repo-first/23-probe-copilot-github.md?plain=1#L226-L242)).
  The Copilot cloud agent does too: told not to read `AGENTS.md` or `docs/`, it followed a rule that
  existed only in the imported file, and its log shows no read of that file (probe issue
  [#7](https://github.com/sumsolution/folkways-agent-sandbox/issues/7) →
  [PR #8](https://github.com/sumsolution/folkways-agent-sandbox/pull/8)).
- **Copilot code review follows neither an import nor a pointer.** It followed rules written directly in
  `AGENTS.md`, `CLAUDE.md`, `REVIEW.md` and `.github/copilot-instructions.md`, and missed the imported rule
  both times
  ([PR #2](https://github.com/sumsolution/folkways-agent-sandbox/pull/2),
  [PR #6](https://github.com/sumsolution/folkways-agent-sandbox/pull/6)). With only a plain "read
  `.folkways/instructions.md` and follow its rules" line in `AGENTS.md`, it missed the rule in that file
  too ([PR #9](https://github.com/sumsolution/folkways-agent-sandbox/pull/9);
  [probe Q1, pointer form](https://github.com/sumsolution/folkways-command/blob/e7781d5/efforts/app-repo-first/23-probe-copilot-github.md?plain=1#L93-L114)).
- **Code review reads skills** from `.claude/skills` and from `.github/skills`, each tested alone with a
  skill named `code-review`
  ([PR #4](https://github.com/sumsolution/folkways-agent-sandbox/pull/4),
  [PR #5](https://github.com/sumsolution/folkways-agent-sandbox/pull/5)). The docs say review-focused
  names are more likely to be picked
  ([20 §7b](https://github.com/sumsolution/folkways-command/blob/991d759c19cb77530bd2619150c84de52de4a5d8/efforts/app-repo-first/20-research-app-repo-architecture.md?plain=1#L349)).
- **Probe limits:** private repo
  [sumsolution/folkways-agent-sandbox](https://github.com/sumsolution/folkways-agent-sandbox), one run
  per case, model chosen by GitHub. The import fixtures for the cloud agent and Copilot CLI used `docs/…`,
  not `.folkways/…`; the code review pointer test used `.folkways/`.
- **Claude Code reads `CLAUDE.md`, not `AGENTS.md`;** a team repo's `CLAUDE.md` holds `@AGENTS.md`
  ([spec §5](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/v1-design/06-v1-spec.md)).
  Datadog uses the same one-line `CLAUDE.md`
  ([post](https://dev.to/datadog-frontend-dev/steering-ai-agents-in-monorepos-with-agentsmd-13g0)).
  Code review reads `CLAUDE.md` on its own
  ([PR #6](https://github.com/sumsolution/folkways-agent-sandbox/pull/6)). octo overwrites both files.

## Decision

**folkways never overwrites or edits the team's text in an app repo's `AGENTS.md` or `CLAUDE.md`. The
repo's team owns them. folkways writes only paths it owns.**

1. **Owned paths:** `.folkways/`, plus the framework skill folders and settings entries the manifest
   already lists (spec §3 table). `sync`, or a hub, replaces only those paths.
2. **Session agents get the instructions by import.** The team's `AGENTS.md` holds
   `@.folkways/instructions.md` (option 4a), as team repos do today. That reaches Claude Code (through
   `CLAUDE.md`'s `@AGENTS.md`), Copilot CLI and the Copilot cloud agent. Adding the line is the repo team's
   edit; `sync` checks for it and warns, and doesn't add it.
3. **Code review gets what it needs from a skill.** Code review doesn't follow the import, so folkways
   ships its review criteria as a framework skill under `.claude/skills/`, a folder it already owns
   (option 4c). The criteria come from the existing `review` skill and the team's compiled overlay. The
   skill's name and description must read as code review, since the review agent picks skills by
   relevance.

Rationale: an overwrite costs the repo's owners their own instructions every time the sync runs, and in
an app repo those owners may not be on the folkways team, so they can't fix it in the hub. Owning paths
keeps every folkways write reviewable as a change to files only folkways writes, and keeps today's
team-repo rule. The import is built, and the probes show it reaches every agent that runs a session. For
code review, which follows neither an import nor a pointer, a skill is the only form that needs no
write to a file the repo already owns or may own: `REVIEW.md` and `.github/copilot-instructions.md` are
the repo's to write, and a marked block puts two writers in one file. Most of `instructions.md` is about running a session (status docs, handoff, pushes),
which a reviewer doesn't need (inferred).

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Overwrite `AGENTS.md` wholesale from the hub (octo) | Simplest; full text in the file every agent reads, no import needed | Erases the repo's own instructions; the reconcile reverts edits without notice; the repo team can't change its own file | Easy to build, and it takes the file away from the people who own the repo. |
| 2 | A marked managed block inside `AGENTS.md` (yuka-repos; folkways before round 12) | Full text reaches every agent, code review included; text outside the markers stays the team's; a check can catch drift | folkways writes inside a team file; edits inside the markers get reverted; the block makes the file longer; diffs land in the team's file | The one form code review reads for sure, at the cost of a shared file with two writers. |
| 3 | Tool-owned `AGENTS.md` plus a repo-owned addendum folder (trick77) | Clear split; the tool can lay out the whole file | The team loses `AGENTS.md` and edits a side folder instead; an app repo that already has an `AGENTS.md` must move its text | Clean ownership, but the wrong owner for an app repo. |
| 4a | folkways-owned `.folkways/`; the team's `AGENTS.md` imports it | folkways never writes the team's file; built today; reaches Claude Code, Copilot CLI and the cloud agent | One line the team must add; code review ignores it | Covers every session agent; code review needs something else. |
| 4b | Same, with a plain-text pointer instead of an import | Readable by any agent in principle | Loading depends on the model; code review ignored it (PR #9) | Weaker than 4a everywhere 4a works, and no better where it doesn't. |
| **4a + 4c** | **4a for sessions, plus a folkways review skill under `.claude/skills/` for code review** | Uses folders folkways already owns; the review skill already exists; code review was seen reading `.claude/skills` | Skills load by relevance, not always; tested with a skill named `code-review`, not folkways' name; a repo's own review skill may compete | Reaches every agent without writing a file the repo owns, and rests on one probe run for review. |
| 4d | 4a, plus a folkways-owned `REVIEW.md` or `.github/copilot-instructions.md` | Read directly by code review, every time | Both are files a repo may already own; `copilot-instructions.md` also loads in every Copilot session, duplicating the import | Reliable for review, and brings back the ownership clash for a different file. |

## Consequences

- **Spec:** the spec §3 ownership table applies to app repos unchanged: `AGENTS.md` is "team, never
  written", and `CLAUDE.md` joins it. What `init` or `setup` does in an app repo with no `AGENTS.md` is
  open; creating it once, as `init` does today, is the obvious carry-over (unconfirmed).
- **`sync` warnings:** besides the missing import line, warn when `CLAUDE.md` exists and doesn't import
  `AGENTS.md`, since Claude Code then misses the instructions.
- **Review skill:** folkways' skill is named `review`. Name it, or describe it, so code review picks it, and
  probe that once: a PR with only the synced folkways skill, checking for a marker in the review. Until
  then, 4c is unconfirmed for folkways' own skill.
- **Hub designs** (architecture A, or each repo pulling, decided elsewhere) may write only folkways-owned
  paths. Whether a hub PR that adds the import line counts as folkways writing `AGENTS.md` goes with that
  decision.
- **Cost:** adoption needs one edit by each repo team. A repo that skips it gets no folkways instructions
  in sessions; skills in `.claude/skills` still load, and code review is unaffected. A missing line is a
  warning, not a failure.
- **Revisit** if code review stops picking the folkways skill or a repeat run contradicts the probe (then
  4d, accepting the clash on `REVIEW.md`), if the cloud agent stops expanding imports, or if repo teams
  consistently ask folkways to manage the file for them.
