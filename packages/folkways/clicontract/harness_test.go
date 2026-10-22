// Package clicontract is the black-box contract suite for the folkways CLI (ADR 0011). It runs any
// implementation as a subprocess against the JSON fixtures in fixtures/; README.md has the format.
package clicontract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
)

// Fixture is one case: steps run in order in the same temp root.
type Fixture struct {
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}

// Step sets up files and git state, then optionally runs the CLI and checks the result.
type Step struct {
	Files  map[string]*string `json:"files"` // written before the call; null deletes
	Git    [][]string         `json:"git"`   // git commands run in ${TEAM} before the call
	Args   []string           `json:"args"`  // nil: no CLI call
	Cwd    string             `json:"cwd"`
	Env    map[string]*string `json:"env"` // null unsets
	Stdin  string             `json:"stdin"`
	Expect *Expect            `json:"expect"`
}

// Expect is checked after the call. Unset fields are not checked, except exit (default 0).
type Expect struct {
	Exit      *int           `json:"exit"`
	Stdout    *string        `json:"stdout"`
	Stderr    *string        `json:"stderr"`
	Files     map[string]any `json:"files"` // string: exact or "re:"; null: absent; true: exists
	Tree      []string       `json:"tree"`
	Unchanged bool           `json:"unchanged"`
	Git       []GitCheck     `json:"git"`
}

// GitCheck runs git in ${TEAM} and compares its trimmed stdout.
type GitCheck struct {
	Args   []string `json:"args"`
	Stdout string   `json:"stdout"`
}

var (
	repoRoot = mustAbs("../../..")
	// versionPkg holds the CLI's version; changesets bump it.
	versionPkg = filepath.Join(repoRoot, "packages/folkways/package.json")
)

func mustAbs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	return a
}

// builtCLI is the Go CLI TestMain built when FOLKWAYS_CLI_CMD is unset.
var builtCLI string

// cliCommand is the command under test: FOLKWAYS_CLI_CMD, else the Go CLI TestMain built from this checkout.
// The CLI runs in a temp dir, so a relative path in the first word is refused: go test runs in the package
// dir, not yours.
func cliCommand() (string, error) {
	c := os.Getenv("FOLKWAYS_CLI_CMD")
	if c == "" {
		return shellQuote(builtCLI), nil
	}
	f := strings.Fields(c)
	if len(f) == 0 {
		return "", fmt.Errorf("FOLKWAYS_CLI_CMD is blank")
	}
	if strings.Contains(f[0], "/") && !filepath.IsAbs(f[0]) {
		return "", fmt.Errorf("FOLKWAYS_CLI_CMD: %s is relative; use an absolute path", f[0])
	}
	return c, nil
}

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run builds ./cmd/folkways into a temp dir, unless FOLKWAYS_CLI_CMD names the binary, then runs the tests.
// The build carries the version in packages/folkways/package.json, which the fixtures expect from `version`.
func run(m *testing.M) int {
	if os.Getenv("FOLKWAYS_CLI_CMD") == "" {
		dir, err := os.MkdirTemp("", "folkways-clicontract-bin-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer os.RemoveAll(dir)
		raw, err := os.ReadFile(versionPkg)
		var pkg struct {
			Version string `json:"version"`
		}
		if err == nil {
			err = json.Unmarshal(raw, &pkg)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, versionPkg+":", err)
			return 1
		}
		builtCLI = filepath.Join(dir, "folkways")
		build := exec.Command("go", "build", "-ldflags", "-X main.version="+pkg.Version, "-o", builtCLI, "./cmd/folkways")
		build.Dir = filepath.Join(repoRoot, "packages/folkways")
		build.Stdout, build.Stderr = os.Stdout, os.Stderr
		if err := build.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "building ./cmd/folkways:", err)
			return 1
		}
	}
	return m.Run()
}

func TestContract(t *testing.T) {
	versions := readVersions(t)
	cli, err := cliCommand()
	if err != nil {
		t.Fatal(err)
	}
	versions["CLI"] = cli
	files, err := filepath.Glob("fixtures/*/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no fixtures: %v", err)
	}
	for _, file := range files {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.ToSlash(file), "fixtures/"), ".json")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runFixture(t, file, versions)
		})
	}
}

func readVersions(t *testing.T) map[string]string {
	raw, err := os.ReadFile(versionPkg)
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	return map[string]string{"VERSION": pkg.Version}
}

func loadFixture(t *testing.T, file string) Fixture {
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var f Fixture
	if err := dec.Decode(&f); err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	if dec.More() {
		t.Fatalf("%s: more than one JSON value", file)
	}
	if f.Description == "" || len(f.Steps) == 0 {
		t.Fatalf("%s: needs a description and at least one step", file)
	}
	return f
}

