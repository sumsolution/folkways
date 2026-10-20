// Package hooks checks that an npm team repo can run the hook binary its settings.json calls.
package hooks

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/sumsolution/folkways/packages/folkways/internal/pkgmgr"
)

// Bin is the Go hook binary, linked by the @folkways/hooks-<os>-<cpu> optional dependency (ADR 0002).
const Bin = "node_modules/.bin/folkways-hook"

// Names are the hooks the binary runs; packages/content/settings.json wires each one.
var Names = []string{"context-band", "git-guard", "precompact", "session-start"}

// HoistPattern makes pnpm link our dependency's bin, which it links only when hoisted.
const HoistPattern = "@folkways/hooks-*"

// PnpmWorkspace is pnpm-workspace.yaml for a new pnpm team repo.
const PnpmWorkspace = "publicHoistPattern:\n  - \"" + HoistPattern + "\"\n"

var (
	shamefullyHoist = regexp.MustCompile(`(?m)^shamefullyHoist:\s*true\s*(#.*)?$`)
	hoistKey        = regexp.MustCompile(`^publicHoistPattern:`)
	hoistInline     = regexp.MustCompile(`^publicHoistPattern:\s*\[(.*)\]`)
	listItem        = regexp.MustCompile(`^\s+-\s*(.*)$`)
	comment         = regexp.MustCompile(`(^|\s)#.*$`)
)

// scalar is a YAML scalar without its quotes or trailing comment: `"a#b" # c` is a#b, `a # c` is a.
func scalar(s string) string {
	t := strings.TrimLeftFunc(s, unicode.IsSpace)
	if t != "" && (t[0] == '"' || t[0] == '\'') {
		if end := strings.IndexByte(t[1:], t[0]); end >= 0 {
			return t[1 : 1+end]
		}
	}
	return strings.TrimSpace(comment.ReplaceAllString(s, ""))
}

// HoistsHooks reports whether a pnpm-workspace.yaml hoists the hook binary: shamefullyHoist: true, or a
// publicHoistPattern list (block or inline) with our pattern or *.
func HoistsHooks(yaml string) bool {
	if shamefullyHoist.MatchString(yaml) {
		return true
	}
	lines := strings.Split(yaml, "\n")
	start := slices.IndexFunc(lines, hoistKey.MatchString)
	if start < 0 {
		return false
	}
	var items []string
	if m := hoistInline.FindStringSubmatch(lines[start]); m != nil {
		items = strings.Split(m[1], ",")
	} else {
		for _, line := range lines[start+1:] {
			if t := strings.TrimSpace(line); t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			m := listItem.FindStringSubmatch(line)
			if m == nil {
				break
			}
			items = append(items, m[1])
		}
	}
	return slices.ContainsFunc(items, func(i string) bool {
		p := scalar(i)
		return p == HoistPattern || p == "*"
	})
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Warnings lists what would stop hooks from running in this team repo; sync prints them as warnings.
func Warnings(team string) []string {
	var warnings []string
	if pm, err := pkgmgr.Detect(team); err == nil && pm == "pnpm" {
		b, err := os.ReadFile(filepath.Join(team, "pnpm-workspace.yaml"))
		if err != nil || !HoistsHooks(string(b)) {
			warnings = append(warnings, `pnpm-workspace.yaml needs publicHoistPattern: ["`+HoistPattern+`"], or pnpm won't link `+Bin+". Add it and run pnpm install.")
		}
	}
	if exists(filepath.Join(team, "node_modules")) && !exists(filepath.Join(team, Bin)) {
		warnings = append(warnings, Bin+" is missing, so hooks will fail. Reinstall without --omit=optional, and check that your registry serves "+HoistPattern+".")
	}
	return warnings
}
