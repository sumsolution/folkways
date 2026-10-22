// Package team loads the team repo configuration: folkways.yaml and .local/me.yaml.
package team

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Review is the review section of folkways.yaml and me.yaml.
type Review struct {
	Trigger string `yaml:"trigger"`
}

// TeamYAML is the part of folkways.yaml the hooks read.
type TeamYAML struct {
	Mode string `yaml:"mode"` // team (default) | solo
	Git  struct {
		CommitConvention string   `yaml:"commit_convention"` // angular (default) | pattern | none
		CommitPattern    string   `yaml:"commit_pattern"`
		Scopes           []string `yaml:"scopes"`
	} `yaml:"git"`
	Hooks   map[string]bool `yaml:"hooks"` // a hook set to false is off; a missing key is on
	Review  Review          `yaml:"review"`
	Handoff struct {
		Bands []int `yaml:"bands"` // token thresholds for context-band
	} `yaml:"handoff"`
}

// MeYAML is the part of .local/me.yaml the hooks read.
type MeYAML struct {
	Handle      string `yaml:"handle"`
	ProjectsDir string `yaml:"projects_dir"`
	Review      Review `yaml:"review"`
}

// Config is the loaded configuration of a team repo.
type Config struct {
	Dir        string // the project root
	Configured bool   // folkways.yaml exists at the project root
	Team       TeamYAML
	Me         MeYAML
	Version    string // the running binary's version; hookrun sets it
}

// Load reads folkways.yaml and .local/me.yaml from the team repo (CLAUDE_PROJECT_DIR, else the cwd).
// Missing or broken files leave the zero value: the checks that need them simply do not fire.
func Load() Config {
	dir := os.Getenv("CLAUDE_PROJECT_DIR")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	return LoadDir(dir)
}

// LoadDir reads the configuration of the team repo at dir.
func LoadDir(dir string) Config {
	c := Config{Dir: dir}
	if info, err := os.Stat(filepath.Join(dir, "folkways.yaml")); err == nil && !info.IsDir() {
		c.Configured = true
	}
	readYAML(filepath.Join(dir, "folkways.yaml"), &c.Team)
	readYAML(filepath.Join(dir, ".local", "me.yaml"), &c.Me)
	return c
}

func readYAML(path string, out any) {
	if raw, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(raw, out)
	}
}

// SafeHandle returns the person's handle, or "" when there is none or it is not a plain directory name
// (a separator or ".." would let it point outside people/<handle>/; whitespace could forge lines in the
// session-start context).
func (c Config) SafeHandle() string {
	h := c.Me.Handle
	if h == "" || h == "." || strings.Contains(h, "..") || strings.ContainsAny(h, "/\\ \t\r\n\v\f") {
		return ""
	}
	return h
}

// Triggers in increasing strictness. Unknown values have no rank.
var triggerRank = map[string]int{"on-demand": 1, "after-commit": 2, "pre-push": 3}

// ReviewTrigger is the effective review.trigger: the stricter of folkways.yaml and .local/me.yaml.
// A person can only tighten the team's setting. A value that is not a known trigger is ignored.
func (c Config) ReviewTrigger() string {
	best := ""
	for _, t := range []string{c.Team.Review.Trigger, c.Me.Review.Trigger} {
		if triggerRank[t] > triggerRank[best] {
			best = t
		}
	}
	return best
}

// HookEnabled reports whether folkways.yaml leaves the hook on: only an explicit `false` turns it off.
func (c Config) HookEnabled(name string) bool {
	on, set := c.Team.Hooks[name]
	return !set || on
}

// Solo reports whether the repo is a solo repo (`mode: solo`).
func (c Config) Solo() bool { return c.Team.Mode == "solo" }

// PersonalRoot is the directory that holds the person's inbox/ and work/: the repo root in a solo repo,
// people/<handle>/ in a team repo. It is "" when a team repo has no safe handle.
func (c Config) PersonalRoot() string {
	if c.Solo() {
		return c.Dir
	}
	if h := c.SafeHandle(); h != "" {
		return filepath.Join(c.Dir, "people", h)
	}
	return ""
}
