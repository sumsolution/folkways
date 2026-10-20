package settings

import "testing"

// The settings cases the sync fixtures can't express (fixtures/sync/COVERAGE.md): they need a fragment
// other than the real one, or in-process identity. The fixtures pin the rest.

func parse(t *testing.T, s string) any {
	t.Helper()
	v, err := Parse([]byte(s))
	if err != nil {
		t.Fatalf("Parse(%s): %v", s, err)
	}
	return v
}

func merge(t *testing.T, current, fragment string) string {
	t.Helper()
	out, err := Merge(parse(t, current), parse(t, fragment).(*Object))
	if err != nil {
		t.Fatal(err)
	}
	return Stringify(out)
}

const ours = `{"type":"command","command":"\"${CLAUDE_PROJECT_DIR:-.}\"/node_modules/.bin/folkways-hook x"}`

func TestMergeDropsHooksKeyWhenNothingIsLeft(t *testing.T) {
	// Also the hooks half of "removes files and hooks the new version dropped".
	got := merge(t, `{"a":1,"hooks":{"Stop":[{"hooks":[`+ours+`]}]}}`, `{}`)
	if want := "{\n  \"a\": 1\n}"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestMergeKeepsDisableAllHooksWithoutFragmentValue(t *testing.T) {
	if got, want := merge(t, `{"disableAllHooks":true}`, `{}`), "{\n  \"disableAllHooks\": true\n}"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := merge(t, `{"disableAllHooks":true}`, `{"disableAllHooks":false}`), "{\n  \"disableAllHooks\": false\n}"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestMergeDoesNotChangeItsInput(t *testing.T) {
	const cur = `{"hooks":{"Stop":[{"hooks":[` + ours + `,{"command":"team"}]}]},"permissions":{"ask":["a"]}}`
	const frag = `{"hooks":{"Stop":[{"hooks":[` + ours + `]}]},"permissions":{"ask":["b"]},"disableAllHooks":false}`
	current, fragment := parse(t, cur), parse(t, frag).(*Object)
	before, fragBefore := Stringify(current), Stringify(fragment)
	if _, err := Merge(current, fragment); err != nil {
		t.Fatal(err)
	}
	if Stringify(current) != before || Stringify(fragment) != fragBefore {
		t.Error("Merge changed its input")
	}
}

func TestStringifyMatchesJSONStringify(t *testing.T) {
	// Expected values are node -e 'console.log(JSON.stringify(JSON.parse(s), null, 2))'.
	cases := map[string]string{
		`{"b":1,"a":[],"c":{},"b":2}`:           "{\n  \"b\": 2,\n  \"a\": [],\n  \"c\": {}\n}",
		`[1.50, 1e2, -0, 1e21, 1e-7, 0.000001]`: "[\n  1.5,\n  100,\n  0,\n  1e+21,\n  1e-7,\n  0.000001\n]",
		`"<&> \u0001\t\"\\é"`:                   "\"<&> \\u0001\\t\\\"\\\\é\"",
		`["\\u2028\u2029"]`:                     "[\n  \"\\\\u2028\u2029\"\n]",
		`[null,true,{"x":[{"y":null}]}]`:        "[\n  null,\n  true,\n  {\n    \"x\": [\n      {\n        \"y\": null\n      }\n    ]\n  }\n]",
	}
	for in, want := range cases {
		if got := Stringify(parse(t, in)); got != want {
			t.Errorf("Stringify(%s) = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{``, `{ not json`, `{} {}`, `{"a":1,}`} {
		if _, err := Parse([]byte(bad)); err == nil {
			t.Errorf("Parse(%q) succeeded", bad)
		}
	}
}
