package mx

import (
	"strings"
)

// DangerCategoryConfig represents the danger_categories configuration from mx.yaml.
// Maps WARN REASON text to pattern lists (REQ-SPC-004-012).
//
// @MX:NOTE: [AUTO] DangerCategoryConfig — default category patterns are conservatively designed
// default values: concurrency(concurrency), resource-leak(resource leak), cleanup(cleanup), security(security)
type DangerCategoryConfig struct {
	// Categories is the category name → pattern list mapping.
	Categories map[string][]string `yaml:"danger_categories"`

	// TestPaths is the list of test file path patterns to exclude during fan-in calculation (REQ-SPC-004-040).
	TestPaths []string `yaml:"test_paths"`
}

// DefaultDangerCategories is the default danger category mapping used when mx.yaml is not configured.
var DefaultDangerCategories = map[string][]string{
	"concurrency": {
		"goroutine leak",
		"unbounded channel",
		"race condition",
	},
	"resource-leak": {
		"missing Close",
		"fd leak",
	},
	"cleanup": {
		"defer missing",
		"Close not called",
	},
	"security": {
		"hardcoded credential",
		"sql injection",
		"xss",
	},
}

// DangerCategoryMatcher matches WARN REASON text to danger categories.
type DangerCategoryMatcher struct {
	config DangerCategoryConfig
}

// NewDangerCategoryMatcher creates a DangerCategoryMatcher with the given configuration.
// Uses default categories when configuration has no Categories.
func NewDangerCategoryMatcher(config DangerCategoryConfig) *DangerCategoryMatcher {
	if len(config.Categories) == 0 {
		config.Categories = DefaultDangerCategories
	}
	return &DangerCategoryMatcher{config: config}
}

// Match verifies if reason text matches one of the patterns in the category.
// Uses case-insensitive partial string matching for pattern matching (REQ-SPC-004-012).
func (m *DangerCategoryMatcher) Match(reason, category string) bool {
	patterns, ok := m.config.Categories[category]
	if !ok {
		return false
	}

	lowerReason := strings.ToLower(reason)
	for _, pattern := range patterns {
		if strings.Contains(lowerReason, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// CategoryOf returns the first category that matches the reason text.
func (m *DangerCategoryMatcher) CategoryOf(reason string) string {
	lowerReason := strings.ToLower(reason)
	for cat, patterns := range m.config.Categories {
		for _, pattern := range patterns {
			if strings.Contains(lowerReason, strings.ToLower(pattern)) {
				return cat
			}
		}
	}
	return ""
}

// ValidateCategory verifies if the given category is a known category.
func (m *DangerCategoryMatcher) ValidateCategory(category string) bool {
	_, ok := m.config.Categories[category]
	return ok
}

// KnownCategories returns all category names defined in the configuration.
func (m *DangerCategoryMatcher) KnownCategories() []string {
	result := make([]string, 0, len(m.config.Categories))
	for cat := range m.config.Categories {
		result = append(result, cat)
	}
	return result
}
