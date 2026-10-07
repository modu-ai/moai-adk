package cli

// update_migrate_test.go — the M3 migration criteria (SPEC-INIT-SHRINK-001
// acceptance.md AC-011..AC-019) plus the leader condition 5b test (D-16:
// the one-run preservation boundary). Every test drives the REAL update
// flow (runTemplateSyncWithReporter through runUpdateCobraCmd) against a
// fixture old-project tree; the probe arms are reached through the injected
// fake runner or the default runner's test-binary refusal (REQ-020: no
// real tool, no real profile).

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// migration fixture paths (the AC-010 fixture shapes, reused end to end).
const (
	migIdenticalSkill = ".claude/skills/moai-foundation-core/SKILL.md"
	migModifiedSkill  = ".claude/skills/moai-workflow-tdd/SKILL.md"
	migForeignSkill   = ".claude/skills/moai-custom/SKILL.md"
	migForeignCommand = ".claude/commands/moai-user-cmd.md"
	migAbsentRecCmd   = ".claude/commands/moai/todo.md"
)

// buildMigrationFixture seeds a record-less old-project tree with one file
// per class and a config section file for the Backup/Restore cycle.
func buildMigrationFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	write := func(rel, content string) {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	identical, err := fs.ReadFile(embedded, migIdenticalSkill)
	if err != nil {
		t.Fatalf("embedded read: %v", err)
	}
	modified, err := fs.ReadFile(embedded, migModifiedSkill)
	if err != nil {
		t.Fatalf("embedded read: %v", err)
	}
	todo, err := fs.ReadFile(embedded, migAbsentRecCmd)
	if err != nil {
		t.Fatalf("embedded read: %v", err)
	}
	write(migIdenticalSkill, string(identical))
	write(migModifiedSkill, string(modified)+"\n<!-- user edit -->\n")
	write(migForeignSkill, "# moai-custom: the user's own skill\n")
	write(migForeignCommand, "# the user's own command\n")
	write(migAbsentRecCmd, string(todo))
	write(".moai/config/sections/llm.yaml", "llm:\n  harness: claude\n")
	write(".claude/settings.json", "{}\n")

	// The manifest tracks the identical copy healthy and the modified copy
	// stale (the edit lands after tracking); the absent-record command and
	// both foreign files carry no record.
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(migIdenticalSkill, manifest.TemplateManaged, manifest.HashBytes(identical)); err != nil {
		t.Fatalf("track: %v", err)
	}
	if err := mgr.Track(migModifiedSkill, manifest.TemplateManaged, manifest.HashBytes(modified)); err != nil {
		t.Fatalf("track: %v", err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	return root
}

// runUpdateCobraCmd drives the real template-sync flow in the scratch
// project (the flow uses "." as its project root) with the given flags.
func runUpdateCobraCmd(t *testing.T, root string, flags map[string]string) (string, string) {
	t.Helper()
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v (stderr: %s)", err, errBuf.String())
	}
	return out.String(), errBuf.String()
}

// readFixtureFile returns a file's bytes, failing the test when absent.
func readFixtureFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

// assertFileAbsent asserts a dropped-root file is gone from the project.
func assertFileAbsent(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Errorf("%s still present (stat err: %v)", rel, err)
	}
}

// assertFilePresent asserts a file exists.
func assertFilePresent(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Errorf("%s missing: %v", rel, err)
	}
}

// TestUpdateLocalModeKeepsFullScope is AC-017: a local-mode project's
// update keeps today's full merge scope — template-carried skills deploy
// and the record survives the cycle.
func TestUpdateLocalModeKeepsFullScope(t *testing.T) {
	root := buildMigrationFixture(t)
	if err := template.ApplyDeployMode(root, "local"); err != nil {
		t.Fatalf("seed record: %v", err)
	}
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "no-plugin": "true"})

	if got := config.ReadDeployMode(root); got != "local" {
		t.Errorf("deployment_mode = %q, want local", got)
	}
	assertFilePresent(t, root, ".claude/skills")
	assertFilePresent(t, root, ".claude/commands")
}

