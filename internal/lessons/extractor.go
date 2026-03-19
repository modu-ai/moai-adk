package lessons

import (
	"fmt"
	"time"
)

// Extractor creates lessons from various system events.
type Extractor struct {
	store LessonStore
}

// NewExtractor creates a new lesson extractor.
func NewExtractor(store LessonStore) *Extractor {
	return &Extractor{store: store}
}

// ExtractFromQualityFailure creates a lesson from a quality gate failure.
func (e *Extractor) ExtractFromQualityFailure(specID, failureType, description string, tags []string) error {
	lesson := &Lesson{
		ID:       fmt.Sprintf("qf-%s-%d", specID, time.Now().UnixMilli()),
		Type:     "quality_failure",
		Source:   specID,
		Pattern:  fmt.Sprintf("Quality gate failure (%s): %s", failureType, description),
		Severity: "high",
		Tags:     append(tags, "quality", failureType),
		Active:   true,
		HitCount: 1,
	}
	return e.store.Save(lesson)
}

// ExtractFromSPECCompletion creates a lesson from SPEC completion metrics.
func (e *Extractor) ExtractFromSPECCompletion(specID string, iterations int, tags []string) error {
	severity := "low"
	if iterations > 5 {
		severity = "high"
	} else if iterations > 3 {
		severity = "medium"
	}

	lesson := &Lesson{
		ID:       fmt.Sprintf("sc-%s-%d", specID, time.Now().UnixMilli()),
		Type:     "spec_completion",
		Source:   specID,
		Pattern:  fmt.Sprintf("SPEC %s completed with %d iterations", specID, iterations),
		Severity: severity,
		Tags:     append(tags, "spec_completion"),
		Active:   true,
		HitCount: 1,
	}
	return e.store.Save(lesson)
}

// ExtractFromChallenge creates lessons from challenge session outcomes.
func (e *Extractor) ExtractFromChallenge(specID string, highValueQuestions, dismissedQuestions []string) error {
	for _, q := range highValueQuestions {
		lesson := &Lesson{
			ID:       fmt.Sprintf("ch-%s-hv-%d", specID, time.Now().UnixMilli()),
			Type:     "challenge_pattern",
			Source:   specID,
			Pattern:  fmt.Sprintf("High-value challenge question: %s", q),
			Severity: "high",
			Tags:     []string{"challenge", "high_value"},
			Active:   true,
			HitCount: 1,
		}
		if err := e.store.Save(lesson); err != nil {
			return err
		}
	}

	for _, q := range dismissedQuestions {
		lesson := &Lesson{
			ID:       fmt.Sprintf("ch-%s-lv-%d", specID, time.Now().UnixMilli()),
			Type:     "challenge_pattern",
			Source:   specID,
			Pattern:  fmt.Sprintf("Low-value challenge question (dismissed): %s", q),
			Severity: "low",
			Tags:     []string{"challenge", "low_value"},
			Active:   true,
			HitCount: 1,
		}
		if err := e.store.Save(lesson); err != nil {
			return err
		}
	}
	return nil
}
