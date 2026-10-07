package cli

// update_migrate_test.go — the M3 migration criteria (SPEC-INIT-SHRINK-001
// acceptance.md AC-011..AC-019), re-baselined to the joint t1547×t1509
// contract (repair round, leader disposition): the plugin probe arms are
// retired with their carrier (SPEC-USER-ASSET-INSTALL-001 M6 — every
// record-less run resolves opted-out and records local), DeployMode is
// Local-only (M7), and the common-asset roots are the user-asset install
// surface whose only removal is the per-file migration
// (migrateProjectCommonAssets — covered by fix_round3_* and review_fix2).
// Every test drives the REAL update flow (runTemplateSyncWithReporter
// through runUpdateCobraCmd) against a fixture old-project tree.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
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

// TestUpdateMigratesLegacyProject is AC-015, re-baselined (repair round):
// the probe arms collapsed — the confirmed arm retired with the plugin
// carrier (M6), so every record-less arm lands the same contract. The
// not-demonstrated and opted-out subtests below keep their historic names;
// this subtest carries the confirmed arm's re-baseline.
func TestUpdateMigratesLegacyProject(t *testing.T) {
	t.Run("confirmed_arm_retired_recordless_local", func(t *testing.T) {
		root := buildMigrationFixture(t)
		runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})

		// The record reads local — the plugin record is retired with its
		// carrier (M6): a record-less project takes the local deployment.
		if got := config.ReadDeployMode(root); got != "local" {
			t.Fatalf("deployment_mode = %q, want local", got)
		}
		// The template sync removes NOTHING (REQ-UPM-002): the old body
		// asserted the confirmed arm's classified removal here, whose only
		// surviving producer is the per-file user-asset migration — covered
		// by the migrateProjectCommonAssets tests (fix_round3_*, review_fix2),
		// not by the sync.
		assertFilePresent(t, root, migIdenticalSkill)
		assertFilePresent(t, root, migModifiedSkill)
		assertFilePresent(t, root, migAbsentRecCmd)
		// Foreign files are preserved byte-for-byte (REQ-013).
		if got := readFixtureFile(t, root, migForeignSkill); !strings.Contains(got, "the user's own skill") {
			t.Errorf("foreign skill content changed:\n%s", got)
		}
		if got := readFixtureFile(t, root, migForeignCommand); !strings.Contains(got, "user's own command") {
			t.Errorf("foreign command content changed:\n%s", got)
		}
		// No migration archive: the archive-then-absorb of user-modified
		// skills is retired — C6 preserves them with a report instead.
		if _, err := os.Stat(filepath.Join(root, ".moai", "archive", "skills", templateMigrationTagForTest)); !os.IsNotExist(err) {
			t.Errorf("record-less local run wrote a migration archive: %v", err)
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

// TestMigrationArchivesModifiedBeforeRemoval keeps only the negative
// control: AC-012's flow-level archive-then-remove subtest was a dead arm of
// the retired confirmed migration (repair round) — the sync never archives
// or removes common assets anymore, and the per-file removal machinery's
// coverage lives in the migrateProjectCommonAssets tests plus the
// archiveMigrationFile unit tests (the symlink-refusal subtest below). The
// control still documents why the unguarded wholesale walk (the legacy
// fresh-install branch's retained path) must never be the default.
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
}

// TestMigrationIdempotent is AC-014, re-baselined to the local record
// (repair round): a second update on a record-less-run project removes
// nothing, archives nothing, and the record is unchanged.
func TestMigrationIdempotent(t *testing.T) {
	root := buildMigrationFixture(t)
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	if got := config.ReadDeployMode(root); got != "local" {
		t.Fatalf("first run record = %q, want local", got)
	}
	archiveDir := filepath.Join(root, ".moai", "archive", "skills", templateMigrationTagForTest)
	var before []string
	if entries, err := os.ReadDir(archiveDir); err == nil {
		for _, e := range entries {
			before = append(before, e.Name())
		}
	}

	// Second run: --force bypasses the version-compare skip (RK-7) so the
	// flow actually runs; the record is present — the trigger short-circuits,
	// and the reconcile preserves the common-asset roots.
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "force": "true"})

	if got := config.ReadDeployMode(root); got != "local" {
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
	assertFilePresent(t, root, migIdenticalSkill)
	assertFilePresent(t, root, migForeignSkill)
}

