package deploy

// Error-path tests for deploy.go — SPEC-CLI-TUX-V3-003 AC-TUX3-018 coverage
// hardening. Each test targets an `if err != nil` branch that was previously
// uncovered, raising the deploy subpackage coverage from 74.7% to >=85%.
//
// Permission-based tests are skipped when running as root (root bypasses Unix
// file-permission checks, making EACCES-based error paths unreachable). The
// file-as-directory test (MkdirAll ENOTDIR) needs no root skip — it works on
// all platforms and privilege levels.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// skipIfRoot skips the test when running as root, since root bypasses Unix
// file-permission checks and EACCES-based error paths become unreachable.
func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("skipped as root: permission-based error path is unreachable")
	}
}

// restorePerm registers a t.Cleanup that restores directory write permissions
// so t.TempDir() can remove the directory at test end. The Chmod error is
// intentionally discarded — this is best-effort cleanup.
func restorePerm(t *testing.T, dir string) {
	t.Helper()
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

// --- CleanMoaiManagedPaths error paths ---

// TestCleanMoaiManagedPaths_StatPermissionError covers the os.Stat non-IsNotExist
// branch (deploy.go:101-102). A 0o000 mode on .claude makes stat of any child
// fail with EACCES (not IsNotExist), triggering the Fail+return path.
func TestCleanMoaiManagedPaths_StatPermissionError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	claudeDir := filepath.Join(root, defs.ClaudeDir)
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, claudeDir)

	if err := os.Chmod(claudeDir, 0o000); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out)
	if err == nil {
		t.Fatal("expected error from stat permission failure, got nil")
	}
	if !strings.Contains(err.Error(), "stat ") {
		t.Errorf("expected 'stat' error, got: %v", err)
	}
}

// TestCleanMoaiManagedPaths_RemoveAllError covers the os.RemoveAll error branch
// for non-glob targets (deploy.go:105-108). With .claude at mode 0o500 (r-x),
// stat succeeds (execute bit set) but RemoveAll fails (no write bit to unlink).
func TestCleanMoaiManagedPaths_RemoveAllError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	claudeDir := filepath.Join(root, defs.ClaudeDir)
	settingsPath := filepath.Join(claudeDir, defs.SettingsJSON)
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, claudeDir)

	if err := os.Chmod(claudeDir, 0o500); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out)
	if err == nil {
		t.Fatal("expected error from remove failure, got nil")
	}
	if !strings.Contains(err.Error(), "remove ") {
		t.Errorf("expected 'remove' error, got: %v", err)
	}
}

// TestCleanMoaiManagedPaths_GlobRemoveAllError covers the glob-match RemoveAll
// error branch (deploy.go:83-86). A moai-* directory under .claude/skills/ is
// matched by Glob, but .claude/skills at mode 0o500 prevents removal.
func TestCleanMoaiManagedPaths_GlobRemoveAllError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	skillsDir := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir)
	matchDir := filepath.Join(skillsDir, "moai-test")
	if err := os.MkdirAll(matchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, skillsDir)

	if err := os.Chmod(skillsDir, 0o500); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out)
	if err == nil {
		t.Fatal("expected error from glob match remove failure, got nil")
	}
	if !strings.Contains(err.Error(), "remove ") {
		t.Errorf("expected 'remove' error, got: %v", err)
	}
}

// TestCleanMoaiManagedPaths_ConfigDirRemoveAllError covers the .moai/config
// RemoveAll non-IsNotExist branch (deploy.go:121-125). With .moai at mode
// 0o500, removing the config subdirectory fails with EACCES.
func TestCleanMoaiManagedPaths_ConfigDirRemoveAllError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	configDir := filepath.Join(moaiDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)

	if err := os.Chmod(moaiDir, 0o500); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out)
	if err == nil {
		t.Fatal("expected error from config dir remove failure, got nil")
	}
	if !strings.Contains(err.Error(), "remove ") {
		t.Errorf("expected 'remove' error, got: %v", err)
	}
}

// TestCleanMoaiManagedPaths_MigrateErrorPropagation covers the error-propagation
// branch from MigrateLegacyMemoryDir (deploy.go:132-134). With .moai at 0o500
// and a legacy memory dir present, the Rename inside MigrateLegacyMemoryDir
// fails and the error propagates through CleanMoaiManagedPaths.
func TestCleanMoaiManagedPaths_MigrateErrorPropagation(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	legacyDir := filepath.Join(moaiDir, "memory")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)

	if err := os.Chmod(moaiDir, 0o500); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out)
	if err == nil {
		t.Fatal("expected error from migrate failure propagation, got nil")
	}
	if !strings.Contains(err.Error(), "migrate ") {
		t.Errorf("expected 'migrate' error, got: %v", err)
	}
}

// --- MigrateLegacyMemoryDir error paths ---

// TestMigrateLegacyMemoryDir_RenameError covers the Rename error branch
// (deploy.go:169-172). Legacy dir exists, state dir does not, and .moai at
// mode 0o500 prevents the Rename (no write permission on parent).
func TestMigrateLegacyMemoryDir_RenameError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	legacyDir := filepath.Join(moaiDir, "memory")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)

	if err := os.Chmod(moaiDir, 0o500); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := MigrateLegacyMemoryDir(root, &out)
	if err == nil {
		t.Fatal("expected error from rename failure, got nil")
	}
	if !strings.Contains(err.Error(), "migrate ") {
		t.Errorf("expected 'migrate' error, got: %v", err)
	}
}

