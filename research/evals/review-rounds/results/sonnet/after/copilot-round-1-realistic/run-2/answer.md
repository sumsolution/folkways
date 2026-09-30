I haven't changed any files. Bash and Edit are disabled in this session, and I have no write tool, so all seven edits failed. Nothing was run either, including `go test`. I read the three files behind the findings, and the fixes below are ready to apply. I'd suggest applying them in a session with write access, or you can paste them in.

All three findings look valid. Each fix is small, and none needs new machinery.

**1. High: global options before the subcommand (`guard.go:66`)**

Replace the `-C`/`-c` loop in `checkGit` with a loop over any leading dash-word:

```go
var globalOperand = map[string]bool{"-C": true, "-c": true, "--git-dir": true, "--work-tree": true, "--namespace": true, "--super-prefix": true, "--config-env": true}

for len(args) > 0 && strings.HasPrefix(args[0], "-") {
	n := 1
	if globalOperand[args[0]] && len(args) >= 2 {
		n = 2
		if args[0] == "-C" {
			dir = resolve(dir, args[1])
		}
	}
	args = args[n:]
}
```

- **Flags:** a subcommand never starts with `-`, so any other leading dash-word is a flag. That covers `--no-pager`, `-p` and `--bare`.
- **`--opt=value`:** it counts as one word, so it needs no special case.
- **Tests:** add cases to `guard_test.go` for `git --no-pager commit -m bad` and `git --no-pager -c a=b -C x commit -m bad`. The push and main-clone guards need a temp-repo test. I'd check the existing contract tests for a fixture before writing one.

**2. Medium: reading whole files (`review.go:28`)**

- **`frontmatterSHA`:** open the file and read through `bufio.Scanner` over `io.LimitReader(f, 64<<10)`. Return `""` if the first line isn't `---`. Otherwise accumulate lines until one starts with `---`, then YAML-parse only that block. Binary files exit after the first line, and reads are capped at 64 KB.
- **Walk:** in `hasFrontmatterReview`, add `if found { break }` after each root's `WalkDir`. `fs.SkipAll` only ends the current walk, so this stops the walk from continuing into the next root once a match is found.
- **Tests:** the existing CRLF, no-frontmatter and unterminated cases in `TestFrontmatterSHA` still cover the behaviour. I'd add a file that is large and binary with no frontmatter.

**3. Medium: heredoc bodies (`shell.go:137`)**

- **`heredocBodyEnd`:** split it out of `skipHeredoc`. It takes `(s, pos, word)` and returns the index after the terminator line.
- **`splitCommands`:** on a top-level `<<WORD` (not `<<<`), record `WORD` as pending and continue the line. Any redirect such as `> notes.md` stays in the command.
- **Newline:** on the next newline, flush the segment and jump past each pending body with `heredocBodyEnd`. Then set `last = i` so the body is never emitted.
- **Tests:** add `cat <<EOF > notes.md\ngit push\nEOF` → `["cat <<EOF > notes.md"]`. Also add `cat <<EOF\ngit push\nEOF\ngit status` → `["cat <<EOF", "git status"]`. And add a `checkCommand` case showing that `git push` inside a heredoc isn't denied.

The contract README says git-guard catches agent mistakes, not a determined bypass. The `--git-dir` and `--work-tree` operand handling is therefore the outer limit. I wouldn't add anything more for global options.

Once you have write access, I can apply all three and add the tests. I'd also run `go test ./...` in `projects/folkways/hooks`.
