// Package team loads the team repo configuration: team.yaml and .local/me.yaml.
package team

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Review is the review section of team.yaml and me.yaml.
type Review struct {
	Trigger string `yaml:"trigger"`
}

// TeamYAML is the part of team.yaml the hooks read.
type TeamYAML struct {
	Git struct {
		Scopes []string `yaml:"scopes"`
	} `yaml:"git"`
	Review Review `yaml:"review"`
}

// MeYAML is the part of .local/me.yaml the hooks read.
type MeYAML struct {
	Handle      string `yaml:"handle"`
	ProjectsDir string `yaml:"projects_dir"`
	Review      Review `yaml:"review"`
}

// Config is the loaded configuration of a team repo.
type Config struct {
	Dir  string // the team repo
	Team TeamYAML
	Me   MeYAML
}

// Load reads team.yaml and .local/me.yaml from the team repo (CLAUDE_PROJECT_DIR, else the cwd).
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
	readYAML(filepath.Join(dir, "team.yaml"), &c.Team)
	readYAML(filepath.Join(dir, ".local", "me.yaml"), &c.Me)
	return c
}

func readYAML(path string, out any) {
	if raw, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(raw, out)
	}
}

// SafeHandle returns the person's handle, or "" when there is none or it is not a plain directory name
// (a separator or ".." would let it point outside people/<handle>/).
func (c Config) SafeHandle() string {
	h := c.Me.Handle
	if h == "" || h == "." || strings.Contains(h, "..") || strings.ContainsAny(h, `/\`) {
		return ""
	}
	return h
}

// Triggers in increasing strictness. Unknown values have no rank.
var triggerRank = map[string]int{"on-demand": 1, "after-commit": 2, "pre-push": 3}

// ReviewTrigger is the effective review.trigger: the stricter of team.yaml and .local/me.yaml.
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
