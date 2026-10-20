// Package manifest reads and writes .folkways/manifest.json, the record of what the last sync wrote.
package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
)

// Path is the manifest's path in a team repo.
const Path = ".folkways/manifest.json"

// Manifest is what the last sync wrote: version, framework skills, a checksum per file, and a checksum per
// team overlay it compiled.
type Manifest struct {
	Version string            `json:"version"`
	Skills  []string          `json:"skills"`
	Files   map[string]string `json:"files"`
	// Overlays maps overlays/<skill>/OVERLAY.md to its checksum at the last sync. Empty when none was compiled.
	Overlays map[string]string `json:"overlays,omitempty"`
}

var (
	skillName   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	sha256Hex   = regexp.MustCompile(`^[a-f0-9]{64}$`)
	overlayPath = regexp.MustCompile(`^overlays/[A-Za-z0-9][A-Za-z0-9._-]*/OVERLAY\.md$`)
	workflow    = regexp.MustCompile(`^\.github/workflows/folkways-[^/]+\.yml$`)
)

// IsOwnedPath reports whether path is a plain relative path in a framework-owned location. sync deletes
// these, so a tampered manifest must not reach outside them.
func IsOwnedPath(path string, skills []string) bool {
	parts := strings.Split(path, "/")
	if strings.Contains(path, `\`) || slices.ContainsFunc(parts, func(p string) bool { return p == "" || p == "." || p == ".." }) {
		return false
	}
	if parts[0] == ".folkways" {
		return len(parts) > 1
	}
	if workflow.MatchString(path) {
		return true
	}
	return parts[0] == ".claude" && len(parts) > 3 && parts[1] == "skills" && slices.Contains(skills, parts[2])
}

func invalid(why string) error {
	return clierr.New("%s: %s. Restore it with 'git checkout -- %s'.", Path, why, Path)
}

func quote(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// validate checks a decoded manifest the way the TS CLI does, in its order, and converts it.
func validate(value any) (*Manifest, error) {
	obj, ok := value.(map[string]any)
	if _, isArray := value.([]any); isArray {
		obj, ok = map[string]any{}, true // an array is an object in JS: it fails on the first field
	}
	if !ok {
		return nil, invalid("not a JSON object")
	}
	m := &Manifest{}
	if m.Version, ok = obj["version"].(string); !ok {
		return nil, invalid("missing version")
	}
	skills, ok := obj["skills"].([]any)
	if !ok {
		return nil, invalid("missing skills")
	}
	m.Skills = []string{}
	for _, s := range skills {
		name, ok := s.(string)
		if !ok || !skillName.MatchString(name) {
			return nil, invalid("unsafe skill name " + quote(s))
		}
		m.Skills = append(m.Skills, name)
	}
	files, ok := obj["files"].(map[string]any)
	if !ok {
		return nil, invalid("missing files")
	}
	m.Files = map[string]string{}
	for _, path := range sortedKeys(files) {
		if !IsOwnedPath(path, m.Skills) {
			return nil, invalid("unsafe path " + quote(path))
		}
		hash, ok := files[path].(string)
		if !ok || !sha256Hex.MatchString(hash) {
			return nil, invalid("bad checksum for " + path)
		}
		m.Files[path] = hash
	}
	if raw, present := obj["overlays"]; present {
		overlays, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid("bad overlays")
		}
		m.Overlays = map[string]string{}
		for _, path := range sortedKeys(overlays) {
			if !overlayPath.MatchString(path) {
				return nil, invalid("unsafe overlay path " + quote(path))
			}
			hash, ok := overlays[path].(string)
			if !ok || !sha256Hex.MatchString(hash) {
				return nil, invalid("bad checksum for " + path)
			}
			m.Overlays[path] = hash
		}
	}
	return m, nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Read returns the team repo's manifest, or nil when it has none. A broken or unsafe manifest is an error.
func Read(team string) (*Manifest, error) {
	b, err := os.ReadFile(filepath.Join(team, Path))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var value any
	if json.Unmarshal(b, &value) != nil {
		value = nil
	}
	return validate(value)
}

// Write writes m to dir/.folkways/manifest.json (dir is a team repo or a staging copy of one). Keys are
// sorted; the TS CLI kept insertion order.
func Write(dir string, m *Manifest) error {
	out := *m
	if out.Skills == nil {
		out.Skills = []string{}
	}
	if out.Files == nil {
		out.Files = map[string]string{}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, Path), buf.Bytes(), 0o644)
}
