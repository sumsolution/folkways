package gitguard

import (
	"reflect"
	"testing"
)

func TestSplitCommands(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"git status", []string{"git status"}},
		{"cd a && git push; git fetch || echo x", []string{"cd a", "git push", "git fetch", "echo x"}},
		{`echo "a && b" && git push`, []string{`echo "a && b"`, "git push"}},
		{`echo 'a; b'`, []string{`echo 'a; b'`}},
		{`git commit -m "$(echo a && b)" && git push`, []string{`git commit -m "$(echo a && b)"`, "git push"}},
		{"git commit -m \"$(cat <<'EOF'\nfeat: x && y\nEOF\n)\" && git push", []string{"git commit -m \"$(cat <<'EOF'\nfeat: x && y\nEOF\n)\"", "git push"}},
		{"a\nb", []string{"a", "b"}},
		{`echo a\&\&b`, []string{`echo a\&\&b`}},
		{"", nil},
	}
	for _, c := range cases {
		if got := splitCommands(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitCommands(%q) = %q, want %q", c.in, got, c.want)
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
