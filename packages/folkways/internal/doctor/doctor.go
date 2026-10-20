// Package doctor checks a team repo (`folkways doctor`). Ported from packages/cli/src/doctor.ts.
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/hooks"
	"github.com/sumsolution/folkways/packages/folkways/internal/manifest"
	"github.com/sumsolution/folkways/packages/folkways/internal/pkgmgr"
	"github.com/sumsolution/folkways/packages/folkways/internal/teamsync"
)

// Finding is one problem, with its fix in the message. Level is "error" or "warning".
type Finding struct {
	Level   string
	Message string
}

// Options for Doctor.
type Options struct {
	Home string // where personal skills live (~/.claude/skills, ...)
}

func errorf(format string, a ...any) Finding   { return Finding{"error", fmt.Sprintf(format, a...)} }
func warningf(format string, a ...any) Finding { return Finding{"warning", fmt.Sprintf(format, a...)} }

// Doctor checks config, versions, overlays, skills, settings and the hook install. b is the running build.
// Each message names its fix.
func Doctor(team string, b bundle.Bundle, opts Options) ([]Finding, error) {
	config := filepath.Join(team, "folkways.yaml")
	if !exists(config) {
		return []Finding{errorf("no folkways.yaml in %s; run the setup skill, or run doctor in the team repo", team)}, nil
	}
	yaml, err := os.ReadFile(config)
	if err != nil {
		return nil, err
	}
	findings := hookValueFindings(string(yaml), hooks.Names)
	// No package.json: the plugin set this repo up and there is no CLI install, sync or settings to check.
	if !exists(filepath.Join(team, "package.json")) {
		return append(findings, shadowFindings(team, opts.Home)...), nil
	}
	m, err := manifest.Read(team)
	if err != nil {
		findings = append(findings, errorf("%s", err.Error()))
		m = nil
	}
	synced := ""
	if m != nil {
		synced = m.Version
	}
	findings = append(findings, versionFindings(team, b.Version, synced)...)
	if m != nil {
		edits, err := teamsync.FindEdits(team, m)
		if err != nil {
			return nil, err
		}
		for _, e := range edits {
			findings = append(findings, editFinding(e))
		}
		stale, err := teamsync.StaleOverlays(team, m)
		if err != nil {
			return nil, err
		}
		for _, p := range stale {
			findings = append(findings, errorf("%s changed since the last sync compiled it; run 'folkways sync'", p))
		}
	} else if !slices.ContainsFunc(findings, func(f Finding) bool { return strings.HasPrefix(f.Message, manifest.Path) }) {
		findings = append(findings, errorf("no %s; run 'folkways sync'", manifest.Path))
	}
	findings = append(findings, shadowFindings(team, opts.Home)...)
	findings = append(findings, settingsFindings(team)...)
	for _, w := range hooks.Warnings(team) {
		findings = append(findings, errorf("%s", w))
	}
	return findings, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// editFinding turns `modified  .folkways/instructions.md` (from FindEdits) into an error: sync refuses until it is resolved.
func editFinding(edit string) Finding {
	how, path, _ := strings.Cut(edit, " ")
	path = strings.TrimLeftFunc(path, isJSSpace)
	return errorf("%s was %s, so sync refuses; move the change into an overlay, or run 'folkways sync --force' to discard it", path, how)
}

// readJSON parses a file the way `JSON.parse(...)` then a truthiness test does: ok is false for unreadable,
// invalid, or falsy (null, false, 0, "") content. Other non-objects give an empty map, as property reads on
// them are undefined.
func readJSON(path string) (obj map[string]any, ok bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var v any
	if json.Unmarshal(b, &v) != nil {
		return nil, false
	}
	switch x := v.(type) {
	case nil:
		return nil, false
	case bool:
		return map[string]any{}, x
	case float64:
		return map[string]any{}, x != 0
	case string:
		return map[string]any{}, x != ""
	case map[string]any:
		return x, true
	}
	return map[string]any{}, true
}

// stringAt reads obj[a][b] as a string, or "".
func stringAt(obj map[string]any, a, b string) (string, bool) {
	inner, _ := obj[a].(map[string]any)
	s, ok := inner[b].(string)
	return s, ok
}

var exact = regexp.MustCompile(`^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)

// versionFindings checks the pinned, installed and running CLI versions, and the version the repo was last synced from.
func versionFindings(team, running, synced string) []Finding {
	install := "npm install"
	if pm, err := pkgmgr.Detect(team); err == nil {
		install = strings.Join(pkgmgr.InstallCommand(pm), " ")
	}
	pkg, _ := readJSON(filepath.Join(team, "package.json"))
	pinned, ok := stringAt(pkg, "devDependencies", bundle.PackageName)
	if !ok {
		pinned, ok = stringAt(pkg, "dependencies", bundle.PackageName)
	}
	if !ok || pinned == "" {
		return []Finding{errorf("package.json doesn't pin %s; add it as an exact devDependency ('folkways upgrade <version>' does this)", bundle.PackageName)}
	}
	if !exact.MatchString(pinned) {
		return []Finding{errorf("package.json pins %s %s, not an exact version; pin one with 'folkways upgrade <version>'", bundle.PackageName, pinned)}
	}

	var findings []Finding
	installedPkg, _ := readJSON(filepath.Join(team, "node_modules", bundle.PackageName, "package.json"))
	installed, _ := installedPkg["version"].(string)
	if installed == "" {
		findings = append(findings, errorf("%s is not installed; run '%s'", bundle.PackageName, install))
	} else if installed != pinned {
		findings = append(findings, errorf("node_modules has %s %s but package.json pins %s (a failed upgrade leaves this); run '%s'", bundle.PackageName, installed, pinned, install))
	}
	// A running CLI that is the installed one is covered by the node_modules error above.
	if running != pinned && running != installed {
		findings = append(findings, warningf("this is folkways %s, but the repo pins %s; run the pinned one with 'npx folkways doctor' from the team repo", running, pinned))
	}
	if synced != "" && synced != pinned {
		findings = append(findings, errorf("the repo was synced from %s, but package.json pins %s; run 'folkways sync'", synced, pinned))
	}
	return findings
}

var personalSkillDirs = []string{".claude/skills", ".agents/skills", ".copilot/skills"}

// dirs lists the folders and symlinks in path, sorted; nothing if it can't be read.
func dirs(path string) []string {
	entries, _ := os.ReadDir(path)
	var names []string
	for _, e := range entries {
		if e.Type()&(os.ModeDir|os.ModeSymlink) != 0 {
			names = append(names, e.Name())
		}
	}
	return names
}

// shadowFindings warns about personal skills with the name of a framework or team skill (ADR 0001 M10).
func shadowFindings(team, home string) []Finding {
	repoSkills := dirs(filepath.Join(team, ".claude/skills"))
	var findings []Finding
	for _, d := range personalSkillDirs {
		for _, name := range dirs(filepath.Join(home, d)) {
			if slices.Contains(repoSkills, name) {
				findings = append(findings, warningf("~/%s/%s shadows the skill %s in this repo; rename or remove it", d, name, name))
			}
		}
	}
	return findings
}

// settingsFindings checks `disableAllHooks` (ADR 0001 M11): the project sets false; a local true turns every hook off.
func settingsFindings(team string) []Finding {
	var findings []Finding
	for _, s := range []struct {
		file     string
		required bool
	}{{".claude/settings.json", true}, {".claude/settings.local.json", false}} {
		path := filepath.Join(team, s.file)
		if !exists(path) {
			if s.required {
				findings = append(findings, errorf("%s is missing; run 'folkways sync'", s.file))
			}
			continue
		}
		settings, ok := readJSON(path)
		v, isBool := settings["disableAllHooks"].(bool)
		switch {
		case !ok:
			findings = append(findings, errorf("%s is not valid JSON; fix it, or hooks and permissions may not load", s.file))
		case s.required && !(isBool && !v):
			findings = append(findings, errorf(`%s must set "disableAllHooks: false" so a user setting can't turn hooks off; run 'folkways sync'`, s.file))
		case !s.required && isBool && v:
			findings = append(findings, errorf("%s sets disableAllHooks: true, which turns every folkways hook off; remove it", s.file))
		}
	}
	return findings
}

