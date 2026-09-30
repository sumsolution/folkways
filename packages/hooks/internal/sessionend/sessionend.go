// Package sessionend implements the session-end hook: it marks the session as not yet closed.
package sessionend

import (
	"os"
	"path/filepath"

	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

// Run writes .local/sessions/<session id>.unclosed in the team repo. The session-close skill clears it;
// session-start offers session-close while one is left.
func Run(p payload.Payload, cfg team.Config) {
	id := p.SessionID()
	if id == "" {
		return
	}
	dir := filepath.Join(cfg.Dir, ".local", "sessions")
	if os.MkdirAll(dir, 0o755) == nil {
		_ = os.WriteFile(filepath.Join(dir, id+".unclosed"), nil, 0o644)
	}
}
