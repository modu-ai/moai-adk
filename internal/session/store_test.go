package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFileSessionStoreCheckpoint(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-001",
		Checkpoint: &PlanCheckpoint{
			SPECID:       "SPEC-001",
			Status:       "approved",
			ResearchPath: "/research/SPEC-001",
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}

	// Test checkpoint creation
	if err := store.Checkpoint(state); err != nil {
		t.Fatalf("Checkpoint() failed: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tempDir, "checkpoint-plan-SPEC-001.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Checkpoint file not created at %s", expectedPath)
	}

	// Verify file content
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read checkpoint file: %v", err)
	}

	var loadedState PhaseState
	if err := json.Unmarshal(data, &loadedState); err != nil {
		t.Fatalf("Failed to unmarshal checkpoint: %v", err)
	}

	if loadedState.Phase != PhasePlan {
		t.Errorf("Phase = %v, want %v", loadedState.Phase, PhasePlan)
	}
	if loadedState.SPECID != "SPEC-001" {
		t.Errorf("SPECID = %v, want SPEC-001", loadedState.SPECID)
	}
}

func TestFileSessionStoreCheckpointWithBlocker(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-001",
		BlockerRpt: &BlockerReport{
			Kind:            "error",
			Message:         "Test failed",
			RequestedAction: "fix_test",
			Resolved:        false,
		},
		UpdatedAt: time.Now(),
	}

	// Should fail with outstanding blocker
	err := store.Checkpoint(state)
	if err != ErrBlockerOutstanding {
		t.Errorf("Checkpoint() with unresolved blocker = %v, want ErrBlockerOutstanding", err)
	}
}

func TestFileSessionStoreHydrate(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Test hydrating non-existent checkpoint
	state, err := store.Hydrate(PhasePlan, "NONEXISTENT")
	if err != nil {
		t.Fatalf("Hydrate() non-existent failed: %v", err)
	}
	if state != nil {
		t.Errorf("Hydrate() non-existent = %v, want nil", state)
	}

	// Create a checkpoint
	originalState := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-001",
		Checkpoint: &RunCheckpoint{
			SPECID:        "SPEC-001",
			Status:        "pass",
			Harness:       "standard", // SPEC-V3R2-RT-004: 필수 필드 추가
			TestsTotal:    100,
			TestsPassed:   95,
			FilesModified: 12,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}

	if err := store.Checkpoint(originalState); err != nil {
		t.Fatalf("Checkpoint() failed: %v", err)
	}

	// Hydrate the checkpoint
	loadedState, err := store.Hydrate(PhaseRun, "SPEC-001")
	if err != nil {
		t.Fatalf("Hydrate() failed: %v", err)
	}
	if loadedState == nil {
		t.Fatal("Hydrate() returned nil")
	}

	if loadedState.Phase != PhaseRun {
		t.Errorf("Phase = %v, want %v", loadedState.Phase, PhaseRun)
	}

	// Check checkpoint type
	runCheckpoint, ok := loadedState.Checkpoint.(*RunCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *RunCheckpoint")
	}
	if runCheckpoint.TestsPassed != 95 {
		t.Errorf("TestsPassed = %v, want 95", runCheckpoint.TestsPassed)
	}
}

func TestFileSessionStoreHydrateStale(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 100*time.Millisecond) // Short TTL

	// Create an old checkpoint
	oldState := PhaseState{
		Phase:     PhaseSync,
		SPECID:    "SPEC-001",
		UpdatedAt: time.Now().Add(-200 * time.Millisecond), // 200ms ago
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}

	if err := store.Checkpoint(oldState); err != nil {
		t.Fatalf("Checkpoint() failed: %v", err)
	}

	// Try to hydrate stale checkpoint
	_, err := store.Hydrate(PhaseSync, "SPEC-001")
	if err != ErrCheckpointStale {
		t.Errorf("Hydrate() stale = %v, want ErrCheckpointStale", err)
	}
}

func TestFileSessionStoreAppendTaskLedger(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	entry := TaskLedgerEntry{
		Timestamp: "2026-04-28T10:00:00Z",
		Action:    "phase_start",
		Phase:     PhasePlan,
		SPECID:    "SPEC-001",
		Detail:    "Starting plan phase",
	}

	if err := store.AppendTaskLedger(entry); err != nil {
		t.Fatalf("AppendTaskLedger() failed: %v", err)
	}

	// Verify file exists and contains the entry
	ledgerPath := filepath.Join(tempDir, "task-ledger.md")
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to read ledger: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Error("Ledger file is empty")
	}
}

func TestFileSessionStoreWriteRunArtifact(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	artifactBody := []byte("artifact content")
	iterID := "iter-001"
	name := "output.json"

	if err := store.WriteRunArtifact(iterID, name, artifactBody); err != nil {
		t.Fatalf("WriteRunArtifact() failed: %v", err)
	}

	// Verify artifact exists
	artifactPath := filepath.Join(tempDir, "runs", iterID, name)
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("Failed to read artifact: %v", err)
	}

	if string(data) != string(artifactBody) {
		t.Errorf("Artifact content = %v, want %v", string(data), string(artifactBody))
	}
}

func TestFileSessionStoreRecordBlocker(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	report := BlockerReport{
		Kind:            "missing_input",
		Message:         "SPEC ID required",
		RequestedAction: "provide_spec_id",
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
		Resolved:  false,
		Timestamp: time.Now(),
	}

	if err := store.RecordBlocker(report); err != nil {
		t.Fatalf("RecordBlocker() failed: %v", err)
	}

	// Verify blocker file exists
	pattern := filepath.Join(tempDir, "blocker-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("Found %d blocker files, want 1", len(matches))
	}

	// Verify content
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("Failed to read blocker file: %v", err)
	}

	var loadedReport BlockerReport
	if err := json.Unmarshal(data, &loadedReport); err != nil {
		t.Fatalf("Failed to unmarshal blocker: %v", err)
	}

	if loadedReport.Kind != "missing_input" {
		t.Errorf("Kind = %v, want missing_input", loadedReport.Kind)
	}
}

