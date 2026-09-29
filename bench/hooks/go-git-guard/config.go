package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type review struct {
	Trigger string `yaml:"trigger"`
}

type teamYAML struct {
	Git struct {
		Scopes []string `yaml:"scopes"`
	} `yaml:"git"`
	Review review `yaml:"review"`
}

type meYAML struct {
	ProjectsDir string `yaml:"projects_dir"`
	Review      review `yaml:"review"`
}

type config struct {
	Dir  string // the team repo
	Team teamYAML
	Me   meYAML
}

// loadConfig reads team.yaml and .local/me.yaml from the team repo (CLAUDE_PROJECT_DIR, else the cwd).
// Missing or broken files leave the zero value: the checks that need them simply do not fire.
func loadConfig() config {
	dir := os.Getenv("CLAUDE_PROJECT_DIR")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	c := config{Dir: dir}
	readYAML(filepath.Join(dir, "team.yaml"), &c.Team)
	readYAML(filepath.Join(dir, ".local", "me.yaml"), &c.Me)
	return c
}

func readYAML(path string, out any) {
	if raw, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(raw, out)
	}
}

// reviewTrigger is the effective review.trigger: the personal setting over the team's.
func (c config) reviewTrigger() string {
	if c.Me.Review.Trigger != "" {
		return c.Me.Review.Trigger
	}
	return c.Team.Review.Trigger
}
