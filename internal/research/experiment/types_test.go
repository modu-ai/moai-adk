package experiment

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// TestExperimentState_Constants는 상태 상수가 기대 값을 가지는지 검증한다.
func TestExperimentState_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state ExperimentState
		want  string
	}{
		{"idle", StateIdle, "idle"},
		{"baseline", StateBaseline, "baseline"},
		{"mutating", StateMutating, "mutating"},
		{"evaluating", StateEvaluating, "evaluating"},
		{"scoring", StateScoring, "scoring"},
		{"complete", StateComplete, "complete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.state) != tt.want {
				t.Errorf("ExperimentState %s = %q, want %q", tt.name, tt.state, tt.want)
			}
		})
	}
}

// TestDecision_Constants는 결정 상수가 기대 값을 가지는지 검증한다.
func TestDecision_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		decision Decision
		want     string
	}{
		{"keep", DecisionKeep, "keep"},
		{"discard", DecisionDiscard, "discard"},
		{"pending", DecisionPending, "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.decision) != tt.want {
				t.Errorf("Decision %s = %q, want %q", tt.name, tt.decision, tt.want)
			}
		})
	}
}

// TestChangeRecord_JSONRoundTrip은 ChangeRecord의 JSON 직렬화/역직렬화를 검증한다.
func TestChangeRecord_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := ChangeRecord{
		Type:    "modification",
		Section: "system_prompt",
		Diff:    "- old line\n+ new line",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 실패: %v", err)
	}

	var decoded ChangeRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 실패: %v", err)
	}

	if decoded != original {
		t.Errorf("라운드트립 결과 불일치:\n  원본: %+v\n  복원: %+v", original, decoded)
	}
}

// TestExperiment_JSONRoundTrip은 Experiment의 JSON 직렬화/역직렬화를 검증한다.
func TestExperiment_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	original := Experiment{
		ID:         "exp-001",
		Target:     "skills/moai-lang-go",
		Hypothesis: "Go 코드 블록 추가 시 정확도 향상",
		Change: ChangeRecord{
			Type:    "addition",
			Section: "examples",
			Diff:    "+ code block here",
		},
		Result: &eval.EvalResult{
			Overall: 0.85,
			PerCriterion: map[string]eval.CriterionResult{
				"accuracy": {Name: "accuracy", Passed: true, Weight: eval.MustPass},
			},
			MustPassOK: true,
			Timestamp:  now,
		},
		Decision:  DecisionKeep,
		Timestamp: now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 실패: %v", err)
	}

	var decoded Experiment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 실패: %v", err)
	}

	// 주요 필드 비교 (time.Time은 JSON 포맷 후 나노초 손실 가능)
	if decoded.ID != original.ID {
		t.Errorf("ID: got %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Target != original.Target {
		t.Errorf("Target: got %q, want %q", decoded.Target, original.Target)
	}
	if decoded.Hypothesis != original.Hypothesis {
		t.Errorf("Hypothesis: got %q, want %q", decoded.Hypothesis, original.Hypothesis)
	}
	if decoded.Change != original.Change {
		t.Errorf("Change: got %+v, want %+v", decoded.Change, original.Change)
	}
	if decoded.Decision != original.Decision {
		t.Errorf("Decision: got %q, want %q", decoded.Decision, original.Decision)
	}
	if decoded.Result == nil {
		t.Fatal("Result가 nil")
	}
	if decoded.Result.Overall != original.Result.Overall {
		t.Errorf("Result.Overall: got %f, want %f", decoded.Result.Overall, original.Result.Overall)
	}
	if decoded.Result.MustPassOK != original.Result.MustPassOK {
		t.Errorf("Result.MustPassOK: got %v, want %v", decoded.Result.MustPassOK, original.Result.MustPassOK)
	}
}

// TestChangelogEntry_JSONRoundTrip은 ChangelogEntry의 JSON 직렬화/역직렬화를 검증한다.
func TestChangelogEntry_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := ChangelogEntry{
		ExperimentID: "exp-003",
		Score:        0.92,
		Change:       "시스템 프롬프트에 예제 추가",
		Reasoning:    "예제가 포함된 프롬프트의 정확도가 높았음",
		Decision:     DecisionKeep,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 실패: %v", err)
	}

	var decoded ChangelogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 실패: %v", err)
	}

	if decoded != original {
		t.Errorf("라운드트립 결과 불일치:\n  원본: %+v\n  복원: %+v", original, decoded)
	}
}
