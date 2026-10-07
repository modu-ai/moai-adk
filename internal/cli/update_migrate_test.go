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

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
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

// TestUpdateMigratesLegacyProject is AC-015: the three arms mapped to the
// probe — confirmed (dedupe + record plugin), not-demonstrated (no dedupe,
// record local), opted-out (full local deploy, record local).
func TestUpdateMigratesLegacyProject(t *testing.T) {
	t.Run("confirmed", func(t *testing.T) {
		root := buildMigrationFixture(t)
		runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

		if got := config.ReadDeployMode(root); got != "plugin" {
			t.Fatalf("deployment_mode = %q, want plugin", got)
		}
		assertFileAbsent(t, root, migIdenticalSkill)
		assertFileAbsent(t, root, migModifiedSkill)
		assertFileAbsent(t, root, migAbsentRecCmd)
		// Foreign files are preserved byte-for-byte (REQ-013).
		if got := readFixtureFile(t, root, migForeignSkill); !strings.Contains(got, "the user's own skill") {
			t.Errorf("foreign skill content changed:\n%s", got)
		}
		if got := readFixtureFile(t, root, migForeignCommand); !strings.Contains(got, "user's own command") {
			t.Errorf("foreign command content changed:\n%s", got)
		}
		// The modified skill's archive holds the pre-run bytes (REQ-012).
		archived := readFixtureFile(t, root,
			".moai/archive/skills/"+templateMigrationTagForTest+"/moai-workflow-tdd/SKILL.md")
		if !strings.Contains(archived, "user edit") {
			t.Errorf("archived copy lost the user's edit:\n%s", archived)
		}
	})

	t.Run("not-demonstrated", func(t *testing.T) {
		root := buildMigrationFixture(t)
		// The default runner refuses under a test binary: every probe read
		// is unreadable → not-demonstrated.
		runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

		if got := config.ReadDeployMode(root); got != "local" {
			t.Fatalf("deployment_mode = %q, want local", got)
		}
		// Nothing was removed: the deployed copies stand untouched.
		assertFilePresent(t, root, migIdenticalSkill)
		assertFilePresent(t, root, migModifiedSkill)
		assertFilePresent(t, root, migForeignSkill)
		// No archive was written (nothing was archived).
		if _, err := os.Stat(filepath.Join(root, ".moai", "archive", "skills", templateMigrationTagForTest)); !os.IsNotExist(err) {
			t.Errorf("not-demonstrated run wrote a migration archive: %v", err)
		}
	})

	t.Run("opted-out", func(t *testing.T) {
		root := buildMigrationFixture(t)
		runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "no-plugin": "true"})

		if got := config.ReadDeployMode(root); got != "local" {
			t.Fatalf("deployment_mode = %q, want local", got)
		}
		// The full local payload deployed (today's path): template-carried
		// skills are present; the record reads local; nothing was archived.
		assertFilePresent(t, root, ".claude/skills")
		if _, err := os.Stat(filepath.Join(root, ".moai", "archive", "skills", templateMigrationTagForTest)); !os.IsNotExist(err) {
			t.Errorf("opted-out run wrote a migration archive: %v", err)
		}
	})
}

// templateMigrationTagForTest reads the migration archive tag through the
// exported constant's package (the classifier owns the layout).
const templateMigrationTagForTest = "init-shrink-migration"

