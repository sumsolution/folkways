package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var hash = strings.Repeat("a", 64)

func writeRaw(t *testing.T, team, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(team, ".folkways"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(team, Path), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestIsOwnedPath(t *testing.T) {
	skills := []string{"start"}
	cases := map[string]bool{
		".folkways/instructions.md":          true,
		".folkways":                          false,
		".github/workflows/folkways-x.yml":   true,
		".github/workflows/other.yml":        false,
		".github/workflows/folkways-x/y.yml": false,
		".claude/skills/start/SKILL.md":      true,
		".claude/skills/start":               false,
		".claude/skills/other/SKILL.md":      false,
		"folkways.yaml":                      false,
		"../../outside":                      false,
		".folkways/../x":                     false,
		".folkways//x":                       false,
		`.folkways\x`:                        false,
	}
	for path, want := range cases {
		if got := IsOwnedPath(path, skills); got != want {
			t.Errorf("IsOwnedPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// From sync.test.ts: 'rejects an unsafe manifest before forced deletion' and 'rejects a manifest that is not JSON'.
func TestReadRejects(t *testing.T) {
	cases := map[string]string{
		`{ nope`:               "not a JSON object",
		`5`:                    "not a JSON object",
		`[]`:                   "missing version",
		`{"version": 1}`:       "missing version",
		`{"version": "0.1.0"}`: "missing skills",
		`{"version": "0.1.0", "skills": ["../../outside"], "files": {}}`:                                           `unsafe skill name "../../outside"`,
		`{"version": "0.1.0", "skills": ["a<b"], "files": {}}`:                                                     `unsafe skill name "a<b"`,
		`{"version": "0.1.0", "skills": [5], "files": {}}`:                                                         `unsafe skill name 5`,
		`{"version": "0.1.0", "skills": [], "files": []}`:                                                          "missing files",
		`{"version": "0.1.0", "skills": [], "files": {"../../outside": "` + hash + `"}}`:                           `unsafe path "../../outside"`,
		`{"version": "0.1.0", "skills": ["start"], "files": {".claude/skills/other/x.md": "` + hash + `"}}`:        `unsafe path ".claude/skills/other/x.md"`,
		`{"version": "0.1.0", "skills": [], "files": {"folkways.yaml": "` + hash + `"}}`:                           `unsafe path "folkways.yaml"`,
		`{"version": "0.1.0", "skills": [], "files": {".folkways/x": "nope"}}`:                                     "bad checksum for .folkways/x",
		`{"version": "0.1.0", "skills": [], "files": {}, "overlays": {"../x": "` + hash + `"}}`:                    `unsafe overlay path "../x"`,
		`{"version": "0.1.0", "skills": [], "files": {}, "overlays": {"overlays/a/b/OVERLAY.md": "` + hash + `"}}`: `unsafe overlay path "overlays/a/b/OVERLAY.md"`,
		`{"version": "0.1.0", "skills": [], "files": {}, "overlays": {"overlays/a/OVERLAY.md": "nope"}}`:           "bad checksum for overlays/a/OVERLAY.md",
		`{"version": "0.1.0", "skills": [], "files": {}, "overlays": []}`:                                          "bad overlays",
		`{"version": "0.1.0", "skills": [], "files": {}, "overlays": null}`:                                        "bad overlays",
	}
	for content, why := range cases {
		team := t.TempDir()
		writeRaw(t, team, content)
		_, err := Read(team)
		want := Path + ": " + why + ". Restore it with 'git checkout -- " + Path + "'."
		if err == nil || err.Error() != want {
			t.Errorf("Read(%s) = %v, want %q", content, err, want)
		}
	}
}

func TestReadMissing(t *testing.T) {
	if m, err := Read(t.TempDir()); m != nil || err != nil {
		t.Errorf("Read = %v, %v; want nil, nil", m, err)
	}
}

func TestWriteThenRead(t *testing.T) {
	team := t.TempDir()
	if err := os.MkdirAll(filepath.Join(team, ".folkways"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := &Manifest{Version: "1.0.0", Files: map[string]string{".folkways/b": hash, ".claude/skills/s/<x>.md": hash}, Skills: []string{"s"}}
	if err := Write(team, m); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(team, Path))
	want := "{\n  \"version\": \"1.0.0\",\n  \"skills\": [\n    \"s\"\n  ],\n  \"files\": {\n" +
		"    \".claude/skills/s/<x>.md\": \"" + hash + "\",\n    \".folkways/b\": \"" + hash + "\"\n  }\n}\n"
	if string(b) != want {
		t.Errorf("wrote\n%s\nwant\n%s", b, want)
	}
	got, err := Read(team)
	if err != nil || got.Version != "1.0.0" || len(got.Files) != 2 || got.Overlays != nil {
		t.Errorf("Read = %+v, %v", got, err)
	}
	if err := Write(team, &Manifest{Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(filepath.Join(team, Path))
	if want := "{\n  \"version\": \"1.0.0\",\n  \"skills\": [],\n  \"files\": {}\n}\n"; string(b) != want {
		t.Errorf("empty manifest:\n%s", b)
	}
}
