package kanban

// role_naming_m2_test.go — SPEC-ROLE-NAMING-CODE-001 M2: the lane-label
// parser admits ONLY the canonical `lane-<n>` shape (AC-RNC-004). The legacy
// `worker-<n>` / `agent-<n>` shapes and the degenerate forms are not-a-lane.

import "testing"

func TestSplitFactoryLaneLabelAdmitsLaneShapesOnly(t *testing.T) {
	t.Parallel()

	n, ok := SplitFactoryLaneLabel("lane-3")
	if !ok || n != 3 {
		t.Fatalf("SplitFactoryLaneLabel(lane-3) = (%d, %v), want (3, true)", n, ok)
	}

	for _, rejected := range []string{
		"worker-3", // legacy prefix — detection only
		"agent-3",  // legacy prefix — detection only
		"lane-0",   // n must be >= 1
		"lane-",    // no number
		"lane-a",   // non-numeric suffix
		"lane-3-x", // second hyphen never parses
		"lane",     // bare prefix is the role value, not a label
	} {
		if _, ok := SplitFactoryLaneLabel(rejected); ok {
			t.Errorf("SplitFactoryLaneLabel(%q) = ok, want not-a-lane-label", rejected)
		}
	}
}

func TestLegacyFactoryValuesAreDetectedWithoutBecomingLanes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		value       string
		legacyLabel bool
		legacyAgent bool
		legacyValue bool
		wantNumber  int
	}{
		{value: "worker", legacyValue: true},
		{value: "agent", legacyValue: true},
		{value: "worker-3", legacyLabel: true, legacyValue: true, wantNumber: 3},
		{value: "agent-4", legacyLabel: true, legacyAgent: true, legacyValue: true, wantNumber: 4},
		{value: "lane-3"},
		{value: "agent-0"},
		{value: "agent-3-extra"},
	} {
		if n, ok := SplitFactoryLegacyLabel(tc.value); ok != tc.legacyLabel || (ok && n != tc.wantNumber) {
			t.Errorf("SplitFactoryLegacyLabel(%q) = (%d, %v), want (%d, %v)", tc.value, n, ok, tc.wantNumber, tc.legacyLabel)
		}
		if n, ok := SplitFactoryLegacyAgentLabel(tc.value); ok != tc.legacyAgent || (ok && n != tc.wantNumber) {
			t.Errorf("SplitFactoryLegacyAgentLabel(%q) = (%d, %v), want (%d, %v)", tc.value, n, ok, tc.wantNumber, tc.legacyAgent)
		}
		if got := IsLegacyFactoryRoleValue(tc.value); got != tc.legacyValue {
			t.Errorf("IsLegacyFactoryRoleValue(%q) = %v, want %v", tc.value, got, tc.legacyValue)
		}
		if _, ok := SplitFactoryLaneLabel(tc.value); ok != (tc.value == "lane-3") {
			t.Errorf("SplitFactoryLaneLabel(%q) = %v, want canonical lane only", tc.value, ok)
		}
	}
}
