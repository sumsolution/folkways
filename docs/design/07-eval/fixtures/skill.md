---
name: commit
description: Write a commit message for a change, in the team's commit format.
---

# Commit

{{STEP0}}

## Step 1: read the change

Read the diff and the task. Work out what changed and why.

## Commit format {#commit-format}

Angular conventional commits.

- Header: `type(scope): subject`. Types: feat, fix, docs, refactor, test, chore, build, ci.
- scope: the module changed. subject: imperative, lowercase, no period, 72 chars max.
- Body: what changed and why, wrapped at 72.

## Step 2: write the message

Write the message in the commit format. Output only the message, in one fenced code block, then
one line on anything you left out and why. Don't run git.
