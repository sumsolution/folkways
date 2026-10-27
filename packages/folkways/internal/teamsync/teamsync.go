// Package teamsync copies the framework content into a team repo (`folkways sync`). A port of
// packages/cli/src/sync.ts.
package teamsync

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
	"github.com/sumsolution/folkways/packages/folkways/internal/files"
	"github.com/sumsolution/folkways/packages/folkways/internal/hooks"
	"github.com/sumsolution/folkways/packages/folkways/internal/manifest"
	"github.com/sumsolution/folkways/packages/folkways/internal/overlays"
	"github.com/sumsolution/folkways/packages/folkways/internal/settings"
	teamcfg "github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// InstructionsImport is the AGENTS.md line that loads the framework instructions.
const InstructionsImport = "@.folkways/instructions.md"

// stagePrefix names the staging folder of a sync in progress, inside the team repo.
const stagePrefix = ".folkways-sync-"

const settingsPath = ".claude/settings.json"

// pushAskRule is the framework's permission rule for pushes (packages/content/settings.json).
const pushAskRule = "Bash(git push:*)"

// ops is the seam the atomicity tests replace.
var ops = files.OS

var overlayFile = regexp.MustCompile(`^([^/]+)/OVERLAY\.md$`)

// Options are sync's flags.
type Options struct {
	// Force discards edits to framework-owned files instead of stopping.
	Force bool
}

// Result is what a sync did.
type Result struct {
	Version   string
	Previous  string // the manifest's version before the sync; "" without a manifest
	Discarded []string
	Warnings  []string
}

// fileCopy is one file sync copies from the bundle into the team repo.
type fileCopy struct {
	src, dest string // src in the bundle's FS, dest relative to the team repo
	// content, when not nil, replaces src (a skill file with a team overlay applied).
	content []byte
}

func (c fileCopy) hash(fsys fs.FS) (string, error) {
	if c.content != nil {
		return files.SHA256Text(c.content), nil
	}
	b, err := fs.ReadFile(fsys, c.src)
	if err != nil {
		return "", err
	}
	return files.SHA256Text(b), nil
}

// assetMap lists every file sync copies from the bundle into a team repo.
func assetMap(fsys fs.FS) ([]fileCopy, error) {
	out := []fileCopy{{src: "instructions.md", dest: ".folkways/instructions.md"}}
	prompts, err := files.ListFS(fsys, "prompts")
	if err != nil {
		return nil, err
	}
	for _, f := range prompts {
		out = append(out, fileCopy{src: "prompts/" + f, dest: ".folkways/prompts/" + f})
	}
	skills, err := files.ListFS(fsys, "skills")
	if err != nil {
		return nil, err
	}
	for _, f := range skills {
		if strings.Contains(f, "/") {
			out = append(out, fileCopy{src: "skills/" + f, dest: ".claude/skills/" + f})
		}
	}
	workflows, err := files.ListFS(fsys, "workflows")
	for _, f := range workflows {
		if !strings.Contains(f, "/") && strings.HasSuffix(f, ".yml") {
			out = append(out, fileCopy{src: "workflows/" + f, dest: ".github/workflows/folkways-" + f})
		}
	}
	return out, err
}

// skillNames are the framework skills in the bundle, sorted.
func skillNames(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, "skills")
	if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil // ReadDir sorts
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func lexists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

// addUnique appends v unless list has it: a JS Set's first insertion order.
func addUnique(list *[]string, v string) {
	if !slices.Contains(*list, v) {
		*list = append(*list, v)
	}
}

// skillFiles lists the files in a team skill folder, as paths relative to the team repo.
func skillFiles(team, skill string) ([]string, error) {
	list, err := files.List(filepath.Join(team, ".claude/skills", skill))
	for i, f := range list {
		list[i] = ".claude/skills/" + skill + "/" + f
	}
	return list, err
}

// FindEdits lists framework-owned files changed since the last sync, e.g. "modified  .folkways/instructions.md".
// Files are checked in sorted path order (TS: manifest order).
func FindEdits(team string, m *manifest.Manifest) ([]string, error) {
	edits := []string{}
	for _, p := range slices.Sorted(maps.Keys(m.Files)) {
		full := filepath.Join(team, p)
		if !exists(full) {
			edits = append(edits, "deleted   "+p)
			continue
		}
		sum, err := files.SHA256(full)
		if err != nil {
			return nil, err
		}
		if sum != m.Files[p] {
			edits = append(edits, "modified  "+p)
		}
	}
	for _, skill := range m.Skills {
		list, err := skillFiles(team, skill)
		if err != nil {
			return nil, err
		}
		for _, p := range list {
			if _, ok := m.Files[p]; !ok {
				edits = append(edits, "added     "+p)
			}
		}
	}
	return edits, nil
}

