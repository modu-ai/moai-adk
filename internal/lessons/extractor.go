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
	allTags := make([]string, 0, len(tags)+2)
	allTags = append(allTags, tags...)
	allTags = append(allTags, "quality", failureType)
	lesson := &Lesson{
		ID:       fmt.Sprintf("qf-%s-%d", specID, time.Now().UnixMilli()),
		Type:     "quality_failure",
		Source:   specID,
		Pattern:  fmt.Sprintf("Quality gate failure (%s): %s", failureType, description),
		Severity: "high",
		Tags:     allTags,
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

	allTags := make([]string, 0, len(tags)+1)
	allTags = append(allTags, tags...)
	allTags = append(allTags, "spec_completion")
	lesson := &Lesson{
		ID:       fmt.Sprintf("sc-%s-%d", specID, time.Now().UnixMilli()),
		Type:     "spec_completion",
		Source:   specID,
		Pattern:  fmt.Sprintf("SPEC %s completed with %d iterations", specID, iterations),
		Severity: severity,
		Tags:     allTags,
		Active:   true,
		HitCount: 1,
	}
	return e.store.Save(lesson)
}

// ExtractFromCritic creates lessons from critic session outcomes.
func (e *Extractor) ExtractFromCritic(specID string, highValueQuestions, dismissedQuestions []string) error {
	ts := time.Now().UnixMilli()
	for i, q := range highValueQuestions {
		lesson := &Lesson{
			ID:       fmt.Sprintf("ch-%s-hv-%d-%d", specID, ts, i),
			Type:     "critic_pattern",
			Source:   specID,
			Pattern:  fmt.Sprintf("High-value critic question: %s", q),
			Severity: "high",
			Tags:     []string{"critic", "high_value"},
			Active:   true,
			HitCount: 1,
		}
		if err := e.store.Save(lesson); err != nil {
			return err
		}
	}

	for i, q := range dismissedQuestions {
		lesson := &Lesson{
			ID:       fmt.Sprintf("ch-%s-lv-%d-%d", specID, ts, i),
			Type:     "critic_pattern",
			Source:   specID,
			Pattern:  fmt.Sprintf("Low-value critic question (dismissed): %s", q),
			Severity: "low",
			Tags:     []string{"critic", "low_value"},
			Active:   true,
			HitCount: 1,
		}
		if err := e.store.Save(lesson); err != nil {
			return err
		}
	}
	return nil
}
