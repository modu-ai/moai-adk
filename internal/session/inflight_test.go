package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDetectInFlightTransition detects an in-progress phase transition.
// SPEC-V3R2-RT-004 AC-14: verifies the DetectInFlightTransition method.
func TestDetectInFlightTransition(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Only a plan checkpoint exists (no run checkpoint).
	planState := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-TEST-001",
		Checkpoint: &PlanCheckpoint{
			SPECID:       "SPEC-TEST-001",
			Status:       "approved",
			ResearchPath: "/research/SPEC-TEST-001",
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}
	if err := store.Checkpoint(planState); err != nil {
		t.Fatalf("Checkpoint(plan) 실패: %v", err)
	}

	// In-flight detection: plan-only must be detected as a plan->run transition.
	fromPhase, toPhase, found, err := store.DetectInFlightTransition("SPEC-TEST-001")
	if err != nil {
		t.Fatalf("DetectInFlightTransition() 에러: %v", err)
	}
	if !found {
		t.Error("in-flight transition이 감지되어야 하는데 found=false")
	}
	if fromPhase != PhasePlan {
		t.Errorf("fromPhase = %v, want %v", fromPhase, PhasePlan)
	}
	if toPhase != PhaseRun {
		t.Errorf("toPhase = %v, want %v", toPhase, PhaseRun)
	}
}

// TestDetectInFlightTransition_AllPhasesComplete verifies there is no in-flight
// transition when all phases are complete.
func TestDetectInFlightTransition_AllPhasesComplete(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	specID := "SPEC-COMPLETE-001"

	// plan, run, and sync are all complete.
	phases := []struct {
		phase Phase
		cp    Checkpoint
	}{
		{PhasePlan, &PlanCheckpoint{SPECID: specID, Status: "approved", ResearchPath: "/r"}},
		{PhaseRun, &RunCheckpoint{SPECID: specID, Status: "pass", Harness: "standard"}},
		{PhaseSync, &SyncCheckpoint{SPECID: specID, DocsSynced: true}},
	}

	for _, p := range phases {
		state := PhaseState{
			Phase:      p.phase,
			SPECID:     specID,
			Checkpoint: p.cp,
			UpdatedAt:  time.Now(),
			Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
		}
		if err := store.Checkpoint(state); err != nil {
			t.Fatalf("Checkpoint(%v) 실패: %v", p.phase, err)
		}
	}

	_, _, found, err := store.DetectInFlightTransition(specID)
	if err != nil {
		t.Fatalf("DetectInFlightTransition() 에러: %v", err)
	}
	if found {
		t.Error("모든 phase 완료 후 in-flight transition이 감지되지 않아야 함")
	}
}

// TestDetectInFlightTransition_NoCheckpoint returns found=false when no
// checkpoint exists.
func TestDetectInFlightTransition_NoCheckpoint(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	_, _, found, err := store.DetectInFlightTransition("NONEXISTENT-SPEC")
	if err != nil {
		t.Fatalf("DetectInFlightTransition() 에러: %v", err)
	}
	if found {
		t.Error("체크포인트 없을 때 found=false이어야 함")
	}
}

// TestDetectInFlightTransition_DifferentSpec ignores checkpoints with a
// different SPEC ID.
func TestDetectInFlightTransition_DifferentSpec(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Plan checkpoint for a different SPEC.
	planState := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-OTHER-001",
		Checkpoint: &PlanCheckpoint{
			SPECID:       "SPEC-OTHER-001",
			Status:       "approved",
			ResearchPath: "/research",
		},
		UpdatedAt:  time.Now(),
		Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
	}
	if err := store.Checkpoint(planState); err != nil {
		t.Fatalf("Checkpoint() 실패: %v", err)
	}

	// Querying a different SPEC must return not found.
	_, _, found, err := store.DetectInFlightTransition("SPEC-TARGET-002")
	if err != nil {
		t.Fatalf("DetectInFlightTransition() 에러: %v", err)
	}
	if found {
		t.Error("다른 SPEC ID 체크포인트는 무시되어야 함")
	}
}

// TestCheckpoint_BlockerOutstanding_FileScan scans on-disk blocker files and
// returns ErrBlockerOutstanding when unresolved blockers exist.
// SPEC-V3R2-RT-004 AC-04: blocker-file scan (file scan, not inline ref).
func TestCheckpoint_BlockerOutstanding_FileScan(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Write a blocker file to disk (without inline BlockerRpt).
	blockerPath := filepath.Join(tempDir, "blocker-run-SPEC-SCAN-001-20260101-000000.json")
	blockerContent := `{
		"kind": "error",
		"message": "test failed",
		"requested_action": "fix_test",
		"provenance": {"source": "session", "origin": "cli", "loaded": "2026-01-01T00:00:00Z"},
		"resolved": false,
		"timestamp": "2026-01-01T00:00:00Z",
		"phase": "run",
		"spec_id": "SPEC-SCAN-001"
	}`
	if err := os.WriteFile(blockerPath, []byte(blockerContent), 0644); err != nil {
		t.Fatalf("blocker 파일 작성 실패: %v", err)
	}

	// Checkpoint attempt without inline BlockerRpt.
	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-SCAN-001",
		Checkpoint: &RunCheckpoint{
			SPECID:  "SPEC-SCAN-001",
			Status:  "pass",
			Harness: "standard",
		},
		UpdatedAt:  time.Now(),
		Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
	}

	err := store.Checkpoint(state)
	if err == nil {
		t.Error("미해결 blocker 파일이 있을 때 ErrBlockerOutstanding을 반환해야 함")
		return
	}
	if err != ErrBlockerOutstanding {
		t.Errorf("에러 = %v, want ErrBlockerOutstanding", err)
	}
}

// TestCheckpoint_AfterBlockerResolved_FileScan verifies that Checkpoint
// succeeds after the blocker is resolved.
func TestCheckpoint_AfterBlockerResolved_FileScan(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Write a resolved blocker file.
	blockerPath := filepath.Join(tempDir, "blocker-run-SPEC-RESOLVED-001-20260101-000000.json")
	blockerContent := `{
		"kind": "error",
		"message": "test failed",
		"requested_action": "fix_test",
		"provenance": {"source": "session", "origin": "cli", "loaded": "2026-01-01T00:00:00Z"},
		"resolved": true,
		"resolution": "fixed",
		"timestamp": "2026-01-01T00:00:00Z",
		"phase": "run",
		"spec_id": "SPEC-RESOLVED-001"
	}`
	if err := os.WriteFile(blockerPath, []byte(blockerContent), 0644); err != nil {
		t.Fatalf("blocker 파일 작성 실패: %v", err)
	}

	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-RESOLVED-001",
		Checkpoint: &RunCheckpoint{
			SPECID:  "SPEC-RESOLVED-001",
			Status:  "pass",
			Harness: "standard",
		},
		UpdatedAt:  time.Now(),
		Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
	}

	if err := store.Checkpoint(state); err != nil {
		t.Errorf("해결된 blocker 후 Checkpoint() 실패: %v", err)
	}
}