// env is one case's world: temp root, placeholders and the CLI's environment.
type env struct {
	t     *testing.T
	vars  map[string]string
	plain *strings.Replacer
	regex *strings.Replacer
	base  map[string]string
}

func newEnv(t *testing.T, versions map[string]string) *env {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	vars := map[string]string{
		"ROOT": root, "TEAM": filepath.Join(root, "team"), "HOME": filepath.Join(root, "home"),
		"BIN": filepath.Join(root, "bin"),
	}
	for k, v := range versions {
		vars[k] = v
	}
	vars["GIT"], _ = exec.LookPath("git") // for a fixture's own git wrapper in ${BIN}
	var plain, regex []string
	for k, v := range vars {
		plain = append(plain, "${"+k+"}", v)
		regex = append(regex, "${"+k+"}", regexp.QuoteMeta(v))
	}
	e := &env{t: t, vars: vars, plain: strings.NewReplacer(plain...), regex: strings.NewReplacer(regex...)}
	for _, d := range []string{vars["HOME"], vars["BIN"]} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	e.writeStubs()
	// Built from scratch: nothing leaks in from the caller (npm_config_user_agent under pnpm, git config).
	e.base = map[string]string{
		"PATH": vars["BIN"] + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME": vars["HOME"], "XDG_CONFIG_HOME": filepath.Join(vars["HOME"], ".config"), "TMPDIR": os.TempDir(), "TZ": "UTC", "LC_ALL": "C",
		"GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1",
		"GIT_AUTHOR_NAME": "test", "GIT_AUTHOR_EMAIL": "test@example.com",
		"GIT_COMMITTER_NAME": "test", "GIT_COMMITTER_EMAIL": "test@example.com",
		"GIT_AUTHOR_DATE": "2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE": "2026-01-01T00:00:00Z",
	}
	return e
}

func (e *env) sub(s string) string { return e.plain.Replace(s) }

// path resolves a fixture path: absolute after substitution, else relative to ${TEAM}.
func (e *env) path(p string) string {
	p = e.sub(p)
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(e.vars["TEAM"], p)
}

// match compares got with a fixture value: exact, or a regex after "re:" (placeholders regex-quoted).
func (e *env) match(want, got string) bool {
	if re, ok := strings.CutPrefix(want, "re:"); ok {
		r, err := regexp.Compile(e.regex.Replace(re))
		if err != nil {
			e.t.Fatalf("bad regex %q: %v", re, err)
		}
		return r.MatchString(got)
	}
	return e.sub(want) == got
}

func (e *env) environ(extra map[string]*string) []string {
	m := maps.Clone(e.base)
	for k, v := range extra {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = e.sub(*v)
		}
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

func runFixture(t *testing.T, file string, versions map[string]string) {
	f := loadFixture(t, file)
	e := newEnv(t, versions)
	for i, s := range f.Steps {
		e.setup(i, s)
		before := e.snapshot()
		var stdout, stderr string
		code := 0
		if s.Args != nil {
			stdout, stderr, code = e.run(s)
		}
		if s.Expect == nil {
			if s.Args != nil && code != 0 {
				t.Fatalf("step %d: exit %d\nstdout:\n%s\nstderr:\n%s", i+1, code, stdout, stderr)
			}
			continue
		}
		x := s.Expect
		want := 0
		if x.Exit != nil {
			want = *x.Exit
		}
		if s.Args != nil && code != want {
			t.Errorf("step %d: exit %d, want %d\nstderr:\n%s", i+1, code, want, stderr)
		}
		if x.Stdout != nil && !e.match(*x.Stdout, stdout) {
			t.Errorf("step %d: stdout:\n%s\nwant:\n%s", i+1, stdout, e.sub(*x.Stdout))
		}
		if x.Stderr != nil && !e.match(*x.Stderr, stderr) {
			t.Errorf("step %d: stderr:\n%s\nwant:\n%s", i+1, stderr, e.sub(*x.Stderr))
		}
		e.checkFiles(i, x.Files)
		if x.Tree != nil {
			e.checkTree(i, x.Tree)
		}
		if x.Unchanged {
			e.checkUnchanged(i, before)
		}
		for _, g := range x.Git {
			if out := e.git(g.Args...); !e.match(g.Stdout, out) {
				t.Errorf("step %d: git %s:\n%s\nwant:\n%s", i+1, strings.Join(g.Args, " "), out, e.sub(g.Stdout))
			}
		}
		if t.Failed() {
			t.FailNow()
		}
	}
}

func (e *env) setup(i int, s Step) {
	keys := make([]string, 0, len(s.Files))
	for k := range s.Files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		p := e.path(k)
		if s.Files[k] == nil {
			if err := os.RemoveAll(p); err != nil {
				e.t.Fatal(err)
			}
			continue
		}
		content := e.sub(*s.Files[k])
		mode := os.FileMode(0o644)
		if strings.HasPrefix(content, "#!") {
			mode = 0o755
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			e.t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), mode); err != nil {
			e.t.Fatal(err)
		}
	}
	for _, g := range s.Git {
		if err := os.MkdirAll(e.vars["TEAM"], 0o755); err != nil {
			e.t.Fatal(err)
		}
		e.git(g...)
	}
}

