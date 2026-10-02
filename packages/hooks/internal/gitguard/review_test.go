package gitguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

const head = "abc1234def5678901234567890abcdef12345678"

func write(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHasReview(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  bool
	}{
		{"empty", nil, false},
		{"effort frontmatter", map[string]string{"efforts/e1/review.md": "---\nsha: " + head + "\n---\nok\n"}, true},
		{"work frontmatter", map[string]string{"people/alex/work/r.md": "---\nsha: " + head + "\n---\n"}, true},
		{"stale frontmatter", map[string]string{"efforts/e1/review.md": "---\nsha: 0000000000000000000000000000000000000000\n---\n"}, false},
		{"frontmatter outside roots", map[string]string{"notes/r.md": "---\nsha: " + head + "\n---\n"}, false},
		{"inbox review", map[string]string{"people/alex/inbox/2026-09-29-review-abc1234.md": "x"}, true},
		{"inbox name without a date", map[string]string{"people/alex/inbox/whenever-review-abc1234.md": "x"}, false},
		{"inbox name with a prefix", map[string]string{"people/alex/inbox/x-2026-09-29-review-abc1234.md": "x"}, false},
		{"inbox stale sha", map[string]string{"people/alex/inbox/2026-09-29-review-abc1235.md": "x"}, false},
		{"inbox wrong suffix", map[string]string{"people/alex/inbox/2026-09-29-review-abc1234.txt": "x"}, false},
		{"inbox not a review", map[string]string{"people/alex/inbox/2026-09-29-note-abc1234.md": "x"}, false},
		{"other person work review", map[string]string{"people/sam/work/r.md": "---\nsha: " + head + "\n---\n"}, false},
		{"other person inbox review", map[string]string{"people/sam/inbox/2026-09-29-review-abc1234.md": "x"}, false},
		{"inbox in wrong place", map[string]string{"efforts/e1/2026-09-29-review-abc1234.md": "x"}, false},
	}
	for _, c := range cases {
		dir := t.TempDir()
		for rel, body := range c.files {
			write(t, dir, rel, body)
		}
		if got := hasReview(dir, personal(dir, "alex"), head); got != c.want {
			t.Errorf("%s: hasReview = %v, want %v", c.name, got, c.want)
		}
	}
	if hasReview(t.TempDir(), personal("x", "alex"), "abc") {
		t.Error("short head must not match")
	}
}

func TestFrontmatterSHA(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.md", "---\r\nsha: xyz\r\n---\r\nbody")
	write(t, dir, "b.md", "no frontmatter\nsha: xyz\n")
	write(t, dir, "c.md", "---\nsha: xyz\n")
	if got := frontmatterSHA(filepath.Join(dir, "a.md")); got != "xyz" {
		t.Errorf("CRLF frontmatter = %q", got)
	}
	for _, f := range []string{"b.md", "c.md", "missing.md"} {
		if got := frontmatterSHA(filepath.Join(dir, f)); got != "" {
			t.Errorf("%s = %q, want empty", f, got)
		}
	}
}

func TestHasReviewNoHandle(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "people/alex/work/r.md", "---\nsha: "+head+"\n---\n")
	write(t, dir, "people/alex/inbox/2026-09-29-review-abc1234.md", "x")
	if hasReview(dir, "", head) {
		t.Error("with no handle, people/** must not count")
	}
	write(t, dir, "efforts/e1/r.md", "---\nsha: "+head+"\n---\n")
	if !hasReview(dir, "", head) {
		t.Error("with no handle, efforts/** must still count")
	}
}

func TestHasReviewHandleIsLiteral(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "people/alex/inbox/2026-09-29-review-abc1234.md", "x")
	write(t, dir, "people/alex/work/r.md", "---\nsha: "+head+"\n---\n")
	for _, h := range []string{"*", "a*", "[a-z]lex", "?lex"} {
		if hasReview(dir, personal(dir, h), head) {
			t.Errorf("handle %q must not act as a glob", h)
		}
	}
	write(t, dir, "people/*/inbox/2026-09-29-review-abc1234.md", "x")
	if !hasReview(dir, personal(dir, "*"), head) {
		t.Error("a literal directory named * is that person's own folder")
	}
}

func TestFrontmatterSHAReadsOnlyTheStart(t *testing.T) {
	dir := t.TempDir()
	big := strings.Repeat("filler: line\n", 20000)
	write(t, dir, "huge.md", "---\nsha: xyz\n---\n"+big)
	write(t, dir, "unterminated.md", "---\nsha: xyz\n"+big)
	write(t, dir, "notfm.md", big+"---\nsha: xyz\n---\n")
	if got := frontmatterSHA(filepath.Join(dir, "huge.md")); got != "xyz" {
		t.Errorf("huge body: %q", got)
	}
	for _, f := range []string{"unterminated.md", "notfm.md"} {
		if got := frontmatterSHA(filepath.Join(dir, f)); got != "" {
			t.Errorf("%s: %q, want empty", f, got)
		}
	}
}

// personal is the team-repo personal root of a handle.
func personal(dir, handle string) string { return filepath.Join(dir, "people", handle) }

func TestHasReviewSoloRoot(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  bool
	}{
		{"inbox review", map[string]string{"inbox/2026-09-29-review-abc1234.md": "x"}, true},
		{"work frontmatter", map[string]string{"work/EX-1-a/r.md": "---\nsha: " + head + "\n---\n"}, true},
		{"stale inbox", map[string]string{"inbox/2026-09-29-review-abc1235.md": "x"}, false},
		{"people/ is not the solo root", map[string]string{"people/alex/inbox/2026-09-29-review-abc1234.md": "x"}, false},
		{"nothing", nil, false},
	}
	for _, c := range cases {
		dir := t.TempDir()
		for rel, body := range c.files {
			write(t, dir, rel, body)
		}
		var cfg team.Config
		cfg.Dir, cfg.Team.Mode = dir, "solo" // no handle: solo needs none
		if got := hasReview(dir, cfg.PersonalRoot(), head); got != c.want {
			t.Errorf("%s: hasReview = %v, want %v", c.name, got, c.want)
		}
	}
}