// TestUpdateForceKeepsPlantedCommonAssets re-aims AC-016 (repair round): the
// plugin-mode thin-deploy premise is retired with the carrier (M7 —
// DeployMode is Local-only). The contract a force run must now hold: a user
// file planted under a common-asset root survives byte-for-byte — the
// deployer skips those roots (REQ-005) and the reconcile lists them
// preserved — and the record never flips.
func TestUpdateForceKeepsPlantedCommonAssets(t *testing.T) {
	root := buildMigrationFixture(t)
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	before := config.ReadDeployMode(root)

	// A user file under a common-asset root, planted before the force run:
	// the deploy must not re-create template copies beside it or touch it.
	plant := filepath.Join(root, ".claude", "skills", "moai-custom", "PLANTED")
	if err := os.MkdirAll(filepath.Dir(plant), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(plant, []byte("x"), 0o644); err != nil {
		t.Fatalf("plant: %v", err)
	}

	runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "force": "true"})

	if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(".claude/skills/moai-custom/PLANTED"))); err != nil {
		t.Errorf("force run removed the planted user file: %v", err)
	} else if string(data) != "x" {
		t.Errorf("force run altered the planted user file: %q", data)
	}
	assertFilePresent(t, root, migForeignSkill)
	if got := config.ReadDeployMode(root); got != before {
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

// TestForceUpdatePreservesUserCommonAssets re-aims AC-019 (repair round,
// leader disposition): the old body pinned SPEC-INIT-SHRINK-001's
// plugin-mode force contract ("--force never re-creates the dropped
// components"), whose dropped-component premise the de-plugin absorption
// retired — DeployMode is Local-only (M7) and the common-asset roots are
// the user-asset install surface (M4), never the sync's removal targets.
// The joint t1547×t1509 contract a force run must hold: the record never
// flips (REQ-018), every common-root file survives byte-for-byte, and the
// summary lists the user's own skill as preserved.
func TestForceUpdatePreservesUserCommonAssets(t *testing.T) {
	root := buildMigrationFixture(t)
	runUpdateCobraCmd(t, root, map[string]string{"yes": "true"})
	foreignBefore := readFixtureFile(t, root, migForeignSkill)
	modifiedBefore := readFixtureFile(t, root, migModifiedSkill)

	out, _ := runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "no-plugin": "true", "force": "true"})

	if got := config.ReadDeployMode(root); got != "local" {
		t.Errorf("force update flipped the record to %q, want local", got)
	}
	if got := readFixtureFile(t, root, migForeignSkill); got != foreignBefore {
		t.Errorf("force update altered the foreign skill:\n%s", got)
	}
	if got := readFixtureFile(t, root, migModifiedSkill); got != modifiedBefore {
		t.Errorf("force update absorbed the user-modified skill:\n%s", got)
	}
	assertFilePresent(t, root, migIdenticalSkill)
	if !strings.Contains(out, "moai-custom") {
		t.Errorf("force-run summary did not list the preserved foreign skill:\n%s", out)
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
			t.Fatal("the run did not abort on the symlinked archive destination")
		}
		// Re-baselined (repair round): the flow's FIRST archive attempt is
		// now the reconciliation's run-dir claim, whose guard refuses the
		// link with the same semantics — the ARCHIVE_SYMLINK code itself
		// stays pinned by the unit subtest above. The contract here is the
		// refusal names the symlinked path.
		if !strings.Contains(err.Error(), ".moai/archive") {
			t.Errorf("abort error does not name the symlink refusal: %v", err)
		}
		// The abort-before-removal contract (OD-3): nothing was removed.
		assertFilePresent(t, root, migIdenticalSkill)
		assertFilePresent(t, root, migModifiedSkill)
		assertFilePresent(t, root, migForeignSkill)
		// The record is unwritten — the next update re-triggers.
		if got := config.ReadDeployMode(root); got != "" {
			t.Errorf("deployment_mode = %q after an aborted run, want empty", got)
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

// (TestMigrationRehomesExistingMirrorEntries was finding 2, design §3
// mirror paragraph: it pinned the confirmed migration's mirror re-home,
// whose producer retired with the plugin carrier — repair round: the re-home
// machinery has no remaining call site, and the dropped-root removal it
// protected against never runs. The surviving user-alias contract is pinned
// by TestMigrationPreservesExistingMirrorEntries.)

// TestUnknownSkillPreservedAcrossLocalUpdates pins the preservation contract
// across the record-less and the recorded-local arm: the user's own skill
// survives both runs byte-for-byte, the record reads local after each.
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
	// The old body pinned the wholesale clean's managed local-only hook
	// removal here ("cleaned with a pre-clean backup") — retired with the
	// wholesale path (repair round): the reconcile classifies a local-only
	// file under a managed root user-owned and preserves it byte-for-byte
	// (REQ-UPM-002/013, the D-15 gate).
	if got, err := os.ReadFile(hook); err != nil {
		t.Fatalf("update removed the user's local-only hook: %v", err)
	} else if string(got) != hookBytes {
		t.Fatalf("update rewrote the user's local-only hook: %q", got)
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
	if err := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}); err != nil {
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

// TestUpdatePreservesUserModifiedPolicies is gate round 4 finding 1's repro
// (card t1547 repair round): in a `both` project the harness-neutral shared
// surfaces (.moai/policies ← .claude/rules/moai) are real deploy paths, so a
// user-modified policies file must classify user-modified and reach the
// merge-or-conflict disposition — a force update preserves the operator's
// bytes and writes the render to the <path>.moai-new sidecar, never an
// overwrite without a sidecar.
func TestUpdatePreservesUserModifiedPolicies(t *testing.T) {
	root := buildMigrationFixture(t)
	// The `both` harness profile deploys the projected shared surfaces.
	writeFixtureFile(t, root, ".moai/config/sections/llm.yaml",
		"llm:\n  harness: both\n")

	// Install the policies file at its deploy path with the template bytes
	// the deploy would carry, track it healthy, THEN land the user's edit —
	// the install-then-edit shape a real project presents.
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	const policiesRel = ".moai/policies/core/moai-constitution.md"
	source, err := fs.ReadFile(embedded, ".claude/rules/moai/core/moai-constitution.md")
	if err != nil {
		t.Fatalf("embedded policies source: %v", err)
	}
	writeFixtureFile(t, root, policiesRel, string(source))
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(policiesRel, manifest.TemplateManaged, manifest.HashBytes(source)); err != nil {
		t.Fatalf("track: %v", err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	const sentinel = "\n<!-- the operator's own policy amendment -->\n"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(policiesRel)),
		[]byte(string(source)+sentinel), 0o644); err != nil {
		t.Fatalf("write user edit: %v", err)
	}

	out, _ := runUpdateCobraCmd(t, root, map[string]string{"yes": "true", "force": "true"})

	// The operator's bytes are restored byte-for-byte...
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(policiesRel)))
	if err != nil {
		t.Fatalf("policies file removed by the force update: %v", err)
	}
	if !strings.Contains(string(got), "the operator's own policy amendment") {
		t.Errorf("force update overwrote the user-modified policies file without preservation:\n%s", out)
	}
	// ...and the render reached the sidecar (REQ-UPM-012).
	sidecar, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(policiesRel+".moai-new")))
	if err != nil {
		t.Fatalf("no conflict sidecar for the user-modified policies file: %v", err)
	}
	if string(sidecar) != string(source) {
		t.Errorf("sidecar does not carry the template render")
	}
}