// JavaScript's \s and String.trim() whitespace, which Go's regexp \s and unicode.IsSpace differ from. The regexps
// below spell the class out, as package-level MustCompile takes a string literal.
func isJSSpace(r rune) bool { return unicode.IsSpace(r) && r != 0x85 || r == 0xfeff }

func trim(s string) string { return strings.TrimFunc(s, isJSSpace) }

var (
	hooksStart = regexp.MustCompile(`^hooks[\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]*:`)
	// JS `.` does not match line terminators.
	keyLine = regexp.MustCompile(`^(?:"([^"]*)"|'([^']*)'|([^\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}"'#][^:#]*?))[\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]*:(?:[\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]+([^\n\r\x{2028}\x{2029}]*))?$`)
)

// How gopkg.in/yaml.v3 decodes a scalar into the hooks' `map[string]bool` (checked against yaml.v3 3.0.1).
var (
	yamlTrue   = []string{"y", "Y", "yes", "Yes", "YES", "on", "On", "ON"}
	yamlFalse  = []string{"n", "N", "no", "No", "NO", "off", "Off", "OFF"}
	plainTrue  = []string{"true", "True", "TRUE"}
	plainFalse = []string{"false", "False", "FALSE", "", "~", "null", "Null", "NULL"}
	nullMap    = []string{"", "{}", "~", "null", "Null", "NULL"}
)

func hasLineTerminator(s string) bool { return strings.ContainsAny(s, "\n\r  ") }

// unquote strips one pair of matching quotes: ^(["'])(.*)\1$ in the TS.
func unquote(s string) (inner string, quoted bool) {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] && !hasLineTerminator(s[1:len(s)-1]) {
		return s[1 : len(s)-1], true
	}
	return s, false
}

