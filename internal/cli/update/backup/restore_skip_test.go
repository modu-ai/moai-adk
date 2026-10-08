package backup

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRestoreMoaiConfigRetained_SkipsArchivedEntries — gate round 10,
// finding 3 (card t1547): the restore must NOT re-create a section file the
// reconciliation pipeline archived and removed — the summary reports the
// removal, and a restore that resurrects the file contradicts it. The
// skipRel filters exclude such backup entries; everything else restores.
func TestRestoreMoaiConfigRetained_SkipsArchivedEntries(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, ".moai-backups", "test")
	sections := filepath.Join(backupDir, "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	write := func(rel, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(sections, rel), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("kept.yaml", "kept: true\n")
	write("stale.yaml", "stale: true\n")
	// The template (target dir) carries only kept.yaml — stale.yaml would be
	// restored as-is without the skip filter.
	configDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "kept.yaml"), []byte("kept: template\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	if _, err := RestoreMoaiConfigRetained(root, backupDir, nil, func(relPath string) bool {
		return relPath == "stale.yaml"
	}); err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}

	if _, err := os.Stat(filepath.Join(configDir, "stale.yaml")); !os.IsNotExist(err) {
		t.Errorf("archived-removed section was resurrected by the restore (err=%v)", err)
	}
	if data, readErr := os.ReadFile(filepath.Join(configDir, "kept.yaml")); readErr != nil || len(data) == 0 {
		t.Errorf("unfiltered section not restored: %q (err=%v)", data, readErr)
	}
}