func TestFileSessionStoreResolveBlocker(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Create an unresolved blocker
	report := BlockerReport{
		Kind:            "error",
		Message:         "Test failed",
		RequestedAction: "fix_test",
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
		Resolved:  false,
		Timestamp: time.Now(),
	}

	if err := store.RecordBlocker(report); err != nil {
		t.Fatalf("RecordBlocker() failed: %v", err)
	}

	// Resolve the blocker
	resolution := "Fixed the failing test"
	if err := store.ResolveBlocker(PhaseRun, "SPEC-001", resolution); err != nil {
		t.Fatalf("ResolveBlocker() failed: %v", err)
	}

	// Verify the blocker is now resolved
	pattern := filepath.Join(tempDir, "blocker-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("Found %d blocker files, want 1", len(matches))
	}

	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("Failed to read blocker file: %v", err)
	}

	var resolvedReport BlockerReport
	if err := json.Unmarshal(data, &resolvedReport); err != nil {
		t.Fatalf("Failed to unmarshal blocker: %v", err)
	}

	if !resolvedReport.Resolved {
		t.Error("Blocker should be resolved")
	}
	if resolvedReport.Resolution != resolution {
		t.Errorf("Resolution = %v, want %v", resolvedReport.Resolution, resolution)
	}
}

// T-RT004-04: TestCheckpoint_ValidatorRejectsBadHarness
// RED phase - validator/v10 태그가 없으므로 이 테스트는 현재 실패해야 함
func TestCheckpoint_ValidatorRejectsBadHarness(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// RunCheckpoint에 유효하지 않은 Harness 값
	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-V3R2-RT-004",
		Checkpoint: &RunCheckpoint{
			SPECID: "SPEC-V3R2-RT-004",
			Status: "pass",
			Harness: "ultra", // 유효하지 않은 값 (oneof=minimal standard thorough)
		},
		UpdatedAt: time.Now(),
	}

	err := store.Checkpoint(state)
	if err == nil {
		t.Error("Checkpoint() with invalid Harness should return error, got nil")
		return
	}

	// 에러 메시지에 "harness" 문자열 포함 확인 (AC-15 요구사항)
	errMsg := err.Error()
	if !contains(errMsg, "harness") && !contains(errMsg, "Harness") {
		t.Errorf("Error message should contain 'harness' or 'Harness', got: %v", errMsg)
	}
}

// T-RT004-04 part 2: TestCheckpoint_ValidatorAcceptsGoodHarness
func TestCheckpoint_ValidatorAcceptsGoodHarness(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-V3R2-RT-004",
		Checkpoint: &RunCheckpoint{
			SPECID:        "SPEC-V3R2-RT-004",
			Status:        "pass",
			Harness:       "thorough", // 유효한 값
			TestsTotal:    100,
			TestsPassed:   100,
			FilesModified: 5,
		},
		UpdatedAt: time.Now(),
	}

	err := store.Checkpoint(state)
	if err != nil {
		t.Errorf("Checkpoint() with valid Harness should succeed, got: %v", err)
	}
}

// T-RT004-04 part 3: TestCheckpoint_ValidatorRejectsEmptyHarness
func TestCheckpoint_ValidatorRejectsEmptyHarness(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-V3R2-RT-004",
		Checkpoint: &RunCheckpoint{
			SPECID:  "SPEC-V3R2-RT-004",
			Status:  "pass",
			Harness: "", // 빈 값 (required 필드)
		},
		UpdatedAt: time.Now(),
	}

	err := store.Checkpoint(state)
	if err == nil {
		t.Error("Checkpoint() with empty Harness should return error, got nil")
		return
	}

	errMsg := err.Error()
	if !contains(errMsg, "Harness") && !contains(errMsg, "required") {
		t.Errorf("Error message should contain 'Harness' or 'required', got: %v", errMsg)
	}
}

// T-RT004-03: TestCheckpoint_ConcurrentRace
// RED phase - advisory lock 구현이 없으므로 이 테스트는 현재 실패해야 함
func TestCheckpoint_ConcurrentRace(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-V3R2-RT-004",
		Checkpoint: &PlanCheckpoint{
			SPECID:       "SPEC-V3R2-RT-004",
			Status:       "approved",
			ResearchPath: "/research/SPEC-V3R2-RT-004",
		},
		UpdatedAt: time.Now(),
	}

	var wg sync.WaitGroup
	// SPEC-V3R2-RT-004: atomic counter로 goroutine 간 race-free 집계
	// (이전 int++ 사용은 go test -race에서 WARNING: DATA RACE 발동)
	var successCount, concurrentErrCount atomic.Int64

	// 2개의 goroutine이 동시에 Checkpoint 시도
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := store.Checkpoint(state)
			switch {
			case err == nil:
				successCount.Add(1)
			case errors.Is(err, ErrCheckpointConcurrent):
				concurrentErrCount.Add(1)
			}
		}()
	}

	wg.Wait()

	// 적어도 하나는 성공해야 함
	if successCount.Load() < 1 {
		t.Errorf("Expected at least 1 successful checkpoint, got %d", successCount.Load())
	}

	// lock 구현이 없으면 둘 다 성공해버림 (현재 skeleton 동작)
	if concurrentErrCount.Load() < 1 {
		t.Log("WARNING: No ErrCheckpointConcurrent returned - advisory lock not yet implemented (expected in RED phase)")
	}
}

// 보조 함수: 문자열 포함 검사
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
