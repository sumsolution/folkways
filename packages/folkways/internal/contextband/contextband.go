// Package contextband implements the context-band hook: it announces each handoff band the context crosses.
package contextband

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/sumsolution/folkways/packages/folkways/internal/output"
	"github.com/sumsolution/folkways/packages/folkways/internal/payload"
	"github.com/sumsolution/folkways/packages/folkways/internal/skill"
	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// What to do at the first band and a middle band (instructions.md, "Context and handoff"). The last band's
// text names a skill, so advice builds it.
const (
	firstText = "Look for a stopping point."
	midText   = "Start no new multi-step work. Propose a handoff at the next checkpoint."
)

// advice returns what to do at band i of count.
func advice(i, count int) string {
	switch i {
	case count - 1:
		return "Write `00-STATUS.md` (`" + skill.Name("handoff") + "` skill) and end the session."
	case 0:
		return firstText
	}
	return midText
}

type state struct {
	Band int `json:"band"` // the threshold last announced, 0 for none
}

// Run reads the tokens in context from the transcript and, when the band rises above the one last announced
// for this session, injects what to do. Without a transcript (Copilot CLI) it does nothing.
func Run(p payload.Payload, cfg team.Config) {
	id := p.SessionID()
	if id == "" || p.TranscriptPath == "" || len(cfg.Team.Handoff.Bands) == 0 {
		return
	}
	f, err := os.Open(p.TranscriptPath)
	if err != nil {
		return
	}
	n := tokens(f)
	f.Close()
	if n == 0 {
		return
	}
	path := filepath.Join(cfg.Dir, ".local", "hooks", "context-band", id+".json")
	var last state
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &last)
	}
	cur, i := band(cfg.Team.Handoff.Bands, n)
	if cur == last.Band {
		return
	}
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	raw, _ := json.Marshal(state{cur})
	if os.WriteFile(path, raw, 0o644) != nil || cur < last.Band {
		return // fell below the last band (compaction): the next crossing is announced again
	}
	text := advice(i, len(cfg.Team.Handoff.Bands))
	event := p.HookEventName
	if event == "" {
		event = "UserPromptSubmit"
	}
	output.Context(event, fmt.Sprintf("folkways: %dK tokens in context, past the %s band. %s", n/1000, label(cur), text))
}

// tokens returns the tokens in context at the last main-chain assistant line with non-zero usage, 0 for none.
func tokens(r io.Reader) int {
	last := 0
	br := bufio.NewReader(r)
	for {
		raw, err := br.ReadBytes('\n') // no line length limit: tool results can be large
		if bytes.Contains(raw, []byte(`"assistant"`)) {
			var line struct {
				Type        string `json:"type"`
				IsSidechain bool   `json:"isSidechain"`
				Message     struct {
					Usage struct {
						Input    int `json:"input_tokens"`
						Read     int `json:"cache_read_input_tokens"`
						Creation int `json:"cache_creation_input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if json.Unmarshal(raw, &line) == nil && line.Type == "assistant" && !line.IsSidechain {
				if u := line.Message.Usage; u.Input+u.Read+u.Creation > 0 {
					last = u.Input + u.Read + u.Creation
				}
			}
		}
		if err != nil {
			return last
		}
	}
}

// band returns the highest threshold at or below n and its index among the sorted thresholds; 0, -1 for none.
func band(bands []int, n int) (int, int) {
	sorted := slices.Sorted(slices.Values(bands))
	for i := len(sorted) - 1; i >= 0; i-- {
		if sorted[i] > 0 && sorted[i] <= n {
			return sorted[i], i
		}
	}
	return 0, -1
}

// label writes a threshold in thousands: 125000 is "125K".
func label(n int) string {
	return strconv.FormatFloat(float64(n)/1000, 'f', -1, 64) + "K"
}
