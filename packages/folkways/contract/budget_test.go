package contract

import (
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

// TestHookBudget is the hook latency budget of ADR 0011: each hook is a process the agent waits for on every
// tool call. Skipped unless FOLKWAYS_HOOK_BUDGET_MS is set (CI sets it in the hooks job). It runs a
// representative call of each hook through `sh -c`, as the tools do, and fails when the p50 exceeds the budget.
func TestHookBudget(t *testing.T) {
	raw := os.Getenv("FOLKWAYS_HOOK_BUDGET_MS")
	if raw == "" {
		t.Skip("set FOLKWAYS_HOOK_BUDGET_MS to run")
	}
	budget, err := strconv.ParseFloat(raw, 64)
	if err != nil || budget <= 0 {
		t.Fatalf("FOLKWAYS_HOOK_BUDGET_MS=%q is not a positive number of milliseconds", raw)
	}
	const runs, warmup = 50, 3

	base := plain{"session_id": "11111111-1111-4111-8111-111111111111", "transcript_path": "${TRANSCRIPT}", "cwd": "${TEAM}", "permission_mode": "default"}
	payload := func(extra plain) plain { return merge(base, extra) }
	calls := []struct {
		name, hook string
		payload    plain
	}{
		{"git-guard commit check", "git-guard", payload(plain{"hook_event_name": "PreToolUse", "tool_name": "Bash", "tool_input": plain{"command": `git commit -m "feat(auth): add login"`}})},
		{"session-start", "session-start", payload(plain{"hook_event_name": "SessionStart", "source": "startup"})},
		{"context-band", "context-band", payload(plain{"hook_event_name": "UserPromptSubmit", "prompt": "next step"})},
		{"precompact", "precompact", payload(plain{"hook_event_name": "PreCompact", "trigger": "auto"})},
	}
	fx := Fixture{}
	fx.Env.Transcript = []any{130000.0}
	vars := buildEnv(t, fx)

	for _, c := range calls {
		t.Run(c.name, func(t *testing.T) {
			step := Step{Payload: c.payload}
			durations := make([]float64, 0, runs)
			for i := 0; i < warmup+runs; i++ {
				start := time.Now()
				r := runHook(t, hookCmd, c.hook, step, vars, true)
				elapsed := time.Since(start)
				if r.Exit != 0 {
					t.Fatalf("exit %d (stderr: %s)", r.Exit, r.Stderr)
				}
				if i >= warmup {
					durations = append(durations, float64(elapsed)/float64(time.Millisecond))
				}
			}
			sort.Float64s(durations)
			p50 := (durations[runs/2-1] + durations[runs/2]) / 2
			t.Logf("p50 %.1f ms, p95 %.1f ms, budget %.1f ms (%d runs)", p50, durations[runs*95/100], budget, runs)
			if p50 > budget {
				t.Errorf("p50 %.1f ms exceeds the %.1f ms budget", p50, budget)
			}
		})
	}
}
