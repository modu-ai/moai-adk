package cli

// update_restore_settings_test.go: SPEC-UPDATE-REPAIR-RC30-001 defect D9 (card t1619).
//
// The user-facing restore (`moai update --restore <dir>`) must:
//   - never write the live .claude/settings.json (REQ-RC30-009, AC-RC30-08);
//   - name .claude/settings.json as not restored, with a manual copy instruction
//     that shows the backup path and the target path (REQ-RC30-009);
//   - list the backed-up paths the restore does not cover (REQ-RC30-009);
//   - count backup retention per run, not by directory-name shape (REQ-RC30-010).

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// restoreRunMarkerD9 is the per-run marker the retention subtest writes into both
// backup roots of a run. Its single line is the run id.
const restoreRunMarkerD9 = ".moai-run"

func restoreSettingsWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), defs.DirPerm); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// restoreSettingsFixture builds a moai project whose live .claude/settings.json
// differs from the copy held in the run backup. It returns the project root and
// the run backup directory.
func restoreSettingsFixture(t *testing.T) (root, backupDir string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	root = t.TempDir()
	restoreSettingsWriteFile(t, filepath.Join(root, ".moai", "config", "sections", "system.yaml"), "version: 1\n")
	restoreSettingsWriteFile(t, filepath.Join(root, ".moai", "config", "sections", "user.yaml"), "user:\n  name: live\n")
	restoreSettingsWriteFile(t, filepath.Join(root, ".claude", "settings.json"), `{"marker":"live"}`)
	restoreSettingsWriteFile(t, filepath.Join(root, ".gitignore"), "node_modules/\n")
	restoreSettingsWriteFile(t, filepath.Join(root, ".moai", "status_line.sh"), "#!/bin/sh\n")

	dir, err := backup.BackupMoaiConfig(root)
	if err != nil || dir == "" {
		t.Fatalf("BackupMoaiConfig: dir=%q err=%v", dir, err)
	}
	if _, _, err := backupInMemoryOnlyFiles(root, dir); err != nil {
		t.Fatalf("backupInMemoryOnlyFiles: %v", err)
	}
	// The backed-up settings.json must differ from the live file, so that a
	// restore which wrote it would show up in the byte comparison.
	restoreSettingsWriteFile(t, filepath.Join(dir, "in-memory-backups", ".claude", "settings.json"), `{"marker":"backup"}`)
	return root, dir
}

func runRestoreForTest(t *testing.T, root, backupDir string) string {
	t.Helper()
	var out bytes.Buffer
	if err := runUpdateRestore(root, backupDir, &out); err != nil {
		t.Fatalf("runUpdateRestore: %v", err)
	}
	return out.String()
}

// retentionRunDirs returns the config backup directory and the namespace backup
// directory that run i of the retention fixture writes.
func retentionRunDirs(root string, i int) (configDir, nsDir string) {
	configDir = filepath.Join(root, defs.BackupsDir, fmt.Sprintf("20261010_1200%02d", i))
	nsDir = filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir, fmt.Sprintf("update-2026-10-10T12-00-%02dZ", i))
	return configDir, nsDir
}

func TestRestoreSettingsAndRetentionD9(t *testing.T) {
	t.Run("a_live_settings_json_unchanged", func(t *testing.T) {
		root, backupDir := restoreSettingsFixture(t)
		livePath := filepath.Join(root, ".claude", "settings.json")
		before, err := os.ReadFile(livePath)
		if err != nil {
			t.Fatalf("read live settings.json before restore: %v", err)
		}

		out := runRestoreForTest(t, root, backupDir)
		if !strings.Contains(out, "Restored .moai/config from") {
			t.Fatalf("control failed: restore did not report completion:\n%s", out)
		}

		after, err := os.ReadFile(livePath)
		if err != nil {
			t.Fatalf("read live settings.json after restore: %v", err)
		}
		if !bytes.Equal(before, after) {
			t.Errorf("restore rewrote the live .claude/settings.json:\nbefore: %s\nafter:  %s", before, after)
		}
	})

	t.Run("b_names_not_restored_settings_json_with_copy_hint", func(t *testing.T) {
		root, backupDir := restoreSettingsFixture(t)
		out := runRestoreForTest(t, root, backupDir)

		if !strings.Contains(out, ".claude/settings.json") {
			t.Errorf("restore output does not name .claude/settings.json:\n%s", out)
		}
		source := filepath.Join(backupDir, "in-memory-backups", ".claude", "settings.json")
		target := filepath.Join(root, ".claude", "settings.json")
		if !strings.Contains(out, source) {
			t.Errorf("manual copy instruction lacks the backup path %s:\n%s", source, out)
		}
		if !strings.Contains(out, target) {
			t.Errorf("manual copy instruction lacks the target path %s:\n%s", target, out)
		}
	})

	t.Run("c_lists_backed_up_uncovered_paths", func(t *testing.T) {
		root, backupDir := restoreSettingsFixture(t)
		out := runRestoreForTest(t, root, backupDir)

		gitignore := filepath.Join(backupDir, "in-memory-backups", ".gitignore")
		if !strings.Contains(out, ".gitignore") || !strings.Contains(out, gitignore) {
			t.Errorf("restore output does not list the backed-up uncovered path .gitignore (%s):\n%s", gitignore, out)
		}
	})

	t.Run("d_keeps_five_newest_runs", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		root := t.TempDir()
		const runs = 7
		const keep = 5

		// Seven runs. Each run writes a config backup and a namespace backup, and
		// both carry the same run marker. The newest `keep` runs must survive
		// intact; both directories of each older run must go.
		for i := 0; i < runs; i++ {
			configDir, nsDir := retentionRunDirs(root, i)
			runID := fmt.Sprintf("2026-10-10T12:00:%02d.000000000Z", i)
			restoreSettingsWriteFile(t, filepath.Join(configDir, restoreRunMarkerD9), runID+"\n")
			restoreSettingsWriteFile(t, filepath.Join(nsDir, restoreRunMarkerD9), runID+"\n")
		}

		backup.CleanupOldBackups(root, keep)

		for i := 0; i < runs; i++ {
			configDir, nsDir := retentionRunDirs(root, i)
			_, cfgErr := os.Stat(configDir)
			_, nsErr := os.Stat(nsDir)
			if i < runs-keep {
				if cfgErr == nil {
					t.Errorf("run %d: config backup %s kept; the oldest %d runs must be removed", i, configDir, runs-keep)
				}
				if nsErr == nil {
					t.Errorf("run %d: namespace backup %s kept; it must be removed with its run", i, nsDir)
				}
				continue
			}
			if cfgErr != nil {
				t.Errorf("run %d: config backup %s removed; the newest %d runs must be kept: %v", i, configDir, keep, cfgErr)
			}
			if nsErr != nil {
				t.Errorf("run %d: namespace backup %s removed; the newest %d runs must be kept: %v", i, nsDir, keep, nsErr)
			}
		}
	})
}
