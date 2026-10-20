// Package pkgmgr picks and drives a team repo's package manager (npm or pnpm).
package pkgmgr

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
)

// Supported are the package managers a team repo can use.
var Supported = []string{"npm", "pnpm"}

func supported(name string) (string, error) {
	if slices.Contains(Supported, name) {
		return name, nil
	}
	return "", clierr.New("%s is not supported; use %s", name, strings.Join(Supported, " or "))
}

// Detect returns the team repo's package manager: the packageManager field, then the lockfile, then npm.
func Detect(team string) (string, error) {
	b, err := os.ReadFile(filepath.Join(team, "package.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if err == nil {
		var pkg struct {
			PackageManager string `json:"packageManager"`
		}
		if err := json.Unmarshal(b, &pkg); err != nil {
			return "", err
		}
		if pkg.PackageManager != "" {
			return supported(strings.Split(pkg.PackageManager, "@")[0])
		}
	}
	if _, err := os.Stat(filepath.Join(team, "pnpm-lock.yaml")); err == nil {
		return "pnpm", nil
	}
	return "npm", nil
}

// FromUserAgent returns the package manager that launched us (npx or pnpm dlx), from the value of
// npm_config_user_agent, or "" when it is neither.
func FromUserAgent(ua string) string {
	name := strings.Split(ua, "/")[0]
	if slices.Contains(Supported, name) {
		return name
	}
	return ""
}

// InstallCommand installs from the lockfile.
func InstallCommand(pm string) []string { return []string{pm, "install"} }

// AddDevExactCommand pins spec (name@version) as an exact dev dependency.
func AddDevExactCommand(pm, spec string) []string {
	verb := "install"
	if pm == "pnpm" {
		verb = "add"
	}
	return []string{pm, verb, "--save-dev", "--save-exact", spec}
}

// PackageManagerField is "pnpm@<installed version>" for package.json's packageManager field, "" for npm.
func PackageManagerField(pm string) (string, error) {
	if pm == "npm" {
		return "", nil
	}
	out, err := exec.Command("pnpm", "--version").Output()
	if err != nil {
		return "", clierr.New("pnpm is not installed; install it or use --package-manager npm")
	}
	return "pnpm@" + strings.TrimSpace(string(out)), nil
}

// Exec runs argv in dir with stdin closed and stdout to stdout (nil discards it). A failure is an error whose
// message is the command's trimmed stderr, or Go's error when stderr is empty: what init and upgrade print.
func Exec(dir string, argv []string, stdout io.Writer) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir, cmd.Stdout = dir, stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if detail := strings.TrimSpace(stderr.String()); detail != "" {
		return errors.New(detail)
	}
	return err
}
