# @folkways/evals

A behavior-eval runner for folkways skills. Private; never published. Why we built it instead of adopting
a tool: [14-research-eval-harness.md](https://github.com/sumsolution/folkways-command/blob/14c47bd/efforts/review-rounds/14-research-eval-harness.md).

For each **arm** (a version of the framework) × **case** × **run**, it:

1. builds a team repo in the OS temp dir with that arm's own `folkways init --no-install`, then lays the
   case's `repo/` files over it and commits;
2. runs one headless `claude -p` there, isolated (below), and saves the answer and what the session loaded;
3. grades every run. `llm` votes go to a judge one answer at a time, shuffled across arms, with nothing
   that names the arm.

Suites live in [suites/](suites), one folder each. The runner finds this repo from its own location, so a
suite can also sit elsewhere, e.g. in a team repo.

## Run

Needs Node 24+, git, and an authenticated `claude` CLI. Every subject run and judge vote is a model call on
your account; `--dry-run` prints the counts first.

```sh
node packages/evals/src/cli.ts run packages/evals/suites/review-rounds --dry-run
node packages/evals/src/cli.ts run packages/evals/suites/review-rounds --jobs 6
node packages/evals/src/cli.ts run packages/evals/suites/review-rounds --model opus --runs 5 --arms after
node packages/evals/src/cli.ts grade packages/evals/suites/review-rounds     # re-grade after a rubric change
```

Results go to `<suite>/results/<model>/`: `summary.md`, `summary.json`, and per run `answer.md`,
`run.json` (the trace: answer, the agent's text and tool calls, skills and tools loaded, cost) and
`grades.json`. A rerun overwrites them.

## Suite layout

```text
<suite>/
├── suite.json                 # arms, default models
└── cases/<case>/
    ├── prompt.md              # frontmatter: run limits; body: the prompt
    ├── graders/<name>.md      # one grader per file
    └── repo/                  # optional: files laid over the team repo
```

`suite.json`:

```json
{
  "arms": [{ "name": "before", "ref": "28fbd63" }, { "name": "after", "ref": "WORKTREE" }],
  "model": "sonnet",
  "judge_model": "sonnet"
}
```

An arm's `ref` is any commit of this repo that has `packages/cli`, or `WORKTREE` for the working tree
(uncommitted changes included). Pin a SHA, not a branch, so the result can be reproduced.

The case format follows [`claude plugin eval`](https://code.claude.com/docs/en/plugin-evals), so cases read
the same in both. Unknown keys are errors.

| `prompt.md` key | Default | |
|---|---|---|
| `runs` | 3 | per arm |
| `model` | suite, then `--model` | subject model |
| `max_turns` | 30 | |
| `timeout_seconds` | 600 | |
| `allowed_tools` | `[Read, Glob, Grep, Skill]` | the only tools that exist in the run |
| `description` | | for people |

| Grader `type` | Keys | Passes when |
|---|---|---|
| `regex` | `pattern`, `flags`, `match` (`contains`, `not_contains`), `target` | the JavaScript regex is (or isn't) found |
| `tool_used` | `tool`, `input_match`, `min` (1), `max` | the count of matching calls is in range |
| `llm` | `focus`, `votes` (3); body = rubric | a strict majority of judge votes is PASS |

`target` and `focus` are `last_message` (default) or `transcript` (the agent's text and tool calls, not tool
results). Every grader takes `weight` (1). A run's score is the weighted share of graders passed.

Write rubrics as concrete PASS and FAIL conditions. The judge sees only the rubric and one answer.

## Isolation

| | |
|---|---|
| Team repo | built in the OS temp dir, outside any git repo, so no host `CLAUDE.md` or `AGENTS.md` loads |
| Settings | `--setting-sources project,local`: user settings, hooks and permissions don't load |
| MCP | `--strict-mcp-config`, no servers |
| Tools | `--tools` = the case's `allowed_tools` |
| Git | fixture commits ignore your git config (`GIT_CONFIG_GLOBAL=/dev/null`) |
| Judge | no tools, no settings sources, an empty directory |

Not isolated: your user memory (`~/.claude/CLAUDE.md`) and user skills still load. `summary.md` flags runs
where a team-repo skill didn't load or an MCP server did; check `run.json` `trace.skills` for extras. For
clean runs use a container or a cloud session. A case that grants `Bash`, `Edit` or `Write` runs them as
you, in the temp team repo.
