package gitguard

import (
	"errors"
	"io"
	"maps"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

var errUnresolvable = errors.New("not resolvable statically")

// resolver turns a shell word into the string the shell would pass, when that is knowable statically:
// quotes, escapes, `~`, the variables the walker knows and the `$(cat <<'EOF' ... EOF)` form of a commit
// message. Any other variable or command makes the word unresolvable.
type resolver struct {
	vars *vars
	cfg  *expand.Config
}

// vars holds the variables the walker knows: HOME and top-level `NAME=literal` assignments. It is
// read-only to expand, so `${W:=x}` is unresolvable.
type vars struct {
	m      map[string]string
	home   string
	frozen bool // after a command that can change variables in ways the walker can't follow
}

func (v *vars) Get(name string) expand.Variable {
	if s, ok := v.m[name]; ok {
		return expand.Variable{Set: true, Kind: expand.String, Str: s}
	}
	return expand.Variable{}
}

func (v *vars) Each(f func(string, expand.Variable) bool) {
	for name := range v.m {
		if !f(name, v.Get(name)) {
			return
		}
	}
}

func newResolver(home string) resolver {
	v := &vars{m: map[string]string{}, home: home}
	if home != "" {
		v.m["HOME"] = home
	}
	config := func() *expand.Config {
		return &expand.Config{Env: v, NoUnset: true,
			ProcSubst: func(*syntax.ProcSubst) (string, error) { return "", errUnresolvable }}
	}
	cfg, doc := config(), config() // an expand.Config is not reentrant: the heredoc body needs its own
	cfg.CmdSubst = func(w io.Writer, cs *syntax.CmdSubst) error { return catHeredoc(doc, w, cs) }
	return resolver{vars: v, cfg: cfg}
}

func (r resolver) get(name string) (string, bool) {
	s, ok := r.vars.m[name]
	return s, ok
}

func (r resolver) set(name, value string) {
	if !r.vars.frozen {
		r.vars.m[name] = value
	}
}

func (r resolver) unset(name string) { delete(r.vars.m, name) }

// stop forgets every variable and ignores later assignments. HOME keeps its starting value, as before
// git-guard tracked variables.
func (r resolver) stop() {
	home, ok := r.vars.m["HOME"]
	clear(r.vars.m)
	if ok && home == r.vars.home {
		r.vars.m["HOME"] = home
	}
	r.vars.frozen = true
}

// forgetChanged forgets the variables set since before was copied.
func (r resolver) forgetChanged(before map[string]string) {
	for name, v := range r.vars.m {
		if old, ok := before[name]; !ok || old != v {
			delete(r.vars.m, name)
		}
	}
}

// scope runs f with a copy of the variables and drops it afterwards, as a subshell does.
func (r resolver) scope(f func()) {
	saved := *r.vars
	r.vars.m = maps.Clone(saved.m)
	f()
	*r.vars = saved
}

// literal returns the value of an assignment's word, which the shell does not split.
func (r resolver) literal(w *syntax.Word) (string, bool) {
	if w == nil {
		return "", true
	}
	s, err := expand.Literal(r.cfg, w)
	return s, err == nil
}

// word returns the value of w and whether it could be resolved.
func (r resolver) word(w *syntax.Word) (string, bool) {
	fields, err := expand.Fields(r.cfg, w)
	if err != nil || len(fields) != 1 {
		return "", false
	}
	return fields[0], true
}

// catHeredoc expands `$(cat <<EOF ... EOF)`, the usual way an agent passes a multi-line commit message.
func catHeredoc(cfg *expand.Config, w io.Writer, cs *syntax.CmdSubst) error {
	if len(cs.Stmts) != 1 {
		return errUnresolvable
	}
	st := cs.Stmts[0]
	call, ok := st.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) != 1 || call.Args[0].Lit() != "cat" || len(st.Redirs) != 1 || st.Redirs[0].Hdoc == nil {
		return errUnresolvable
	}
	body, err := expand.Document(cfg, st.Redirs[0].Hdoc)
	if err == nil {
		_, err = io.WriteString(w, body)
	}
	return err
}
