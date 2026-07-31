package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// SPEC-UPDATE-REINSTALL-LOOP-002 M3 — residue cleanup for v3-confirmed projects
// (REQ-RIL2-018..023 / AC-RIL2-010, AC-RIL2-011, AC-RIL2-012, AC-RIL2-019)
// ---------------------------------------------------------------------------

// residueDeprecatedRel is a plain-file entry of defs.DeprecatedPaths used as the
// residue fixture. A file (not a directory) keeps the fixture minimal; the
// directory-expansion path is exercised separately by the .agency fixture in
// TestResidueCleanup_PreservesAgencyMigrationPreStep.
const residueDeprecatedRel = ".claude/agents/moai/planner.md"

// residueWriteFile writes content at projectRoot/rel, creating parents.
func residueWriteFile(t *testing.T, projectRoot, rel, content string) string {
	t.Helper()
	abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
	return abs
}

// newV3ConfirmedProject builds a t.TempDir() project whose system.yaml marker is
// present and whose moai.version is v3-confirmed (REQ-RIL2-002 after M1).
func newV3ConfirmedProject(t *testing.T, version string) string {
	t.Helper()
	root := t.TempDir()
	residueWriteFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n  version: "+version+"\n")
	return root
}

// residueSnapshotTree returns the slash-relative paths of every regular file
// under root, sorted. Used to assert the residue path performs no wipe and no
// forced redeploy (REQ-RIL2-020).
func residueSnapshotTree(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(out)
	return out
}

// AC-RIL2-010 (REQ-RIL2-018, REQ-RIL2-020) — a v3-confirmed project carrying
// residue is cleaned WITHOUT entering the destructive full reinstall: no
// PRESERVE snapshot directory (.moai/backups/v2-to-v3-*), no tree wipe, and no
// forced template redeployment. The residue path accepts no template deployer at
// all, so "force-deploy was not invoked" is asserted through its only observable
// consequence: not one template file appears in the tree.
func TestResidueCleanup_NoDestructiveReinstall(t *testing.T) {
	root := newV3ConfirmedProject(t, "3.0.1")
	residueWriteFile(t, root, residueDeprecatedRel, "legacy planner\n")
	// A user-authored file that a tree wipe or forced redeploy would disturb.
	residueWriteFile(t, root, ".moai/project/product.md", "user content\n")

	before := residueSnapshotTree(t, root)

	var out bytes.Buffer
	result, err := runV3ResidueCleanup(root, false, false, &out)
	if err != nil {
		t.Fatalf("runV3ResidueCleanup: %v", err)
	}

	// (a) the residue is nonetheless gone
	if _, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(residueDeprecatedRel))); !os.IsNotExist(statErr) {
		t.Fatalf("residue %s still present after cleanup (stat err: %v)", residueDeprecatedRel, statErr)
	}
	if len(result.Removed) != 1 || result.Removed[0] != residueDeprecatedRel {
		t.Fatalf("Removed = %v, want exactly [%s]", result.Removed, residueDeprecatedRel)
	}

	// (b) no PRESERVE snapshot directory was created
	if entries, readErr := os.ReadDir(filepath.Join(root, ".moai", "backups")); readErr == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "v2-to-v3-") {
				t.Fatalf("clean-reinstall PRESERVE snapshot created: .moai/backups/%s", e.Name())
			}
		}
	}

	// (c) no tree wipe and no forced redeploy: every pre-existing file except the
	// residue survives, and the only additions live under the residue path's own
	// backup directory.
	after := residueSnapshotTree(t, root)
	afterSet := make(map[string]bool, len(after))
	for _, p := range after {
		afterSet[p] = true
	}
	for _, p := range before {
		if p == residueDeprecatedRel {
			continue
		}
		if !afterSet[p] {
			t.Errorf("pre-existing file disappeared (tree wipe): %s", p)
		}
	}
	beforeSet := make(map[string]bool, len(before))
	for _, p := range before {
		beforeSet[p] = true
	}
	for _, p := range after {
		if beforeSet[p] {
			continue
		}
		if strings.HasPrefix(p, ".moai/backup/") {
			continue // the residue path's own backup — expected
		}
		t.Errorf("unexpected file added (forced redeploy?): %s", p)
	}
}

