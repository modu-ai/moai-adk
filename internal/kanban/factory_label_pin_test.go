package kanban

// factory_label_pin_test.go — the REQ-AP-013 role-vocabulary pin,
// label-prefix side (SPEC-AUTONOMY-PRECONDITION-001 M2, AC-AP-018).
//
// The kanban carrier is the worker-label prefix FactoryLaneLabel emits; the
// CLI role token alone would not catch its drift, which is why REQ-AP-013
// names two carriers and one assertion per carrier package.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryLabelPrefixPinsGuardConstant pins AC-AP-018 (kanban limb): the
// guard's role-value constant equals the prefix of FactoryLaneLabel(1), and
// equals neither of the package's two retired label prefixes.
func TestFactoryLabelPrefixPinsGuardConstant(t *testing.T) {
	prefix, _, _ := strings.Cut(FactoryLaneLabel(1), "-")
	if prefix != config.FactoryRoleWorker {
		t.Fatalf("factory worker label prefix %q != guard constant %q — a one-sided rename must turn this test RED (AC-AP-018)",
			prefix, config.FactoryRoleWorker)
	}
	if config.FactoryRoleWorker == factoryLegacyLaneRole || config.FactoryRoleWorker == factoryLegacyAgentRole {
		t.Fatalf("guard constant %q equals a retired label prefix (%q / %q) — swapping live and legacy spellings must not pass (AC-AP-018)",
			config.FactoryRoleWorker, factoryLegacyLaneRole, factoryLegacyAgentRole)
	}
}
