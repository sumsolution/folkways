package doctor

import (
	"slices"
	"testing"
)

var known = []string{"context-band", "git-guard", "precompact", "session-start"}

func TestHookValueFindings(t *testing.T) {
	cases := []struct {
		name, yaml string
		want       []Finding
	}{
		{"no hooks block", "mode: team\n", nil},
		{"null map", "hooks: ~ # none\n", nil},
		{"true and false with comments", "hooks:\n  git-guard: false # off\n  precompact: true\n", nil},
		{"CRLF", "hooks:\r\n  git-guard: \"false\"\r\n  precompact: false\r\n", []Finding{
			{"error", `folkways.yaml hooks.git-guard: "false" is not a boolean, so the hook stays on; write true or false`}}},
		{"tab", "hooks:\n\tgit-guard: false\n", []Finding{
			{"error", "folkways.yaml hooks: line 2 is indented with a tab, which YAML forbids, so the hooks ignore folkways.yaml; indent with spaces"}}},
		{"inline map", "hooks: { git-guard: false }\n", []Finding{
			{"warning", "folkways.yaml hooks: write one hook per line (git-guard: false) so doctor can check the values"}}},
		{"YAML 1.1 word", "hooks:\n  git-guard: off\n", []Finding{
			{"error", "folkways.yaml hooks.git-guard: off reads as false (hook off); write false"}}},
		{"unknown name", "hooks:\n  nope: false\n", []Finding{
			{"warning", "folkways.yaml hooks.nope: no hook by that name (hooks: context-band, git-guard, precompact, session-start)"}}},
		{"alias", "hooks:\n  git-guard: *off\n", []Finding{
			{"error", "folkways.yaml hooks.git-guard: doctor can't check *off (a tag, anchor or alias); write true or false"}}},
		{"empty", "hooks:\n  git-guard:\n", []Finding{
			{"error", "folkways.yaml hooks.git-guard: (empty) reads as false (hook off); write false"}}},
	}
	for _, c := range cases {
		if got := hookValueFindings(c.yaml, known); !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStripComment(t *testing.T) {
	for in, want := range map[string]string{
		`false # c`:   "false",
		`"a # b" # c`: `"a # b"`,
		`a#b`:         "a#b",
		`# only`:      "",
		`'x' y`:       `'x'`,
	} {
		if got := stripComment(in); got != want {
			t.Errorf("stripComment(%q) = %q, want %q", in, got, want)
		}
	}
}
