package hookrun

import (
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/folkways/internal/payload"
	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// A panicking hook must not escape Run: the binary would exit 2 and Copilot CLI would deny the tool call.
func TestRunRecoversFromPanic(t *testing.T) {
	t.Chdir(t.TempDir()) // no folkways.yaml: only session-start runs
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	orig := hooks["session-start"]
	hooks["session-start"] = func(payload.Payload, team.Config) { panic("boom") }
	defer func() { hooks["session-start"] = orig }()

	called := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Run let a panic escape: %v", r)
			}
		}()
		Run("session-start", "dev", strings.NewReader(`{"hook_event_name":"SessionStart"}`))
		called = true
	}()
	if !called {
		t.Fatal("Run did not return")
	}
}
