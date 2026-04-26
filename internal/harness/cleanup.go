// Package harness — cleanup.go provides CleanupTracker and CleanupOnFailure,
// used to roll back partial meta-harness output when generation fails
// mid-way (SPEC-V3R3-PROJECT-HARNESS-001 REQ-PH-004, REQ-PH-010, T-P2-03).
package harness

import (
	"errors"
	"fmt"
	"os"
)

// CleanupTracker records file paths written during a meta-harness
// invocation so they can be removed atomically on failure.
//
// The zero value is ready to use (no constructor required).
type CleanupTracker struct {
	paths []string
}

// Track appends path to the tracker's internal list.
// Duplicate paths are allowed; both will be attempted on Cleanup.
func (t *CleanupTracker) Track(path string) {
	t.paths = append(t.paths, path)
}

// Tracked returns a defensive copy of the tracked path list.
func (t *CleanupTracker) Tracked() []string {
	if len(t.paths) == 0 {
		return []string{}
	}
	result := make([]string, len(t.paths))
	copy(result, t.paths)
	return result
}

// Cleanup removes every tracked path from the file system.
//
// Removal strategy:
//   - os.Remove is attempted first (files and empty directories).
//   - os.IsNotExist errors are silently ignored (idempotent behaviour).
//   - All other errors are collected and returned as a multi-error.
//
// Cleanup is idempotent: re-running it on already-removed paths
// does not return an error.
func (t *CleanupTracker) Cleanup() error {
	var errs []error
	for _, p := range t.paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("harness: Cleanup: remove %q: %w", p, err))
		}
	}
	return errors.Join(errs...)
}

// CleanupOnFailure is a convenience wrapper that calls tracker.Cleanup
// when err is non-nil, then wraps the original error together with any
// cleanup error.
//
// If err is nil, CleanupOnFailure is a no-op and returns nil.
func CleanupOnFailure(tracker *CleanupTracker, err error) error {
	if err == nil {
		return nil
	}
	if cleanupErr := tracker.Cleanup(); cleanupErr != nil {
		return fmt.Errorf("%w (cleanup also failed: %v)", err, cleanupErr)
	}
	return err
}
