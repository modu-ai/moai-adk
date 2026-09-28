package kanban

// factory_label_pin_test.go — the REQ-AP-013 role-vocabulary pin,
// label-prefix side (SPEC-AUTONOMY-PRECONDITION-001 M2, AC-AP-018).
//
// The kanban carrier is the lane-label prefix FactoryLaneLabel emits; the
// CLI role token alone would not catch its drift, which is why REQ-AP-013
// names two carriers and one assertion per carrier package.
//
// SPEC-ROLE-NAMING-CODE-001 M4: the guard constant is FactoryRoleLane =
// `lane`, so the full three-way equality (marker value == CLI token ==
// label prefix) is restored and asserted here through the constant — the
// M1 interim tripwire is gone (REQ-RNC-012). The legacy prefixes
// (`worker`, `agent`) stay detection-only.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryLabelPrefixPinsGuardConstant pins the kanban limb of AC-AP-018:
// FactoryLaneLabel composes the canonical `lane-<n>` prefix, that prefix
// equals the guard's role-value constant (REQ-RNC-012's three-way equality,
// label-prefix carrier), and it is neither legacy prefix.
func TestFactoryLabelPrefixPinsGuardConstant(t *testing.T) {
	prefix, _, _ := strings.Cut(FactoryLaneLabel(1), "-")
	if prefix != config.FactoryRoleLane {
		t.Fatalf("factory lane label prefix %q != guard value constant %q — the REQ-AP-013 equality (marker value == lane-label prefix, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-012) regressed",
			prefix, config.FactoryRoleLane)
	}
	if prefix != "lane" {
		t.Fatalf("factory lane label prefix %q != %q — the M1 vocabulary swap regressed (AC-AP-018, SPEC-ROLE-NAMING-CODE-001)",
			prefix, "lane")
	}
	if prefix == factoryLegacyWorkerRole || prefix == factoryLegacyAgentRole {
		t.Fatalf("lane label prefix %q equals a legacy label prefix (%q / %q) — swapping live and legacy spellings must not pass (AC-AP-018)",
			prefix, factoryLegacyWorkerRole, factoryLegacyAgentRole)
	}
}
