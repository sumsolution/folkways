package gitguard

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// hasReview reports whether the repo holds a review of head: a file under efforts/** or <personal>/work/**
// with frontmatter `sha: <head>`, or an inbox file <personal>/inbox/*-review-<sha7>.md where <sha7> is the
// first 7 characters of head. <personal> is the personal root (Config.PersonalRoot: people/<handle>/ in a team
// repo, the repo root in a solo repo); only the person's own folder counts. With no personal root only
// efforts/** does.
func hasReview(dir, personal, head string) bool {
	return slices.Contains(frontmatterSHAs(dir, personal), head) ||
		len(head) >= 7 && slices.Contains(inboxSHAs(personal), head[:7])
}

// reviewedSHAs are the commits the repo holds reviews of, as written: full SHAs from frontmatter, 7-character
// prefixes from inbox file names.
func reviewedSHAs(dir, personal string) []string {
	return append(frontmatterSHAs(dir, personal), inboxSHAs(personal)...)
}

func frontmatterSHAs(dir, personal string) []string {
	roots := []string{filepath.Join(dir, "efforts")}
	if personal != "" {
		roots = append(roots, filepath.Join(personal, "work"))
	}
	var shas []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				if sha := frontmatterSHA(p); sha != "" {
					shas = append(shas, sha)
				}
			}
			return nil
		})
	}
	return shas
}

// inboxReview matches inbox review file names `<YYYY-MM-DD>-review-<sha7>.md`.
var inboxReview = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-review-([0-9a-f]{7})\.md$`)

// inboxSHAs reads the person's inbox directory literally (the handle is never a glob).
func inboxSHAs(personal string) []string {
	if personal == "" {
		return nil
	}
	var shas []string
	entries, _ := os.ReadDir(filepath.Join(personal, "inbox"))
	for _, e := range entries {
		if m := inboxReview.FindStringSubmatch(e.Name()); m != nil && !e.IsDir() {
			shas = append(shas, m[1])
		}
	}
	return shas
}

// coversHead reports whether a review of reviewed covers head in repo: reviewed (a full SHA or a prefix) is an
// ancestor of head and every path changed between them, renames as delete plus add, matches one of skip.
func coversHead(repo, reviewed, head string, skip []string) bool {
	if len(skip) == 0 || strings.HasPrefix(reviewed, "-") {
		return false
	}
	base, err := gitOut(repo, "rev-parse", "--verify", "--quiet", reviewed+"^{commit}")
	if err != nil {
		return false
	}
	if _, err := gitOut(repo, "merge-base", "--is-ancestor", base, head); err != nil {
		return false
	}
	out, err := gitOut(repo, "diff", "--name-only", "--no-renames", "-z", base, head)
	if err != nil {
		return false
	}
	for _, p := range strings.Split(out, "\x00") {
		if p != "" && !slices.ContainsFunc(skip, func(g string) bool { return matchPath(g, p) }) {
			return false
		}
	}
	return true
}

// matchPath matches a slash-separated repo path against a glob: `**` is any number of whole segments, other
// segments follow path.Match; a trailing `**` needs at least one segment. A malformed pattern matches nothing.
func matchPath(pattern, p string) bool {
	return matchSegments(strings.Split(pattern, "/"), strings.Split(p, "/"))
}

func matchSegments(pat, segs []string) bool {
	if len(pat) == 0 {
		return len(segs) == 0
	}
	if pat[0] == "**" {
		if len(pat) == 1 {
			return len(segs) > 0 // a trailing ** needs something under it
		}
		for i := 0; i <= len(segs); i++ {
			if matchSegments(pat[1:], segs[i:]) {
				return true
			}
		}
		return false
	}
	if len(segs) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], segs[0])
	return ok && err == nil && matchSegments(pat[1:], segs[1:])
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
