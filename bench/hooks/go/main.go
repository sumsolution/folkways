// Bench prototype (not a production hook). Spec: ../README.md#the-bench-hook
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	commit = regexp.MustCompile(`\bgit\s+commit\b[^|;&]*?\s-m\s*(?:"([^"]*)"|'([^']*)')`)
	header = regexp.MustCompile(`^(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(([a-z0-9._-]+)\))?!?: \S`)
)

type payload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type team struct {
	Git struct {
		Scopes []string `yaml:"scopes"`
	} `yaml:"git"`
}

type decision struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

func deny(reason string) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(struct {
		HookSpecificOutput decision `json:"hookSpecificOutput"`
	}{decision{"PreToolUse", "deny", reason}})
}

func run() error {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var p payload
	if err := json.Unmarshal(in, &p); err != nil {
		return err
	}
	if p.ToolName != "Bash" {
		return nil
	}
	m := commit.FindStringSubmatch(p.ToolInput.Command)
	if m == nil {
		return nil
	}
	raw, err := os.ReadFile("team.yaml")
	if err != nil {
		return err
	}
	var t team
	if err := yaml.Unmarshal(raw, &t); err != nil {
		return err
	}
	subject, _, _ := strings.Cut(m[1]+m[2], "\n")
	h := header.FindStringSubmatch(subject)
	if h == nil {
		return deny(fmt.Sprintf("Commit message is not Angular \"type(scope): subject\": \"%s\".", subject))
	}
	scopes := t.Git.Scopes
	for _, s := range scopes {
		if s == h[1] {
			return nil
		}
	}
	if len(scopes) > 0 {
		return deny(fmt.Sprintf("Scope \"%s\" is not in team.yaml git.scopes (%s).", h[1], strings.Join(scopes, ", ")))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "folkways:", err)
		os.Exit(1)
	}
}
