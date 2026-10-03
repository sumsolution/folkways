---
name: tests
description: Writes and improves automated tests, test-first. Use when the user says "write tests", "add a test", "TDD", "test this", "reproduce this bug", "cover this with tests", "improve these tests", "why is this test failing" or "this test is flaky", or before fixing a bug or building a feature with acceptance criteria. Not for manual checks in an environment (use qa) or reviewing someone's change (use review).
---

# tests

Tests come first. A bug gets a failing test that reproduces it before any fix; a feature gets tests for
its acceptance criteria before the code. Tests describe behavior a user or caller can observe, so they
survive refactors and fail when the code is wrong.

## Config

- `folkways.yaml`: `repos` (default branches), `git.scopes`.
- `.local/me.yaml`: `projects_dir`.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## The law

**No production code without a failing test first.** Wrote code first? Set it aside (`git stash`),
write the test, watch it fail, then bring the code back only as far as the test demands. Never delete
the user's uncommitted work to follow this rule; stash it and say so.

## Workflow

1. **Find the repo and work in a worktree.** Code changes go in `<projects_dir>/.worktrees/<repo>/<branch>`,
   never in the user's clone. If you are in the user's clone, stop and say so. The branch follows the
   team's branch pattern (the `commit` skill).
2. **Read the house style first.** Read two or three tests next to the code you are testing: naming,
   layout, fixtures, helpers, assertion style, how the suite runs. Follow them (see
   [House style](#house-style)). Find the exact test command in the README, `AGENTS.md` or the CI config.
3. **State what to prove.**
   - **Bug:** the reproduction steps and the expected behavior. Ask for them if the report has none.
   - **Feature:** the acceptance criteria from the ticket, one test each. No criteria: write them
     as a short list, and confirm with the user before testing them.
4. **Red.** Write the smallest test for one behavior. Run it. Read the failure: it must fail because
   the behavior is missing or wrong, not because of a typo, an import error or a bad fixture. A test
   that passes at once proves nothing; fix the test, not the code.
5. **Green.** Write the least code that passes. No extras.
6. **Repeat, then edge cases.** After the acceptance criteria pass, add edge cases and the business
   rules that would cost most if wrong: boundaries, empty and null input, error paths, time zones,
   concurrency, permissions. One behavior per test, named for the behavior.
7. **Refactor with the tests green.** Run the whole suite, not only your new tests.
8. **Verify, then report.** Follow the [verification gate](#verification-gate). Say which tests you
   added, which failed first and why, and the command and result. Commits and pushes need the user's
   approval (the `commit` skill).

A failing test you were asked to diagnose ("why is this failing") is in scope. Find the root cause
before changing anything: read the failure, reproduce it alone, change one variable at a time. Decide
whether the test or the code is wrong; fix the one that is. After three failed fixes, stop and question
the design or the test itself, and tell the user. A bug in running code with no test yet starts at
step 3.

## Quality

- **Test behavior, not implementation.** Assert on outputs, state changes and calls across a boundary,
  not on private methods, call order or internals. A refactor that keeps behavior must not break tests.
- **Don't mock what you own.** Use the real code. Fake only what is slow, nondeterministic or outside
  the team's control (network, clock, third-party services), at the boundary. Prefer a simple fake over
  a mock library.
- **Would it fail if the code were wrong?** Check by breaking the code on purpose, once, for any test
  you doubt.
- **One reason to fail.** Independent tests, no shared state, no order dependence, no sleeps.
- **Coverage is a signal, not a gate.** Use it to find untested code, never to justify a test that
  asserts nothing. Don't chase a percentage.
- **Call out poor tests and bad patterns:** tests that mock the unit under test, assert only that
  nothing threw, duplicate the implementation, are skipped or commented out, or flake. If it is in your
  change's scope, fix it. If not, write it up with the `side-quest` skill and keep going.

## Rationalizations

| Excuse | Reality |
|---|---|
| "Too simple to need a test first" | Simple code breaks too. The test takes a minute. |
| "I'll add tests after" | A test written after passes at once and proves nothing about the need. |
| "I already wrote it and it works" | You checked by hand once. Stash it and let the test drive. |
| "Testing this is too hard" | Hard to test means hard to use. Simplify the design or the seam. |
| "The suite is slow, I'll run it at the end" | You will find breakage late. Run the one test now, the suite before you claim. |
| "It should pass" | Should is a guess. Run it. |

## Red flags: stop and start over

- Production code written before a failing test.
- A new test that passed on its first run.
- You can't say why the test failed.
- A test disabled, skipped or loosened to get green.
- "Just this once", "should work now", "probably fine".

## Verification gate

Before you say a test passes, or that work is done: run the exact command fresh, read the full output
and exit code, and check that it supports the claim (right tests ran, none skipped, count matches).
Report what you ran and what it printed. No "should pass". A line of another agent's report is not
evidence; run it yourself.

## House style {#house-style}

Default: follow the surrounding tests. With none to follow: one test file per unit, next to the code or
in the repo's existing test folder; names read as sentences describing behavior ("rejects an expired
token"); arrange, act, assert with a blank line between; data built by small helpers or factories, not
shared mutable fixtures; no logic (loops, conditionals) in a test body.

## Test types {#test-types}

Default: unit tests for logic and business rules; integration tests where behavior crosses a boundary
(database, HTTP handler, message queue) and a fake would hide the risk. Write the lowest level that
can show the behavior. End-to-end and browser tests are out of scope; manual verification is the `qa`
skill.

## Coverage {#coverage}

Default: report coverage for the files you changed if the repo already measures it, and mention
uncovered branches that matter. Set no threshold and never fail work on a number.

## Headless

No questions are possible. Work on a non-default branch in a worktree only; never on the default branch
or in the user's clone. Nothing needing approval happens: no push, no tracker or wiki writes. Missing reproduction steps or acceptance criteria: state the
assumption in the output and test that. A failing test you can't make pass is reported with the command,
the output and your diagnosis; never skip, disable or delete it. Flag every judgment call, and every
poor test you found but left alone.
