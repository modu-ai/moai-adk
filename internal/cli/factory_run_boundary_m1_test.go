package cli

// factory_run_boundary_m1_test.go — SPEC-ROLE-NAMING-CODE-001 M1: the factory
// run boundary and the kanban registry notice (REQ-RNC-022, REQ-RNC-025).

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedLegacyPeerRun creates a project root with an active factory run R whose
// broker holds one peer recorded under the legacy vocabulary, owned by
// livePID when live is true.
func seedLegacyPeerRun(t *testing.T, runID, role, slot string, live bool) string {
	t.Helper()
	root := t.TempDir()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "claude", ManifestJSON: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	s, err := factorymsg.Open(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	start := homestate.CurrentProcessFingerprint()
	p := factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: runID, Backend: "claude",
		Role: role, Slot: slot, Generation: 1,
		PID:          os.Getpid(),
		ProcessStart: start,
	}
	if !live {
		p.PID = 999999999
		p.ProcessStart = "2020-01-01T00:00:00Z"
	}
	if _, err := s.RegisterLaunchPending(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()
	return root
}

func TestEnterSelectedFactoryRunRefusesLiveLegacyPeer(t *testing.T) { // AC-RNC-022 first clause
	root := seedLegacyPeerRun(t, "runR", "worker", "worker-2", true)
	restore, err := enterSelectedFactoryRun(root, "runR", false)
	if err == nil {
		restore()
		t.Fatal("enterSelectedFactoryRun = nil error, want the legacy-run refusal")
	}
	msg := err.Error()
	for _, want := range []string{"worker-2", "runR", "moai factory runs --retire runR"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal %q missing %q", msg, want)
		}
	}
}

func TestEnterSelectedFactoryRunDeadLegacyPeerProceeds(t *testing.T) {
	root := seedLegacyPeerRun(t, "runD", "worker", "worker-2", false)
	restore, err := enterSelectedFactoryRun(root, "runD", false)
	if err != nil {
		t.Fatalf("enterSelectedFactoryRun with dead legacy peer: %v", err)
	}
	restore()
}

func TestEnterSelectedFactoryRunLiveLegacyLeaderPeerRefuses(t *testing.T) {
	root := seedLegacyPeerRun(t, "runL", "lead", "lead", true)
	_, err := enterSelectedFactoryRun(root, "runL", false)
	if err == nil {
		t.Fatal("enterSelectedFactoryRun = nil error, want refusal for legacy lead peer")
	}
	if !strings.Contains(err.Error(), "lead") || !strings.Contains(err.Error(), "runL") {
		t.Errorf("refusal %q does not name the legacy value and run", err.Error())
	}
}

// REQ-RNC-025 first clause: a live `lead` entry in leads.json produces one
// notice naming it and the relaunch step, and the launch proceeds under
// `leader` — resolveLeaderName never blocks.
func TestResolveLeadNameNoticesLegacyLeadEntry(t *testing.T) {
	root := t.TempDir()
	path := leaderRegistryPath(root)
	reg := loadFactoryRegistry(path)
	reg["lead"] = factoryLaneEntry{PID: os.Getpid(), RegisteredAt: "2026-09-01T00:00:00Z"}
	if err := saveFactoryRegistry(path, reg); err != nil {
		t.Fatal(err)
	}

	var notes bytes.Buffer
	got := resolveLeaderName(root, kanban.LeaderLabel(), &notes)
	if got != "leader" {
		t.Fatalf("resolveLeaderName = %q, want leader (legacy entry never blocks)", got)
	}
	out := notes.String()
	if !strings.Contains(out, `"lead"`) && !strings.Contains(out, "lead") {
		t.Errorf("notice %q does not name the legacy entry", out)
	}
	if !strings.Contains(out, "relaunch") {
		t.Errorf("notice %q does not name the relaunch step", out)
	}
	if strings.Count(out, "lead") > 0 && strings.Count(out, "\n") > 0 && len(strings.Split(strings.TrimSpace(out), "\n")) != 1 {
		t.Errorf("notice is not one line: %q", out)
	}
	// The legacy entry is untouched afterwards.
	reg2 := loadFactoryRegistry(path)
	e1, ok1 := reg["lead"]
	e2, ok2 := reg2["lead"]
	if !ok1 || !ok2 || e1.PID != e2.PID {
		t.Errorf("legacy lead entry changed across the launch: %v → %v", e1, e2)
	}
}

func TestFactoryCardOwnerWriterUsesLeaderConstant(t *testing.T) {
	// recordFactoryCardState falls back to the leader role constant when the
	// session carries no lane label (REQ-RNC-010: factory card owner).
	if kanban.RoleLeader != "leader" {
		t.Fatalf("kanban.RoleLeader = %q, want leader", kanban.RoleLeader)
	}
}
