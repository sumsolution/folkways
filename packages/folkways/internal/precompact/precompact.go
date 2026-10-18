// Package precompact implements the precompact hook: a compaction means a context band was missed.
package precompact

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sumsolution/folkways/packages/folkways/internal/payload"
	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// Run appends one JSON line (time, session id, trigger) to .local/precompact.log. It prints nothing, so it
// never blocks compaction.
func Run(p payload.Payload, cfg team.Config) {
	line, _ := json.Marshal(struct {
		Time      string `json:"time"`
		SessionID string `json:"session_id"`
		Trigger   string `json:"trigger"`
	}{time.Now().UTC().Format(time.RFC3339), p.SessionID(), p.Trigger})
	dir := filepath.Join(cfg.Dir, ".local")
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "precompact.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}
