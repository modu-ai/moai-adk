package core

import (
	"errors"
	"fmt"
	"testing"
)

// TestErrNotImplemented_IsSemantics verifies that errors.Is works for wrapped ErrNotImplemented.
func TestErrNotImplemented_IsSemantics(t *testing.T) {
	wrapped := fmt.Errorf("OpenFile: %w", ErrNotImplemented)

	if !errors.Is(wrapped, ErrNotImplemented) {
		t.Error("expected errors.Is(wrapped, ErrNotImplemented) to be true")
	}
	if errors.Is(wrapped, ErrCapabilityUnsupported) {
		t.Error("expected errors.Is(wrapped, ErrCapabilityUnsupported) to be false")
	}
}

// TestErrCapabilityUnsupported_IsSemantics verifies error wrapping for ErrCapabilityUnsupported.
func TestErrCapabilityUnsupported_IsSemantics(t *testing.T) {
	wrapped := fmt.Errorf("textDocument/references: %w", ErrCapabilityUnsupported)

	if !errors.Is(wrapped, ErrCapabilityUnsupported) {
		t.Error("expected errors.Is(wrapped, ErrCapabilityUnsupported) to be true")
	}
	if errors.Is(wrapped, ErrNotImplemented) {
		t.Error("expected errors.Is(wrapped, ErrNotImplemented) to be false")
	}
}

// TestErrCapabilityUnsupported_DirectComparison verifies direct sentinel comparison.
func TestErrCapabilityUnsupported_DirectComparison(t *testing.T) {
	if ErrCapabilityUnsupported == nil {
		t.Fatal("ErrCapabilityUnsupported must not be nil")
	}
	if ErrCapabilityUnsupported == ErrNotImplemented {
		t.Error("ErrCapabilityUnsupported and ErrNotImplemented must be distinct")
	}
}

// TestErrNotImplemented_DirectComparison verifies direct sentinel comparison.
func TestErrNotImplemented_DirectComparison(t *testing.T) {
	if ErrNotImplemented == nil {
		t.Fatal("ErrNotImplemented must not be nil")
	}
}

// TestSentinel_MultiLevelWrapping verifies that errors.Is traverses multiple wrapping levels.
func TestSentinel_MultiLevelWrapping(t *testing.T) {
	inner := fmt.Errorf("capability: %w", ErrCapabilityUnsupported)
	outer := fmt.Errorf("query failed: %w", inner)

	if !errors.Is(outer, ErrCapabilityUnsupported) {
		t.Error("expected errors.Is to unwrap through multiple levels")
	}
}
