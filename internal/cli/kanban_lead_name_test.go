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

// TestLeadNameArgs_InjectsWhenUnnamed is the core case: a bare lead gets the
// explicit bare-role name, which is what survives /clear.
func TestLeadNameArgs_InjectsWhenUnnamed(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	got := leadNameArgs([]string{"-p", "work"})
	want := []string{"--name", "lead"}
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

// TestLeadNameArgs_InjectsWithoutRunID pins the gate t133 REMOVED. The name no
// longer embeds the run id, so there is no degenerate `lead-` form to guard
// against and no reason to skip the injection when the environment carries no
// id: the name is `lead` either way. A regression that reinstates the old gate
// would leave a bare lead unnamed again, which is the failure the injection
// exists to prevent.
func TestLeadNameArgs_InjectsWithoutRunID(t *testing.T) {
	clearAllKanbanEnv(t)

	got := leadNameArgs(nil)
	if len(got) != 2 || got[0] != "--name" || got[1] != "lead" {
		t.Errorf("leadNameArgs with no run id = %q, want [--name lead]", got)
	}
}

// TestLeadNameArgs_LabelIsNotCompanionShape guards the reclassification hazard:
// the injected name must never satisfy the companion-shape discriminator, or a
// re-parse of the argv would route the lead down the companion branch.
func TestLeadNameArgs_LabelIsNotCompanionShape(t *testing.T) {
	clearAllKanbanEnv(t)

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

// TestEnterKanbanMode_AdoptsOperatorLeadRunID is the assertion whose ABSENCE let
// the divergence ship green: the prior suite checked only that leadNameArgs
// returned nil for an operator-named lead, never that the run id the launcher
// published matched the one in that name. It did not — the launcher minted a
// fresh id beside it, and the SessionStart notice, which reads the environment,
// then printed companion commands for a run the session was not on.
//
// Everything downstream of the mint is asserted, not just the id itself: the
// leader socket path is derived from the same value and is what a companion
// would address.
func TestEnterKanbanMode_AdoptsOperatorLeadRunID(t *testing.T) {
	clearAllKanbanEnv(t)

	args := []string{"--name", "lead-abc123"}
	label, ok := parseLeadLabel(args)
	if !ok {
		t.Fatalf("parseLeadLabel(%q) did not recognize the lead name", args)
	}
	restore := enterKanbanMode("", label)
	defer restore()

	if got := os.Getenv(config.EnvMoaiKanbanID); got != "abc123" {
		t.Errorf("%s = %q, want %q (the id from the operator's name)", config.EnvMoaiKanbanID, got, "abc123")
	}
	if got, want := os.Getenv(config.EnvMoaiKanbanLeadAddr), "/tmp/moai-socket-kanban/abc123"; got != want {
		t.Errorf("%s = %q, want %q", config.EnvMoaiKanbanLeadAddr, got, want)
	}
	// The operator's name still wins — adoption must not also inject a second
	// --name and hand claude two.
	if got := leadNameArgs(args); got != nil {
		t.Errorf("leadNameArgs(%q) = %q, want nil (operator name wins)", args, got)
	}
}

// TestEnterKanbanMode_MintsWithoutLeadName is the other half of the contract: a
// lead with no usable name in argv still gets a run id, so the bare
// `moai cc -k` launch is unchanged by adoption.
func TestEnterKanbanMode_MintsWithoutLeadName(t *testing.T) {
	for _, c := range []struct {
		name  string
		label string
	}{
		{"no name at all", ""},
		{"non-lead name", "board-watch"},
		{"the bare lead name carries no id to adopt", "lead"},
		{"a bump number is not a run id", "lead-2"},
		{"lead prefix with no id", "lead-"},
		{"uppercase is not a run id shape", "lead-ABC123"},
		{"a second hyphen is not a run id shape", "lead-a-b"},
	} {
		t.Run(c.name, func(t *testing.T) {
			clearAllKanbanEnv(t)
			restore := enterKanbanMode("", c.label)
			defer restore()

			if got := os.Getenv(config.EnvMoaiKanbanID); got == "" {
				t.Errorf("%s is empty, want a freshly minted run id", config.EnvMoaiKanbanID)
			}
		})
	}
}

// TestParseLeadLabel covers the four name forms claude accepts and the shapes
// that must NOT read as a lead label.
func TestParseLeadLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"the bare lead name", []string{"--name", "lead"}, "lead"},
		{"a bumped lead name", []string{"--name", "lead-1"}, "lead-1"},
		{"--name value", []string{"--name", "lead-abc123"}, "lead-abc123"},
		{"--name=value", []string{"--name=lead-abc123"}, "lead-abc123"},
		{"-n value", []string{"-n", "lead-abc123"}, "lead-abc123"},
		{"-n=value", []string{"-n=lead-abc123"}, "lead-abc123"},
		{"absent", []string{"-p", "work"}, ""},
		{"a non-lead name is not ours", []string{"--name", "board-watch"}, ""},
		{"a companion name is not ours", []string{"--name", "run-abc123"}, ""},
		{"past the pass-through marker is not ours", []string{"--", "--name", "lead-abc123"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseLeadLabel(c.args)
			if c.want == "" {
				if ok {
					t.Errorf("parseLeadLabel(%q) = %q, want no match", c.args, got)
				}
				return
			}
			if !ok || got != c.want {
				t.Errorf("parseLeadLabel(%q) = %q/%v, want %q/true", c.args, got, ok, c.want)
			}
		})
	}
}

