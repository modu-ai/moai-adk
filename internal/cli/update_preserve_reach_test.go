// Package cli — update_preserve_reach_test.go
//
// Extends the user-area preservation assertion to the destructive surfaces the
// existing byte-identity guard never reached (REQ-UGE-009 .. REQ-UGE-011).
//
// The gap this closes is NOT coverage. Both runCleanReinstall and the backup
// subsystem are already executed by other tests in this package; what no test
// asserted was that user-owned directories are BYTE-IDENTICAL across those
// paths. snapshotDir — the helper that makes that assertion — was used in
// exactly one file (update_safety_test.go) against exactly one entry point
// (deploy.CleanMoaiManagedPaths).
//
// Each guard here targets a surface that actually deletes:
//
//	runCleanReinstall            → the os.RemoveAll(abs) loop over scanDeprecatedPaths
//	BackupMoaiConfig             → the failure-rollback os.RemoveAll(backupDir)
//	CleanupOldBackups            → the rotation delete (historically inverted — see below)
//	MigrateLegacyMemoryDir       → the both-exist backup-then-remove branch
//
// Every fixture uses t.TempDir() per CLAUDE.local.md §6 HARD.

package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// userOwnedGuardPaths is the set whose byte-identity every guard in this file
// asserts. It matches the set the existing TestMoaiUpdate_PreservesUserArea
// guards, so the two files make the same claim over different entry points.
func userOwnedGuardPaths(root string) []string {
	return []string{
		filepath.Join(root, ".moai", "harness"),
		filepath.Join(root, ".claude", "agents", "harness"),
		filepath.Join(root, ".claude", "skills", "harness-ios-patterns"),
	}
}

// seedUserOwnedArea writes content into all three user-owned directories.
//
// .moai/harness is seeded as a real directory on purpose: AC-UGE-009F's
// falsification injects ".moai/harness" into scanDeprecatedPaths, and the
// backup stage that runs before the removal loop behaves differently for an
// existing directory than for an absent path. Seeding it is what makes that
// falsification exercise the real code path.
func seedUserOwnedArea(t *testing.T, root string) {
	t.Helper()
	writeFixture(t, root, ".moai/harness/usage-log.jsonl", "{\"event\":\"user\"}\n")
	writeFixture(t, root, ".moai/harness/notes.md", "user harness notes\n")
	writeFixture(t, root, ".claude/agents/harness/hns-my-specialist.md", "user agent\n")
	writeFixture(t, root, ".claude/skills/harness-ios-patterns/SKILL.md", "user skill\n")
}

// TestCleanReinstall_PreservesUserArea drives the real runCleanReinstall
// orchestrator — including its destructive Step 4, the os.RemoveAll(abs) loop
// over scanDeprecatedPaths — and asserts the user-owned tree is byte-identical
// across it (REQ-UGE-009).
//
// runCleanReinstall does NOT call CleanMoaiManagedPaths (verified by grep: 0
// matches in update_clean_install.go), so the existing guard over that function
// says nothing about this path.
func TestCleanReinstall_PreservesUserArea(t *testing.T) {
	root := t.TempDir()

	// v2 fingerprint — without it Step 1 returns a no-op and no destructive
	// step ever runs, so the guard would pass vacuously.
	writeFixture(t, root, ".moai/config/sections/system.yaml", "moai:\n    version: v2.16.1\n")
	writeFixture(t, root, ".agency/index.md", "legacy agency content\n")

	// A deprecated path that really exists, so Step 4's removal loop has a
	// target and the destructive surface is actually traversed.
	writeFixture(t, root, ".claude/commands/agency/agency.md", "deprecated command\n")

	seedUserOwnedArea(t, root)

	userPaths := userOwnedGuardPaths(root)
	pre := make([]map[string]string, len(userPaths))
	for i, p := range userPaths {
		pre[i] = snapshotDir(t, p)
		if len(pre[i]) == 0 {
			t.Fatalf("fixture seeding failed: %s is empty, so the byte-identity assertion would be vacuous", p)
		}
	}

	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}
	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}

	// Direction 1 — the user area is byte-identical.
	for i, p := range userPaths {
		post := snapshotDir(t, p)
		if !mapsEqual(pre[i], post) {
			t.Errorf("user area changed: %s\npre:  %v\npost: %v", p, pre[i], post)
		}
	}

	// Direction 2 — the destructive step actually ran. Without this the guard
	// would pass against an orchestrator that removed nothing at all.
	if _, err := os.Stat(filepath.Join(root, ".claude", "commands", "agency", "agency.md")); !os.IsNotExist(err) {
		t.Errorf("deprecated path survived Step 4; the destructive surface was not traversed (stat err = %v)", err)
	}
}

