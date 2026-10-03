---
name: review
description: Reviews a code change against its ticket and writes a verdict and findings file. Use when the user says "review this", "review my PR", "code review", "look over this diff", "is this ready to merge", "check this branch", or when `folkways review` runs the fresh-process review gate. Code only, not ADRs or other documents (use adr), writing tests (use tests) or fixing the code.
---

# review

Review a change the way a careful senior teammate would: know what it is meant to do, check that it
does it, prove risky claims by running the code, and say plainly what must change. Every teammate's
agent gives the same baseline review; team overlays add team rules.

## Config

- `folkways.yaml`: `repos` (default branches, owners), `tracker`, `review.trigger`.
- `.local/me.yaml`: `projects_dir`.
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Inputs

You need the repo, the branch or PR, the base, and the HEAD SHA to review. Ask for any you can't find.
The ticket and effort path are optional but make the review much better; ask once.

When `folkways review` started you, the prompt gives all of these and you have no context from the
author. Keep it that way: don't read the author's session notes as a substitute for the ticket.

## Workflow

1. **Load intent before the diff.** Read the ticket, the effort's `00-STATUS.md` and design docs, any
   ADRs it links, and `knowledge/` entries on the touched area. Write down in two or three lines what
   the change must do, what it must not break, and its non-goals. If there is no ticket, review
   against the PR description and note the missing ticket as a process finding.
2. **Pin the SHA.** Review exactly one SHA. Check it out in a detached worktree
   (`git worktree add --detach <projects_dir>/.worktrees/<repo>/review-<sha7> <sha>`), never in the
   user's clone. Record the SHA in the output.
3. **Read the diff against intent.** Does it do what the ticket asks, all of it, and nothing else?
   Then read for correctness, failure paths, security, data handling, and tests (see
   [What to look for](#what-to-look-for)).
4. **Verify by running.** Run the test suite at the SHA. For each risky claim (a race, an edge case,
   an error path), write a throwaway test or script and run it. When docs and source disagree, the
   source is the citation. Delete throwaway code; report what you ran. When the review file is written, remove the worktree
   (`git worktree remove <path>`).
5. **Write the review** in the [Output](#output) format. Verdict first. Every finding cites evidence:
   `path:line` at the SHA, a command and its output, or a doc link.
6. **Draft locally, post only on approval.** Save the file; show the verdict and findings. Offer to post
   comments on the PR. Posting is the user's call: it appears to come from them.

## What to look for {#what-to-look-for}

- **Intent:** missing acceptance criteria, scope creep, behavior the ticket didn't ask for.
- **Correctness:** edge cases, error handling, concurrency, off-by-one, time zones, null and empty input.
- **Security and data:** input validation, authz on new paths, secrets, logging of personal data.
- **Tests:** do they test behavior, and would they fail if the code were wrong? Flag tests that mock
  the unit under test or assert on implementation details.
- **Standards:** follow the repo's patterns, README and `AGENTS.md`. If a repo pattern is outdated, say so,
  keep the PR's scope, and suggest a follow-up ticket (the `side-quest` skill writes it).
- **Size:** friction shows in review rounds and time open more than in line count. If the change
  mixes independent outcomes, suggest a split.
- **Proportion:** is the complexity in line with the intent and its stated non-goals? A finding that
  falls under a stated non-goal is a `question` or `thought`, not a blocking `issue`. After two or more
  review rounds on the same area, review the design, not the diff: say whether the approach still fits.

## Findings format {#findings-format}

Label each finding with [Conventional Comments](https://conventionalcomments.org/):
`praise`, `nitpick`, `suggestion`, `issue`, `todo`, `question`, `thought`, `chore`, `note`, plus a
decoration:

| Decoration | Means |
|---|---|
| `blocking` | Must change before merge |
| `before-qa` | Must change before QA or release, may merge behind a flag |
| `decision` | Needs a call from the owner; state the options |
| `non-blocking` | Author's choice |

`nitpick` is always non-blocking. Include at least one `praise` when something is done well; be
specific about what.

## Verdict {#verdict}

One of:

- **Ready:** no blocking findings.
- **Ready after fixes:** blocking findings are small and clear; no re-review needed.
- **Changes needed:** blocking findings need a re-review.
- **Needs a decision:** a `decision` finding blocks progress.

## Output {#output}

Write `<effort>/NN-review-<sha7>.md`, where `NN` is the effort's next topic number. With no effort,
use `<personal root>/inbox/<YYYY-MM-DD>-review-<sha7>.md`. A re-review is a new file, never an edit of the old one.

```markdown
---
sha: <full SHA>
---
# Review: <PR or branch title>

**Verdict:** Changes needed. <one sentence why>
**Reviewed:** <repo>@<sha7> against <base>, <date>. **Intent:** <ticket link or "no ticket">.
**Ran:** `<test command>`: <result>. <throwaway checks, one line each>

| ID | Label | Where | Finding | Evidence |
|---|---|---|---|---|
| 1 | issue (blocking) | `src/auth.ts:42` | Expired tokens pass validation | Throwaway test: `exp` in the past returns 200 |
| 2 | praise | `src/auth.test.ts` | Clear table-driven cases | — |

## Notes
<only what the table can't hold; link follow-up tickets>
```

Tone: friendly, direct, succinct. Say what to change and why in one or two sentences per finding.
Metrics carry provenance: environment, time window, and the date you read them.

## Headless

You can't ask questions or post anything. Missing inputs: review what you have and say what was
missing in the header. Write the review file and nothing else. Flag every judgment call as a
`decision` finding with the options, so a reviewer can settle it.
