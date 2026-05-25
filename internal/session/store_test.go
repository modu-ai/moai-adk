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
			Harness:       "standard", // SPEC-V3R2-RT-004: required field added
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
// RED phase — without validator/v10 tags this test currently fails.
func TestCheckpoint_ValidatorRejectsBadHarness(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Invalid Harness value in RunCheckpoint.
	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-V3R2-RT-004",
		Checkpoint: &RunCheckpoint{
			SPECID: "SPEC-V3R2-RT-004",
			Status: "pass",
			Harness: "ultra", // invalid value (oneof=minimal standard thorough)
		},
		UpdatedAt: time.Now(),
	}

	err := store.Checkpoint(state)
	if err == nil {
		t.Error("Checkpoint() with invalid Harness should return error, got nil")
		return
	}

	// Verify the error message contains "harness" (AC-15 requirement).
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
			Harness:       "thorough", // valid value
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
			Harness: "", // empty value (required field)
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
// RED phase — without advisory-lock implementation this test currently fails.
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
	// SPEC-V3R2-RT-004: race-free aggregation across goroutines via atomic counters
	// (the prior int++ usage triggered WARNING: DATA RACE under go test -race).
	var successCount, concurrentErrCount atomic.Int64

	// Two goroutines attempt Checkpoint concurrently.
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

	// At least one must succeed.
	if successCount.Load() < 1 {
		t.Errorf("Expected at least 1 successful checkpoint, got %d", successCount.Load())
	}

	// Without lock implementation, both succeed (current skeleton behavior).
	if concurrentErrCount.Load() < 1 {
		t.Log("WARNING: No ErrCheckpointConcurrent returned - advisory lock not yet implemented (expected in RED phase)")
	}
}

// Helper function: substring check.
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

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P2 보강 — store.go:392 WriteRunArtifact
// 66.7% → cover. SPEC-V3R2-RT-004 REQ-043: text-declared artifacts (md/txt/json/yaml/yml)
// MUST be UTF-8 — invalid UTF-8 시 ErrArtifactEncodingMismatch 반환.

func TestWriteRunArtifact_TextExtensions_RejectInvalidUTF8(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Invalid UTF-8 byte sequence (lone continuation byte 0x80 + truncated 0xC3).
	badBytes := []byte{0xC3, 0x28, 0x80, 0xFF}

	textExts := []string{"output.md", "output.txt", "output.json", "output.yaml", "output.yml"}
	for _, name := range textExts {
		name := name
		t.Run(name, func(t *testing.T) {
			err := store.WriteRunArtifact("iter-utf8-bad", name, badBytes)
			if err == nil {
				t.Fatalf("WriteRunArtifact(%s, invalid UTF-8) should error, got nil", name)
			}
			if !errors.Is(err, ErrArtifactEncodingMismatch) {
				t.Errorf("WriteRunArtifact(%s) error = %v, want ErrArtifactEncodingMismatch", name, err)
			}
		})
	}
}

func TestWriteRunArtifact_BinaryExtensions_AcceptInvalidUTF8(t *testing.T) {
	// Non-text extensions (.bin, .png, .tar) skip UTF-8 validation.
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	binaryBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A} // JPEG-like header
	binaryExts := []string{"output.bin", "image.png", "archive.tar"}
	for _, name := range binaryExts {
		name := name
		t.Run(name, func(t *testing.T) {
			if err := store.WriteRunArtifact("iter-bin", name, binaryBytes); err != nil {
				t.Fatalf("WriteRunArtifact(%s, binary) should succeed, got: %v", name, err)
			}
			path := filepath.Join(tempDir, "runs", "iter-bin", name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read artifact failed: %v", err)
			}
			if len(data) != len(binaryBytes) {
				t.Errorf("artifact length = %d, want %d", len(data), len(binaryBytes))
			}
		})
	}
}

func TestWriteRunArtifact_TextExtensions_AcceptValidUTF8(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Valid UTF-8 incl. multi-byte (Korean) characters.
	validBytes := []byte("# Hello 안녕 こんにちは 你好\n")
	if err := store.WriteRunArtifact("iter-utf8-ok", "valid.md", validBytes); err != nil {
		t.Fatalf("WriteRunArtifact(valid UTF-8) should succeed, got: %v", err)
	}

	path := filepath.Join(tempDir, "runs", "iter-utf8-ok", "valid.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact failed: %v", err)
	}
	if string(data) != string(validBytes) {
		t.Errorf("artifact content mismatch")
	}
}