// TestBackupSubsystem_DestructiveSurfaces guards the two places the backup
// subsystem actually deletes (REQ-UGE-010). BackupMoaiConfig's main path is a
// pure copy, so a "user area unchanged" assertion over it would be near-vacuous
// — these two surfaces are where data can be lost.
func TestBackupSubsystem_DestructiveSurfaces(t *testing.T) {
	// Surface (a) — the failure rollback. When BackupMoaiConfig fails partway,
	// it removes the backup directory IT created. It must not touch sibling
	// backups from earlier runs, nor the user area.
	t.Run("rollback_removes_only_own_backup_dir", func(t *testing.T) {
		root := t.TempDir()
		seedUserOwnedArea(t, root)

		// A sibling backup from an earlier run that MUST survive.
		siblingBackup := filepath.Join(root, defs.BackupsDir, "20200101_000000")
		writeFixture(t, filepath.Join(siblingBackup, "config"), "sections/system.yaml", "earlier run\n")

		configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
		writeFixture(t, configDir, "sections/system.yaml", "moai:\n    version: v3.0.0\n")

		// Trigger: a dangling symlink inside the config tree. filepath.Walk
		// reports it via Lstat as a non-directory, then os.ReadFile follows it
		// and fails with ENOENT, so the copy loop errors and the rollback runs.
		// This is a content-shaped trigger, not a permission-bit one, so root
		// does not bypass it.
		if err := os.Symlink(filepath.Join(configDir, "no-such-target"), filepath.Join(configDir, "dangling.yaml")); err != nil {
			// Windows refuses symlink creation without the privilege. There is
			// no portable content-shaped way to fail this copy loop, so the
			// subtest reports rather than silently passing.
			t.Skipf("cannot create a symlink on %s (%v); the rollback branch has no portable content-shaped trigger", runtime.GOOS, err)
		}

		userPre := make([]map[string]string, len(userOwnedGuardPaths(root)))
		for i, p := range userOwnedGuardPaths(root) {
			userPre[i] = snapshotDir(t, p)
		}
		siblingPre := snapshotDir(t, siblingBackup)
		if len(siblingPre) == 0 {
			t.Fatalf("sibling backup fixture is empty; its survival assertion would be vacuous")
		}

		backupDir, err := backup.BackupMoaiConfig(root)
		if err == nil {
			t.Fatalf("BackupMoaiConfig unexpectedly succeeded; the rollback branch was not reached")
		}

		// Its own backup directory is gone.
		if backupDir != "" {
			if _, statErr := os.Stat(backupDir); !os.IsNotExist(statErr) {
				t.Errorf("rollback left its own backup dir behind: %s (stat err = %v)", backupDir, statErr)
			}
		}

		// The sibling backup survived — the rollback radius did not widen to
		// the whole backups root.
		if post := snapshotDir(t, siblingBackup); !mapsEqual(siblingPre, post) {
			t.Errorf("sibling backup changed during rollback\npre:  %v\npost: %v", siblingPre, post)
		}

		// The user area survived.
		for i, p := range userOwnedGuardPaths(root) {
			if post := snapshotDir(t, p); !mapsEqual(userPre[i], post) {
				t.Errorf("user area changed during rollback: %s\npre:  %v\npost: %v", p, userPre[i], post)
			}
		}
	})

	// Surface (b) — rotation. CleanupOldBackups keeps the NEWEST keepCount and
	// deletes the oldest excess. A prior revision deleted backups[keepCount:] —
	// the newest — destroying the most recent restore points on every rotation.
	// This subtest asserts WHICH backups survive, not how many: a count-only
	// assertion passes against exactly that inversion.
	t.Run("rotation_keeps_newest", func(t *testing.T) {
		root := t.TempDir()
		seedUserOwnedArea(t, root)

		// Five timestamped backups, oldest to newest by name.
		stamps := []string{
			"20200101_000000",
			"20210101_000000",
			"20220101_000000",
			"20230101_000000",
			"20240101_000000",
		}
		for _, s := range stamps {
			writeFixture(t, filepath.Join(root, defs.BackupsDir, s), "marker.txt", s+"\n")
		}

		const keepCount = 2
		deleted := backup.CleanupOldBackups(root, keepCount)
		if want := len(stamps) - keepCount; deleted != want {
			t.Errorf("CleanupOldBackups deleted %d; want %d", deleted, want)
		}

		// Identity assertion — the survivors must be the NEWEST keepCount.
		wantSurvivors := stamps[len(stamps)-keepCount:]
		wantDeleted := stamps[:len(stamps)-keepCount]

		for _, s := range wantSurvivors {
			if _, err := os.Stat(filepath.Join(root, defs.BackupsDir, s)); err != nil {
				t.Errorf("newest backup %s was deleted; rotation destroyed the most recent restore point (stat err = %v)", s, err)
			}
		}
		for _, s := range wantDeleted {
			if _, err := os.Stat(filepath.Join(root, defs.BackupsDir, s)); !os.IsNotExist(err) {
				t.Errorf("oldest excess backup %s survived rotation (stat err = %v)", s, err)
			}
		}

		// The rotation must not reach outside the backups root.
		for _, p := range userOwnedGuardPaths(root) {
			if len(snapshotDir(t, p)) == 0 {
				t.Errorf("user area emptied by rotation: %s", p)
			}
		}
	})
}

