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
