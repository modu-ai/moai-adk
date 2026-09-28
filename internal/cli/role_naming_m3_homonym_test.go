package cli

// role_naming_m3_homonym_test.go — SPEC-ROLE-NAMING-CODE-001 M3, AC-RNC-023:
// the pre-flight population's two rows (the `moai cg` mixed-backend refusal
// in factory.go and kanban.go) carry their CG qualifier next to each role
// noun. Qualify, never rename — the sentinels and the surrounding sentence
// stay.

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

	kanbanErr := rejectKanbanOnCG([]string{"-k"})
	if kanbanErr == nil {
		t.Fatal("rejectKanbanOnCG(-k) returned nil, want the mixed-backend refusal")
	}
	if !strings.Contains(kanbanErr.Error(), "(CG leader Claude, CG teammates GLM)") {
		t.Errorf("kanban refusal lacks the CG qualifiers: %v", kanbanErr)
	}
}