// TestLeadLabelNeverReadsAsCompanion guards the branch that would be broken by
// recognizing lead names: resolveKanbanBranch must still route a lead-named
// `-k` launch down the lead branch, not the companion one.
func TestLeadLabelNeverReadsAsCompanion(t *testing.T) {
	t.Parallel()

	args := []string{"--name", "lead-abc123"}
	if _, isCompanion := parseCompanionLabel(args); isCompanion {
		t.Fatalf("parseCompanionLabel(%q) matched a lead name", args)
	}
	if branch := resolveKanbanBranch(true, false); branch != kanbanBranchLead {
		t.Errorf("resolveKanbanBranch = %v, want the lead branch", branch)
	}
}

// TestLeadRunID_AdoptsEnvironmentRunID pins the carrier that REPLACED the name
// round-trip. With the run id gone from the session name, a relaunched lead
// recovers its id from MOAI_KANBAN_ID or not at all — so this is the whole
// continuity path, and a regression here silently forks a relaunch onto a
// second run id (the notice header and the lead socket path both follow it).
func TestLeadRunID_AdoptsEnvironmentRunID(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	if got := leadRunID(""); got != "abc123" {
		t.Errorf("leadRunID(\"\") = %q, want %q (adopted from the environment)", got, "abc123")
	}
	if got := leadRunID("lead"); got != "abc123" {
		t.Errorf("leadRunID(\"lead\") = %q, want %q (the bare name carries no id)", got, "abc123")
	}
}

// TestLeadRunID_LegacyNameWinsOverEnvironment pins the migration order: an
// operator still pasting an old `lead-<run-id>` launch line lands on the run
// that name states, not on whatever the environment happened to hold.
func TestLeadRunID_LegacyNameWinsOverEnvironment(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "stale1")

	if got := leadRunID("lead-abc123"); got != "abc123" {
		t.Errorf("leadRunID(\"lead-abc123\") = %q, want %q (the pasted name wins)", got, "abc123")
	}
}

// TestLeadRunID_BumpNumberIsNotARunID is the distinction the launcher must not
// lose: `lead-2` names the second live lead on this machine, not run 2.
// Adopting it would publish a run id and a lead socket path that no other
// session shares.
func TestLeadRunID_BumpNumberIsNotARunID(t *testing.T) {
	clearAllKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	if got := leadRunID(kanban.LeadNumberLabel(2)); got != "abc123" {
		t.Errorf("leadRunID(%q) = %q, want %q (a bump number is not a run id)", kanban.LeadNumberLabel(2), got, "abc123")
	}
}

// TestResolveLeadName_BumpsPastALiveClaim asserts the collision behavior the
// bare name makes possible: a second lead on one machine takes the next free
// number rather than answering to the same name as the first, which is what
// keeps every session addressable by name alone.
func TestResolveLeadName_BumpsPastALiveClaim(t *testing.T) {
	root := t.TempDir()

	first := resolveLeadName(root, kanban.LeadLabel(), nil)
	if first != kanban.LeadLabel() {
		t.Fatalf("first lead launched as %q, want the bare %q", first, kanban.LeadLabel())
	}
	// This process holds the claim, so it is alive by construction.
	second := resolveLeadName(root, kanban.LeadLabel(), nil)
	if want := kanban.LeadNumberLabel(1); second != want {
		t.Errorf("second lead launched as %q, want %q", second, want)
	}
}

// TestResolveLeadName_SeparateFromCompanions pins the namespace split: the lead
// registry is its own file, so a companion holding `plan` can never bump a lead
// and vice versa.
func TestResolveLeadName_SeparateFromCompanions(t *testing.T) {
	root := t.TempDir()

	resolveCompanionName(root, kanban.CompanionLabel("plan"), nil)
	if got := resolveLeadName(root, kanban.LeadLabel(), nil); got != kanban.LeadLabel() {
		t.Errorf("lead launched as %q after a companion claim, want the bare %q", got, kanban.LeadLabel())
	}
	if leadRegistryPath(root) == companionRegistryPath(root) {
		t.Error("lead and companion registries share one path")
	}
}

// TestAppendLeadName_OperatorNameWins asserts the helper preserves the gate it
// wraps: an operator-named lead gets no injected name, bumped or otherwise.
func TestAppendLeadName_OperatorNameWins(t *testing.T) {
	root := t.TempDir()

	args := []string{"--name", "board-watch"}
	got, name := appendLeadName(args, root, nil)
	if len(got) != len(args) {
		t.Errorf("appendLeadName appended to an operator-named lead: %q", got)
	}
	// The operator's own name is still REPORTED, so the title registered
	// downstream is the name the session actually answers to (issue #1596).
	if name != "board-watch" {
		t.Errorf("appendLeadName reported %q for an operator-named lead, want %q", name, "board-watch")
	}
}
