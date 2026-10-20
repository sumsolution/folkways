// Package hookrun runs one folkways hook by name. `folkways hook <name>` and `folkways-hook <name>` both
// call it, so the two binaries behave the same until folkways-hook goes (ADR 0011).
package hookrun

import (
	"fmt"
	"io"
	"os"

	"github.com/sumsolution/folkways/packages/folkways/internal/contextband"
	"github.com/sumsolution/folkways/packages/folkways/internal/gitguard"
	"github.com/sumsolution/folkways/packages/folkways/internal/payload"
	"github.com/sumsolution/folkways/packages/folkways/internal/precompact"
	"github.com/sumsolution/folkways/packages/folkways/internal/sessionstart"
	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// hooks maps a hook name to its implementation.
var hooks = map[string]func(p payload.Payload, cfg team.Config){
	"git-guard":     gitguard.Run,
	"context-band":  contextband.Run,
	"precompact":    precompact.Run,
	"session-start": sessionstart.Run,
}

// Run runs the named hook on the payload in r. It never panics and never exits: the caller exits 0,
// because Copilot CLI denies the tool call when a preToolUse hook exits nonzero or crashes.
// An unknown name does nothing.
func Run(name string, r io.Reader) {
	defer func() {
		if v := recover(); v != nil {
			fmt.Fprintln(os.Stderr, "folkways: internal error:", v)
		}
	}()
	run, ok := hooks[name]
	if !ok {
		return
	}
	p, err := payload.Read(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "folkways:", err)
		return
	}
	cfg := team.Load()
	// No folkways.yaml: nothing runs, except session-start, which makes the one-time offer.
	if !cfg.Configured && name != "session-start" {
		return
	}
	if cfg.Configured && !cfg.HookEnabled(name) {
		return
	}
	run(p, cfg)
}
