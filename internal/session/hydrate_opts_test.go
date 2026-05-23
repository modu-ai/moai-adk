package session

import (
	"testing"
	"time"
)

// TestHydrateWithOpts_SkipStaleCheck verifies that the SkipStaleCheck option
// bypasses the staleness check. SPEC-V3R2-RT-004 AC-06: --resume flag integration.
func TestHydrateWithOpts_SkipStaleCheck(t *testing.T) {
	tempDir := t.TempDir()
	// Use a very short TTL to mark as stale immediately.
	store := NewFileSessionStore(tempDir, 1*time.Millisecond)

	oldState := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-RESUME-001",
		Checkpoint: &RunCheckpoint{
			SPECID:  "SPEC-RESUME-001",
			Status:  "pass",
			Harness: "standard",
		},
		UpdatedAt: time.Now().Add(-1 * time.Hour), // 1 hour ago (stale)
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}

	// After store creation, write directly (bypasses TTL).
	storeFull := NewFileSessionStore(tempDir, 3600*time.Second)
	if err := storeFull.Checkpoint(oldState); err != nil {
		t.Fatalf("Checkpoint() 실패: %v", err)
	}

	// SkipStaleCheck=false must return ErrCheckpointStale.
	_, err := store.HydrateWithOpts(PhaseRun, "SPEC-RESUME-001", HydrateOpts{SkipStaleCheck: false})
	if err != ErrCheckpointStale {
		t.Errorf("SkipStaleCheck=false 시 ErrCheckpointStale 기대, got: %v", err)
	}

	// SkipStaleCheck=true must succeed.
	state, err := store.HydrateWithOpts(PhaseRun, "SPEC-RESUME-001", HydrateOpts{SkipStaleCheck: true})
	if err != nil {
		t.Errorf("SkipStaleCheck=true 시 에러가 없어야 함, got: %v", err)
	}
	if state == nil {
		t.Error("SkipStaleCheck=true 시 nil이 아닌 state를 반환해야 함")
	}
}

// TestHydrateWithOpts_Default verifies that default options (SkipStaleCheck=false)
// behave identically to the existing Hydrate.
func TestHydrateWithOpts_Default(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-OPTS-001",
		Checkpoint: &PlanCheckpoint{
			SPECID:       "SPEC-OPTS-001",
			Status:       "approved",
			ResearchPath: "/research",
		},
		UpdatedAt:  time.Now(),
		Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
	}

	if err := store.Checkpoint(state); err != nil {
		t.Fatalf("Checkpoint() 실패: %v", err)
	}

	// Hydrate with default options.
	loaded, err := store.HydrateWithOpts(PhasePlan, "SPEC-OPTS-001", HydrateOpts{})
	if err != nil {
		t.Fatalf("HydrateWithOpts() 실패: %v", err)
	}
	if loaded == nil {
		t.Fatal("nil이 아닌 state를 반환해야 함")
	}
	if loaded.SPECID != "SPEC-OPTS-001" {
		t.Errorf("SPECID = %v, want SPEC-OPTS-001", loaded.SPECID)
	}
}

// TestHydrateWithOpts_NotExists returns nil for non-existent checkpoints.
func TestHydrateWithOpts_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	state, err := store.HydrateWithOpts(PhasePlan, "SPEC-NONE", HydrateOpts{})
	if err != nil {
		t.Fatalf("HydrateWithOpts() 에러: %v", err)
	}
	if state != nil {
		t.Error("존재하지 않는 checkpoint는 nil을 반환해야 함")
	}
}