// EditsMessage is the refusal for edited framework files; fix is its last line.
func EditsMessage(edits []string, fix string) string {
	lines := []string{"framework files were edited:"}
	for _, e := range edits {
		lines = append(lines, "  "+e)
	}
	return strings.Join(append(lines, "Move team changes into overlays/, then rerun.", fix), "\n")
}

// blockingParent is the first existing parent of dest that isn't a folder, e.g. a `.folkways` file.
func blockingParent(team, dest string) string {
	parts := strings.Split(dest, "/")
	for i := 1; i < len(parts); i++ {
		p := strings.Join(parts[:i], "/")
		info, err := os.Lstat(filepath.Join(team, p))
		if err != nil {
			return ""
		}
		if !info.IsDir() {
			return p
		}
	}
	return ""
}

// findOverwrites lists team files this version would overwrite that the manifest doesn't track, e.g.
// "replaced  .folkways/x.md". Without a manifest (adopting a repo) that is every destination, plus extra
// files in framework skill folders. Files that already match are fine.
func findOverwrites(team string, fsys fs.FS, m *manifest.Manifest, copies []fileCopy) ([]string, error) {
	tracked := map[string]bool{}
	var ownedSkills []string
	if m != nil {
		for p := range m.Files {
			tracked[p] = true
		}
		for _, s := range m.Skills {
			ownedSkills = append(ownedSkills, ".claude/skills/"+s+"/")
		}
	}
	var out []string
	dests := make([]string, 0, len(copies)+2)
	for _, c := range copies {
		dests = append(dests, c.dest)
	}
	for _, dest := range append(dests, settingsPath, manifest.Path) {
		if parent := blockingParent(team, dest); parent != "" {
			addUnique(&out, "replaced  "+parent)
		}
	}
	for _, c := range copies {
		owned := slices.ContainsFunc(ownedSkills, func(dir string) bool { return strings.HasPrefix(c.dest, dir) })
		if tracked[c.dest] || owned || blockingParent(team, c.dest) != "" {
			continue
		}
		full := filepath.Join(team, c.dest)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			addUnique(&out, "replaced  "+c.dest)
			continue
		}
		have, err := files.SHA256(full)
		if err != nil {
			return nil, err
		}
		want, err := c.hash(fsys)
		if err != nil {
			return nil, err
		}
		if have != want {
			addUnique(&out, "replaced  "+c.dest)
		}
	}
	if m == nil {
		// Adopting: files in framework skill folders that this version doesn't ship would be left behind.
		shipped := map[string]bool{}
		for _, c := range copies {
			shipped[c.dest] = true
		}
		skills, err := skillNames(fsys)
		if err != nil {
			return nil, err
		}
		for _, skill := range skills {
			if blockingParent(team, ".claude/skills/"+skill+"/x") != "" {
				continue
			}
			list, err := skillFiles(team, skill)
			if err != nil {
				return nil, err
			}
			for _, p := range list {
				if !shipped[p] {
					addUnique(&out, "replaced  "+p)
				}
			}
		}
	}
	return out, nil
}

// replacedPaths are the paths this sync replaces or removes: the last sync's files, edits in its skills,
// forced overwrites, and every existing destination, so a failed swap can put all of them back.
func replacedPaths(team string, m *manifest.Manifest, overwrites, incoming []string) ([]string, error) {
	var paths, skills []string
	if m != nil {
		paths, skills = slices.Sorted(maps.Keys(m.Files)), m.Skills
	}
	for _, p := range incoming {
		addUnique(&paths, p)
	}
	for _, skill := range skills {
		list, err := skillFiles(team, skill)
		if err != nil {
			return nil, err
		}
		for _, p := range list {
			addUnique(&paths, p)
		}
	}
	for _, line := range overwrites {
		addUnique(&paths, strings.TrimPrefix(line, "replaced  "))
	}
	addUnique(&paths, settingsPath)
	addUnique(&paths, manifest.Path)
	var out []string
	for _, p := range paths {
		if blockingParent(team, p) == "" && lexists(filepath.Join(team, p)) {
			out = append(out, p)
		}
	}
	return out, nil
}

