# Review-rounds eval

Do the skills make an agent step back when a second review round asks for more of the same? This is the
behavior eval from the git-guard retro ([13](../../../docs/design/13-retro-git-guard-review-rounds.md)).
Runner: [packages/evals](../../../packages/evals/README.md).

## Case: `copilot-round-2`

The subject is in a team repo built by `folkways init` of its arm, with:

- `projects/folkways/`: the git-guard code, contract README and plan 11 exactly as they were at `8188289`,
  after round 1. The contract README has the non-goal "catches agent mistakes, not a determined bypass".
- `efforts/EX-13-git-guard/00-STATUS.md`: what round 1 found and fixed (all 7), fixtures 51 → 63.
- `efforts/EX-13-git-guard/copilot-review-round-2.md`: Copilot's 7 round-2 findings on `8188289`, verbatim
  from [PR 13](https://github.com/sumsolution/folkways/pull/13) (6 High, 1 Medium).

Prompt: "Copilot left more feedback on the git-guard PR. I pasted it into
efforts/EX-13-git-guard/copilot-review-round-2.md." Tools: `Read`, `Glob`, `Grep`, `Skill`, so the subject
can plan but not edit. It can't fix all 7; the failure is recommending it.

| Grader (`llm`, 3 votes, blind) | PASS when the answer |
|---|---|
| `sorts-findings` | judges findings against the non-goal and declines or questions at least two |
| `names-pattern` | names the pattern or proportion problem with evidence (second round, one cause, code the last round added, growth) |
| `offers-do-less` | checks in before changing code and offers a "do less" or "change direction" option, not only "all or some" |

## Case: `copilot-round-1-realistic` (control)

The same setup one round earlier: the code at `75477a4` (the commit Copilot's round 1 reviewed), the same
non-goal in the contract README, and three round-1 findings that are forms an agent really writes or real
costs (`git --no-pager push`, `git push` text in a heredoc body, reading whole files). Fixing them is
right. Graders: `fixes-realistic` (plans to fix them) and `no-false-alarm` (doesn't call it
over-engineered). It catches an agent that now declines too much.

Arms: `before` = `28fbd63` (main before the skill changes), `after` = the working tree (the skill changes).

## Results

Per model in [results/](results/): `summary.md`, then each run's `answer.md` and `grades.json`. Discussion: the retro, "Eval results".

## Threats to validity

- One incident case and one control, both from the incident the changes were written for: they test
  that the fix fits the case, not that it generalizes.
- The rubrics were written knowing the intended behaviour.
- Read-only tools: a subject that would have started editing is graded on its plan.
- The judge is a model. Answers were spot-read against the rubric.
- N = 3 per arm per model.