// TestUpdateOptedOutMigrationPreservesManagedRootLocals — SPEC-UPDATE-
// MIGRATION-001 (card t1547, gate round 10 finding 1): a record-less project
// run with --no-plugin takes the migration arm, whose WHOLESALE removal is
// exactly its classified dropped-root list. The managed roots reconcile:
// a local-only rule file survives byte-for-byte, and the legacy
// .moai/memory tree still migrates (finding 4). Before the split, this arm
// wiped the managed roots wholesale — deleting what the dry-run preview had
// promised to preserve.
func TestUpdateOptedOutMigrationPreservesManagedRootLocals(t *testing.T) {
	root := buildMigrationFixture(t)
	const localRule = ".claude/rules/moai/local-rule.md"
	const localRuleBytes = "# operator's dev-only rule — exists in no template\n"
	abs := filepath.Join(root, filepath.FromSlash(localRule))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte(localRuleBytes), 0o644); err != nil {
		t.Fatalf("write local rule: %v", err)
	}
	// Legacy .moai/memory tree (finding 4): the wholesale clean used to run
	// MigrateLegacyMemoryDir at the end of its walk; the reconcile path must
	// carry it too.
	memCkpt := filepath.Join(root, ".moai", "memory", "checkpoints")
	if err := os.MkdirAll(memCkpt, 0o755); err != nil {
		t.Fatalf("mkdir memory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memCkpt, "state.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write checkpoint: %v", err)
	}

	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "no-plugin": "true"})

	// The managed-root local survived byte-for-byte.
	if got := readFixtureFile(t, root, localRule); got != localRuleBytes {
		t.Errorf("managed-root local rule content = %q, want byte-identical %q", got, localRuleBytes)
	}
	// The legacy memory tree migrated (renamed to .moai/state — no state dir
	// pre-existed in the fixture).
	if _, err := os.Stat(filepath.Join(root, ".moai", "memory")); !os.IsNotExist(err) {
		t.Errorf("legacy .moai/memory still present (err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "checkpoints", "state.json")); err != nil {
		t.Errorf("legacy memory not migrated to .moai/state: %v", err)
	}
}

// TestMigrationRemovesIdenticalDroppedComponents is AC-011: identical
// dropped components are removed with the count printed, no archive of an
// identical component is written, and the removal ran through the
// classified-set executor — a foreign file the classified set never
// contains survives the same run untouched.
func TestMigrationRemovesIdenticalDroppedComponents(t *testing.T) {
	root := buildMigrationFixture(t)
	out, _ := runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

	assertFileAbsent(t, root, migIdenticalSkill)
	assertFileAbsent(t, root, migAbsentRecCmd)
	// The counts line names the three classes (REQ-010).
	if !strings.Contains(out, "migration: classified ") {
		t.Errorf("counts line missing from update output:\n%s", out)
	}
	// No archive of an IDENTICAL component exists (OD-3 settled (a)).
	if _, err := os.Stat(filepath.Join(root, ".moai", "archive", "skills",
		templateMigrationTagForTest, "moai-foundation-core")); !os.IsNotExist(err) {
		t.Errorf("an identical component was archived: %v", err)
	}
	// The classified-set executor: the foreign file is not in the removal
	// list, so the same Clean step that removed the classified files left
	// it untouched (the global walk did not run over the dropped roots).
	assertFilePresent(t, root, migForeignSkill)
}

// TestMigrationArchivesModifiedBeforeRemoval is AC-012: modified classified
// files archive (per file, never a whole directory) before removal; an
// archive failure anywhere in the batch aborts with nothing removed. The
// negative control shows the unguarded pre-fix path losing the file.
func TestMigrationArchivesModifiedBeforeRemoval(t *testing.T) {
	t.Run("negative_control_unguarded_path_loses_the_file", func(t *testing.T) {
		// The pre-fix shape (P-08): the global managed-roots walk removes
		// the modified skill with NO archive — the raw template carries the
		// path, so the pre-clean backup exempts it.
		root := buildMigrationFixture(t)
		embedded, err := template.EmbeddedTemplates()
		if err != nil {
			t.Fatalf("load embedded templates: %v", err)
		}
		var sink strings.Builder
		if err := deploy.CleanMoaiManagedPaths(root, &sink, embedded); err != nil {
			t.Fatalf("unguarded clean: %v", err)
		}
		assertFileAbsent(t, root, migModifiedSkill)
		if _, err := os.Stat(filepath.Join(root, ".moai", "archive", "skills",
			templateMigrationTagForTest, "moai-workflow-tdd")); !os.IsNotExist(err) {
			t.Fatalf("the unguarded control unexpectedly archived: %v", err)
		}
	})

	t.Run("migration_archives_then_removes", func(t *testing.T) {
		root := buildMigrationFixture(t)
		runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

		// The modified skill member archived with its pre-run bytes, and the
		// file is gone; the command file (absent manifest record → modified)
		// likewise.
		archivedSkill := readFixtureFile(t, root,
			".moai/archive/skills/"+templateMigrationTagForTest+"/moai-workflow-tdd/SKILL.md")
		if !strings.Contains(archivedSkill, "user edit") {
			t.Errorf("skill archive lost the user's bytes:\n%s", archivedSkill)
		}
		assertFileAbsent(t, root, migModifiedSkill)
		archivedCmd := readFixtureFile(t, root,
			".moai/archive/files/"+templateMigrationTagForTest+"/.claude/commands/moai/todo.md")
		if archivedCmd == "" {
			t.Error("command archive is empty")
		}
		assertFileAbsent(t, root, migAbsentRecCmd)
		// Per-file archive unit: the identical member of the SAME directory
		// family was not archived (no directory-level copy).
		if _, err := os.Stat(filepath.Join(root, ".moai", "archive", "skills",
			templateMigrationTagForTest, "moai-foundation-core")); !os.IsNotExist(err) {
			t.Error("a directory-level archive copied an identical member")
		}
	})
}

