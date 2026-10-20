// Package sessionstart implements the session-start hook: it reminds the start procedure, and in a repo
// without folkways.yaml makes the one-time setup offer.
package sessionstart

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/output"
	"github.com/sumsolution/folkways/packages/folkways/internal/payload"
	"github.com/sumsolution/folkways/packages/folkways/internal/skill"
	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// Run injects the start reminder, plus a 00-STATUS.md update after a compaction, then the config block.
// In an unconfigured repo it offers setup once instead.
func Run(p payload.Payload, cfg team.Config) {
	if !cfg.Configured {
		if offer(cfg.Dir) {
			output.Context("SessionStart", "folkways isn't set up in this repo. folkways runs from a brain repo: start sessions there, and code changes happen in worktrees of repos like this one. No brain repo yet? Run the `"+skill.Name("setup")+"` skill to create one. This won't be shown again here.")
		}
		return
	}
	text := "folkways: follow the `" + skill.Name("start") + "` skill for this session."
	if p.Source == "compact" {
		text += " The context was just compacted: update the effort's `00-STATUS.md` now, before anything else."
	}
	output.Context("SessionStart", text+configBlock(cfg))
}

const maxFile = 4096

// configBlock is the derived values and the two config files as read, so skills need not read them.
func configBlock(cfg team.Config) string {
	root := "none"
	if r := cfg.PersonalRoot(); r != "" {
		if rel, err := filepath.Rel(cfg.Dir, r); err == nil {
			root = filepath.ToSlash(rel)
		}
	}
	trigger := cfg.ReviewTrigger()
	if trigger == "" {
		trigger = "none"
	}
	return "\n\nfolkways config (from the session-start hook; skills use these values instead of reading the files):\n" +
		"- personal root: " + root + "\n- review trigger: " + trigger + "\n\n" +
		section(cfg.Dir, "folkways.yaml", "") + "\n\n" +
		section(cfg.Dir, ".local/me.yaml", " Run the `"+skill.Name("onboard")+"` skill to create it.")
}

// section is one file as a yaml block, or a one-line pointer when it is missing or too large to print.
func section(dir, name, ifMissing string) string {
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	switch {
	case os.IsNotExist(err):
		return name + ": missing." + ifMissing
	case err != nil:
		return name + ": can't be read. Check the file."
	case len(raw) > maxFile:
		return name + ": over 4096 bytes, not shown. Read the file."
	}
	return name + ":\n```yaml\n" + strings.TrimSuffix(string(raw), "\n") + "\n```"
}

// offer records the repo in ${XDG_CONFIG_HOME:-$HOME/.config}/folkways/offered and reports whether it was new.
// A file that cannot be read or written means no offer: silence beats repeating. Nothing is written in the repo.
func offer(projectDir string) bool {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		base = filepath.Join(home, ".config")
	}
	file := filepath.Join(base, "folkways", "offered")
	key := repoKey(projectDir)
	if info, err := os.Stat(filepath.Join(key, "folkways.yaml")); err == nil && !info.IsDir() {
		return false // a subfolder of a configured repo
	}
	raw, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return false
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == key {
			return false
		}
	}
	if os.MkdirAll(filepath.Dir(file), 0o755) != nil {
		return false
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.WriteString(key + "\n")
	return err == nil
}

// repoKey is the top level of the repository holding the project dir, else the project dir, as an absolute path.
func repoKey(dir string) string {
	if out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
		if top := strings.TrimSpace(string(out)); filepath.IsAbs(top) {
			return top
		}
	}
	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}
	return dir
}
