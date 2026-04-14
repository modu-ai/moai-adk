package evolution_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

// TestEvaluateGraduation_Observation verifies that a single observation
// stays at StatusObservation.
func TestEvaluateGraduation_Observation(t *testing.T) {
	entry := &evolution.LearningEntry{
		Observations: 1,
		Confidence:   0.1,
		Status:       evolution.StatusObservation,
	}
	got := evolution.EvaluateGraduation(entry)
	if got != evolution.StatusObservation {
		t.Fatalf("expected StatusObservation, got %q", got)
	}
}

// TestEvaluateGraduation_Heuristic verifies that 3 observations promote to
// StatusHeuristic.
func TestEvaluateGraduation_Heuristic(t *testing.T) {
	entry := &evolution.LearningEntry{
		Observations: evolution.HeuristicThreshold,
		Confidence:   0.5,
		Status:       evolution.StatusObservation,
	}
	got := evolution.EvaluateGraduation(entry)
	if got != evolution.StatusHeuristic {
		t.Fatalf("expected StatusHeuristic, got %q", got)
	}
}

// TestEvaluateGraduation_Rule verifies that 5 observations + confidence >=
// 0.80 promote to StatusRule.
func TestEvaluateGraduation_Rule(t *testing.T) {
	entry := &evolution.LearningEntry{
		Observations: evolution.RuleThreshold,
		Confidence:   evolution.RuleConfidenceThreshold,
		Status:       evolution.StatusHeuristic,
	}
	got := evolution.EvaluateGraduation(entry)
	if got != evolution.StatusRule {
		t.Fatalf("expected StatusRule, got %q", got)
	}
}

// TestEvaluateGraduation_RuleBelowConfidence verifies that 5 observations
// with confidence < 0.80 do NOT reach StatusRule.
func TestEvaluateGraduation_RuleBelowConfidence(t *testing.T) {
	entry := &evolution.LearningEntry{
		Observations: evolution.RuleThreshold,
		Confidence:   0.79,
		Status:       evolution.StatusHeuristic,
	}
	got := evolution.EvaluateGraduation(entry)
	if got == evolution.StatusRule {
		t.Fatal("expected NOT StatusRule when confidence < threshold, got StatusRule")
	}
}

// TestEvaluateGraduation_HighConfidence verifies that 10+ observations
// promote to StatusHighConfidence.
func TestEvaluateGraduation_HighConfidence(t *testing.T) {
	entry := &evolution.LearningEntry{
		Observations: evolution.HighConfidenceThreshold,
		Confidence:   0.95,
		Status:       evolution.StatusRule,
	}
	got := evolution.EvaluateGraduation(entry)
	if got != evolution.StatusHighConfidence {
		t.Fatalf("expected StatusHighConfidence, got %q", got)
	}
}

// TestEvaluateGraduation_AntiPatternPreserved verifies that an anti-pattern
// entry is never upgraded by the graduation engine.
func TestEvaluateGraduation_AntiPatternPreserved(t *testing.T) {
	entry := &evolution.LearningEntry{
		Observations: 20,
		Confidence:   1.0,
		Status:       evolution.StatusAntiPattern,
	}
	got := evolution.EvaluateGraduation(entry)
	if got != evolution.StatusAntiPattern {
		t.Fatalf("expected StatusAntiPattern to be preserved, got %q", got)
	}
}

// TestCalculateConfidence_ConsistentSuccess verifies that a streak of
// successful outcomes yields a high confidence score.
func TestCalculateConfidence_ConsistentSuccess(t *testing.T) {
	entry := makeEntryWithSuccessOutcomes(10)
	conf := evolution.CalculateConfidence(entry)
	if conf < 0.70 {
		t.Fatalf("expected confidence >= 0.70 for consistent success, got %.2f", conf)
	}
}

// TestCalculateConfidence_MixedOutcomes verifies that mixed outcomes
// produce a moderate confidence score.
func TestCalculateConfidence_MixedOutcomes(t *testing.T) {
	entry := makeEntryWithMixedOutcomes(10)
	conf := evolution.CalculateConfidence(entry)
	// Mixed should be below the consistent-success baseline.
	if conf >= 0.85 {
		t.Fatalf("expected confidence < 0.85 for mixed outcomes, got %.2f", conf)
	}
}

// TestCalculateConfidence_ZeroObservations verifies that an empty entry
// returns confidence 0.
func TestCalculateConfidence_ZeroObservations(t *testing.T) {
	entry := &evolution.LearningEntry{}
	conf := evolution.CalculateConfidence(entry)
	if conf != 0.0 {
		t.Fatalf("expected 0.0 for zero observations, got %.2f", conf)
	}
}

