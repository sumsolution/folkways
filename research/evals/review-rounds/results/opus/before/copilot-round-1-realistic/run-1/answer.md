I checked all three findings against the code. Each one is real. In two places the right fix goes further than Copilot suggests.

I read the files in the working tree of `projects/folkways`, not with `git show 75477a4`, because I have no shell in this session. So I haven't confirmed that the checkout is at 75477a4. The line numbers do match Copilot's.

## 1. Global options skip the guard: agree, High

`guard.go:66` only strips `-C` and `-c`. So in `git --no-pager push`, the subcommand is read as `--no-pager` and nothing is checked. Agents often add `--no-pager` by habit, so this lets ordinary mistakes through. It isn't a deliberate bypass, so it counts under the contract (`test/hooks/README.md:81`).

**Fix:** skip git's global options before reading the subcommand:
- Options without a value: `-p`, `-P`, `--no-pager`, `--paginate`, `--bare`, `--no-replace-objects`, `--literal-pathspecs`, `--no-optional-locks`, and similar.
- Options with a value: `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`, `--config-env`, in both the `--opt=v` and `--opt v` forms.

`--git-dir` and `--work-tree` also change which repo the command acts on. Copilot doesn't mention this. The simplest safe choice is to treat `--work-tree=X` like `-C X` for the main-clone check, and resolve `--git-dir` too.

**Tests:** `git --no-pager push`, `git -P reset --hard` in a main clone, and `git --git-dir=... push`.

## 2. Review scan reads whole files: agree, but Medium is generous

`review.go:54` reads each whole file with `os.ReadFile`. `efforts/` gets QA screenshots and traces, so this cost grows over time. There's also a smaller issue: `fs.SkipAll` at `review.go:30` only stops the current walk. The loop then walks the next root anyway. The result is still correct, but the work is wasted.

**Fix:**
- Read `frontmatterSHA` from a `bufio.Reader`. Return "" if the first line isn't `---`, and stop at the closing `---` or after a size limit (say 8 KB).
- Stop the loop over roots once a match is found.

Review files can still sit anywhere under the roots.

## 3. Heredoc bodies parsed as commands: agree, and it's worse than reported

`splitCommands` (`shell.go:113`) has no heredoc handling at the top level. `skipSubst` has it, but only inside `$(...)`. Copilot's example is a false deny. The bigger problem is an apostrophe inside the body, such as "don't", which starts a single quote that never closes. Everything after it is swallowed:

```
cat > notes.md <<'EOF'
don't
EOF
git push
```

Here `git push` is never checked, so a normal push gets through without the review. Agents write prose with apostrophes all the time.

**Fix:** don't just call `skipHeredoc` from `splitCommands`. It jumps from `<<` straight to the end of the body, so `cat <<EOF > f && git push` would skip the `&& git push` on the first line.
- On `<<WORD` (not `<<<`), record the delimiter and keep lexing the rest of that line.
- At the newline, skip the bodies in order. This also handles two heredocs on one line.
- `skipSubst` has the same problem. It's lower risk, but it's worth using the same helper there.

**Tests:**
- Copilot's case: the heredoc body contains `git push`, so the command is allowed.
- An apostrophe in the body followed by a real `git push`, which must be denied.
- `cat <<EOF > f && git push` with a body, which must be denied.

## Next steps

This session can only read files, so I haven't changed anything. When we have a shell, the plan is:
1. Create a worktree on `feat/hooks-go-git-guard` at `projects/.worktrees/folkways/feat-hooks-go-git-guard`.
2. Write the tests first, then the three fixes, as separate `fix(hooks): …` commits.
3. Run `folkways review` on the new HEAD.
4. Ask you before pushing.

I'd also add these assessments to `00-STATUS.md` as round-1 dispositions. Should I go ahead? And do you want to reply to Copilot's comments on the PR yourself, or should I draft replies?
