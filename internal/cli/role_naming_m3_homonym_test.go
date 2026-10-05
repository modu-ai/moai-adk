package cli

// role_naming_m3_homonym_test.go — SPEC-ROLE-NAMING-CODE-001 M3, AC-RNC-023:
// the pre-flight population's row (the `moai cg` mixed-backend refusal in
// factory.go) carries its CG qualifier next to each role noun. Qualify, never
// rename — the sentinel and the surrounding sentence stay. The second row, in
// the removed kanban helper, left with it (SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a).

import (
	"strings"
	"testing"
)

func TestCGMixedBackendRefusalsCarryQualifier(t *testing.T) {
	factoryErr := rejectFactoryOnCG([]string{"-f"})
	if factoryErr == nil {
		t.Fatal("rejectFactoryOnCG(-f 2) returned nil, want the mixed-backend refusal")
	}
	if !strings.Contains(factoryErr.Error(), "(CG leader Claude, CG teammates GLM)") {
		t.Errorf("factory refusal lacks the CG qualifiers: %v", factoryErr)
	}
}
