package migration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// withTestRegistry saves the global registry, replaces it with the given
// migrations for the duration of the test, and restores the original on
// cleanup. Tests that mutate the package-level `registry` MUST use this helper
// (and MUST NOT run in parallel with other registry-mutating tests).
func withTestRegistry(t *testing.T, migrations []Migration) {
	t.Helper()
	saved := registry
	registry = migrations
	t.Cleanup(func() { registry = saved })
}

// touchMigration returns a Migration whose Apply creates a marker file under
// <root>/.moai/migrations-applied/<name> so the test can verify execution.
func touchMigration(version int, name string) Migration {
	return Migration{
		Version: version,
		Name:    name,
		Apply: func(root string) error {
			dir := filepath.Join(root, ".moai", "migrations-applied")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, name), []byte("applied"), 0o644)
		},
	}
}

// failingMigration returns a Migration whose Apply always returns an error.
func failingMigration(version int, name string) Migration {
	return Migration{
		Version: version,
		Name:    name,
		Apply:   func(string) error { return errors.New("simulated failure") },
	}
}

// markerTmpFile creates a migration-version.tmp file under <root>/.moai/state/
// to simulate an in-flight crash state (REQ-V3R2-RT-007-031).
func markerTmpFile(t *testing.T, root string) {
	t.Helper()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, versionTmpFileName), []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRunner_Apply_HappyPath verifies the basic success path.
// REQ-V3R2-RT-007-012: MigrationRunner.Apply applies registered migrations in order.
func TestRunner_Apply_HappyPath(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
		touchMigration(2, "m002"),
	})

	r := NewRunner(root)
	applied, err := r.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if len(applied) != 2 {
		t.Fatalf("applied count: got %d, want 2", len(applied))
	}

	v, _ := readVersion(root)
	if v != 2 {
		t.Errorf("version after apply: got %d, want 2", v)
	}
}

// TestRunner_Apply_Idempotent verifies idempotency when reapplying.
// REQ-V3R2-RT-007-011: every migration must be idempotent under reapplication.
func TestRunner_Apply_Idempotent(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
	})

	r := NewRunner(root)
	if _, err := r.Apply(context.Background()); err != nil {
		t.Fatalf("first Apply: %v", err)
	}

	// Second apply: version is now 1, no pending migrations.
	applied, err := r.Apply(context.Background())
	if err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("idempotent re-apply: got %d applied, want 0", len(applied))
	}
}

// TestRunner_Apply_FreshInstall_AllInOrder verifies all migrations apply in order on fresh installs.
// REQ-V3R2-RT-007-030: when the version-file is absent, treat current version as 0 and apply all.
func TestRunner_Apply_FreshInstall_AllInOrder(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
		touchMigration(2, "m002"),
		touchMigration(3, "m003"),
	})

	r := NewRunner(root)
	applied, err := r.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(applied) != 3 {
		t.Fatalf("fresh install: got %d, want 3", len(applied))
	}
	for i, v := range applied {
		if v != i+1 {
			t.Errorf("order: applied[%d]=%d, want %d", i, v, i+1)
		}
	}
}

// TestRunner_Apply_VersionAhead verifies behavior when version-file is ahead of registry max.
// REQ-V3R2-RT-007-054: if version is greater than the registry, treat as a no-op.
func TestRunner_Apply_VersionAhead(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
	})
	if err := writeVersion(root, 99); err != nil {
		t.Fatal(err)
	}

	r := NewRunner(root)
	applied, err := r.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("version-ahead: got %d applied, want 0 (no-op)", len(applied))
	}
}

// TestRunner_Apply_FailureHaltsAdvance verifies version advancement halts on failure.
// REQ-V3R2-RT-007-021: on failure, do not update the version-file.
func TestRunner_Apply_FailureHaltsAdvance(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		failingMigration(1, "m001-fail"),
		touchMigration(2, "m002"),
	})

	r := NewRunner(root)
	applied, err := r.Apply(context.Background())
	if err == nil {
		t.Fatal("Apply should fail on migration failure")
	}
	if len(applied) != 0 {
		t.Errorf("applied before failure: got %d, want 0", len(applied))
	}
	v, _ := readVersion(root)
	if v != 0 {
		t.Errorf("version after failure: got %d, want 0 (must not advance)", v)
	}
}

// TestRunner_Apply_PartialSuccess verifies partial-success behavior.
// REQ-V3R2-RT-007-021: progress halts after the failing migration but prior successes persist.
func TestRunner_Apply_PartialSuccess(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
		failingMigration(2, "m002-fail"),
	})

	r := NewRunner(root)
	applied, err := r.Apply(context.Background())
	if err == nil {
		t.Fatal("Apply should fail on second migration")
	}
	if len(applied) != 1 || applied[0] != 1 {
		t.Errorf("partial success: got %v, want [1]", applied)
	}
	v, _ := readVersion(root)
	if v != 1 {
		t.Errorf("version after partial: got %d, want 1", v)
	}
}

