// Package output writes the JSON objects hooks print on stdout.
package output

import (
	"encoding/json"
	"io"
	"os"
)

// Stdout is where hooks print; tests replace it.
var Stdout io.Writer = os.Stdout

// Deny denies a tool call with one JSON object both tools understand:
// the nested form is Claude Code's, the top-level form Copilot CLI's.
func Deny(reason string) {
	type decision struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	}
	enc := json.NewEncoder(Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(struct {
		HookSpecificOutput       decision `json:"hookSpecificOutput"`
		PermissionDecision       string   `json:"permissionDecision"`
		PermissionDecisionReason string   `json:"permissionDecisionReason"`
	}{decision{"PreToolUse", "deny", reason}, "deny", reason})
}
