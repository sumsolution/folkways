package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func missing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("%s should not exist", path)
	}
}

func parse(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func matches(t *testing.T, got, pattern string) {
	t.Helper()
	if !strings.Contains(got, pattern) {
		t.Errorf("%q does not contain %q", got, pattern)
	}
}

func TestMismatch(t *testing.T) {
	t.Run("matches a deep subset", func(t *testing.T) {
		if m := mismatch(parse(t, `{"a":{"b":1}}`), parse(t, `{"a":{"b":1,"c":2},"d":3}`), "$"); m != "" {
			t.Error(m)
		}
		matches(t, mismatch(parse(t, `{"a":{"b":1}}`), parse(t, `{"a":{"b":2}}`), "$"), "$.a.b")
	})
	t.Run("treats re: strings as case-insensitive regexes", func(t *testing.T) {
		if m := mismatch(parse(t, `{"r":"re:scope.*api"}`), parse(t, `{"r":"Scope \"API\" is not allowed"}`), "$"); m != "" {
			t.Error(m)
		}
		if mismatch(parse(t, `{"r":"re:scope"}`), parse(t, `{"r":"nothing"}`), "$") == "" {
			t.Error("want a mismatch")
		}
		if mismatch(parse(t, `{"r":"re:scope"}`), parse(t, `{}`), "$") == "" {
			t.Error("want a mismatch for a missing key")
		}
	})
	t.Run("dot matches a newline in a regex", func(t *testing.T) {
		if m := mismatch("re:a.b", "a\nb", "$"); m != "" {
			t.Error(m)
		}
	})
}

func TestFillAndMerge(t *testing.T) {
	t.Run("replaces placeholders in keys and values", func(t *testing.T) {
		got := fill(plain{"${TEAM}/x": []any{"${HEAD7}", 1.0}}, Vars{"TEAM": "/t", "HEAD7": "abc1234"})
		want := plain{"/t/x": []any{"abc1234", 1.0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v", got)
		}
	})
	t.Run("merges objects and replaces arrays", func(t *testing.T) {
		got := merge(parse(t, `{"git":{"scopes":["a"],"merge":"squash"}}`).(plain), parse(t, `{"git":{"scopes":[]}}`).(plain))
		want := parse(t, `{"git":{"scopes":[],"merge":"squash"}}`)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v", got)
		}
	})
}

func TestBuildEnv(t *testing.T) {
	t.Run("builds a team repo, a main clone and a worktree", func(t *testing.T) {
		var f Fixture
		f.Env.Team = optional{Set: true, Value: parse(t, `{"git":{"scopes":[]}}`).(plain)}
		f.Env.Files = map[string]string{"efforts/x/r-${HEAD7}.md": "sha: ${HEAD}\n"}
		vars := buildEnv(t, f)
		matches(t, read(t, filepath.Join(vars["TEAM"], "folkways.yaml")), `"scopes":[]`)
		matches(t, read(t, filepath.Join(vars["TEAM"], ".local/me.yaml")), "projects_dir: "+js(vars["PROJECTS"]))
		if got := read(t, filepath.Join(vars["TEAM"], "efforts/x/r-"+vars["HEAD7"]+".md")); got != "sha: "+vars["HEAD"]+"\n" {
			t.Errorf("review file: %q", got)
		}
		if !strings.HasPrefix(read(t, filepath.Join(vars["WORKTREE"], ".git")), "gitdir:") {
			t.Error("the worktree's .git should be a file")
		}
		if vars["HEAD"] == vars["OLD_HEAD"] {
			t.Error("HEAD should differ from OLD_HEAD")
		}
		if vars["OLD_HEAD7"] != vars["OLD_HEAD"][:7] {
			t.Error("OLD_HEAD7")
		}
	})
	t.Run("omits folkways.yaml when team is null", func(t *testing.T) {
		var f Fixture
		f.Env.Team = optional{Set: true, Null: true}
		missing(t, filepath.Join(buildEnv(t, f)["TEAM"], "folkways.yaml"))
	})
	t.Run("seeds the offered file, or blocks it with a directory", func(t *testing.T) {
		var f Fixture
		f.Env.Offered = []any{"${TEAM}", "/other"}
		seeded := buildEnv(t, f)
		if got := read(t, filepath.Join(seeded["XDG"], "folkways", "offered")); got != seeded["TEAM"]+"\n/other\n" {
			t.Errorf("offered: %q", got)
		}
		f.Env.Offered = "unwritable"
		blocked := buildEnv(t, f)
		if _, err := os.ReadFile(filepath.Join(blocked["XDG"], "folkways", "offered")); err == nil {
			t.Error("offered should not be readable as a file")
		}
		missing(t, filepath.Join(buildEnv(t, Fixture{})["XDG"], "folkways", "offered"))
	})
	t.Run("makes the team repo a git repo with commits and their SHAs as variables", func(t *testing.T) {
		var f Fixture
		f.Env.Commits = []struct {
			Message string            `json:"message"`
			Files   map[string]string `json:"files"`
		}{{Files: map[string]string{"a.md": "a\n"}}, {Files: map[string]string{"b-${COMMIT_1_7}.md": "b\n"}}}
		vars := buildEnv(t, f)
		if git(t, vars["TEAM"], "rev-parse", "HEAD") != vars["TEAM_HEAD"] {
			t.Error("TEAM_HEAD")
		}
		if git(t, vars["TEAM"], "rev-parse", "HEAD~1") != vars["COMMIT_1"] {
			t.Error("COMMIT_1")
		}
		if read(t, filepath.Join(vars["TEAM"], "b-"+vars["COMMIT_1_7"]+".md")) != "b\n" {
			t.Error("placeholder in a later commit's file name")
		}
		if got := git(t, vars["TEAM"], "status", "--porcelain"); got != "" {
			t.Errorf("status: %q", got)
		}
	})
	t.Run("omits me.yaml when me is null", func(t *testing.T) {
		var f Fixture
		f.Env.Me = optional{Set: true, Null: true}
		missing(t, filepath.Join(buildEnv(t, f)["TEAM"], ".local/me.yaml"))
	})
}

func TestHookEnv(t *testing.T) {
	vars := buildEnv(t, Fixture{})
	env := hookEnv(vars, true)
	if env["HOME"] != filepath.Dir(vars["PROJECTS"]) {
		t.Errorf("HOME %q", env["HOME"])
	}
	if !strings.HasPrefix(vars["WORKTREE"], env["HOME"]+"/") || !strings.HasPrefix(vars["MAIN"], env["HOME"]+"/") {
		t.Error("the worktree and main clone should be under HOME")
	}
	if env["CLAUDE_PROJECT_DIR"] != vars["TEAM"] {
		t.Error("CLAUDE_PROJECT_DIR")
	}
	if _, ok := hookEnv(vars, false)["CLAUDE_PROJECT_DIR"]; ok {
		t.Error("CLAUDE_PROJECT_DIR should be unset")
	}
	if hookEnv(vars, "/sub")["CLAUDE_PROJECT_DIR"] != "/sub" {
		t.Error("a string sets CLAUDE_PROJECT_DIR")
	}
	if env["XDG_CONFIG_HOME"] != vars["XDG"] {
		t.Error("XDG_CONFIG_HOME")
	}
}

func TestRunHookEnvironment(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "/developers/own")
	vars := buildEnv(t, Fixture{})
	r := runHook(t, `cat >/dev/null; printf '{"p":"%s","h":"%s"}' "$CLAUDE_PROJECT_DIR" "$HOME"; :`, "x", Step{Payload: plain{}}, vars, false)
	if r.Exit != 0 || r.Stdout != `{"p":"","h":"`+filepath.Dir(vars["PROJECTS"])+`"}` {
		t.Errorf("%+v", r)
	}
}

