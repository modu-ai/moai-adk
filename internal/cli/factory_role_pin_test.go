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

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryRoleTokenPinsGuardConstant pins AC-AP-018 (CLI limb): the
// guard's role-value constant equals the live `-f` role token this package
// parses, and equals no retired alias — the rename that merely swaps the live
// and legacy spellings must not pass.
func TestFactoryRoleTokenPinsGuardConstant(t *testing.T) {
	if factoryWorkerRoleToken != config.FactoryRoleWorker {
		t.Fatalf("factory role token %q != guard constant %q — a one-sided rename must turn this test RED (AC-AP-018)",
			factoryWorkerRoleToken, config.FactoryRoleWorker)
	}
	if config.FactoryRoleWorker == factoryLegacyAgentRoleToken {
		t.Fatalf("guard constant %q equals the retired alias %q — swapping the live and legacy spellings must not pass (AC-AP-018)",
			config.FactoryRoleWorker, factoryLegacyAgentRoleToken)
	}
}
