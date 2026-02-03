package foundation

import (
	"encoding/json"
	"testing"
)

func TestPillarString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Pillar
		want string
	}{
		{name: "Tested", p: Tested, want: "tested"},
		{name: "Readable", p: Readable, want: "readable"},
		{name: "Understandable", p: Understandable, want: "understandable"},
		{name: "Secured", p: Secured, want: "secured"},
		{name: "Trackable", p: Trackable, want: "trackable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.p.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPillarIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Pillar
		want bool
	}{
		{name: "Tested", p: Tested, want: true},
		{name: "Readable", p: Readable, want: true},
		{name: "Understandable", p: Understandable, want: true},
		{name: "Secured", p: Secured, want: true},
		{name: "Trackable", p: Trackable, want: true},
		{name: "empty", p: Pillar(""), want: false},
		{name: "invalid", p: Pillar("invalid"), want: false},
		{name: "partial", p: Pillar("test"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.p.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllPillars(t *testing.T) {
	t.Parallel()

	pillars := AllPillars()

	if len(pillars) != 5 {
		t.Fatalf("AllPillars() returned %d pillars, want 5", len(pillars))
	}

	expected := map[Pillar]bool{
		Tested:         true,
		Readable:       true,
		Understandable: true,
		Secured:        true,
		Trackable:      true,
	}

	for _, p := range pillars {
		if !expected[p] {
			t.Errorf("unexpected pillar: %s", p)
		}
		delete(expected, p)
	}

	for p := range expected {
		t.Errorf("missing pillar: %s", p)
	}
}

func TestPillarStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    PillarStatus
		want string
	}{
		{name: "Pass", s: StatusPass, want: "pass"},
		{name: "Warning", s: StatusWarning, want: "warning"},
		{name: "Critical", s: StatusCritical, want: "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.s.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewAssessment(t *testing.T) {
	t.Parallel()

	a := NewAssessment()

	if a == nil {
		t.Fatal("NewAssessment() returned nil")
	}

	if len(a.Scores) != 5 {
		t.Errorf("NewAssessment() has %d scores, want 5", len(a.Scores))
	}

	for _, p := range AllPillars() {
		ps, ok := a.Scores[p]
		if !ok {
			t.Errorf("missing pillar score for %s", p)
			continue
		}
		if ps.Score != 0.0 {
			t.Errorf("pillar %s initial score = %f, want 0.0", p, ps.Score)
		}
		if ps.Status != StatusCritical {
			t.Errorf("pillar %s initial status = %q, want %q", p, ps.Status, StatusCritical)
		}
		if ps.Issues == nil {
			t.Errorf("pillar %s issues should not be nil", p)
		}
	}
}

func TestAssessmentSetScore(t *testing.T) {
	t.Parallel()

	t.Run("valid_score", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		err := a.SetScore(Tested, 0.90, []string{"good coverage"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Scores[Tested].Score != 0.90 {
			t.Errorf("score = %f, want 0.90", a.Scores[Tested].Score)
		}
	})

	t.Run("zero_score", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		err := a.SetScore(Tested, 0.0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Scores[Tested].Score != 0.0 {
			t.Errorf("score = %f, want 0.0", a.Scores[Tested].Score)
		}
		if a.Scores[Tested].Issues == nil {
			t.Error("issues should be empty slice, not nil")
		}
	})

	t.Run("max_score", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		err := a.SetScore(Tested, 1.0, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Scores[Tested].Score != 1.0 {
			t.Errorf("score = %f, want 1.0", a.Scores[Tested].Score)
		}
	})

	t.Run("invalid_pillar", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		err := a.SetScore(Pillar("invalid"), 0.5, nil)
		if err == nil {
			t.Error("expected error for invalid pillar, got nil")
		}
	})

	t.Run("negative_score", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		err := a.SetScore(Tested, -0.1, nil)
		if err == nil {
			t.Error("expected error for negative score, got nil")
		}
	})

	t.Run("score_above_one", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		err := a.SetScore(Tested, 1.1, nil)
		if err == nil {
			t.Error("expected error for score > 1.0, got nil")
		}
	})
}

func TestAssessmentEvaluate(t *testing.T) {
	t.Parallel()

	t.Run("all_pass", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.90, nil)
		}
		a.Evaluate()
		for _, p := range AllPillars() {
			if a.Scores[p].Status != StatusPass {
				t.Errorf("pillar %s status = %q, want %q", p, a.Scores[p].Status, StatusPass)
			}
		}
	})

	t.Run("mixed_statuses", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		_ = a.SetScore(Tested, 0.90, nil)
		_ = a.SetScore(Readable, 0.70, nil)
		_ = a.SetScore(Understandable, 0.40, nil)
		_ = a.SetScore(Secured, 0.85, nil)
		_ = a.SetScore(Trackable, 0.50, nil)
		a.Evaluate()

		if a.Scores[Tested].Status != StatusPass {
			t.Errorf("Tested status = %q, want %q", a.Scores[Tested].Status, StatusPass)
		}
		if a.Scores[Readable].Status != StatusWarning {
			t.Errorf("Readable status = %q, want %q", a.Scores[Readable].Status, StatusWarning)
		}
		if a.Scores[Understandable].Status != StatusCritical {
			t.Errorf("Understandable status = %q, want %q", a.Scores[Understandable].Status, StatusCritical)
		}
		if a.Scores[Secured].Status != StatusPass {
			t.Errorf("Secured status = %q, want %q", a.Scores[Secured].Status, StatusPass)
		}
		if a.Scores[Trackable].Status != StatusWarning {
			t.Errorf("Trackable status = %q, want %q", a.Scores[Trackable].Status, StatusWarning)
		}
	})

	t.Run("boundary_at_085", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		_ = a.SetScore(Tested, 0.85, nil)
		_ = a.SetScore(Readable, 0.8499, nil)
		_ = a.SetScore(Understandable, 0.85, nil)
		_ = a.SetScore(Secured, 0.85, nil)
		_ = a.SetScore(Trackable, 0.85, nil)
		a.Evaluate()

		if a.Scores[Tested].Status != StatusPass {
			t.Errorf("score 0.85 should be Pass, got %q", a.Scores[Tested].Status)
		}
		if a.Scores[Readable].Status != StatusWarning {
			t.Errorf("score 0.8499 should be Warning, got %q", a.Scores[Readable].Status)
		}
	})

	t.Run("boundary_at_050", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		_ = a.SetScore(Tested, 0.50, nil)
		_ = a.SetScore(Readable, 0.4999, nil)
		_ = a.SetScore(Understandable, 0.50, nil)
		_ = a.SetScore(Secured, 0.50, nil)
		_ = a.SetScore(Trackable, 0.50, nil)
		a.Evaluate()

		if a.Scores[Tested].Status != StatusWarning {
			t.Errorf("score 0.50 should be Warning, got %q", a.Scores[Tested].Status)
		}
		if a.Scores[Readable].Status != StatusCritical {
			t.Errorf("score 0.4999 should be Critical, got %q", a.Scores[Readable].Status)
		}
	})
}

