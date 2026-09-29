package gitguard

import (
	"slices"
	"strings"

	"github.com/sumsolution/folkways/hooks/internal/team"
	"mvdan.cc/sh/v3/syntax"
)

// dirset is the set of directories a command may run in. unknown means it may also run somewhere that
// cannot be determined statically (after `cd $SOMETHING`); git checks skip that part, so they fail open.
type dirset struct {
	dirs    []string
	unknown bool
}

func union(sets ...dirset) dirset {
	var out dirset
	for _, s := range sets {
		out.unknown = out.unknown || s.unknown
		for _, d := range s.dirs {
			if !slices.Contains(out.dirs, d) {
				out.dirs = append(out.dirs, d)
			}
		}
	}
	return out
}

func (s dirset) empty() bool { return len(s.dirs) == 0 && !s.unknown }

// moved is where a successful `cd target` leaves the shell; !known makes every destination unknown.
func (s dirset) moved(target string, known bool) dirset {
	if !known {
		return dirset{unknown: true}
	}
	out := dirset{unknown: s.unknown}
	for _, d := range s.dirs {
		out.dirs = append(out.dirs, resolve(d, target))
	}
	return out
}

// walker walks a parsed command line and records the first deny reason.
type walker struct {
	cfg    team.Config
	res    resolver
	reason string
}

// list walks statements in sequence. It returns the directories after the last statement if it
// succeeded and if it failed (they differ after a `cd`, which fails without moving).
//
// `a && b` runs b where a succeeded; `a || b` where a failed; after `;` or a newline a statement may run
// in either. A cd in a pipeline, in the background or in a ( ) subshell does not affect what follows.
func (w *walker) list(stmts []*syntax.Stmt, in dirset) (succ, fail dirset) {
	succ, fail = in, in
	for i, s := range stmts {
		d := in
		if i > 0 {
			d = union(succ, fail)
		}
		succ, fail = w.stmt(s, d)
	}
	return succ, fail
}

// subst walks the command substitutions and process substitutions inside n (a word, redirect,
// assignment or expression) as commands that run in in.
func (w *walker) subst(n syntax.Node, in dirset) {
	syntax.Walk(n, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.CmdSubst:
			w.list(x.Stmts, in)
			return false
		case *syntax.ProcSubst:
			w.list(x.Stmts, in)
			return false
		}
		return w.reason == ""
	})
}

func (w *walker) stmt(s *syntax.Stmt, in dirset) (succ, fail dirset) {
	succ, fail = in, in
	if w.reason != "" {
		return
	}
	for _, r := range s.Redirs {
		if r.Word != nil {
			w.subst(r.Word, in)
		}
		if r.Hdoc != nil {
			w.subst(r.Hdoc, in)
		}
	}
	switch c := s.Cmd.(type) {
	case *syntax.CallExpr:
		succ, fail = w.call(c, in)
	case *syntax.BinaryCmd:
		switch c.Op {
		case syntax.AndStmt:
			ls, lf := w.stmt(c.X, in)
			rs, rf := w.stmt(c.Y, ls)
			succ, fail = rs, union(lf, rf)
		case syntax.OrStmt:
			ls, lf := w.stmt(c.X, in)
			if lf.empty() {
				lf = union(ls, lf)
			}
			rs, rf := w.stmt(c.Y, lf)
			succ, fail = union(ls, rs), rf
		default: // | and |&: both sides run in the current directory and cd does not persist
			w.stmt(c.X, in)
			w.stmt(c.Y, in)
		}
	case *syntax.Subshell:
		w.list(c.Stmts, in)
	case *syntax.Block:
		succ, fail = w.list(c.Stmts, in)
	case *syntax.IfClause:
		out := w.ifClause(c, in)
		succ, fail = out, out
	case *syntax.WhileClause:
		cs, cf := w.list(c.Cond, in)
		ds, df := w.list(c.Do, cs)
		out := union(in, cs, cf, ds, df)
		succ, fail = out, out
	case *syntax.ForClause:
		switch l := c.Loop.(type) {
		case *syntax.WordIter:
			for _, item := range l.Items {
				w.subst(item, in)
			}
		case *syntax.CStyleLoop:
			w.subst(l, in)
		}
		ds, df := w.list(c.Do, in)
		out := union(in, ds, df)
		succ, fail = out, out
	case *syntax.CaseClause:
		w.subst(c.Word, in)
		out := in
		for _, item := range c.Items {
			for _, p := range item.Patterns {
				w.subst(p, in)
			}
			is, i2 := w.list(item.Stmts, in)
			out = union(out, is, i2)
		}
		succ, fail = out, out
	case *syntax.FuncDecl:
		w.stmt(c.Body, in) // checked where defined; calls are not followed
	case *syntax.TimeClause:
		if c.Stmt != nil {
			succ, fail = w.stmt(c.Stmt, in)
		}
	case *syntax.CoprocClause:
		w.stmt(c.Stmt, in)
	case nil:
	default: // [[ ]], (( )), let, declare, ...: only substitutions matter
		w.subst(c, in)
	}
	if s.Negated {
		succ, fail = fail, succ
	}
	if s.Background {
		succ, fail = in, in
	}
	return succ, fail
}

