package cli

// kanban_dispatch_test.go holds the SPEC-FACTORY-BOOTSTRAP-001 dispatch-truth-
// table tests. The four-row table at spec.md §A.2 is the semantic core of this
// SPEC; these tests pin each row against the resolveKanbanBranch decision
// function the cc.go / glm.go dispatch sites call.

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// clearAllKanbanEnv unsets EVERY kanban variable so each dispatch case starts
// from a known-absent state. t.Setenv registers the restore.
func clearAllKanbanEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanSpec,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanSettingsInjected,
		// t118: the per-lane cap seed is fill-if-absent, so an ambient value
		// would mask it in the assertions below.
		config.EnvClaudeCodeMaxConcurrentSubagents,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// TestResolveKanbanBranchTruthTable is AC-FB-001..004 + AC-FB-027: the four-row
// truth table at spec.md §A.2, including the breaking-change 5th case
// (companion-shape --name alone → no-op).
func TestResolveKanbanBranchTruthTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		kanbanEnabled bool
		isCompanion   bool
		want          kanbanBranch
	}{
		{"row1: -k alone → lead", true, false, kanbanBranchLead},
		{"row2: -k --name <companion> → companion", true, true, kanbanBranchCompanion},
		{"row3: --name <non-companion> alone → no-op", false, false, kanbanBranchNone},
		{"row4: -k --name <non-companion> → lead", true, false, kanbanBranchLead},
		{"row5 (BREAKING): --name <companion-shape> alone → no-op", false, true, kanbanBranchNone},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := resolveKanbanBranch(c.kanbanEnabled, c.isCompanion)
			if got != c.want {
				t.Errorf("resolveKanbanBranch(kanbanEnabled=%v, isCompanion=%v) = %v, want %v",
					c.kanbanEnabled, c.isCompanion, got, c.want)
			}
		})
	}
}

// TestDispatchOutcome_LeadEnvState is AC-FB-001 / AC-FB-006 / AC-FB-007: driving
// the lead branch (kanbanBranchLead) through the real enterKanbanMode produces the correct env
// state (MOAI_KANBAN set, MOAI_KANBAN_ID set, MOAI_KANBAN_LABEL unset).
func TestDispatchOutcome_LeadEnvState(t *testing.T) {
	clearAllKanbanEnv(t)
	if resolveKanbanBranch(true, false) != kanbanBranchLead {
		t.Fatal("expected lead branch")
	}
	restore := enterKanbanMode("SPEC-FOO-001", "")
	defer restore()

	if os.Getenv(config.EnvMoaiKanban) == "" {
		t.Errorf("MOAI_KANBAN not set (lead must carry the chain seed)")
	}
	if os.Getenv(config.EnvMoaiKanbanID) == "" {
		t.Errorf("MOAI_KANBAN_ID not set")
	}
	if os.Getenv(config.EnvMoaiKanbanSpec) != "SPEC-FOO-001" {
		t.Errorf("MOAI_KANBAN_SPEC = %q, want SPEC-FOO-001", os.Getenv(config.EnvMoaiKanbanSpec))
	}
	if _, present := os.LookupEnv(config.EnvMoaiKanbanLabel); present {
		t.Errorf("MOAI_KANBAN_LABEL must NOT be set on a lead")
	}
}

// TestDispatchOutcome_CompanionEnvState is AC-FB-002 / AC-FB-006: driving the
// companion branch through enterKanbanCompanionMode produces the correct env
// state (MOAI_KANBAN_LABEL set, MOAI_KANBAN unset, MOAI_KANBAN_ID unset), and
// t118 seeds the per-lane agent cap — the companion's subagent fan-out is
// what the cap bounds.
func TestDispatchOutcome_CompanionEnvState(t *testing.T) {
	clearAllKanbanEnv(t)
	if resolveKanbanBranch(true, true) != kanbanBranchCompanion {
		t.Fatal("expected companion branch")
	}
	restore := enterKanbanCompanionMode("runner-abc123")
	defer restore()

	if os.Getenv(config.EnvMoaiKanbanLabel) != "runner-abc123" {
		t.Errorf("MOAI_KANBAN_LABEL = %q, want \"runner-abc123\"", os.Getenv(config.EnvMoaiKanbanLabel))
	}
	if _, present := os.LookupEnv(config.EnvMoaiKanban); present {
		t.Errorf("MOAI_KANBAN must NOT be set on a companion")
	}
	if got := os.Getenv(config.EnvClaudeCodeMaxConcurrentSubagents); got != "10" {
		t.Errorf("%s = %q, want 10 (the per-lane cap on a companion)", config.EnvClaudeCodeMaxConcurrentSubagents, got)
	}
}

// TestDispatchOutcome_NoOpEnvState is AC-FB-003 / AC-FB-027: the no-op branch
// leaves every kanban env var unset — including the breaking-change case where
// --name carries the companion shape but -k is absent.
func TestDispatchOutcome_NoOpEnvState(t *testing.T) {
	clearAllKanbanEnv(t)
	if resolveKanbanBranch(false, true) != kanbanBranchNone {
		t.Fatal("expected no-op branch for companion-shape --name alone (AC-FB-027)")
	}
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanLabel, config.EnvMoaiKanbanSpec} {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s set on a no-op dispatch", key)
		}
	}
}
