package sessionstart

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumsolution/folkways/packages/hooks/internal/output"
	"github.com/sumsolution/folkways/packages/hooks/internal/payload"
	"github.com/sumsolution/folkways/packages/hooks/internal/team"
)

func run(t *testing.T) string {
	t.Helper()
	var out bytes.Buffer
	prev := output.Stdout
	output.Stdout = &out
	t.Cleanup(func() { output.Stdout = prev })
	Run(payload.Payload{}, team.Config{})
	return out.String()
}

func TestRunNamesTheStartSkill(t *testing.T) {
	t.Setenv("FOLKWAYS_SKILL_NAMESPACE", "")
	if got := run(t); !strings.Contains(got, "follow the `start` skill") {
		t.Errorf("session-start printed %q", got)
	}
}

func TestRunNamesTheStartSkillInANamespace(t *testing.T) {
	t.Setenv("FOLKWAYS_SKILL_NAMESPACE", "folkways")
	if got := run(t); !strings.Contains(got, "follow the `folkways:start` skill") {
		t.Errorf("session-start printed %q", got)
	}
}
