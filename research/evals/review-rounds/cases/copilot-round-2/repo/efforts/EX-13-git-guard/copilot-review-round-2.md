# Copilot review, round 2 (on 8188289)

Pasted from PR #13. Copilot's summary: "Several supported command forms can bypass git-guard checks, and
the shell parser can also deny a command that would not run." Findings: 6 High, 1 Medium.

## 1. High: Resolve cd targets including $HOME and tilde paths (`hooks/internal/gitguard/guard.go:142`)

A normal `cd "$HOME/projects/app-wt" && git push` from the team repo changes into the worktree in Bash, but this branch ignores the `cd`. The guard checks the team directory instead of the worktree; it can miss the push when it cannot read a HEAD there, and the same pattern misses a protected main clone. Resolve common `$HOME` and `~` directory forms or conservatively handle an unknown target instead of treating the original directory as the successful `cd` destination; add both forms to the guard tests.

## 2. High: Support attached -C and -c global Git options (`hooks/internal/gitguard/guard.go:187`)

Git accepts attached short global options. `git -C/path/to/worktree push` leaves `-C/path/to/worktree` as the subcommand here, so the review check never runs; `git -ccore.foo=bar reset --hard` similarly bypasses the main-clone check. Consume the attached `-C` operand while updating `dir`, and consume an attached `-c` setting before dispatch; cover both forms in the guard tests.

## 3. High: Resolve Git directory options before validating repository targets (`hooks/internal/gitguard/guard.go:190`)

`--git-dir` and `--work-tree` are skipped without changing the directory passed to either guard. From a team repo, `git --git-dir=/projects/app/.git --work-tree=/projects/app reset --hard` operates on the main clone, but `checkMainClone` inspects the team repo and permits it. A push targeting another repository can likewise be checked against the wrong HEAD (or fail open outside a repository). Resolve these options together with `-C` before checking the effective Git target; add regression cases for both guards without `-C`.

## 4. High: Parse attached commit message options (`hooks/internal/gitguard/guard.go:217`)

Bash combines a short option and an adjacent quoted argument: `git commit -m"bad subject"` and `git commit -am"bad subject"` become `-mbad subject` and `-ambad subject` in `words`. Neither matches here, so `commitMessage` reports no message and the invalid subject bypasses the commit check. Extract attached `-m`/`-am` values and test both forms.

## 5. High: Inspect command substitutions in unquoted heredocs (`hooks/internal/gitguard/shell.go:279`)

An unquoted heredoc expands command substitutions before handing text to its command. `cat <<EOF` with `$(git push)` on a body line executes that push, but removing the whole body here hides it from the guard. Inspect executable substitutions in unquoted heredoc bodies while leaving plain body text and quoted-delimiter bodies as data; test both cases.

## 6. High: Handle backslash-newline continuations consistently (`hooks/internal/gitguard/shell.go:323`)

Bash removes a backslash-newline before splitting words, but `words` appends the newline to the following word. In `git ` followed by a backslash-newline and `push`, `parseSeq` keeps the command together while `checkGit` sees a subcommand containing a newline; an unreviewed push passes. Handle line continuations consistently in sequence and word parsing (also for `cd` targets), and add a multiline push regression case.

## 7. Medium: Recognize shell comments before command splitting (`hooks/internal/gitguard/shell.go:162`)

Shell comments are not recognized before splitting on `;` and `&&`. With `echo ok # note; git push`, the shell only runs `echo`, but this parser treats `git push` as another command and can deny a harmless tool call. Skip unquoted `#` comments from a token boundary through the newline, without skipping commands on the next line; add an inline-comment regression test.