func TestTranscriptJSONL(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(transcriptJSONL([]any{130000.0, plain{"sidechain": 5.0}})), "\n")
	main, side := parse(t, lines[0]).(plain), parse(t, lines[1]).(plain)
	u := main["message"].(plain)["usage"].(plain)
	if sum := u["input_tokens"].(float64) + u["cache_read_input_tokens"].(float64) + u["cache_creation_input_tokens"].(float64); sum != 130000 {
		t.Errorf("usage sums to %v", sum)
	}
	if main["isSidechain"] != false || side["isSidechain"] != true {
		t.Error("isSidechain")
	}
}

func TestCheck(t *testing.T) {
	vars := Vars{"TEAM": "/nonexistent"}
	t.Run("passes empty stdout when null is expected", func(t *testing.T) {
		if p := check(Expect{}, Result{Stdout: "\n"}, vars); len(p) != 0 {
			t.Error(p)
		}
	})
	t.Run("reports exit code, stdout and file problems", func(t *testing.T) {
		p := check(Expect{Stdout: plain{"a": 1.0}, Exists: []string{"${TEAM}/x"}}, Result{Exit: 2, Stdout: "nope", Stderr: "boom"}, vars)
		if len(p) != 3 {
			t.Fatalf("problems: %v", p)
		}
		matches(t, p[0], "exit 2")
		matches(t, p[2], "/nonexistent/x")
	})
}

func TestNegatedAndListedPatterns(t *testing.T) {
	if m := mismatch("!re:125k", "band 200K", "$"); m != "" {
		t.Error(m)
	}
	matches(t, mismatch("!re:125k", "band 125K", "$"), "must not")
	list := []any{"!re:125K", "!re:150K", "re:200K"}
	if m := mismatch(list, "band 200K", "$"); m != "" {
		t.Error(m)
	}
	if mismatch(list, "125K and 200K", "$") == "" || mismatch(list, "nothing", "$") == "" {
		t.Error("want a mismatch")
	}
	matches(t, mismatch("re:(?!x)y", "y", "$"), "invalid regex")
}