// TestCommonAssetCleanFilterNormalizesSeparators is the Windows-shaped repro
// for the clean-scope exclusion (repair round, leader directive): on Windows
// ManagedCleanTargets builds DisplayPath with filepath.Join, so the display
// path carries backslashes while projectCommonAssetRels uses slash roots —
// a raw prefix comparison misses, the exclusion silently lets the
// common-asset roots back into the clean scope, and the cleanup deletes
// skills the migration just preserved (user-modified copy loss).
// Preservation is platform-independent: the comparison normalizes to slash
// form first.
func TestCommonAssetCleanFilterNormalizesSeparators(t *testing.T) {
	// Windows-shaped inputs, constructed directly (the shape filepath.Join
	// produces on GOOS=windows).
	windowsShaped := func(rel string) string {
		return strings.ReplaceAll(rel, "/", "\\")
	}
	for _, path := range []string{
		windowsShaped(".claude/skills/moai-foundation-core/SKILL.md"),
		windowsShaped(".claude/skills"),
		windowsShaped(".agents/skills/moai-custom/SKILL.md"),
		windowsShaped(".claude/agents/moai/manager-develop.md"),
		windowsShaped(".codex/agents/moai/manager-develop.toml"),
	} {
		if !isCommonAssetCleanTarget(path) {
			t.Errorf("common-asset filter missed the Windows-shaped path %q — the clean scope would delete preserved assets", path)
		}
	}
	// The slash forms keep matching, and non-common roots stay outside.
	for _, path := range []string{".claude/skills/moai-foundation-core/SKILL.md", ".claude/rules/moai/core"} {
		want := path != ".claude/rules/moai/core"
		if got := isCommonAssetCleanTarget(path); got != want {
			t.Errorf("isCommonAssetCleanTarget(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestMigrationRefusesRemovalWhenUserInstallMissing is the Gate-A repro
// (repair round, leader scope addition #5): the removal arm runs ONLY when
// THIS invocation's user-asset install succeeded. A failed, cancelled, or
// skipped install leaves every project-side asset in place — even one whose
// counterpart record and bytes would otherwise confirm removal, because the
// confirmation evidence a removal rests on is THIS run's install.
func TestMigrationRefusesRemovalWhenUserInstallMissing(t *testing.T) {
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)

	// The counterpart record and bytes WOULD confirm removal (the identical
	// skill's user copy is installed and current) — the missing install
	// still refuses it.
	if err := migrateProjectCommonAssets(root, home, false, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	assertFilePresent(t, root, migIdenticalSkill)
}

// TestUserCounterpartMustBeCurrentVersion is the Gate-B repro (repair round,
// leader scope addition #5): a counterpart record agreeing with an
// OLD-version user file must NOT confirm removal — the counterpart must be
// the CURRENT version's content, byte-equal to the source this binary
// installs from. The old-record-agrees-with-old-file pair is exactly what a
// refused update (a symlink refusal, --templates-only) leaves behind while
// the deletion proceeds, destroying the newest remaining copy.
func TestUserCounterpartMustBeCurrentVersion(t *testing.T) {
	home := t.TempDir()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	const rel = ".claude/skills/moai-foundation-core/SKILL.md"
	stale := []byte("OLD-VERSION CONTENT — no longer what this binary ships\n")
	sum := sha256.Sum256(stale)
	um := &userassets.Manifest{Files: map[string]userassets.FileEntry{
		"claude-skills/moai-foundation-core/SKILL.md": {
			SHA256: hex.EncodeToString(sum[:]), Bundle: "core",
			InstalledAt: "t0", MoaiVersion: "vOld",
		},
	}}
	dir := userassets.RootBySlugDir(home, userassets.RootSlug("claude-skills"))
	if dir == "" {
		t.Fatal("claude-skills root dir unresolved")
	}
	if err := os.MkdirAll(filepath.Join(dir, "moai-foundation-core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "moai-foundation-core", "SKILL.md"), stale, 0o644); err != nil {
		t.Fatal(err)
	}

	// The record matches the stale disk bytes (the F4 gate passes) — the
	// current-version gate must still refuse.
	if userCounterpartConfirmed(um, embedded, home, rel) {
		t.Error("a stale old-version counterpart confirmed removal — the current-version check is missing")
	}

	// The control: the CURRENT version's bytes DO confirm.
	current, err := fs.ReadFile(embedded, rel)
	if err != nil {
		t.Fatalf("embedded source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "moai-foundation-core", "SKILL.md"), current, 0o644); err != nil {
		t.Fatal(err)
	}
	sum = sha256.Sum256(current)
	um.Files["claude-skills/moai-foundation-core/SKILL.md"] = userassets.FileEntry{
		SHA256: hex.EncodeToString(sum[:]), Bundle: "core",
		InstalledAt: "t1", MoaiVersion: "vCurrent",
	}
	if !userCounterpartConfirmed(um, embedded, home, rel) {
		t.Error("the current-version counterpart did not confirm — the gate over-refuses")
	}
}

// TestCancelledUpdateKeepsProjectManagedAssets is the cancel-path repro
// (repair round, gate r5 finding): the project migration's removal arm runs
// AFTER the confirmation gate — cancelling the update leaves the project's
// managed assets byte-intact and the record unwritten. The original defect's
// deletion-before-prompt lived in runUpdate (observed by the gate overlay's
// cancelled-migration repro); this test pins the fixed contract at the gate
// the fix placed it behind.
func TestCancelledUpdateKeepsProjectManagedAssets(t *testing.T) {
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	prev := confirmViaPreviewFn
	confirmViaPreviewFn = func(merge.MergeAnalysis, string) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmViaPreviewFn = prev })

	skipped, err := runTemplateSyncWithProgress(cmd, true)
	if err != nil {
		t.Fatalf("runTemplateSyncWithProgress: %v", err)
	}
	if !skipped {
		t.Errorf("a cancelled run must report skipped, got false")
	}
	if !strings.Contains(out.String(), "Merge cancelled by user") {
		t.Errorf("cancellation banner missing from output:\n%s", out.String())
	}
	// The project's managed asset survives the cancelled run byte-for-byte.
	assertFilePresent(t, root, migIdenticalSkill)
	if got := config.ReadDeployMode(root); got != "" {
		t.Errorf("deployment_mode = %q after a cancelled run, want empty", got)
	}
}
