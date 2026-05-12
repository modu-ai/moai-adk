package harness_test

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// TestEvaluatorPriorJudgmentLeak는 evaluator spawn 프롬프트에서
// 이전 iteration 판단 흔적을 탐지하는 DetectPriorJudgmentLeak 함수를 검증합니다.
// AC-HRN-002-07 leaf 시나리오를 포함합니다.
func TestEvaluatorPriorJudgmentLeak(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		spawnPrompt string
		wantErr     bool
		wantErrIs   error
	}{
		// AC-HRN-002-07.a: 깨끗한 프롬프트 — 오류 없음
		{
			name: "clean prompt no leak",
			spawnPrompt: `You are a fresh evaluator.
BRIEF: Build a landing page.
Sprint Contract: criteria.yaml
Artifact path: /output/index.html`,
			wantErr: false,
		},
		// AC-HRN-002-07.a.i: 숫자 iteration 참조 — ErrPriorJudgmentLeak
		{
			name: "numbered iteration reference",
			spawnPrompt: `Iteration 3 produced a score of 0.82.
Based on the previous result, please re-evaluate.
BRIEF: Build a landing page.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
		// AC-HRN-002-07.a.ii: 이전 evaluator 언급 — ErrPriorJudgmentLeak
		{
			name: "paraphrased prior evaluator rationale",
			spawnPrompt: `The previous evaluator reasoned that the design was incomplete.
Please consider this when scoring.
BRIEF: Build a landing page.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
		// AC-HRN-002-07.b: Score: 서브스트링 — ErrPriorJudgmentLeak
		{
			name: "score substring present",
			spawnPrompt: `Score: 0.85
Feedback: The design looks good but needs improvement.
BRIEF: Build a landing page.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
		// Feedback: 서브스트링 — ErrPriorJudgmentLeak
		{
			name: "feedback substring present",
			spawnPrompt: `Feedback: Please fix the color contrast issue.
BRIEF: Build a landing page.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
		// Verdict: 서브스트링 — ErrPriorJudgmentLeak
		{
			name: "verdict substring present",
			spawnPrompt: `Verdict: FAIL
The artifact does not meet the required criteria.
BRIEF: Build a landing page.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
		// AC-HRN-002-07.c: 깨끗한 BRIEF 참조 — 오류 없음
		{
			name: "clean brief reference no leak",
			spawnPrompt: `You are a fresh evaluator for this iteration.
Review the BRIEF at .moai/design/brief/BRIEF-001.md.
Check the Sprint Contract at .moai/sprints/my-spec/contract.yaml.
Artifact: /output/index.html`,
			wantErr: false,
		},
		// lowercase iteration N 패턴 — ErrPriorJudgmentLeak
		{
			name: "lowercase iteration reference",
			spawnPrompt: `In iteration 2, the evaluator found issues.
Please address those issues.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
		// prior evaluator 대소문자 변형 — ErrPriorJudgmentLeak
		{
			name: "prior evaluator uppercase variant",
			spawnPrompt: `The Prior Evaluator found that the design was incomplete.
Please improve accordingly.`,
			wantErr:   true,
			wantErrIs: harness.ErrPriorJudgmentLeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := harness.DetectPriorJudgmentLeak(tt.spawnPrompt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectPriorJudgmentLeak() error = nil, want non-nil (ErrPriorJudgmentLeak)")
					return
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("DetectPriorJudgmentLeak() error = %v, want errors.Is(%v)", err, tt.wantErrIs)
				}
			} else {
				if err != nil {
					t.Errorf("DetectPriorJudgmentLeak() error = %v, want nil", err)
				}
			}
		})
	}
}
