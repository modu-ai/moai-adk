package cli

// migrate_removal_backup_test.go — REQ-RC30-004 / AC-RC30-03 (card t1668,
// SPEC-UPDATE-REMOVAL-FAMILY-001 M1): every file the project migration
// removes lands a byte-identical copy in the RUN backup before the removal,
// and the migration runs after the run's backup stage. The order witness is
// structural: the removed-asset copies must live in the SAME run directory
// the Backup stage created (the one holding backup_metadata.json) — a
// migration that ran before the backup stage would have no run directory to
// write into and would create its own.

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// runUpdateFlowWithProgress drives the full update flow (the same entry
// point `moai update` uses) in the scratch project with --yes, so the
// Backup stage, the migration, and the deploy all run in their production
// order.
func runUpdateFlowWithProgress(t *testing.T, root string) string {
	t.Helper()
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&strings.Builder{})
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if _, err := runTemplateSyncWithProgress(cmd, true); err != nil {
		t.Fatalf("runTemplateSyncWithProgress: %v", err)
	}
	return out.String()
}

// removalBackupCopy locates the run-backup copy of a removed project asset:
// the unique .moai-backups/<run>/removed-assets/<rel> file. It returns the
// run directory and the copy's absolute path, failing the test when the copy
// is absent or ambiguous — the copy IS the criterion under judgment.
func removalBackupCopy(t *testing.T, root, rel string) (string, string) {
	t.Helper()
	pattern := filepath.Join(root, ".moai-backups", "*", "removed-assets", filepath.FromSlash(rel))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob removal backups: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one run-backup copy of %s under .moai-backups, found %d: %v", rel, len(matches), matches)
	}
	suffix := filepath.Join("removed-assets", filepath.FromSlash(rel))
	if !strings.HasSuffix(matches[0], suffix) {
		t.Fatalf("copy path %s does not carry the expected layout %s", matches[0], suffix)
	}
	return strings.TrimSuffix(matches[0], suffix), matches[0]
}

// TestMigrationBackupBeforeRemoval is AC-RC30-03 (U-02, REQ-RC30-004): with
// the removal switch on in the test fixture, the full update flow removes the
// confirmed identical skill AND (1) the run backup holds a byte-identical
// copy of every removed file, (2) the copy lives in the run directory the
// Backup stage created — the migration ran after the backup stage, and (3)
// the dropped manifest entry stays dropped after the whole flow (the
// lost-update guard's happy path).
func TestMigrationBackupBeforeRemoval(t *testing.T) {
	enableRemovalForTest(t)
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	runUpdateFlowWithProgress(t, root)

	// The removal really happened (the fixture meets every removal
	// precondition: template-managed record, matching bytes, confirmed
	// counterpart, switch on).
	assertFileAbsent(t, root, migIdenticalSkill)

	// (1) The run backup holds a byte-identical copy of the removed file.
	runDir, copyPath := removalBackupCopy(t, root, migIdenticalSkill)
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	wantBytes, err := fs.ReadFile(embedded, migIdenticalSkill)
	if err != nil {
		t.Fatalf("embedded read: %v", err)
	}
	gotBytes, err := os.ReadFile(copyPath)
	if err != nil {
		t.Fatalf("read removal backup copy: %v", err)
	}
	if string(gotBytes) != string(wantBytes) {
		t.Errorf("the run-backup copy of %s is not byte-identical to the removed file (%d vs %d bytes)",
			migIdenticalSkill, len(gotBytes), len(wantBytes))
	}

	// (2) The copy shares the run directory the Backup stage created: the
	// stage's own artifact (backup_metadata.json) sits beside removed-assets/.
	if _, err := os.Stat(filepath.Join(runDir, "backup_metadata.json")); err != nil {
		t.Errorf("the removed-asset copy is not in the Backup stage's run directory %s (backup_metadata.json: %v) — the migration did not run behind the backup stage", runDir, err)
	}

	// (3) The dropped manifest entry survived the FULL flow as dropped: a
	// later stale-view Save would resurrect it (the lost-update window
	// plan-audit-iter2 R1 reproduced).
	disk := manifest.NewManager()
	if _, err := disk.Load(root); err != nil {
		t.Fatalf("reload manifest after the flow: %v", err)
	}
	if _, found := disk.GetEntry(migIdenticalSkill); found {
		t.Errorf("the removed file's manifest entry resurrected after the full update flow")
	}
}

