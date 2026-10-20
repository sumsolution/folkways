// Package overlays compiles a team overlay (overlays/<skill>/OVERLAY.md, ADR 0001) into a framework skill.
// sync calls CompileSkillFiles. A line-for-line port of packages/cli/src/overlays.ts.
//
// An OVERLAY.md holds `## Add` (extra rules) and `## Override {#id}` (team only: replaces the skill section
// with that ID).
package overlays

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
)

// Compiled is the result of applying an overlay to a skill.
type Compiled struct {
	Files    map[string]string // only the files that change, by path relative to the skill folder
	Warnings []string
}

type override struct{ id, body string }

// parsed is an OVERLAY.md split into its sections.
type parsed struct {
	add       string     // body of `## Add`, trimmed; empty when absent
	overrides []override // section ID and replacement body, trimmed, in file order
	other     []string   // titles of level-2 headings that are neither Add nor Override
}

var (
	headingRe  = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
	idRe       = regexp.MustCompile(`\{#([a-z0-9-]+)\}$`)
	overrideRe = regexp.MustCompile(`^Override\s+\{#([a-z0-9-]+)\}$`)
	fenceRe    = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
)

type heading struct {
	line, level int
	title       string
}

// scan returns the Markdown headings outside fenced code blocks.
func scan(lines []string) []heading {
	var headings []heading
	open := ""
	for i, line := range lines {
		fence := ""
		if m := fenceRe.FindStringSubmatch(line); m != nil {
			fence = m[1]
		}
		if open != "" {
			// A fence closes on the same character, at least as long, with nothing after it.
			if fence != "" && fence[0] == open[0] && len(fence) >= len(open) && strings.TrimSpace(line) == fence {
				open = ""
			}
			continue
		}
		if fence != "" {
			open = fence
			continue
		}
		if m := headingRe.FindStringSubmatch(line); m != nil {
			headings = append(headings, heading{line: i, level: len(m[1]), title: m[2]})
		}
	}
	return headings
}

// sectionEnd is the index of the line that ends the section starting at h (next heading at the same or a
// higher level).
func sectionEnd(all []heading, h heading, total int) int {
	for _, x := range all {
		if x.line > h.line && x.level <= h.level {
			return x.line
		}
	}
	return total
}

func headingID(title string) string {
	if m := idRe.FindStringSubmatch(title); m != nil {
		return m[1]
	}
	return ""
}

type section struct{ title, body string }

func level2Sections(text string) []section {
	lines := strings.Split(text, "\n")
	all := scan(lines)
	var out []section
	for _, h := range all {
		if h.level != 2 {
			continue
		}
		end := sectionEnd(all, h, len(lines))
		out = append(out, section{title: h.title, body: strings.TrimSpace(strings.Join(lines[h.line+1:end], "\n"))})
	}
	return out
}

func parseOverlay(text string) (parsed, error) {
	var out parsed
	for _, s := range level2Sections(text) {
		m := overrideRe.FindStringSubmatch(s.title)
		switch {
		case s.title == "Add":
			if out.add == "" {
				out.add = s.body
			} else if s.body != "" {
				out.add += "\n\n" + s.body
			}
		case m != nil:
			id := m[1]
			if slices.ContainsFunc(out.overrides, func(o override) bool { return o.id == id }) {
				return parsed{}, clierr.New("overlay overrides #%s twice", id)
			}
			out.overrides = append(out.overrides, override{id, s.body})
		default:
			out.other = append(out.other, s.title)
		}
	}
	return out, nil
}

func replaceSection(skill, id, body string) (string, error) {
	lines := strings.Split(skill, "\n")
	all := scan(lines)
	var matches []heading
	for _, x := range all {
		if headingID(x.title) == id {
			matches = append(matches, x)
		}
	}
	if len(matches) > 1 {
		return "", clierr.New("the skill has #%s in more than one heading", id)
	}
	if len(matches) == 0 {
		return "", clierr.New("team overlay overrides #%s, but the skill has no section with that ID", id)
	}
	h := matches[0]
	end := sectionEnd(all, h, len(lines))
	out := append(slices.Clone(lines[:h.line+1]), "", "<!-- team override -->", body, "")
	return strings.Join(append(out, lines[end:]...), "\n"), nil
}

// countID counts the headings (outside fences) that carry section id.
func countID(text, id string) int {
	n := 0
	for _, h := range scan(strings.Split(text, "\n")) {
		if headingID(h.title) == id {
			n++
		}
	}
	return n
}

// checkNesting fails if one overridden section sits inside another: the outer replacement would remove the
// inner one.
func checkNesting(files map[string]string, markdown []string, overrides []override) error {
	type idSection struct {
		h  heading
		id string
	}
	for _, f := range markdown {
		lines := strings.Split(files[f], "\n")
		all := scan(lines)
		var sections []idSection
		for _, h := range all {
			if id := headingID(h.title); id != "" && slices.ContainsFunc(overrides, func(o override) bool { return o.id == id }) {
				sections = append(sections, idSection{h, id})
			}
		}
		for _, outer := range sections {
			end := sectionEnd(all, outer.h, len(lines))
			for _, inner := range sections {
				if inner.h.line > outer.h.line && inner.h.line < end {
					return clierr.New("team overlay overrides both #%s and #%s, but #%s sits inside #%s; override one",
						outer.id, inner.id, inner.id, outer.id)
				}
			}
		}
	}
	return nil
}

func unappliedWarnings(o parsed) []string {
	var out []string
	for _, title := range o.other {
		out = append(out, "overlays: ignored ## "+title+"; team overlays hold ## Add and ## Override {#id} only")
	}
	return out
}

// CompileSkillFiles applies overlay (the OVERLAY.md text) to a skill's files, keyed by path relative to the
// skill folder. An override replaces the section with that ID in whichever markdown file has it (SKILL.md or
// a reference); team additions go at the end of SKILL.md. An ID no file has, or that more than one heading
// has, fails.
func CompileSkillFiles(files map[string]string, overlay string) (Compiled, error) {
	o, err := parseOverlay(overlay)
	if err != nil {
		return Compiled{}, err
	}
	out := map[string]string{}
	current := func(path string) string {
		if s, ok := out[path]; ok {
			return s
		}
		return files[path]
	}
	markdown := slices.DeleteFunc(slices.Sorted(maps.Keys(files)), func(f string) bool { return !strings.HasSuffix(f, ".md") })
	if err := checkNesting(files, markdown, o.overrides); err != nil {
		return Compiled{}, err
	}
	for _, ov := range o.overrides {
		var holders []string
		for _, f := range markdown {
			if countID(current(f), ov.id) > 0 {
				holders = append(holders, f)
			}
		}
		if len(holders) == 0 {
			return Compiled{}, clierr.New("team overlay overrides #%s, but the skill has no section with that ID", ov.id)
		}
		if len(holders) > 1 {
			return Compiled{}, clierr.New("the skill has #%s in more than one file: %s", ov.id, strings.Join(holders, ", "))
		}
		s, err := replaceSection(current(holders[0]), ov.id, ov.body)
		if err != nil {
			return Compiled{}, err
		}
		out[holders[0]] = s
	}
	if o.add != "" {
		if _, ok := files["SKILL.md"]; !ok {
			return Compiled{}, clierr.New("team overlay adds rules, but the skill has no SKILL.md")
		}
		out["SKILL.md"] = strings.TrimRightFunc(current("SKILL.md"), unicode.IsSpace) + "\n\n## Team additions\n\n" + o.add + "\n"
	}
	return Compiled{Files: out, Warnings: unappliedWarnings(o)}, nil
}