// TestUpdateNeverFlipsModeRecord is AC-018: update (with or without
// --force) leaves the record unchanged and names the init re-entry as the
// switch surface.
func TestUpdateNeverFlipsModeRecord(t *testing.T) {
	for _, force := range []string{"true", "false"} {
		root := buildMigrationFixture(t)
		if err := template.ApplyDeployMode(root, "local"); err != nil {
			t.Fatalf("seed record: %v", err)
		}
		flags := map[string]string{"yes": "true", "no-plugin": "true", "force": force}
		out, _ := runUpdateCobraCmd(t, root, flags)

		if got := config.ReadDeployMode(root); got != "local" {
			t.Errorf("force=%s: deployment_mode flipped to %q", force, got)
		}
		if !strings.Contains(out, "deploy mode: local") || !strings.Contains(out, "moai init") {
			t.Errorf("force=%s: switch guidance missing from output:\n%s", force, out)
		}
	}
}

// TestNotDemonstratedPreservationEndsAtNextLocalUpdate pins the D-16 loss
// path (leader condition 5b): the not-demonstrated migration preserves the
// foreign file for THAT run; the next update — the record now local, the
// full deployer and today's Clean walk — removes it (backed up). The
// boundary is executable, not prose: preservation is one migration run
// (acceptance.md Edge Cases, the narrowed promise).
func TestNotDemonstratedPreservationEndsAtNextLocalUpdate(t *testing.T) {
	root := buildMigrationFixture(t)

	// Run 1 — the not-demonstrated migration: the foreign file survives.
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	if got := config.ReadDeployMode(root); got != "local" {
		t.Fatalf("run 1 record = %q, want local", got)
	}
	foreignBefore := readFixtureFile(t, root, migForeignSkill)
	if !strings.Contains(foreignBefore, "the user's own skill") {
		t.Fatalf("run 1 lost the foreign file:\n%s", foreignBefore)
	}

	// Run 2 — the recorded-local update: today's full deployer + today's
	// Clean walk. --force bypasses the version-compare skip (RK-7) so the
	// walk actually runs. The managed-glob hit (.claude/skills/moai-custom)
	// is backed up and removed — exactly where the one-run promise ends.
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "no-plugin": "true", "force": "true"})
	assertFileAbsent(t, root, migForeignSkill)

	// The removal was backed up (the P-08 rule: files the template does not
	// carry reach the pre-clean backup) — the loss is recoverable.
	matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "pre-clean",
		".claude", "skills", "moai-custom", "SKILL.md"))
	if len(matches) == 0 {
		t.Error("the removed foreign skill left no pre-clean backup copy")
	}
}

// ---------------------------------------------------------------------------
// Card t1438 card-review findings 1 and 2
// ---------------------------------------------------------------------------

// plantSymlinkedArchiveRoot replaces the migration archive root with a
// symlink pointing at an outside directory — the planted-write shape the
// symlink guard must refuse (finding 1). Returns the outside target.
func plantSymlinkedArchiveRoot(t *testing.T, root string) string {
	t.Helper()
	archiveDir := filepath.Join(root, ".moai", "archive")
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(archiveDir), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	if err := os.Symlink(outside, archiveDir); err != nil {
		t.Fatalf("symlink the archive root: %v", err)
	}
	return outside
}

