package migration_test

import (
	"testing"
)

// TestRunner_Apply_HappyPath verifies the basic success path.
// REQ-V3R2-RT-007-012: MigrationRunner.Apply applies registered migrations in order.
func TestRunner_Apply_HappyPath(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRunner_Apply_Idempotent verifies idempotency when reapplying the same migration.
// REQ-V3R2-RT-007-011: every migration must be idempotent under reapplication.
func TestRunner_Apply_Idempotent(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRunner_Apply_FreshInstall_AllInOrder verifies that all migrations apply in order on fresh installs.
// REQ-V3R2-RT-007-030: when the version-file is absent, treat current version as 0 and apply all migrations.
func TestRunner_Apply_FreshInstall_AllInOrder(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRunner_Apply_VersionAhead verifies behavior when the version-file is ahead of the registry's max.
// REQ-V3R2-RT-007-054: if version is greater than the registry, treat as a no-op.
func TestRunner_Apply_VersionAhead(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRunner_Apply_FailureHaltsAdvance verifies that version advancement halts on migration failure.
// REQ-V3R2-RT-007-021: on failure, do not update the version-file.
func TestRunner_Apply_FailureHaltsAdvance(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRunner_Apply_PartialSuccess verifies partial-success behavior.
// REQ-V3R2-RT-007-021: progress halts after the failing migration but prior successes persist.
func TestRunner_Apply_PartialSuccess(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}

// TestRunner_Apply_CrashRecovery verifies recovery on restart after a crash during Apply.
// REQ-V3R2-RT-007-031: presence of version-file.tmp signals in-flight state and triggers reapplication.
func TestRunner_Apply_CrashRecovery(t *testing.T) {
	// RED: the runner package does not yet exist.
	t.Skip("waiting for migration package implementation")
}