// mkdirAll is mkdirSync(dir, {recursive: true}): it returns the topmost folder it created, "" for none.
func mkdirAll(dir string) (string, error) {
	first := ""
	for d := dir; !lexists(d); d = filepath.Dir(d) {
		first = d
		if filepath.Dir(d) == d {
			break
		}
	}
	return first, os.MkdirAll(dir, 0o755)
}

// swap moves the outgoing team paths aside and the staged incoming paths (manifest last) into place. On any
// failure it undoes every move, so the team repo keeps the old set of files.
func swap(team, stage string, outgoing, incoming []string) error {
	type move struct{ from, to string }
	var done []move
	var created []string
	moveOne := func(from, to string) error {
		made, err := mkdirAll(filepath.Dir(to))
		if made != "" {
			created = append(created, made)
		}
		if err != nil {
			return err
		}
		if err := ops.Rename(from, to); err != nil {
			return err
		}
		done = append(done, move{from, to})
		return nil
	}
	err := func() error {
		for _, p := range outgoing {
			if err := moveOne(filepath.Join(team, p), filepath.Join(stage, "old", p)); err != nil {
				return err
			}
		}
		for _, p := range incoming {
			if err := moveOne(filepath.Join(stage, "new", p), filepath.Join(team, p)); err != nil {
				return err
			}
		}
		return nil
	}()
	if err == nil {
		return nil
	}
	undo := func() error {
		for i := len(done) - 1; i >= 0; i-- {
			if err := os.MkdirAll(filepath.Dir(done[i].from), 0o755); err != nil {
				return err
			}
			if err := ops.Rename(done[i].to, done[i].from); err != nil {
				return err
			}
		}
		for i := len(created) - 1; i >= 0; i-- {
			if err := files.PruneEmptyDirs(created[i]); err != nil {
				return err
			}
		}
		return nil
	}()
	if undo != nil {
		return &restoreError{&clierr.Error{Msg: fmt.Sprintf("sync failed and could not restore the old files; they are in %s/old.\n%v\n%v", stage, err, undo)}}
	}
	return err
}

// restoreError is a failed sync whose undo failed too: the staging folder holds the old files, so it stays.
// It wraps a *clierr.Error, as TS RestoreError extends FolkwaysError.
type restoreError struct{ cause *clierr.Error }

func (e *restoreError) Error() string { return e.cause.Msg }
func (e *restoreError) Unwrap() error { return e.cause }

type teamOverlay struct{ path, skill string }

// teamOverlays are the team's overlays: overlays/<skill>/OVERLAY.md.
func teamOverlays(team string) ([]teamOverlay, error) {
	list, err := files.List(filepath.Join(team, "overlays"))
	if err != nil {
		return nil, err
	}
	var out []teamOverlay
	for _, f := range list {
		if m := overlayFile.FindStringSubmatch(f); m != nil {
			out = append(out, teamOverlay{path: "overlays/" + f, skill: m[1]})
		}
	}
	return out, nil
}

// compileOverlays compiles each team overlay into the copies of its framework skill (ADR 0001). It only
// reads, so an unknown section ID stops the sync before anything is written. It returns each compiled
// overlay's checksum and the warnings.
func compileOverlays(team string, fsys fs.FS, copies []fileCopy) (map[string]string, []string, error) {
	shipped, err := skillNames(fsys)
	if err != nil {
		return nil, nil, err
	}
	list, err := teamOverlays(team)
	if err != nil {
		return nil, nil, err
	}
	sums := map[string]string{}
	var warnings []string
	for _, o := range list {
		if !slices.Contains(shipped, o.skill) {
			warnings = append(warnings, fmt.Sprintf("%s: no framework skill named %s; the overlay is not applied", o.path, o.skill))
			continue
		}
		prefix := ".claude/skills/" + o.skill + "/"
		skillFiles := map[string]string{}
		for _, c := range copies {
			if strings.HasPrefix(c.dest, prefix) && strings.HasSuffix(c.dest, ".md") {
				b, err := fs.ReadFile(fsys, c.src)
				if err != nil {
					return nil, nil, err
				}
				skillFiles[strings.TrimPrefix(c.dest, prefix)] = string(b)
			}
		}
		text, err := os.ReadFile(filepath.Join(team, o.path))
		if err != nil {
			return nil, nil, err
		}
		compiled, err := overlays.CompileSkillFiles(skillFiles, string(text))
		if err != nil {
			var ce *clierr.Error
			if errors.As(err, &ce) {
				return nil, nil, clierr.New("%s: %s", o.path, ce.Msg)
			}
			return nil, nil, err
		}
		for i := range copies {
			if rel, ok := strings.CutPrefix(copies[i].dest, prefix); ok {
				if s, ok := compiled.Files[rel]; ok {
					copies[i].content = []byte(s)
				}
			}
		}
		for _, w := range compiled.Warnings {
			warnings = append(warnings, o.path+": "+strings.TrimPrefix(w, "overlays: "))
		}
		sums[o.path] = files.SHA256Text(text)
	}
	return sums, warnings, nil
}

