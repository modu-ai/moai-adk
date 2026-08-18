package cli

// kanban_companion_name_test.go pins the bare-role companion naming policy
// (card t56): a companion is launched under its bare role name, a label held
// by a live session is bumped to the next free number, and the companion
// branch publishes no run id — the t21 incident class (lead MOAI_KANBAN_ID
// vs companion suffix mismatch) is removed at its root by the latter.

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestResolveCompanionName_FreeLabelUnchanged: a bare role nobody holds
// launches under itself, and the claim is registered so the NEXT companion
// bumps.
func TestResolveCompanionName_FreeLabelUnchanged(t *testing.T) {
	root := t.TempDir()
	if got := resolveCompanionName(root, "planner", nil); got != "planner" {
		t.Fatalf("free label = %q, want planner", got)
	}
	reg := loadFactoryRegistry(companionRegistryPath(root))
	if e, ok := reg["planner"]; !ok || e.PID != os.Getpid() {
		t.Errorf("planner not registered to this pid: %+v", reg)
	}
}

// TestResolveCompanionName_LiveClaimBumps: a label held by a live session is
// bumped to the next free number for its role, and the bump is stated on the
// notes writer the operator reads at launch.
func TestResolveCompanionName_LiveClaimBumps(t *testing.T) {
	root := t.TempDir()
	// Simulate two live holders: planner and planner-1.
	if err := saveFactoryRegistry(companionRegistryPath(root), map[string]factoryWorkerEntry{
		"planner":   {PID: 11100},
		"planner-1": {PID: 11101},
	}); err != nil {
		t.Fatalf("seed registry: %v", err)
	}
	probe := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid == 11100 || pid == 11101 }
	defer func() { factoryProcessAlive = probe }()

	var notes bytes.Buffer
	got := resolveCompanionName(root, "planner", &notes)
	if got != "planner-2" {
		t.Fatalf("bumped label = %q, want planner-2 (planner and planner-1 are live)", got)
	}
	if !strings.Contains(notes.String(), "planner-2") {
		t.Errorf("operator note missing the final label: %q", notes.String())
	}
	reg := loadFactoryRegistry(companionRegistryPath(root))
	if e, ok := reg["planner-2"]; !ok || e.PID != os.Getpid() {
		t.Errorf("planner-2 not registered to this pid: %+v", reg)
	}
}

// TestResolveCompanionName_DeadClaimReclaimed: a crashed companion leaves a
// dead pid behind; a dead claim frees the name so a relaunch reuses it instead
// of counting up forever.
func TestResolveCompanionName_DeadClaimReclaimed(t *testing.T) {
	root := t.TempDir()
	if err := saveFactoryRegistry(companionRegistryPath(root), map[string]factoryWorkerEntry{
		"planner": {PID: 11100},
	}); err != nil {
		t.Fatalf("seed registry: %v", err)
	}
	probe := factoryProcessAlive
	factoryProcessAlive = func(int) bool { return false }
	defer func() { factoryProcessAlive = probe }()

	if got := resolveCompanionName(root, "planner", nil); got != "planner" {
		t.Fatalf("dead claim should free the label, got %q", got)
	}
}

// TestResolveCompanionName_LegacySuffixHeldBumpsToNumbered: a legacy
// run-id-suffixed label still parses, and when held the bump produces the
// numbered form for the role — never a second hyphen.
func TestResolveCompanionName_LegacySuffixHeldBumpsToNumbered(t *testing.T) {
	root := t.TempDir()
	if err := saveFactoryRegistry(companionRegistryPath(root), map[string]factoryWorkerEntry{
		"planner-abc123": {PID: 11100},
	}); err != nil {
		t.Fatalf("seed registry: %v", err)
	}
	probe := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid == 11100 }
	defer func() { factoryProcessAlive = probe }()

	got := resolveCompanionName(root, "planner-abc123", nil)
	if got != "planner-1" {
		t.Fatalf("bumped legacy label = %q, want planner-1", got)
	}
	if _, _, ok := kanban.SplitCompanionLabel(got); !ok {
		t.Errorf("bumped label %q does not parse as a companion label", got)
	}
}

// TestResolveCompanionName_NonCompanionLabelUntouched: the bump machinery is
// reached only after the companion-shape branch, but it must also hold its own
// — a label that does not parse is registered and returned as supplied, not
// mangled.
func TestResolveCompanionName_NonCompanionLabelUntouched(t *testing.T) {
	root := t.TempDir()
	if got := resolveCompanionName(root, "board-watch", nil); got != "board-watch" {
		t.Fatalf("non-companion label = %q, want it back unchanged", got)
	}
}

// TestEnterKanbanCompanionModeDoesNotPublishRunID is the t21 root removal: a
// companion's membership is its label, and no companion surface carries a run
// id anymore. Publishing one derived from the label suffix (a collision
// number under the new policy) would recreate exactly the mismatch class the
// bare-name policy removes.
func TestEnterKanbanCompanionModeDoesNotPublishRunID(t *testing.T) {
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanLabel} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}

	restore := enterKanbanCompanionMode("planner")
	defer restore()

	if got := os.Getenv(config.EnvMoaiKanbanLabel); got != "planner" {
		t.Errorf("%s = %q, want %q", config.EnvMoaiKanbanLabel, got, "planner")
	}
	if _, present := os.LookupEnv(config.EnvMoaiKanbanID); present {
		t.Errorf("%s must NOT be set on a companion (the run id is leader-owned state; "+
			"a companion-published id is the t21 mismatch root)", config.EnvMoaiKanbanID)
	}
	if _, present := os.LookupEnv(config.EnvMoaiKanban); present {
		t.Errorf("%s must NOT be set on a companion (it seeds the chain)", config.EnvMoaiKanban)
	}
}

// TestCompanionRoleBareLabel: the role the session record carries is read back
// from a bare label unchanged.
func TestCompanionRoleBareLabel(t *testing.T) {
	t.Parallel()
	if got := companionRole("planner"); got != "planner" {
		t.Errorf("companionRole(planner) = %q, want planner", got)
	}
	if got := companionRole("planner-2"); got != "planner" {
		t.Errorf("companionRole(planner-2) = %q, want planner", got)
	}
}