func (e *env) git(args ...string) string {
	for i := range args {
		args[i] = e.sub(args[i])
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = e.vars["TEAM"]
	cmd.Env = e.environ(nil)
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = ee.Stderr
		}
		e.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr)
	}
	return strings.TrimSpace(string(out))
}

func (e *env) run(s Step) (string, string, int) {
	args := make([]string, len(s.Args))
	for i, a := range s.Args {
		args[i] = shellQuote(e.sub(a))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", e.vars["CLI"]+" "+strings.Join(args, " "))
	cmd.Dir = e.vars["ROOT"]
	if s.Cwd != "" {
		cmd.Dir = e.path(s.Cwd)
	} else if _, err := os.Stat(e.vars["TEAM"]); err == nil {
		cmd.Dir = e.vars["TEAM"]
	}
	cmd.Env = e.environ(s.Env)
	cmd.WaitDelay = 5 * time.Second // the timeout kills sh, not a child holding stdout open
	cmd.Stdin = strings.NewReader(e.sub(s.Stdin))
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		e.t.Fatalf("run %v: %v", s.Args, err)
	}
	return stdout.String(), stderr.String(), code
}

func (e *env) checkFiles(i int, files map[string]any) {
	for _, k := range slices.Sorted(maps.Keys(files)) {
		want := files[k]
		p := e.path(k)
		raw, err := os.ReadFile(p)
		_, serr := os.Lstat(p)
		exists := serr == nil
		switch w := want.(type) {
		case nil:
			if exists {
				e.t.Errorf("step %d: %s exists, want absent", i+1, k)
			}
		case bool:
			if exists != w {
				e.t.Errorf("step %d: %s exists=%v, want %v", i+1, k, exists, w)
			}
		case string:
			if err != nil {
				e.t.Errorf("step %d: %s: %v", i+1, k, err)
			} else if !e.match(w, string(raw)) {
				e.t.Errorf("step %d: %s:\n%s\nwant:\n%s", i+1, k, raw, e.sub(w))
			}
		default:
			e.t.Fatalf("step %d: files.%s: want a string, null or true", i+1, k)
		}
	}
}

// snapshot is every file under ${TEAM} except .git/, with its content and mode. Nil if no ${TEAM}.
func (e *env) snapshot() map[string]string {
	team := e.vars["TEAM"]
	if _, err := os.Stat(team); err != nil {
		return nil
	}
	out := map[string]string{}
	err := filepath.WalkDir(team, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(team, p)
		rel = filepath.ToSlash(rel)
		if rel == ".git" && d.IsDir() {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		raw, _ := os.ReadFile(p)
		out[rel] = fmt.Sprintf("%v\n%s", info.Mode(), raw)
		return nil
	})
	if err != nil {
		e.t.Fatal(err)
	}
	return out
}

// checkTree: every file matches a pattern and every pattern matches a file. A pattern is a path.Match
// glob; one ending in /** matches everything under that folder.
func (e *env) checkTree(i int, patterns []string) {
	files := e.snapshot()
	used := make([]bool, len(patterns))
	var extra []string
	for f := range files {
		hit := false
		for j, p := range patterns {
			if treeMatch(e.sub(p), f) {
				used[j], hit = true, true
			}
		}
		if !hit {
			extra = append(extra, f)
		}
	}
	sort.Strings(extra)
	if len(extra) > 0 {
		e.t.Errorf("step %d: tree has files no pattern lists:\n  %s", i+1, strings.Join(extra, "\n  "))
	}
	for j, p := range patterns {
		if !used[j] {
			e.t.Errorf("step %d: tree pattern %q matches no file", i+1, p)
		}
	}
}

func treeMatch(pattern, file string) bool {
	if dir, ok := strings.CutSuffix(pattern, "/**"); ok {
		return strings.HasPrefix(file, dir+"/")
	}
	ok, err := path.Match(pattern, file)
	return err == nil && ok
}

func (e *env) checkUnchanged(i int, before map[string]string) {
	after := e.snapshot()
	var diff []string
	for f, v := range before {
		if a, ok := after[f]; !ok {
			diff = append(diff, "removed  "+f)
		} else if a != v {
			diff = append(diff, "changed  "+f)
		}
	}
	for f := range after {
		if _, ok := before[f]; !ok {
			diff = append(diff, "added    "+f)
		}
	}
	sort.Strings(diff)
	if len(diff) > 0 {
		e.t.Errorf("step %d: tree changed:\n  %s", i+1, strings.Join(diff, "\n  "))
	}
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
