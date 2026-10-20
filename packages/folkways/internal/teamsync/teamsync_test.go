package teamsync

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
	"github.com/sumsolution/folkways/packages/folkways/internal/files"
	"github.com/sumsolution/folkways/packages/folkways/internal/manifest"
)

// The unit cases the sync fixtures can't express (fixtures/sync/COVERAGE.md): failure injection, the
// executable bit, and StaleOverlays. They use a synthetic bundle, as the TS tests' makePackage does.

const fragment = `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"\"${CLAUDE_PROJECT_DIR:-.}\"/node_modules/.bin/folkways-hook session-start"}]}]},"permissions":{"ask":["Bash(git push:*)"]}}`

func file(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s), Mode: 0o644} }

// pkg is a bundle like makePackage plus addFrameworkContent: one skill with a reference, a prompt, a workflow.
func pkg(version string) bundle.Bundle {
	return bundle.Bundle{Version: version, FS: fstest.MapFS{
		"instructions.md":               file("# Way of working\n"),
		"settings.json":                 file(fragment),
		"skills/start/SKILL.md":         file("---\nname: start\ndescription: test\n---\nbody\n## Steps {#steps}\n\nold steps\n"),
		"skills/start/references/a.md":  file("ref\n"),
		"prompts/review.md":             file("review prompt\n"),
		"workflows/knowledge-audit.yml": file("name: audit\n"),
	}}
}

// next03 is the next version: it drops the skill's reference file, changes a prompt and adds one.
func next03() bundle.Bundle {
	b := pkg("0.3.0")
	m := b.FS.(fstest.MapFS)
	delete(m, "skills/start/references/a.md")
	m["prompts/review.md"] = file("review prompt v3\n")
	m["prompts/new.md"] = file("new prompt\n")
	return b
}

// team is a team repo synced at 0.2.0.
func team(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "AGENTS.md", "@.folkways/instructions.md\n")
	write(t, dir, "folkways.yaml", "team: Example Team\n")
	if _, err := Sync(dir, pkg("0.2.0"), Options{}); err != nil {
		t.Fatal(err)
	}
	return dir
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// snapshot is every file in the team repo with its content and mode.
func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		out[rel] = info.Mode().String() + " " + read(t, dir, rel)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// failOn makes the nth call to Copy or Rename fail (and every later one, when always is set).
func failOn(t *testing.T, op string, n int, always bool) {
	t.Helper()
	real := ops
	t.Cleanup(func() { ops = real })
	calls := 0
	fail := func() error {
		calls++
		if calls == n || (always && calls > n) {
			return errors.New("injected " + op + " failure")
		}
		return nil
	}
	ops = files.Ops{Copy: real.Copy, Rename: real.Rename}
	if op == "copy" {
		ops.Copy = func(fsys fs.FS, src, dest string) error {
			if err := fail(); err != nil {
				return err
			}
			return real.Copy(fsys, src, dest)
		}
	} else {
		ops.Rename = func(from, to string) error {
			if err := fail(); err != nil {
				return err
			}
			return real.Rename(from, to)
		}
	}
}

func noStage(t *testing.T, dir string) {
	t.Helper()
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".folkways-") {
			t.Errorf("left %s behind", e.Name())
		}
	}
}

