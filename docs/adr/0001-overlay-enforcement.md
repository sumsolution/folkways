# 0001. Mores are enforced by mechanism; folkways are layered by overlays

| | |
|---|---|
| Status | accepted |
| Date | 2026-09-29 |
| Owner | Summer Jordan |

## Issue

Spec §3.2 says a skill's behaviour is framework skill + team overlay + personal overlay: the team may
add and override by section ID, a person may only add. Every skill's step 0 reads the overlay files.
That rule exists only as instructions, and models merge instructions; they don't enforce precedence.
Nothing stops a personal overlay from rewriting the commit format, or a personal skill from replacing
a team one. Skills and hooks are about to be built on §3.2, so the model has to be settled first.

Which team rules must hold regardless of what the model or a personal overlay says, what enforces
them, and how are the remaining rules layered?

## Context

- **Neither tool enforces instruction precedence.** Claude Code concatenates instruction files and calls
  them "context, not enforced configuration"
  ([memory](https://code.claude.com/docs/en/memory#claudemd-vs-auto-memory)). Copilot CLI "does not define a
  general precedence order" ([add-custom-instructions](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions)).
  [07 §1–2](../design/07-research-overlay-enforcement.md#1-how-instructions-compose-in-claude-code).
- **Settings are enforced.** Deny rules can't be loosened by any other scope; `PreToolUse` hooks block
  before allow rules ([permissions](https://code.claude.com/docs/en/permissions#settings-precedence)). Copilot
  CLI reads the same `.claude/settings.json` hooks ([hooks-reference](https://docs.github.com/en/copilot/reference/hooks-reference)).
- **Hooks from a team repo can be turned off by the user** (`disableAllHooks` in local settings or
  `--settings`, `--bare`; a project `"disableAllHooks": false` beats a user-settings `true`); only
  managed settings or Copilot policy hooks can't be. A team repo can't deploy either. Tested here: a
  project hook held under `--dangerously-skip-permissions` and fell to one `--settings` flag
  ([07 §4.1](../design/07-research-overlay-enforcement.md#41-empirical-hooks-and-rules-under-bypass-mode)).
- **Skill shadowing:** in Claude Code a personal skill of the same name replaces the project skill; in
  Copilot CLI the project skill wins ([07 §3](../design/07-research-overlay-enforcement.md#3-skill-loading-and-name-collisions)).
- **Eval** ([07 §6](../design/07-research-overlay-enforcement.md#6-eval-does-the-model-keep-the-team-rule), Sonnet,
  fresh process per run): with "apply the overlays" and no stated precedence, a personal `## Override`
  won 10/10 and a persuasive "team lead approved" `## Add` won 8/10. With one framework-written
  precedence paragraph: 0 failures in every cell (10/10 on the deciding cells). Compiling personal
  additions into the shared skill dropped the harmless one in 2 of 6 runs.
- **Prior decisions:** agent-side enforcement only, no CI in project repos (spec §1, round 7);
  personal review trigger may differ from the team's (round 7); personal skills out of v1 (spec §9, item 4);
  overlays never touch framework files (round 11).
- **Framing** (Summer, raised for this workstream): Sumner (1906), whose *folkways* gave the project its
  name ([02-naming](../design/02-naming.md)), split group norms into *folkways* (customs people may vary)
  and *mores* (norms the group enforces).

## Decision

**Option 3: mores are team config enforced by mechanism; folkways are prose layered by overlays, with
team overrides compiled in and precedence stated by the framework.**

1. **Every rule is either a more or a folkway**, listed below. A new rule is classified when it is added.
2. **Mores live in config** (`team.yaml`, now `folkways.yaml` ([16](../design/16-plan-config-and-solo.md)), framework `settings.json`), which changes by PR. They are
   enforced by hooks, permission rules and `folkways sync` / `doctor`. Skills may *describe* a more;
   the hook reads the config, so an overlay can't change it.
3. **Folkways live in skill sections and overlays.** At `sync`, the team overlay is compiled into the
   synced `SKILL.md` (overrides replace sections by ID; an unknown ID fails the sync). The personal
   overlay is read at runtime in step 0, under a precedence paragraph the framework writes, not each skill.
4. **Visibility over tamper-proofing.** `doctor` reports what it can't prevent: personal skills that
   shadow a framework or team skill, hooks disabled in user or local settings, personal overlays with
   `## Override` or override wording, and compiled skills older than their overlay.
5. **Hardening is optional and documented, not required:** managed settings (`allowManagedHooksOnly`),
   Copilot policy hooks, a GitHub push ruleset on team paths, and
   `permissions.disableBypassPermissionsMode` (works from the team repo too, but a person may need bypass mode in a sandbox). They need an org or repo admin.

Rationale: the eval shows the folkway layer works *when precedence is stated*, and fails badly when it
isn't, so the statement must be the framework's job. Mores can't rest on that: 18 of the eval's
failures would have been committed, and every one had an Angular header a format check rejects. From a
team repo, mechanism is accident-proof, not tamper-proof; for a 3–8 person team, the threat is model
lapses and well-meant overlays, not a colleague disabling hooks, and `doctor` makes that visible.
Compiling team overrides removes one instruction-following step for the rules the team cares most
about and moves the override-ID check to sync time. Keeping personal additions at runtime keeps them
out of the shared file (the eval shows why) and needs no per-person build.

### Mores: enforced by mechanism

| # | Rule | Mechanism | Configured in | Status |
|---|---|---|---|---|
| M1 | Commit header format and allowed scopes | `git-guard` on `git commit`: checks the message | `team.yaml` `git.commit_convention`, `git.scopes` (+ proposed `git.commit_pattern` for non-Angular formats) | planned, spec §7.1 |
| M2 | Push needs the user's approval | permission `ask: Bash(git push:*)`; never auto-approved, even in bypass mode | framework `assets/settings.json` | built |
| M3 | A review covers HEAD before push (trigger `pre-push`) | `git-guard` on `git push` | `team.yaml` `review.trigger` | planned, spec §7.1 |
| M4 | The main clone's checkout is never changed | `git-guard` denies `checkout`, `switch`, `pull`, `reset`, `rebase` there | `.local/me.yaml` `projects_dir` | planned, spec §7.1 |
| M5 | Branch names follow the team pattern | `git-guard` on `git push` and branch creation | `team.yaml` `git.branch_pattern` | **new** |
| M6 | Headless runs: no push to a default branch, nothing destructive | `folkways run` passes deny rules (`--disallowedTools`); `git-guard` checks headless | `team.yaml` `repos[].default_branch` | **new** (spec §7.5 is prose today) |
| M7 | Framework files are not edited | `sync` checksum check | `.folkways/manifest.json` | built |
| M8 | Personal overlays only add | `lintPersonalOverlay` in `sync` (warning) and `doctor` (error). **Detect only**: at runtime it's the step-0 prose (a folkway), and the lint misses a quiet contradiction | framework | built in `sync` (`src/overlays.ts`); `doctor` not built |
| M9 | Team overrides apply | compiled into `SKILL.md` at `sync`; unknown section ID fails the sync | `overlays/<skill>/OVERLAY.md` | built |
| M10 | Framework and team skills aren't shadowed | `doctor` lists same-name skills in `~/.claude/skills`, `~/.agents/skills`, `~/.copilot/skills` | framework | **new**, detect only |
| M11 | Hooks stay on | framework settings set `"disableAllHooks": false`, which beats a user-settings `true`; `doctor` warns on `true` in local settings | framework `assets/settings.json` | **new**; closes one route, detects the other |
| M12 | `team.yaml`, `overlays/`, team skills change by PR | CODEOWNERS routes reviews; optional push ruleset "Restrict file paths" with a maintainer bypass | team repo, GitHub settings | recommended by `setup` |

### Folkways: prose; team overrides by section ID, people add

Commit body content and examples (the header is M1) · PR description · ticket templates and examples ·
ADR template, examples and wording · review focus areas and extra checks · test style notes · research
source notes · tone and writing style · QA evidence format · status-doc and session-close details ·
where side quests go.

### Wiring (status)

- Done: `sync.ts`: after copying a framework skill, if `overlays/<skill>/OVERLAY.md` exists, write
  `compileSkill(framework, { team })`. The manifest records the compiled file's checksum (so edit
  detection still works) and the overlay's checksum (so `doctor` can spot a stale compile). A
  `FolkwaysError` from an unknown ID stops the sync before anything is written.
- `lintPersonalOverlay` over `people/*/overlays/*/OVERLAY.md`: done in `sync` (warnings); `doctor` not built.
- Done: framework skills: step 0 is one shared block: read `team.yaml` and `.local/me.yaml`, then the
  personal overlay, under the precedence text tested as `hybrid`
  ([07-eval/fixtures/step0-hybrid.md](../../research/evals/overlay-enforcement/fixtures/step0-hybrid.md)). A framework-repo test
  checks every `assets/skills/*/SKILL.md` contains it.
- `compileSkill`'s personal branch exists for the eval (Option 4); production doesn't use it.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Keep spec §3.2: runtime read of three files; each skill states the precedence | No new code; eval shows it holds when the text is present | Every skill author must remember the paragraph, and without it personal overlays won 18/20; team overrides depend on a file read; no mores concept | Works in the lab, fails on the first skill that forgets one paragraph. |
| 2 | Instructions only, made stronger: framework-owned precedence block, lint, `SessionStart` reminders | Cheap; fixes the forgotten-paragraph risk; eval shows 0 failures with it | Every rule, commit format included, still rests on the model; Opus-vs-Sonnet gap shows it varies by model | Good enough for folkways, not for rules whose failure lands in a shared repo. |
| **3** | **Mores by mechanism, folkways by overlay (team compiled, personal at runtime), `doctor` for visibility** | Mores hold whatever the overlay or model does; folkways get the eval's best layout; bypass is visible | Mores must be expressible as config; `sync` gains a compile step; a determined member can still turn hooks off | Accident-proof for the rules that matter, at the cost of one sync step and config for every more. |
| 4 | Compile everything per person into `~/.claude/skills/<name>`, using Claude Code's personal-over-project rule | One file per person; overlays can't be skipped | Copilot CLI inverts the rule (project wins); writes outside the repo; eval: compiled personal adds dropped 2/6; clashes if a person ever joins two teams | Uses a shadowing rule that one of the two tools reverses. |
| 5 | Require admin enforcement: managed settings, `allowManagedHooksOnly`, Copilot `policy.d`, GitHub push rulesets | Tamper-proof | A team repo can't deploy any of it; needs an org or machine admin; not portable across companies | Right for regulated teams, too heavy as the default; offered as hardening. |

## Consequences

- **Spec §3.2** rewritten to match (this PR, headed `(ADR 0001)`).
- **Spec follow-ups** (other sections, not in this PR): §3.1 `sync` compiles team overlays and records
  both checksums; §4.1 adds `git.commit_pattern`; §6 step 0 becomes the shared block; §7.1 `git-guard`
  gains M5 and M6 and reads every more from `team.yaml`; §7.2 `doctor` gains M8, M10, M11; §7.5 headless
  limits become deny rules; §8.1 `setup` offers CODEOWNERS and the optional push ruleset.
- **Code:** wire `src/overlays.ts` into `sync.ts` and `doctor` (description above). `assets/settings.json`
  gains `"disableAllHooks": false`, and `mergeSettings` must carry that key (today it merges hooks and `ask` only). `git-guard` must stay
  fast: Copilot CLI hook timeouts fail open.
- **Skills:** a skill section that documents a more says where it's configured and that `git-guard`
  checks it. Team overrides of that section change the prose only.
- **Teams:** put rules that must hold in `team.yaml`, not in an overlay. Review `overlays/` like code.
- **Docs:** `assets/instructions.md` and the team README explain mores vs folkways in two lines each.
  `AGENTS.md` ("Step 0 reads `team.yaml`, `.local/me.yaml` and both overlays") changes to the new step 0.
- **Closed:** round 7 let a person set their own review trigger, which could loosen M3. The trigger is now a
  floor: `.local/me.yaml` can only tighten the team's value (spec §4.2).
- **Open:** interactive bypass mode and push-ruleset behaviour on PR merges are untested
  ([07, unknowns](../design/07-research-overlay-enforcement.md#unknown-or-unverified)).