// TestMigrationIdempotent is AC-014: a second update on a migrated project
// removes nothing, archives nothing, and the record is unchanged.
func TestMigrationIdempotent(t *testing.T) {
	root := buildMigrationFixture(t)
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	if got := config.ReadDeployMode(root); got != "plugin" {
		t.Fatalf("first run record = %q, want plugin", got)
	}
	archiveDir := filepath.Join(root, ".moai", "archive", "skills", templateMigrationTagForTest)
	var before []string
	if entries, err := os.ReadDir(archiveDir); err == nil {
		for _, e := range entries {
			before = append(before, e.Name())
		}
	}

	// Second run: --force bypasses the version-compare skip (RK-7) so the
	// flow actually runs; the record is present — step 1 short-circuits, and
	// the thin deploy rewrites nothing under the dropped roots.
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "force": "true"})

	if got := config.ReadDeployMode(root); got != "plugin" {
		t.Errorf("second run flipped the record to %q", got)
	}
	var after []string
	if entries, err := os.ReadDir(archiveDir); err == nil {
		for _, e := range entries {
			after = append(after, e.Name())
		}
	}
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("second run changed the archive set: before=%v after=%v", before, after)
	}
	assertFileAbsent(t, root, migIdenticalSkill)
	assertFilePresent(t, root, migForeignSkill)
}

// TestUpdatePluginModeSkipsDroppedRedeploy is AC-016: a plugin-mode
// project's update deploys the thin set — no dropped component is
// re-deployed — and the recorded value is byte-identical after the run.
func TestUpdatePluginModeSkipsDroppedRedeploy(t *testing.T) {
	root := buildMigrationFixture(t)
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	before := config.ReadDeployMode(root)

	// A user file under a dropped root, planted AFTER the migration: the
	// thin redeploy must not re-create template copies beside it.
	plant := filepath.Join(root, ".claude", "skills", "moai-custom", "PLANTED")
	if err := os.MkdirAll(filepath.Dir(plant), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(plant, []byte("x"), 0o644); err != nil {
		t.Fatalf("plant: %v", err)
	}

	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "force": "true"})

	assertFileAbsent(t, root, migIdenticalSkill)
	assertFileAbsent(t, root, migAbsentRecCmd)
	if got := config.ReadDeployMode(root); got != before || got != "plugin" {
		t.Errorf("deployment_mode = %q, want the recorded %q", got, before)
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

// TestUpdateForceDoesNotResurrectDropped is AC-019: --force on a
// plugin-mode project never re-creates the dropped components.
func TestUpdateForceDoesNotResurrectDropped(t *testing.T) {
	root := buildMigrationFixture(t)
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "force": "true"})

	assertFileAbsent(t, root, migIdenticalSkill)
	assertFileAbsent(t, root, migModifiedSkill)
	assertFileAbsent(t, root, migAbsentRecCmd)
	// The only file left under .claude/skills is the preserved foreign
	// skill; the classified executor's file-level removal leaves no
	// template copy behind. (Empty directory shells may remain — the
	// removal unit is the classified file.)
	var skillFiles []string
	_ = filepath.WalkDir(filepath.Join(root, ".claude", "skills"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			skillFiles = append(skillFiles, filepath.ToSlash(p))
		}
		return nil
	})
	if len(skillFiles) != 1 || !strings.Contains(skillFiles[0], "moai-custom") {
		t.Errorf("force update left unexpected skill files: %v", skillFiles)
	}
}

