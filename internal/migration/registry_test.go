package migration_test

import (
	"testing"
)

// TestRegistry_DuplicateVersion_Panics verifies that a panic occurs when a duplicate Version is declared.
// REQ-V3R2-RT-007-053: registering two migrations with the same Version must panic.
func TestRegistry_DuplicateVersion_Panics(t *testing.T) {
	// RED: the registry package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRegistry_Pending verifies the pending-migration list relative to the current version.
// REQ-V3R2-RT-007-012: Pending(current) returns migrations with Version > current.
func TestRegistry_Pending(t *testing.T) {
	// RED: the registry package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRegistry_Highest verifies the registry's maximum version.
// REQ-V3R2-RT-007-016: the registry is compile-time static and exposes a max-version lookup.
func TestRegistry_Highest(t *testing.T) {
	// RED: the registry package does not yet exist.
	t.Skip("waiting for migration package implementation")
}
