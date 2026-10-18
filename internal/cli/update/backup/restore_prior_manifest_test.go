package backup

// restore_prior_manifest_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card t1865,
// card-review finding 1): manifest.prior.json is run-backup bookkeeping — the
// project manifest as the run loaded it, written at the root of the run backup
// directory by the Backup stage. Like the run marker it must never be copied
// into the project by a restore. The restore walk over the backup root
// (RestoreMoaiConfigLegacy) used to copy every file it did not know.
//
// The file name is spelled out here on purpose: the tests pin the name that is
// really written on disk, independent of the constant the restore code uses.

import (
	"os"
	"path/filepath"
	"testing"
)

const priorManifestLiteral = "manifest.prior.json"

// priorManifestRestoreFixture builds a project with an empty config dir and a
// run backup that holds one real config file (the reach control), the backup
// marker, manifest.prior.json at its root, and a same-named file in a
// subdirectory (user content, which must still be restored).
func priorManifestRestoreFixture(t *testing.T, withSections bool) (projectRoot, backupDir, configDir string) {
	t.Helper()
	projectRoot = t.TempDir()
	backupDir = t.TempDir()
	configDir = filepath.Join(projectRoot, ".moai", "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(rel, content string) {
		p := filepath.Join(backupDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(BackupMarkerFile, `{"timestamp":"2026-10-18T00:00:00Z","backup_type":"config"}`)
	write(priorManifestLiteral, `{"version":"v0","files":{}}`+"\n")
	write("notes/"+priorManifestLiteral, "user content in a subdirectory\n")
	if withSections {
		write("sections/system.yaml", "version: 1\n")
	} else {
		write("system.yaml", "version: 1\n")
	}
	return projectRoot, backupDir, configDir
}

func assertPriorManifestNotRestored(t *testing.T, configDir string, reachRel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(configDir, filepath.FromSlash(reachRel))); err != nil {
		t.Fatalf("control: the backed-up %s was not restored, the restore did not reach its walk: %v", reachRel, err)
	}
	if _, err := os.Stat(filepath.Join(configDir, priorManifestLiteral)); !os.IsNotExist(err) {
		t.Errorf("%s from the backup root was copied into the config tree (stat err: %v)", priorManifestLiteral, err)
	}
	// A same-named file below the root is not the bookkeeping file.
	if got, err := os.ReadFile(filepath.Join(configDir, "notes", priorManifestLiteral)); err != nil || string(got) != "user content in a subdirectory\n" {
		t.Errorf("a same-named file in a subdirectory was not restored as user content (err=%v)", err)
	}
}

// The Legacy walk (a backup without sections/).
func TestRestoreLegacySkipsPriorManifest(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestRestoreFixture(t, false)
	if err := RestoreMoaiConfigLegacy(projectRoot, backupDir, configDir); err != nil {
		t.Fatalf("RestoreMoaiConfigLegacy: %v", err)
	}
	assertPriorManifestNotRestored(t, configDir, "system.yaml")
}

// RestoreMoaiConfigRetained on a backup without sections/ falls through to the
// Legacy walk.
func TestRestoreRetainedWithoutSectionsSkipsPriorManifest(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestRestoreFixture(t, false)
	if _, err := RestoreMoaiConfigRetained(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}
	assertPriorManifestNotRestored(t, configDir, "system.yaml")
}

// RestoreMoaiConfigRetained on a backup with sections/ walks sections/ only and
// takes YAML files only; the file at the backup root must stay out.
func TestRestoreRetainedWithSectionsSkipsPriorManifest(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestRestoreFixture(t, true)
	if _, err := RestoreMoaiConfigRetained(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}
	if _, err := os.Stat(filepath.Join(configDir, "sections", "system.yaml")); err != nil {
		t.Fatalf("control: the backed-up section was not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(configDir, priorManifestLiteral)); !os.IsNotExist(err) {
		t.Errorf("%s was copied into the config tree (stat err: %v)", priorManifestLiteral, err)
	}
}

// The entry point the --restore command uses (marker check, two passes).
func TestRestoreFromBackupDirSkipsPriorManifest(t *testing.T) {
	projectRoot, backupDir, configDir := priorManifestRestoreFixture(t, false)
	if err := RestoreFromBackupDir(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreFromBackupDir: %v", err)
	}
	assertPriorManifestNotRestored(t, configDir, "system.yaml")
}
