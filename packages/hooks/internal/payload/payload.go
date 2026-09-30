// Package payload parses the JSON a hook receives on stdin from Claude Code or Copilot CLI.
package payload

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

// Payload covers Claude Code and Copilot CLI PascalCase (tool_name, tool_input) and camelCase (toolName, toolArgs).
type Payload struct {
	ToolName  string          `json:"tool_name"`
	ToolNameC string          `json:"toolName"`
	ToolInput json.RawMessage `json:"tool_input"`
	ToolArgs  json.RawMessage `json:"toolArgs"`
	Cwd       string          `json:"cwd"`

	SessionIDS     string `json:"session_id"`
	SessionIDC     string `json:"sessionId"`
	HookEventName  string `json:"hook_event_name"`
	TranscriptPath string `json:"transcript_path"`
	Source         string `json:"source"`  // SessionStart: startup, resume, compact, new, ...
	Trigger        string `json:"trigger"` // PreCompact: auto or manual
}

var safeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// SessionID returns the session id, or "" when there is none or it is not safe as a file name.
func (p Payload) SessionID() string {
	id := p.SessionIDS
	if id == "" {
		id = p.SessionIDC
	}
	if !safeID.MatchString(id) {
		return ""
	}
	return id
}

// Read parses a payload from r.
func Read(r io.Reader) (Payload, error) {
	var p Payload
	in, err := io.ReadAll(r)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(in, &p)
	return p, err
}

// BashCommand returns the Bash command of a payload, or "" for other tools.
func (p Payload) BashCommand() string {
	name := p.ToolName
	if name == "" {
		name = p.ToolNameC
	}
	if !strings.EqualFold(name, "bash") {
		return ""
	}
	raw := p.ToolInput
	if len(raw) == 0 {
		raw = p.ToolArgs
	}
	var s string
	if json.Unmarshal(raw, &s) == nil { // arguments arrive as a JSON string in some shapes
		raw = json.RawMessage(s)
	}
	var args struct {
		Command string `json:"command"`
	}
	_ = json.Unmarshal(raw, &args)
	return args.Command
}