// AC-RIL2-011 (REQ-RIL2-019, REQ-RIL2-023) — the deprecated path is removed and
// a backup copy exists; the same run against a directory lacking the project
// marker performs no removal.
func TestResidueCleanup_RemovesDeprecatedOnV3Project(t *testing.T) {
	t.Run("v3_project_with_marker", func(t *testing.T) {
		root := newV3ConfirmedProject(t, "3.0.1")
		residueWriteFile(t, root, residueDeprecatedRel, "legacy planner\n")

		var out bytes.Buffer
		result, err := runV3ResidueCleanup(root, false, false, &out)
		if err != nil {
			t.Fatalf("runV3ResidueCleanup: %v", err)
		}
		if result.Skipped {
			t.Fatalf("Skipped = true on a genuine MoAI project")
		}
		if _, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(residueDeprecatedRel))); !os.IsNotExist(statErr) {
			t.Errorf("residue %s not removed", residueDeprecatedRel)
		}
		if result.BackupDir == "" {
			t.Fatalf("BackupDir empty — nothing was backed up before removal")
		}
		backupCopy := filepath.Join(result.BackupDir, filepath.FromSlash(residueDeprecatedRel))
		data, readErr := os.ReadFile(backupCopy)
		if readErr != nil {
			t.Fatalf("backup copy missing at %s: %v", backupCopy, readErr)
		}
		if string(data) != "legacy planner\n" {
			t.Errorf("backup copy content = %q, want %q", string(data), "legacy planner\n")
		}
	})

	t.Run("no_project_marker", func(t *testing.T) {
		// Same residue, but NO .moai/config/sections/system.yaml.
		root := t.TempDir()
		abs := residueWriteFile(t, root, residueDeprecatedRel, "legacy planner\n")

		var out bytes.Buffer
		result, err := runV3ResidueCleanup(root, false, false, &out)
		if err != nil {
			t.Fatalf("runV3ResidueCleanup: %v", err)
		}
		if !result.Skipped {
			t.Errorf("Skipped = false on a directory lacking the project marker")
		}
		if len(result.Removed) != 0 {
			t.Errorf("Removed = %v, want empty on a non-MoAI directory", result.Removed)
		}
		if _, statErr := os.Lstat(abs); statErr != nil {
			t.Errorf("residue removed on a directory lacking the project marker: %v", statErr)
		}
	})
}

// AC-RIL2-012 (REQ-RIL2-022) — the direct anti-loop assertion for issue #1243:
// run 1 removes at least one path, run 2 removes exactly zero.
func TestResidueCleanup_IdempotentAcrossTwoRuns(t *testing.T) {
	root := newV3ConfirmedProject(t, "3.0.1")
	residueWriteFile(t, root, residueDeprecatedRel, "legacy planner\n")

	var out bytes.Buffer
	first, err := runV3ResidueCleanup(root, false, false, &out)
	if err != nil {
		t.Fatalf("run 1: %v", err)
	}
	if len(first.Removed) < 1 {
		t.Fatalf("run 1 removed %d paths, want >= 1", len(first.Removed))
	}

	second, err := runV3ResidueCleanup(root, false, false, &out)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if len(second.Removed) != 0 {
		t.Fatalf("run 2 removed %v, want exactly 0 — the removal is not converging", second.Removed)
	}
}

// AC-RIL2-019 (REQ-RIL2-021) — the .agency/ migration pre-step still fires
// alongside the residue sweep. Both happen in the same run: the migration's
// archive destination exists on disk AND the deprecated path is removed.
//
// The migration copies .agency/ to .agency.archived/ and leaves both in place.
// Because .agency.archived is itself a defs.DeprecatedPaths entry, a sweep that
// scanned AFTER the migration would delete the archive the migration just wrote.
// The sweep therefore operates on the pre-migration residue snapshot.
func TestResidueCleanup_PreservesAgencyMigrationPreStep(t *testing.T) {
	root := newV3ConfirmedProject(t, "3.0.1")
	residueWriteFile(t, root, ".agency/learnings/note.md", "a learning\n")
	residueWriteFile(t, root, residueDeprecatedRel, "legacy planner\n")

	var out bytes.Buffer
	result, err := runV3ResidueCleanup(root, false, false, &out)
	if err != nil {
		t.Fatalf("runV3ResidueCleanup: %v", err)
	}

	// (a) the migration pre-step fired — its archive destination exists
	archive := filepath.Join(root, ".agency.archived")
	if info, statErr := os.Stat(archive); statErr != nil || !info.IsDir() {
		t.Fatalf(".agency/ migration pre-step did not fire: .agency.archived missing (%v)", statErr)
	}

	// (b) the deprecated path was removed by the residue sweep in the same run
	if _, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(residueDeprecatedRel))); !os.IsNotExist(statErr) {
		t.Errorf("residue %s not removed alongside the migration", residueDeprecatedRel)
	}
	found := false
	for _, rel := range result.Removed {
		if rel == residueDeprecatedRel {
			found = true
		}
		if rel == ".agency.archived" {
			t.Errorf("sweep removed the migration's own archive destination .agency.archived")
		}
	}
	if !found {
		t.Errorf("Removed = %v, want it to contain %s", result.Removed, residueDeprecatedRel)
	}
}
