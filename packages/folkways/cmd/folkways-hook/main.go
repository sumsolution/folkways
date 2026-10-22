// Command folkways-hook runs one folkways hook: folkways-hook <hook> < payload.json.
// It always exits 0: Copilot CLI denies the tool call when a preToolUse hook exits nonzero or crashes.
// It goes once the plugin and team repos run `folkways hook <name>` (ADR 0011).
package main

import (
	"fmt"
	"os"

	"github.com/sumsolution/folkways/packages/folkways/internal/hookrun"
)

// version is set at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		return
	}
	if os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	hookrun.Run(os.Args[1], version, os.Stdin)
}