// TestDetectDuplicate_MatchFound verifies that a near-identical observation
// string is detected as a duplicate.
func TestDetectDuplicate_MatchFound(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	// Create an existing learning.
	existing := &evolution.LearningEntry{
		ID:          "LEARN-20260101-001",
		SkillID:     "moai-lang-go",
		ZoneID:      "best-practices",
		Observation: "go test fails when context is missing from function signatures",
		Status:      evolution.StatusObservation,
		Observations: 1,
		Created:     time.Now(),
		Updated:     time.Now(),
	}
	if err := evolution.CreateLearning(projectRoot, existing); err != nil {
		t.Fatalf("setup: failed to create learning: %v", err)
	}

	// Query with a sufficiently similar observation.
	dup, err := evolution.DetectDuplicate(
		projectRoot,
		"go test fails when context is missing from function signature",
	)
	if err != nil {
		t.Fatalf("DetectDuplicate: %v", err)
	}
	if dup == nil {
		t.Fatal("expected duplicate to be detected, got nil")
	}
	if dup.ID != existing.ID {
		t.Fatalf("expected duplicate ID %q, got %q", existing.ID, dup.ID)
	}
}

// TestDetectDuplicate_NoMatch verifies that a dissimilar observation is not
// flagged as a duplicate.
func TestDetectDuplicate_NoMatch(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	existing := &evolution.LearningEntry{
		ID:          "LEARN-20260101-001",
		SkillID:     "moai-lang-go",
		ZoneID:      "best-practices",
		Observation: "go test fails when context is missing",
		Status:      evolution.StatusObservation,
		Observations: 1,
		Created:     time.Now(),
		Updated:     time.Now(),
	}
	if err := evolution.CreateLearning(projectRoot, existing); err != nil {
		t.Fatalf("setup: %v", err)
	}

	dup, err := evolution.DetectDuplicate(projectRoot, "completely unrelated observation about TypeScript")
	if err != nil {
		t.Fatalf("DetectDuplicate: %v", err)
	}
	if dup != nil {
		t.Fatalf("expected no duplicate, got %v", dup)
	}
}

// TestDetectAntiPattern_CriticalFailure verifies that a high-error-rate entry
// is flagged as an anti-pattern.
func TestDetectAntiPattern_CriticalFailure(t *testing.T) {
	entry := makeCriticalFailureEntry()
	if !evolution.DetectAntiPattern(entry) {
		t.Fatal("expected DetectAntiPattern to return true for critical failure")
	}
}

// TestDetectAntiPattern_NormalEntry verifies that a normal entry is not
// flagged as an anti-pattern.
func TestDetectAntiPattern_NormalEntry(t *testing.T) {
	entry := makeEntryWithSuccessOutcomes(5)
	if evolution.DetectAntiPattern(entry) {
		t.Fatal("expected DetectAntiPattern to return false for normal entry")
	}
}

// --- helpers ---

func makeEntryWithSuccessOutcomes(n int) *evolution.LearningEntry {
	e := &evolution.LearningEntry{
		Observations: n,
		Status:       evolution.StatusObservation,
	}
	for i := 0; i < n; i++ {
		e.Evidence = append(e.Evidence, evolution.EvidenceEntry{
			SessionID: fmt.Sprintf("sess-%d", i),
			Date:      "2026-04-11",
			Context:   "success: tool invocation completed",
		})
	}
	return e
}

func makeEntryWithMixedOutcomes(n int) *evolution.LearningEntry {
	e := &evolution.LearningEntry{
		Observations: n,
		Status:       evolution.StatusObservation,
	}
	for i := 0; i < n; i++ {
		ctx := "success: completed"
		if i%2 == 0 {
			ctx = "error: build failed"
		}
		e.Evidence = append(e.Evidence, evolution.EvidenceEntry{
			SessionID: fmt.Sprintf("sess-%d", i),
			Date:      "2026-04-11",
			Context:   ctx,
		})
	}
	return e
}

func makeCriticalFailureEntry() *evolution.LearningEntry {
	e := &evolution.LearningEntry{
		Observations: 1,
		Status:       evolution.StatusObservation,
	}
	// A single critical failure: the observation itself signals failure.
	e.Evidence = append(e.Evidence, evolution.EvidenceEntry{
		SessionID: "sess-critical",
		Date:      "2026-04-11",
		Context:   "error: critical build failure — score dropped > 0.20",
	})
	return e
}
