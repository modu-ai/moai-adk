package session

import (
	"strings"
	"testing"
	"time"
)

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — hydrate.go HydrateForPrompt + checkpointStatus
// 두 함수가 baseline 0% coverage 였음. cache-prefix DO-NOT-REORDER 계약 + 모든 phase
// status 분기를 검증한다. fake store 사용 (외부 io 없이 in-memory state 주입).

// fakeSessionStore implements SessionStore for the hydrate tests. Only Hydrate
// is meaningfully used by HydrateForPrompt; the other methods return safe
// defaults to satisfy the interface contract.
type fakeSessionStore struct {
	state    *PhaseState
	hydrateErr error
}

func (f *fakeSessionStore) Checkpoint(state PhaseState) error                      { return nil }
func (f *fakeSessionStore) Hydrate(phase Phase, specID string) (*PhaseState, error) { return f.state, f.hydrateErr }
func (f *fakeSessionStore) HydrateWithOpts(phase Phase, specID string, opts HydrateOpts) (*PhaseState, error) {
	return f.state, f.hydrateErr
}
func (f *fakeSessionStore) AppendTaskLedger(entry TaskLedgerEntry) error        { return nil }
func (f *fakeSessionStore) WriteRunArtifact(iterID, name string, body []byte) error { return nil }
func (f *fakeSessionStore) RecordBlocker(report BlockerReport) error            { return nil }
func (f *fakeSessionStore) ResolveBlocker(phase Phase, specID string, resolution string) error {
	return nil
}
func (f *fakeSessionStore) DetectInFlightTransition(specID string) (Phase, Phase, bool, error) {
	return "", "", false, nil
}
func (f *fakeSessionStore) MergeTeamCheckpoints(specID string, phase Phase, agentNames []string) (*PhaseState, error) {
	return nil, nil
}

func TestHydrateForPrompt_NilState(t *testing.T) {
	// store.Hydrate 가 nil state 반환 → 3-tuple 모두 빈 문자열
	store := &fakeSessionStore{state: nil}
	sysP, userC, sysC, err := HydrateForPrompt(PhasePlan, "SPEC-NONE", store)
	if err != nil {
		t.Fatalf("HydrateForPrompt nil state should not error, got: %v", err)
	}
	if sysP != "" || userC != "" || sysC != "" {
		t.Errorf("nil state should return empty fragments, got: sys=%q user=%q sysCtx=%q",
			sysP, userC, sysC)
	}
}

func TestHydrateForPrompt_HydrateError(t *testing.T) {
	// store.Hydrate 가 ErrCheckpointStale 반환 → wrapped error
	store := &fakeSessionStore{state: nil, hydrateErr: ErrCheckpointStale}
	_, _, _, err := HydrateForPrompt(PhaseRun, "SPEC-STALE", store)
	if err == nil {
		t.Fatal("HydrateForPrompt should propagate stale error, got nil")
	}
	if !strings.Contains(err.Error(), "hydrate phase state") {
		t.Errorf("error message should wrap with 'hydrate phase state', got: %v", err)
	}
}

func TestHydrateForPrompt_CacheOrderInvariant(t *testing.T) {
	// cache-prefix DO-NOT-REORDER 계약. (systemPrompt, userContext, systemContext)
	// 순서가 Anthropic cache key 의 구성 요소이므로 reorder 금지.
	// 본 테스트는 각 fragment 의 prefix 식별 토큰으로 순서를 정적 검증한다.
	store := &fakeSessionStore{
		state: &PhaseState{
			Phase:  PhasePlan,
			SPECID: "SPEC-CACHE-ORDER",
			Checkpoint: &PlanCheckpoint{
				SPECID:       "SPEC-CACHE-ORDER",
				Status:       "approved",
				ResearchPath: "/r",
			},
			UpdatedAt:  time.Now(),
			Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now()},
		},
	}

	sysP, userC, sysC, err := HydrateForPrompt(PhasePlan, "SPEC-CACHE-ORDER", store)
	if err != nil {
		t.Fatalf("HydrateForPrompt failed: %v", err)
	}
	// (1) systemPrompt: phase= / spec= / status= 식별 토큰
	if !strings.Contains(sysP, "phase=plan") || !strings.Contains(sysP, "spec=SPEC-CACHE-ORDER") || !strings.Contains(sysP, "status=approved") {
		t.Errorf("systemPrompt missing required tokens, got: %q", sysP)
	}
	// (2) userContext: spec_id= 식별 토큰
	if !strings.Contains(userC, "spec_id=SPEC-CACHE-ORDER") {
		t.Errorf("userContext missing spec_id token, got: %q", userC)
	}
	// (3) systemContext: provenance_source= / origin= 식별 토큰
	if !strings.Contains(sysC, "provenance_source=session") || !strings.Contains(sysC, "origin=cli") {
		t.Errorf("systemContext missing provenance tokens, got: %q", sysC)
	}
}