func TestAssessmentOverallStatus(t *testing.T) {
	t.Parallel()

	t.Run("all_pass", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.90, nil)
		}
		a.Evaluate()
		if got := a.OverallStatus(); got != StatusPass {
			t.Errorf("OverallStatus() = %q, want %q", got, StatusPass)
		}
	})

	t.Run("one_warning", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.90, nil)
		}
		_ = a.SetScore(Readable, 0.70, nil)
		a.Evaluate()
		if got := a.OverallStatus(); got != StatusWarning {
			t.Errorf("OverallStatus() = %q, want %q", got, StatusWarning)
		}
	})

	t.Run("one_critical", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.90, nil)
		}
		_ = a.SetScore(Secured, 0.30, nil)
		a.Evaluate()
		if got := a.OverallStatus(); got != StatusCritical {
			t.Errorf("OverallStatus() = %q, want %q", got, StatusCritical)
		}
	})

	t.Run("default_critical", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		// Default scores are 0.0, all should be critical.
		if got := a.OverallStatus(); got != StatusCritical {
			t.Errorf("OverallStatus() on default = %q, want %q", got, StatusCritical)
		}
	})
}

func TestAssessmentIsPass(t *testing.T) {
	t.Parallel()

	t.Run("all_above_threshold", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.90, nil)
		}
		if !a.IsPass() {
			t.Error("IsPass() should return true when all scores >= 0.85")
		}
	})

	t.Run("all_at_threshold", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.85, nil)
		}
		if !a.IsPass() {
			t.Error("IsPass() should return true when all scores == 0.85")
		}
	})

	t.Run("one_below_threshold", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.90, nil)
		}
		_ = a.SetScore(Tested, 0.84, nil)
		if a.IsPass() {
			t.Error("IsPass() should return false when one pillar < 0.85")
		}
	})

	t.Run("all_zero", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		if a.IsPass() {
			t.Error("IsPass() should return false when all scores are 0.0")
		}
	})

	t.Run("all_perfect", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 1.0, nil)
		}
		if !a.IsPass() {
			t.Error("IsPass() should return true when all scores are 1.0")
		}
	})
}

