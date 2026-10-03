---
name: research
description: Investigates a question across the team's sources and writes a findings file with cited evidence and confidence. Use when the user says "research", "look into", "investigate how X works", "find out", "what does service Y do", "discovery", "compare options for" or "spike", or when starting an effort that needs preliminary findings. Not for writing a decision record (use adr) or tracing a bug in running code (use tests to reproduce it).
---

# research

Answer a question well enough that the user can decide with little or well-understood risk. Read the
team's own knowledge and precedent before the web, check every claim against a source, and say how sure
you are. The user makes the call.

## Config

- `folkways.yaml`: `research.sources`, `repos`, `docs`, `models`.
- `.local/me.yaml`: `projects_dir`, `tool`, `models`.
- The personal root, for `<personal root>` paths.

The session-start hook puts these values in the session context; read `folkways.yaml` and `.local/me.yaml` only if they aren't there.

## Workflow

1. **Frame the question.** Write the question, what decision it feeds, and what would count as enough,
   in two or three lines. Ask if you can't tell. Note the effort folder (from the `start` skill) or use
   `<personal root>/inbox/`.
2. **Plan the sources.** Take them in the [source order](#source-order); drop those that don't apply to
   this topic and say which. Check each tool you need is available (see [Tools](#tools-and-fallbacks)).
3. **Search, most authoritative first.** Stop when the question is answered with enough confidence.
   Record for each source what you searched and what you found, including a negative result.
4. **Fan out when it pays.** See [Subagents](#subagents). Independent lookups run in parallel; a
   question that needs one chain of reasoning stays with you.
5. **Check citations.** Open a sample of what subagents cited (all of it for any claim a decision rests
   on). A claim that doesn't hold up is dropped or re-researched, and the failed check triggers
   [escalation](#escalation).
6. **Write the file** in the [output](#output) format. Every claim links to evidence; mark inferences.
7. **Report back briefly:** the answer, the confidence, known risks, open questions, and where the file
   is. Say whether it is enough to decide. Don't paste the file.
8. **Offer to lift durable findings.** See [Knowledge](#knowledge).

## Source order {#source-order}

Use the order in `folkways.yaml` `research.sources`; each entry has a `name`, a `kind` (`local`, `repos`,
`url`, `mcp`, `web`) and notes. Without config, use: `knowledge/`, the team's repos, other teams'
repos, internal docs and API portal, docs wiki, tracker, monitoring, chat, then the web.

- `knowledge/` first: it may already answer the question, or say what is stale.
- **Team and org precedent before the web:** ADRs, architecture and team docs, existing code. Web
  results are for what the team's own sources don't cover, and read them last.
- Skip what doesn't apply; don't touch monitoring for a library comparison.

## Tools and fallbacks {#tools-and-fallbacks}

| Need | Use | If unavailable |
|---|---|---|
| Understand code in one repo | Local clone (see [clones](#clones)) | Ask |
| Search many repos | `gh search code`, or local clones | Local clones, else ask |
| Any source with `kind: mcp` | That server from `folkways.yaml` | Ask the user to paste, export or link |
| A source with `kind: url` | Fetch it; note any `notes` (VPN, login) | Ask the user to open it and paste or link |
| External docs and practice | Web, after team sources | Say the web was not searched |

**An unavailable tool means ask the user.** Never skip a source the question depends on without
saying so. If the user can't help, list the source under open questions and lower the confidence.

## Clones {#clones}

The user's clones under `projects_dir` belong to the user.

- Never check out, pull, reset or rebase in them.
- Run `git fetch`, then read `origin/<default>` (default branch from `folkways.yaml` `repos`) with
  `git grep`, `git show origin/<default>:<path>`, or, for a deep read, a detached worktree at
  `<projects_dir>/.worktrees/<repo>/research-<sha7>`, removed with `git worktree remove` when you
  finish. Record the full SHA you read.
- **Missing repo:** clone it into `projects_dir` when the search is deep (many files) or the repo will
  be referenced again. Cloning writes to the user's disk: tell them the repo and path and clone on their
  yes. A repo listed in `folkways.yaml` `repos` is pre-approved; say so as you clone. Then the same rules apply.

## Subagents {#subagents}

Use them to fan out independent lookups (one per repo, source or option).

- **Cheap model first** (`models.<tool>.cheap`); the main agent orchestrates and reviews.
- **Each gets a self-contained prompt:** the question, the one source to search, the clone rules, the
  output format (short summary, each claim with a link or `path:line` at a SHA, confidence, what it
  could not check). It has none of your context.
- **You review** what comes back and spot-check citations. No subagent writes the research file.
- If no subagents are available, do the lookups yourself, one after another.

## Escalation {#escalation}

Re-run or re-check with the strong model (`models.<tool>.strong`), or read the evidence yourself, when:

- sources conflict,
- a claim a decision depends on is at stake,
- a sampled citation fails your check,
- confidence is low.

Say in the file that you escalated and what changed.

## Output

Write `NN-<topic>.md` in the effort folder, `NN` being the next topic number (`00-STATUS.md` is 00).
With no effort, use `<personal root>/inbox/<YYYY-MM-DD>-research-<slug>.md`. Use the template in
[references/output-template.md](references/output-template.md)
(`.claude/skills/research/references/output-template.md`); read it before writing.

- **Every claim links to evidence:** a docs URL, code as `path:line` at a SHA, a ticket or page. Cite
  every number: a number without a link is an assertion.
- **Each finding gets a confidence:** high (read at the source, or two sources agree), medium (one
  good source, or a source that may be stale), low (inferred, or secondhand).
- **Mark inferences** as inferred, with what they rest on. Don't state them as fact.
- **Quotes are short.** Paraphrase and link.
- **List every SHA read** per repo, and the date, so a reader can tell what is stale.
- **Metrics carry provenance:** environment, time window, and the date you read them.
- **Say what you didn't find** and which sources you skipped, and why.

## Done

Enough to decide with little risk or well-understood risk. Give the user the decision options, your
confidence and the known risks, then let them call it. Don't keep researching to raise confidence past
what the decision needs.

## Knowledge

`knowledge/` is the team's durable memory, and it lands by PR. Whether research findings get lifted
there during research or at session close is undecided. So: when you find a clearly durable fact (it
won't change next month, and others will need it), mention it in your summary. At session close, offer
the candidates. Write to `knowledge/` only on the user's explicit approval.

## Headless

No questions are possible. Write the research file and nothing else: no external writes, no
knowledge changes, no clones outside `projects_dir`. A source that needs a tool you don't have is
skipped and listed under open questions, the confidence lowered. Flag every low-confidence finding
and every judgment call in the file's header so a reviewer can settle them.