// TestMigrateLegacyMemoryDir_RemoveAllLegacyError covers the RemoveAll error
// branch when both legacy and state dirs exist (deploy.go:176-179). With .moai
// at 0o500, removing the legacy directory fails (no write on parent).
func TestMigrateLegacyMemoryDir_RemoveAllLegacyError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	legacyDir := filepath.Join(moaiDir, "memory")
	stateDir := filepath.Join(moaiDir, defs.StateSubdir)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)

	if err := os.Chmod(moaiDir, 0o500); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := MigrateLegacyMemoryDir(root, &out)
	if err == nil {
		t.Fatal("expected error from legacy remove failure, got nil")
	}
	if !strings.Contains(err.Error(), "remove legacy ") {
		t.Errorf("expected 'remove legacy' error, got: %v", err)
	}
}

// --- ScaffoldEvolutionDir error paths ---

// TestScaffoldEvolutionDir_MkdirAllError covers the MkdirAll error branch
// (deploy.go:202-204). A regular file at the .moai/evolution path causes
// MkdirAll for the telemetry subdirectory to fail with ENOTDIR. No root skip
// needed — this error is independent of file permissions.
func TestScaffoldEvolutionDir_MkdirAllError(t *testing.T) {
	root := t.TempDir()
	evolutionPath := filepath.Join(root, ".moai", "evolution")
	if err := os.MkdirAll(filepath.Dir(evolutionPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a FILE where a directory is expected — MkdirAll returns ENOTDIR.
	if err := os.WriteFile(evolutionPath, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := ScaffoldEvolutionDir(root)
	if err == nil {
		t.Fatal("expected error from MkdirAll failure, got nil")
	}
	if !strings.Contains(err.Error(), "create ") {
		t.Errorf("expected 'create' error, got: %v", err)
	}
}

// TestScaffoldEvolutionDir_GitkeepWriteError covers the .gitkeep WriteFile error
// branch (deploy.go:209-211). The telemetry subdirectory exists at mode 0o500,
// so Stat(.gitkeep) returns IsNotExist but WriteFile fails (no write on parent).
func TestScaffoldEvolutionDir_GitkeepWriteError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	evolutionDir := filepath.Join(root, ".moai", "evolution")
	telemetryDir := filepath.Join(evolutionDir, "telemetry")
	if err := os.MkdirAll(telemetryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, telemetryDir)

	if err := os.Chmod(telemetryDir, 0o500); err != nil {
		t.Fatal(err)
	}

	err := ScaffoldEvolutionDir(root)
	if err == nil {
		t.Fatal("expected error from gitkeep write failure, got nil")
	}
	if !strings.Contains(err.Error(), "create ") {
		t.Errorf("expected 'create' error, got: %v", err)
	}
}

// TestScaffoldEvolutionDir_ManifestWriteError covers the manifest.yaml WriteFile
// error branch (deploy.go:228-230). All subdirectories and .gitkeep files
// pre-exist (loop succeeds), but .moai/evolution at mode 0o500 prevents
// creating manifest.yaml.
func TestScaffoldEvolutionDir_ManifestWriteError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	evolutionDir := filepath.Join(root, ".moai", "evolution")
	for _, sub := range []string{"telemetry", "learnings", "new-skills"} {
		dir := filepath.Join(evolutionDir, sub)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		gk := filepath.Join(dir, ".gitkeep")
		if err := os.WriteFile(gk, []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	restorePerm(t, evolutionDir)

	if err := os.Chmod(evolutionDir, 0o500); err != nil {
		t.Fatal(err)
	}

	err := ScaffoldEvolutionDir(root)
	if err == nil {
		t.Fatal("expected error from manifest write failure, got nil")
	}
	if !strings.Contains(err.Error(), "create ") {
		t.Errorf("expected 'create' error, got: %v", err)
	}
}

// TestScaffoldEvolutionDir_ChangelogWriteError covers the changelog.md WriteFile
// error branch (deploy.go:244-246). All subdirectories, .gitkeep files, and
// manifest.yaml pre-exist, but .moai/evolution at mode 0o500 prevents creating
// changelog.md.
func TestScaffoldEvolutionDir_ChangelogWriteError(t *testing.T) {
	skipIfRoot(t)

	root := t.TempDir()
	evolutionDir := filepath.Join(root, ".moai", "evolution")
	for _, sub := range []string{"telemetry", "learnings", "new-skills"} {
		dir := filepath.Join(evolutionDir, sub)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		gk := filepath.Join(dir, ".gitkeep")
		if err := os.WriteFile(gk, []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, evolutionDir)

	if err := os.Chmod(evolutionDir, 0o500); err != nil {
		t.Fatal(err)
	}

	err := ScaffoldEvolutionDir(root)
	if err == nil {
		t.Fatal("expected error from changelog write failure, got nil")
	}
	if !strings.Contains(err.Error(), "create ") {
		t.Errorf("expected 'create' error, got: %v", err)
	}
}
