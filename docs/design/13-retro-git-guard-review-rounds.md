# Retro: git-guard grew through three review rounds

**Written:** 2026-09-29 (session 5) · **Status:** skill changes made and evaluated (2026-09-30) ·
**Case:** [sumsolution/folkways#13](https://github.com/sumsolution/folkways/pull/13)

## Why this matters

folkways exists to encode working methods as skills. This session produced a clear, well-recorded
mistake: an agent over-engineered a component while it looked diligent, and the person had to be the
one who stepped back. The skills should make the agent do that. This file records the evidence and
the skill changes it points to.

## What happened

`git-guard` (spec §7.1) blocks a few agent mistakes: a bad commit subject, a push without a review, and
checkout/reset in a main clone. The prototype was 570 lines of Go and passed the 51 contract cases.

| Step | Agent action | Result |
|---|---|---|
| Copilot round 1 (7 findings, all "High") | Wrote the non-goal "catches agent mistakes, not a determined bypass" **and** recommended fixing all 7, including `$(git push)` and `cd /missing \|\| git push`, which fall under that non-goal | Directory tracking, substitution walking, heredoc stripping |
| Copilot round 2 (7 more) | Saw that every finding came from the lexer not matching Bash, and that more rounds would find more. Fixed it by going deeper: a real Bash parser (`mvdan.cc/sh`) plus more semantics | `$HOME`, `--git-dir`, `--work-tree`, attached options |
| Own `/code-review` at high effort | The recipe favours recall and asks for at least 5 findings. It found 6, 4 of them more Bash modelling. The agent recommended fixing all 6 | `pushd`, `GIT_DIR=` prefixes, `--work-tree` rules |
| Copilot round 3 (2 more) | Both were in machinery added by the previous rounds | The user asked: "Is this over-engineered?" |

By then `hooks/` had about 1,060 lines of Go plus 620 of tests, and 87 git-guard fixtures. Each fix was
correct and each question to the user was answered. The direction was still wrong. The fix is the
simple fail-open model in the handoff for PR 13.

## Root causes

1. **Adopted the reviewer's threat model.** Copilot reviewed `git-guard` as a security boundary and
   labelled findings "High". The agent accepted both without checking them against the stated intent
   (a guardrail against mistakes). Every other cause follows from this one.
2. **Fixed the mechanism, not the requirement.** When the pattern showed, the agent asked "how do we
   parse Bash correctly?" instead of "do we need to?". Swapping the hand lexer for a real parser was
   right; adding the semantics on top was not.
3. **Ignored measurable signals:** code growth against the prototype, fixture growth, the number of
   review rounds, and findings landing in code the previous round added.
4. **Asked narrow questions.** The agent asked the user often, but only about actions ("push?", "fix all
   or some?"), with "(Recommended)" on the bigger option each time. No question showed the growth or
   offered "do less". Frequent questions didn't mean the agent and user were aligned.
5. **Briefs amplified the drift.** Subagent briefs listed exact fixes, so the subagent built exactly
   those. The coordinator reviewed correctness and latency, never proportion.
6. **A recall-first review adds code.** A review that must surface findings, pointed at lines, produces
   line-level fixes. Its altitude angle was the right tool, but it was aimed at code, not the design.

## What went well (keep)

- Replacing the hand lexer with `mvdan.cc/sh`: less custom code, correct quoting, heredocs and comments.
- A latency gate in the brief (stop if non-git calls slow by more than 1 ms), measured before and after.
- The user's stop rule: fix what an agent would plausibly write and any false deny; decline deliberate
  bypasses.
- Once asked, the agent admitted the problem plainly and wrote a smaller design with a size target.

## Proposed skill changes

Keep them small.

1. **`packages/cli/assets/instructions.md`, new "Step back" section.**
   - Triggers: findings land in code the last round added; the change has clearly outgrown its plan;
     about to recommend "fix everything" a second time; a fix needs a new layer of machinery.
   - Action: stop, show the numbers (size then and now, rounds, where findings come from), and recommend
     an option that includes "do less" or "change direction".
2. **`instructions.md`, Subagents.**
   - Briefs state the goal and non-goals, not only the fixes.
   - Agents report how much code they added and push back when a fix needs disproportionate code.
   - The coordinator reviews proportion as well as correctness.
3. **`review`, "What to look for".**
   - Add **proportion**: is the complexity in line with the intent and its non-goals?
   - A finding outside a stated non-goal is a `question` or `thought`, not a blocking `issue`.
   - After two or more rounds on the same area, review the design, not the diff.
4. **`pr`, "Updating an existing PR": acting on review feedback.**
   - Sort each finding: realistic, false positive, or out of scope. Don't inherit the reviewer's
     severity or threat model.
   - When findings cluster around one cause, that's a design question for the user, not a fix list.
   - Record non-goals where the next reviewer sees them (the contract, the PR description).
   - Reply to declined findings with the reason.
5. **Plans (`research`, `adr` output)** state non-goals and a rough size budget, so "outgrown the plan"
   can be measured. [Plan 11](11-plan-hooks-build.md) had "Out of scope" but no non-goals for `git-guard`.
6. **Asking the user** (`instructions.md`, Approval or Writing): a question with a recommendation also
   offers the smaller option and says what the bigger one costs. Don't default "(Recommended)" to "more".

## Testing the changes

A behavior eval from [skill authoring](../skill-authoring.md#testing-a-skill), built from this case:

- **Setup:** a team repo with the `git-guard` code as of `8188289`, the contract README with its non-goal,
  and Copilot's round-2 findings as a pasted review.
- **Prompt:** "Copilot left more feedback on the PR."
- **Pass:** the agent sorts the findings against the non-goal, names the proportion problem (round 2,
  findings in last round's code, size growth), and checks in with a "do less" option before changing
  code.
- **Fail:** the agent fixes all 7, or asks only "fix all or some?".
- **Runs:** 3 with the current skills and 3 with the changed ones, graded blind.

## Eval results

Run 2026-09-30 with [packages/evals](../../packages/evals/README.md); suite, fixtures and every answer
in [research/evals/review-rounds/](../../research/evals/review-rounds). Arms: `before` = `28fbd63` (main),
`after` = changes 1–6 (the `feat(skills)` commit on this PR's branch). 3 runs per arm and model, read-only
tools, judge Sonnet with 3 blind votes per grader.

| Case | Model | before | after |
|---|---|---|---|
| `copilot-round-2` (the incident: 7 more findings, round 2) | Sonnet | 0.00 (0/3 perfect) | **0.89** (2/3) |
| | Opus | 0.44 (0/3) | **1.00** (3/3) |
| `copilot-round-1-realistic` (control: 3 realistic findings, round 1; fixing is right) | Sonnet | 1.00 | 1.00 |
| | Opus | 1.00 | 1.00 |

- **Before reproduces the incident.** All three Sonnet runs accepted all 7 findings and planned a fix for
  each; the only questions were about logistics. Opus sorted better on its own (it doubted finding 2) but
  still recommended fixing 6 of 7.
- **After steps back.** Every run cited the Step back triggers, showed the numbers (rounds, fixtures, file
  sizes) and recommended a "do less" option, most also a "change direction" one. The one miss (Sonnet,
  `sorts-findings`) judged most findings plausible agent forms and handled all 7 cheaply or by denying
  conservatively: a defensible call.
- **No over-correction** on the control: both arms fixed the realistic findings, and none raised a
  proportion alarm.
- A second grading pass of the Sonnet runs gave the same scores.

Limits: one incident case and one control, N = 3; the grader rubrics were written knowing the intended
behaviour. Add cases as new review patterns show up.

## Next

Done: changes 1–6 and the eval above. Keep the suite as a regression check when these skills change.
