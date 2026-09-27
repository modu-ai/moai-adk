package kanban

import "testing"

// TestFactoryLaneLabelJoinSplitRoundTrips covers the label vocabulary:
// joining a number and splitting the result must be inverses.
func TestFactoryLaneLabelJoinSplitRoundTrips(t *testing.T) {
	t.Parallel()

	for _, n := range []int{1, 2, 9, 10, 42, 100} {
		label := FactoryLaneLabel(n)
		got, ok := SplitFactoryLaneLabel(label)
		if !ok || got != n {
			t.Errorf("round trip %d -> %q -> (%d, %v)", n, label, got, ok)
		}
	}
}

// TestSplitFactoryLaneLabelRejectsNonShapes asserts the shape IS the
// discriminator: anything that is not `worker-<n>` or the legacy `lane-<n>`
// (n >= 1) reads as "not a worker", never as an error.
func TestSplitFactoryLaneLabelRejectsNonShapes(t *testing.T) {
	t.Parallel()

	for _, label := range []string{
		"", "lane", "lane-", "lane-0", "lane--3", "lane-a",
		"lane-3-extra", "Lane-3", "lanes-3",
		"plan-lane-3", // first hyphen is the boundary; role is "plan"
		"agent-3",     // the legacy agent shape has its own parser
		"worker--3", "worker-3-extra",
	} {
		if _, ok := SplitFactoryLaneLabel(label); ok {
			t.Errorf("SplitFactoryLaneLabel(%q) admitted a non-lane shape", label)
		}
	}
}

// TestFactoryLaneLabelNeverKanbanShape is the no-cross-talk property: a
// factory lane label satisfies neither kanban discriminator, and no kanban
// label satisfies the lane one — the three branch selectors can never both
// match the same name.
func TestFactoryLaneLabelNeverKanbanShape(t *testing.T) {
	t.Parallel()

	if _, _, ok := SplitCompanionLabel(FactoryLaneLabel(3)); ok {
		t.Error("a lane label must not satisfy the companion shape")
	}
	if _, ok := SplitLeadLabel(FactoryLaneLabel(3)); ok {
		t.Error("a lane label must not satisfy the lead shape")
	}
	if _, ok := SplitFactoryLaneLabel(CompanionLabel("run")); ok {
		t.Error("a companion label must not satisfy the lane shape")
	}
	if _, ok := SplitFactoryLaneLabel(CompanionNumberLabel("run", 1)); ok {
		t.Error("a bumped companion label must not satisfy the lane shape")
	}
	if _, ok := SplitFactoryLaneLabel(LeadLabel()); ok {
		t.Error("a lead label must not satisfy the lane shape")
	}
}

// TestWithRoleAdmitsLane asserts the record role set admits the factory
// lane role while still discarding arbitrary strings.
func TestWithRoleAdmitsLane(t *testing.T) {
	t.Parallel()

	rec := NewRecord("sess-1", "", BackendClaude).WithRole(RoleLane)
	if rec.Role != RoleLane {
		t.Errorf("WithRole(lane) = %q, want %q", rec.Role, RoleLane)
	}
	if got := NewRecord("sess-2", "", BackendClaude).WithRole("worker"); got.Role != "" {
		t.Errorf("WithRole(worker) = %q, want the role dropped", got.Role)
	}
	if got := NewRecord("sess-3", "", BackendClaude).WithRole("foreman"); got.Role != "" {
		t.Errorf("WithRole(foreman) = %q, want the role dropped", got.Role)
	}
}