func TestSyncIsAtomic(t *testing.T) {
	cases := []struct {
		op string
		n  int
	}{{"copy", 1}, {"copy", 3}, {"copy", 5}, {"rename", 1}, {"rename", 4}, {"rename", 8}, {"rename", 12}}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %d fails", c.op, c.n), func(t *testing.T) {
			dir := team(t)
			before := snapshot(t, dir)
			failOn(t, c.op, c.n, false)
			if _, err := Sync(dir, next03(), Options{}); err == nil || !strings.Contains(err.Error(), "injected "+c.op+" failure") {
				t.Fatalf("err = %v, want the injected failure", err)
			}
			if after := snapshot(t, dir); !maps.Equal(after, before) {
				t.Errorf("team repo changed:\nbefore %v\nafter  %v", before, after)
			}
			noStage(t, dir)

			ops = files.OS
			r, err := Sync(dir, next03(), Options{})
			if err != nil {
				t.Fatal(err)
			}
			if r.Version != "0.3.0" || r.Previous != "0.2.0" || len(r.Discarded) != 0 {
				t.Errorf("result = %+v", r)
			}
			if got := read(t, dir, ".folkways/prompts/review.md"); got != "review prompt v3\n" {
				t.Errorf("review.md = %q", got)
			}
			if _, err := os.Stat(filepath.Join(dir, ".claude/skills/start/references")); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("references/ is still there: %v", err)
			}
		})
	}
}

func TestSyncAdoptingKeepsFilesWhenSwapFails(t *testing.T) {
	dir := team(t)
	if err := os.Remove(filepath.Join(dir, manifest.Path)); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, dir)
	failOn(t, "rename", 5, false)
	if _, err := Sync(dir, pkg("0.2.0"), Options{}); err == nil || !strings.Contains(err.Error(), "injected") {
		t.Fatalf("err = %v, want the injected failure", err)
	}
	if after := snapshot(t, dir); !maps.Equal(after, before) {
		t.Errorf("team repo changed:\nbefore %v\nafter  %v", before, after)
	}
}

func TestSyncKeepsStageWhenRestoreFails(t *testing.T) {
	dir := team(t)
	failOn(t, "rename", 4, true)
	_, err := Sync(dir, next03(), Options{})
	if err == nil || !strings.Contains(err.Error(), "sync failed and could not restore the old files; they are in ") {
		t.Fatalf("err = %v", err)
	}
	if !errors.As(err, new(*clierr.Error)) {
		t.Error("a restore failure is not a *clierr.Error")
	}
	entries, _ := os.ReadDir(dir)
	kept := false
	for _, e := range entries {
		kept = kept || strings.HasPrefix(e.Name(), stagePrefix)
	}
	if !kept {
		t.Error("the staging folder with the old files was removed")
	}
}

func TestSyncPreservesExecutableBit(t *testing.T) {
	dir := team(t)
	b := pkg("0.3.0")
	b.FS.(fstest.MapFS)["skills/start/scripts/run.sh"] = &fstest.MapFile{Data: []byte("#!/bin/sh\n"), Mode: 0o755}
	if _, err := Sync(dir, b, Options{}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, ".claude/skills/start/scripts/run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("mode = %v, want executable", info.Mode())
	}
}

func TestStaleOverlays(t *testing.T) {
	dir := team(t)
	write(t, dir, "overlays/start/OVERLAY.md", "## Override {#steps}\n\nteam steps\n")
	write(t, dir, "overlays/other/OVERLAY.md", "## Add\n\nx\n") // not a framework skill: never stale
	m, _ := manifest.Read(dir)
	stale, err := StaleOverlays(dir, m)
	if err != nil || len(stale) != 1 || stale[0] != "overlays/start/OVERLAY.md" {
		t.Fatalf("before sync: %v, %v", stale, err)
	}
	if _, err := Sync(dir, pkg("0.2.0"), Options{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, dir, ".claude/skills/start/SKILL.md"), "<!-- team override -->\nteam steps\n") {
		t.Error("overlay not compiled")
	}
	m, _ = manifest.Read(dir)
	if stale, err := StaleOverlays(dir, m); err != nil || len(stale) != 0 {
		t.Fatalf("after sync: %v, %v", stale, err)
	}
	if err := os.Remove(filepath.Join(dir, "overlays/start/OVERLAY.md")); err != nil {
		t.Fatal(err)
	}
	if stale, err := StaleOverlays(dir, m); err != nil || len(stale) != 1 || stale[0] != "overlays/start/OVERLAY.md" {
		t.Fatalf("after removal: %v, %v", stale, err)
	}
}
