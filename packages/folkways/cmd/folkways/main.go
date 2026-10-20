// Command folkways is the folkways CLI (ADR 0011). `folkways hook <name>` runs one hook in-process; the
// other commands dispatch to their packages, which the go-cli port fills in.
package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/bundle"
	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
	"github.com/sumsolution/folkways/packages/folkways/internal/doctor"
	"github.com/sumsolution/folkways/packages/folkways/internal/git"
	"github.com/sumsolution/folkways/packages/folkways/internal/hookrun"
	"github.com/sumsolution/folkways/packages/folkways/internal/pkgmgr"
	"github.com/sumsolution/folkways/packages/folkways/internal/scaffold"
	"github.com/sumsolution/folkways/packages/folkways/internal/teamsync"
	"github.com/sumsolution/folkways/packages/folkways/internal/upgrade"
)

// version is set at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

const usage = `usage: folkways <command> [options]

  init <path> [--no-install] [--package-manager npm|pnpm]
                               scaffold a team repo at <path>, pin this version, sync, commit
                               (npx @folkways/cli init <path>, or pnpm dlx @folkways/cli init <path>)
                               --package-manager  default: the one that ran this, else npm
  sync [--force]               copy the installed version's files into this team repo
                               --force  discard edits to framework-owned files and overwrite
                                        differing team files at framework paths
  upgrade <version>            install <version> on branch chore/folkways-<version>, sync, commit
                               (uses the team repo's package manager)
  hook <name>                  run a hook by hand; reads the payload on stdin
  doctor                       check config, versions, overlays, skills, settings and hooks;
                               prints each problem with its fix, exits 1 on errors
  version                      print the installed version
  review | run                 not built yet`

func main() {
	// The hook path comes first and always exits 0: Copilot CLI denies the tool call when a preToolUse
	// hook exits nonzero or crashes.
	if len(os.Args) >= 2 && os.Args[1] == "hook" {
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "folkways: usage: folkways hook <name> < payload.json")
			return
		}
		hookrun.Run(os.Args[2], os.Stdin)
		return
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run runs every command except hook and returns the exit code. Errors print as `folkways: <message>`.
func run(args []string, stdout, stderr io.Writer) int {
	code, err := dispatch(args, stdout, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "folkways: %v\n", err)
		return 1
	}
	return code
}

