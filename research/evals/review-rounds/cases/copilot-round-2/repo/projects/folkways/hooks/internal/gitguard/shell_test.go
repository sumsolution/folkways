package gitguard

import (
	"reflect"
	"strings"
	"testing"
)

// flat renders parsed elements as "text<op>" for comparison; a subshell is rendered as "(...)" around its content.
func flat(es []elem) string {
	var b strings.Builder
	for _, e := range es {
		if e.sub != nil {
			b.WriteString("(" + flat(e.sub) + ")")
		} else {
			b.WriteString(e.text)
		}
		b.WriteString("<" + e.op + ">")
	}
	return b.String()
}

func TestParseSeq(t *testing.T) {
	cases := []struct{ in, want string }{
		{"git status", "git status<>"},
		{"cd a && git push; git fetch || echo x", "cd a<&&>git push<;>git fetch<||>echo x<>"},
		{`echo "a && b" && git push`, `echo "a && b"<&&>git push<>`},
		{`echo 'a; b'`, `echo 'a; b'<>`},
		{`git commit -m "$(echo a && b)" && git push`, `git commit -m "$(echo a && b)"<&&>git push<>`},
		{"a\nb", "a<\n>b<>"},
		{"a &&\nb", "a<&&>b<>"},
		{`echo a\&\&b`, `echo a\&\&b<>`},
		{"(cd a && git push) && git fetch", "(cd a<&&>git push<>)<&&>git fetch<>"},
		{"(cd a; git push)", "(cd a<;>git push<>)<>"},
		{"{ cd a; git push; }", "{ cd a<;>git push<;>}<>"},
		{"cat <(git push) | wc", "cat <(git push)<|>wc<>"},
		{"echo `a && b` && c", "echo `a && b`<&&>c<>"},
		{"", ""},
	}
	for _, c := range cases {
		els, _ := parseSeq(c.in, 0, false)
		if got := flat(els); got != c.want {
			t.Errorf("parseSeq(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestStripHeredocs(t *testing.T) {
	cases := []struct{ in, want string }{
		{"cat <<EOF > n.md\ngit push\nEOF\ngit status", "cat <<EOF > n.md\ngit status"},
		{"cat <<'EOF'\ngit push\nEOF", "cat <<'EOF'\n"},
		{"cat <<\"EOF\"\ngit push\nEOF\n", "cat <<\"EOF\"\n"},
		{"cat <<-EOF\n\tgit push\n\tEOF\necho ok", "cat <<-EOF\necho ok"},
		{"cat <<EOF\n  EOF\ngit push\nEOF\necho ok", "cat <<EOF\necho ok"}, // an indented EOF is not the terminator
		{"cat <<A <<B\none\nA\ntwo\nB\necho ok", "cat <<A <<B\necho ok"},
		{"cat <<EOF && git push\nbody\nEOF", "cat <<EOF && git push\n"},
		{"cat <<EOF\nnever ends\ngit push", "cat <<EOF\n"},
		{"cat <<< 'git push'", "cat <<< 'git push'"},
		{`echo "<<EOF"` + "\ngit push", `echo "<<EOF"` + "\ngit push"},
		{"git commit -m \"$(cat <<'EOF'\nfeat: x\nEOF\n)\"", "git commit -m \"$(cat <<'EOF'\nfeat: x\nEOF\n)\""},
	}
	for _, c := range cases {
		if got := stripHeredocs(c.in); got != c.want {
			t.Errorf("stripHeredocs(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSubstitutions(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"echo hi", nil},
		{"echo $(git push)", []string{"git push"}},
		{`echo "a $(git push) b"`, []string{"git push"}},
		{"echo 'a $(git push)'", nil},
		{"echo `git push`", []string{"git push"}},
		{`echo "x ` + "`git push`" + `"`, []string{"git push"}},
		{"cat <(git push)", []string{"git push"}},
		{"echo $(a $(b))", []string{"a $(b)"}},
		{`echo \$(git push)`, nil},
		{"echo $(unterminated", []string{"unterminated"}},
	}
	for _, c := range cases {
		if got := substitutions(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("substitutions(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"git commit -m msg", []string{"git", "commit", "-m", "msg"}},
		{`git commit -m "feat: a b"`, []string{"git", "commit", "-m", "feat: a b"}},
		{`git commit -m 'feat: "a"'`, []string{"git", "commit", "-m", `feat: "a"`}},
		{`a\ b c`, []string{"a b", "c"}},
		{`x "a \"q\" b"`, []string{"x", `a "q" b`}},
		{`git commit -m "$(echo hi)"`, []string{"git", "commit", "-m", "$(echo hi)"}},
		{"  a\t b  ", []string{"a", "b"}},
	}
	for _, c := range cases {
		if got := words(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("words(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSubjectOf(t *testing.T) {
	msg := "$(cat <<'EOF'\nfeat(cli): add x\n\nbody\nEOF\n)"
	if got, ok := subjectOf(msg); !ok || got != "feat(cli): add x" {
		t.Errorf("heredoc subject = %q, %v", got, ok)
	}
	if _, ok := subjectOf("$(date)"); ok {
		t.Error("plain command substitution should not be evaluable")
	}
	if got, ok := subjectOf("\n  fix: y\nmore"); !ok || got != "fix: y" {
		t.Errorf("multi-line subject = %q, %v", got, ok)
	}
}
