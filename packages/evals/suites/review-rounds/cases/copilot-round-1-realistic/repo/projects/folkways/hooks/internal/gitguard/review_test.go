package gitguard

import (
	"os"
	"path/filepath"
	"testing"
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
		{"inbox review, any date text", map[string]string{"people/alex/inbox/whenever-review-abc1234.md": "x"}, true},
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
		if got := hasReview(dir, "alex", head); got != c.want {
			t.Errorf("%s: hasReview = %v, want %v", c.name, got, c.want)
		}
	}
	if hasReview(t.TempDir(), "alex", "abc") {
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
