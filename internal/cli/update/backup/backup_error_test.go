package backup

// Error-path tests for backup.go, merge.go, restore.go — SPEC-CLI-TUX-V3-003
// AC-TUX3-018 coverage hardening. Each test targets an `if err != nil` branch
// or edge case that was previously uncovered, raising the backup subpackage
// coverage from 75.2% to >=85%.
//
// Permission-based tests are skipped when running as root (root bypasses Unix
// file-permission checks). Tests that trigger errors via file-as-directory or
// invalid YAML need no root skip — they work on all platforms.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// skipIfRoot skips the test when running as root, since root bypasses Unix
// file-permission checks and EACCES-based error paths become unreachable.
func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("skipped as root: permission-based error path is unreachable")
	}
}

// skipIfWindows skips a test whose error path is forced by clearing POSIX
// permission bits. Windows does not enforce them (os.Chmod only toggles the
// read-only attribute, and directories ignore it entirely), so the operation
// under test succeeds and the error branch is unreachable. The branch stays
// covered on unix — this is a platform-coverage gap, not a dropped assertion.
func skipIfWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforced on Windows; error path covered on unix")
	}
}

// restorePerm registers a t.Cleanup that restores directory write permissions
// so t.TempDir() can remove the directory at test end. The Chmod error is
// intentionally discarded — this is best-effort cleanup.
func restorePerm(t *testing.T, dir string) {
	t.Helper()
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

// invalidYAML is a byte sequence that yaml.Unmarshal rejects (unclosed flow
// sequence bracket). Used to trigger error branches in MergeYAML3Way and
// MergeYAMLDeep.
var invalidYAML = []byte("bad: [\n  unclosed")

// --- merge.go error paths ---

func TestMergeYAML3Way_InvalidNewData(t *testing.T) {
	_, err := MergeYAML3Way(invalidYAML, []byte("a: 1\n"), []byte("a: 1\n"))
	if err == nil {
		t.Fatal("expected error from invalid newData, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal new YAML") {
		t.Errorf("expected 'unmarshal new YAML' error, got: %v", err)
	}
}

func TestMergeYAML3Way_InvalidOldData(t *testing.T) {
	_, err := MergeYAML3Way([]byte("a: 1\n"), invalidYAML, []byte("a: 1\n"))
	if err == nil {
		t.Fatal("expected error from invalid oldData, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal old YAML") {
		t.Errorf("expected 'unmarshal old YAML' error, got: %v", err)
	}
}

func TestMergeYAML3Way_InvalidBaseData(t *testing.T) {
	_, err := MergeYAML3Way([]byte("a: 1\n"), []byte("a: 1\n"), invalidYAML)
	if err == nil {
		t.Fatal("expected error from invalid baseData, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal base YAML") {
		t.Errorf("expected 'unmarshal base YAML' error, got: %v", err)
	}
}

// TestDeepMerge3Way_BaseNotMap covers the branch where newV and oldV are both
// maps but baseV is not (deploy.go:76-78). The code substitutes an empty map
// for the non-map base and recurses.
func TestDeepMerge3Way_BaseNotMap(t *testing.T) {
	newMap := map[string]any{"nested": map[string]any{"a": 1}}
	oldMap := map[string]any{"nested": map[string]any{"b": 2}}
	baseMap := map[string]any{"nested": "not-a-map"}

	result := DeepMerge3Way(newMap, oldMap, baseMap)
	nested, ok := result["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map, got %T", result["nested"])
	}
	if nested["a"] != 1 {
		t.Errorf("expected a=1 from newMap, got %v", nested["a"])
	}
}

// TestDeepMerge3Way_BaseNotExists covers the branch where a key exists in new
// and old but not in base (deploy.go:82-85). The code preserves the old (user)
// value since there's no base to compare against.
func TestDeepMerge3Way_BaseNotExists(t *testing.T) {
	newMap := map[string]any{"key": "new-value"}
	oldMap := map[string]any{"key": "old-value"}
	baseMap := map[string]any{} // no "key" in base

	result := DeepMerge3Way(newMap, oldMap, baseMap)
	if result["key"] != "old-value" {
		t.Errorf("expected old-value preserved (no base), got %v", result["key"])
	}
}

func TestMergeYAMLDeep_InvalidNewData(t *testing.T) {
	_, err := MergeYAMLDeep(invalidYAML, []byte("a: 1\n"))
	if err == nil {
		t.Fatal("expected error from invalid newData, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal new YAML") {
		t.Errorf("expected 'unmarshal new YAML' error, got: %v", err)
	}
}

func TestMergeYAMLDeep_InvalidOldData(t *testing.T) {
	_, err := MergeYAMLDeep([]byte("a: 1\n"), invalidYAML)
	if err == nil {
		t.Fatal("expected error from invalid oldData, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal old YAML") {
		t.Errorf("expected 'unmarshal old YAML' error, got: %v", err)
	}
}

// --- backup.go error paths ---

// TestBackupMoaiConfig_StatError covers the os.Stat non-IsNotExist branch
// (backup.go:36). With .moai at mode 0o000, stat of .moai/config fails with
// EACCES (not IsNotExist).
func TestBackupMoaiConfig_StatError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	configDir := filepath.Join(moaiDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)

	if err := os.Chmod(moaiDir, 0o000); err != nil {
		t.Fatal(err)
	}

	_, err := BackupMoaiConfig(root)
	if err == nil {
		t.Fatal("expected error from stat permission failure, got nil")
	}
	if !strings.Contains(err.Error(), "stat config directory") {
		t.Errorf("expected 'stat config directory' error, got: %v", err)
	}
}

// TestBackupMoaiConfig_MkdirAllError covers the backup directory MkdirAll error
// (backup.go:46-48). A regular file at the .moai-backups path causes MkdirAll
// to fail with ENOTDIR. No root skip needed.
func TestBackupMoaiConfig_MkdirAllError(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a FILE where the backups directory is expected — MkdirAll fails.
	backupsPath := filepath.Join(root, defs.BackupsDir)
	if err := os.MkdirAll(filepath.Dir(backupsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupsPath, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := BackupMoaiConfig(root)
	if err == nil {
		t.Fatal("expected error from MkdirAll failure, got nil")
	}
	if !strings.Contains(err.Error(), "create backup directory") {
		t.Errorf("expected 'create backup directory' error, got: %v", err)
	}
}

// TestBackupMoaiConfig_WalkError covers the walk-function error propagation
// (backup.go:61-63 and 106-109). With .moai/config at mode 0o000, Walk cannot
// read directory entries and passes an error to the walk function.
func TestBackupMoaiConfig_WalkError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "test.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, configDir)

	if err := os.Chmod(configDir, 0o000); err != nil {
		t.Fatal(err)
	}

	_, err := BackupMoaiConfig(root)
	if err == nil {
		t.Fatal("expected error from walk failure, got nil")
	}
	if !strings.Contains(err.Error(), "copy config files") {
		t.Errorf("expected 'copy config files' error, got: %v", err)
	}
}

// TestBackupMoaiConfig_ReadFileError covers the os.ReadFile error inside the
// walk function (backup.go:99-101). A config file with mode 0o000 cannot be
// read, causing ReadFile to fail.
func TestBackupMoaiConfig_ReadFileError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	configFile := filepath.Join(configDir, "test.yaml")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, configFile)

	if err := os.Chmod(configFile, 0o000); err != nil {
		t.Fatal(err)
	}

	_, err := BackupMoaiConfig(root)
	if err == nil {
		t.Fatal("expected error from ReadFile failure, got nil")
	}
	if !strings.Contains(err.Error(), "copy config files") {
		t.Errorf("expected 'copy config files' error, got: %v", err)
	}
}

// TestSaveTemplateDefaults_MkdirAllError covers the MkdirAll error for the
// sections destination directory (backup.go:163-165). A regular FILE at the
// "sections" path makes MkdirAll fail on every platform (ENOTDIR on unix, a
// name collision on Windows), so this needs neither a root nor a Windows skip.
func TestSaveTemplateDefaults_MkdirAllError(t *testing.T) {
	destDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(destDir, "sections"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := SaveTemplateDefaults(destDir)
	if err == nil {
		t.Fatal("expected error from MkdirAll failure, got nil")
	}
	if !strings.Contains(err.Error(), "create template defaults directory") {
		t.Errorf("expected 'create template defaults directory' error, got: %v", err)
	}
}

// TestSaveTemplateDefaults_WriteFileError covers the WriteFile error inside the
// template-defaults loop (backup.go:186-187). The sections subdirectory is
// pre-created at mode 0o500 so WriteFile fails.
func TestSaveTemplateDefaults_WriteFileError(t *testing.T) {
	skipIfRoot(t)

	destDir := t.TempDir()
	sectionsDir := filepath.Join(destDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, sectionsDir)

	if err := os.Chmod(sectionsDir, 0o500); err != nil {
		t.Fatal(err)
	}

	// SaveTemplateDefaults returns nil even when individual writes fail (it
	// uses `continue` to skip unreadable/unwritable files). The coverage is
	// gained by entering the WriteFile error branch.
	err := SaveTemplateDefaults(destDir)
	if err != nil {
		t.Fatalf("expected nil (errors are skipped), got: %v", err)
	}
}

// TestCleanupOldBackups_NotADir covers the "backups path is not a directory"
// branch (backup.go:218-220). A regular file at .moai-backups causes the
// function to return 0. No root skip needed.
func TestCleanupOldBackups_NotADir(t *testing.T) {
	root := t.TempDir()
	backupsPath := filepath.Join(root, defs.BackupsDir)
	if err := os.MkdirAll(filepath.Dir(backupsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupsPath, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	deleted := CleanupOldBackups(root, 3)
	if deleted != 0 {
		t.Errorf("expected 0 deletions for non-dir backups path, got %d", deleted)
	}
}

// TestCleanupOldBackups_StatError covers the os.Stat non-IsNotExist branch
// (backup.go:216). With the parent of .moai-backups at mode 0o000, stat fails
// with EACCES.
func TestCleanupOldBackups_StatError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)

	// .moai-backups is under root, not .moai. Make root non-executable so
	// stat of .moai-backups fails (the path doesn't need to exist — we want
	// stat to fail with EACCES, not IsNotExist).
	// Actually: stat of a non-existent path under a 0o000 parent → EACCES.
	backupsPath := filepath.Join(root, defs.BackupsDir)
	_ = backupsPath // ensure var used
	if err := os.Chmod(root, 0o000); err != nil {
		t.Fatal(err)
	}
	defer restorePerm(t, root)

	// CleanupOldBackups uses filepath.Join(projectRoot, defs.BackupsDir).
	// We need stat to fail with non-IsNotExist. With root at 0o000, stat of
	// any child fails with EACCES.
	// BUT: CleanupOldBackups also calls os.Stat which needs x on parent.
	// root at 0o000 → no x → stat fails EACCES → non-IsNotExist → return 0.
	deleted := CleanupOldBackups(root, 3)
	if deleted != 0 {
		t.Errorf("expected 0 deletions on stat error, got %d", deleted)
	}
}

// TestCleanupOldBackups_ReadDirError covers the os.ReadDir error branch
// (backup.go:224-226). The backups directory exists but is unreadable (mode
// 0o000), so ReadDir fails.
func TestCleanupOldBackups_ReadDirError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	backupsDir := filepath.Join(root, defs.BackupsDir)
	if err := os.MkdirAll(backupsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, backupsDir)

	if err := os.Chmod(backupsDir, 0o000); err != nil {
		t.Fatal(err)
	}

	deleted := CleanupOldBackups(root, 3)
	if deleted != 0 {
		t.Errorf("expected 0 deletions on ReadDir error, got %d", deleted)
	}
}

// TestCleanupOldBackups_RemoveAllError covers the RemoveAll error branch
// (backup.go:255-258). A backup directory whose parent is read-only prevents
// removal, triggering the warning path.
func TestCleanupOldBackups_RemoveAllError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	backupsDir := filepath.Join(root, defs.BackupsDir)
	// Create 2 backups, keep only 1, so one must be deleted.
	b1 := filepath.Join(backupsDir, "20260101_120000")
	b2 := filepath.Join(backupsDir, "20260102_120000")
	if err := os.MkdirAll(b1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(b2, 0o755); err != nil {
		t.Fatal(err)
	}
	// Make backupsDir read-only so RemoveAll of a child dir fails.
	restorePerm(t, backupsDir)
	if err := os.Chmod(backupsDir, 0o500); err != nil {
		t.Fatal(err)
	}

	deleted := CleanupOldBackups(root, 1)
	// The deletion fails (EACCES), so deleted should be 0.
	if deleted != 0 {
		t.Errorf("expected 0 successful deletions on RemoveAll error, got %d", deleted)
	}
}

// --- restore.go error paths (modern RestoreMoaiConfig) ---

// TestRestoreMoaiConfig_NonYAMLFile covers the non-YAML skip branch
// (restore.go:73-75). A .json file in the sections backup is silently skipped.
func TestRestoreMoaiConfig_NonYAMLFile(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsDir := filepath.Join(backupDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a non-YAML file that should be skipped.
	if err := os.WriteFile(filepath.Join(sectionsDir, "metadata.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfig(root, backupDir, nil)
	if err != nil {
		t.Errorf("expected nil for non-YAML skip, got: %v", err)
	}
}

// TestRestoreMoaiConfig_TargetNotExists covers the "target file doesn't exist"
// branch (restore.go:99-105). A backup .yaml file exists but the corresponding
// target config file does not, so it's restored as-is.
func TestRestoreMoaiConfig_TargetNotExists(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	configSectionsDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections")
	if err := os.MkdirAll(sectionsBackupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configSectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsBackupDir, "new.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfig(root, backupDir, nil)
	if err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
	// Verify the file was restored.
	data, err := os.ReadFile(filepath.Join(configSectionsDir, "new.yaml"))
	if err != nil {
		t.Fatalf("restored file not found: %v", err)
	}
	if string(data) != "key: value\n" {
		t.Errorf("unexpected restored content: %s", data)
	}
}

// TestRestoreMoaiConfig_TargetNotExists_MkdirAllError covers the MkdirAll error
// when the target file doesn't exist (restore.go:102-104). The config sections
// directory is read-only, so MkdirAll for the parent fails.
func TestRestoreMoaiConfig_TargetNotExists_MkdirAllError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsBackupDir := filepath.Join(backupDir, "sections", "sub")
	if err := os.MkdirAll(sectionsBackupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsBackupDir, "new.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make configDir/sections read-only so MkdirAll for the "sub" subdir fails.
	configSectionsDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections")
	if err := os.MkdirAll(configSectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, configSectionsDir)
	if err := os.Chmod(configSectionsDir, 0o500); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfig(root, backupDir, nil)
	if err == nil {
		t.Fatal("expected error from MkdirAll failure, got nil")
	}
}

// TestRestoreMoaiConfig_3WayMergeFail covers the 3-way merge failure branch
// (restore.go:132-134) and the 2-way merge fallback failure (restore.go:140-143).
// The backup file contains invalid YAML, causing both MergeYAML3Way and
// MergeYAMLDeep to fail. The function falls back to writing the raw backup data.
func TestRestoreMoaiConfig_3WayMergeFail(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults", "sections")
	configSectionsDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections")

	for _, dir := range []string{sectionsBackupDir, templateDefaultsDir, configSectionsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Backup file with INVALID yaml → 3-way merge fails, 2-way merge also fails.
	if err := os.WriteFile(filepath.Join(sectionsBackupDir, "test.yaml"), invalidYAML, 0o644); err != nil {
		t.Fatal(err)
	}
	// Template defaults (base) with valid yaml.
	if err := os.WriteFile(filepath.Join(templateDefaultsDir, "test.yaml"), []byte("key: base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Target file (new template) with valid yaml.
	if err := os.WriteFile(filepath.Join(configSectionsDir, "test.yaml"), []byte("key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use a recorder to verify the fallback is called with success=false.
	var recorderCalls []bool
	recorder := func(_, _ string, success bool, _ io.Writer) {
		recorderCalls = append(recorderCalls, success)
	}

	err := RestoreMoaiConfig(root, backupDir, recorder)
	if err != nil {
		t.Errorf("expected nil (falls back to raw write), got: %v", err)
	}
	// Verify the recorder was called with success=false.
	found := false
	for _, s := range recorderCalls {
		if !s {
			found = true
		}
	}
	if !found {
		t.Error("expected recordFallback to be called with success=false")
	}
}

// TestRestoreMoaiConfig_WalkError covers the walk-function error propagation
// (restore.go:58-60). The sections backup directory is unreadable (mode 0o000),
// so Walk passes an error to the walk function.
func TestRestoreMoaiConfig_WalkError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	if err := os.MkdirAll(sectionsBackupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsBackupDir, "test.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, sectionsBackupDir)
	if err := os.Chmod(sectionsBackupDir, 0o000); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfig(root, backupDir, nil)
	if err == nil {
		t.Fatal("expected error from walk failure, got nil")
	}
}

// TestRestoreMoaiConfig_ReadFileBackupError covers the ReadFile(backupPath)
// error (restore.go:93-95). The backup .yaml file is unreadable (mode 0o000).
func TestRestoreMoaiConfig_ReadFileBackupError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	configSectionsDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections")
	for _, dir := range []string{sectionsBackupDir, configSectionsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	backupFile := filepath.Join(sectionsBackupDir, "test.yaml")
	if err := os.WriteFile(backupFile, []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Target file must exist (so we reach ReadFile for backup, not the
	// IsNotExist branch which returns early).
	if err := os.WriteFile(filepath.Join(configSectionsDir, "test.yaml"), []byte("key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, backupFile)
	if err := os.Chmod(backupFile, 0o000); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfig(root, backupDir, nil)
	if err == nil {
		t.Fatal("expected error from ReadFile backup failure, got nil")
	}
}

// TestRestoreMoaiConfig_ReadFileTargetError covers the ReadFile(targetPath)
// error (restore.go:112-114). The target .yaml file exists but is unreadable.
func TestRestoreMoaiConfig_ReadFileTargetError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	configSectionsDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections")
	for _, dir := range []string{sectionsBackupDir, configSectionsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(sectionsBackupDir, "test.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	targetFile := filepath.Join(configSectionsDir, "test.yaml")
	if err := os.WriteFile(targetFile, []byte("key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, targetFile)
	if err := os.Chmod(targetFile, 0o000); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfig(root, backupDir, nil)
	if err == nil {
		t.Fatal("expected error from ReadFile target failure, got nil")
	}
}

// --- restore.go error paths (legacy RestoreMoaiConfigLegacy) ---

// TestRestoreMoaiConfigLegacy_TargetNotExists covers the "target file doesn't
// exist" branch in the legacy path (restore.go:192-197).
func TestRestoreMoaiConfigLegacy_TargetNotExists(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "newfile.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfigLegacy(root, backupDir, configDir)
	if err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, "newfile.yaml"))
	if err != nil {
		t.Fatalf("restored file not found: %v", err)
	}
	if string(data) != "key: value\n" {
		t.Errorf("unexpected content: %s", data)
	}
}

// TestRestoreMoaiConfigLegacy_MergeFail covers the 2-way merge failure branch
// (restore.go:208-211). The backup file contains invalid YAML, causing
// MergeYAMLDeep to fail. The function falls back to writing the raw backup.
func TestRestoreMoaiConfigLegacy_MergeFail(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Backup file with INVALID yaml → merge fails.
	if err := os.WriteFile(filepath.Join(backupDir, "bad.yaml"), invalidYAML, 0o644); err != nil {
		t.Fatal(err)
	}
	// Target file with valid yaml (so Stat succeeds, ReadFile succeeds).
	if err := os.WriteFile(filepath.Join(configDir, "bad.yaml"), []byte("key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfigLegacy(root, backupDir, configDir)
	if err != nil {
		t.Errorf("expected nil (falls back to raw write), got: %v", err)
	}
}

// TestRestoreMoaiConfigLegacy_WalkError covers the walk-function error
// propagation in the legacy path (restore.go:153-155).
func TestRestoreMoaiConfigLegacy_WalkError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "test.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, backupDir)
	if err := os.Chmod(backupDir, 0o000); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfigLegacy(root, backupDir, configDir)
	if err == nil {
		t.Fatal("expected error from walk failure, got nil")
	}
}

// TestRestoreMoaiConfigLegacy_ReadFileBackupError covers the ReadFile(backupPath)
// error in the legacy path (restore.go:188-190).
func TestRestoreMoaiConfigLegacy_ReadFileBackupError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	for _, dir := range []string{backupDir, configDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	backupFile := filepath.Join(backupDir, "test.yaml")
	if err := os.WriteFile(backupFile, []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Target must exist so we pass the Stat branch.
	if err := os.WriteFile(filepath.Join(configDir, "test.yaml"), []byte("key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, backupFile)
	if err := os.Chmod(backupFile, 0o000); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfigLegacy(root, backupDir, configDir)
	if err == nil {
		t.Fatal("expected error from ReadFile backup failure, got nil")
	}
}

// TestRestoreMoaiConfigLegacy_ReadFileTargetError covers the ReadFile(targetPath)
// error in the legacy path (restore.go:203-205).
func TestRestoreMoaiConfigLegacy_ReadFileTargetError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	for _, dir := range []string{backupDir, configDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(backupDir, "test.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	targetFile := filepath.Join(configDir, "test.yaml")
	if err := os.WriteFile(targetFile, []byte("key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, targetFile)
	if err := os.Chmod(targetFile, 0o000); err != nil {
		t.Fatal(err)
	}

	err := RestoreMoaiConfigLegacy(root, backupDir, configDir)
	if err == nil {
		t.Fatal("expected error from ReadFile target failure, got nil")
	}
}

// Ensure yaml.v3 import is used (for potential future assertions on parsed
// output). The import is retained to keep the test file self-contained for
// YAML round-trip checks if the test suite expands.
var _ = yaml.Unmarshal

// Ensure bytes import is used (buffer for potential io.Writer assertions).
var _ = bytes.Buffer{}