// TestMigrationArchiveRefusesSymlinkedArchiveDestination is finding 1: the
// archive write never goes through a symlink. The current per-file migration
// preserves user-modified assets in place and does not require an archive.
func TestMigrationArchiveRefusesSymlinkedArchiveDestination(t *testing.T) {
	t.Run("unit_archive_through_symlink_refused", func(t *testing.T) {
		root := buildMigrationFixture(t)
		outside := plantSymlinkedArchiveRoot(t, root)

		err := archiveMigrationFile(root, migModifiedSkill)
		if err == nil {
			t.Fatal("the archive write went through a symlinked archive destination without refusal")
		}
		if !strings.Contains(err.Error(), "ARCHIVE_SYMLINK") {
			t.Errorf("error does not carry the ARCHIVE_SYMLINK code: %v", err)
		}
		if !strings.Contains(err.Error(), ".moai/archive") {
			t.Errorf("error does not name the symlinked path: %v", err)
		}
		// Nothing was written through the link.
		if entries, _ := os.ReadDir(outside); len(entries) != 0 {
			t.Errorf("the symlink target received archive writes: %v", entries)
		}
	})

	t.Run("flow_preserves_user_assets_without_archive", func(t *testing.T) {
		root := buildMigrationFixture(t)
		outside := plantSymlinkedArchiveRoot(t, root)
		home := installMigrationUserCounterparts(t)
		if err := migrateProjectCommonAssets(root, home, nil, func(string, ...interface{}) {}); err != nil {
			t.Fatal(err)
		}
		if err := tryRunUpdateCobraCmd(t, root, map[string]string{"yes": "true"}); err != nil {
			t.Fatal(err)
		}
		assertFileAbsent(t, root, migIdenticalSkill)
		if got := readFixtureFile(t, root, migModifiedSkill); !strings.Contains(got, "user edit") {
			t.Fatalf("modified skill bytes lost: %q", got)
		}
		if got := readFixtureFile(t, root, migForeignSkill); !strings.Contains(got, "the user's own skill") {
			t.Fatalf("foreign skill bytes lost: %q", got)
		}
		if entries, err := os.ReadDir(outside); err != nil || len(entries) != 0 {
			t.Fatalf("archive target changed: %v (%v)", entries, err)
		}
	})
}

// tryRunUpdateCobraCmd is runUpdateCobraCmd's error-tolerant form: the caller
// asserts on the returned error instead of the helper failing the test.
func tryRunUpdateCobraCmd(t *testing.T, root string, flags map[string]string) error {
	t.Helper()
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	return runTemplateSyncWithReporter(cmd, nil, true)
}

// TestMigrationPreservesExistingMirrorEntries checks that migration does not
// rewrite user-owned project aliases or provision retired project mirrors.
func TestMigrationPreservesExistingMirrorEntries(t *testing.T) {
	root := buildMigrationFixture(t)
	mirrorEntry := filepath.Join(root, ".agents", "skills", "moai-foundation-core")
	if err := os.MkdirAll(filepath.Dir(mirrorEntry), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join("..", "..", ".claude", "skills", "moai-foundation-core")
	if err := os.Symlink(target, mirrorEntry); err != nil {
		t.Fatal(err)
	}
	userEntry := filepath.Join(root, ".agents", "skills", "moai-user-own")
	if err := os.MkdirAll(userEntry, 0o755); err != nil {
		t.Fatal(err)
	}
	home := installMigrationUserCounterparts(t)
	if err := migrateProjectCommonAssets(root, home, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	if got, err := os.Readlink(mirrorEntry); err != nil || got != target {
		t.Fatalf("user alias changed: %q (%v)", got, err)
	}
	assertFileAbsent(t, root, migIdenticalSkill)
	assertFilePresent(t, root, migModifiedSkill)
	assertFilePresent(t, root, migForeignSkill)
	if _, err := os.Lstat(filepath.Join(root, ".agents", "skills", "moai-workflow-tdd")); !os.IsNotExist(err) {
		t.Fatalf("retired mirror provisioned: %v", err)
	}
	if entries, err := os.ReadDir(userEntry); err != nil || len(entries) != 0 {
		t.Fatalf("user entry changed: %v (%v)", entries, err)
	}
}

// Exercise the same user-install phase that precedes project migration in
// runUpdate, with both roots confined to test-owned temporary directories.
func installMigrationUserCounterparts(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := runUserAssetUpdatePhase(home, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	want, err := fs.ReadFile(embedded, migIdenticalSkill)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(migIdenticalSkill)))
	if err != nil || string(got) != string(want) {
		t.Fatalf("user counterpart differs: %v", err)
	}
	return home
}