func dispatch(args []string, stdout, stderr io.Writer) (int, error) {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 1, nil
	}
	command, rest := args[0], args[1:]
	b := bundle.Embedded(version)
	switch command {
	case "init":
		pos, opts, err := parseArgs(rest, []string{"no-install"}, []string{"package-manager"}, true)
		if err != nil {
			return 1, err
		}
		if len(pos) != 1 {
			return 1, clierr.New("usage: folkways init <path> [--no-install] [--package-manager npm|pnpm]")
		}
		pm, set := opts["package-manager"]
		if !set {
			pm = pkgmgr.FromUserAgent(os.Getenv("npm_config_user_agent"))
		}
		if !set && pm == "" {
			pm = "npm"
		}
		if !slices.Contains(pkgmgr.Supported, pm) {
			return 1, clierr.New("--package-manager must be %s", strings.Join(pkgmgr.Supported, " or "))
		}
		dest, err := scaffold.Init(pos[0], b, scaffold.Options{Install: opts["no-install"] == "", PackageManager: pm})
		if err != nil {
			return 1, err
		}
		fmt.Fprintf(stdout, "Created %s with folkways %s.\n", dest, version)
		fmt.Fprintf(stdout, "Next: open Claude Code or Copilot CLI in %s and run the setup skill.\n", dest)
	case "sync":
		_, opts, err := parseArgs(rest, []string{"force"}, nil, false)
		if err != nil {
			return 1, err
		}
		team, err := teamRoot()
		if err != nil {
			return 1, err
		}
		r, err := teamsync.Sync(team, b, teamsync.Options{Force: opts["force"] != ""})
		if err != nil {
			return 1, err
		}
		if len(r.Discarded) > 0 {
			fmt.Fprintf(stdout, "Discarded edits:\n  %s\n", strings.Join(r.Discarded, "\n  "))
		}
		for _, w := range r.Warnings {
			fmt.Fprintf(stdout, "warning: %s\n", w)
		}
		from := ""
		if r.Previous != "" && r.Previous != r.Version {
			from = r.Previous + " -> "
		}
		fmt.Fprintf(stdout, "Synced folkways %s%s. Review with git status, then commit.\n", from, r.Version)
	case "upgrade":
		pos, _, err := parseArgs(rest, nil, nil, true)
		if err != nil {
			return 1, err
		}
		if len(pos) != 1 {
			return 1, clierr.New("usage: folkways upgrade <version>")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return 1, err
		}
		r, err := upgrade.Upgrade(cwd, pos[0], upgrade.PackageInstall)
		if err != nil {
			return 1, err
		}
		if r.Branch == "" {
			fmt.Fprintf(stdout, "Already at %s; nothing to change.\n", r.To)
		} else {
			fmt.Fprintf(stdout, "Committed %s -> %s on %s. Review, push, open a PR.\n", r.From, r.To, r.Branch)
		}
	case "doctor":
		if _, _, err := parseArgs(rest, nil, nil, false); err != nil {
			return 1, err
		}
		team, err := teamRoot()
		if err != nil {
			return 1, err
		}
		home, _ := os.UserHomeDir()
		findings, err := doctor.Doctor(team, b, doctor.Options{Home: home})
		if err != nil {
			return 1, err
		}
		errs := 0
		for _, f := range findings {
			fmt.Fprintf(stdout, "%s: %s\n", f.Level, f.Message)
			if f.Level == "error" {
				errs++
			}
		}
		if len(findings) == 0 {
			fmt.Fprintln(stdout, "No problems found.")
		} else {
			fmt.Fprintf(stdout, "%d %s, %d %s.\n", errs, plural(errs, "error"), len(findings)-errs, plural(len(findings)-errs, "warning"))
		}
		if errs > 0 {
			return 1, nil
		}
	case "version", "--version":
		if _, _, err := parseArgs(rest, nil, nil, false); err != nil {
			return 1, err
		}
		fmt.Fprintln(stdout, version)
	case "review", "run":
		return 1, clierr.New("'%s' is not built yet", command)
	case "help", "--help", "-h":
		if _, _, err := parseArgs(rest, nil, nil, false); err != nil {
			return 1, err
		}
		fmt.Fprintln(stdout, usage)
	default:
		fmt.Fprintln(stderr, usage)
		return 1, nil
	}
	return 0, nil
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// teamRoot is the top of the git repo around the cwd.
func teamRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := git.Run(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", clierr.New("run this inside the team repo")
	}
	return root, nil
}

// parseArgs parses long options anywhere in args, like Node's util.parseArgs in strict mode: boolean flags
// map to "true", string flags take `--name value` or `--name=value`, and `--` ends the options. The error
// texts start as Node's do.
func parseArgs(args, bools, strs []string, positionals bool) ([]string, map[string]string, error) {
	var pos []string
	opts := map[string]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			pos = append(pos, a)
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		switch {
		case strings.HasPrefix(a, "--") && slices.Contains(bools, name):
			if hasValue {
				return nil, nil, clierr.New("Option '--%s' does not take an argument", name)
			}
			opts[name] = "true"
		case strings.HasPrefix(a, "--") && slices.Contains(strs, name):
			if !hasValue {
				if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
					return nil, nil, clierr.New("Option '--%s <value>' argument missing", name)
				}
				i++
				value = args[i]
			}
			opts[name] = value
		default:
			return nil, nil, clierr.New("Unknown option '%s'", strings.SplitN(a, "=", 2)[0])
		}
	}
	if !positionals && len(pos) > 0 {
		return nil, nil, clierr.New("Unexpected argument '%s'. This command does not take positional arguments", pos[0])
	}
	return pos, opts, nil
}