// StaleOverlays lists team overlays changed, added or removed since the last sync compiled them.
func StaleOverlays(team string, m *manifest.Manifest) ([]string, error) {
	list, err := teamOverlays(team)
	if err != nil {
		return nil, err
	}
	paths := slices.Collect(maps.Keys(m.Overlays))
	for _, o := range list {
		if slices.Contains(m.Skills, o.skill) {
			paths = append(paths, o.path)
		}
	}
	slices.Sort(paths)
	out := []string{}
	for _, p := range slices.Compact(paths) {
		full := filepath.Join(team, p)
		if !exists(full) {
			out = append(out, p)
			continue
		}
		sum, err := files.SHA256(full)
		if err != nil {
			return nil, err
		}
		if recorded, ok := m.Overlays[p]; !ok || recorded != sum {
			out = append(out, p)
		}
	}
	return out, nil
}

// mergedSettings is the merged .claude/settings.json content. It only reads, so sync can fail before
// writing anything.
func mergedSettings(teamDir string, fsys fs.FS) (string, error) {
	p := filepath.Join(teamDir, settingsPath)
	var current any = &settings.Object{}
	if exists(p) {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		if current, err = settings.Parse(b); err != nil {
			return "", clierr.New("%s is not valid JSON", p)
		}
	}
	b, err := fs.ReadFile(fsys, "settings.json")
	if err != nil {
		return "", err
	}
	fragment, _ := settings.Parse(b) // the embedded fragment is a JSON object; content tests check it
	merged, err := settings.Merge(current, fragment.(*settings.Object))
	if err != nil {
		return "", err
	}
	// The push ask rule is the fallback for push.brain: always when hooks are off; git-guard asks otherwise
	// (ADR 0014). It reads folkways.yaml only: the committed file must not depend on one person's me.yaml.
	if (teamcfg.Config{Team: teamcfg.LoadDir(teamDir).Team}).PushApproval("brain") != "always" {
		settings.DropAsk(merged, pushAskRule)
	}
	return settings.Stringify(merged) + "\n", nil
}

