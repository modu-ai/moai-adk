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

	// Set a non-default permission on the source file
	brandVoiceSrc := filepath.Join(dir, ".agency", "context", "brand-voice.md")
	if err := os.Chmod(brandVoiceSrc, 0o640); err != nil {
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

	dst := filepath.Join(dir, ".moai", "project", "brand", "brand-voice.md")
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
