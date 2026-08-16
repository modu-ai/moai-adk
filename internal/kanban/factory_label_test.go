package kanban

import "testing"

// TestFactoryWorkerLabelJoinSplitRoundTrips covers the label vocabulary:
// joining a number and splitting the result must be inverses.
func TestFactoryWorkerLabelJoinSplitRoundTrips(t *testing.T) {
	t.Parallel()

	for _, n := range []int{1, 2, 9, 10, 42, 100} {
		label := FactoryWorkerLabel(n)
		got, ok := SplitFactoryWorkerLabel(label)
		if !ok || got != n {
			t.Errorf("round trip %d -> %q -> (%d, %v)", n, label, got, ok)
		}
	}
}

// TestSplitFactoryWorkerLabelRejectsNonShapes asserts the shape IS the
// discriminator: anything that is not `worker-<n>` (n >= 1) reads as "not a
// worker", never as an error.
func TestSplitFactoryWorkerLabelRejectsNonShapes(t *testing.T) {
	t.Parallel()

	for _, label := range []string{
		"", "worker", "worker-", "worker-0", "worker--3", "worker-a",
		"worker-3-extra", "Worker-3", "workers-3",
		"plan-worker-3", // first hyphen is the boundary; role is "plan"
	} {
		if _, ok := SplitFactoryWorkerLabel(label); ok {
			t.Errorf("SplitFactoryWorkerLabel(%q) admitted a non-worker shape", label)
		}
	}
}

// TestFactoryWorkerLabelNeverKanbanShape is the no-cross-talk property: a
// factory worker label satisfies neither kanban discriminator, and no kanban
// label satisfies the worker one — the three branch selectors can never both
// match the same name.
func TestFactoryWorkerLabelNeverKanbanShape(t *testing.T) {
	t.Parallel()

	if _, _, ok := SplitCompanionLabel(FactoryWorkerLabel(3)); ok {
		t.Error("a worker label must not satisfy the companion shape")
	}
	if _, ok := SplitLeadLabel(FactoryWorkerLabel(3)); ok {
		t.Error("a worker label must not satisfy the lead shape")
	}
	if _, ok := SplitFactoryWorkerLabel(CompanionLabel("run")); ok {
		t.Error("a companion label must not satisfy the worker shape")
	}
	if _, ok := SplitFactoryWorkerLabel(CompanionNumberLabel("run", 1)); ok {
		t.Error("a bumped companion label must not satisfy the worker shape")
	}
	if _, ok := SplitFactoryWorkerLabel(LeadLabel("tjlgt1")); ok {
		t.Error("a lead label must not satisfy the worker shape")
	}
}

// TestWithRoleAdmitsWorker asserts the record role set admits the factory
// worker role while still discarding arbitrary strings.
func TestWithRoleAdmitsWorker(t *testing.T) {
	t.Parallel()

	rec := NewRecord("sess-1", "", BackendClaude).WithRole(RoleWorker)
	if rec.Role != RoleWorker {
		t.Errorf("WithRole(worker) = %q, want %q", rec.Role, RoleWorker)
	}
	if got := NewRecord("sess-2", "", BackendClaude).WithRole("foreman"); got.Role != "" {
		t.Errorf("WithRole(foreman) = %q, want the role dropped", got.Role)
	}
}
