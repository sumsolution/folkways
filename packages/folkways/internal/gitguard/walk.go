package gitguard

import (
	"maps"
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
	ask    string // the first ask reason; a deny anywhere wins
	cds    int    // cd, pushd and popd commands walked so far
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
		if c.Op != syntax.AndStmt {
			return w.stmt(c.Y, w.stmt(c.X, cur, false), false)
		}
		// In `a && b && c`, b and c see what runs before them, but may not run: after the chain, what
		// they changed is unknown.
		chain := andChain(c)
		cur = w.stmt(chain[0], cur, top)
		dir, cds := cur, w.cds
		w.res.forgetSet(func() {
			for _, s := range chain[1:] {
				cur = w.stmt(s, cur, top)
			}
		})
		if cur != dir || w.cds != cds {
			return unknown
		}
		return cur
	case *syntax.Subshell: // never changes the outer directory or variables
		w.res.scope(func() { w.list(c.Stmts, cur, top) })
		return cur
	case *syntax.Block:
		return w.list(c.Stmts, cur, top)
	case *syntax.ForClause, *syntax.WhileClause: // the next iteration sees what the body changed
		before := maps.Clone(w.res.vars.m)
		after := w.inner(c, cur)
		if after == cur && maps.Equal(before, w.res.vars.m) {
			return cur
		}
		return w.inner(c, unknown)
	case nil:
		return cur
	default: // if, while, for, case, functions, [[ ]], ...: a cd inside makes the directory unknown
		return w.inner(c, cur)
	}
}

// inner walks the statements and substitutions nested in n (a compound command, word, redirect or
// assignment). Nested statements are not tracked; substitutions run like subshells. A variable assigned
// here becomes unknown.
func (w *walker) inner(n syntax.Node, cur cwd) cwd {
	syntax.Walk(n, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.Stmt:
			cur = w.stmt(x, cur, false)
			return false
		case *syntax.CmdSubst:
			w.res.scope(func() { w.list(x.Stmts, cur, true) })
			return false
		case *syntax.ProcSubst:
			w.res.scope(func() { w.list(x.Stmts, cur, true) })
			return false
		case *syntax.Redirect: // `{W}>file` sets W to the fd number
			if x.N != nil && strings.HasPrefix(x.N.Value, "{") {
				w.res.unset(strings.Trim(x.N.Value, "{}"))
			}
		case *syntax.Assign:
			if x.Name != nil {
				w.res.unset(x.Name.Value)
			}
		case *syntax.WordIter:
			w.res.unset(x.Name.Value)
		case *syntax.BinaryArithm:
			if x.Op == syntax.Assgn || x.Op >= syntax.AddAssgn && x.Op <= syntax.PowAssgn {
				w.unsetArithm(x.X)
			}
		case *syntax.UnaryArithm:
			if x.Op == syntax.Inc || x.Op == syntax.Dec {
				w.unsetArithm(x.X)
			}
		case *syntax.ParamExp:
			if x.Exp != nil && (x.Exp.Op == syntax.AssignUnset || x.Exp.Op == syntax.AssignUnsetOrNull) {
				w.res.unset(x.Param.Value)
			}
		case *syntax.FuncDecl, *syntax.CoprocClause, *syntax.LetClause:
			w.res.stop()
		case *syntax.DeclClause:
			if x.Variant.Value == "nameref" || declaresNameref(x) {
				w.res.stop()
			}
		}
		return w.reason == ""
	})
	return cur
}

// call walks one simple command: its substitutions, then cd and git themselves. A top-level
// assignment statement records each plain `NAME=value` whose value resolves.
func (w *walker) call(c *syntax.CallExpr, cur cwd, top bool) cwd {
	record := top && len(c.Args) == 0
	for _, a := range c.Assigns {
		w.inner(a, cur)
		if record && a.Name != nil && !a.Append && !a.Naked && a.Index == nil && a.Array == nil {
			if v, ok := w.res.literal(a.Value); ok {
				w.res.set(a.Name.Value, v)
			}
		}
	}
	for _, a := range c.Args {
		w.inner(a, cur)
	}
	if len(c.Args) == 0 || w.reason != "" {
		return cur
	}
	name, ok := w.res.word(c.Args[0])
	if !ok || !keepsVars(name, c.Args[1:], w.res) {
		w.res.stop()
	}
	if !ok {
		return cur
	}
	switch name {
	case "cd":
		w.cds++
		target, ok := w.cdTarget(c.Args[1:])
		if !ok || !top {
			return unknown
		}
		return cur.cd(target)
	case "pushd", "popd":
		w.cds++
		return unknown
	case "git":
		v := w.git(c.Args[1:], c.Assigns, cur)
		w.reason = v.deny
		if w.ask == "" {
			w.ask = v.ask
		}
	}
	return cur
}