func TestAssessmentOverallScore(t *testing.T) {
	t.Parallel()

	t.Run("equal_scores", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 0.80, nil)
		}
		if got := a.OverallScore(); got != 0.80 {
			t.Errorf("OverallScore() = %f, want 0.80", got)
		}
	})

	t.Run("varied_scores", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		_ = a.SetScore(Tested, 0.90, nil)
		_ = a.SetScore(Readable, 0.85, nil)
		_ = a.SetScore(Understandable, 1.00, nil)
		_ = a.SetScore(Secured, 0.75, nil)
		_ = a.SetScore(Trackable, 0.80, nil)
		// (0.90 + 0.85 + 1.00 + 0.75 + 0.80) / 5 = 4.30 / 5 = 0.86
		got := a.OverallScore()
		if got != 0.86 {
			t.Errorf("OverallScore() = %f, want 0.86", got)
		}
	})

	t.Run("all_zero", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		if got := a.OverallScore(); got != 0.0 {
			t.Errorf("OverallScore() = %f, want 0.0", got)
		}
	})

	t.Run("all_perfect", func(t *testing.T) {
		t.Parallel()
		a := NewAssessment()
		for _, p := range AllPillars() {
			_ = a.SetScore(p, 1.0, nil)
		}
		if got := a.OverallScore(); got != 1.0 {
			t.Errorf("OverallScore() = %f, want 1.0", got)
		}
	})
}

func TestAssessmentJSONRoundTrip(t *testing.T) {
	t.Parallel()

	a := NewAssessment()
	_ = a.SetScore(Tested, 0.90, []string{"good coverage"})
	_ = a.SetScore(Readable, 0.85, nil)
	_ = a.SetScore(Understandable, 1.00, nil)
	_ = a.SetScore(Secured, 0.75, []string{"missing input validation"})
	_ = a.SetScore(Trackable, 0.80, nil)
	a.Evaluate()

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Assessment
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(got.Scores) != 5 {
		t.Errorf("unmarshaled scores count = %d, want 5", len(got.Scores))
	}

	for _, p := range AllPillars() {
		orig := a.Scores[p]
		restored := got.Scores[p]
		if restored == nil {
			t.Errorf("missing pillar %s after unmarshal", p)
			continue
		}
		if restored.Score != orig.Score {
			t.Errorf("pillar %s score = %f, want %f", p, restored.Score, orig.Score)
		}
		if restored.Status != orig.Status {
			t.Errorf("pillar %s status = %q, want %q", p, restored.Status, orig.Status)
		}
	}
}

func TestPillarScoreJSONRoundTrip(t *testing.T) {
	t.Parallel()

	ps := &PillarScore{
		Pillar: Tested,
		Score:  0.90,
		Status: StatusPass,
		Issues: []string{"minor issue"},
	}

	data, err := json.Marshal(ps)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got PillarScore
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Pillar != ps.Pillar {
		t.Errorf("Pillar = %q, want %q", got.Pillar, ps.Pillar)
	}
	if got.Score != ps.Score {
		t.Errorf("Score = %f, want %f", got.Score, ps.Score)
	}
	if got.Status != ps.Status {
		t.Errorf("Status = %q, want %q", got.Status, ps.Status)
	}
	if len(got.Issues) != len(ps.Issues) {
		t.Errorf("Issues length = %d, want %d", len(got.Issues), len(ps.Issues))
	}
}
