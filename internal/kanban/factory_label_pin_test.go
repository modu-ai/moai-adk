package kanban

// factory_label_pin_test.go — the REQ-AP-013 role-vocabulary pin,
// label-prefix side (SPEC-AUTONOMY-PRECONDITION-001 M2, AC-AP-018).
//
// The kanban carrier is the lane-label prefix FactoryLaneLabel emits; the
// CLI role token alone would not catch its drift, which is why REQ-AP-013
// names two carriers and one assertion per carrier package.
//
// SPEC-ROLE-NAMING-CODE-001 M1 state: the label prefix is `lane` and the
// legacy prefixes (`worker`, `agent`) are detection-only. The third limb of
// the original pin — prefix == config.FactoryRoleWorker — is M4's
// (SPEC-ROLE-NAMING-CODE-001): M4 flips the guard constant from `worker` to
// `lane`, at which point the full three-way equality (marker value == CLI
// token == label prefix) is restored and re-asserted here. Until then the
// constant still reads `worker`, which is now a legacy value, so the equality
// limb is deliberately not asserted in this interim state.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryLabelPrefixPinsGuardConstant pins the kanban limb of AC-AP-018
// for the M1 state: FactoryLaneLabel composes the canonical `lane-<n>` prefix,
// which is neither legacy prefix.
func TestFactoryLabelPrefixPinsGuardConstant(t *testing.T) {
	prefix, _, _ := strings.Cut(FactoryLaneLabel(1), "-")
	if prefix != "lane" {
		t.Fatalf("factory lane label prefix %q != %q — the M1 vocabulary swap regressed (AC-AP-018, SPEC-ROLE-NAMING-CODE-001)",
			prefix, "lane")
	}
	if prefix == factoryLegacyWorkerRole || prefix == factoryLegacyAgentRole {
		t.Fatalf("lane label prefix %q equals a legacy label prefix (%q / %q) — swapping live and legacy spellings must not pass (AC-AP-018)",
			prefix, factoryLegacyWorkerRole, factoryLegacyAgentRole)
	}
	// M4 tripwire: when config.FactoryRoleWorker flips to `lane`, restore the
	// original equality limb (prefix == config.FactoryRoleWorker) per
	// SPEC-AUTONOMY-PRECONDITION-001 AC-AP-018.
	if config.FactoryRoleWorker == prefix {
		t.Fatalf("guard constant %q now equals the lane prefix — M4 has flipped the constant; restore the AC-AP-018 equality assertion in this test",
			config.FactoryRoleWorker)
	}
}
