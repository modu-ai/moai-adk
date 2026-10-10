// restore_retention_test.go — U-07 (AC-RC30-08, REQ-RC30-009, REQ-RC30-010):
// the user-facing restore never writes the live .claude/settings.json, names it
// as not restored with the manual copy instruction, lists the backed-up paths
// it does not cover, and keeps the newest five backup runs in both backup roots.
package cli

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

// restoreRunMarker is the file one update run writes at the root of each backup
// directory it creates; its single line is the run id.
const restoreRunMarker = ".moai-run"

// writeRestoreFixtureFile writes content to path, creating its parent directories.
func writeRestoreFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), defs.DirPerm); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// buildRestoreSettingsFixture builds a moai project whose live
// .claude/settings.json differs from the copy in its run backup. It returns the
// project root and the run backup directory.
func buildRestoreSettingsFixture(t *testing.T) (root, backupDir string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	root = t.TempDir()
	writeRestoreFixtureFile(t, filepath.Join(root, ".moai", "config", "sections", "system.yaml"), "version: 1\n")
	writeRestoreFixtureFile(t, filepath.Join(root, ".moai", "config", "sections", "user.yaml"), "user:\n  name: live\n")
	writeRestoreFixtureFile(t, filepath.Join(root, ".claude", "settings.json"), `{"marker":"live"}`)
	writeRestoreFixtureFile(t, filepath.Join(root, ".gitignore"), "node_modules/\n")
	writeRestoreFixtureFile(t, filepath.Join(root, ".moai", "status_line.sh"), "#!/bin/sh\n")

	dir, err := backup.BackupMoaiConfig(root)
	if err != nil || dir == "" {
		t.Fatalf("BackupMoaiConfig: dir=%q err=%v", dir, err)
	}
	if _, _, err := backupInMemoryOnlyFiles(root, dir); err != nil {
		t.Fatalf("backupInMemoryOnlyFiles: %v", err)
	}
	// The backed-up settings.json differs from the live file, so a restore that
	// wrote it would change the live bytes.
	writeRestoreFixtureFile(t, filepath.Join(dir, diskBackupSubdir, ".claude", "settings.json"), `{"marker":"backup"}`)
	return root, dir
}

// runRestoreCapturingOutput runs the user-facing restore and returns its output.
func runRestoreCapturingOutput(t *testing.T, root, backupDir string) string {
	t.Helper()
	var out bytes.Buffer
	if err := runUpdateRestore(root, backupDir, &out); err != nil {
		t.Fatalf("runUpdateRestore: %v", err)
	}
	return out.String()
}

// restoreRunDirs returns the config and namespace backup directories of run i
// in the retention fixture.
func restoreRunDirs(root string, i int) (configDir, nsDir string) {
	configDir = filepath.Join(root, defs.BackupsDir, fmt.Sprintf("20261010_1200%02d", i))
	nsDir = filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir, fmt.Sprintf("update-2026-10-10T12-00-%02dZ", i))
	return configDir, nsDir
}

func TestRestoreSettingsAndRetention(t *testing.T) {
	t.Run("a_live_settings_json_unchanged", func(t *testing.T) {
		root, backupDir := buildRestoreSettingsFixture(t)
		livePath := filepath.Join(root, ".claude", "settings.json")
		before, err := os.ReadFile(livePath)
		if err != nil {
			t.Fatalf("read live settings.json before restore: %v", err)
		}

		out := runRestoreCapturingOutput(t, root, backupDir)
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

	t.Run("b_names_settings_json_not_restored_with_copy_instruction", func(t *testing.T) {
		root, backupDir := buildRestoreSettingsFixture(t)
		out := runRestoreCapturingOutput(t, root, backupDir)

		if !strings.Contains(out, ".claude/settings.json") {
			t.Errorf("restore output does not name .claude/settings.json as not restored:\n%s", out)
		}
		source := filepath.Join(backupDir, diskBackupSubdir, ".claude", "settings.json")
		target := filepath.Join(root, ".claude", "settings.json")
		if !strings.Contains(out, source) {
			t.Errorf("manual copy instruction lacks the backup path %s:\n%s", source, out)
		}
		if !strings.Contains(out, target) {
			t.Errorf("manual copy instruction lacks the target path %s:\n%s", target, out)
		}
	})

	t.Run("c_lists_backed_up_paths_the_restore_does_not_cover", func(t *testing.T) {
		root, backupDir := buildRestoreSettingsFixture(t)
		out := runRestoreCapturingOutput(t, root, backupDir)

		gitignore := filepath.Join(backupDir, diskBackupSubdir, ".gitignore")
		if !strings.Contains(out, ".gitignore") || !strings.Contains(out, gitignore) {
			t.Errorf("restore output does not list the uncovered backed-up path .gitignore (%s):\n%s", gitignore, out)
		}
	})

	t.Run("d_keeps_the_five_newest_runs_in_both_roots", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		root := t.TempDir()
		const runs = 7
		const keep = 5

		// Each run writes one config directory and one namespace directory, and
		// both carry the run's marker. The newest keep runs survive whole; both
		// directories of each older run are removed.
		for i := 0; i < runs; i++ {
			configDir, nsDir := restoreRunDirs(root, i)
			runID := fmt.Sprintf("2026-10-10T12:00:%02d.000000000Z", i)
			writeRestoreFixtureFile(t, filepath.Join(configDir, restoreRunMarker), runID+"\n")
			writeRestoreFixtureFile(t, filepath.Join(nsDir, restoreRunMarker), runID+"\n")
		}

		backup.CleanupOldBackups(root, keep)

		for i := 0; i < runs; i++ {
			configDir, nsDir := restoreRunDirs(root, i)
			_, cfgErr := os.Stat(configDir)
			_, nsErr := os.Stat(nsDir)
			if i < runs-keep {
				if cfgErr == nil {
					t.Errorf("run %d: config directory %s kept; every directory of an older run must be removed", i, configDir)
				}
				if nsErr == nil {
					t.Errorf("run %d: namespace directory %s kept; every directory of an older run must be removed", i, nsDir)
				}
				continue
			}
			if cfgErr != nil {
				t.Errorf("run %d: config directory %s removed; the newest %d runs must be kept: %v", i, configDir, keep, cfgErr)
			}
			if nsErr != nil {
				t.Errorf("run %d: namespace directory %s removed; the newest %d runs must be kept: %v", i, nsDir, keep, nsErr)
			}
		}
	})
}
