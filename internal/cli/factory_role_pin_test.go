package cli

// factory_role_pin_test.go — the REQ-AP-013 role-vocabulary pin, CLI-token
// side (SPEC-AUTONOMY-PRECONDITION-001 M2, AC-AP-018).
//
// The assertion lives INSIDE the carrier's own package because deriving the
// guard's value from this constant is unavailable: internal/cli imports
// internal/config, so internal/config cannot import internal/cli back
// without a cycle, and both carriers are unexported besides (spec.md
// REQ-AP-013). The assertion guarantees the failure — a one-sided rename
// turns this test RED — not the derivation.
//
// SPEC-ROLE-NAMING-CODE-001 M4: the full three-way equality (guard constant
// == -f role token == lane-label prefix) is restored — the M1 tripwire is
// gone (REQ-RNC-012).

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryRoleTokenPinsGuardConstant pins AC-AP-018 (CLI limb): the live
// `-f` role token this package parses equals the guard's role-value constant
// (REQ-RNC-012's three-way equality, CLI carrier), the token is the canonical
// `lane`, and the retired spellings are not accepted anywhere on the token
// path — a rename that merely swaps the live and legacy spellings must not
// pass.
func TestFactoryRoleTokenPinsGuardConstant(t *testing.T) {
	if factoryLaneRoleToken != config.FactoryRoleLane {
		t.Fatalf("-f role token %q != guard value constant %q — the REQ-AP-013 equality (marker value == -f token, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-012) regressed",
			factoryLaneRoleToken, config.FactoryRoleLane)
	}
	if factoryLaneRoleToken != "lane" {
		t.Fatalf("factory role token %q != %q — the canonical token must stay lane (AC-AP-018)",
			factoryLaneRoleToken, "lane")
	}
	for _, legacy := range []string{"worker", "agent"} {
		p, err := parseFactoryFlag([]string{"-f", legacy})
		if err == nil {
			t.Fatalf("-f %s parsed as %+v, want the legacy-token refusal (AC-AP-018)", legacy, p)
		}
	}
}
