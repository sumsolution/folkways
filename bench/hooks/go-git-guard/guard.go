package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	header    = regexp.MustCompile(`^(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(([a-z0-9._-]+)\))?!?: \S`)
	shortMsg  = regexp.MustCompile(`^-[a-zA-Z]*m$`) // -m, -am
	envAssign = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
	mainGuard = map[string]bool{"checkout": true, "switch": true, "pull": true, "reset": true, "rebase": true}
)

// checkCommand walks every command of a chain and returns the first deny reason ("" to allow).
func checkCommand(cmd, payloadCwd string, cfg config) string {
	cur := payloadCwd
	if cur == "" {
		cur, _ = os.Getwd()
	}
	for _, seg := range splitCommands(cmd) {
		w := words(seg)
		for len(w) > 0 && envAssign.MatchString(w[0]) {
			w = w[1:]
		}
		switch {
		case len(w) >= 2 && w[0] == "cd":
			cur = resolve(cur, w[1])
		case len(w) >= 2 && w[0] == "git":
			if reason := checkGit(w[1:], cur, cfg); reason != "" {
				return reason
			}
		}
	}
	return ""
}

func resolve(base, dir string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Join(base, dir)
}

// checkGit checks one `git [-C dir] [-c k=v] <sub> args...` (args excludes the leading "git").
func checkGit(args []string, dir string, cfg config) string {
	for len(args) >= 2 && (args[0] == "-C" || args[0] == "-c") {
		if args[0] == "-C" {
			dir = resolve(dir, args[1])
		}
		args = args[2:]
	}
	if len(args) == 0 {
		return ""
	}
	sub, rest := args[0], args[1:]
	switch {
	case sub == "commit":
		return checkCommit(rest, cfg)
	case sub == "push":
		return checkPush(dir, cfg)
	case mainGuard[sub]:
		return checkMainClone(sub, dir, cfg)
	}
	return ""
}

// commitMessage returns the first -m/-am/--message value; ok is false when there is none.
func commitMessage(args []string) (string, bool) {
	for i, a := range args {
		switch {
		case shortMsg.MatchString(a) || a == "--message":
			if i+1 < len(args) {
				return args[i+1], true
			}
		case strings.HasPrefix(a, "--message="):
			return strings.TrimPrefix(a, "--message="), true
		}
	}
	return "", false
}

// subjectOf returns the first non-empty line of a message, reading through `$(cat <<'EOF' ... EOF)`.
// ok is false for other command substitutions, which cannot be evaluated here.
func subjectOf(msg string) (string, bool) {
	if strings.HasPrefix(msg, "$(") {
		at := strings.Index(msg, "<<")
		if at < 0 {
			return "", false
		}
		word, j := heredocStart(msg, at)
		nl := strings.IndexByte(msg[j:], '\n')
		if word == "" || nl < 0 {
			return "", false
		}
		msg = msg[j+nl+1:]
	}
	for _, line := range strings.Split(msg, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line, true
		}
	}
	return "", true
}

func checkCommit(args []string, cfg config) string {
	msg, ok := commitMessage(args)
	if !ok {
		return ""
	}
	subject, ok := subjectOf(msg)
	if !ok {
		return ""
	}
	m := header.FindStringSubmatch(subject)
	if m == nil {
		return fmt.Sprintf("Commit message is not Angular \"type(scope): subject\": \"%s\".", subject)
	}
	scopes := cfg.Team.Git.Scopes
	if len(scopes) == 0 {
		return ""
	}
	if m[1] == "" {
		return fmt.Sprintf("Commit message needs a scope: \"type(scope): subject\", one of team.yaml git.scopes (%s).", strings.Join(scopes, ", "))
	}
	for _, s := range scopes {
		if s == m[1] {
			return ""
		}
	}
	return fmt.Sprintf("Scope \"%s\" is not in team.yaml git.scopes (%s).", m[1], strings.Join(scopes, ", "))
}

func gitOut(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

func checkPush(dir string, cfg config) string {
	if cfg.reviewTrigger() != "pre-push" {
		return ""
	}
	head, err := gitOut(dir, "rev-parse", "HEAD")
	if err != nil || head == "" {
		return "" // not a repo we can read: fail open
	}
	if hasReview(cfg.Dir, head) {
		return ""
	}
	return fmt.Sprintf("Push blocked: review.trigger is pre-push and no review of HEAD %s exists. Run the review first (a file under efforts/ or people/<handle>/work/ with frontmatter sha: %s).", head[:min(7, len(head))], head)
}

// hasReview reports whether a file under efforts/** or people/*/work/** has frontmatter `sha: <head>`.
func hasReview(team, head string) bool {
	roots, _ := filepath.Glob(filepath.Join(team, "people", "*", "work"))
	roots = append(roots, filepath.Join(team, "efforts"))
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

// checkMainClone denies checkout/switch/pull/reset/rebase in a main clone (`.git` is a directory)
// under projects_dir. Worktrees have a `.git` file and pass.
func checkMainClone(sub, dir string, cfg config) string {
	if cfg.Me.ProjectsDir == "" {
		return ""
	}
	projects, err1 := filepath.EvalSymlinks(cfg.Me.ProjectsDir)
	target, err2 := filepath.EvalSymlinks(dir)
	if err1 != nil || err2 != nil || !within(projects, target) {
		return ""
	}
	top, err := gitOut(target, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	if info, err := os.Stat(filepath.Join(top, ".git")); err != nil || !info.IsDir() {
		return ""
	}
	return fmt.Sprintf("git %s in the main clone %s is not allowed. Create a worktree instead (git worktree add) and work there.", sub, top)
}

func within(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
