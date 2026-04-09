package experiment

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// defaultConfig는 테스트에서 공통으로 사용하는 기본 설정이다.
func defaultConfig() LoopConfig {
	return LoopConfig{
		MaxExperiments:      10,
		TargetScore:         0.90,
		StagnationThreshold: 0.01,
		StagnationPatience:  3,
	}
}

// makeEvalResult는 테스트용 EvalResult를 생성한다.
func makeEvalResult(overall float64, mustPassOK bool) *eval.EvalResult {
	return &eval.EvalResult{
		Overall: overall,
		PerCriterion: map[string]eval.CriterionResult{
			"accuracy": {Name: "accuracy", Passed: mustPassOK, Weight: eval.MustPass},
		},
		MustPassOK: mustPassOK,
		Timestamp:  time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
}

// makeExp는 테스트용 Experiment를 생성한다.
func makeExp(id string, overall float64, mustPassOK bool) *Experiment {
	return &Experiment{
		ID:         id,
		Target:     "test-target",
		Hypothesis: "테스트 가설",
		Change:     ChangeRecord{Type: "modification", Section: "prompt", Diff: "diff"},
		Result:     makeEvalResult(overall, mustPassOK),
		Decision:   DecisionPending,
		Timestamp:  time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
}

// TestNewLoop은 Loop 생성 시 초기 상태를 검증한다.
func TestNewLoop(t *testing.T) {
	t.Parallel()

	l := NewLoop(defaultConfig())

	if l.State() != StateIdle {
		t.Errorf("초기 상태: got %q, want %q", l.State(), StateIdle)
	}
	if l.BestScore() != 0.0 {
		t.Errorf("초기 BestScore: got %f, want 0.0", l.BestScore())
	}
	if l.ExperimentCount() != 0 {
		t.Errorf("초기 ExperimentCount: got %d, want 0", l.ExperimentCount())
	}
}

// TestSetBaseline은 베이스라인 설정 후 상태와 점수를 검증한다.
func TestSetBaseline(t *testing.T) {
	t.Parallel()

	l := NewLoop(defaultConfig())
	baseline := makeEvalResult(0.70, true)

	l.SetBaseline(baseline)

	if l.State() != StateBaseline {
		t.Errorf("상태: got %q, want %q", l.State(), StateBaseline)
	}
	if l.BestScore() != 0.70 {
		t.Errorf("BestScore: got %f, want 0.70", l.BestScore())
	}
}

// TestShouldContinue는 루프 지속 조건을 테이블 기반으로 검증한다.
func TestShouldContinue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Loop
		want  bool
	}{
		{
			name: "초기_상태_지속",
			setup: func() *Loop {
				l := NewLoop(defaultConfig())
				l.SetBaseline(makeEvalResult(0.50, true))
				return l
			},
			want: true,
		},
		{
			name: "최대_실험_수_도달",
			setup: func() *Loop {
				cfg := defaultConfig()
				cfg.MaxExperiments = 3
				l := NewLoop(cfg)
				l.SetBaseline(makeEvalResult(0.50, true))
				// 3개 실험 기록 (점수 계속 상승하여 정체 없음)
				l.RecordExperiment(makeExp("1", 0.55, true))
				l.RecordExperiment(makeExp("2", 0.60, true))
				l.RecordExperiment(makeExp("3", 0.65, true))
				return l
			},
			want: false,
		},
		{
			name: "목표_점수_3회_연속_달성",
			setup: func() *Loop {
				cfg := defaultConfig()
				cfg.TargetScore = 0.90
				l := NewLoop(cfg)
				l.SetBaseline(makeEvalResult(0.50, true))
				// 목표 점수 3회 연속 달성
				l.RecordExperiment(makeExp("1", 0.91, true))
				l.RecordExperiment(makeExp("2", 0.92, true))
				l.RecordExperiment(makeExp("3", 0.93, true))
				return l
			},
			want: false,
		},
		{
			name: "정체_인내_초과",
			setup: func() *Loop {
				cfg := defaultConfig()
				cfg.StagnationPatience = 3
				cfg.StagnationThreshold = 0.05
				l := NewLoop(cfg)
				l.SetBaseline(makeEvalResult(0.50, true))
				// 미미한 개선으로 정체 3회
				l.RecordExperiment(makeExp("1", 0.51, true))
				l.RecordExperiment(makeExp("2", 0.51, true))
				l.RecordExperiment(makeExp("3", 0.51, true))
				return l
			},
			want: false,
		},
		{
			name: "완료_상태",
			setup: func() *Loop {
				l := NewLoop(defaultConfig())
				l.state = StateComplete
				return l
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := tt.setup()
			if got := l.ShouldContinue(); got != tt.want {
				t.Errorf("ShouldContinue: got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRecordExperiment는 실험 기록 시 Decision 로직을 테이블 기반으로 검증한다.
func TestRecordExperiment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		baseline     float64
		expScore     float64
		mustPassOK   bool
		nilResult    bool
		wantDecision Decision
		wantBest     float64
	}{
		{
			name:         "더_높은_점수_keep",
			baseline:     0.50,
			expScore:     0.80,
			mustPassOK:   true,
			wantDecision: DecisionKeep,
			wantBest:     0.80,
		},
		{
			name:         "더_낮은_점수_discard",
			baseline:     0.80,
			expScore:     0.60,
			mustPassOK:   true,
			wantDecision: DecisionDiscard,
			wantBest:     0.80,
		},
		{
			name:         "must_pass_실패_높은_점수도_discard",
			baseline:     0.50,
			expScore:     0.90,
			mustPassOK:   false,
			wantDecision: DecisionDiscard,
			wantBest:     0.50,
		},
		{
			name:         "nil_result_discard",
			baseline:     0.50,
			nilResult:    true,
			wantDecision: DecisionDiscard,
			wantBest:     0.50,
		},
		{
			name:         "동일_점수_discard",
			baseline:     0.70,
			expScore:     0.70,
			mustPassOK:   true,
			wantDecision: DecisionDiscard,
			wantBest:     0.70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := NewLoop(defaultConfig())
			l.SetBaseline(makeEvalResult(tt.baseline, true))

			var exp *Experiment
			if tt.nilResult {
				exp = &Experiment{
					ID:        "test",
					Result:    nil,
					Timestamp: time.Now(),
				}
			} else {
				exp = makeExp("test", tt.expScore, tt.mustPassOK)
			}

			decision := l.RecordExperiment(exp)

			if decision != tt.wantDecision {
				t.Errorf("Decision: got %q, want %q", decision, tt.wantDecision)
			}
			if l.BestScore() != tt.wantBest {
				t.Errorf("BestScore: got %f, want %f", l.BestScore(), tt.wantBest)
			}
		})
	}
}

// TestStagnationCounter는 정체 카운터의 증가/초기화를 검증한다.
func TestStagnationCounter(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.StagnationThreshold = 0.05
	cfg.StagnationPatience = 10 // 높게 설정하여 조기 종료 방지
	l := NewLoop(cfg)
	l.SetBaseline(makeEvalResult(0.50, true))

	// 미미한 개선 → 정체 카운터 증가
	l.RecordExperiment(makeExp("1", 0.51, true))
	if l.stagnationCount != 1 {
		t.Errorf("1회 후 stagnationCount: got %d, want 1", l.stagnationCount)
	}

	// 미미한 개선 → 정체 카운터 증가
	l.RecordExperiment(makeExp("2", 0.51, true))
	if l.stagnationCount != 2 {
		t.Errorf("2회 후 stagnationCount: got %d, want 2", l.stagnationCount)
	}

	// 큰 개선 → 정체 카운터 초기화
	l.RecordExperiment(makeExp("3", 0.70, true))
	if l.stagnationCount != 0 {
		t.Errorf("큰 개선 후 stagnationCount: got %d, want 0", l.stagnationCount)
	}
}

// TestTargetHitCounter는 목표 점수 달성 카운터의 증가/초기화를 검증한다.
func TestTargetHitCounter(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.TargetScore = 0.80
	cfg.MaxExperiments = 20 // 높게 설정하여 조기 종료 방지
	l := NewLoop(cfg)
	l.SetBaseline(makeEvalResult(0.50, true))

	// 목표 달성 → 카운터 증가
	l.RecordExperiment(makeExp("1", 0.85, true))
	if l.targetHitCount != 1 {
		t.Errorf("1회 후 targetHitCount: got %d, want 1", l.targetHitCount)
	}

	// 목표 달성 → 카운터 증가
	l.RecordExperiment(makeExp("2", 0.90, true))
	if l.targetHitCount != 2 {
		t.Errorf("2회 후 targetHitCount: got %d, want 2", l.targetHitCount)
	}

	// 목표 미달 → 카운터 초기화
	l.RecordExperiment(makeExp("3", 0.60, true))
	if l.targetHitCount != 0 {
		t.Errorf("미달 후 targetHitCount: got %d, want 0", l.targetHitCount)
	}
}

// TestExperimentCount는 실험 카운트가 정확히 증가하는지 검증한다.
func TestExperimentCount(t *testing.T) {
	t.Parallel()

	l := NewLoop(defaultConfig())
	l.SetBaseline(makeEvalResult(0.50, true))

	for i := 0; i < 5; i++ {
		l.RecordExperiment(makeExp("exp", 0.50+float64(i)*0.05, true))
	}

	if l.ExperimentCount() != 5 {
		t.Errorf("ExperimentCount: got %d, want 5", l.ExperimentCount())
	}
}
