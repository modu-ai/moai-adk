// factory_m4_test.go — SPEC-FACTORY-SELF-DISPATCH-001 M4, AC-SD-017 hook arm
// (card t1240): the contract sign guard's role gate widens in the deny
// direction to the lane-refusal predicate (REQ-SD-017 widening REQ-AP-011) —
// the same three clauses the cli side evaluates (marker set, lane label
// non-empty, backend naming the Codex harness) — while an environment
// carrying none of the three stays allowed exactly as before.
package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestSD_AC017_WidenedRoleGateDenyAndAllow — the widened role gate denies a
// lane-label-only environment and a Codex MCP environment, and allows an
// environment carrying none of the three lane variables.
func TestSD_AC017_WidenedRoleGateDenyAndAllow(t *testing.T) {
	const call = "moai contract sign SPEC-X --signer llm"
	assertDeny := func(where, decision, reason string) {
		t.Helper()
		if decision != DecisionDeny {
			t.Fatalf("%s: decision = %q (reason %q), want deny", where, decision, reason)
		}
		if !strings.Contains(reason, "CONTRACT_SIGN_AGENT_VIOLATION:") {
			t.Errorf("%s: deny reason = %q, want the CONTRACT_SIGN_AGENT_VIOLATION: sentinel", where, reason)
		}
	}

	// No marker anywhere in this test: every arm below runs marker-free, so
	// only the widened clauses can produce the denies.
	t.Setenv(config.EnvFactoryRole, "")

	// Label-only environment: the marker is unset, the lane label is set.
	t.Setenv(config.EnvMoaiKanbanLabel, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, "")
	decision, reason := checkContractSign(signGuardInput(t, call))
	assertDeny("label-only", decision, reason)

	// Codex MCP environment: exactly the variables the frozen Codex MCP
	// env_vars allowlist forwards for a lane — a lane label and
	// MOAI_KANBAN_BACKEND=gpt, no role marker. The deny here rides the
	// backend clause.
	t.Setenv(config.EnvMoaiKanbanLabel, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
	decision, reason = checkContractSign(signGuardInput(t, call))
	assertDeny("codex-mcp", decision, reason)

	// None of the three lane variables: allowed exactly as before — the allow
	// direction is a requirement (the leader's own decide path rides it).
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvMoaiKanbanBackend, "")
	decision, reason = checkContractSign(signGuardInput(t, call))
	if decision != "" || reason != "" {
		t.Errorf("no-lane-vars: decision = %q reason = %q, want the silent allow", decision, reason)
	}

	// The marker clause itself: the stamped marker denies through the same
	// gate (the launcher-side capture exercises it from the cli package; this
	// arm pins the clause inside the gate's own package).
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	decision, reason = checkContractSign(signGuardInput(t, call))
	assertDeny("marker", decision, reason)
}
