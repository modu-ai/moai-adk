package deploy

// Inventory + remaining branch tests for the clean path
// (SPEC-CLI-CLEAN-SYMLINK-001 M4 sweep). InventoryManagedPaths is the
// read-only deletion preview sharing ManagedCleanTargets with the removal
// itself — its file-listing contract (forward-slash paths relative to the
// project root, advisory skips) is behavior in its own right, exercised
// here including its symlink handling (links are listed, never traversed).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestInventoryManagedPathsListsFilesAndLinks verifies the read-only
// inventory surface: files under managed roots are listed as
// project-root-relative forward-slash paths, the glob root contributes its
// matches, and a symlink entry is listed without any traversal of its
// target.
func TestInventoryManagedPathsListsFilesAndLinks(t *testing.T) {
	root := t.TempDir()
	rulesFile := filepath.Join(root, defs.ClaudeDir, defs.RulesMoaiSubdir, "core", "rule.md")
	if err := os.MkdirAll(filepath.Dir(rulesFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rulesFile, []byte("rule"), 0o644); err != nil {
		t.Fatal(err)
	}
	extraSkill := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "moai-extra", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(extraSkill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extraSkill, []byte("skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkRel := filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir)
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, linkRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, filepath.Join(root, "inv-target"), filepath.Join(root, linkRel))

	got := InventoryManagedPaths(root)
	want := map[string]bool{
		".claude/rules/moai/core/rule.md":    false,
		".claude/skills/moai-extra/SKILL.md": false,
		".claude/agents/moai":                false, // the link entry itself
	}
	for _, p := range got {
		if w, ok := want[p]; ok && !w {
			want[p] = true
		}
	}
	for p, seen := range want {
		if !seen {
			t.Errorf("inventory missing %s (got %v)", p, got)
		}
	}
	// The inventory must not have followed the link into its target's tree:
	// nothing outside the managed roots may appear.
	for _, p := range got {
		if strings.HasPrefix(p, "inv-target") {
			t.Errorf("inventory followed a symlink out of the managed roots: %s", p)
		}
	}
}

// TestBackupThenRemove_RealFileTemplateCarried covers the real-file branch
// where the template carries the exact relative path: removal without
// backup (deployment rewrites the file, so its only copy is never at
// stake) — the non-link sibling of the carried-file-link unit test.
func TestBackupThenRemove_RealFileTemplateCarried(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "carried-real.txt")
	if err := os.WriteFile(p, []byte("carried"), 0o644); err != nil {
		t.Fatal(err)
	}

	backedUp, note, err := backupThenRemove(p, "carried-real.txt",
		filepath.Join(root, "bk"), preCleanTestFS("carried-real.txt"))
	if err != nil {
		t.Fatalf("backupThenRemove: %v", err)
	}
	if backedUp != 0 || note != "" {
		t.Errorf("backedUp=%d note=%q, want 0/empty (template carries the path)", backedUp, note)
	}
	if _, serr := os.Lstat(p); !os.IsNotExist(serr) {
		t.Errorf("carried real file not removed: %v", serr)
	}
}

// TestCleanMoaiManagedPaths_ConfigRootLstatError covers the config root's
// internal stat-error branch (the eighth root has no pre-check, so a
// non-IsNotExist Lstat failure surfaces from backupThenRemove itself).
func TestCleanMoaiManagedPaths_ConfigRootLstatError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	root := t.TempDir()
	moaiDir := filepath.Join(root, defs.MoAIDir)
	if err := os.MkdirAll(filepath.Join(moaiDir, defs.ConfigSubdir), 0o755); err != nil {
		t.Fatal(err)
	}
	restorePerm(t, moaiDir)
	if err := os.Chmod(moaiDir, 0o000); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out, preCleanTestFS())
	if err == nil {
		t.Fatal("expected an error from the config root's stat failure, got nil")
	}
	if !strings.Contains(err.Error(), "stat "+filepath.Join(defs.MoAIDir, defs.ConfigSubdir)) {
		t.Errorf("error does not name the config root stat: %v", err)
	}
}

// TestCleanMoaiManagedPaths_LiveFileSymlinkBackupError covers the file-link
// disposition's backup-failure branch: the abort ordering applies to links
// too — a failed backup of the target bytes returns before the link removal.
func TestCleanMoaiManagedPaths_LiveFileSymlinkBackupError(t *testing.T) {
	skipIfRoot(t)
	skipIfWindows(t)

	outsideRoot := t.TempDir()
	outsideFile := filepath.Join(outsideRoot, "unreadable.json")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0o000); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.ClaudeDir), 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON)
	makeSymlink(t, outsideFile, settingsPath)

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/settings.json.tmpl"))
	if err == nil {
		t.Fatal("expected a backup error from the unreadable link target, got nil")
	}
	if !strings.Contains(err.Error(), "back up") {
		t.Errorf("error does not attribute the backup failure: %v", err)
	}
	// Abort ordering: the link survives the failed backup (a failed backup
	// must never be followed by the removal it was taken to survive).
	if fi, lerr := os.Lstat(settingsPath); lerr != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("file symlink removed despite backup failure: %v", lerr)
	}
}
