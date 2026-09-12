package backup

// settings_snapshot_restore_test.go is the run-phase M1 verification item of
// SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 (plan.md Decision D5): the user-facing
// restore command (`moai update --restore` → RestoreFromBackupDir →
// RestoreMoaiConfig) never writes the live .claude/settings.json. Case 5 of the
// promotion rule ("the user file is reverted after an abort") can therefore
// only arise from a hand edit, not from the restore command.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

func writeFileRel(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), defs.DirPerm); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func TestRestoreFromBackupDir_NeverWritesLiveSettingsJSON(t *testing.T) {
	const live = `{"marker":"live-render"}`
	const backedUpUser = `{"marker":"user-file-in-backup"}`

	cases := []struct {
		name       string
		seedBackup func(t *testing.T, backupDir string)
		// written is a path under .moai/config the restore must have produced,
		// proving the restore actually ran (a no-op restore would pass the
		// settings.json assertion vacuously).
		written string
	}{
		{
			name: "sections_backup",
			seedBackup: func(t *testing.T, backupDir string) {
				writeFileRel(t, backupDir, "sections/user.yaml", "user:\n  name: restored\n")
				writeFileRel(t, backupDir, "in-memory-backups/.claude/settings.json", backedUpUser)
			},
			written: ".moai/config/sections/user.yaml",
		},
		{
			name: "legacy_backup_without_sections",
			seedBackup: func(t *testing.T, backupDir string) {
				writeFileRel(t, backupDir, "in-memory-backups/.claude/settings.json", backedUpUser)
				writeFileRel(t, backupDir, ".claude/settings.json", backedUpUser)
			},
			written: ".moai/config/in-memory-backups/.claude/settings.json",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFileRel(t, root, ".claude/settings.json", live)
			if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), defs.DirPerm); err != nil {
				t.Fatalf("mkdir config: %v", err)
			}
			backupDir := t.TempDir()
			writeFileRel(t, backupDir, BackupMarkerFile, "{}")
			tc.seedBackup(t, backupDir)

			if err := RestoreFromBackupDir(root, backupDir, nil); err != nil {
				t.Fatalf("RestoreFromBackupDir: %v", err)
			}

			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(tc.written))); err != nil {
				t.Fatalf("restore did not write %s (control): %v", tc.written, err)
			}
			got, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
			if err != nil {
				t.Fatalf("read live settings.json: %v", err)
			}
			if !bytes.Equal(got, []byte(live)) {
				t.Errorf("restore rewrote the live .claude/settings.json:\ngot:  %s\nwant: %s", got, live)
			}
		})
	}
}