// keepVars lists commands that leave the shell's variables alone: git, builtins that set none
// (cd, pushd and popd set PWD and OLDPWD, handled in keepsVars) and programs agents often chain with git.
// Any other command stops tracking: a builtin, function or alias could set any variable.
var keepVars = map[string]bool{"git": true, "gh": true, "cd": true, "pushd": true, "popd": true,
	"echo": true, "printf": true, "pwd": true, "test": true, "[": true, "true": true, "false": true, ":": true,
	"ls": true, "cat": true, "mkdir": true, "rm": true, "cp": true, "mv": true, "grep": true, "head": true,
	"tail": true, "wc": true, "go": true, "npm": true, "make": true}

// keepsVars reports whether `name args...` leaves the tracked variables as they are.
func keepsVars(name string, args []*syntax.Word, res resolver) bool {
	switch name {
	case "cd", "pushd", "popd":
		res.unset("PWD")
		res.unset("OLDPWD")
	case "printf": // `printf -v NAME` sets NAME
		for _, a := range args {
			if v, ok := res.word(a); !ok || strings.HasPrefix(v, "-v") {
				return false
			}
		}
	}
	return keepVars[name]
}

// andChain returns the statements of an `a && b && c` chain in order.
func andChain(c *syntax.BinaryCmd) []*syntax.Stmt {
	if x, ok := c.X.Cmd.(*syntax.BinaryCmd); ok && x.Op == syntax.AndStmt && !c.X.Negated && !c.X.Background &&
		len(c.X.Redirs) == 0 {
		return append(andChain(x), c.Y)
	}
	return []*syntax.Stmt{c.X, c.Y}
}

// unsetArithm makes the variable an arithmetic assignment writes unknown; `W[0]=5` writes W.
func (w *walker) unsetArithm(x syntax.ArithmExpr) {
	if word, ok := x.(*syntax.Word); ok {
		if name := word.Lit(); name != "" {
			w.res.unset(name)
		} else if p, ok := word.Parts[0].(*syntax.ParamExp); ok && p.Short && len(word.Parts) == 1 {
			w.res.unset(p.Param.Value)
		}
	}
}

// declaresNameref reports a `declare -n` (or `local -n`, `typeset -n`).
func declaresNameref(d *syntax.DeclClause) bool {
	for _, a := range d.Args {
		if a.Naked && a.Name == nil && a.Value != nil {
			if f := a.Value.Lit(); strings.HasPrefix(f, "-") && strings.Contains(f, "n") {
				return true
			}
		}
	}
	return false
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
		home, ok := w.res.get("HOME") // plain `cd` goes home
		return home, ok && home != ""
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
func (w *walker) git(words []*syntax.Word, assigns []*syntax.Assign, cur cwd) verdict {
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
				return verdict{}
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
		return verdict{}
	}
	sub, rest := args[i].val, args[i+1:]
	switch {
	case sub == "commit":
		return verdict{deny: checkCommit(rest, w.cfg)}
	case sub == "push" && !cur.known:
		return unknownPush(w.cfg)
	case !cur.known:
		return verdict{}
	case sub == "push":
		return checkPushVerdict(cur.dir, rest, w.cfg)
	case mainGuard[sub]:
		return verdict{deny: checkMainClone(sub, cur.dir, w.cfg)}
	}
	return verdict{}
}

// literalPrefix returns the text of the leading literal part of w ("" if it starts with anything else).
func literalPrefix(w *syntax.Word) string {
	if lit, ok := w.Parts[0].(*syntax.Lit); ok {
		return lit.Value
	}
	return ""
}
