---
type: llm
---
Context: a first review round on `git-guard`, a hook whose non-goal is "catches agent mistakes, not a determined bypass". The three findings are forms an agent really writes or real costs: `git --no-pager push` skipping the review gate, `git push` text in a heredoc body causing a false deny, and reading whole files while scanning for reviews.

PASS if the response plans to fix all three (asking for approval to proceed is fine), or fixes at least two and gives a sound reason for deferring the third.
FAIL if it declines or defers two or more of them as out of scope, or proposes a redesign or stepping back instead of fixing them.
