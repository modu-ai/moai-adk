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