// hookValue is what the hook binary reads for a raw `hooks:` value: ok is false when it is not a boolean (the hook stays on).
func hookValue(raw string) (value, ok bool) {
	s, quoted := unquote(raw)
	switch {
	case !quoted && slices.Contains(plainTrue, s):
		return true, true
	case !quoted && slices.Contains(plainFalse, s):
		return false, true
	case slices.Contains(yamlTrue, s):
		return true, true
	case slices.Contains(yamlFalse, s):
		return false, true
	}
	return false, false
}

// stripComment is a raw YAML value without its trailing comment: a leading quoted string is kept whole,
// otherwise a `#` at the start or after whitespace starts the comment.
func stripComment(v string) string {
	if v != "" && (v[0] == '"' || v[0] == '\'') {
		if end := strings.IndexByte(v[1:], v[0]); end >= 0 && !hasLineTerminator(v[1:1+end]) {
			return trim(v[:end+2])
		}
	}
	for i := 0; i < len(v); i++ {
		if v[i] != '#' {
			continue
		}
		if r, _ := utf8.DecodeLastRuneInString(v[:i]); i == 0 || isJSSpace(r) {
			return trim(v[:i])
		}
	}
	return trim(v)
}

type hookEntry struct {
	name, raw string
	nested    bool
}

// matchKey splits a `key: value` line; ok is false when it is not one.
func matchKey(line string) (name, raw string, ok bool) {
	m := keyLine.FindStringSubmatchIndex(line)
	if m == nil {
		return "", "", false
	}
	for g := 1; g <= 3; g++ {
		if m[2*g] >= 0 {
			name = line[m[2*g]:m[2*g+1]]
			break
		}
	}
	if m[8] >= 0 {
		raw = line[m[8]:m[9]]
	}
	return trim(name), raw, true
}

// hookValueFindings checks the `hooks:` block of folkways.yaml: each value is `true` or `false`. The hook binary
// (yaml.v3) also reads YAML 1.1 words (`off`, `no`) and empty values as booleans, and anything else leaves the
// hook on; `run.sh` in the plugin tests only some of these, so doctor asks for the two plain words.
func hookValueFindings(yaml string, known []string) []Finding {
	lines := strings.Split(yaml, "\n")
	for i := 0; i < len(lines)-1; i++ { // split on \r?\n
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	start := slices.IndexFunc(lines, hooksStart.MatchString)
	if start < 0 {
		return nil
	}
	if inline := stripComment(hooksStart.ReplaceAllString(lines[start], "")); !slices.Contains(nullMap, inline) {
		return []Finding{warningf("folkways.yaml hooks: write one hook per line (git-guard: false) so doctor can check the values")}
	}

	var entries []*hookEntry
	indent := -1
loop:
	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		if t := trim(line); t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		if spaces < len(line) && line[spaces] == '\t' {
			return []Finding{errorf("folkways.yaml hooks: line %d is indented with a tab, which YAML forbids, so the hooks ignore folkways.yaml; indent with spaces", i+1)}
		}
		if spaces == 0 {
			break
		}
		if indent < 0 {
			indent = spaces
		}
		switch {
		case spaces > indent:
			// Deeper lines belong to the entry above: a nested map, or a plain value continued on the next line.
			if len(entries) == 0 {
				continue
			}
			last := entries[len(entries)-1]
			if _, _, isKey := matchKey(trim(line)); isKey {
				last.nested = true
			} else if !last.nested {
				last.raw = trim(last.raw + " " + stripComment(line))
			}
		case spaces < indent:
			break loop
		default:
			name, raw, ok := matchKey(trim(line))
			if !ok {
				break loop
			}
			entries = append(entries, &hookEntry{name: name, raw: stripComment(raw)})
		}
	}

	var findings []Finding
	for _, e := range entries {
		where := "folkways.yaml hooks." + e.name
		if !slices.Contains(known, e.name) {
			sorted := slices.Clone(known)
			sort.Strings(sorted)
			findings = append(findings, warningf("%s: no hook by that name (hooks: %s)", where, strings.Join(sorted, ", ")))
			continue
		}
		if e.nested && e.raw == "" {
			findings = append(findings, errorf("%s: a nested map is not a boolean, so the hook stays on; write true or false", where))
			continue
		}
		if e.raw == "true" || e.raw == "false" {
			continue
		}
		if e.raw != "" && strings.ContainsRune("!*&", rune(e.raw[0])) {
			findings = append(findings, errorf("%s: doctor can't check %s (a tag, anchor or alias); write true or false", where, e.raw))
			continue
		}
		shown := e.raw
		if shown == "" {
			shown = "(empty)"
		}
		if value, ok := hookValue(e.raw); !ok {
			findings = append(findings, errorf("%s: %s is not a boolean, so the hook stays on; write true or false", where, shown))
		} else if value {
			findings = append(findings, errorf("%s: %s reads as true (hook on); write true", where, shown))
		} else {
			findings = append(findings, errorf("%s: %s reads as false (hook off); write false", where, shown))
		}
	}
	return findings
}
