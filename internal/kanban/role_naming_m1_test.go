package kanban

// role_naming_m1_test.go — SPEC-ROLE-NAMING-CODE-001 M1: the persisted
// vocabulary moves to `leader` / `lane` / `lane-<n>` and the legacy spellings
// (`lead`, `worker`, `agent`, `worker-<n>`, `agent-<n>`) survive only as
// detection values (REQ-RNC-009, REQ-RNC-010).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestRoleLeadValueIsLeader(t *testing.T) {
	if RoleLeader != "leader" {
		t.Errorf("RoleLeader = %q, want %q", RoleLeader, "leader")
	}
	if got := LeaderLabel(); got != "leader" {
		t.Errorf("LeaderLabel() = %q, want leader", got)
	}
	if got := LeaderNumberLabel(2); got != "leader-2" {
		t.Errorf("LeaderNumberLabel(2) = %q, want leader-2", got)
	}
	if got := LeaderNumberLabel(7); got != "leader-7" {
		t.Errorf("LeaderNumberLabel(7) = %q, want leader-7", got)
	}
}

func TestSplitLeadLabelLeaderForms(t *testing.T) {
	tests := []struct {
		label  string
		ok     bool
		suffix string
	}{
		{"leader", true, ""},      // bare form
		{"leader-2", true, "2"},   // collision bump (digits)
		{"leader-r7", true, "r7"}, // run-id composed form
		{"lead", false, ""},       // legacy — no longer a lead label at all
		{"lead-7", false, ""},     // legacy bumped
		{"lane-1", false, ""},     // factory lane — never a leader label
	}
	for _, tc := range tests {
		suffix, ok := SplitLeaderLabel(tc.label)
		if ok != tc.ok || suffix != tc.suffix {
			t.Errorf("SplitLeaderLabel(%q) = (%q, %v), want (%q, %v)", tc.label, suffix, ok, tc.suffix, tc.ok)
		}
	}
}

func TestFactoryLaneLabelPrefixIsLane(t *testing.T) {
	if got := FactoryLaneLabel(3); got != "lane-3" {
		t.Errorf("FactoryLaneLabel(3) = %q, want lane-3", got)
	}
}

func TestIsLegacyLeadLabelDetection(t *testing.T) {
	for _, value := range []string{"lead", "lead-7", "lead-abc123"} {
		if !IsLegacyLeaderSpelling(value) {
			t.Errorf("IsLegacyLeaderSpelling(%q) = false, want true", value)
		}
	}
	for _, value := range []string{"leader", "leader-2", "leader-r7", "lane", ""} {
		if IsLegacyLeaderSpelling(value) {
			t.Errorf("IsLegacyLeaderSpelling(%q) = true, want false", value)
		}
	}
}

// REQ-RNC-022: a live legacy registry claim refuses the lane join with the
// retire-and-relaunch message; a dead one is stale exactly like a dead
// new-vocabulary claim.
func TestClaimFactoryWorkerRefusesLiveLegacyClaim(t *testing.T) {
	root := t.TempDir()
	if err := seedWorkerRow(t, root, "worker-3", 999999); err != nil {
		t.Fatal(err)
	}
	alive := func(int) bool { return true }
	_, err := ClaimFactoryLane(root, "", true, os.Getpid(), "newrun", alive)
	var legacy *FactoryLegacyRunError
	if !errors.As(err, &legacy) {
		t.Fatalf("ClaimFactoryLane live legacy row: err = %v, want *FactoryLegacyRunError", err)
	}
	if legacy.Label != "worker-3" {
		t.Errorf("legacy error label = %q, want worker-3", legacy.Label)
	}
	if legacy.RunID != "run7" {
		t.Errorf("legacy error run id = %q, want run7", legacy.RunID)
	}
	msg := legacy.Error()
	for _, want := range []string{"worker-3", "run7", "moai factory runs --retire run7"} {
		if !strings.Contains(msg, want) {
			t.Errorf("legacy error message %q missing %q", msg, want)
		}
	}
	// Nothing was written: the registry still holds exactly the legacy row.
	reg := LoadFactoryRegistry(FactoryRegistryPath(root))
	if len(reg) != 1 {
		t.Fatalf("registry holds %d rows after refused claim, want 1 (legacy row untouched)", len(reg))
	}
	if _, ok := reg["worker-3"]; !ok {
		t.Errorf("legacy row worker-3 vanished after refusal")
	}
}

func TestClaimFactoryWorkerDeadLegacyClaimIsStale(t *testing.T) {
	root := t.TempDir()
	if err := seedWorkerRow(t, root, "worker-3", 999999); err != nil {
		t.Fatal(err)
	}
	alive := func(int) bool { return false } // the legacy claim's pid is dead
	claim, err := ClaimFactoryLane(root, "", true, os.Getpid(), "newrun", alive)
	if err != nil {
		t.Fatalf("ClaimFactoryLane with dead legacy row: %v", err)
	}
	if claim.Label != "lane-1" {
		t.Errorf("claimed label = %q, want lane-1 (dead legacy pruned, canonical numbering)", claim.Label)
	}
}

