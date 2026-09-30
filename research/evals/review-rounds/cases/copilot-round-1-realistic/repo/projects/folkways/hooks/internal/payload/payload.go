// Package payload parses the JSON a hook receives on stdin from Claude Code or Copilot CLI.
package payload

import (
	"encoding/json"
	"io"
	"strings"
)

// Payload covers Claude Code and Copilot CLI PascalCase (tool_name, tool_input) and camelCase (toolName, toolArgs).
type Payload struct {
	ToolName  string          `json:"tool_name"`
	ToolNameC string          `json:"toolName"`
	ToolInput json.RawMessage `json:"tool_input"`
	ToolArgs  json.RawMessage `json:"toolArgs"`
	Cwd       string          `json:"cwd"`
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
