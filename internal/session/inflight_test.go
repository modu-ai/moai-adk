package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDetectInFlightTransition은 진행 중인 phase 전환을 감지합니다.
// SPEC-V3R2-RT-004 AC-14: DetectInFlightTransition 메서드 검증.
func TestDetectInFlightTransition(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// plan checkpoint만 존재 (run checkpoint 없음)
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

	// in-flight 감지 — plan만 있으면 plan→run 전환 중으로 감지해야 함
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

// TestDetectInFlightTransition_AllPhasesComplete는 모든 phase가 완료되면
// in-flight 전환이 없음을 검증합니다.
func TestDetectInFlightTransition_AllPhasesComplete(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	specID := "SPEC-COMPLETE-001"

	// plan, run, sync 모두 완료
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

// TestDetectInFlightTransition_NoCheckpoint는 체크포인트가 없을 때
// found=false를 반환합니다.
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

// TestDetectInFlightTransition_DifferentSpec은 다른 SPEC ID의 checkpoint를
// 무시합니다.
func TestDetectInFlightTransition_DifferentSpec(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// 다른 SPEC의 plan checkpoint
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

	// 다른 SPEC를 조회하면 not found이어야 함
	_, _, found, err := store.DetectInFlightTransition("SPEC-TARGET-002")
	if err != nil {
		t.Fatalf("DetectInFlightTransition() 에러: %v", err)
	}
	if found {
		t.Error("다른 SPEC ID 체크포인트는 무시되어야 함")
	}
}

// TestCheckpoint_BlockerOutstanding_FileScan은 디스크의 blocker 파일을 스캔하여
// 미해결 blocker가 있을 때 ErrBlockerOutstanding을 반환합니다.
// SPEC-V3R2-RT-004 AC-04: blocker-file scan (inline ref가 아닌 파일 스캔).
func TestCheckpoint_BlockerOutstanding_FileScan(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// 디스크에 blocker 파일 작성 (inline BlockerRpt 없이)
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

	// inline BlockerRpt 없이 Checkpoint 시도
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

// TestCheckpoint_AfterBlockerResolved_FileScan은 blocker가 해결된 후
// Checkpoint가 성공하는지 검증합니다.
func TestCheckpoint_AfterBlockerResolved_FileScan(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// 해결된 blocker 파일 작성
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
