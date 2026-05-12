package session

import (
	"testing"
	"time"
)

// TestHydrateWithOpts_SkipStaleCheck은 SkipStaleCheck 옵션이 staleness 검사를
// 우회하는지 검증합니다. SPEC-V3R2-RT-004 AC-06: --resume 플래그 연동.
func TestHydrateWithOpts_SkipStaleCheck(t *testing.T) {
	tempDir := t.TempDir()
	// 매우 짧은 TTL로 즉시 stale 처리
	store := NewFileSessionStore(tempDir, 1*time.Millisecond)

	oldState := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-RESUME-001",
		Checkpoint: &RunCheckpoint{
			SPECID:  "SPEC-RESUME-001",
			Status:  "pass",
			Harness: "standard",
		},
		UpdatedAt: time.Now().Add(-1 * time.Hour), // 1시간 전 (stale)
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}

	// 스토어 생성 후 직접 파일 쓰기 (TTL 우회)
	storeFull := NewFileSessionStore(tempDir, 3600*time.Second)
	if err := storeFull.Checkpoint(oldState); err != nil {
		t.Fatalf("Checkpoint() 실패: %v", err)
	}

	// SkipStaleCheck=false이면 ErrCheckpointStale 반환해야 함
	_, err := store.HydrateWithOpts(PhaseRun, "SPEC-RESUME-001", HydrateOpts{SkipStaleCheck: false})
	if err != ErrCheckpointStale {
		t.Errorf("SkipStaleCheck=false 시 ErrCheckpointStale 기대, got: %v", err)
	}

	// SkipStaleCheck=true이면 성공해야 함
	state, err := store.HydrateWithOpts(PhaseRun, "SPEC-RESUME-001", HydrateOpts{SkipStaleCheck: true})
	if err != nil {
		t.Errorf("SkipStaleCheck=true 시 에러가 없어야 함, got: %v", err)
	}
	if state == nil {
		t.Error("SkipStaleCheck=true 시 nil이 아닌 state를 반환해야 함")
	}
}

// TestHydrateWithOpts_Default은 기본 옵션(SkipStaleCheck=false)이
// 기존 Hydrate와 동일하게 동작합니다.
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

	// 기본 옵션으로 Hydrate
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

// TestHydrateWithOpts_NotExists은 존재하지 않는 checkpoint에 대해 nil을 반환합니다.
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