// P3 (plan-audit debt): a live legacy claim with an empty run_id belongs to
// no run — it is neither refused nor counted as a lane.
func TestClaimFactoryWorkerIgnoresLegacyClaimOfNoRun(t *testing.T) {
	root := t.TempDir()
	// run_id '' is the shape of every legacy import; the row holds a live pid
	// but names no run.
	if err := seedWorkerRowEmptyRun(t, root, "worker-5", 999998); err != nil {
		t.Fatal(err)
	}
	alive := func(int) bool { return true }
	claim, err := ClaimFactoryLane(root, "", true, os.Getpid(), "newrun", alive)
	if err != nil {
		t.Fatalf("ClaimFactoryLane with empty-run_id legacy row: %v", err)
	}
	if claim.Label != "lane-1" {
		t.Errorf("claimed label = %q, want lane-1 (empty-run_id legacy claim is no run)", claim.Label)
	}
}

func TestClaimFactoryWorkerWritesLaneLabel(t *testing.T) {
	root := t.TempDir()
	alive := func(int) bool { return true }
	claim, err := ClaimFactoryLane(root, "", true, os.Getpid(), "newrun", alive)
	if err != nil {
		t.Fatalf("ClaimFactoryLane: %v", err)
	}
	if claim.Label != "lane-1" {
		t.Errorf("claimed label = %q, want lane-1", claim.Label)
	}
	reg := LoadFactoryRegistry(FactoryRegistryPath(root))
	if _, ok := reg["lane-1"]; !ok {
		t.Errorf("registry missing lane-1 after claim: %v", reg)
	}
}

// REQ-RNC-025 board clause: a board role declaration `lead` loses board write
// access; `leader` is admitted. DeclareRole has no production caller — this is
// exercised from tests only.
func TestBoardGuardRefusesLegacyLeadDeclaration(t *testing.T) {
	root := t.TempDir()
	if err := DeclareRole(root, "legacy-session", "lead", "lead"); err != nil {
		t.Fatalf("DeclareRole: %v", err)
	}
	before, err := os.ReadFile(filepath.Join(BoardDir(root), "roles", "legacy-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	err = requireLeaderRole(root, "legacy-session")
	if !errors.Is(err, ErrNotSoleWriter) {
		t.Fatalf("requireLeaderRole(lead declaration) err = %v, want ErrNotSoleWriter", err)
	}
	if !strings.Contains(err.Error(), "lead") {
		t.Errorf("refusal %q does not name the legacy role lead", err.Error())
	}
	after, err := os.ReadFile(filepath.Join(BoardDir(root), "roles", "legacy-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("declaration file changed on refusal")
	}
}

func TestBoardGuardAdmitsLeaderDeclaration(t *testing.T) {
	root := t.TempDir()
	if err := DeclareRole(root, "leader-session", RoleLeader, LeaderLabel()); err != nil {
		t.Fatalf("DeclareRole: %v", err)
	}
	if err := requireLeaderRole(root, "leader-session"); err != nil {
		t.Fatalf("requireLeaderRole(leader declaration) = %v, want nil", err)
	}
}

// seedWorkerRow inserts a registry claim whose run_id is run7 — the shape a
// pre-change binary of a run with an id would have written (the factory DB
// claim INSERT records run_id since the run boundary landed).
func seedWorkerRow(t *testing.T, root, label string, pid int) error {
	t.Helper()
	return seedWorkerRowRun(t, root, label, pid, "run7")
}

func seedWorkerRowEmptyRun(t *testing.T, root, label string, pid int) error {
	t.Helper()
	return seedWorkerRowRun(t, root, label, pid, "")
}

func seedWorkerRowRun(t *testing.T, root, label string, pid int, runID string) error {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	_, err = db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at,run_id) VALUES(?,?,?,?,?)`,
		label, pid, "2026-09-26T00:00:00Z", "2026-09-26T00:00:00Z", runID)
	return err
}

// TestFactoryAgentLabelLegacyRendererOnly pins the legacy renderer: it exists
// for detection and fixtures, nothing launches under it and no reader maps it
// to a lane (REQ-RNC-009).
func TestFactoryAgentLabelLegacyRendererOnly(t *testing.T) {
	if got := FactoryLegacyAgentLabel(2); got != "agent-2" {
		t.Errorf("FactoryLegacyAgentLabel(2) = %q, want agent-2", got)
	}
	if _, ok := SplitFactoryLaneLabel(FactoryLegacyAgentLabel(2)); ok {
		t.Errorf("the legacy agent label must not parse as a lane")
	}
}
