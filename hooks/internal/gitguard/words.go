package gitguard

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// resolver turns shell words into the string the shell would pass, when that is knowable statically:
// literals, quotes, backslash escapes, `~`, `~/x`, $HOME and ${HOME}. Anything else is "not resolvable".
type resolver struct {
	home   string
	homeOK bool
}

// word returns the value of w and whether it could be resolved.
func (r resolver) word(w *syntax.Word) (string, bool) {
	var b strings.Builder
	for i, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			v := p.Value
			if i == 0 && strings.HasPrefix(v, "~") {
				if !r.homeOK || (v != "~" && !strings.HasPrefix(v, "~/")) {
					return "", false // ~user is not resolvable
				}
				b.WriteString(r.home)
				v = v[1:]
			}
			s, ok := unescape(v, false)
			if !ok {
				return "", false
			}
			b.WriteString(s)
		case *syntax.SglQuoted:
			if p.Dollar {
				return "", false // $'...' has escapes we do not evaluate
			}
			b.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, q := range p.Parts {
				switch q := q.(type) {
				case *syntax.Lit:
					s, _ := unescape(q.Value, true)
					b.WriteString(s)
				case *syntax.ParamExp:
					s, ok := r.param(q)
					if !ok {
						return "", false
					}
					b.WriteString(s)
				default:
					return "", false
				}
			}
		case *syntax.ParamExp:
			s, ok := r.param(p)
			if !ok {
				return "", false
			}
			b.WriteString(s)
		default:
			return "", false
		}
	}
	return b.String(), true
}

// param resolves a plain $HOME or ${HOME}.
func (r resolver) param(p *syntax.ParamExp) (string, bool) {
	plain := p.Param != nil && p.Param.Value == "HOME" && p.Flags == nil && !p.Excl && !p.Length && !p.Width && !p.IsSet &&
		p.NestedParam == nil && p.Index == nil && len(p.Modifiers) == 0 && p.Slice == nil && p.Repl == nil && p.Names == 0 && p.Exp == nil
	if !plain || !r.homeOK {
		return "", false
	}
	return r.home, true
}

// unescape removes backslash escapes from a literal. Inside double quotes only \$ \` \" \\ and a
// line continuation are escapes. Outside quotes an unescaped glob character makes the word unresolvable.
func unescape(v string, quoted bool) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case c == '\\' && i+1 < len(v):
			next := v[i+1]
			i++
			switch {
			case next == '\n':
			case quoted && !strings.ContainsRune("$`\"\\", rune(next)):
				b.WriteByte('\\')
				b.WriteByte(next)
			default:
				b.WriteByte(next)
			}
		case !quoted && strings.IndexByte("*?[", c) >= 0:
			return "", false
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), true
}

// heredocBody returns the text of a heredoc body word. A body made only of literal text (a quoted
// delimiter, or no expansions) is taken as written.
func (r resolver) heredocBody(w *syntax.Word) (string, bool) {
	var b strings.Builder
	for _, part := range w.Parts {
		lit, ok := part.(*syntax.Lit)
		if !ok {
			return r.word(w)
		}
		b.WriteString(lit.Value)
	}
	return b.String(), true
}
