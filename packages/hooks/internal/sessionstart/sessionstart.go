// Package sessionstart implements the session-start hook: it reminds the start procedure.
package sessionstart

import (
	"github.com/sumsolution/folkways/packages/hooks/internal/output"
	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/skill"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

// Run injects the start reminder, plus a 00-STATUS.md update after a compaction.
func Run(p payload.Payload, _ team.Config) {
	text := "folkways: follow the `" + skill.Name("start") + "` skill for this session."
	if p.Source == "compact" {
		text += " The context was just compacted: update the effort's `00-STATUS.md` now, before anything else."
	}
	output.Context("SessionStart", text)
}
