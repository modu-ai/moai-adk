package cli

// kanban_lead_name_test.go pins the lead-session name injection: a lead
// launched as a bare `moai cc -k` carries only an AI-generated title, which
// claude discards on /clear, so the launcher supplies an explicit
// `--name lead-<run-id>` instead. The operator's own name always wins.

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestOperatorSuppliedName covers every form claude accepts for a session name,
// plus the two cases that must read as "no name": absent entirely, and present
// only past the pass-through marker.
func TestOperatorSuppliedName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"absent", []string{"-p", "work"}, false},
		{"empty", nil, false},
		{"--name value", []string{"--name", "board-watch"}, true},
		{"--name=value", []string{"--name=board-watch"}, true},
		{"-n value", []string{"-n", "board-watch"}, true},
		{"-n=value", []string{"-n=board-watch"}, true},
		{"companion-shape name still counts", []string{"--name", "run-abc123"}, true},
		{"past the pass-through marker is not ours", []string{"--", "--name", "x"}, false},
		{"before the marker still counts", []string{"--name", "x", "--", "--print"}, true},
		// A bare `--name` with no following token is still an operator name
		// declaration: moai must not append a second one and hand claude two.
		{"dangling --name", []string{"--name"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := operatorSuppliedName(c.args); got != c.want {
				t.Errorf("operatorSuppliedName(%q) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

// TestLeadNameArgs_InjectsWhenUnnamed is the core case: a bare lead gets an
// explicit name derived from the run id enterKanbanMode published.
func TestLeadNameArgs_InjectsWhenUnnamed(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	got := leadNameArgs([]string{"-p", "work"})
	want := []string{"--name", "lead-abc123"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("leadNameArgs = %q, want %q", got, want)
	}
}

// TestLeadNameArgs_NeverOverridesOperatorName is the requirement that an
// operator who named their lead by hand keeps that name — in every form.
func TestLeadNameArgs_NeverOverridesOperatorName(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	for _, args := range [][]string{
		{"--name", "board-watch"},
		{"--name=board-watch"},
		{"-n", "board-watch"},
		{"-n=board-watch"},
	} {
		if got := leadNameArgs(args); got != nil {
			t.Errorf("leadNameArgs(%q) = %q, want nil (operator name wins)", args, got)
		}
	}
}

// TestLeadNameArgs_NoRunIDNoName is the fail-open gate: without a run id the
// injection is skipped rather than producing the nonsense name `lead-`.
func TestLeadNameArgs_NoRunIDNoName(t *testing.T) {
	clearAllKanbanEnv(t)

	if got := leadNameArgs(nil); got != nil {
		t.Errorf("leadNameArgs with no run id = %q, want nil", got)
	}
}

// TestLeadNameArgs_LabelIsNotCompanionShape guards the reclassification hazard:
// the injected name must never satisfy the companion-shape discriminator, or a
// re-parse of the argv would route the lead down the companion branch.
func TestLeadNameArgs_LabelIsNotCompanionShape(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	args := leadNameArgs(nil)
	if len(args) != 2 {
		t.Fatalf("leadNameArgs = %q, want a --name pair", args)
	}
	if _, _, isCompanion := kanban.SplitCompanionLabel(args[1]); isCompanion {
		t.Errorf("injected lead label %q reads as a companion label", args[1])
	}
	if _, ok := parseCompanionLabel(args); ok {
		t.Errorf("injected lead label %q is picked up by parseCompanionLabel", args[1])
	}
}

// TestLeadNameArgs_UsesRunIDFromEnterKanbanMode wires the helper to the real
// producer rather than a hand-set variable, so a change to where the run id is
// published breaks this test instead of passing silently.
func TestLeadNameArgs_UsesRunIDFromEnterKanbanMode(t *testing.T) {
	clearAllKanbanEnv(t)
	restore := enterKanbanMode("")
	defer restore()

	runID := os.Getenv(config.EnvMoaiKanbanID)
	if runID == "" {
		t.Fatal("enterKanbanMode published no run id")
	}
	args := leadNameArgs(nil)
	if len(args) != 2 || args[1] != kanban.LeadLabel(runID) {
		t.Errorf("leadNameArgs = %q, want [--name %s]", args, kanban.LeadLabel(runID))
	}
}
