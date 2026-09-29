// Bench prototype (not a production hook): git-guard for the black-box contract suite in test/hooks.
// Usage: folkways-hook git-guard < payload.json. Always exits 0 (a crash would make Copilot CLI deny the tool).
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// payload covers Claude Code / Copilot PascalCase (tool_name, tool_input) and camelCase (toolName, toolArgs).
type payload struct {
	ToolName  string          `json:"tool_name"`
	ToolNameC string          `json:"toolName"`
	ToolInput json.RawMessage `json:"tool_input"`
	ToolArgs  json.RawMessage `json:"toolArgs"`
	Cwd       string          `json:"cwd"`
}

// bashCommand returns the Bash command of a payload, or "" for other tools.
func (p payload) bashCommand() string {
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

// deny prints the one JSON object both tools understand: the nested form is Claude Code's, the top-level Copilot's.
func deny(reason string) {
	type decision struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(struct {
		HookSpecificOutput       decision `json:"hookSpecificOutput"`
		PermissionDecision       string   `json:"permissionDecision"`
		PermissionDecisionReason string   `json:"permissionDecisionReason"`
	}{decision{"PreToolUse", "deny", reason}, "deny", reason})
}

func main() {
	// Fail open on a bug: a panic exits 2, and Copilot CLI denies the tool call on any nonzero preToolUse exit.
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "folkways: internal error:", r)
			os.Exit(0)
		}
	}()
	if len(os.Args) < 2 || os.Args[1] != "git-guard" {
		return
	}
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	var p payload
	if err := json.Unmarshal(in, &p); err != nil {
		fmt.Fprintln(os.Stderr, "folkways:", err)
		return
	}
	cmd := p.bashCommand()
	if cmd == "" {
		return
	}
	cfg := loadConfig()
	if reason := checkCommand(cmd, p.Cwd, cfg); reason != "" {
		deny(reason)
	}
}
