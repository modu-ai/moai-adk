package sandbox

import (
	"errors"
	"testing"
)

// TestErrors_SentinelMatching verifies that sentinel errors can be
// matched using errors.Is.
// REQ-V3R2-RT-003-012, REQ-V3R2-RT-003-041, REQ-V3R2-RT-003-043
func TestErrors_SentinelMatching(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
	}{
		{
			name:   "ErrSandboxBackendUnavailable",
			err:    ErrSandboxBackendUnavailable,
			target: ErrSandboxBackendUnavailable,
		},
		{
			name:   "ErrSandboxProfileInvalid",
			err:    ErrSandboxProfileInvalid,
			target: ErrSandboxProfileInvalid,
		},
		{
			name:   "ErrSandboxRequired",
			err:    ErrSandboxRequired,
			target: ErrSandboxRequired,
		},
		{
			name:   "ErrSandboxOutputTruncated",
			err:    ErrSandboxOutputTruncated,
			target: ErrSandboxOutputTruncated,
		},
		{
			name:   "ErrSandboxSetuidDenied",
			err:    ErrSandboxSetuidDenied,
			target: ErrSandboxSetuidDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.target) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.target)
			}
		})
	}
}

// TestErrors_Wrapping verifies that sentinel errors support proper
// error wrapping with %w.
func TestErrors_Wrapping(t *testing.T) {
	baseErr := ErrSandboxBackendUnavailable
	wrappedErr := errors.New("context: %w", baseErr)

	if !errors.Is(wrappedErr, ErrSandboxBackendUnavailable) {
		t.Error("Wrapped error should still match sentinel via errors.Is")
	}

	// Test double wrapping
	doubleWrapped := errors.New("outer: %w", wrappedErr)
	if !errors.Is(doubleWrapped, ErrSandboxBackendUnavailable) {
		t.Error("Double-wrapped error should still match sentinel")
	}
}
