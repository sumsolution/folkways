// Command folkways-hook runs one folkways hook: folkways-hook <hook> < payload.json.
// It always exits 0: Copilot CLI denies the tool call when a preToolUse hook exits nonzero or crashes.
package main

import (
	"fmt"
	"os"

	"github.com/sumsolution/folkways/packages/hooks/internal/contextband"
	"github.com/sumsolution/folkways/packages/hooks/internal/gitguard"
	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/precompact"
	"github.com/sumsolution/folkways/packages/hooks/internal/sessionend"
	"github.com/sumsolution/folkways/packages/hooks/internal/sessionstart"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

// version is set at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

// hooks maps a hook name to its implementation.
var hooks = map[string]func(p payload.Payload, cfg team.Config){
	"git-guard":     gitguard.Run,
	"context-band":  contextband.Run,
	"precompact":    precompact.Run,
	"session-start": sessionstart.Run,
	"session-end":   sessionend.Run,
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "folkways: internal error:", r)
		}
	}()
	if len(os.Args) < 2 {
		return
	}
	if os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	run, ok := hooks[os.Args[1]]
	if !ok {
		return
	}
	p, err := payload.Read(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "folkways:", err)
		return
	}
	run(p, team.Load())
}
