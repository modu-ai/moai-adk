package backup

// restore_prior_manifest_tmp_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, card-review round 4, finding B): the temporary file the prior-manifest
// writer creates before its rename (`manifest.prior.json.tmp-*`) is backup
// bookkeeping exactly like the final name. A failed write used to leave it at the
// backup root, and the legacy restore walk copied it into .moai/config. The
// walks must skip it at the backup root, as they skip the final name; a
// same-named file in a subdirectory is user content and is still restored.
//
// The name is spelled out as a literal, independent of the constant the restore
// code uses.

import (
	"os"
	"path/filepath"
	"testing"
)

const priorManifestTmpLiteral = "manifest.prior.json.tmp-4242"

func priorManifestTmpFixture(t *testing.T, withSections bool) (projectRoot, backupDir, configDir string) {
	t.Helper()
	projectRoot, backupDir, configDir = priorManifestRestoreFixture(t, withSections)
	for rel, content := range map[string]string{
		priorManifestTmpLiteral:            `{"version":"v0","files":{}}` + "\n",
		"notes/" + priorManifestTmpLiteral: "user content in a subdirectory\n",
	} {
		p := filepath.Join(backupDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return projectRoot, backupDir, configDir
}

func assertPriorManifestTmpNotRestored(t *testing.T, configDir, reachRel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(configDir, filepath.FromSlash(reachRel))); err != nil {
		t.Fatalf("control: the backed-up %s was not restored, the restore did not reach its walk: %v", reachRel, err)
	}
	if _, err := os.Stat(filepath.Join(configDir, priorManifestTmpLiteral)); !os.IsNotExist(err) {
		t.Errorf("%s from the backup root was copied into the config tree (stat err: %v)", priorManifestTmpLiteral, err)
	}
	if got, err := os.ReadFile(filepath.Join(configDir, "notes", priorManifestTmpLiteral)); err != nil || string(got) != "user content in a subdirectory\n" {
		t.Errorf("a same-named file in a subdirectory was not restored as user content (err=%v)", err)
	}
}

func TestRestoreLegacySkipsPriorManifestTemporaryFile(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestTmpFixture(t, false)
	if err := RestoreMoaiConfigLegacy(projectRoot, backupDir, configDir); err != nil {
		t.Fatalf("RestoreMoaiConfigLegacy: %v", err)
	}
	assertPriorManifestTmpNotRestored(t, configDir, "system.yaml")
}

func TestRestoreRetainedWithoutSectionsSkipsPriorManifestTemporaryFile(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestTmpFixture(t, false)
	if _, err := RestoreMoaiConfigRetained(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}
	assertPriorManifestTmpNotRestored(t, configDir, "system.yaml")
}

func TestRestoreRetainedWithSectionsSkipsPriorManifestTemporaryFile(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestTmpFixture(t, true)
	if _, err := RestoreMoaiConfigRetained(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}
	if _, err := os.Stat(filepath.Join(configDir, "sections", "system.yaml")); err != nil {
		t.Fatalf("control: the backed-up section was not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(configDir, priorManifestTmpLiteral)); !os.IsNotExist(err) {
		t.Errorf("%s was copied into the config tree (stat err: %v)", priorManifestTmpLiteral, err)
	}
}

func TestRestoreFromBackupDirSkipsPriorManifestTemporaryFile(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestTmpFixture(t, false)
	if err := RestoreFromBackupDir(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreFromBackupDir: %v", err)
	}
	assertPriorManifestTmpNotRestored(t, configDir, "system.yaml")
}
