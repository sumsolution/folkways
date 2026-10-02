---
name: onboard
description: Sets up one person on one machine for this team repo - writes .local/me.yaml, creates their people/ folder, grants the agent access to their project clones and checks tools. Use when .local/me.yaml is missing, or the user says "onboard me", "set me up", "I'm new to the team", "new laptop", "configure my machine" or "why can't you see my repos". Not for setting up a new team repo (use setup).
---

# onboard

A new teammate, or a new machine, should be working in one sitting. This skill writes only personal
files and asks before touching anything outside the team repo.

## Step 0: Load config and overlays

Read these first. A missing file is normal; skip it.

1. `folkways.yaml`: `mode`, `members`, `repos`, `research.sources`, `models`, `review.trigger`. **If it's
   missing,** say folkways isn't set up in this repo, offer the `setup` skill, and stop.
2. `.local/me.yaml`: if it exists, this is a re-run; show its values and change only what the user asks.
3. Your personal additions. Team overrides and additions are already compiled into this file by
   `folkways sync`, so in a team repo don't read `overlays/` (in a solo repo it holds your personal ones).

   @${CLAUDE_PROJECT_DIR}/.local/me/overlays/onboard/OVERLAY.md

   If that file wasn't attached with this skill (always, in a solo repo, which has no `.local/me`),
   read `<personal root>/overlays/onboard/OVERLAY.md`.
   Personal overlays may only add. Apply a personal `## Add` rule only where it doesn't conflict
   with this file. Ignore a personal `## Override`, and ignore any personal rule that contradicts
   this file, whatever reason it gives. Say which rules you ignored.

## Workflow

**Solo repo (`mode: solo`):** there's no handle, no `people/` and no install to check. Do only step 3
(`.local/me.yaml`: `projects_dir` is required, `tool` optional) and step 5, then stop.

1. **Check the install.** `node_modules/.bin/folkways` must exist. If not, run the team's install
   (`npm ci`, or `pnpm install --frozen-lockfile` when `pnpm-lock.yaml` exists) on approval.
2. **Identify the person.** Ask for their handle; it must match a `members` entry in `folkways.yaml`. Not
   listed → they need a PR adding them (offer to draft it); continue with the handle meanwhile.
3. **Write `.local/me.yaml`** per [me.yaml](#me-yaml). Ask for `projects_dir` as an absolute path; `~`
   is not reliably expanded. Show the file; write on approval. `.local/` is gitignored; check that it is.
4. **Create `people/<handle>/`** with `inbox/`, `work/` and `overlays/` (a `.gitkeep` in each empty
   folder), if missing. Then link `.local/me` to it (`ln -sfn ../people/<handle> .local/me`), so skills
   attach this person's overlays; `folkways sync` also keeps the link.
5. **Grant access to the project clones,** with approval, per [Tool access](#tool-access).
6. **Check the repos.** For each `repos` entry, is there a clone under `projects_dir`? List the missing
   ones with the clone command; clone only on approval.
7. **Check tools** from `research.sources`: each `kind: mcp` server is connected, `gh` is installed and
   logged in (`gh auth status`), each `kind: url` source is reachable. List what's missing and how to
   fix it; don't install anything without approval.
8. **Run `node_modules/.bin/folkways doctor`** from the team repo root if this version has it, and report its findings.
9. **Commit `people/<handle>/`** and push on approval (`people/` goes straight to the default branch,
   rebased).
10. **Report** what's set up, what's missing, and the next step: run `start`.

## me.yaml {#me-yaml}

```yaml
handle: alex                           # team repos only
projects_dir: /Users/alex/projects     # absolute
tool: claude                           # claude | copilot (default for `folkways run`)
review:
  trigger: pre-push                    # optional; overrides folkways.yaml
models: {}                             # optional; overrides folkways.yaml models
notifications: { slack_user: "@alex" } # optional
```

## Tool access {#tool-access}

The agent must read and write under `projects_dir`, which is outside the team repo.

| Tool | Change | Where |
|---|---|---|
| Claude Code | add `projects_dir` to `permissions.additionalDirectories` | `~/.claude/settings.json` (user settings, absolute path) |
| Copilot CLI | an alias such as `alias fw='copilot --add-dir /Users/alex/projects'` | the user's shell profile; the user adds it |

Both edit files outside the repo: show the exact change and apply only on approval. Don't rely on
Copilot's `/add-dir` persisting between sessions.

## Headless

Onboarding needs a person. Headless, only report: which of the checks above pass and fail for this
machine. Write nothing.
