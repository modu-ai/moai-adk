package backup

// restore_run_marker_test.go — D9 (SPEC-UPDATE-REPAIR-RC30-001, REQ-RC30-010):
// the run marker is bookkeeping. The legacy restore walks every file at the
// backup root, so a marker written there must not land in the config tree.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreLegacySkipsRunMarker(t *testing.T) {
	backupDir := t.TempDir()
	projectRoot := t.TempDir()
	configDir := filepath.Join(projectRoot, ".moai", "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "system.yaml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, ".moai-run"), []byte("2026-10-10T12:00:00.000000000Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RestoreMoaiConfigLegacy(projectRoot, backupDir, configDir); err != nil {
		t.Fatalf("RestoreMoaiConfigLegacy: %v", err)
	}

	if _, err := os.Stat(filepath.Join(configDir, "system.yaml")); err != nil {
		t.Fatalf("control: the backed-up system.yaml was not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(configDir, ".moai-run")); !os.IsNotExist(err) {
		t.Errorf("the run marker was copied into the config tree (stat err: %v)", err)
	}
}
