package gitguard

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// hasReview reports whether the team repo holds a review of head: a file under efforts/** or
// people/<handle>/work/** with frontmatter `sha: <head>`, or an inbox file people/<handle>/inbox/*-review-<sha7>.md
// where <sha7> is the first 7 characters of head. Only the person's own folder counts; with no handle
// only efforts/** does.
func hasReview(dir, handle, head string) bool {
	return hasFrontmatterReview(dir, handle, head) || hasInboxReview(dir, handle, head)
}

func hasFrontmatterReview(dir, handle, head string) bool {
	roots := []string{filepath.Join(dir, "efforts")}
	if handle != "" {
		roots = append(roots, filepath.Join(dir, "people", handle, "work"))
	}
	found := false
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && frontmatterSHA(p) == head {
				found = true
				return fs.SkipAll
			}
			return nil
		})
	}
	return found
}

// hasInboxReview matches by file name only: the date part is not validated.
func hasInboxReview(dir, handle, head string) bool {
	if handle == "" || len(head) < 7 {
		return false
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "people", handle, "inbox", "*-review-"+head[:7]+".md"))
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

// frontmatterSHA returns the `sha` of a file's leading `---` YAML block ("" if none).
func frontmatterSHA(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return ""
	}
	block, _, ok := strings.Cut(rest, "\n---")
	if !ok {
		return ""
	}
	var fm struct {
		SHA string `yaml:"sha"`
	}
	_ = yaml.Unmarshal([]byte(block), &fm)
	return fm.SHA
}
