// todo_triage_model_free_test.go — SPEC-JEV-GOAL-DIST-001 M8c (REQ-JEVG-014,
// the triage half of AC-JEVG-012's method). `moai todo triage` stays
// model-free: no requirement in this chain routes a model answer into it, and
// the premise-death rejection recorded in docs/jev-negative-results.md makes
// that a settled decision rather than an unmade one. The scan carries its own
// positive control, because a zero-hit and a broken search are
// indistinguishable without one.
package cli

import (
	"os"
	"strings"
	"testing"
)

func TestTodoTriageStaysModelFree(t *testing.T) {
	body, err := os.ReadFile("todo_triage.go")
	if err != nil {
		t.Fatalf("read todo_triage.go: %v", err)
	}
	if strings.Contains(strings.ToLower(string(body)), "jev") {
		t.Error("todo_triage.go references the judgment capability — moai todo triage stays model-free (REQ-JEVG-014)")
	}

	// SPEC-JEV-GUARD-001 (t1083) withdrew jev_skill_suggest.go, the former
	// positive control; mcp_jev.go is the live MCP wrapper that still carries
	// the capability, so the control keeps proving the scan fires.
	control, err := os.ReadFile("mcp_jev.go")
	if err != nil {
		t.Fatalf("read positive control: %v", err)
	}
	if !strings.Contains(strings.ToLower(string(control)), "jev") {
		t.Fatal("positive control failed: mcp_jev.go no longer references the capability, so the zero-result above establishes nothing")
	}
}
