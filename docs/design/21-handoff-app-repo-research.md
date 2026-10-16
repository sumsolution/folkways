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
| Questions to Summer (§8) | Asked, **not answered yet**. Don't assume answers |

## Next: research to do

### 1. Verify the `~` claims (open the primary pages)

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

### 3. Hands-on tests (unknowns 1–3 in 20 §7)

Only with Summer's approval for anything that creates repos or pushes. Use throwaway repos and synthetic
names (`Example Team`, `example.com`).

1. A marketplace repo holding one tiny team plugin (a skill and a `SessionStart` hook), and an app repo
   whose `.claude/settings.json` registers it. Fresh clone; record the Claude Code trust and install flow
   and whether the hook sees `CLAUDE_PROJECT_DIR`.
2. Same app repo with Copilot CLI: does it auto-register the marketplace from repo settings? From which
   file?
3. Record each result in 20 (replace `~` with ✓ or a correction), with the version tested.

## How to work

- Keep findings in 20; mark confidence (✓ opened, ~ summary only, tested with version).
- Commit to `claude/determined-clarke-k9kpdh`; the PR is #37. Docs only, no changeset.
- Summer approves pushes outside this branch and any repo creation.

## Resume prompt

`Resume docs/design/21-handoff-app-repo-research.md on branch claude/determined-clarke-k9kpdh: read it and 20-research-app-repo-architecture.md, then start "Next" step 1 (verify the ~ claims). Collaborate; don't decide.`
