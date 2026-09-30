// Package sessionstart implements the session-start hook: it reminds the start procedure and what is pending.
package sessionstart

import (
	"path/filepath"
	"strings"

	"github.com/sumsolution/folkways/packages/hooks/internal/output"
	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

// Run injects the start reminder, plus a 00-STATUS.md update after a compaction and a session-close offer
// for each unclosed-session marker.
func Run(p payload.Payload, cfg team.Config) {
	parts := []string{"folkways: follow the `start` skill for this session."}
	if p.Source == "compact" {
		parts = append(parts, "The context was just compacted: update the effort's `00-STATUS.md` now, before anything else.")
	}
	markers, _ := filepath.Glob(filepath.Join(cfg.Dir, ".local", "sessions", "*.unclosed"))
	if len(markers) > 0 {
		ids := make([]string, len(markers))
		for i, m := range markers {
			ids[i] = strings.TrimSuffix(filepath.Base(m), ".unclosed")
		}
		parts = append(parts, "Earlier sessions were not closed ("+strings.Join(ids, ", ")+
			", markers in `.local/sessions/`): offer to run the `session-close` skill for them.")
	}
	output.Context("SessionStart", strings.Join(parts, " "))
}