func (w *walker) ifClause(c *syntax.IfClause, in dirset) dirset {
	cs, cf := w.list(c.Cond, in)
	ts, tf := w.list(c.Then, cs)
	out := union(cs, cf, ts, tf)
	if c.Else != nil {
		out = union(out, w.ifClause(c.Else, union(cf, in)))
	}
	return out
}

// call walks one simple command: its substitutions, then cd and git themselves.
func (w *walker) call(c *syntax.CallExpr, in dirset) (succ, fail dirset) {
	succ, fail = in, in
	for _, a := range c.Assigns {
		w.subst(a, in)
	}
	for _, a := range c.Args {
		w.subst(a, in)
	}
	if len(c.Args) == 0 || w.reason != "" {
		return
	}
	name, ok := w.res.word(c.Args[0])
	if !ok {
		return
	}
	switch name {
	case "cd":
		return in.moved(w.cdTarget(c.Args[1:])), in
	case "git":
		for _, d := range in.dirs {
			if w.reason = w.res.git(c.Args[1:], d, w.cfg); w.reason != "" {
				break
			}
		}
	}
	return
}

// cdTarget returns the directory `cd args...` goes to, and whether it is known.
func (w *walker) cdTarget(args []*syntax.Word) (string, bool) {
	for len(args) > 0 {
		v, ok := w.res.word(args[0])
		if !ok {
			return "", false
		}
		if v != "--" && v != "-L" && v != "-P" && v != "-e" && v != "-@" {
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		return w.res.home, w.res.homeOK // plain `cd` goes home
	}
	v, ok := w.res.word(args[0])
	if !ok || v == "-" || v == "" {
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

// Git global options that take an operand: `--opt value` or `--opt=value`.
var gitOptWithArg = map[string]bool{"--git-dir": true, "--work-tree": true, "--namespace": true, "--config-env": true}

// Git global options without an operand (`--exec-path` takes one only as `--exec-path=<path>`).
var gitFlag = map[string]bool{"-p": true, "-P": true, "--paginate": true, "--no-pager": true, "--no-replace-objects": true,
	"--no-optional-locks": true, "--bare": true, "--literal-pathspecs": true, "--glob-pathspecs": true,
	"--noglob-pathspecs": true, "--icase-pathspecs": true, "--no-lazy-fetch": true, "--exec-path": true}

// git checks one `git [global options] <sub> args...` run in dir. A word that cannot be resolved where
// it matters (the subcommand, a directory option) means the target repo is unknown: fail open.
func (r resolver) git(words []*syntax.Word, dir string, cfg team.Config) string {
	args := make([]arg, len(words))
	for i, word := range words {
		args[i].w = word
		args[i].val, args[i].ok = r.word(word)
	}
	var gitDir, workTree string
	i := 0
options:
	for i < len(args) {
		if !args[i].ok {
			return ""
		}
		v := args[i].val
		// operand returns the option's operand and the number of words used, or ok=false when it is unknown.
		operand := func(attached string, hasAttached bool) (val string, n int, ok bool) {
			if hasAttached {
				return attached, 1, true
			}
			if i+1 >= len(args) {
				return "", 1, false
			}
			return args[i+1].val, 2, args[i+1].ok
		}
		name, eqVal, hasEq := strings.Cut(v, "=")
		var (
			val    string
			n      int
			ok     bool
			target *string
			dirOpt bool
		)
		switch {
		case v == "-C" || (strings.HasPrefix(v, "-C") && len(v) > 2):
			val, n, ok = operand(v[2:], len(v) > 2)
			dirOpt = true
		case v == "-c" || (strings.HasPrefix(v, "-c") && len(v) > 2):
			_, n, _ = operand("", len(v) > 2)
		case gitOptWithArg[name]:
			val, n, ok = operand(eqVal, hasEq)
			switch name {
			case "--git-dir":
				target, dirOpt = &gitDir, true
			case "--work-tree":
				target, dirOpt = &workTree, true
			}
		case gitFlag[name]:
			n = 1
		default:
			break options
		}
		if dirOpt {
			if !ok {
				return ""
			}
			switch {
			case target != nil:
				*target = resolve(dir, val)
			default:
				dir = resolve(dir, val)
			}
		}
		i += n
	}
	repo := dir
	switch {
	case workTree != "":
		repo = workTree
	case gitDir != "" && strings.HasSuffix(gitDir, "/.git"):
		repo = strings.TrimSuffix(gitDir, "/.git")
	case gitDir != "":
		repo = gitDir
	}
	if i >= len(args) || !args[i].ok {
		return ""
	}
	sub, rest := args[i].val, args[i+1:]
	switch {
	case sub == "commit":
		return checkCommit(r, rest, cfg)
	case sub == "push":
		return checkPush(repo, cfg)
	case mainGuard[sub]:
		return checkMainClone(sub, repo, cfg)
	}
	return ""
}
