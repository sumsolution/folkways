---
name: qa
description: Verifies a change by hand in a running environment and records evidence (screenshots, traces, links) against the acceptance criteria. Use when the user says "QA this", "test it in dev", "check it works locally", "verify the fix is deployed", "click through it" or "does it work in the browser". Not for writing automated tests (use tests) or reviewing code (use review).
---

# qa

Manual verification of a change in `local` or `dev`, criterion by criterion, with evidence someone else
can check. End-to-end test writing is out of scope for v1.

## Config

- `folkways.yaml`: `qa.allowed_envs`, `repos`.
- `.local/me.yaml`: `projects_dir`.
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

1. **Load the criteria.** The ticket's acceptance criteria, or the PR's "How to verify". No criteria →
   ask the user what "working" means before testing anything.
2. **Pick the environment.** Environments in `qa.allowed_envs` (default `local`, `dev`) need no
   approval. Any other environment needs explicit approval for this session, every time. Never run
   anything that writes data in a shared environment without approval.
3. **Confirm the build.** Check the environment runs the SHA under test (a version endpoint, deploy
   log, or build info). Record it. Testing the wrong build is the most common QA failure.
4. **Verify each criterion.** Drive the browser, API or CLI through the criterion's Given / When. Capture
   evidence per [Evidence](#evidence). Try the obvious edge next to each criterion (empty input, a
   second click, an expired session).
5. **Write the QA notes** per [Output](#output) and show them. Found a defect → describe it as a bug
   (the `ticket` skill's bug shape) in the notes; create a ticket only on approval.

## Evidence {#evidence}

- A screenshot or recording for UI behavior; a request and response (headers trimmed of secrets) for
  APIs; a trace or log link for backend behavior.
- Every link carries its environment and time window.
- Never capture secrets, tokens or personal data. Use test accounts.

## Output {#output}

`<effort>/NN-qa-<env>-<sha7>.md`; with no effort,
`<personal root>/inbox/<YYYY-MM-DD>-qa-<env>-<sha7>.md`:

```markdown
# QA: EX-123 in dev

**Build:** web-app@1a2b3c4 (from /version) · **Env:** dev · **Date:** 2026-09-29 · **Result:** 1 of 2 pass

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Expired token returns 401 | pass | screenshots/01.png, trace link |
| 2 | Refresh retries once | **fail**: retries forever | log link (10:02-10:05 UTC) |

## Defects
<bug write-ups, if any>
```

## Headless

Only environments in `qa.allowed_envs`, read-only actions, no approvals possible. Write the QA notes;
mark any criterion you couldn't verify as `not verified` with the reason.
