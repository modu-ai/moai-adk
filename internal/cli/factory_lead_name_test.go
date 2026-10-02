package cli

// factory_lead_name_test.go pins the leader-session name injection: a leader
// launched as a bare `moai cc -f` carries only an AI-generated title, which
// claude discards on /clear, so the launcher supplies an explicit
// `--name leader` instead. The operator's own name always wins.

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
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvFactoryRunID, "abc123")

	got := leaderNameArgs([]string{"-p", "work"})
	want := []string{"--name", "leader"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("leaderNameArgs = %q, want %q", got, want)
	}
}

// TestLeadNameArgs_NeverOverridesOperatorName is the requirement that an
// operator who named their lead by hand keeps that name — in every form.
func TestLeadNameArgs_NeverOverridesOperatorName(t *testing.T) {
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvFactoryRunID, "abc123")

	for _, args := range [][]string{
		{"--name", "board-watch"},
		{"--name=board-watch"},
		{"-n", "board-watch"},
		{"-n=board-watch"},
	} {
		if got := leaderNameArgs(args); got != nil {
			t.Errorf("leaderNameArgs(%q) = %q, want nil (operator name wins)", args, got)
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
	clearFactoryTestEnv(t)

	got := leaderNameArgs(nil)
	if len(got) != 2 || got[0] != "--name" || got[1] != "leader" {
		t.Errorf("leaderNameArgs with no run id = %q, want [--name leader]", got)
	}
}

// TestLeadNameArgs_LabelIsNotLaneShape guards the reclassification hazard: the
// injected name must never satisfy the lane-shape discriminator, or a re-parse
// of the argv would route the leader down the lane branch.
func TestLeadNameArgs_LabelIsNotLaneShape(t *testing.T) {
	clearFactoryTestEnv(t)

	args := leaderNameArgs(nil)
	if len(args) != 2 {
		t.Fatalf("leaderNameArgs = %q, want a --name pair", args)
	}
	if _, isLane := kanban.SplitFactoryLaneLabel(args[1]); isLane {
		t.Errorf("injected leader label %q reads as a lane label", args[1])
	}
	if _, ok := parseFactoryLaneLabel(args); ok {
		t.Errorf("injected leader label %q is picked up by parseFactoryLaneLabel", args[1])
	}
}

// TestEnterFactoryLeaderMode_AdoptsOperatorLeadRunID is the assertion whose
// ABSENCE let a divergence ship green: the prior suite checked only that
// leaderNameArgs returned nil for an operator-named leader, never that the run
// id the launcher published matched the one in that name. It did not — the
// launcher minted a fresh id beside it, and the SessionStart notice, which
// reads the environment, then printed lane commands for a run the session was
// not on.
//
// Everything downstream of the mint is asserted, not just the id itself: the
// leader socket path is derived from the same value and is what a lane would
// address.
func TestEnterFactoryLeaderMode_AdoptsOperatorLeadRunID(t *testing.T) {
	clearFactoryTestEnv(t)

	args := []string{"--name", "leader-abc123"}
	label, ok := parseLeaderLabel(args)
	if !ok {
		t.Fatalf("parseLeaderLabel(%q) did not recognize the leader name", args)
	}
	restore := enterFactoryLeaderMode(1, label)
	defer restore()

	if got := os.Getenv(config.EnvFactoryRunID); got != "abc123" {
		t.Errorf("%s = %q, want %q (the id from the operator's name)", config.EnvFactoryRunID, got, "abc123")
	}
	if got, want := os.Getenv(config.EnvFactoryLeadAddr), "/tmp/moai-socket-factory/abc123"; got != want {
		t.Errorf("%s = %q, want %q", config.EnvFactoryLeadAddr, got, want)
	}
	// The operator's name still wins — adoption must not also inject a second
	// --name and hand claude two.
	if got := leaderNameArgs(args); got != nil {
		t.Errorf("leaderNameArgs(%q) = %q, want nil (operator name wins)", args, got)
	}
}

// TestEnterFactoryLeaderMode_MintsWithoutLeadName is the other half of the
// contract: a leader with no usable name in argv still gets a run id, so the
// bare `moai cc -f` launch is unchanged by adoption.
func TestEnterFactoryLeaderMode_MintsWithoutLeadName(t *testing.T) {
	for _, c := range []struct {
		name  string
		label string
	}{
		{"no name at all", ""},
		{"non-lead name", "board-watch"},
		{"the bare leader name carries no id to adopt", "leader"},
		{"a bump number is not a run id", "leader-2"},
		{"lead prefix with no id", "leader-"},
		{"uppercase is not a run id shape", "leader-ABC123"},
		{"a second hyphen is not a run id shape", "leader-a-b"},
	} {
		t.Run(c.name, func(t *testing.T) {
			clearFactoryTestEnv(t)
			restore := enterFactoryLeaderMode(1, c.label)
			defer restore()

			if got := os.Getenv(config.EnvFactoryRunID); got == "" {
				t.Errorf("%s is empty, want a freshly minted run id", config.EnvFactoryRunID)
			}
		})
	}
}

// TestParseLeaderLabel covers the four name forms claude accepts and the shapes
// that must NOT read as a lead label.
func TestParseLeaderLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"the bare leader name", []string{"--name", "leader"}, "leader"},
		{"a bumped leader name", []string{"--name", "leader-1"}, "leader-1"},
		{"--name value", []string{"--name", "leader-abc123"}, "leader-abc123"},
		{"--name=value", []string{"--name=leader-abc123"}, "leader-abc123"},
		{"-n value", []string{"-n", "leader-abc123"}, "leader-abc123"},
		{"-n=value", []string{"-n=leader-abc123"}, "leader-abc123"},
		{"absent", []string{"-p", "work"}, ""},
		{"a non-lead name is not ours", []string{"--name", "board-watch"}, ""},
		{"a lane name is not ours", []string{"--name", "lane-2"}, ""},
		{"past the pass-through marker is not ours", []string{"--", "--name", "leader-abc123"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseLeaderLabel(c.args)
			if c.want == "" {
				if ok {
					t.Errorf("parseLeaderLabel(%q) = %q, want no match", c.args, got)
				}
				return
			}
			if !ok || got != c.want {
				t.Errorf("parseLeaderLabel(%q) = %q/%v, want %q/true", c.args, got, ok, c.want)
			}
		})
	}
}

