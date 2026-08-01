package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// newRestoreFixture builds a project tree with a real .moai/config, takes a
// real update backup of it, and returns (projectRoot, backupDir).
func newRestoreFixture(t *testing.T) (string, string) {
	t.Helper()

	projectRoot := t.TempDir()
	writeFixtureFile(t, projectRoot, ".moai/config/sections/system.yaml",
		"moai:\n  version: 3.0.0\n  template_version: 3.0.0\n")
	writeFixtureFile(t, projectRoot, ".moai/config/sections/user.yaml",
		"user:\n  name: fixture-operator\n")

	backupDir, err := backup.BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	if backupDir == "" {
		t.Fatal("BackupMoaiConfig returned an empty backup dir; fixture is wrong")
	}
	return projectRoot, backupDir
}

// hashTree returns a content hash over every regular file under root, so two
// invocations can be compared for byte-level tree equality.
func hashTree(t *testing.T, root string) string {
	t.Helper()

	var entries []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		sum := sha256.Sum256(data)
		entries = append(entries, filepath.ToSlash(rel)+":"+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatalf("hash tree %s: %v", root, err)
	}
	sort.Strings(entries)
	total := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(total[:])
}

// TestRestore_ProceedsWithoutProjectMarker verifies AC-UDS-002 (REQ-UDS-021,
// REQ-UDS-022): the restore entry point runs on a tree whose project marker
// was destroyed — that absence is the damage it exists to repair.
func TestRestore_ProceedsWithoutProjectMarker(t *testing.T) {
	projectRoot, backupDir := newRestoreFixture(t)

	marker := filepath.Join(projectRoot, ".moai", "config", "sections", "system.yaml")
	if err := os.Remove(marker); err != nil {
		t.Fatalf("remove project marker: %v", err)
	}
	if checkProjectMarker(projectRoot) == nil {
		t.Fatal("fixture precondition: the marker gate must reject this tree")
	}

	if err := runUpdateRestore(projectRoot, backupDir, io.Discard); err != nil {
		t.Fatalf("restore entry point refused a marker-less tree: %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("system.yaml was not restored: %v", err)
	}
	if checkProjectMarker(projectRoot) != nil {
		t.Error("the restored tree is still not recognised as a moai project")
	}
}

// TestUpdate_RejectsTreeWithoutProjectMarker verifies AC-UDS-003
// (REQ-UDS-025): the ordinary update path still refuses a marker-less tree.
// The bypass is scoped to the restore entry point alone.
func TestUpdate_RejectsTreeWithoutProjectMarker(t *testing.T) {
	bare := t.TempDir()

	err := checkProjectMarker(bare)
	if err == nil {
		t.Fatal("the ordinary update path accepted a tree with no project marker")
	}
	if !strings.Contains(err.Error(), "not a moai project") {
		t.Errorf("gate error does not carry the expected message: %v", err)
	}

	// A tree that DOES carry the marker passes, so the gate is not a constant.
	ok := t.TempDir()
	writeFixtureFile(t, ok, ".moai/config/sections/system.yaml", "moai:\n  version: 3.0.0\n")
	if err := checkProjectMarker(ok); err != nil {
		t.Errorf("the gate rejected a valid project tree: %v", err)
	}
}

// TestRestore_IdempotentAndRefusesForeignDir verifies AC-UDS-004
// (REQ-UDS-023, REQ-UDS-024).
func TestRestore_IdempotentAndRefusesForeignDir(t *testing.T) {
	projectRoot, backupDir := newRestoreFixture(t)

	configDir := filepath.Join(projectRoot, ".moai", "config")
	if err := os.RemoveAll(configDir); err != nil {
		t.Fatalf("wipe config dir: %v", err)
	}

	if err := runUpdateRestore(projectRoot, backupDir, io.Discard); err != nil {
		t.Fatalf("first restore: %v", err)
	}
	afterFirst := hashTree(t, configDir)

	if err := runUpdateRestore(projectRoot, backupDir, io.Discard); err != nil {
		t.Fatalf("second restore: %v", err)
	}
	afterSecond := hashTree(t, configDir)

	if afterFirst != afterSecond {
		t.Errorf("restore is not idempotent: tree hash %s after one apply, %s after two",
			afterFirst, afterSecond)
	}

	// REQ-UDS-024: a directory that is not a recognisable update backup is
	// refused, and the tree is left untouched.
	foreign := t.TempDir()
	writeFixtureFile(t, foreign, "sections/system.yaml", "moai:\n  version: 9.9.9\n")

	before := hashTree(t, configDir)
	err := runUpdateRestore(projectRoot, foreign, io.Discard)
	if err == nil {
		t.Fatal("restore accepted a directory that is not an update backup")
	}
	if !strings.Contains(err.Error(), backup.BackupMarkerFile) {
		t.Errorf("refusal does not name the missing marker file: %v", err)
	}
	if after := hashTree(t, configDir); after != before {
		t.Error("the refused restore modified the project tree")
	}
}
