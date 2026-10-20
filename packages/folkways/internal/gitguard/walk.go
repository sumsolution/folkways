package gitguard

import (
	"path/filepath"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/team"
	"mvdan.cc/sh/v3/syntax"
)

// cwd is the directory the next command runs in. When it is not known, the push and main-clone
// checks are skipped (fail open): git-guard catches agent mistakes, it is not a sandbox.
type cwd struct {
	dir   string
	known bool
}

var unknown = cwd{}

// walker walks the simple commands of a parsed command line in source order and records the first deny reason.
type walker struct {
	cfg    team.Config
	res    resolver
	reason string
}

// list walks statements in order. top is true where a `cd` is tracked: a top-level `&&` or `;` chain,
// also inside `{ }`, `( )` and substitutions.
func (w *walker) list(stmts []*syntax.Stmt, cur cwd, top bool) cwd {
	for _, s := range stmts {
		cur = w.stmt(s, cur, top)
	}
	return cur
}

func (w *walker) stmt(s *syntax.Stmt, cur cwd, top bool) cwd {
	if w.reason != "" {
		return cur
	}
	top = top && !s.Background
	for _, r := range s.Redirs {
		w.inner(r, cur)
	}
	switch c := s.Cmd.(type) {
	case *syntax.CallExpr:
		return w.call(c, cur, top)
	case *syntax.BinaryCmd: // a cd in `||` or a pipeline could run or not: only `&&` keeps tracking it
		top = top && c.Op == syntax.AndStmt
		return w.stmt(c.Y, w.stmt(c.X, cur, top), top)
	case *syntax.Subshell: // never changes the outer directory
		w.list(c.Stmts, cur, top)
		return cur
	case *syntax.Block:
		return w.list(c.Stmts, cur, top)
	case nil:
		return cur
	default: // if, while, for, case, functions, [[ ]], ...: a cd inside makes the directory unknown
		return w.inner(c, cur)
	}
}

// inner walks the statements and substitutions nested in n (a compound command, word, redirect or
// assignment). Nested statements are not tracked; substitutions run like subshells.
func (w *walker) inner(n syntax.Node, cur cwd) cwd {
	syntax.Walk(n, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.Stmt:
			cur = w.stmt(x, cur, false)
			return false
		case *syntax.CmdSubst:
			w.list(x.Stmts, cur, true)
			return false
		case *syntax.ProcSubst:
			w.list(x.Stmts, cur, true)
			return false
		}
		return w.reason == ""
	})
	return cur
}

// call walks one simple command: its substitutions, then cd and git themselves.
func (w *walker) call(c *syntax.CallExpr, cur cwd, top bool) cwd {
	for _, a := range c.Assigns {
		w.inner(a, cur)
	}
	for _, a := range c.Args {
		w.inner(a, cur)
	}
	if len(c.Args) == 0 || w.reason != "" {
		return cur
	}
	name, ok := w.res.word(c.Args[0])
	if !ok {
		return cur
	}
	switch name {
	case "cd":
		target, ok := w.cdTarget(c.Args[1:])
		if !ok || !top {
			return unknown
		}
		return cur.cd(target)
	case "pushd", "popd":
		return unknown
	case "git":
		w.reason = w.git(c.Args[1:], c.Assigns, cur)
	}
	return cur
}

// cd returns where `cd target` leaves the shell. An absolute target is always known.
func (c cwd) cd(target string) cwd {
	if filepath.IsAbs(target) {
		return cwd{filepath.Clean(target), true}
	}
	if !c.known {
		return unknown
	}
	return cwd{filepath.Join(c.dir, target), true}
}

// cdTarget returns the directory `cd args...` goes to, and whether it is known.
func (w *walker) cdTarget(args []*syntax.Word) (string, bool) {
	for len(args) > 0 {
		v, ok := w.res.word(args[0])
		if !ok {
			return "", false
		}
		if v != "--" && v != "-L" && v != "-P" {
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		return w.res.home, w.res.home != "" // plain `cd` goes home
	}
	v, ok := w.res.word(args[0])
	if !ok || v == "" || v[0] == '-' {
		return "", false
	}
	return v, true
}

// arg is one resolved word of a git command line.
type arg struct {
	w   *syntax.Word
	val string
	ok  bool
}

// Git global options without an operand.
var gitFlag = map[string]bool{"-p": true, "-P": true, "--paginate": true, "--no-pager": true,
	"--no-replace-objects": true, "--no-optional-locks": true, "--literal-pathspecs": true, "--no-lazy-fetch": true}

// Git global options whose operand may be the next word. They move the repo, so the directory becomes unknown.
var gitOptWithArg = map[string]bool{"--git-dir": true, "--work-tree": true, "--namespace": true, "--config-env": true}

// git checks one `git [global options] <sub> args...` run in cur. The commit message check always runs.
// The push and main-clone checks run only when the directory is known: `-C <resolvable>`, `-c` and the
// flags above keep it known; any other global option, or a GIT_DIR/GIT_WORK_TREE prefix, makes it unknown.
func (w *walker) git(words []*syntax.Word, assigns []*syntax.Assign, cur cwd) string {
	args := make([]arg, len(words))
	for i, word := range words {
		args[i] = arg{w: word}
		args[i].val, args[i].ok = w.res.word(word)
	}
	for _, a := range assigns {
		if a.Name != nil && (a.Name.Value == "GIT_DIR" || a.Name.Value == "GIT_WORK_TREE") {
			cur = unknown
		}
	}
	i := 0
	for i < len(args) {
		v := args[i].val
		if !args[i].ok { // `-C"$D"` is still an option; an unresolvable subcommand is not checked
			if !strings.HasPrefix(literalPrefix(args[i].w), "-") {
				return ""
			}
			cur = unknown
			i++
			continue
		}
		if !strings.HasPrefix(v, "-") {
			break
		}
		name, _, hasEq := strings.Cut(v, "=")
		switch {
		case v == "-C" || v == "-c":
			if v == "-C" {
				if i+1 < len(args) && args[i+1].ok {
					cur = cur.cd(args[i+1].val)
				} else {
					cur = unknown
				}
			}
			i += 2
		case strings.HasPrefix(v, "-C"):
			cur = cur.cd(v[2:])
			i++
		case strings.HasPrefix(v, "-c") || gitFlag[v]:
			i++
		case gitOptWithArg[name] && !hasEq:
			cur = unknown
			i += 2
		default:
			cur = unknown
			i++
		}
	}
	if i >= len(args) {
		return ""
	}
	sub, rest := args[i].val, args[i+1:]
	switch {
	case sub == "commit":
		return checkCommit(rest, w.cfg)
	case !cur.known:
		return ""
	case sub == "push":
		return checkPush(cur.dir, w.cfg)
	case mainGuard[sub]:
		return checkMainClone(sub, cur.dir, w.cfg)
	}
	return ""
}

// literalPrefix returns the text of the leading literal part of w ("" if it starts with anything else).
func literalPrefix(w *syntax.Word) string {
	if lit, ok := w.Parts[0].(*syntax.Lit); ok {
		return lit.Value
	}
	return ""
}
