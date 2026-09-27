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
// M2 state (SPEC-ROLE-NAMING-CODE-001): the CLI parses exactly one role
// token, `lane`, and refuses the legacy spellings. The full three-way
// equality (guard constant == -f token == lane-label prefix) is restored at
// M4 when config.FactoryRoleWorker flips to "lane"; until then the guard
// constant is the M4 tripwire documented in progress.md §E.2 (the M1 note).

import (
	"testing"
)

// TestFactoryRoleTokenPinsGuardConstant pins AC-AP-018 (CLI limb): the live
// `-f` role token this package parses is the canonical `lane`, and the
// retired spellings are not accepted anywhere on the token path — a rename
// that merely swaps the live and legacy spellings must not pass.
func TestFactoryRoleTokenPinsGuardConstant(t *testing.T) {
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
