package deploy

// Pre-clean backup tests for CleanMoaiManagedPaths (card t111). The contract
// under test is the REQ-UDS-008 generalization: every regular file the
// template filesystem does not carry at the same relative path reaches the
// run's .moai-backups/<timestamp>/pre-clean/ tree BEFORE the managed root is
// removed, and a backup failure aborts before any removal happens.
//
// The template side is a fstest.MapFS rather than the real embed so the
// managed/unmanaged split is an input the test controls, not a fact about
// today's template tree.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// preCleanTestFS builds a template FS carrying exactly the given paths.
func preCleanTestFS(paths ...string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for _, p := range paths {
		fsys[p] = &fstest.MapFile{Data: []byte("template:" + p)}
	}
	return fsys
}

// preCleanBackupRoot finds the single pre-clean backup directory the run
// created under root, failing the test when the count is not exactly one.
func preCleanBackupRoot(t *testing.T, root string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, defs.BackupsDir, "*", preCleanBackupSubdir))
	if err != nil {
		t.Fatalf("glob backup root: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly 1 pre-clean backup dir, got %d: %v", len(matches), matches)
	}
	return matches[0]
}

// TestCleanMoaiManagedPaths_BackupsUnmanagedFiles is the acceptance case: a
// file under a managed root that the template does not carry survives into
// the backup while the root itself is removed, and a file the template DOES
// carry is not copied — deployment rewrites it, so backing it up would only
// duplicate what the update overwrites.
func TestCleanMoaiManagedPaths_BackupsUnmanagedFiles(t *testing.T) {
	root := t.TempDir()
	rulesDir := filepath.Join(root, defs.ClaudeDir, defs.RulesMoaiSubdir)
	if err := os.MkdirAll(filepath.Join(rulesDir, "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	unmanaged := filepath.Join(rulesDir, "core", "local-only.md")
	if err := os.WriteFile(unmanaged, []byte("user-authored rule"), 0o644); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(rulesDir, "core", "core-keep.md")
	if err := os.WriteFile(managed, []byte("stale on disk"), 0o644); err != nil {
		t.Fatal(err)
	}

	tmplFS := preCleanTestFS(".claude/rules/moai/core/core-keep.md")

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, tmplFS); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	// The root is gone — the cleanup behavior is unchanged.
	if _, err := os.Stat(rulesDir); !os.IsNotExist(err) {
		t.Errorf("managed root still exists after clean: %v", err)
	}

	// The unmanaged file reached the backup, content intact.
	backup := filepath.Join(preCleanBackupRoot(t, root),
		defs.ClaudeDir, defs.RulesMoaiSubdir, "core", "local-only.md")
	data, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("unmanaged file not backed up: %v", err)
	}
	if string(data) != "user-authored rule" {
		t.Errorf("backup content = %q, want the original bytes", data)
	}

	// The template-managed file was not copied — the backup holds only what
	// would otherwise be lost.
	managedBackup := filepath.Join(preCleanBackupRoot(t, root),
		defs.ClaudeDir, defs.RulesMoaiSubdir, "core", "core-keep.md")
	if _, err := os.Stat(managedBackup); !os.IsNotExist(err) {
		t.Errorf("template-managed file was backed up (deployment rewrites it): %v", err)
	}

	if !strings.Contains(out.String(), "backed up 1 unmanaged file") {
		t.Errorf("progress output does not report the backup:\n%s", out.String())
	}
}

// TestCleanMoaiManagedPaths_ConfigTreeReachesBackup covers the .moai/config
// root — the 2026-08-15 incident's heaviest loss (a user's astgrep-rules tree
// vanished with no backup anywhere).
func TestCleanMoaiManagedPaths_ConfigTreeReachesBackup(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(filepath.Join(cfgDir, "astgrep-rules", "go"), 0o755); err != nil {
		t.Fatal(err)
	}
	ruleFile := filepath.Join(cfgDir, "astgrep-rules", "go", "hardcoding.yml")
	if err := os.WriteFile(ruleFile, []byte("rules: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".moai/config/sections/language.yaml")); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	if _, err := os.Stat(cfgDir); !os.IsNotExist(err) {
		t.Errorf("config dir still exists after clean: %v", err)
	}
	backup := filepath.Join(preCleanBackupRoot(t, root),
		defs.MoAIDir, defs.ConfigSubdir, "astgrep-rules", "go", "hardcoding.yml")
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("astgrep rule not backed up: %v", err)
	}
}

// TestCleanMoaiManagedPaths_SettingsFileBackedUp covers the FILE target
// shape: settings.json is runtime-generated (the template carries only the
// .tmpl), so a customized settings.json counts as unmanaged and is backed up
// before removal.
func TestCleanMoaiManagedPaths_SettingsFileBackedUp(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.ClaudeDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON),
		[]byte("{\"custom\":true}"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	// The template carries settings.json.tmpl, never settings.json itself.
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/settings.json.tmpl")); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	backup := filepath.Join(preCleanBackupRoot(t, root), defs.ClaudeDir, defs.SettingsJSON)
	data, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("settings.json not backed up: %v", err)
	}
	if !strings.Contains(string(data), "custom") {
		t.Errorf("backup content = %q, want the user's file", data)
	}
}

// TestCleanMoaiManagedPaths_GlobMatchBackedUp covers the glob target shape:
// a moai-prefixed skills directory the template does not carry.
func TestCleanMoaiManagedPaths_GlobMatchBackedUp(t *testing.T) {
	root := t.TempDir()
	extra := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "moai-extra")
	if err := os.MkdirAll(extra, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extra, "SKILL.md"), []byte("user skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, preCleanTestFS()); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	backup := filepath.Join(preCleanBackupRoot(t, root),
		defs.ClaudeDir, defs.SkillsSubdir, "moai-extra", "SKILL.md")
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("glob-matched file not backed up: %v", err)
	}
}

// TestCleanMoaiManagedPaths_BackupFailureAbortsRemoval is the abort-ordering
// contract: when the backup copy fails, the removal must not run — a failed
// backup must never be followed by the destruction it was taken to survive.
func TestCleanMoaiManagedPaths_BackupFailureAbortsRemoval(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	rulesDir := filepath.Join(root, defs.ClaudeDir, defs.RulesMoaiSubdir)
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	locked := filepath.Join(rulesDir, "unreadable.md")
	if err := os.WriteFile(locked, []byte("cannot read me"), 0o000); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out, preCleanTestFS())
	if err == nil {
		t.Fatal("expected error from backup failure, got nil")
	}
	if !strings.Contains(err.Error(), "back up") {
		t.Errorf("expected a 'back up' error, got: %v", err)
	}

	// The root and the file are still on disk — the abort happened before
	// any removal.
	if _, statErr := os.Stat(locked); statErr != nil {
		t.Errorf("unmanaged file was removed despite backup failure: %v", statErr)
	}
	if _, statErr := os.Stat(rulesDir); statErr != nil {
		t.Errorf("managed root was removed despite backup failure: %v", statErr)
	}
}