// TestNotDemonstratedPreservationHoldsAtNextLocalUpdate — SPEC-UPDATE-
// MIGRATION-001 (card t1547): the preservation contract supersedes
// SPEC-INIT-SHRINK-001's D-16 narrowed promise ("preservation is one
// migration run", leader condition 5b) whose removal-at-next-local-update
// behavior this test previously pinned. The next recorded-local update
// RECONCILES the managed roots: a foreign file under the moai* managed glob
// (template-carriage gate — not carried, no manifest record) classifies
// user-owned and survives byte-for-byte, listed in the reconciliation
// summary (REQ-UPM-002/013; AC-UPM-001's flow-level form). The old
// managed-glob removal-with-backup was exactly the D-15 loss path
// REQ-UPM-002 exists to close (leader-approved test-contract update,
// 2026-10-07).
func TestNotDemonstratedPreservationHoldsAtNextLocalUpdate(t *testing.T) {
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

	// Run 2 — the recorded-local update reconciles the managed roots: the
	// foreign skill is user-owned (never removal-eligible) and survives
	// byte-for-byte, and the summary lists it as preserved.
	out, _ := runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "no-plugin": "true", "force": "true"})
	if got := readFixtureFile(t, root, migForeignSkill); got != foreignBefore {
		t.Errorf("run 2 altered the foreign skill:\n%s", got)
	}
	if !strings.Contains(out, "moai-custom") {
		t.Errorf("run 2 summary did not list the preserved foreign skill:\n%s", out)
	}
}