// TestMigrateLegacyMemoryDir_PreservesUserArea drives both destructive branches
// of MigrateLegacyMemoryDir deterministically (REQ-UGE-011). The branches are
// selected by whether .moai/state/ exists, so the fixtures differ only in that.
func TestMigrateLegacyMemoryDir_PreservesUserArea(t *testing.T) {
	// Branch 1 — .moai/state/ absent: the legacy directory is renamed.
	t.Run("rename_when_state_absent", func(t *testing.T) {
		root := t.TempDir()
		seedUserOwnedArea(t, root)
		writeFixture(t, root, ".moai/memory/MEMORY.md", "user memory index\n")

		stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)
		if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
			t.Fatalf("fixture precondition failed: .moai/state must be absent to select the rename branch (stat err = %v)", err)
		}

		userPre := make([]map[string]string, len(userOwnedGuardPaths(root)))
		for i, p := range userOwnedGuardPaths(root) {
			userPre[i] = snapshotDir(t, p)
		}

		if err := deploy.MigrateLegacyMemoryDir(root, io.Discard); err != nil {
			t.Fatalf("MigrateLegacyMemoryDir: %v", err)
		}

		// The rename actually happened — content moved, not vanished.
		moved := filepath.Join(stateDir, "MEMORY.md")
		data, err := os.ReadFile(moved)
		if err != nil {
			t.Fatalf("renamed content missing at %s: %v", moved, err)
		}
		if string(data) != "user memory index\n" {
			t.Errorf("renamed content differs: %q", string(data))
		}

		for i, p := range userOwnedGuardPaths(root) {
			if post := snapshotDir(t, p); !mapsEqual(userPre[i], post) {
				t.Errorf("user area changed: %s\npre:  %v\npost: %v", p, userPre[i], post)
			}
		}
	})

	// Branch 2 — both exist: the legacy directory is backed up, THEN removed.
	//
	// Asserting only "the legacy directory is gone" would pass against a
	// mutation that removes it without backing it up first — which is exactly
	// the unrecoverable data loss REQ-UDS-008 exists to prevent. The load-
	// bearing assertion is therefore that the BACKUP exists and its content
	// matches what was removed.
	t.Run("backup_then_remove_when_both_exist", func(t *testing.T) {
		root := t.TempDir()
		seedUserOwnedArea(t, root)

		const legacyContent = "user memory that must survive\n"
		writeFixture(t, root, ".moai/memory/MEMORY.md", legacyContent)
		writeFixture(t, root, ".moai/state/existing.json", "{}\n")

		userPre := make([]map[string]string, len(userOwnedGuardPaths(root)))
		for i, p := range userOwnedGuardPaths(root) {
			userPre[i] = snapshotDir(t, p)
		}

		if err := deploy.MigrateLegacyMemoryDir(root, io.Discard); err != nil {
			t.Fatalf("MigrateLegacyMemoryDir: %v", err)
		}

		legacyDir := filepath.Join(root, defs.MoAIDir, "memory")
		if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
			t.Errorf("legacy .moai/memory survived the both-exist branch (stat err = %v)", err)
		}

		// The removed content must be recoverable from the backup.
		found := findLegacyMemoryBackup(t, filepath.Join(root, defs.BackupsDir), "MEMORY.md")
		if found == "" {
			t.Fatalf("no legacy-memory backup of MEMORY.md found under %s — the legacy directory was removed without a backup, so user data is unrecoverable",
				filepath.Join(root, defs.BackupsDir))
		}
		data, err := os.ReadFile(found)
		if err != nil {
			t.Fatalf("read legacy-memory backup %s: %v", found, err)
		}
		if string(data) != legacyContent {
			t.Errorf("legacy-memory backup content differs\n got: %q\nwant: %q", string(data), legacyContent)
		}

		for i, p := range userOwnedGuardPaths(root) {
			if post := snapshotDir(t, p); !mapsEqual(userPre[i], post) {
				t.Errorf("user area changed: %s\npre:  %v\npost: %v", p, userPre[i], post)
			}
		}
	})
}

// findLegacyMemoryBackup returns the path of the first file named leaf that
// lives under a "legacy-memory" directory anywhere below backupsRoot, or "" if
// none exists. The backup directory is timestamped at run time, so the search
// is by structure rather than by a predicted path.
func findLegacyMemoryBackup(t *testing.T, backupsRoot, leaf string) string {
	t.Helper()
	var found string
	_ = filepath.WalkDir(backupsRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || found != "" {
			return nil //nolint:nilerr // absent backups root is a legitimate "not found"
		}
		if filepath.Base(p) == leaf && strings.Contains(filepath.ToSlash(p), "/legacy-memory/") {
			found = p
		}
		return nil
	})
	return found
}
