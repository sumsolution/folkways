package gitguard

import (
	"path/filepath"
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
	case "pushd":
		return in.moved(w.pushdTarget(c.Args[1:])), in
	case "popd":
		return dirset{unknown: true}, in
	case "git":
		// The commit message check does not depend on the directory, so it runs even when the
		// directories are unknown; the checks that need the repo are skipped for an unknown one.
		starts := in.dirs
		if len(starts) == 0 || in.unknown {
			starts = append(slices.Clone(starts), "")
		}
		for _, d := range starts {
			if w.reason = w.res.git(c.Args[1:], c.Assigns, d, d != "", w.cfg); w.reason != "" {
				break
			}
		}
	}
	return
}

// pushdTarget returns the directory `pushd dir` goes to. Options and stack rotations (+N, -N) are unknown.
func (w *walker) pushdTarget(args []*syntax.Word) (string, bool) {
	if len(args) == 0 {
		return "", false // swaps the top two entries of the stack
	}
	v, ok := w.res.word(args[0])
	if !ok || v == "" || v == "--" || v[0] == '-' || v[0] == '+' {
		return "", false
	}
	return v, true
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

// literalPrefix returns the text of the leading literal part of w ("" if it starts with anything else).
func literalPrefix(w *syntax.Word) string {
	if lit, ok := w.Parts[0].(*syntax.Lit); ok {
		return lit.Value
	}
	return ""
}

// arg is one resolved word of a git command line.
type arg struct {
	w   *syntax.Word
	val string
	ok  bool
}

// Git global options without an operand (`--exec-path` takes one only as `--exec-path=<path>`).
var gitFlag = map[string]bool{"-p": true, "-P": true, "--paginate": true, "--no-pager": true, "--no-replace-objects": true,
	"--no-optional-locks": true, "--bare": true, "--literal-pathspecs": true, "--glob-pathspecs": true,
	"--noglob-pathspecs": true, "--icase-pathspecs": true, "--no-lazy-fetch": true, "--exec-path": true}

// Git global options that take an operand: `--opt value` or `--opt=value`.
var gitOptWithArg = map[string]bool{"--git-dir": true, "--work-tree": true, "--namespace": true, "--config-env": true}

// gitPath is a directory operand of git: the raw text, and whether it could be resolved statically.
type gitPath struct {
	raw string
	set bool
	ok  bool
}

// git checks one `git [global options] <sub> args...` (assigns are the `NAME=value` prefixes of the command)
// run in dir; dirKnown is false when dir cannot be determined. The commit message check needs no
// directory. The push and main-clone checks need the repo: they run only when it is known, so anything
// unresolvable (a directory option, GIT_DIR) fails open. Only an unresolvable subcommand skips everything.
func (r resolver) git(words []*syntax.Word, assigns []*syntax.Assign, dir string, dirKnown bool, cfg team.Config) string {
	args := make([]arg, len(words))
	for i, word := range words {
		args[i].w = word
		args[i].val, args[i].ok = r.word(word)
	}
	var gitDir, workTree gitPath
	for _, a := range assigns { // GIT_DIR and GIT_WORK_TREE from the environment; options win
		if a.Name == nil || a.Value == nil || (a.Name.Value != "GIT_DIR" && a.Name.Value != "GIT_WORK_TREE") {
			continue
		}
		v, ok := r.word(a.Value)
		p := gitPath{raw: v, set: v != "" || !ok, ok: ok}
		if a.Name.Value == "GIT_DIR" {
			gitDir = p
		} else {
			workTree = p
		}
	}
	i := 0
options:
	for i < len(args) {
		v, unresolved := args[i].val, !args[i].ok
		if unresolved { // an option with an unresolvable operand (`--git-dir="$D"`, `-C$D`) is still an option
			if v = literalPrefix(args[i].w); !strings.HasPrefix(v, "-") {
				return ""
			}
		}
		// operand returns the option's operand (attached or the next word), the words used, and whether it is known.
		operand := func(attached string, hasAttached bool) (val string, n int, ok bool) {
			if hasAttached {
				return attached, 1, !unresolved
			}
			if i+1 >= len(args) {
				return "", 1, false
			}
			return args[i+1].val, 2, args[i+1].ok
		}
		name, eqVal, hasEq := strings.Cut(v, "=")
		switch {
		case v == "-C" || (strings.HasPrefix(v, "-C") && len(v) > 2):
			val, n, ok := operand(v[2:], len(v) > 2 || unresolved)
			switch {
			case !ok:
				dirKnown = false
			case filepath.IsAbs(val):
				dir, dirKnown = filepath.Clean(val), true
			case dirKnown:
				dir = resolve(dir, val)
			}
			i += n
		case v == "-c" || (strings.HasPrefix(v, "-c") && len(v) > 2):
			_, n, _ := operand("", len(v) > 2 || unresolved)
			i += n
		case gitOptWithArg[name]:
			val, n, ok := operand(eqVal, hasEq)
			switch name {
			case "--git-dir":
				gitDir = gitPath{raw: val, set: true, ok: ok}
			case "--work-tree":
				workTree = gitPath{raw: val, set: true, ok: ok}
			}
			i += n
		case gitFlag[name]:
			i++
		default:
			break options
		}
	}
	if i >= len(args) || !args[i].ok {
		return ""
	}
	sub, rest := args[i].val, args[i+1:]
	if sub == "commit" {
		return checkCommit(r, rest, cfg)
	}
	if sub != "push" && !mainGuard[sub] {
		return ""
	}
	repo, ok := gitRepo(dir, dirKnown, gitDir, workTree)
	switch {
	case !ok:
		return ""
	case sub == "push":
		return checkPush(repo, cfg)
	default:
		return checkMainClone(sub, repo, cfg)
	}
}

// gitRepo is the repo git would act on: with a git dir given, the work tree if given, else the parent of a
// git dir named `.git`, else the git dir itself; with no git dir, the repo of the current directory
// (a work tree alone does not select a repo). ok is false when it cannot be determined.
func gitRepo(dir string, dirKnown bool, gitDir, workTree gitPath) (string, bool) {
	resolvePath := func(p gitPath) (string, bool) {
		switch {
		case !p.ok:
			return "", false
		case filepath.IsAbs(p.raw):
			return filepath.Clean(p.raw), true
		case dirKnown:
			return resolve(dir, p.raw), true
		}
		return "", false
	}
	if !gitDir.set {
		return dir, dirKnown
	}
	if workTree.set {
		return resolvePath(workTree)
	}
	gd, ok := resolvePath(gitDir)
	if !ok {
		return "", false
	}
	if filepath.Base(gd) == ".git" {
		return filepath.Dir(gd), true
	}
	return gd, true
}
