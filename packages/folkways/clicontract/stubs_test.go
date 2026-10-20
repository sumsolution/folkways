package clicontract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The external tools the CLI runs, stubbed on PATH. Each stub is a shell script that runs this test
// binary with FOLKWAYS_CLI_STUB=<name>; TestMain hands it to runStub. README.md lists what each does.
var stubs = []string{"npm", "pnpm"}

// Versions the stubs report for --version.
var stubVersions = map[string]string{"npm": "10.0.0", "pnpm": "9.0.0"}

func (e *env) writeStubs() {
	self, err := os.Executable()
	if err != nil {
		e.t.Fatal(err)
	}
	for _, name := range stubs {
		e.writeStub(filepath.Join(e.vars["BIN"], name), name, self)
	}
}

func (e *env) writeStub(file, name, self string) {
	script := fmt.Sprintf("#!/bin/sh\nFOLKWAYS_CLI_STUB=%s FOLKWAYS_CLI_CALLS=%s FOLKWAYS_CLI_UNDER_TEST=%s exec %s \"$@\"\n",
		shellQuote(name), shellQuote(e.vars["CALLS"]), shellQuote(e.vars["CLI"]), shellQuote(self))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(script), 0o755); err != nil {
		e.t.Fatal(err)
	}
}

// runStub records the call as "<cwd>: <name> <args>" in ${CALLS}, then does the minimal work.
func runStub(name string, args []string) int {
	cwd, _ := os.Getwd()
	if f, err := os.OpenFile(os.Getenv("FOLKWAYS_CLI_CALLS"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		fmt.Fprintf(f, "%s: %s\n", cwd, strings.TrimSpace(name+" "+strings.Join(args, " ")))
		f.Close()
	}
	key := "STUB_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_FAIL"
	if msg := os.Getenv(key); msg != "" {
		fmt.Fprintln(os.Stderr, msg)
		return 1
	}
	if len(args) == 1 && args[0] == "--version" {
		fmt.Println(stubVersions[name])
		return 0
	}
	if err := packageManager(name, args); err != nil {
		fmt.Fprintf(os.Stderr, "%s stub: %v\n", name, err)
		return 1
	}
	return 0
}

// packageManager handles `install` (no args) and `install|add --save-dev --save-exact <pkg>@<version>`.
func packageManager(name string, args []string) error {
	if len(args) == 0 || (args[0] != "install" && args[0] != "add") {
		return nil
	}
	if len(args) > 1 {
		spec := args[len(args)-1]
		at := strings.LastIndex(spec, "@")
		if at <= 0 {
			return fmt.Errorf("want <pkg>@<version>, got %q", spec)
		}
		if err := pin(spec[:at], spec[at+1:]); err != nil {
			return err
		}
	}
	return install(name)
}

// pin rewrites the package's version in the devDependencies or dependencies block, keeping the key order.
func pin(pkg, version string) error {
	raw, err := os.ReadFile("package.json")
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`("(?:devD|d)ependencies"\s*:\s*\{[^}]*"` + regexp.QuoteMeta(pkg) + `"\s*:\s*)"[^"]*"`)
	if !re.Match(raw) {
		return fmt.Errorf("package.json has no %s dependency to pin", pkg)
	}
	return os.WriteFile("package.json", re.ReplaceAll(raw, []byte(`${1}"`+version+`"`)), 0o644)
}

// install writes what the CLI reads after an install: the lockfile, @folkways/cli at the pinned
// version, node_modules/.bin/folkways (the CLI under test) and node_modules/.bin/folkways-hook (exits 0).
func install(name string) error {
	raw, err := os.ReadFile("package.json")
	if err != nil {
		return err
	}
	var manifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return err
	}
	version := manifest.DevDependencies["@folkways/cli"]
	if version == "" {
		version = manifest.Dependencies["@folkways/cli"]
	}
	lock := map[string]string{"npm": "package-lock.json", "pnpm": "pnpm-lock.yaml"}[name]
	files := map[string]string{
		lock: "stub lockfile\n",
		"node_modules/@folkways/cli/package.json": fmt.Sprintf(`{"name":"@folkways/cli","version":%q}`+"\n", version),
		"node_modules/.bin/folkways":              "#!/bin/sh\nexec " + os.Getenv("FOLKWAYS_CLI_UNDER_TEST") + " \"$@\"\n",
		"node_modules/.bin/folkways-hook":         "#!/bin/sh\nexit 0\n",
	}
	for p, content := range files {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasPrefix(content, "#!") {
			mode = 0o755
		}
		if err := os.WriteFile(p, []byte(content), mode); err != nil {
			return err
		}
	}
	return nil
}
