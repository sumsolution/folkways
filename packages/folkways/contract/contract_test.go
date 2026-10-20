// Black-box hook contract suite. Runs any implementation:
//
//	FOLKWAYS_HOOK_CMD="/abs/path/to/folkways-hook" go test ./contract
//	FOLKWAYS_HOOK_CMD="/abs/path/to/folkways hook" FOLKWAYS_HOOKS=git-guard,precompact go test ./contract
//
// The command gets the hook name appended. Without FOLKWAYS_HOOK_CMD, TestMain builds ./cmd/folkways and
// runs "<bin> hook". FOLKWAYS_HOOKS limits the run to some hooks. See README.md.
package contract

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// hookCmd is the command under test; the hook name is appended to it.
var hookCmd string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	// Isolate git from the developer's config (signing, hooks, identity). Child processes inherit this.
	for k, v := range map[string]string{
		"GIT_CONFIG_GLOBAL":   "/dev/null",
		"GIT_CONFIG_NOSYSTEM": "1",
		"GIT_AUTHOR_NAME":     "test", "GIT_AUTHOR_EMAIL": "test@example.com",
		"GIT_COMMITTER_NAME": "test", "GIT_COMMITTER_EMAIL": "test@example.com",
	} {
		os.Setenv(k, v)
	}
	hookCmd = os.Getenv("FOLKWAYS_HOOK_CMD")
	if hookCmd == "" {
		dir, err := os.MkdirTemp("", "folkways-contract-bin-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer os.RemoveAll(dir)
		bin := filepath.Join(dir, "folkways")
		build := exec.Command("go", "build", "-o", bin, "./cmd/folkways")
		build.Dir = ".." // the module root
		build.Stdout, build.Stderr = os.Stdout, os.Stderr
		if err := build.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "building ./cmd/folkways:", err)
			return 1
		}
		hookCmd = "'" + strings.ReplaceAll(bin, "'", `'\''`) + "' hook"
	}
	return m.Run()
}

// selectedHooks is FOLKWAYS_HOOKS (comma-separated) or every hook. An unknown name, or an empty selection, fails.
func selectedHooks(t *testing.T) []string {
	only := os.Getenv("FOLKWAYS_HOOKS")
	if only == "" {
		return hooks
	}
	var out []string
	for _, o := range strings.Split(only, ",") {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if !slices.Contains(hooks, o) {
			t.Fatalf("FOLKWAYS_HOOKS names unknown hook %q (known: %s)", o, strings.Join(hooks, ", "))
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		t.Fatalf("FOLKWAYS_HOOKS=%q selects no hook", only)
	}
	return out
}

func TestContract(t *testing.T) {
	for _, hook := range selectedHooks(t) {
		t.Run(hook, func(t *testing.T) {
			fixtures := loadFixtures(t, hook)
			if len(fixtures) == 0 {
				t.Fatalf("%s has no fixtures", hook)
			}
			for _, fx := range fixtures {
				t.Run(fx.Name+": "+fx.Description, func(t *testing.T) {
					t.Parallel()
					vars := buildEnv(t, fx.Fixture)
					cpd := fx.Env.ClaudeProjectDir
					if cpd == nil {
						cpd = true
					}
					var problems []string
					for i, step := range fx.Steps {
						result := runHook(t, hookCmd, hook, step, vars, cpd)
						for _, p := range check(step.Expect, result, vars) {
							if len(fx.Steps) > 1 {
								p = fmt.Sprintf("step %d: %s", i+1, p)
							}
							problems = append(problems, p)
						}
					}
					for _, p := range problems {
						t.Error(p)
					}
				})
			}
		})
	}
}

func TestFixturesExist(t *testing.T) {
	for _, hook := range hooks {
		if len(loadFixtures(t, hook)) == 0 {
			t.Errorf("%s has no fixtures", hook)
		}
	}
}