// TestLeadLabelNeverReadsAsLane guards the branch that would be broken by
// recognizing leader names: resolveFactoryBranch must still route a
// leader-named `-f` launch down the leader branch, not the lane one.
func TestLeadLabelNeverReadsAsLane(t *testing.T) {
	t.Parallel()

	args := []string{"--name", "leader-abc123"}
	_, isLane := parseFactoryLaneLabel(args)
	if isLane {
		t.Fatalf("parseFactoryLaneLabel(%q) matched a leader name", args)
	}
	if branch := resolveFactoryBranch(true, isLane); branch != factoryBranchLeader {
		t.Errorf("resolveFactoryBranch = %v, want the leader branch", branch)
	}
}

// TestLeadRunID_AdoptsEnvironmentRunID pins the carrier that REPLACED the name
// round-trip. With the run id gone from the session name, a relaunched lead
// recovers its id from MOAI_KANBAN_ID or not at all — so this is the whole
// continuity path, and a regression here silently forks a relaunch onto a
// second run id (the notice header and the lead socket path both follow it).
func TestLeadRunID_AdoptsEnvironmentRunID(t *testing.T) {
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvFactoryRunID, "abc123")

	if got := leaderRunID(""); got != "abc123" {
		t.Errorf("leaderRunID(\"\") = %q, want %q (adopted from the environment)", got, "abc123")
	}
	if got := leaderRunID("lead"); got != "abc123" {
		t.Errorf("leaderRunID(\"lead\") = %q, want %q (the bare name carries no id)", got, "abc123")
	}
}

// TestLeadRunID_LegacyNameWinsOverEnvironment pins the migration order: an
// operator still pasting an old `lead-<run-id>` launch line lands on the run
// that name states, not on whatever the environment happened to hold.
func TestLeadRunID_LegacyNameWinsOverEnvironment(t *testing.T) {
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvFactoryRunID, "stale1")

	if got := leaderRunID("leader-abc123"); got != "abc123" {
		t.Errorf("leaderRunID(\"lead-abc123\") = %q, want %q (the pasted name wins)", got, "abc123")
	}
}

// TestLeadRunID_BumpNumberIsNotARunID is the distinction the launcher must not
// lose: `lead-2` names the second live lead on this machine, not run 2.
// Adopting it would publish a run id and a lead socket path that no other
// session shares.
func TestLeadRunID_BumpNumberIsNotARunID(t *testing.T) {
	clearFactoryTestEnv(t)
	t.Setenv(config.EnvFactoryRunID, "abc123")

	if got := leaderRunID(kanban.LeaderNumberLabel(2)); got != "abc123" {
		t.Errorf("leaderRunID(%q) = %q, want %q (a bump number is not a run id)", kanban.LeaderNumberLabel(2), got, "abc123")
	}
}

// TestResolveLeadName_BumpsPastALiveClaim asserts the collision behavior the
// bare name makes possible: a second lead on one machine takes the next free
// number rather than answering to the same name as the first, which is what
// keeps every session addressable by name alone.
func TestResolveLeadName_BumpsPastALiveClaim(t *testing.T) {
	root := t.TempDir()

	first := resolveLeaderName(root, kanban.LeaderLabel(), nil)
	if first != kanban.LeaderLabel() {
		t.Fatalf("first lead launched as %q, want the bare %q", first, kanban.LeaderLabel())
	}
	// This process holds the claim, so it is alive by construction.
	second := resolveLeaderName(root, kanban.LeaderLabel(), nil)
	if want := kanban.LeaderNumberLabel(1); second != want {
		t.Errorf("second lead launched as %q, want %q", second, want)
	}
}

// TestResolveLeadName_SeparateFromLanes pins the namespace split: the leader
// registry is its own file, so a lane claim can never bump a leader and vice
// versa.
func TestResolveLeadName_SeparateFromLanes(t *testing.T) {
	root := t.TempDir()

	if leaderRegistryPath(root) == factoryRegistryPath(root) {
		t.Error("leader and lane registries share one path")
	}
	if got := resolveLeaderName(root, kanban.LeaderLabel(), nil); got != kanban.LeaderLabel() {
		t.Errorf("leader launched as %q on a fresh root, want the bare %q", got, kanban.LeaderLabel())
	}
}

// TestAppendLeadName_OperatorNameWins asserts the helper preserves the gate it
// wraps: an operator-named lead gets no injected name, bumped or otherwise.
func TestAppendLeadName_OperatorNameWins(t *testing.T) {
	root := t.TempDir()

	args := []string{"--name", "board-watch"}
	got, name := appendLeaderName(args, root, nil)
	if len(got) != len(args) {
		t.Errorf("appendLeaderName appended to an operator-named lead: %q", got)
	}
	// The operator's own name is still REPORTED, so the title registered
	// downstream is the name the session actually answers to (issue #1596).
	if name != "board-watch" {
		t.Errorf("appendLeaderName reported %q for an operator-named lead, want %q", name, "board-watch")
	}
}