// TestRemovalBackupRefusesPlantedSymlinkDestination — sync-audit D1 (card
// t1668): the removal-backup DESTINATION is project-controlled content. A
// symlink planted under the run backup directory must never route the
// pre-removal copy outside the project, and a symlink planted at the copy
// path itself must never turn the backup into a no-copy write onto the
// source. Both shapes refuse the removal: the file stays project-side and
// the run reports the refusal.
func TestRemovalBackupRefusesPlantedSymlinkDestination(t *testing.T) {
	t.Run("removed_assets_dir_symlinked_outside", func(t *testing.T) {
		enableRemovalForTest(t)
		root := buildMigrationFixture(t)
		home := installMigrationUserCounterparts(t)
		outside := t.TempDir()

		// The run backup directory the migration will reuse (the rc path
		// makes the destination deterministic), with removed-assets planted
		// as a symlink pointing outside the project.
		runDir := filepath.Join(root, ".moai-backups", "20260101_000000")
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(runDir, "removed-assets")); err != nil {
			t.Fatal(err)
		}

		rc := &migrationRunContext{runBackupDir: runDir}
		var lines []string
		removed, err := migrateProjectCommonAssets(root, home, true, nil, func(format string, args ...interface{}) {
			lines = append(lines, fmt.Sprintf(format, args...))
		}, rc)
		if err != nil {
			t.Fatal(err)
		}
		if len(removed) != 0 {
			t.Errorf("the migration removed %v despite the unsafe backup destination", removed)
		}
		assertFilePresent(t, root, migIdenticalSkill)
		if !slices.ContainsFunc(lines, func(l string) bool {
			return strings.Contains(l, "removal refused") && strings.Contains(l, migIdenticalSkill)
		}) {
			t.Errorf("no removal-refused line for %s in the migration output: %q", migIdenticalSkill, lines)
		}
		if entries, _ := os.ReadDir(outside); len(entries) != 0 {
			t.Errorf("the backup write escaped through the planted symlink: %v", entries)
		}
	})

	t.Run("dst_symlink_onto_source_refused", func(t *testing.T) {
		root := t.TempDir()
		src := filepath.Join(root, "asset.md")
		if err := os.WriteFile(src, []byte("the only copy"), 0o644); err != nil {
			t.Fatal(err)
		}
		runDir := filepath.Join(root, ".moai-backups", "20260101_000000")
		if err := os.MkdirAll(filepath.Join(runDir, "removed-assets"), 0o755); err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(runDir, "removed-assets", "asset.md")
		if err := os.Symlink(src, dst); err != nil {
			t.Fatal(err)
		}

		rc := &migrationRunContext{runBackupDir: runDir}
		err := backupRemovedAsset(root, rc, "asset.md", []byte("the only copy"))
		if err == nil {
			t.Fatal("the backup write followed a symlink planted at the destination — the copy is a no-write and os.Remove would take the only copy")
		}
		if !strings.Contains(err.Error(), "BACKUP_DEST_UNSAFE") {
			t.Errorf("the refusal does not carry the destination-unsafe sentinel: %v", err)
		}
		// The planted link is intact and no regular-file copy replaced it.
		if info, statErr := os.Lstat(dst); statErr != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("the planted symlink was disturbed: %v (%v)", info, statErr)
		}
	})
}

// TestEmptiedSkillDirRemoved is AC-RC30-11 (U-10, REQ-RC30-004): a skill
// directory the migration empties by removing its last file goes WITH that
// file, and a directory that still holds user content stays.
func TestEmptiedSkillDirRemoved(t *testing.T) {
	t.Run("emptied_dir_removed_with_its_last_file", func(t *testing.T) {
		enableRemovalForTest(t)
		root := buildMigrationFixture(t)
		home := installMigrationUserCounterparts(t)

		if _, err := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}); err != nil {
			t.Fatal(err)
		}
		assertFileAbsent(t, root, migIdenticalSkill)
		// The fixture's moai-foundation-core directory held only SKILL.md:
		// its last file gone, the directory itself must be gone too.
		if _, err := os.Stat(filepath.Join(root, ".claude", "skills", "moai-foundation-core")); !os.IsNotExist(err) {
			t.Errorf("the emptied skill directory survived its last file's removal (stat err: %v)", err)
		}
		// The skills ROOT itself is never a pruning target.
		if _, err := os.Stat(filepath.Join(root, ".claude", "skills")); err != nil {
			t.Errorf("the skills root was pruned: %v", err)
		}
		// Manifest consistency: the dropped file's entry is gone (the
		// entries-drop loop keys on the FILE, so pruning the directory must
		// not disturb it).
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		if _, found := mgr.GetEntry(migIdenticalSkill); found {
			t.Errorf("the removed file's manifest entry survived the removal")
		}
	})

	t.Run("dir_with_user_content_stays", func(t *testing.T) {
		enableRemovalForTest(t)
		root := buildMigrationFixture(t)
		home := installMigrationUserCounterparts(t)
		// A user file beside the managed one: after SKILL.md is removed the
		// directory still holds user content, so it must stay.
		notes := filepath.Join(root, ".claude", "skills", "moai-foundation-core", "NOTES.md")
		if err := os.MkdirAll(filepath.Dir(notes), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(notes, []byte("# the user's own notes\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}); err != nil {
			t.Fatal(err)
		}
		assertFileAbsent(t, root, migIdenticalSkill)
		if data, err := os.ReadFile(notes); err != nil {
			t.Errorf("the user's file in the not-emptied directory was lost: %v", err)
		} else if string(data) != "# the user's own notes\n" {
			t.Errorf("the user's file in the not-emptied directory was altered: %q", data)
		}
		if _, err := os.Stat(filepath.Join(root, ".claude", "skills", "moai-foundation-core")); err != nil {
			t.Errorf("a directory still holding user content was removed: %v", err)
		}
	})
}
