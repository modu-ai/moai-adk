// Package harness — meta_invocation_test.go tests the FROZEN guard
// (frozen_guard.go) and cleanup tracker (cleanup.go) introduced in
// SPEC-V3R3-PROJECT-HARNESS-001 Phase 2 (T-P2-02, T-P2-03, T-P2-04).
package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// IsAllowedPath — allowed prefixes
// ---------------------------------------------------------------------------

// TestIsAllowedPath_Allowed verifies that each allowed write-prefix
// returns (true, nil).
func TestIsAllowedPath_Allowed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{"agents my-harness root", ".claude/agents/my-harness/"},
		{"agents my-harness deep", ".claude/agents/my-harness/ios-architect.md"},
		{"skills my-harness- root", ".claude/skills/my-harness-ios-patterns/"},
		{"skills my-harness- deep", ".claude/skills/my-harness-swiftui/SKILL.md"},
		{"moai harness root", ".moai/harness/"},
		{"moai harness file", ".moai/harness/main.md"},
		{"moai harness deep", ".moai/harness/chaining-rules.yaml"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := IsAllowedPath(tc.path)
			if err != nil {
				t.Fatalf("IsAllowedPath(%q): unexpected error: %v", tc.path, err)
			}
			if !got {
				t.Errorf("IsAllowedPath(%q) = false, want true", tc.path)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsAllowedPath — forbidden prefixes (FROZEN violation)
// ---------------------------------------------------------------------------

// TestIsAllowedPath_Forbidden verifies that each moai-managed prefix
// returns (false, *FrozenViolationError).
func TestIsAllowedPath_Forbidden(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{"agents moai direct", ".claude/agents/moai/ios-architect.md"},
		{"skills moai-foundation", ".claude/skills/moai-foundation-core/SKILL.md"},
		{"skills moai/ prefix", ".claude/skills/moai/workflows/run.md"},
		{"rules moai", ".claude/rules/moai/core/moai-constitution.md"},
		{"agents moai root", ".claude/agents/moai/"},
		{"skills moai- root", ".claude/skills/moai-"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := IsAllowedPath(tc.path)
			if err == nil {
				t.Fatalf("IsAllowedPath(%q): expected error, got nil (allowed=%v)", tc.path, got)
			}
			var fve *FrozenViolationError
			if !errors.As(err, &fve) {
				t.Errorf("IsAllowedPath(%q): error is not *FrozenViolationError: %T %v", tc.path, err, err)
			}
			if fve != nil && fve.Path != tc.path {
				t.Errorf("FrozenViolationError.Path = %q, want %q", fve.Path, tc.path)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsAllowedPath — neutral paths (not allowed, not forbidden)
// ---------------------------------------------------------------------------

// TestIsAllowedPath_Neutral verifies that paths matching neither the
// allowed nor the forbidden prefix list return (false, nil).
func TestIsAllowedPath_Neutral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{"internal harness file", "internal/harness/foo.go"},
		{"cmd main", "cmd/main.go"},
		{"random path", "some/other/file.txt"},
		{"claude agents other", ".claude/agents/other-agent/foo.md"},
		{"moai config", ".moai/config/sections/quality.yaml"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := IsAllowedPath(tc.path)
			if err != nil {
				t.Fatalf("IsAllowedPath(%q): unexpected error: %v", tc.path, err)
			}
			if got {
				t.Errorf("IsAllowedPath(%q) = true, want false (neutral path)", tc.path)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsAllowedPath — traversal rejection
// ---------------------------------------------------------------------------

// TestIsAllowedPath_Traversal verifies that paths containing ".."
// are rejected with an error.
func TestIsAllowedPath_Traversal(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{"dotdot start", "../evil/path"},
		{"dotdot middle", ".claude/agents/../../evil"},
		{"dotdot end", ".moai/harness/.."},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := IsAllowedPath(tc.path)
			if err == nil {
				t.Errorf("IsAllowedPath(%q): expected error for traversal, got nil", tc.path)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsAllowedPath — empty path rejection
// ---------------------------------------------------------------------------

// TestIsAllowedPath_EmptyPath verifies that an empty string is rejected
// with an error.
func TestIsAllowedPath_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := IsAllowedPath("")
	if err == nil {
		t.Error("IsAllowedPath(\"\"): expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// EnsureAllowed
// ---------------------------------------------------------------------------

// TestEnsureAllowed_Reject verifies that EnsureAllowed returns a
// *FrozenViolationError whose Path field matches the forbidden input.
func TestEnsureAllowed_Reject(t *testing.T) {
	t.Parallel()

	forbidden := ".claude/agents/moai/something.md"
	err := EnsureAllowed(forbidden)
	if err == nil {
		t.Fatal("EnsureAllowed: expected error, got nil")
	}

	var fve *FrozenViolationError
	if !errors.As(err, &fve) {
		t.Fatalf("EnsureAllowed: error is not *FrozenViolationError: %T", err)
	}
	if fve.Path != forbidden {
		t.Errorf("FrozenViolationError.Path = %q, want %q", fve.Path, forbidden)
	}
}

// TestEnsureAllowed_Pass verifies that EnsureAllowed returns nil for
// explicitly allowed paths.
func TestEnsureAllowed_Pass(t *testing.T) {
	t.Parallel()

	err := EnsureAllowed(".claude/agents/my-harness/ios-architect.md")
	if err != nil {
		t.Errorf("EnsureAllowed (allowed path): unexpected error: %v", err)
	}
}

// TestEnsureAllowed_Neutral verifies that EnsureAllowed returns an error
// for neutral paths (not in the allowed-prefix list) because the guard
// blocks writes to unlisted paths by default.
func TestEnsureAllowed_Neutral(t *testing.T) {
	t.Parallel()

	err := EnsureAllowed("internal/harness/foo.go")
	if err == nil {
		t.Error("EnsureAllowed (neutral path): expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// FrozenViolationError — message format
// ---------------------------------------------------------------------------

// TestFrozenGuard_FROZEN_VIOLATION_Message verifies the error message
// format required by the SPEC: "FROZEN_VIOLATION: <path>: <reason>".
func TestFrozenGuard_FROZEN_VIOLATION_Message(t *testing.T) {
	t.Parallel()

	fve := &FrozenViolationError{
		Path:   ".claude/agents/moai/test.md",
		Reason: "moai-managed area is FROZEN",
	}

	msg := fve.Error()
	if !strings.HasPrefix(msg, "FROZEN_VIOLATION:") {
		t.Errorf("FrozenViolationError.Error() = %q; want prefix FROZEN_VIOLATION:", msg)
	}
	if !strings.Contains(msg, fve.Path) {
		t.Errorf("FrozenViolationError.Error() = %q; must contain path %q", msg, fve.Path)
	}
	if !strings.Contains(msg, fve.Reason) {
		t.Errorf("FrozenViolationError.Error() = %q; must contain reason %q", msg, fve.Reason)
	}
}

// ---------------------------------------------------------------------------
// CleanupTracker — track and cleanup
// ---------------------------------------------------------------------------

// TestCleanupTracker_Track_And_Cleanup creates real temporary files,
// tracks them, calls Cleanup, and verifies they are removed.
func TestCleanupTracker_Track_And_Cleanup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create two test files.
	fileA := filepath.Join(dir, "file_a.txt")
	fileB := filepath.Join(dir, "subdir", "file_b.txt")

	if err := os.MkdirAll(filepath.Dir(fileB), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for _, f := range []string{fileA, fileB} {
		if err := os.WriteFile(f, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", f, err)
		}
	}

	tracker := &CleanupTracker{}
	tracker.Track(fileA)
	tracker.Track(fileB)

	if n := len(tracker.Tracked()); n != 2 {
		t.Fatalf("Tracked() len = %d, want 2", n)
	}

	if err := tracker.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	for _, f := range []string{fileA, fileB} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("after Cleanup, file %s still exists", f)
		}
	}
}

// ---------------------------------------------------------------------------
// CleanupTracker — idempotency
// ---------------------------------------------------------------------------

// TestCleanupTracker_Idempotent verifies that calling Cleanup twice on
// already-removed files does not return an error.
func TestCleanupTracker_Idempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := filepath.Join(dir, "once.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tracker := &CleanupTracker{}
	tracker.Track(f)

	if err := tracker.Cleanup(); err != nil {
		t.Fatalf("first Cleanup: %v", err)
	}
	// Second call — file already gone.
	if err := tracker.Cleanup(); err != nil {
		t.Errorf("second Cleanup (idempotent): unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CleanupOnFailure — with error
// ---------------------------------------------------------------------------

// TestCleanupOnFailure_WithError verifies that when a non-nil error is
// passed, CleanupOnFailure invokes the tracker and removes tracked files.
func TestCleanupOnFailure_WithError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := filepath.Join(dir, "partial.md")
	if err := os.WriteFile(f, []byte("partial"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tracker := &CleanupTracker{}
	tracker.Track(f)

	origErr := errors.New("meta-harness failed mid-way")
	returned := CleanupOnFailure(tracker, origErr)
	if returned == nil {
		t.Fatal("CleanupOnFailure: expected non-nil error, got nil")
	}
	if !errors.Is(returned, origErr) {
		t.Errorf("CleanupOnFailure: returned error does not wrap original: %v", returned)
	}

	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Errorf("CleanupOnFailure: partial file %s still exists after cleanup", f)
	}
}

// ---------------------------------------------------------------------------
// CleanupOnFailure — without error
// ---------------------------------------------------------------------------

// TestCleanupOnFailure_WithoutError verifies that when err == nil,
// CleanupOnFailure is a no-op and tracked files are left intact.
func TestCleanupOnFailure_WithoutError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := filepath.Join(dir, "keep.md")
	if err := os.WriteFile(f, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tracker := &CleanupTracker{}
	tracker.Track(f)

	returned := CleanupOnFailure(tracker, nil)
	if returned != nil {
		t.Errorf("CleanupOnFailure(nil): expected nil, got %v", returned)
	}

	if _, err := os.Stat(f); os.IsNotExist(err) {
		t.Errorf("CleanupOnFailure(nil): file %s was removed, should not have been", f)
	}
}