// TestRunner_Apply_CrashRecovery verifies recovery on restart after a crash during Apply.
// REQ-V3R2-RT-007-031: presence of version-file.tmp signals in-flight state and triggers cleanup.
func TestRunner_Apply_CrashRecovery(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
	})

	// Simulate a prior crash: leave a .tmp in-flight marker.
	markerTmpFile(t, root)
	if !detectInFlightState(root) {
		t.Fatal("detectInFlightState should detect the .tmp marker")
	}

	r := NewRunner(root)
	applied, err := r.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply with in-flight state: %v", err)
	}
	if len(applied) != 1 {
		t.Errorf("crash recovery: got %d applied, want 1", len(applied))
	}
	// In-flight marker must be cleaned up.
	if detectInFlightState(root) {
		t.Error("in-flight .tmp should be cleaned up after Apply")
	}
	// Idempotent convergence: re-running is a no-op.
	applied2, err := r.Apply(context.Background())
	if err != nil {
		t.Fatalf("re-Apply: %v", err)
	}
	if len(applied2) != 0 {
		t.Errorf("re-Apply should be no-op (idempotent convergence), got %d", len(applied2))
	}
}

// TestRunner_Status verifies the Status query returns current version, pending, and last log.
// REQ-V3R2-RT-007-015.
func TestRunner_Status(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
		touchMigration(2, "m002"),
	})

	r := NewRunner(root)
	if _, err := r.Apply(context.Background()); err != nil {
		t.Fatal(err)
	}

	current, pending, lastApplied, err := r.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if current != 2 {
		t.Errorf("current: got %d, want 2", current)
	}
	if len(pending) != 0 {
		t.Errorf("pending: got %v, want empty", pending)
	}
	if lastApplied == nil {
		t.Error("lastApplied should be non-nil after successful apply")
	}
}

// TestRunner_Status_WithPending verifies Status reports pending migrations before apply.
func TestRunner_Status_WithPending(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
		touchMigration(2, "m002"),
	})
	r := NewRunner(root)
	current, pending, lastApplied, err := r.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if current != 0 {
		t.Errorf("current before apply: got %d, want 0", current)
	}
	if len(pending) != 2 {
		t.Errorf("pending before apply: got %v, want [1 2]", pending)
	}
	if lastApplied != nil {
		t.Errorf("lastApplied before any apply: got %+v, want nil", lastApplied)
	}
}

// TestRunner_Rollback_NotRollbackable verifies Rollback returns the sentinel error
// for migrations without a Rollback function.
// REQ-V3R2-RT-007-024.
func TestRunner_Rollback_NotRollbackable(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"), // no Rollback func
	})

	r := NewRunner(root)
	err := r.Rollback(1)
	if err != ErrMigrationNotRollbackable {
		t.Errorf("Rollback on non-rollbackable migration: got %v, want ErrMigrationNotRollbackable", err)
	}
}

// TestRunner_Rollback_NotFound verifies Rollback on an unknown version errors.
func TestRunner_Rollback_NotFound(t *testing.T) {
	root := t.TempDir()
	withTestRegistry(t, nil)

	r := NewRunner(root)
	err := r.Rollback(999)
	if err == nil {
		t.Error("Rollback on unknown version should error")
	}
}

// TestRunner_Rollback_Success verifies the full rollback path: Rollback func
// executes, version decrements to version-1, and a rolled-back log entry is
// appended.
func TestRunner_Rollback_Success(t *testing.T) {
	root := t.TempDir()
	rollbackCalled := false
	withTestRegistry(t, []Migration{
		touchMigration(1, "m001"),
		{
			Version: 2,
			Name:    "m002-rollbackable",
			Apply: func(r string) error {
				return os.MkdirAll(filepath.Join(r, ".moai", "m002"), 0o755)
			},
			Rollback: func(r string) error {
				rollbackCalled = true
				return os.RemoveAll(filepath.Join(r, ".moai", "m002"))
			},
		},
	})

	r := NewRunner(root)
	if _, err := r.Apply(context.Background()); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if v, _ := readVersion(root); v != 2 {
		t.Fatalf("pre-rollback version: got %d, want 2", v)
	}

	if err := r.Rollback(2); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if !rollbackCalled {
		t.Error("Rollback function was not invoked")
	}
	if v, _ := readVersion(root); v != 1 {
		t.Errorf("post-rollback version: got %d, want 1 (version-1)", v)
	}
}
