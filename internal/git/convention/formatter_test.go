package convention

import (
	"regexp"
	"strings"
	"testing"
)

func TestFormatViolation_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		v        Violation
		contains string
	}{
		{
			name:     "pattern without suggestion",
			v:        Violation{Type: ViolationPattern, Expected: `^(feat|fix): .+`},
			contains: "Pattern mismatch",
		},
		{
			name:     "pattern with suggestion",
			v:        Violation{Type: ViolationPattern, Expected: "pattern", Suggestion: "feat: do something"},
			contains: "Suggestion: feat: do something",
		},
		{
			name:     "invalid type",
			v:        Violation{Type: ViolationInvalidType, Actual: "foo", Expected: "feat, fix"},
			contains: `Invalid type "foo"`,
		},
		{
			name:     "invalid scope",
			v:        Violation{Type: ViolationInvalidScope, Actual: "bar", Expected: "api, cli"},
			contains: `Invalid scope "bar"`,
		},
		{
			name:     "max length",
			v:        Violation{Type: ViolationMaxLength, Actual: "120 characters", Expected: "max 100 characters"},
			contains: "Header too long",
		},
		{
			name:     "required",
			v:        Violation{Type: ViolationRequired, Field: "header"},
			contains: "Missing required field: header",
		},
		{
			name:     "unknown type fallback",
			v:        Violation{Type: "other", Expected: "a", Actual: "b"},
			contains: "expected a, got b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatViolation(tt.v)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("FormatViolation() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestFormatError_ValidResult(t *testing.T) {
	result := ValidationResult{Valid: true, Message: "feat: add feature"}
	conv := &Convention{Name: "test"}

	got := FormatError(result, conv)
	if got != "" {
		t.Errorf("FormatError(valid) = %q, want empty string", got)
	}
}

func TestFormatError_SingleViolation(t *testing.T) {
	conv := &Convention{
		Name: "conventional-commits",
		Examples: []string{
			"feat: example one",
		},
	}

	result := ValidationResult{
		Valid:   false,
		Message: "bad message",
		Violations: []Violation{
			{
				Type:     ViolationPattern,
				Expected: `^(feat|fix): .+`,
				Actual:   "bad message",
			},
		},
	}

	got := FormatError(result, conv)

	if !strings.Contains(got, "violates conventional-commits") {
		t.Error("expected convention name in output")
	}
	if !strings.Contains(got, `"bad message"`) {
		t.Error("expected quoted message in output")
	}
	if !strings.Contains(got, "[PATTERN]") {
		t.Error("expected [PATTERN] icon in output")
	}
	if !strings.Contains(got, "feat: example one") {
		t.Error("expected example in output")
	}
}

func TestFormatError_MultipleViolations(t *testing.T) {
	conv := &Convention{
		Name:    "test",
		Pattern: regexp.MustCompile(`^.+`),
	}

	result := ValidationResult{
		Valid:   false,
		Message: "bad",
		Violations: []Violation{
			{Type: ViolationPattern, Expected: "pattern"},
			{Type: ViolationMaxLength, Actual: "120 chars", Expected: "max 100 chars"},
		},
	}

	got := FormatError(result, conv)

	if !strings.Contains(got, "[PATTERN]") {
		t.Error("expected [PATTERN] in output")
	}
	if !strings.Contains(got, "[LENGTH]") {
		t.Error("expected [LENGTH] in output")
	}
}

func TestFormatError_NoExamples(t *testing.T) {
	conv := &Convention{
		Name:     "test",
		Examples: []string{},
	}

	result := ValidationResult{
		Valid:   false,
		Message: "bad",
		Violations: []Violation{
			{Type: ViolationRequired, Field: "header"},
		},
	}

	got := FormatError(result, conv)
	if strings.Contains(got, "Examples") {
		t.Error("should not include examples section when empty")
	}
}

func TestFormatBatchSummary_AllValid(t *testing.T) {
	conv := &Convention{Name: "conventional-commits"}

	results := []ValidationResult{
		{Valid: true, Message: "feat: one"},
		{Valid: true, Message: "fix: two"},
		{Valid: true, Message: "docs: three"},
	}

	got := FormatBatchSummary(results, conv)
	if !strings.Contains(got, "All 3 commits follow conventional-commits convention") {
		t.Errorf("FormatBatchSummary(all valid) = %q, want all-valid message", got)
	}
}

func TestFormatBatchSummary_WithViolations(t *testing.T) {
	conv := &Convention{Name: "test"}

	results := []ValidationResult{
		{Valid: true, Message: "feat: one"},
		{Valid: false, Message: "bad", Violations: []Violation{{Type: ViolationPattern, Expected: "p"}}},
		{Valid: true, Message: "fix: three"},
	}

	got := FormatBatchSummary(results, conv)
	if !strings.Contains(got, "1 of 3 commits violate test convention") {
		t.Errorf("FormatBatchSummary = %q, want violation count", got)
	}
}

func TestFormatBatchSummary_Empty(t *testing.T) {
	conv := &Convention{Name: "test"}

	got := FormatBatchSummary(nil, conv)
	if got != "No commits to validate." {
		t.Errorf("FormatBatchSummary(nil) = %q, want empty message", got)
	}

	got = FormatBatchSummary([]ValidationResult{}, conv)
	if got != "No commits to validate." {
		t.Errorf("FormatBatchSummary([]) = %q, want empty message", got)
	}
}

func TestViolationIcon_AllTypes(t *testing.T) {
	tests := []struct {
		vt   ViolationType
		want string
	}{
		{ViolationPattern, "[PATTERN]"},
		{ViolationInvalidType, "[TYPE]"},
		{ViolationInvalidScope, "[SCOPE]"},
		{ViolationMaxLength, "[LENGTH]"},
		{ViolationRequired, "[REQUIRED]"},
		{"unknown", "[ERROR]"},
	}

	for _, tt := range tests {
		t.Run(string(tt.vt), func(t *testing.T) {
			got := violationIcon(tt.vt)
			if got != tt.want {
				t.Errorf("violationIcon(%q) = %q, want %q", tt.vt, got, tt.want)
			}
		})
	}
}
