//go:build windows

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMigrateAgency_WindowsNoop verifies AC-MIGRATE-011b:
// Windows prints a one-time notice instead of applying Unix permission bits.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-012b
func TestMigrateAgency_WindowsNoop(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stderrBuf bytes.Buffer
	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		stderr:      &stderrBuf,
	}

	if _, err := m.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Stderr must contain the Windows notice
	output := stderrBuf.String()
	const wantMsg = "Windows: Unix permission bits not applicable, ACL preserved as-is"
	if !strings.Contains(output, wantMsg) {
		t.Errorf("expected stderr to contain %q, got:\n%s", wantMsg, output)
	}

	// File content must match source
	src, _ := os.ReadFile(filepath.Join(dir, ".agency", "context", "brand-voice.md"))
	dst, err := os.ReadFile(filepath.Join(dir, ".moai", "project", "brand", "brand-voice.md"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(src, dst) {
		t.Error("brand-voice.md content mismatch after Windows migration")
	}
}
