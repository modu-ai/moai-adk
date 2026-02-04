package foundation

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrInvalidRequirementType",
			err:  ErrInvalidRequirementType,
			want: "invalid requirement type",
		},
		{
			name: "ErrInvalidPillar",
			err:  ErrInvalidPillar,
			want: "invalid TRUST 5 pillar",
		},
		{
			name: "ErrAssessmentFailed",
			err:  ErrAssessmentFailed,
			want: "assessment failed: not all pillars meet threshold",
		},
		{
			name: "ErrInvalidPhaseTransition",
			err:  ErrInvalidPhaseTransition,
			want: "invalid phase transition",
		},
		{
			name: "ErrUnsupportedLanguage",
			err:  ErrUnsupportedLanguage,
			want: "unsupported language",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err == nil {
				t.Fatal("error should not be nil")
			}
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrInvalidRequirementType,
		ErrInvalidPillar,
		ErrAssessmentFailed,
		ErrInvalidPhaseTransition,
		ErrUnsupportedLanguage,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinel errors at index %d and %d should be distinct", i, j)
			}
		}
	}
}

func TestRequirementNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "with_ID", id: "REQ-001", want: "requirement not found: REQ-001"},
		{name: "empty_ID", id: "", want: "requirement not found: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &RequirementNotFoundError{ID: tt.id}
			if err.Error() != tt.want {
				t.Errorf("got %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRequirementNotFoundErrorImplementsError(t *testing.T) {
	t.Parallel()

	var err error = &RequirementNotFoundError{ID: "REQ-001"}
	// Verify the error implements the error interface by using it
	_ = err.Error()
}

func TestLanguageNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "by_ID", query: "brainfuck", want: "language not found: brainfuck"},
		{name: "by_extension", query: ".xyz", want: "language not found: .xyz"},
		{name: "empty_query", query: "", want: "language not found: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &LanguageNotFoundError{Query: tt.query}
			if err.Error() != tt.want {
				t.Errorf("got %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestLanguageNotFoundErrorImplementsError(t *testing.T) {
	t.Parallel()

	var err error = &LanguageNotFoundError{Query: ".xyz"}
	// Verify the error implements the error interface by using it
	_ = err.Error()
}

func TestErrorWrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sentinel error
	}{
		{name: "ErrInvalidRequirementType", sentinel: ErrInvalidRequirementType},
		{name: "ErrInvalidPillar", sentinel: ErrInvalidPillar},
		{name: "ErrAssessmentFailed", sentinel: ErrAssessmentFailed},
		{name: "ErrInvalidPhaseTransition", sentinel: ErrInvalidPhaseTransition},
		{name: "ErrUnsupportedLanguage", sentinel: ErrUnsupportedLanguage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wrapped := fmt.Errorf("context: %w", tt.sentinel)
			if !errors.Is(wrapped, tt.sentinel) {
				t.Errorf("wrapped error should match sentinel %v", tt.sentinel)
			}
		})
	}
}
