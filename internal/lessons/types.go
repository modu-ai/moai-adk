// Package lessons provides a learning system that captures patterns from
// quality gate failures, SPEC completions, and challenge sessions to improve
// future development cycles.
package lessons

import (
	"time"
)

// Lesson represents a single learned pattern or insight.
type Lesson struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`       // "quality_failure", "spec_completion", "challenge_pattern"
	Source    string    `json:"source"`     // SPEC-ID or event that triggered this lesson
	Pattern   string    `json:"pattern"`    // The pattern description
	Severity  string    `json:"severity"`   // "high", "medium", "low"
	Tags      []string  `json:"tags"`       // Searchable tags
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HitCount  int       `json:"hit_count"`  // Times this lesson was relevant
	Active    bool      `json:"active"`     // Whether lesson is active or archived
}

// LessonFilter defines criteria for querying lessons.
type LessonFilter struct {
	Type     string   // Filter by lesson type
	Tags     []string // Filter by tags (any match)
	Active   *bool    // Filter by active status (nil = all)
	MinHits  int      // Minimum hit count
	Limit    int      // Max results (0 = unlimited)
}

// LessonStore defines the interface for lesson persistence.
type LessonStore interface {
	// Save persists a lesson. If a lesson with the same ID exists, it updates.
	Save(lesson *Lesson) error

	// List returns lessons matching the filter criteria.
	List(filter LessonFilter) ([]*Lesson, error)

	// Get retrieves a single lesson by ID.
	Get(id string) (*Lesson, error)

	// Archive marks a lesson as inactive.
	Archive(id string) error

	// Count returns the number of active lessons.
	Count() (int, error)
}
