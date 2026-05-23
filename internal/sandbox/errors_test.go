package sandbox

import (
	"errors"
	"fmt"
	"testing"
)

// TestErrors_SentinelMatching verifies that all sentinel errors can be detected
// via errors.Is after wrapping.
// RED: fails until errors.go is created with sentinel values.
func TestErrors_SentinelMatching(t *testing.T) {
	t.Parallel()

	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrSandboxBackendUnavailable", ErrSandboxBackendUnavailable},
		{"ErrSandboxProfileInvalid", ErrSandboxProfileInvalid},
		{"ErrSandboxRequired", ErrSandboxRequired},
		{"ErrSandboxOutputTruncated", ErrSandboxOutputTruncated},
		{"ErrSandboxSetuidDenied", ErrSandboxSetuidDenied},
	}

	for _, tc := range sentinels {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Direct comparison.
			if !errors.Is(tc.err, tc.err) {
				t.Errorf("%s: errors.Is(err, err) returned false", tc.name)
			}

			// Comparison after wrapping.
			wrapped := fmt.Errorf("outer: %w", tc.err)
			if !errors.Is(wrapped, tc.err) {
				t.Errorf("%s: errors.Is(wrapped, sentinel) returned false", tc.name)
			}

			// The sentinel must not be nil.
			if tc.err == nil {
				t.Errorf("%s must not be nil", tc.name)
			}

			// Error() must return a non-empty string.
			if tc.err.Error() == "" {
				t.Errorf("%s Error() returned empty string", tc.name)
			}
		})
	}
}

// TestErrors_Wrapping verifies that sentinel errors support wrapping via Unwrap().
// RED: fails until errors.go provides unwrappable sentinels.
func TestErrors_Wrapping(t *testing.T) {
	t.Parallel()

	// SandboxBackendUnavailable is the primary "fail-hard" sentinel — verify wrapping
	base := ErrSandboxBackendUnavailable
	wrapped := fmt.Errorf("backend check failed: %w", base)

	if !errors.Is(wrapped, ErrSandboxBackendUnavailable) {
		t.Error("wrapped ErrSandboxBackendUnavailable not detectable via errors.Is")
	}

	// Double-wrapping
	doubleWrapped := fmt.Errorf("outer: %w", wrapped)
	if !errors.Is(doubleWrapped, ErrSandboxBackendUnavailable) {
		t.Error("double-wrapped error not detectable via errors.Is")
	}

	// Distinct sentinels must not match each other
	if errors.Is(ErrSandboxProfileInvalid, ErrSandboxBackendUnavailable) {
		t.Error("ErrSandboxProfileInvalid should not match ErrSandboxBackendUnavailable")
	}
	if errors.Is(ErrSandboxRequired, ErrSandboxOutputTruncated) {
		t.Error("ErrSandboxRequired should not match ErrSandboxOutputTruncated")
	}
}
