package lessons

import "fmt"

// Injector retrieves relevant lessons for injection into new SPEC contexts.
type Injector struct {
	store LessonStore
	limit int
}

// NewInjector creates a new lesson injector.
func NewInjector(store LessonStore, limit int) *Injector {
	if limit <= 0 {
		limit = 5
	}
	return &Injector{store: store, limit: limit}
}

// GetRelevantLessons returns the most relevant active lessons for the given tags.
func (i *Injector) GetRelevantLessons(tags []string) ([]*Lesson, error) {
	active := true
	filter := LessonFilter{
		Tags:   tags,
		Active: &active,
		Limit:  i.limit,
	}
	return i.store.List(filter)
}

// GetQualityLessons returns lessons from quality gate failures.
func (i *Injector) GetQualityLessons() ([]*Lesson, error) {
	active := true
	filter := LessonFilter{
		Type:   "quality_failure",
		Active: &active,
		Limit:  i.limit,
	}
	return i.store.List(filter)
}

// GetCriticLessons returns high-value lessons from critic sessions.
func (i *Injector) GetCriticLessons() ([]*Lesson, error) {
	active := true
	filter := LessonFilter{
		Type:   "critic_pattern",
		Tags:   []string{"high_value"},
		Active: &active,
		Limit:  i.limit,
	}
	return i.store.List(filter)
}

// FormatForInjection formats lessons as a string suitable for SPEC context injection.
func (i *Injector) FormatForInjection(lessons []*Lesson) string {
	if len(lessons) == 0 {
		return ""
	}

	result := "## Lessons from Previous SPECs\n\n"
	for idx, l := range lessons {
		result += fmt.Sprintf("%d. [%s] %s (Source: %s, Hits: %d)\n",
			idx+1, l.Severity, l.Pattern, l.Source, l.HitCount)
	}
	return result
}
