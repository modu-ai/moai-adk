//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMigrateAgency_POSIXPermission verifies AC-MIGRATE-011a:
// POSIX permission bits are preserved after migration.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-012a
func TestMigrateAgency_POSIXPermission(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Set a non-default permission on a source file that is still migrated by copyFile.
	// (Phase 2 brand migration was removed; learnings → observations exercises the
	// same permission-preserving copyFile path.)
	learnSrc := filepath.Join(dir, ".agency", "learnings", "LEARN-001.md")
	if err := os.Chmod(learnSrc, 0o640); err != nil {
		t.Fatalf("Chmod source: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	if _, err := m.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	dst := filepath.Join(dir, ".moai", "research", "observations", "LEARN-001.md")
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Stat dst: %v", err)
	}

	// Mask against 0o7777 and compare
	got := info.Mode().Perm()
	want := os.FileMode(0o640)
	if got != want {
		t.Errorf("expected permission %o, got %o", want, got)
	}
}
