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

const (
	priorManifestTmpLiteral = "manifest.prior.json.tmp-4242"
	priorManifestTmpDir     = "manifest.prior.json.tmp-notes"
)

func priorManifestTmpFixture(t *testing.T, withSections bool) (projectRoot, backupDir, configDir string) {
	t.Helper()
	projectRoot, backupDir, configDir = priorManifestRestoreFixture(t, withSections)
	for rel, content := range map[string]string{
		priorManifestTmpLiteral:            `{"version":"v0","files":{}}` + "\n",
		"notes/" + priorManifestTmpLiteral: "user content in a subdirectory\n",
		// User directories that merely START with a bookkeeping name: everything below
		// them is user content (card-review round 5, finding N1).
		priorManifestTmpDir + "/keep.txt":                "below a directory named like the temporary prefix\n",
		priorManifestLiteral + ".d/x":                    "below a directory named like the final name plus a suffix\n",
		priorManifestTmpDir + "/" + priorManifestLiteral: "the exact bookkeeping name below such a directory\n",
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
	for rel, want := range map[string]string{
		priorManifestTmpDir + "/keep.txt":                "below a directory named like the temporary prefix\n",
		priorManifestLiteral + ".d/x":                    "below a directory named like the final name plus a suffix\n",
		priorManifestTmpDir + "/" + priorManifestLiteral: "the exact bookkeeping name below such a directory\n",
	} {
		if got, err := os.ReadFile(filepath.Join(configDir, filepath.FromSlash(rel))); err != nil || string(got) != want {
			t.Errorf("%s was not restored as user content (err=%v, got %q)", rel, err, got)
		}
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

// IsPriorManifestBookkeeping is the one predicate every restore walk asks. It
// names ONLY a root-level entry: the final name, or a name that starts with the
// temporary prefix. Anything with a path separator is below the root.
func TestIsPriorManifestBookkeeping(t *testing.T) {
	join := filepath.Join
	for _, c := range []struct {
		name string
		rel  string
		want bool
	}{
		{"the final name at the root", "manifest.prior.json", true},
		{"a temporary file at the root", "manifest.prior.json.tmp-4242", true},
		{"a temporary file with a long random suffix", "manifest.prior.json.tmp-3588344372", true},
		{"the empty path", "", false},
		{"the prefix alone is a root-level name", "manifest.prior.json.tmp-", true},
		{"a file below a directory named like the temporary prefix", join("manifest.prior.json.tmp-notes", "keep.txt"), false},
		{"a directory named like the temporary prefix", "manifest.prior.json.tmp-notes" + string(filepath.Separator), false},
		{"the final name below a prefixed directory", join("manifest.prior.json.tmp-notes", "manifest.prior.json"), false},
		{"a directory named like the final name plus a suffix", join("manifest.prior.json.d", "x"), false},
		{"the final name in a subdirectory", join("notes", "manifest.prior.json"), false},
		{"a temporary name in a subdirectory", join("notes", "manifest.prior.json.tmp-4242"), false},
		{"a name that merely ends with the final name", "xmanifest.prior.json", false},
		{"a name that merely extends the final name", "manifest.prior.jsonx", false},
		{"the temporary prefix without its dash", "manifest.prior.json.tmp", false},
		{"another letter case is another name", "Manifest.Prior.json", false},
	} {
		if got := IsPriorManifestBookkeeping(c.rel); got != c.want {
			t.Errorf("%s: IsPriorManifestBookkeeping(%q) = %v, want %v", c.name, c.rel, got, c.want)
		}
	}
}
