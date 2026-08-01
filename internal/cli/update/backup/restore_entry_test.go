package backup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// newBackupFixture builds a project tree with a real .moai/config and takes a
// real backup of it.
func newBackupFixture(t *testing.T) (projectRoot, backupDir string) {
	t.Helper()

	projectRoot = t.TempDir()
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "system.yaml"),
		[]byte("moai:\n  version: 3.0.0\n"), 0o644); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}

	backupDir, err := BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	if backupDir == "" {
		t.Fatal("BackupMoaiConfig returned an empty backup dir")
	}
	return projectRoot, backupDir
}

func TestRestoreFromBackupDir_RestoresDestroyedMarker(t *testing.T) {
	projectRoot, backupDir := newBackupFixture(t)

	marker := filepath.Join(projectRoot, ".moai", "config", "sections", "system.yaml")
	if err := os.Remove(marker); err != nil {
		t.Fatalf("remove marker: %v", err)
	}

	if err := RestoreFromBackupDir(projectRoot, backupDir, nil); err != nil {
		t.Fatalf("RestoreFromBackupDir: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("system.yaml was not restored: %v", err)
	}
}

func TestRestoreFromBackupDir_RefusesInvalidInput(t *testing.T) {
	projectRoot, backupDir := newBackupFixture(t)

	foreignDir := t.TempDir()

	notADir := filepath.Join(t.TempDir(), "plain.txt")
	if err := os.WriteFile(notADir, []byte("x"), 0o644); err != nil {
		t.Fatalf("write plain file: %v", err)
	}

	tests := []struct {
		name        string
		projectRoot string
		backupDir   string
		wantMarker  bool
	}{
		{name: "empty project root", projectRoot: "", backupDir: backupDir},
		{name: "empty backup dir", projectRoot: projectRoot, backupDir: ""},
		{name: "missing backup dir", projectRoot: projectRoot, backupDir: filepath.Join(foreignDir, "absent")},
		{name: "backup path is a file", projectRoot: projectRoot, backupDir: notADir, wantMarker: true},
		{name: "directory without marker file", projectRoot: projectRoot, backupDir: foreignDir, wantMarker: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RestoreFromBackupDir(tt.projectRoot, tt.backupDir, nil)
			if err == nil {
				t.Fatal("expected a refusal, got nil")
			}
			if tt.wantMarker && !errors.Is(err, ErrNotAnUpdateBackup) {
				t.Errorf("error does not wrap ErrNotAnUpdateBackup: %v", err)
			}
		})
	}
}
