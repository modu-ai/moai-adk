package session

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPhaseStateMarshalUnmarshal(t *testing.T) {
	original := PhaseState{
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

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	// Unmarshal
	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	// Verify fields
	if unmarshaled.Phase != original.Phase {
		t.Errorf("Phase = %v, want %v", unmarshaled.Phase, original.Phase)
	}
	if unmarshaled.SPECID != original.SPECID {
		t.Errorf("SPECID = %v, want %v", unmarshaled.SPECID, original.SPECID)
	}

	// Verify checkpoint type and content
	pc, ok := unmarshaled.Checkpoint.(*PlanCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *PlanCheckpoint")
	}
	if pc.Status != "approved" {
		t.Errorf("Checkpoint.Status = %v, want approved", pc.Status)
	}
	if pc.ResearchPath != "/research/SPEC-001" {
		t.Errorf("Checkpoint.ResearchPath = %v, want /research/SPEC-001", pc.ResearchPath)
	}
}

func TestPhaseStateMarshalUnmarshalRunCheckpoint(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-002",
		Checkpoint: &RunCheckpoint{
			SPECID:        "SPEC-002",
			Status:        "pass",
			TestsTotal:    100,
			TestsPassed:   95,
			FilesModified: 12,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "project",
			Origin: "/path/to/config",
			Loaded: time.Now(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	rc, ok := unmarshaled.Checkpoint.(*RunCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *RunCheckpoint")
	}
	if rc.TestsTotal != 100 {
		t.Errorf("TestsTotal = %v, want 100", rc.TestsTotal)
	}
	if rc.TestsPassed != 95 {
		t.Errorf("TestsPassed = %v, want 95", rc.TestsPassed)
	}
}

func TestPhaseStateMarshalUnmarshalSyncCheckpoint(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseSync,
		SPECID: "SPEC-003",
		Checkpoint: &SyncCheckpoint{
			SPECID:     "SPEC-003",
			PRNumber:   123,
			PRURL:      "https://github.com/modu-ai/moai-adk/pull/123",
			DocsSynced: true,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "user",
			Origin: "manual_input",
			Loaded: time.Now(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	sc, ok := unmarshaled.Checkpoint.(*SyncCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *SyncCheckpoint")
	}
	if sc.PRNumber != 123 {
		t.Errorf("PRNumber = %v, want 123", sc.PRNumber)
	}
	if !sc.DocsSynced {
		t.Errorf("DocsSynced = %v, want true", sc.DocsSynced)
	}
}

func TestPhaseStateWithBlocker(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseReview,
		SPECID: "SPEC-004",
		BlockerRpt: &BlockerReport{
			Kind:            "quality_gate",
			Message:         "Lint errors found",
			RequestedAction: "fix_lint",
			Resolved:        false,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "hook",
			Origin: "pre-commit",
			Loaded: time.Now(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if unmarshaled.BlockerRpt == nil {
		t.Fatal("BlockerRpt should not be nil")
	}
	if unmarshaled.BlockerRpt.Kind != "quality_gate" {
		t.Errorf("BlockerRpt.Kind = %v, want quality_gate", unmarshaled.BlockerRpt.Kind)
	}
}

func TestProvenanceTag(t *testing.T) {
	tag := ProvenanceTag{
		Source: "project",
		Origin: "/path/to/config",
		Loaded: time.Now(),
	}

	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled ProvenanceTag
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if unmarshaled.Source != tag.Source {
		t.Errorf("Source = %v, want %v", unmarshaled.Source, tag.Source)
	}
	if unmarshaled.Origin != tag.Origin {
		t.Errorf("Origin = %v, want %v", unmarshaled.Origin, tag.Origin)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P1 보강 — state.go:29 MarshalJSON 0% → cover.
//
// 기존 TestPhaseStateMarshalUnmarshal 시리즈는 `json.Marshal(original)` 를 value
// receiver로 호출하기 때문에 pointer-receiver `(*PhaseState).MarshalJSON` 메서드를
// 실제로 트리거하지 못한다. 본 테스트는 `json.Marshal(&original)` 로 명시적 호출하여
// custom MarshalJSON 코드 경로 (Checkpoint interface polymorphism 직렬화) 를 검증한다.
func TestPhaseStateMarshalJSONPointerReceiver_RoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		state      PhaseState
		assertFunc func(t *testing.T, round *PhaseState)
	}{
		{
			name: "PlanCheckpoint",
			state: PhaseState{
				Phase:  PhasePlan,
				SPECID: "SPEC-MARSHAL-001",
				Checkpoint: &PlanCheckpoint{
					SPECID:       "SPEC-MARSHAL-001",
					Status:       "approved",
					ResearchPath: "/research/marshal-001",
				},
				UpdatedAt:  time.Now().Truncate(time.Second),
				Provenance: ProvenanceTag{Source: "user", Origin: "test", Loaded: time.Now().Truncate(time.Second)},
			},
			assertFunc: func(t *testing.T, round *PhaseState) {
				t.Helper()
				pc, ok := round.Checkpoint.(*PlanCheckpoint)
				if !ok {
					t.Fatalf("Checkpoint = %T, want *PlanCheckpoint", round.Checkpoint)
				}
				if pc.ResearchPath != "/research/marshal-001" {
					t.Errorf("ResearchPath = %q, want /research/marshal-001", pc.ResearchPath)
				}
			},
		},
		{
			name: "RunCheckpoint",
			state: PhaseState{
				Phase:  PhaseRun,
				SPECID: "SPEC-MARSHAL-002",
				Checkpoint: &RunCheckpoint{
					SPECID:        "SPEC-MARSHAL-002",
					Status:        "pass",
					Harness:       "thorough",
					TestsTotal:    50,
					TestsPassed:   50,
					FilesModified: 7,
				},
				UpdatedAt:  time.Now().Truncate(time.Second),
				Provenance: ProvenanceTag{Source: "session", Origin: "cli", Loaded: time.Now().Truncate(time.Second)},
			},
			assertFunc: func(t *testing.T, round *PhaseState) {
				t.Helper()
				rc, ok := round.Checkpoint.(*RunCheckpoint)
				if !ok {
					t.Fatalf("Checkpoint = %T, want *RunCheckpoint", round.Checkpoint)
				}
				if rc.TestsTotal != 50 || rc.TestsPassed != 50 || rc.FilesModified != 7 {
					t.Errorf("RunCheckpoint sums mismatch: total=%d passed=%d modified=%d",
						rc.TestsTotal, rc.TestsPassed, rc.FilesModified)
				}
			},
		},
		{
			name: "SyncCheckpoint",
			state: PhaseState{
				Phase:  PhaseSync,
				SPECID: "SPEC-MARSHAL-003",
				Checkpoint: &SyncCheckpoint{
					SPECID:     "SPEC-MARSHAL-003",
					PRNumber:   456,
					PRURL:      "https://example.invalid/pr/456",
					DocsSynced: true,
				},
				UpdatedAt:  time.Now().Truncate(time.Second),
				Provenance: ProvenanceTag{Source: "project", Origin: "/path", Loaded: time.Now().Truncate(time.Second)},
			},
			assertFunc: func(t *testing.T, round *PhaseState) {
				t.Helper()
				sc, ok := round.Checkpoint.(*SyncCheckpoint)
				if !ok {
					t.Fatalf("Checkpoint = %T, want *SyncCheckpoint", round.Checkpoint)
				}
				if sc.PRNumber != 456 || !sc.DocsSynced {
					t.Errorf("SyncCheckpoint fields mismatch: pr=%d synced=%v", sc.PRNumber, sc.DocsSynced)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Pointer-receiver explicit invocation — triggers (*PhaseState).MarshalJSON.
			data, err := json.Marshal(&tt.state)
			if err != nil {
				t.Fatalf("Marshal(&state) failed: %v", err)
			}
			// Sanity: serialized payload must include the "checkpoint" key.
			if !strings.Contains(string(data), `"checkpoint"`) {
				t.Errorf("serialized payload missing 'checkpoint' key: %s", string(data))
			}

			var round PhaseState
			if err := json.Unmarshal(data, &round); err != nil {
				t.Fatalf("Unmarshal() failed: %v", err)
			}
			if round.Phase != tt.state.Phase {
				t.Errorf("Phase = %v, want %v", round.Phase, tt.state.Phase)
			}
			if round.SPECID != tt.state.SPECID {
				t.Errorf("SPECID = %q, want %q", round.SPECID, tt.state.SPECID)
			}
			tt.assertFunc(t, &round)
		})
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P1 보강 — MarshalJSON nil-checkpoint branch.
//
// `ps.Checkpoint == nil` 분기 — Checkpoint 가 없는 PhaseState 도 안전하게 직렬화되어야
// 한다 (Checkpoint 키 자체가 omitempty 로 누락 가능).
func TestPhaseStateMarshalJSONPointerReceiver_NilCheckpoint(t *testing.T) {
	state := PhaseState{
		Phase:      PhaseReview,
		SPECID:     "SPEC-NIL-CP-001",
		Checkpoint: nil, // explicit nil
		UpdatedAt:  time.Now().Truncate(time.Second),
		Provenance: ProvenanceTag{Source: "hook", Origin: "pre-commit", Loaded: time.Now().Truncate(time.Second)},
	}

	data, err := json.Marshal(&state)
	if err != nil {
		t.Fatalf("Marshal(&state) with nil Checkpoint failed: %v", err)
	}

	var round PhaseState
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}
	if round.Checkpoint != nil {
		t.Errorf("round.Checkpoint = %v, want nil (omitempty)", round.Checkpoint)
	}
	if round.SPECID != "SPEC-NIL-CP-001" {
		t.Errorf("SPECID = %q, want SPEC-NIL-CP-001", round.SPECID)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 P4 보강 — UnmarshalJSON 77.3% → cover edge cases.
//
// (a) Phase 가 알려진 3종 (Plan/Run/Sync) 외 (예: PhaseReview) → switch default 분기
// (b) checkpoint 필드가 잘못된 JSON 구조 → 명시적 에러 반환
func TestPhaseStateUnmarshalJSON_UnknownPhase(t *testing.T) {
	// PhaseReview 는 PhaseState.UnmarshalJSON switch 의 어떤 case 에도 매치되지 않음
	// → Checkpoint 필드는 nil 로 남는다 (switch default 미정의 분기).
	raw := []byte(`{
		"phase": "review",
		"spec_id": "SPEC-UNKNOWN-PHASE",
		"checkpoint": {"some_key": "some_value"},
		"updated_at": "2026-05-25T00:00:00Z",
		"provenance": {"source": "hook", "origin": "test", "loaded": "2026-05-25T00:00:00Z"}
	}`)

	var ps PhaseState
	if err := json.Unmarshal(raw, &ps); err != nil {
		t.Fatalf("Unmarshal() unknown-phase should not error, got: %v", err)
	}
	if ps.Phase != PhaseReview {
		t.Errorf("Phase = %v, want PhaseReview", ps.Phase)
	}
	if ps.Checkpoint != nil {
		t.Errorf("Checkpoint = %v, want nil (unknown phase has no concrete type)", ps.Checkpoint)
	}
}

func TestPhaseStateUnmarshalJSON_MalformedCheckpoint(t *testing.T) {
	// `checkpoint` 가 잘못된 타입 (string) → 외층 Unmarshal 단계에서 에러
	raw := []byte(`{
		"phase": "plan",
		"spec_id": "SPEC-MALFORMED-001",
		"checkpoint": "not-an-object",
		"updated_at": "2026-05-25T00:00:00Z",
		"provenance": {"source": "session", "origin": "test", "loaded": "2026-05-25T00:00:00Z"}
	}`)

	var ps PhaseState
	err := json.Unmarshal(raw, &ps)
	if err == nil {
		t.Fatal("Unmarshal() malformed checkpoint should error, got nil")
	}
}

func TestPhaseStateUnmarshalJSON_PlanCheckpointBadStatus(t *testing.T) {
	// PhasePlan 케이스의 PlanCheckpoint 내부 직렬화 분기 — 잘못된 status 도
	// json.Unmarshal 단계는 통과한다 (Validate 가 별도 호출). 본 테스트는 Unmarshal
	// 자체 성공 + Checkpoint 가 PlanCheckpoint 로 복원되는 path 만 검증.
	raw := []byte(`{
		"phase": "plan",
		"spec_id": "SPEC-PLAN-BAD-STATUS",
		"checkpoint": {"spec_id": "SPEC-PLAN-BAD-STATUS", "status": "invalid-status", "research_path": "/r"},
		"updated_at": "2026-05-25T00:00:00Z",
		"provenance": {"source": "session", "origin": "test", "loaded": "2026-05-25T00:00:00Z"}
	}`)

	var ps PhaseState
	if err := json.Unmarshal(raw, &ps); err != nil {
		t.Fatalf("Unmarshal() should succeed (Validate is separate), got: %v", err)
	}
	pc, ok := ps.Checkpoint.(*PlanCheckpoint)
	if !ok {
		t.Fatalf("Checkpoint = %T, want *PlanCheckpoint", ps.Checkpoint)
	}
	if pc.Status != "invalid-status" {
		t.Errorf("Status = %q, want invalid-status (Unmarshal preserves raw value)", pc.Status)
	}
}
