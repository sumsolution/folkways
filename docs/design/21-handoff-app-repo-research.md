# Handoff: app-repo-first research, continued with full internet access

**Written:** 2026-10-16 (session 11) · **For:** the next session, in an environment with full network
access · **Index:** [00-STATUS](00-STATUS.md)

## Why a new session

Session 11's container blocked most outbound sites. Web search worked, but WebFetch and curl could reach
only `code.claude.com` and GitHub clones. So most claims in [20](20-research-app-repo-architecture.md)
outside the Claude Code docs are marked `~` (search summary, page not opened). This session verifies them
and runs the hands-on tests.

## Ground rules (Summer)

- **Collaborate, don't decide.** Collect data, examples and options for Summer to review. No
  recommendation unless asked; no design changes to the spec or code.
- The premise: engineers keep working in **app repos**. The team still has a central agent repo that holds
  shared knowledge, understands every team repo, and **acts as the sync between them**. The brain-repo
  model (sessions start in the team repo) lacked support; Summer heard several concerns about it.
- The example is [summer-jordan/octo-team-context](https://github.com/summer-jordan/octo-team-context),
  a fork of `TheJeans/octo-team-context`, read in full at `2e9869c`. Section 3 of 20 reviews it.

## Where things are

| What | State |
|---|---|
| Branch | `claude/determined-clarke-k9kpdh` |
| PR | [sumsolution/folkways#37](https://github.com/sumsolution/folkways/pull/37), docs only, open |
| Research doc | [20-research-app-repo-architecture.md](20-research-app-repo-architecture.md): three flows, options per flow, four composed architectures (A CI hub, B team repo as plugin marketplace, C package per app repo, D per machine), what each changes, unknowns (§7), questions (§8) |
| Questions to Summer (§8) | **Answered** 2026-10-16; see 20 §0. Biggest: app repos must carry everything an agent needs, and **Copilot in GitHub** (cloud agent, code review) is first class, a scope change from spec §1. Efforts/reviews location still open (Summer wants to think it through). Per-repo context and synthesis are phase 1 |
| Marketplace tests | Done offline, 20 §7a. Claude Code: works end to end after folder trust. Copilot CLI: registers the marketplace from `.claude/settings.json` and installs the Claude-format plugin; auto-install, skill run and hook run not observed (no Copilot login here) |

## Next: research to do

### 0. Copilot in GitHub (new top priority, from Summer's answer 2)

**Done in session 12.** Docs: 20 §7b (GitHub docs read from a github/docs clone, since docs.github.com
was blocked). Hands-on: throwaway repo `sumsolution/folkways-copilot-probe` (results in 20 §7b; the
cloud agent committed its Copilot token in PR 2, so Summer was advised to delete the repo). Hook answers
from Summer are in 20 §7b "Hook delivery". The original brief follows for reference.

**Open after session 12:**

| # | Question | Who or what settles it |
|---|---|---|
| 1 | Copilot CLI with a login: does a repo-enabled plugin auto-install; do its skill and hook run (20 §7a rows 2, 4, 5)? | Local test with a Copilot login |
| 2 | Trust and marketplace prompts on a fresh machine; do new worktrees prompt again? | Local test on a clean setup |
| 3 | How `session-start`'s content reaches cloud runs (instructions, generated file, or a hook) | Summer to revisit |
| 4 | Server-managed settings for Claude Code cloud sessions in the beta org | Summer, if needed |
| 5 | Security acceptance: downloaded binary, GitHub App with PR write, read-only scanner | Beta company's security team |
| 6 | Enterprise `strictKnownMarketplaces` blocking repo-level marketplaces | Beta team's admin settings |
| 7 | Plugin update and pinning per surface (Copilot ignores repo-level `autoUpdate`) | Docs, then a test |
| 8 | Windows: no Windows hook binary today; Copilot ignores bash-only hooks on Windows | Design question |
| 9 | Efforts and reviews location; bots opening PRs into app repos | Summer (answers 4, 5) |
| 10 | Cost and quality of per-repo context generation | Trial on real repos |
| 11 | Hooks in Copilot code review (not observed) | Low priority |

What does each GitHub-hosted Copilot agent read from a repo? Primary docs, then a test on a throwaway repo
if Summer approves.

| Agent | Find out |
|---|---|
| Copilot cloud coding agent | Instruction files (`AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, `.github/instructions/*`), skills dirs (`.github/skills`, `.claude/skills`, `.agents/skills`), hooks (`.github/hooks/*.json`? `.claude/settings.json`?), MCP config, `copilot-setup-steps.yml` (can it install a hook binary?), whether it installs plugins or honors `enabledPlugins` |
| Copilot code review | Which instruction files it reads, any skills or path-scoped rules, size limits |
| Org custom instructions | Which of these agents they reach, precedence over repo files |

Also: Claude Code on the web (cloud sessions) with a repo-registered private marketplace: does it install?
Same for the Claude GitHub Action. These decide how much must be committed files vs plugin.

### 1. Verify the `~` claims (open the primary pages)

**Done in session 13**, results in 20 §7c. Same blocked sites as session 12. Still `~`: the two
github.blog changelog posts, Claude Code issues 30064 and 44224 (the docs now say the behavior is by
design), agentpatterns.ai, the dev.to posts and the Ruler page. New leads for step 2 from the seylox
sequel: devnewsletter's meta-repo write-up, the codified-context paper, `woditschka/agentic-coding-reference`.
The original brief follows for reference.

| Claim in 20 | Open |
|---|---|
| Copilot CLI plugins, marketplaces, `.claude-plugin/marketplace.json` support, repo-level marketplace registration | https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-plugin-reference and the about-plugins page; the Copilot CLI config docs |
| Copilot enterprise `managed-settings.json` (`.github-private`), keys, GA vs preview | https://docs.github.com/en/copilot/reference/enterprise-administrators/enterprise-managed-settings, the [GA changelog](https://github.blog/changelog/2026-07-01-enterprise-managed-settings-json-is-generally-available/) |
| Copilot org custom instructions: surfaces, precedence vs repo files, size limit, whether Copilot CLI reads them | the [changelog](https://github.blog/changelog/2026-04-02-copilot-organization-custom-instructions-are-generally-available/) and docs |
| Does Copilot CLI read `extraKnownMarketplaces` from `.claude/settings.json`? | CLI docs; then test (below) |
| Claude Code `--add-dir` skill loading and the open bugs | https://code.claude.com/docs/en/skills, anthropics/claude-code issues 30064, 44224 |
| Central-repo and meta-repo patterns | agentpatterns.ai central-repo page, seylox meta-repo post, RepoSwarm, Arvore Repo Hub, the dev.to multi-repo post |
| Cross-repo sync actions | BetaHuhn/repo-file-sync-action README (auth, PR mode, templating), netlify/file-sync-action |

### 2. Widen the prior art (new searches)

- Other "team context hub" implementations, and the octo upstream (`TheJeans/octo-team-context`; the
  "From Drift to Sync" WWC 2026 talk). Search found nothing; try GitHub search directly.
- Real teams' experience with central agent repos vs per-repo instructions: what they hit and why
  (retros, issues, talks). Section 8 Q1 needs evidence on the concerns.
- How Claude Code team plugins behave in practice: trust prompts, cache, `autoUpdate`, pinning a `ref`,
  hooks in plugins vs repo hooks (`CLAUDE_PROJECT_DIR`).
- GitHub App vs PAT for a hub with PR write on many repos; what security teams usually accept.
- Per-repo context generation: tools and costs (octo's loop, RepoSwarm, DeepWiki, `claude -p` with a
  fixed prompt).

### 3. Hands-on tests still open

Only with Summer's approval for anything that creates repos or pushes. Use throwaway repos and synthetic
names (`Example Team`, `example.com`). Session 11's offline fixture is described in 20 §7a; rebuild it the
same way.

1. Copilot CLI with a login: does `enabledPlugins` from `.claude/settings.json` auto-install at session
   start; does the plugin skill run; does the plugin `SessionStart` hook run with `CLAUDE_PROJECT_DIR`?
2. The interactive Claude Code trust and marketplace prompts on a fresh machine (what a teammate sees).
3. How hooks reach agents in a clone without `node_modules` (20 §0, last bullet).
4. Record each result in 20 (replace `~` with ✓ or a correction), with the version tested.

## How to work

- Keep findings in 20; mark confidence (✓ opened, ~ summary only, tested with version).
- Commit to `claude/determined-clarke-k9kpdh`; the PR is #37. Docs only, no changeset.
- Summer approves pushes outside this branch and any repo creation.
- **If docs sites are blocked** (session 12: docs.github.com, github.blog and WebFetch were; code.claude.com
  and GitHub clones worked): read GitHub's docs from source with a sparse clone of `github/docs`
  (`git clone --depth 1 --filter=blob:none --sparse https://github.com/github/docs.git`, then
  `git sparse-checkout set content/copilot data/reusables/copilot`). Fetch Claude Code pages as Markdown:
  `curl https://code.claude.com/docs/en/<page>.md`. Blog posts and third-party pages stay `~` if blocked.
- Never ask an agent under test to print its environment: the session 12 probe leaked a Copilot token
  that way. The probe repo has been deleted.

## Resume prompt

`Resume docs/design/21-handoff-app-repo-research.md on branch claude/determined-clarke-k9kpdh: read it and 20-research-app-repo-architecture.md (especially §7b), then start "Next" step 2 (widen the prior art). Steps 0 and 1 are done (§7b, §7c). Collaborate; don't decide.`
