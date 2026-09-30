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
	r := resolver{home: "/home/u", homeOK: true}
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
		{`~other/x`, "", false},
		{`$OTHER`, "", false},
		{`${HOME:-x}`, "", false},
		{`${HOME%/*}`, "", false},
		{`$(pwd)`, "", false},
		{"`pwd`", "", false},
		{`"a $(pwd)"`, "", false},
		{`$'a\tb'`, "", false},
		{`*.md`, "", false},
		{`a\*b`, "a*b", true},
	}
	for _, c := range cases {
		got, ok := r.word(firstWord(t, c.src))
		if got != c.want || ok != c.ok {
			t.Errorf("word(%s) = %q, %v; want %q, %v", c.src, got, ok, c.want, c.ok)
		}
	}
	if _, ok := (resolver{}).word(firstWord(t, `~/x`)); ok {
		t.Error("~ without a home directory must not resolve")
	}
	if _, ok := (resolver{}).word(firstWord(t, `$HOME`)); ok {
		t.Error("$HOME without a home directory must not resolve")
	}
}

func TestDirset(t *testing.T) {
	a := dirset{dirs: []string{"/a", "/b"}}
	if got := a.moved("c", true); len(got.dirs) != 2 || got.dirs[0] != "/a/c" || got.unknown {
		t.Errorf("moved = %+v", got)
	}
	if got := a.moved("/abs", true); got.dirs[0] != "/abs" {
		t.Errorf("moved absolute = %+v", got)
	}
	if got := a.moved("", false); len(got.dirs) != 0 || !got.unknown {
		t.Errorf("unknown move = %+v", got)
	}
	u := union(a, dirset{dirs: []string{"/b", "/c"}, unknown: true})
	if len(u.dirs) != 3 || !u.unknown {
		t.Errorf("union = %+v", u)
	}
	if !(dirset{}).empty() || a.empty() || (dirset{unknown: true}).empty() {
		t.Error("empty")
	}
}
