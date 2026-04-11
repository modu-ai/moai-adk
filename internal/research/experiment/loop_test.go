package experiment

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// defaultConfig returns the default configuration used in tests.
func defaultConfig() LoopConfig {
	return LoopConfig{
		MaxExperiments:      10,
		TargetScore:         0.90,
		StagnationThreshold: 0.01,
		StagnationPatience:  3,
	}
}

// makeEvalResult creates an EvalResult for testing.
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

// makeExp creates an Experiment for testing.
func makeExp(id string, overall float64, mustPassOK bool) *Experiment {
	return &Experiment{
		ID:         id,
		Target:     "test-target",
		Hypothesis: "test hypothesis",
		Change:     ChangeRecord{Type: "modification", Section: "prompt", Diff: "diff"},
		Result:     makeEvalResult(overall, mustPassOK),
		Decision:   DecisionPending,
		Timestamp:  time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
}

// TestNewLoop verifies the initial state when creating a Loop.
func TestNewLoop(t *testing.T) {
	t.Parallel()

	l := NewLoop(defaultConfig())

	if l.State() != StateIdle {
		t.Errorf("initial state: got %q, want %q", l.State(), StateIdle)
	}
	if l.BestScore() != 0.0 {
		t.Errorf("initial BestScore: got %f, want 0.0", l.BestScore())
	}
	if l.ExperimentCount() != 0 {
		t.Errorf("initial ExperimentCount: got %d, want 0", l.ExperimentCount())
	}
}

// TestSetBaseline verifies state and score after setting a baseline.
func TestSetBaseline(t *testing.T) {
	t.Parallel()

	l := NewLoop(defaultConfig())
	baseline := makeEvalResult(0.70, true)

	l.SetBaseline(baseline)

	if l.State() != StateBaseline {
		t.Errorf("state: got %q, want %q", l.State(), StateBaseline)
	}
	if l.BestScore() != 0.70 {
		t.Errorf("BestScore: got %f, want 0.70", l.BestScore())
	}
}

// TestShouldContinue verifies loop continuation conditions using table-driven tests.
func TestShouldContinue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Loop
		want  bool
	}{
		{
			name: "initial_state_continues",
			setup: func() *Loop {
				l := NewLoop(defaultConfig())
				l.SetBaseline(makeEvalResult(0.50, true))
				return l
			},
			want: true,
		},
		{
			name: "max_experiments_reached",
			setup: func() *Loop {
				cfg := defaultConfig()
				cfg.MaxExperiments = 3
				l := NewLoop(cfg)
				l.SetBaseline(makeEvalResult(0.50, true))
				// 3 experiments with continuously rising scores (no stagnation)
				l.RecordExperiment(makeExp("1", 0.55, true))
				l.RecordExperiment(makeExp("2", 0.60, true))
				l.RecordExperiment(makeExp("3", 0.65, true))
				return l
			},
			want: false,
		},
		{
			name: "target_score_hit_3_consecutive_times",
			setup: func() *Loop {
				cfg := defaultConfig()
				cfg.TargetScore = 0.90
				l := NewLoop(cfg)
				l.SetBaseline(makeEvalResult(0.50, true))
				// Hit target score 3 consecutive times
				l.RecordExperiment(makeExp("1", 0.91, true))
				l.RecordExperiment(makeExp("2", 0.92, true))
				l.RecordExperiment(makeExp("3", 0.93, true))
				return l
			},
			want: false,
		},
		{
			name: "stagnation_patience_exceeded",
			setup: func() *Loop {
				cfg := defaultConfig()
				cfg.StagnationPatience = 3
				cfg.StagnationThreshold = 0.05
				l := NewLoop(cfg)
				l.SetBaseline(makeEvalResult(0.50, true))
				// Minimal improvement causing 3 stagnation events
				l.RecordExperiment(makeExp("1", 0.51, true))
				l.RecordExperiment(makeExp("2", 0.51, true))
				l.RecordExperiment(makeExp("3", 0.51, true))
				return l
			},
			want: false,
		},
		{
			name: "complete_state",
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

// TestRecordExperiment verifies Decision logic when recording experiments using table-driven tests.
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
			name:         "higher_score_keep",
			baseline:     0.50,
			expScore:     0.80,
			mustPassOK:   true,
			wantDecision: DecisionKeep,
			wantBest:     0.80,
		},
		{
			name:         "lower_score_discard",
			baseline:     0.80,
			expScore:     0.60,
			mustPassOK:   true,
			wantDecision: DecisionDiscard,
			wantBest:     0.80,
		},
		{
			name:         "must_pass_failure_high_score_still_discard",
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
			name:         "equal_score_discard",
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

// TestStagnationCounter verifies that the stagnation counter increments and resets correctly.
func TestStagnationCounter(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.StagnationThreshold = 0.05
	cfg.StagnationPatience = 10 // set high to prevent early termination
	l := NewLoop(cfg)
	l.SetBaseline(makeEvalResult(0.50, true))

	// Minimal improvement → stagnation counter increments
	l.RecordExperiment(makeExp("1", 0.51, true))
	if l.stagnationCount != 1 {
		t.Errorf("stagnationCount after 1 minimal improvement: got %d, want 1", l.stagnationCount)
	}

	// Minimal improvement → stagnation counter increments
	l.RecordExperiment(makeExp("2", 0.51, true))
	if l.stagnationCount != 2 {
		t.Errorf("stagnationCount after 2 minimal improvements: got %d, want 2", l.stagnationCount)
	}

	// Large improvement → stagnation counter resets
	l.RecordExperiment(makeExp("3", 0.70, true))
	if l.stagnationCount != 0 {
		t.Errorf("stagnationCount after large improvement: got %d, want 0", l.stagnationCount)
	}
}

// TestTargetHitCounter verifies that the target hit counter increments and resets correctly.
func TestTargetHitCounter(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.TargetScore = 0.80
	cfg.MaxExperiments = 20 // set high to prevent early termination
	l := NewLoop(cfg)
	l.SetBaseline(makeEvalResult(0.50, true))

	// Target hit → counter increments
	l.RecordExperiment(makeExp("1", 0.85, true))
	if l.targetHitCount != 1 {
		t.Errorf("targetHitCount after 1 hit: got %d, want 1", l.targetHitCount)
	}

	// Target hit → counter increments
	l.RecordExperiment(makeExp("2", 0.90, true))
	if l.targetHitCount != 2 {
		t.Errorf("targetHitCount after 2 hits: got %d, want 2", l.targetHitCount)
	}

	// Target missed → counter resets
	l.RecordExperiment(makeExp("3", 0.60, true))
	if l.targetHitCount != 0 {
		t.Errorf("targetHitCount after miss: got %d, want 0", l.targetHitCount)
	}
}

// TestExperimentCount verifies that the experiment count increments correctly.
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