func TestHydrateForPrompt_AllPhases(t *testing.T) {
	tests := []struct {
		name           string
		state          *PhaseState
		expectStatus   string
		expectSpecPart string
	}{
		{
			name: "PlanCheckpoint_status_approved",
			state: &PhaseState{
				Phase: PhasePlan, SPECID: "SPEC-PL-001",
				Checkpoint: &PlanCheckpoint{SPECID: "SPEC-PL-001", Status: "approved"},
				UpdatedAt:  time.Now(),
				Provenance: ProvenanceTag{Source: "session", Origin: "test"},
			},
			expectStatus:   "approved",
			expectSpecPart: "SPEC-PL-001",
		},
		{
			name: "RunCheckpoint_status_pass",
			state: &PhaseState{
				Phase: PhaseRun, SPECID: "SPEC-RUN-001",
				Checkpoint: &RunCheckpoint{SPECID: "SPEC-RUN-001", Status: "pass", Harness: "standard"},
				UpdatedAt:  time.Now(),
				Provenance: ProvenanceTag{Source: "session", Origin: "test"},
			},
			expectStatus:   "pass",
			expectSpecPart: "SPEC-RUN-001",
		},
		{
			name: "SyncCheckpoint_synced",
			state: &PhaseState{
				Phase: PhaseSync, SPECID: "SPEC-SY-001",
				Checkpoint: &SyncCheckpoint{SPECID: "SPEC-SY-001", DocsSynced: true},
				UpdatedAt:  time.Now(),
				Provenance: ProvenanceTag{Source: "session", Origin: "test"},
			},
			expectStatus:   "synced",
			expectSpecPart: "SPEC-SY-001",
		},
		{
			name: "SyncCheckpoint_pending",
			state: &PhaseState{
				Phase: PhaseSync, SPECID: "SPEC-SY-002",
				Checkpoint: &SyncCheckpoint{SPECID: "SPEC-SY-002", DocsSynced: false},
				UpdatedAt:  time.Now(),
				Provenance: ProvenanceTag{Source: "session", Origin: "test"},
			},
			expectStatus:   "pending",
			expectSpecPart: "SPEC-SY-002",
		},
		{
			name: "NoCheckpoint_no_checkpoint",
			state: &PhaseState{
				Phase: PhaseReview, SPECID: "SPEC-NOCP-001",
				Checkpoint: nil,
				UpdatedAt:  time.Now(),
				Provenance: ProvenanceTag{Source: "session", Origin: "test"},
			},
			expectStatus:   "no_checkpoint",
			expectSpecPart: "SPEC-NOCP-001",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeSessionStore{state: tt.state}
			sysP, userC, _, err := HydrateForPrompt(tt.state.Phase, tt.state.SPECID, store)
			if err != nil {
				t.Fatalf("HydrateForPrompt failed: %v", err)
			}
			if !strings.Contains(sysP, "status="+tt.expectStatus) {
				t.Errorf("systemPrompt should contain status=%q, got: %q", tt.expectStatus, sysP)
			}
			if !strings.Contains(userC, tt.expectSpecPart) {
				t.Errorf("userContext should contain %q, got: %q", tt.expectSpecPart, userC)
			}
		})
	}
}

// TestCheckpointStatus_UnknownConcreteType verifies the switch default branch.
// 알려진 Checkpoint 구현체 (Plan/Run/Sync) 외의 type 에 대해 "unknown" 반환.
func TestCheckpointStatus_UnknownConcreteType(t *testing.T) {
	state := &PhaseState{
		Phase:      PhaseReview,
		SPECID:     "SPEC-UNK-CP",
		Checkpoint: &unknownCheckpoint{},
		UpdatedAt:  time.Now(),
	}
	got := checkpointStatus(state)
	if got != "unknown" {
		t.Errorf("checkpointStatus unknown-type = %q, want unknown", got)
	}
}

// unknownCheckpoint is a synthetic Checkpoint implementation used to exercise
// the `default` switch branch in checkpointStatus.
type unknownCheckpoint struct{}

func (u *unknownCheckpoint) PhaseName() Phase { return PhaseReview }
func (u *unknownCheckpoint) Validate() error  { return nil }
