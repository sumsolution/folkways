package gitguard

import (
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

// firstWord parses src and returns the word after the command name of its first simple command.
func firstWord(t *testing.T, src string) *syntax.Word {
	t.Helper()
	f, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader("x "+src), "")
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return f.Stmts[0].Cmd.(*syntax.CallExpr).Args[1]
}

func TestResolverWord(t *testing.T) {
	r := newResolver("/home/u")
	cases := []struct {
		src  string
		want string
		ok   bool
	}{
		{`plain`, "plain", true},
		{`"double quoted"`, "double quoted", true},
		{`'single $HOME'`, "single $HOME", true},
		{`a\ b`, "a b", true},
		{`"a\"b"`, `a"b`, true},
		{`"a\nb"`, `a\nb`, true},
		{`pre"mid"'post'`, "premidpost", true},
		{`~`, "/home/u", true},
		{`~/work`, "/home/u/work", true},
		{`"$HOME/work"`, "/home/u/work", true},
		{`${HOME}/work`, "/home/u/work", true},
		{`$HOME`, "/home/u", true},
		{`"~/work"`, "~/work", true}, // quoted: no tilde expansion
		{`$OTHER`, "", false},
		{`${OTHER}/x`, "", false},
		{`$1`, "", false},
		{`${HOME%/*}`, "/home", true},
		{`$(pwd)`, "", false},
		{"`pwd`", "", false},
		{`"a $(pwd)"`, "", false},
		{`$'a\tb'`, "a\tb", true},
		{`<(pwd)`, "", false},
		{`a\*b`, "a*b", true},
		{"\"$(cat <<'EOF'\nfeat: x\n\nbody\nEOF\n)\"", "feat: x\n\nbody", true},
		{"\"$(cat <<EOF\n$HOME\nEOF\n)\"", "/home/u", true},
		{"\"$(cat <<EOF\n$(date)\nEOF\n)\"", "", false},
		{"\"$(cat -n <<EOF\nx\nEOF\n)\"", "", false},
	}
	for _, c := range cases {
		got, ok := r.word(firstWord(t, c.src))
		if got != c.want || ok != c.ok {
			t.Errorf("word(%s) = %q, %v; want %q, %v", c.src, got, ok, c.want, c.ok)
		}
	}
	if _, ok := newResolver("").word(firstWord(t, `$HOME`)); ok {
		t.Error("$HOME without a home directory must not resolve")
	}
}
