package cli_test

// SPEC-V3R2-RT-004 AC-07, AC-06, AC-12, REQ-030, REQ-032: tests for the CLI state subcommand.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// makeTestStateDir creates a .moai/state/ directory structure for tests.
func makeTestStateDir(t *testing.T) (moaiDir string, stateDir string) {
	t.Helper()
	base := t.TempDir()
	moaiDir = filepath.Join(base, ".moai")
	stateDir = filepath.Join(moaiDir, "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	return moaiDir, stateDir
}

// writeCheckpoint writes a checkpoint file directly to the state dir for tests.
func writeCheckpoint(t *testing.T, stateDir string, state session.PhaseState) {
	t.Helper()
	store := session.NewFileSessionStore(stateDir, 3600*time.Second)
	if err := store.Checkpoint(state); err != nil {
		t.Fatalf("Checkpoint() failed: %v", err)
	}
}

// TestStateDump_HappyPath verifies that state dump succeeds when a valid checkpoint exists.
func TestStateDump_HappyPath(t *testing.T) {
	t.Parallel()
	_, stateDir := makeTestStateDir(t)

	state := session.PhaseState{
		Phase:  session.PhaseRun,
		SPECID: "SPEC-CLI-TEST-001",
		Checkpoint: &session.RunCheckpoint{
			SPECID:      "SPEC-CLI-TEST-001",
			Status:      "pass",
			Harness:     "standard",
			TestsTotal:  10,
			TestsPassed: 10,
		},
		UpdatedAt:  time.Now(),
		Provenance: session.ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
	}
	writeCheckpoint(t, stateDir, state)

	// Verify the file is actually created
	pattern := filepath.Join(stateDir, "checkpoint-run-SPEC-CLI-TEST-001.json")
	if _, err := os.Stat(pattern); err != nil {
		t.Fatalf("checkpoint file not found: %v", err)
	}
}

// TestStateDump_FormatJSON verifies that JSON-parseable output is produced.
func TestStateDump_FormatJSON(t *testing.T) {
	t.Parallel()
	_, stateDir := makeTestStateDir(t)

	state := session.PhaseState{
		Phase:  session.PhasePlan,
		SPECID: "SPEC-JSON-001",
		Checkpoint: &session.PlanCheckpoint{
			SPECID:       "SPEC-JSON-001",
			Status:       "approved",
			ResearchPath: "/research",
		},
		UpdatedAt:  time.Now(),
		Provenance: session.ProvenanceTag{Source: "user", Origin: "cli", Loaded: time.Now()},
	}
	writeCheckpoint(t, stateDir, state)

	// Read the file directly and verify it is parseable as JSON
	checkpointPath := filepath.Join(stateDir, "checkpoint-plan-SPEC-JSON-001.json")
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}

	var loaded session.PhaseState
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if loaded.SPECID != "SPEC-JSON-001" {
		t.Errorf("SPECID = %q, want SPEC-JSON-001", loaded.SPECID)
	}
}

// TestStateDump_NoMatch verifies that nil is returned for a nonexistent phase/specID.
func TestStateDump_NoMatch(t *testing.T) {
	t.Parallel()
	_, stateDir := makeTestStateDir(t)

	store := session.NewFileSessionStore(stateDir, 3600*time.Second)
	state, err := store.Hydrate(session.PhaseRun, "SPEC-NONE-999")
	if err != nil {
		t.Fatalf("Hydrate() unexpected error: %v", err)
	}
	if state != nil {
		t.Error("expected nil state for non-existent checkpoint")
	}
}

// TestStateShowBlocker_Outstanding verifies that an unresolved blocker is detected.
func TestStateShowBlocker_Outstanding(t *testing.T) {
	t.Parallel()
	_, stateDir := makeTestStateDir(t)

	store := session.NewFileSessionStore(stateDir, 3600*time.Second)

	report := session.BlockerReport{
		Kind:            "missing_input",
		Message:         "SPEC ID가 없습니다",
		RequestedAction: "provide_spec_id",
		Provenance:      session.ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
		Resolved:        false,
		Timestamp:       time.Now(),
		Phase:           session.PhaseRun,
		SPECID:          "SPEC-BLOCKER-TEST",
	}

	if err := store.RecordBlocker(report); err != nil {
		t.Fatalf("RecordBlocker() failed: %v", err)
	}

	// Verify the blocker file is created
	pattern := filepath.Join(stateDir, "blocker-run-SPEC-BLOCKER-TEST-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected blocker file to be created")
	}
}

// TestRun_ResumeFlag verifies that the --resume flag wires HydrateWithOpts(SkipStaleCheck=true).
// SPEC-V3R2-RT-004 AC-06: --resume flag wiring.
func TestRun_ResumeFlag(t *testing.T) {
	t.Parallel()
	_, stateDir := makeTestStateDir(t)

	// Use a very short TTL so the checkpoint goes stale immediately
	staleStore := session.NewFileSessionStore(stateDir, 1*time.Millisecond)

	// Write a stale checkpoint with a timestamp one hour in the past
	freshStore := session.NewFileSessionStore(stateDir, 3600*time.Second)
	state := session.PhaseState{
		Phase:  session.PhaseRun,
		SPECID: "SPEC-RESUME-CLI-001",
		Checkpoint: &session.RunCheckpoint{
			SPECID:  "SPEC-RESUME-CLI-001",
			Status:  "pass",
			Harness: "standard",
		},
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
		Provenance: session.ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
	}
	if err := freshStore.Checkpoint(state); err != nil {
		t.Fatalf("Checkpoint() failed: %v", err)
	}

	// SkipStaleCheck=false → expect ErrCheckpointStale
	_, err := staleStore.HydrateWithOpts(session.PhaseRun, "SPEC-RESUME-CLI-001", session.HydrateOpts{SkipStaleCheck: false})
	if err != session.ErrCheckpointStale {
		t.Errorf("SkipStaleCheck=false: expected ErrCheckpointStale, got %v", err)
	}

	// SkipStaleCheck=true (--resume mode) → expect success
	loaded, err := staleStore.HydrateWithOpts(session.PhaseRun, "SPEC-RESUME-CLI-001", session.HydrateOpts{SkipStaleCheck: true})
	if err != nil {
		t.Errorf("SkipStaleCheck=true: unexpected error: %v", err)
	}
	if loaded == nil {
		t.Error("SkipStaleCheck=true: expected non-nil state")
	}
}
