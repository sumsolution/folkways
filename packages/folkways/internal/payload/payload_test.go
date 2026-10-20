package payload

import (
	"strings"
	"testing"
)

func TestBashCommand(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"claude", `{"tool_name":"Bash","tool_input":{"command":"git push"},"cwd":"/x"}`, "git push"},
		{"pascal string input", `{"tool_name":"bash","tool_input":"{\"command\":\"ls\"}"}`, "ls"},
		{"camel", `{"toolName":"bash","toolArgs":"{\"command\":\"git status\"}"}`, "git status"},
		{"other tool", `{"tool_name":"Read","tool_input":{"command":"x"}}`, ""},
		{"no tool", `{}`, ""},
	}
	for _, c := range cases {
		p, err := Read(strings.NewReader(c.in))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := p.BashCommand(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
	if _, err := Read(strings.NewReader("not json")); err == nil {
		t.Error("bad JSON should be an error")
	}
}

func TestSessionID(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"snake", `{"session_id":"1111-aa"}`, "1111-aa"},
		{"camel", `{"sessionId":"2222-bb"}`, "2222-bb"},
		{"none", `{}`, ""},
		{"separator", `{"session_id":"../x"}`, ""},
		{"dotdot", `{"session_id":".."}`, ""},
		{"backslash", `{"session_id":"a\\b"}`, ""},
		{"newline", `{"session_id":"a\nb"}`, ""},
	}
	for _, c := range cases {
		p, _ := Read(strings.NewReader(c.in))
		if got := p.SessionID(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
