# Copilot review, round 1 (on 75477a4)

Pasted from PR #13 (three of Copilot's findings). Findings: 1 High, 2 Medium.

## 1. High: Handle Git global options before subcommand detection (`hooks/internal/gitguard/guard.go:66`)

`checkGit` consumes only `-C` and `-c` before selecting the subcommand. Git accepts other global options, so `git --no-pager push` sets `sub` to `--no-pager` here and bypasses the pre-push review gate; the same applies to `git --no-pager reset` in a main clone. Recognize supported Git global flags and their operands before dispatching the real subcommand, and test both guarded operations with a global option.

## 2. Medium: Avoid reading entire files during review scans (`hooks/internal/gitguard/review.go:28`)

On every pre-push check without a matching review, `WalkDir` calls `frontmatterSHA` on every file in `efforts/` and the personal work tree; that function reads each entire file, including binary attachments, before checking for frontmatter. Large files or a mature repo can stall the hook or consume excessive memory. Read only the leading frontmatter from a stream (stop immediately if it is absent), and stop walking other roots once a match is found; preserve the contract's allowance for review files anywhere under these roots.

## 3. Medium: Ignore heredoc bodies when parsing commands (`hooks/internal/gitguard/shell.go:137`)

`splitCommands` treats newlines inside a top-level heredoc as command separators. For example, `cat <<EOF > notes.md\ngit push\nEOF` writes the words `git push` to a file, but this parser checks them as an actual push and can deny the tool call. Skip heredoc bodies while identifying executable commands, and cover a top-level heredoc containing `git push` in the lexer tests.
