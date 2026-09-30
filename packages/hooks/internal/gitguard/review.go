package gitguard

import (
	"bufio"
	"io"
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
		if found {
			return true
		}
	}
	return false
}

// hasInboxReview reads the person's inbox directory literally (the handle is never a glob) and matches by
// file name only: the date part is not validated.
func hasInboxReview(dir, handle, head string) bool {
	if handle == "" || len(head) < 7 {
		return false
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "people", handle, "inbox"))
	suffix := "-review-" + head[:7] + ".md"
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			return true
		}
	}
	return false
}

// Frontmatter limits: a review file's frontmatter is small; never read more than this of any file.
const (
	maxFrontmatterBytes = 64 << 10
	maxFrontmatterLines = 500
)

// frontmatterSHA returns the `sha` of a file's leading `---` YAML block ("" if none). It reads only the
// start of the file, and gives up when the first line is not `---` or the block does not close within the limits.
func frontmatterSHA(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(io.LimitReader(f, maxFrontmatterBytes))
	first, err := r.ReadString('\n')
	if err != nil || strings.TrimRight(first, "\r\n") != "---" {
		return ""
	}
	var block strings.Builder
	for n := 0; n < maxFrontmatterLines; n++ {
		line, err := r.ReadString('\n')
		if strings.TrimRight(line, "\r\n") == "---" {
			var fm struct {
				SHA string `yaml:"sha"`
			}
			_ = yaml.Unmarshal([]byte(block.String()), &fm)
			return fm.SHA
		}
		block.WriteString(strings.TrimRight(line, "\r\n") + "\n")
		if err != nil {
			return ""
		}
	}
	return ""
}
