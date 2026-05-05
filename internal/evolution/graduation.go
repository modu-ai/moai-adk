package evolution

import (
	"math"
	"strings"
	"time"
)

// @MX:NOTE: [AUTO] Graduation pipeline: observation→heuristic (3x)→rule (5x+0.80 conf)→high-confidence (10x)→graduated
//
// EvaluateGraduation returns the target status tier for entry based on its
// current observation count and confidence score.
//
// Tier rules:
//   - StatusAntiPattern: preserved unconditionally.
//   - StatusHighConfidence: Observations >= HighConfidenceThreshold.
//   - StatusRule: Observations >= RuleThreshold AND Confidence >= RuleConfidenceThreshold.
//   - StatusHeuristic: Observations >= HeuristicThreshold.
//   - StatusObservation: otherwise.
func EvaluateGraduation(entry *LearningEntry) Status {
	// Anti-patterns are immutable by the engine.
	if entry.Status == StatusAntiPattern {
		return StatusAntiPattern
	}
	// Already graduated — keep.
	if entry.Status == StatusGraduated {
		return StatusGraduated
	}

	switch {
	case entry.Observations >= HighConfidenceThreshold:
		return StatusHighConfidence
	case entry.Observations >= RuleThreshold && entry.Confidence >= RuleConfidenceThreshold:
		return StatusRule
	case entry.Observations >= HeuristicThreshold:
		return StatusHeuristic
	default:
		return StatusObservation
	}
}

// CalculateConfidence computes a confidence score in [0, 1] for entry using
// three weighted components:
//
//   - outcome_consistency (weight 0.4): ratio of success-like evidence entries
//     to total evidence.
//   - frequency (weight 0.3): observation count normalised against
//     HighConfidenceThreshold.
//   - recency (weight 0.3): exponential decay based on days since last update.
//
// Returns 0 when the entry has no observations.
func CalculateConfidence(entry *LearningEntry) float64 {
	if entry.Observations == 0 || len(entry.Evidence) == 0 {
		return 0.0
	}

	// Outcome consistency: count evidence entries that are NOT errors.
	successCount := 0
	for _, ev := range entry.Evidence {
		if !isErrorContext(ev.Context) {
			successCount++
		}
	}
	consistency := float64(successCount) / float64(len(entry.Evidence))

	// Frequency: capped at 1.0.
	frequency := math.Min(float64(entry.Observations)/float64(HighConfidenceThreshold), 1.0)

	// Recency: exponential decay with half-life of 7 days.
	// score = exp(-lambda * days) where lambda = ln(2)/7.
	daysSince := 0.0
	if !entry.Updated.IsZero() {
		daysSince = time.Since(entry.Updated).Hours() / 24
	}
	lambda := math.Log(2) / 7.0
	recency := math.Exp(-lambda * daysSince)

	const (
		wConsistency = 0.4
		wFrequency   = 0.3
		wRecency     = 0.3
	)

	score := wConsistency*consistency + wFrequency*frequency + wRecency*recency
	// Clamp to [0, 1].
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score
}

// isErrorContext reports whether an evidence context string indicates an error.
func isErrorContext(ctx string) bool {
	lower := strings.ToLower(ctx)
	for _, kw := range []string{"error", "fail", "critical", "fatal", "panic"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// DetectDuplicate searches existing learnings for an entry whose observation
// string is sufficiently similar to the provided observation.
//
// Similarity is measured by token overlap: if more than 60% of the tokens in
// the query appear in the existing observation (case-insensitive), the entries
// are considered duplicates.
//
// Returns the first matching LearningEntry, or nil if no duplicate is found.
func DetectDuplicate(projectRoot, observation string) (*LearningEntry, error) {
	entries, err := ListLearnings(projectRoot, LearningFilter{ExcludeArchived: true})
	if err != nil {
		return nil, err
	}

	queryTokens := tokenise(observation)
	if len(queryTokens) == 0 {
		return nil, nil
	}

	for _, e := range entries {
		existing := tokenise(e.Observation)
		if len(existing) == 0 {
			continue
		}
		overlap := tokenOverlap(queryTokens, existing)
		if overlap >= 0.60 {
			return e, nil
		}
	}
	return nil, nil
}

// DetectAntiPattern reports whether entry should be immediately classified as
// an anti-pattern.  The heuristic checks for the presence of "critical" error
// keywords in the evidence context strings.
func DetectAntiPattern(entry *LearningEntry) bool {
	for _, ev := range entry.Evidence {
		lower := strings.ToLower(ev.Context)
		for _, kw := range []string{"critical", "score dropped", "fatal"} {
			if strings.Contains(lower, kw) {
				return true
			}
		}
	}
	return false
}

// tokenise splits s into lowercase words suitable for overlap comparison.
func tokenise(s string) []string {
	// Replace punctuation with spaces and split on whitespace.
	var buf strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == ' ' {
			buf.WriteRune(r)
		} else {
			buf.WriteRune(' ')
		}
	}
	parts := strings.Fields(buf.String())

	// Filter stop words.
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "in": true,
		"of": true, "to": true, "and": true, "for": true, "or": true,
		"with": true, "that": true, "this": true, "when": true,
	}
	result := parts[:0]
	for _, p := range parts {
		if !stopWords[p] {
			result = append(result, p)
		}
	}
	return result
}

// tokenOverlap returns the Jaccard-like overlap between two token sets:
// |intersection| / |query|.
func tokenOverlap(query, target []string) float64 {
	if len(query) == 0 {
		return 0
	}
	targetSet := make(map[string]bool, len(target))
	for _, t := range target {
		targetSet[t] = true
	}
	matches := 0
	for _, q := range query {
		if targetSet[q] {
			matches++
		}
	}
	return float64(matches) / float64(len(query))
}
