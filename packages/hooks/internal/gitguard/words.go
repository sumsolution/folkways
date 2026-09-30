package gitguard

import (
	"errors"
	"io"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

var errUnresolvable = errors.New("not resolvable statically")

// resolver turns a shell word into the string the shell would pass, when that is knowable statically:
// quotes, escapes, `~`, $HOME and the `$(cat <<'EOF' ... EOF)` form of a commit message. Any other
// variable or command makes the word unresolvable.
type resolver struct {
	home string
	cfg  *expand.Config
}

func newResolver(home string) resolver {
	var env []string
	if home != "" {
		env = append(env, "HOME="+home)
	}
	config := func() *expand.Config {
		return &expand.Config{Env: expand.ListEnviron(env...), NoUnset: true,
			ProcSubst: func(*syntax.ProcSubst) (string, error) { return "", errUnresolvable }}
	}
	cfg, doc := config(), config() // an expand.Config is not reentrant: the heredoc body needs its own
	cfg.CmdSubst = func(w io.Writer, cs *syntax.CmdSubst) error { return catHeredoc(doc, w, cs) }
	return resolver{home: home, cfg: cfg}
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