// TestPreservedSectionSurvivesRestoreUntouched pins gate round 23, finding 2
// (P2): the Restore Settings skip filter excluded only the reconciliation's
// archived-removed section files, so a PRESERVED local-only sections file was
// re-merged from the config backup and rewritten with normalized YAML bytes
// (4-space indentation collapsed) — contradicting the summary's preserved
// report (REQ-UPM-002: user-owned files are never rewritten). The skip set
// must carry the preserved paths too.
func TestPreservedSectionSurvivesRestoreUntouched(t *testing.T) {
	root := buildMigrationFixture(t)

	// A local-only sections file: never template-carried, never manifest-
	// recorded — the reconciliation classifies it user-owned and preserves it
	// byte-for-byte. The 4-space indentation is the canary: any YAML round-
	// trip rewrites it.
	localOnly := ".moai/config/sections/local-only.yaml"
	content := "deploy:\n    target:\n        region: local\n"
	abs := filepath.Join(root, filepath.FromSlash(localOnly))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	out, _ := runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	if !strings.Contains(out, "local-only.yaml") {
		t.Errorf("summary did not report the local-only section as preserved:\n%s", out)
	}
	if got := readFixtureFile(t, root, localOnly); got != content {
		t.Errorf("the update rewrote the preserved local-only section file:\n--- before ---\n%s\n--- after ---\n%s", content, got)
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
// archive write never goes THROUGH a symlink. A symlinked archive
// destination (or parent component) aborts the archive — which aborts the
// migration before any removal — with an error naming the symlink.
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

	t.Run("flow_aborts_before_any_removal", func(t *testing.T) {
		root := buildMigrationFixture(t)
		plantSymlinkedArchiveRoot(t, root)

		err := tryRunUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
		if err == nil {
			t.Fatal("the confirmed migration did not abort on the symlinked archive destination")
		}
		if !strings.Contains(err.Error(), "ARCHIVE_SYMLINK") {
			t.Errorf("abort error does not name the symlink refusal: %v", err)
		}
		// The abort-before-removal contract (OD-3): nothing was removed.
		assertFilePresent(t, root, migIdenticalSkill)
		assertFilePresent(t, root, migModifiedSkill)
		assertFilePresent(t, root, migForeignSkill)
		// The record is unwritten — the next update re-triggers.
		if got := config.ReadDeployMode(root); got != "" {
			t.Errorf("deployment_mode = %q after an aborted migration, want empty", got)
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

// TestMigrationRehomesExistingMirrorEntries is finding 2 (design §3 mirror
// paragraph): on a confirmed migration the KEPT mirror entries are re-homed
// to real directory copies rendered from the embedded tree BEFORE the
// dropped-root removal runs — a kept symlink would dangle into the removed
// .claude/skills. The re-home converts existing entries; it never provisions
// new ones (REQ-019), and a non-link entry is the user's and stays untouched.
func TestMigrationRehomesExistingMirrorEntries(t *testing.T) {
	root := buildMigrationFixture(t)
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}

	// A pre-existing mirror entry in the local deploy's P-11 symlink shape.
	mirrorEntry := filepath.Join(root, ".agents", "skills", "moai-foundation-core")
	if err := os.MkdirAll(filepath.Dir(mirrorEntry), 0o755); err != nil {
		t.Fatalf("mkdir mirror root: %v", err)
	}
	if err := os.Symlink(filepath.Join("..", "..", ".claude", "skills", "moai-foundation-core"), mirrorEntry); err != nil {
		t.Fatalf("seed mirror symlink: %v", err)
	}
	// A non-link entry is the user's: untouched by the re-home.
	userEntry := filepath.Join(root, ".agents", "skills", "moai-user-own")
	if err := os.MkdirAll(userEntry, 0o755); err != nil {
		t.Fatalf("mkdir user mirror entry: %v", err)
	}

	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

	// The entry is now a REAL directory (not a symlink), holding the
	// embedded skill bytes.
	info, err := os.Lstat(mirrorEntry)
	if err != nil {
		t.Fatalf("re-homed mirror entry missing: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("the mirror entry is still a symlink — it dangles once .claude/skills is removed")
	}
	if !info.IsDir() {
		t.Fatalf("re-homed mirror entry is not a directory: %v", info.Mode())
	}
	copied, err := os.ReadFile(filepath.Join(mirrorEntry, "SKILL.md"))
	if err != nil {
		t.Fatalf("re-homed copy missing SKILL.md: %v", err)
	}
	embeddedData, err := fs.ReadFile(embedded, migIdenticalSkill)
	if err != nil {
		t.Fatalf("embedded read: %v", err)
	}
	if string(copied) != string(embeddedData) {
		t.Error("re-homed copy does not carry the embedded skill bytes")
	}

	// The classified removal still ran (after the re-home).
	assertFileAbsent(t, root, migIdenticalSkill)

	// No provisioning: a catalog skill with no pre-existing entry gains none.
	if _, err := os.Lstat(filepath.Join(root, ".agents", "skills", "moai-workflow-tdd")); !os.IsNotExist(err) {
		t.Errorf("the migration provisioned a mirror entry for a skill that had none: %v", err)
	}
	// The user's own entry stays untouched (an empty dir is still empty).
	if entries, _ := os.ReadDir(userEntry); len(entries) != 0 {
		t.Errorf("the user's own mirror entry was modified: %v", entries)
	}
}
func TestUnknownSkillPreservedAcrossLocalUpdates(t *testing.T) {
	root := buildMigrationFixture(t)
	before := readFixtureFile(t, root, migForeignSkill)
	hook := filepath.Join(root, ".claude", "hooks", "moai", "local-only-review.sh")
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	const hookBytes = "local hook must remain recoverable"
	if err := os.WriteFile(hook, []byte(hookBytes), 0o644); err != nil {
		t.Fatal(err)
	}
	for i, flags := range []map[string]string{
		{"yes": "true"},
		{"yes": "true", "no-plugin": "true", "force": "true"},
	} {
		runUpdateCobraCmd(t, root, flags)
		if got := config.ReadDeployMode(root); got != "local" {
			t.Errorf("run%d mode=%q, want local", i+1, got)
		}
		if got := readFixtureFile(t, root, migForeignSkill); got != before {
			t.Fatalf("run%d altered unknown skill: %q, want %q", i+1, got, before)
		}
	}
	if _, err := os.Stat(hook); !os.IsNotExist(err) {
		t.Fatalf("managed local-only hook not cleaned: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "pre-clean", ".claude", "hooks", "moai", "local-only-review.sh"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("managed hook has no pre-clean backup: %v", err)
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != hookBytes {
			t.Fatalf("managed hook backup differs: %q, %v", data, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Card t1438 card-review findings 1 and 2
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
