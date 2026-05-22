package cli_test

import (
	"testing"
)

// TestMigrationStatus_Human verifies the human-readable status output.
// REQ-V3R2-RT-007-015: moai migration status prints the current version, pending, and last applied.
func TestMigrationStatus_Human(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationStatus_JSON verifies JSON-format output.
// REQ-V3R2-RT-007-041: the --json flag emits machine-readable JSON.
func TestMigrationStatus_JSON(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRun_AppliesPending verifies that pending migrations are applied.
// REQ-V3R2-RT-007-040: moai migration run applies pending migrations.
func TestMigrationRun_AppliesPending(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRun_NoPending verifies behavior when there are no pending migrations.
func TestMigrationRun_NoPending(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRollback_NoRollbackable verifies the case where rollback is not possible.
// REQ-V3R2-RT-007-024: if Rollback is nil, MigrationNotRollbackable is returned.
func TestMigrationRollback_NoRollbackable(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRollback_Succeeds verifies the rollback success path.
// REQ-V3R2-RT-007-042: migrations that declare Rollback support rollback.
func TestMigrationRollback_Succeeds(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRollback_M001_Rejected verifies that m001 rollback is rejected.
// REQ-V3R2-RT-007-024: m001 is a CRITICAL bug-fix and therefore not rollbackable.
func TestMigrationRollback_M001_Rejected(t *testing.T) {
	// RED: the migration CLI does not yet exist
	t.Skip("waiting for migration CLI implementation")
}
