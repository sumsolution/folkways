package contract

// Harness for the black-box hook contract suite (README.md). Builds a synthetic team repo per fixture,
// runs `$FOLKWAYS_HOOK_CMD <hook>` through `sh -c` like the hook runners do, and checks the result.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

var hooks = []string{"context-band", "precompact", "session-start", "git-guard"}

// plain is a JSON object. Values are what encoding/json produces: map, slice, string, float64, bool, nil.
type plain = map[string]any

// optional is an env field that can be absent, null or an object (`env.team`, `env.me`).
type optional struct {
	Set, Null bool
	Value     plain
}

func (o *optional) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(bytes.TrimSpace(b)) == "null" {
		o.Null = true
		return nil
	}
	return json.Unmarshal(b, &o.Value)
}

// Fixture is one fixtures/<hook>/<case>.json file; the fields are documented in README.md.
type Fixture struct {
	Description string `json:"description"`
	Env         struct {
		Team  optional          `json:"team"`
		Me    optional          `json:"me"`
		Files map[string]string `json:"files"`
		// ClaudeProjectDir: absent or true = the team repo, false = unset, a string = that path.
		ClaudeProjectDir any `json:"claudeProjectDir"`
		Commits          []struct {
			Message string            `json:"message"`
			Files   map[string]string `json:"files"`
		} `json:"commits"`
		// Transcript lines: a number, {"sidechain": n} or an object written as is.
		Transcript []any `json:"transcript"`
		// Offered: a list of lines, or "unwritable".
		Offered any `json:"offered"`
	} `json:"env"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Payload    any    `json:"payload"`
	Cwd        string `json:"cwd"`
	Transcript []any  `json:"transcript"`
	Expect     Expect `json:"expect"`
}

type Expect struct {
	Exit     *int              `json:"exit"`
	Stdout   plain             `json:"stdout"` // nil: stdout is empty
	Exists   []string          `json:"exists"`
	Absent   []string          `json:"absent"`
	Contains map[string]string `json:"contains"`
}

// Vars are the placeholders: TEAM, PROJECTS, MAIN, WORKTREE, HEAD, HEAD7, OLD_HEAD, OLD_HEAD7, TRANSCRIPT,
// XDG, and COMMIT_n, COMMIT_n_7, TEAM_HEAD, TEAM_HEAD7 when the fixture has env.commits.
type Vars map[string]string

var baseTeam = plain{
	"team":    plain{"name": "Example Team", "slug": "example"},
	"members": []any{plain{"handle": "alex", "github": "alex-gh", "name": "Alex Doe"}},
	"git":     plain{"commit_convention": "angular", "scopes": []any{"web-app", "auth", "ci"}, "branch_pattern": "{type}/{ticket}-{slug}", "merge": "squash"},
	"review":  plain{"trigger": "pre-push"},
	"handoff": plain{"bands": []any{125000.0, 150000.0, 200000.0}},
}

type namedFixture struct {
	Name string
	Fixture
}

func loadFixtures(t testing.TB, hook string) []namedFixture {
	t.Helper()
	dir := filepath.Join("fixtures", hook)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []namedFixture
	for _, e := range entries { // ReadDir sorts by name
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		f := namedFixture{Name: strings.TrimSuffix(e.Name(), ".json")}
		if err := json.Unmarshal(b, &f.Fixture); err != nil {
			t.Fatalf("%s/%s: %v", hook, e.Name(), err)
		}
		out = append(out, f)
	}
	return out
}

func isPlain(v any) bool { _, ok := v.(plain); return ok }

// merge deep-merges objects; arrays and scalars are replaced.
func merge(base, over plain) plain {
	out := plain{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range over {
		if ov, ok := v.(plain); ok && isPlain(out[k]) {
			out[k] = merge(out[k].(plain), ov)
		} else {
			out[k] = v
		}
	}
	return out
}

var placeholder = regexp.MustCompile(`\$\{(\w+)\}`)

func fillString(s string, vars Vars) string {
	return placeholder.ReplaceAllStringFunc(s, func(m string) string {
		if v, ok := vars[m[2:len(m)-1]]; ok {
			return v
		}
		return m
	})
}

// fill replaces `${NAME}` placeholders in every string (keys too) inside value; unknown names stay.
func fill(value any, vars Vars) any {
	switch v := value.(type) {
	case string:
		return fillString(v, vars)
	case []any:
		out := make([]any, len(v))
		for i, e := range v {
			out[i] = fill(e, vars)
		}
		return out
	case plain:
		out := make(plain, len(v))
		for k, e := range v {
			out[fillString(k, vars)] = fill(e, vars)
		}
		return out
	}
	return value
}

func fillMap(m map[string]string, vars Vars) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[fillString(k, vars)] = fillString(v, vars)
	}
	return out
}

func fillList(l []string, vars Vars) []string {
	out := make([]string, len(l))
	for i, s := range l {
		out[i] = fillString(s, vars)
	}
	return out
}

func js(v any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Sprint(v)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

type missingValue struct{}

// mismatch is a deep subset match. It returns the first mismatch as a path and reason, or "".
// Strings starting `re:` are case-insensitive, dot-all regexes.
func mismatch(expected, actual any, path string) string {
	if s, ok := expected.(string); ok && isPattern(s) {
		return matchPattern(s, actual, path)
	}
	// An array of patterns, against a string: every one must hold.
	if l, ok := expected.([]any); ok && len(l) > 0 {
		if _, isStr := actual.(string); isStr && allPatterns(l) {
			for _, p := range l {
				if m := matchPattern(p.(string), actual, path); m != "" {
					return m
				}
			}
			return ""
		}
	}
	switch e := expected.(type) {
	case plain:
		a, ok := actual.(plain)
		if !ok {
			return fmt.Sprintf("%s: expected an object, got %s", path, jsOrUndefined(actual))
		}
		keys := make([]string, 0, len(e))
		for k := range e {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			var av any = missingValue{}
			if x, ok := a[k]; ok {
				av = x
			}
			if m := mismatch(e[k], av, path+"."+k); m != "" {
				return m
			}
		}
		return ""
	case []any:
		a, ok := actual.([]any)
		if !ok || len(a) != len(e) {
			return fmt.Sprintf("%s: expected %s, got %s", path, js(e), jsOrUndefined(actual))
		}
		for i, v := range e {
			if m := mismatch(v, a[i], fmt.Sprintf("%s[%d]", path, i)); m != "" {
				return m
			}
		}
		return ""
	}
	if _, miss := actual.(missingValue); !miss && expected == actual {
		return ""
	}
	return fmt.Sprintf("%s: expected %s, got %s", path, js(expected), jsOrUndefined(actual))
}

func isPattern(s string) bool { return strings.HasPrefix(s, "re:") || strings.HasPrefix(s, "!re:") }

func allPatterns(l []any) bool {
	for _, v := range l {
		if s, ok := v.(string); !ok || !isPattern(s) {
			return false
		}
	}
	return true
}

// matchPattern checks `re:<regex>` (must match) or `!re:<regex>` (must not match) against a string; both are
// case-insensitive and dot matches newline.
func matchPattern(p string, actual any, path string) string {
	negate := strings.HasPrefix(p, "!")
	pat := strings.TrimPrefix(strings.TrimPrefix(p, "!"), "re:")
	a, isStr := actual.(string)
	if !isStr {
		return fmt.Sprintf("%s: %s does not match /%s/i", path, jsOrUndefined(actual), pat)
	}
	re, err := regexp.Compile("(?is)" + pat)
	if err != nil {
		return fmt.Sprintf("%s: invalid regex /%s/: %v", path, pat, err)
	}
	if re.MatchString(a) == negate {
		if negate {
			return fmt.Sprintf("%s: %s matches /%s/i, which it must not", path, js(a), pat)
		}
		return fmt.Sprintf("%s: %s does not match /%s/i", path, js(a), pat)
	}
	return ""
}

func jsOrUndefined(v any) string {
	if _, ok := v.(missingValue); ok {
		return "undefined"
	}
	return js(v)
}

// yamlText is a minimal YAML writer for the config shapes above (flow style, JSON-compatible), keys sorted.
func yamlText(v plain) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %s\n", k, js(v[k]))
	}
	return b.String()
}

func assistantLine(tokens int, sidechain bool) plain {
	// Realistic split: most of the context is a cache read.
	input := min(tokens, 8)
	creation := min(tokens-input, 1500)
	return plain{
		"type":        "assistant",
		"isSidechain": sidechain,
		"message": plain{
			"role": "assistant", "model": "claude-example",
			"usage": plain{"input_tokens": input, "cache_creation_input_tokens": creation, "cache_read_input_tokens": tokens - input - creation, "output_tokens": 120},
		},
	}
}

// transcriptJSONL: a number is a main-chain assistant line with that many tokens in context, {"sidechain": n} a
// subagent line, any other object is written as is.
func transcriptJSONL(lines []any) string {
	var b strings.Builder
	for _, l := range lines {
		switch v := l.(type) {
		case float64:
			l = assistantLine(int(v), false)
		case int:
			l = assistantLine(v, false)
		case plain:
			if n, ok := v["sidechain"].(float64); ok {
				l = assistantLine(int(n), true)
			} else if n, ok := v["sidechain"].(int); ok {
				l = assistantLine(n, true)
			}
		}
		b.WriteString(js(l) + "\n")
	}
	if len(lines) == 0 {
		return "\n"
	}
	return b.String()
}

func mustWrite(t testing.TB, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// git runs git in dir and returns trimmed stdout.
func git(t testing.TB, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, stderr.String())
	}
	return strings.TrimSpace(string(out))
}

// tmpDir is a temp dir removed after the test, symlinks resolved so paths compare equal to git's.
func tmpDir(t testing.TB) string {
	t.Helper()
	// Not t.TempDir: its name holds the test name, which fixtures' text checks can trip over.
	d, err := os.MkdirTemp("", "folkways-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	if d, err = filepath.EvalSymlinks(d); err != nil {
		t.Fatal(err)
	}
	return d
}

// buildEnv creates the synthetic environment: a team repo with folkways.yaml (unless env.team is null) and
// .local/me.yaml, and a projects dir holding a main clone (`web-app`, two commits on main) and a worktree of it
// (`.worktrees/web-app/feat-orders`).
func buildEnv(t testing.TB, f Fixture) Vars {
	t.Helper()
	root := tmpDir(t)
	team := filepath.Join(root, "team")
	projects := filepath.Join(root, "projects")
	main := filepath.Join(projects, "web-app")
	worktree := filepath.Join(projects, ".worktrees", "web-app", "feat-orders")
	for _, d := range []string{team, main} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	git(t, main, "init", "-q", "-b", "main")
	mustWrite(t, filepath.Join(main, "README.md"), "one\n")
	git(t, main, "add", "-A")
	git(t, main, "commit", "-q", "-m", "chore: first")
	oldHead := git(t, main, "rev-parse", "HEAD")
	mustWrite(t, filepath.Join(main, "README.md"), "two\n")
	git(t, main, "commit", "-q", "-am", "chore: second")
	git(t, main, "worktree", "add", "-q", "-b", "feat/EX-1-orders", worktree)
	mustWrite(t, filepath.Join(worktree, "orders.md"), "orders\n")
	git(t, worktree, "add", "-A")
	git(t, worktree, "commit", "-q", "-m", "feat(web-app): add orders")
	head := git(t, worktree, "rev-parse", "HEAD")

	vars := Vars{
		"TEAM": team, "PROJECTS": projects, "MAIN": main, "WORKTREE": worktree,
		"HEAD": head, "HEAD7": head[:7], "OLD_HEAD": oldHead, "OLD_HEAD7": oldHead[:7],
		"TRANSCRIPT": filepath.Join(root, "transcript.jsonl"), "XDG": filepath.Join(root, "xdg"),
	}
	env := f.Env
	if !env.Team.Null {
		mustWrite(t, filepath.Join(team, "folkways.yaml"), yamlText(merge(baseTeam, env.Team.Value)))
	}
	if !env.Me.Null {
		me := merge(plain{"handle": "alex", "projects_dir": projects, "tool": "claude"}, fill(env.Me.Value, vars).(plain))
		mustWrite(t, filepath.Join(team, ".local", "me.yaml"), yamlText(me))
	}
	for path, content := range fillMap(env.Files, vars) {
		mustWrite(t, filepath.Join(team, path), content)
	}
	switch o := env.Offered.(type) {
	case string:
		if o == "unwritable" {
			if err := os.MkdirAll(filepath.Join(vars["XDG"], "folkways", "offered"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	case []any:
		var b strings.Builder
		for _, l := range o {
			b.WriteString(fillString(fmt.Sprint(l), vars) + "\n")
		}
		mustWrite(t, filepath.Join(vars["XDG"], "folkways", "offered"), b.String())
	}
	if env.Commits != nil {
		git(t, team, "init", "-q", "-b", "main")
		mustWrite(t, filepath.Join(team, ".git", "info", "exclude"), ".local/\n")
		git(t, team, "add", "-A")
		git(t, team, "commit", "-q", "-m", "chore: init")
		for i, c := range env.Commits {
			for path, content := range fillMap(c.Files, vars) {
				mustWrite(t, filepath.Join(team, path), content)
			}
			git(t, team, "add", "-A")
			msg := c.Message
			if msg == "" {
				msg = fmt.Sprintf("chore: commit %d", i+1)
			}
			git(t, team, "commit", "-q", "-m", msg)
			sha := git(t, team, "rev-parse", "HEAD")
			vars[fmt.Sprintf("COMMIT_%d", i+1)] = sha
			vars[fmt.Sprintf("COMMIT_%d_7", i+1)] = sha[:7]
		}
		vars["TEAM_HEAD"] = git(t, team, "rev-parse", "HEAD")
		vars["TEAM_HEAD7"] = vars["TEAM_HEAD"][:7]
	}
	if env.Transcript != nil {
		mustWrite(t, vars["TRANSCRIPT"], transcriptJSONL(env.Transcript))
	}
	return vars
}

// hookEnv is the environment the hook process gets on top of the inherited one. HOME is the temp root that
// holds the team repo and the projects directory, so fixtures can `cd ~/projects/web-app` or
// `cd "$HOME/projects/web-app"`. claudeProjectDir: true = the team repo, false = unset, a string = that path.
// CLAUDE_PROJECT_DIR is always dropped from the inherited environment (see runHook).
func hookEnv(vars Vars, claudeProjectDir any) map[string]string {
	env := map[string]string{
		"HOME":            filepath.Dir(vars["PROJECTS"]),
		"XDG_CONFIG_HOME": vars["XDG"], // never the developer's own offered file
	}
	switch c := claudeProjectDir.(type) {
	case string:
		env["CLAUDE_PROJECT_DIR"] = c
	case bool:
		if c {
			env["CLAUDE_PROJECT_DIR"] = vars["TEAM"]
		}
	}
	return env
}

type Result struct {
	Exit           int // -1: killed by the timeout or not started
	Stdout, Stderr string
}

const hookTimeout = 10 * time.Second

// runHook runs `/bin/sh -c "<command> <hook>"` with the payload on stdin, as both tools do.
func runHook(t testing.TB, command, hook string, step Step, vars Vars, claudeProjectDir any) Result {
	t.Helper()
	if step.Transcript != nil {
		mustWrite(t, vars["TRANSCRIPT"], transcriptJSONL(step.Transcript))
	}
	over := hookEnv(vars, fill(claudeProjectDir, vars))
	var environ []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if _, replaced := over[k]; !replaced && k != "CLAUDE_PROJECT_DIR" {
			environ = append(environ, kv)
		}
	}
	for k, v := range over {
		environ = append(environ, k+"="+v)
	}
	cwd := vars["TEAM"]
	if step.Cwd != "" {
		cwd = fillString(step.Cwd, vars)
	}
	payload, err := json.Marshal(fill(step.Payload, vars))
	if err != nil {
		t.Fatal(err)
	}
	return execHook(command+" "+hook, cwd, environ, payload)
}

func execHook(shellCommand, cwd string, environ []string, stdin []byte) Result {
	ctx, cancel := context.WithTimeout(context.Background(), hookTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", shellCommand)
	cmd.Dir = cwd
	cmd.Env = environ
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.WaitDelay = time.Second
	err := cmd.Run()
	r := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		r.Exit = -1
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() >= 0 {
			r.Exit = ee.ExitCode()
		} else if r.Stderr == "" {
			r.Stderr = err.Error()
		}
	}
	return r
}

// check returns every problem with a result, empty when it meets the expectation.
func check(expect Expect, r Result, vars Vars) []string {
	var problems []string
	want := fill(toPlain(expect), vars).(plain)
	wantExit := 0
	if expect.Exit != nil {
		wantExit = *expect.Exit
	}
	if r.Exit != wantExit {
		problems = append(problems, fmt.Sprintf("exit %d, want %d (stderr: %s)", r.Exit, wantExit, truncate(strings.TrimSpace(r.Stderr), 200)))
	}
	if stdout, ok := want["stdout"].(plain); !ok || stdout == nil {
		if strings.TrimSpace(r.Stdout) != "" {
			problems = append(problems, "stdout should be empty, got: "+truncate(strings.TrimSpace(r.Stdout), 300))
		}
	} else {
		var got any
		if err := json.Unmarshal([]byte(r.Stdout), &got); err != nil {
			problems = append(problems, "stdout is not JSON: "+truncate(strings.TrimSpace(r.Stdout), 300))
		} else if m := mismatch(stdout, got, "$"); m != "" {
			problems = append(problems, "stdout "+m)
		}
	}
	exists, _ := want["exists"].([]any)
	for _, p := range exists {
		if _, err := os.Stat(p.(string)); err != nil {
			problems = append(problems, "missing file "+p.(string))
		}
	}
	absent, _ := want["absent"].([]any)
	for _, p := range absent {
		if _, err := os.Stat(p.(string)); err == nil {
			problems = append(problems, "unexpected file "+p.(string))
		}
	}
	contains, _ := want["contains"].(plain)
	paths := make([]string, 0, len(contains))
	for p := range contains {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		re := contains[p].(string)
		text, err := os.ReadFile(p)
		rx, rxErr := regexp.Compile("(?is)" + re)
		if err != nil || rxErr != nil || !rx.Match(text) {
			problems = append(problems, fmt.Sprintf("%s does not match /%s/i", p, re))
		}
	}
	return problems
}

// toPlain turns an Expect into the generic form so fill can replace placeholders in it.
func toPlain(e Expect) plain {
	out := plain{"stdout": e.Stdout}
	if e.Stdout == nil {
		out["stdout"] = nil
	}
	out["exists"] = toAny(e.Exists)
	out["absent"] = toAny(e.Absent)
	c := plain{}
	for k, v := range e.Contains {
		c[k] = v
	}
	out["contains"] = c
	return out
}

func toAny(l []string) []any {
	out := make([]any, len(l))
	for i, s := range l {
		out[i] = s
	}
	return out
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
