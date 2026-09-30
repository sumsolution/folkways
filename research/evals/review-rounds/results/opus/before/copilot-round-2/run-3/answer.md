I've checked all 7 findings against the code at `projects/folkways/hooks/internal/gitguard/` (8188289). I haven't changed any code yet, because this session has no edit or shell tools.

**Summary:** 6 findings are real and 1 is probably wrong. Finding 7 is also worse than Copilot says: besides blocking a harmless command, it can let a real push through unchecked.

| # | Verdict | Why (code) | Recommended fix |
|---|---|---|---|
| 1 `cd ~/…`, `cd "$HOME/…"` | **Real, fix** | `cdTarget` gives up on `$`, `` ` `` and `~` (`guard.go:142`). With no target, `moved` stays nil, so the next command is checked in the original directory (`guard.go:86-93`). Agents write `cd ~/projects/x && git push` all the time. | Expand `~`, `~/…`, `$HOME` and `${HOME}` with `os.UserHomeDir()`. Leave other variables unresolved and add them to the contract's out-of-scope line. |
| 2 `-C/path`, `-ccore.x=y` | **Reject (to confirm)** | As far as I know, git's `handle_options` matches `-C` and `-c` exactly, so the attached forms fail with `unknown option` and nothing runs. Nothing is bypassed. | Confirm with `git -C/tmp status` before replying to Copilot. If git does accept the attached form, the fix is small, in `guard.go:187-193`. |
| 3 `--git-dir` / `--work-tree` | **Real, fix** | These two options are skipped without changing `dir` (`guard.go:190-193`). As a result, `checkMainClone` and `checkPush` look at the wrong repo. | Collect both options, resolve them against `dir`, and pass them to the `gitOut` checks (`rev-parse HEAD`, `--show-toplevel`) so git picks the target. The same gap exists for a `GIT_DIR=… git reset` prefix, because `commandWords` drops env assignments (`guard.go:126`). Fix it together or note it. |
| 4 `-m"msg"`, `-am"msg"` | **Real, fix (cheap)** | `shortMsg` is `^-[a-zA-Z]*m$` (`guard.go:19`), which only matches a separate argument. Git's option parser accepts a value attached to `-m`. | Match `-<flags>m<value>`, where the flags before `m` take no argument (`a`, `s`, `n`, `v`, `q`, `e`). An empty value means the message is the next word. |
| 5 `$(…)` in unquoted heredoc | **Real, lower priority** | `stripHeredocs` drops every heredoc body (`shell.go:276-283`). An agent pushing from inside a heredoc body is unlikely. | Track whether the delimiter was quoted (`'EOF'`, `"EOF"`, `\EOF`). For unquoted bodies, check `$( )` and backticks, and treat quotes in the body as literal text; the current `substitutions()` treats them as quoting. The contract line "heredoc bodies are data" needs updating too. |
| 6 backslash-newline | **Real, fix first** | `words` writes the newline into the next word (`shell.go:323-325`), and `appendDouble` does the same inside double quotes (`shell.go:351`). `git \⏎push` becomes subcommand `"\npush"`, so the push is not checked. Multi-line commands with `\` are very common agent output. | In both places, drop a backslash followed by a newline without closing the word. `cdTarget` goes through `commandWords`, so this also fixes it. |
| 7 comments | **Real, and worse than reported** | Copilot's case, `echo ok # note; git push`, is a false deny caused by the split at `shell.go:162`. There is also a missed check: an apostrophe in a comment (`# don't push yet`) starts a quote in `skipSingle` (`shell.go:143`) that runs to the next `'`. Any `git push` on the lines in between goes unchecked. | Skip `#` to the end of the line when it starts a word (at the start, or after whitespace or `;&\|()` and newline). Do it in `stripHeredocs` (renamed, since it already tracks quotes, substitutions and heredocs) so the parsers after it never see comments. Test both the false deny and the apostrophe case. |

**Order I'd suggest:** 6, 7, 1, 4, 3, 5. The first four are forms agents actually produce, which is what the contract says git-guard is for ("catches agent mistakes, not a determined bypass", `test/hooks/README.md:87`).

**Doing the work:**
- The code changes go in a worktree at `projects/.worktrees/folkways/feat/hooks-go-git-guard`; `projects/.worktrees/` is empty right now. Nothing runs in the main clone.
- Each fix gets Go unit tests plus contract fixtures (currently 63).
- The "Git global options" and "heredoc" lines in the contract README change with fixes 3 and 5.
- `00-STATUS.md` gets the triage, and Copilot's thread gets a reply for finding 2 explaining why it's rejected.

To go ahead I need a session with edit and shell access, and your OK on two points: rejecting finding 2 once it's confirmed, and fixing finding 5 now rather than parking it as a side-quest.