// Sync copies b's files into the team repo at team. It does not commit.
func Sync(team string, b bundle.Bundle, opts Options) (_ Result, err error) {
	entries, err := os.ReadDir(team)
	if err != nil {
		return Result{}, err
	}
	for _, e := range entries {
		if leftover := e.Name(); strings.HasPrefix(leftover, stagePrefix) {
			return Result{}, clierr.New("a previous sync was interrupted and left %s/ (the files it replaced are in %s/old).\n"+
				"Restore the committed files with 'git checkout -- .', delete %s/, then rerun.", leftover, leftover, leftover)
		}
	}
	m, err := manifest.Read(team)
	if err != nil {
		return Result{}, err
	}
	discarded := []string{}
	if m != nil {
		if discarded, err = FindEdits(team, m); err != nil {
			return Result{}, err
		}
	}
	if len(discarded) > 0 && !opts.Force {
		return Result{}, clierr.New("%s", EditsMessage(discarded, "Or rerun with --force to discard the edits."))
	}

	skills, err := skillNames(b.FS)
	if err != nil {
		return Result{}, err
	}
	// Without a manifest (adopting a repo), findOverwrites compares skill folders file by file instead.
	if m != nil {
		var clashes []string
		for _, s := range skills {
			if !slices.Contains(m.Skills, s) && exists(filepath.Join(team, ".claude/skills", s)) {
				clashes = append(clashes, ".claude/skills/"+s)
			}
		}
		if len(clashes) > 0 {
			return Result{}, clierr.New("team skills clash with new framework skills: %s\nRename the team skills, then rerun.",
				strings.Join(clashes, ", "))
		}
	}

	copies, err := assetMap(b.FS)
	if err != nil {
		return Result{}, err
	}
	overlaySums, overlayWarnings, err := compileOverlays(team, b.FS, copies)
	if err != nil {
		return Result{}, err
	}
	overwrites, err := findOverwrites(team, b.FS, m, copies)
	if err != nil {
		return Result{}, err
	}
	if len(overwrites) > 0 && !opts.Force {
		lines := []string{"sync would overwrite team files that differ from this version:"}
		for _, o := range overwrites {
			lines = append(lines, "  "+o)
		}
		lines = append(lines, "Move team changes into overlays/, then rerun.", "Or rerun with --force to overwrite them.")
		return Result{}, clierr.New("%s", strings.Join(lines, "\n"))
	}

	merged, err := mergedSettings(team, b.FS)
	if err != nil {
		return Result{}, err
	}

	// Stage the new files inside the team repo (same filesystem, so renames are atomic), then swap.
	stage, err := os.MkdirTemp(team, stagePrefix+"*")
	if err != nil {
		return Result{}, err
	}
	defer func() {
		if !errors.As(err, new(*restoreError)) {
			os.RemoveAll(stage)
		}
	}()
	sums := map[string]string{}
	for _, c := range copies {
		staged := filepath.Join(stage, "new", c.dest)
		if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return Result{}, err
		}
		if c.content != nil {
			err = os.WriteFile(staged, c.content, 0o644)
		} else {
			err = ops.Copy(b.FS, c.src, staged)
		}
		if err != nil {
			return Result{}, err
		}
		if sums[c.dest], err = files.SHA256(staged); err != nil {
			return Result{}, err
		}
	}
	if err := os.MkdirAll(filepath.Join(stage, "new/.claude"), 0o755); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(stage, "new", settingsPath), []byte(merged), 0o644); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(filepath.Join(stage, "new/.folkways"), 0o755); err != nil {
		return Result{}, err
	}
	next := &manifest.Manifest{Version: b.Version, Skills: skills, Files: sums}
	if len(overlaySums) > 0 {
		next.Overlays = overlaySums
	}
	if err := manifest.Write(filepath.Join(stage, "new"), next); err != nil {
		return Result{}, err
	}
	incoming := make([]string, 0, len(copies)+2)
	for _, c := range copies {
		incoming = append(incoming, c.dest)
	}
	incoming = append(incoming, settingsPath, manifest.Path)
	outgoing, err := replacedPaths(team, m, overwrites, incoming)
	if err != nil {
		return Result{}, err
	}
	if err := swap(team, stage, outgoing, incoming); err != nil {
		return Result{}, err
	}
	if err := os.RemoveAll(stage); err != nil {
		return Result{}, err
	}
	var prune []string
	if m != nil {
		for _, s := range m.Skills {
			prune = append(prune, filepath.Join(team, ".claude/skills", s))
		}
	}
	for _, dir := range append(prune, filepath.Join(team, ".folkways"), filepath.Join(team, ".github")) {
		if err := files.PruneEmptyDirs(dir); err != nil {
			return Result{}, err
		}
	}

	warnings := append(overlayWarnings, hooks.Warnings()...)
	if !exists(filepath.Join(team, "folkways.yaml")) {
		warnings = append(warnings, "no folkways.yaml; the hooks stay silent until the repo has one.")
	}
	agents, err := os.ReadFile(filepath.Join(team, "AGENTS.md"))
	imports := err == nil && slices.ContainsFunc(strings.Split(string(agents), "\n"), func(l string) bool {
		return strings.TrimSpace(l) == InstructionsImport
	})
	if !imports {
		warnings = append(warnings, "AGENTS.md has no '"+InstructionsImport+"' line; agents won't see the framework instructions.")
	}

	previous := ""
	if m != nil {
		previous = m.Version
	}
	return Result{Version: b.Version, Previous: previous, Discarded: append(discarded, overwrites...), Warnings: warnings}, nil
}
