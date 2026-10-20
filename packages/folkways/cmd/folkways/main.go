// Command folkways is the folkways CLI (ADR 0011). `folkways hook <name>` runs one hook; the other
// commands are still in the TypeScript CLI until the Go port reaches parity.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sumsolution/folkways/packages/folkways/internal/hookrun"
)

// version is set at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

const usage = `usage: folkways <command>

  hook <name>   run a hook; reads the payload on stdin
  version       print the version`

func main() {
	// The hook path comes first and always exits 0: Copilot CLI denies the tool call when a preToolUse
	// hook exits nonzero or crashes.
	if len(os.Args) >= 2 && os.Args[1] == "hook" {
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: folkways hook <name> < payload.json")
			return
		}
		hookrun.Run(os.Args[2], os.Stdin)
		return
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run runs every command except hook and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 1
	}
	switch args[0] {
	case "version", "--version":
		fmt.Fprintln(stdout, version)
		return 0
	case "help", "--help", "-h":
		fmt.Fprintln(stdout, usage)
		return 0
	}
	fmt.Fprintln(stderr, usage)
	return 1
}