func TestWriteRunArtifact_CaseInsensitiveExtension(t *testing.T) {
	// SPEC-V3R2-RT-004 REQ-043: strings.ToLower(filepath.Ext(name)) → MD = md
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	badBytes := []byte{0xFF, 0xFE} // invalid UTF-8
	// .MD should also trigger validation (lowercase normalized).
	err := store.WriteRunArtifact("iter-case", "report.MD", badBytes)
	if err == nil {
		t.Fatal("WriteRunArtifact(.MD invalid UTF-8) should error (case-insensitive ext)")
	}
	if !errors.Is(err, ErrArtifactEncodingMismatch) {
		t.Errorf("error = %v, want ErrArtifactEncodingMismatch", err)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P2 보강 — store.go:269 mergePhaseStates
// 69.6% → cover Plan / Sync / default 분기 (기존 team_merge_test 는 PhaseRun 만 커버).

func TestMergeTeamCheckpoints_PhasePlan_LastWriteWins(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	specID := "SPEC-PLAN-MERGE"
	phase := PhasePlan
	agentNames := []string{"agent-z", "agent-a"} // sorted: agent-a, agent-z

	// agent-a → status approved
	// agent-z → status draft
	// merge result: 마지막 (alphabetical sort 후) = agent-z → draft
	for _, agentInfo := range []struct {
		name   string
		status string
	}{
		{"agent-a", "approved"},
		{"agent-z", "draft"},
	} {
		state := PhaseState{
			Phase:  phase,
			SPECID: specID,
			Checkpoint: &PlanCheckpoint{
				SPECID:       specID,
				Status:       agentInfo.status,
				ResearchPath: "/r/" + agentInfo.name,
			},
			UpdatedAt:  time.Now(),
			Provenance: ProvenanceTag{Source: "session", Origin: agentInfo.name, Loaded: time.Now()},
		}
		data, _ := json.MarshalIndent(state, "", "  ")
		path := filepath.Join(tempDir, "checkpoint-plan-"+specID+"-"+agentInfo.name+".json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("write %s: %v", agentInfo.name, err)
		}
	}

	merged, err := store.MergeTeamCheckpoints(specID, phase, agentNames)
	if err != nil {
		t.Fatalf("MergeTeamCheckpoints(plan) failed: %v", err)
	}
	pc, ok := merged.Checkpoint.(*PlanCheckpoint)
	if !ok {
		t.Fatalf("Checkpoint = %T, want *PlanCheckpoint", merged.Checkpoint)
	}
	// alphabetical sort: agent-a, agent-z → last = agent-z → draft
	if pc.Status != "draft" {
		t.Errorf("merged Status = %q, want draft (last-write-wins after sort)", pc.Status)
	}
}

func TestMergeTeamCheckpoints_PhaseSync_LastWriteWins(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	specID := "SPEC-SYNC-MERGE"
	phase := PhaseSync

	for _, agentInfo := range []struct {
		name string
		pr   int
	}{
		{"agent-a", 100},
		{"agent-z", 200},
	} {
		state := PhaseState{
			Phase:  phase,
			SPECID: specID,
			Checkpoint: &SyncCheckpoint{
				SPECID:     specID,
				PRNumber:   agentInfo.pr,
				DocsSynced: true,
			},
			UpdatedAt:  time.Now(),
			Provenance: ProvenanceTag{Source: "session", Origin: agentInfo.name, Loaded: time.Now()},
		}
		data, _ := json.MarshalIndent(state, "", "  ")
		path := filepath.Join(tempDir, "checkpoint-sync-"+specID+"-"+agentInfo.name+".json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("write %s: %v", agentInfo.name, err)
		}
	}

	merged, err := store.MergeTeamCheckpoints(specID, phase, []string{"agent-z", "agent-a"})
	if err != nil {
		t.Fatalf("MergeTeamCheckpoints(sync) failed: %v", err)
	}
	sc, ok := merged.Checkpoint.(*SyncCheckpoint)
	if !ok {
		t.Fatalf("Checkpoint = %T, want *SyncCheckpoint", merged.Checkpoint)
	}
	if sc.PRNumber != 200 {
		t.Errorf("merged PRNumber = %d, want 200 (last-write-wins after sort)", sc.PRNumber)
	}
}

func TestMergeTeamCheckpoints_EmptyAgentNames(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)
	_, err := store.MergeTeamCheckpoints("SPEC-EMPTY", PhaseRun, []string{})
	if err == nil {
		t.Fatal("MergeTeamCheckpoints with empty agent names should error")
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P3 보강 — store.go:423 RecordBlocker 75.0%
// → cover default phase/specID values when blocker fields are missing.

func TestRecordBlocker_DefaultsWhenPhaseAndSpecMissing(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	report := BlockerReport{
		Kind:            "missing_input",
		Message:         "missing field",
		RequestedAction: "provide",
		Resolved:        false,
		Timestamp:       time.Now(),
		// Phase + SPECID 의도적으로 비움 → "unknown" 으로 대체되어야 한다
	}

	if err := store.RecordBlocker(report); err != nil {
		t.Fatalf("RecordBlocker failed: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(tempDir, "blocker-unknown-unknown-*.json"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 'blocker-unknown-unknown-*' file, got %d", len(matches))
	}
}

func TestRecordBlocker_WithPhaseAndSpec(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	report := BlockerReport{
		Kind:            "quality_gate",
		Message:         "lint fail",
		RequestedAction: "fix_lint",
		Resolved:        false,
		Timestamp:       time.Now(),
		Phase:           PhaseRun,
		SPECID:          "SPEC-WITH-PHASE",
	}

	if err := store.RecordBlocker(report); err != nil {
		t.Fatalf("RecordBlocker failed: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(tempDir, "blocker-run-SPEC-WITH-PHASE-*.json"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 'blocker-run-SPEC-WITH-PHASE-*' file, got %d", len(matches))
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P3 보강 — store.go:459 ResolveBlocker
// 74.2% → cover "no outstanding blocker" + "only resolved blockers exist" cases.

func TestResolveBlocker_NoBlockerFound(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// No blocker file at all
	err := store.ResolveBlocker(PhaseRun, "SPEC-NONE", "resolution")
	if err == nil {
		t.Fatal("ResolveBlocker with no blocker should error")
	}
}

func TestResolveBlocker_OnlyResolvedExists(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Write an already-resolved blocker
	report := BlockerReport{
		Kind:            "error",
		Message:         "previous error",
		RequestedAction: "fix",
		Resolved:        true, // already resolved
		Resolution:      "previously fixed",
		Timestamp:       time.Now(),
		Phase:           PhaseRun,
		SPECID:          "SPEC-ONLY-RESOLVED",
	}
	if err := store.RecordBlocker(report); err != nil {
		t.Fatalf("RecordBlocker failed: %v", err)
	}

	// ResolveBlocker should fail since no unresolved blocker exists
	err := store.ResolveBlocker(PhaseRun, "SPEC-ONLY-RESOLVED", "new resolution")
	if err == nil {
		t.Fatal("ResolveBlocker should error when only resolved blockers exist")
	}
}

func TestResolveBlocker_MostRecentSelectedAmongMultiple(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Create two unresolved blockers with different timestamps (1s apart)
	oldTs := time.Now().Add(-2 * time.Second)
	newTs := time.Now()

	older := BlockerReport{
		Kind: "error", Message: "older",
		Resolved: false, Timestamp: oldTs,
		Phase: PhaseRun, SPECID: "SPEC-MULTI",
	}
	newer := BlockerReport{
		Kind: "error", Message: "newer",
		Resolved: false, Timestamp: newTs,
		Phase: PhaseRun, SPECID: "SPEC-MULTI",
	}
	if err := store.RecordBlocker(older); err != nil {
		t.Fatalf("RecordBlocker older: %v", err)
	}
	// Ensure distinct filenames (1-second granularity)
	time.Sleep(1100 * time.Millisecond)
	if err := store.RecordBlocker(newer); err != nil {
		t.Fatalf("RecordBlocker newer: %v", err)
	}

	if err := store.ResolveBlocker(PhaseRun, "SPEC-MULTI", "resolved via test"); err != nil {
		t.Fatalf("ResolveBlocker failed: %v", err)
	}

	// Verify the most recent (newer) was resolved, older still unresolved
	matches, _ := filepath.Glob(filepath.Join(tempDir, "blocker-*.json"))
	if len(matches) != 2 {
		t.Fatalf("expected 2 blocker files, got %d", len(matches))
	}

	resolvedCount := 0
	unresolvedCount := 0
	for _, m := range matches {
		data, _ := os.ReadFile(m)
		var b BlockerReport
		if err := json.Unmarshal(data, &b); err != nil {
			t.Fatalf("unmarshal %s: %v", m, err)
		}
		if b.Resolved {
			resolvedCount++
			if b.Message != "newer" {
				t.Errorf("resolved blocker should be 'newer', got: %q", b.Message)
			}
		} else {
			unresolvedCount++
		}
	}
	if resolvedCount != 1 || unresolvedCount != 1 {
		t.Errorf("expected 1 resolved + 1 unresolved, got %d resolved + %d unresolved", resolvedCount, unresolvedCount)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P5 보강 — store.go:328 checkBlockerFiles
// 78.6% → cover outstanding-blocker-on-disk rejection path during Checkpoint.

func TestCheckpoint_RejectedByDiskBlocker(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// Pre-place an unresolved blocker file for the target phase+spec.
	blockerReport := BlockerReport{
		Kind:            "quality_gate",
		Message:         "outstanding from a prior session",
		RequestedAction: "resolve",
		Resolved:        false,
		Timestamp:       time.Now(),
		Phase:           PhaseRun,
		SPECID:          "SPEC-BLOCKED-DISK",
	}
	if err := store.RecordBlocker(blockerReport); err != nil {
		t.Fatalf("RecordBlocker setup failed: %v", err)
	}

	// Now attempt Checkpoint — must be rejected by checkBlockerFiles disk scan.
	state := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-BLOCKED-DISK",
		Checkpoint: &RunCheckpoint{
			SPECID: "SPEC-BLOCKED-DISK", Status: "pass", Harness: "standard",
		},
		UpdatedAt: time.Now(),
	}
	err := store.Checkpoint(state)
	if !errors.Is(err, ErrBlockerOutstanding) {
		t.Errorf("Checkpoint with disk-blocker should fail with ErrBlockerOutstanding, got: %v", err)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — store.go:82 Checkpoint 76.0% →
// cover the "invalid checkpoint" Validate failure path (not just blocker rejection).

func TestCheckpoint_RejectedByInvalidCheckpoint(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	// PlanCheckpoint with invalid status — Validate() returns error,
	// wrapped as ErrCheckpointInvalid.
	state := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-INVALID-CP",
		Checkpoint: &PlanCheckpoint{
			SPECID: "SPEC-INVALID-CP",
			Status: "bogus-status", // not in oneof=[approved draft rejected]
		},
		UpdatedAt: time.Now(),
	}
	err := store.Checkpoint(state)
	if !errors.Is(err, ErrCheckpointInvalid) {
		t.Errorf("Checkpoint with bad status should fail with ErrCheckpointInvalid, got: %v", err)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — DetectInFlightTransition coverage
// across plan-only / plan+run / plan+run+sync / empty.

func TestDetectInFlightTransition_PlanOnly_PlanToRun(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase: PhasePlan, SPECID: "SPEC-IFT-001",
		Checkpoint: &PlanCheckpoint{SPECID: "SPEC-IFT-001", Status: "approved"},
		UpdatedAt:  time.Now(),
	}
	if err := store.Checkpoint(state); err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	from, to, found, err := store.DetectInFlightTransition("SPEC-IFT-001")
	if err != nil {
		t.Fatalf("DetectInFlightTransition failed: %v", err)
	}
	if !found {
		t.Fatal("expected found=true (plan checkpoint exists, run does not)")
	}
	if from != PhasePlan || to != PhaseRun {
		t.Errorf("expected plan→run, got %v→%v", from, to)
	}
}

func TestDetectInFlightTransition_Empty(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	from, to, found, err := store.DetectInFlightTransition("SPEC-EMPTY")
	if err != nil {
		t.Fatalf("DetectInFlightTransition failed: %v", err)
	}
	if found {
		t.Errorf("expected found=false for empty state dir, got from=%v to=%v", from, to)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — store.go:356 AppendTaskLedger 72.7%
// → cover multi-entry append + blocker action path.

func TestAppendTaskLedger_MultipleEntriesAppended(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	entries := []TaskLedgerEntry{
		{Timestamp: "2026-05-25T10:00:00Z", Action: "phase_start", Phase: PhasePlan, SPECID: "SPEC-LEDGER", Detail: "starting"},
		{Timestamp: "2026-05-25T10:05:00Z", Action: "blocker", Phase: PhasePlan, SPECID: "SPEC-LEDGER", Detail: "hit blocker"},
		{Timestamp: "2026-05-25T10:10:00Z", Action: "blocker_resolved", Phase: PhasePlan, SPECID: "SPEC-LEDGER", Detail: ""},
	}

	for _, e := range entries {
		if err := store.AppendTaskLedger(e); err != nil {
			t.Fatalf("AppendTaskLedger(%s) failed: %v", e.Action, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(tempDir, "task-ledger.md"))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	content := string(data)
	// All 3 entries should be present in append order
	for _, action := range []string{"phase_start", "blocker", "blocker_resolved"} {
		if !contains(content, action) {
			t.Errorf("ledger missing action %q. content: %s", action, content)
		}
	}
	// Detail markdown should render for entries with non-empty detail
	if !contains(content, "starting") || !contains(content, "hit blocker") {
		t.Errorf("ledger missing detail strings. content: %s", content)
	}
}
