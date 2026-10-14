package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

// The per-file budget of superseded backup versions ships as three. The default in
// code and the value the shipped state.yaml template declares must agree, so a
// project that never edits its state.yaml prunes with the documented default.
func TestBackupPastVersionsKeep_DefaultAndTemplateAgree(t *testing.T) {
	if DefaultBackupPastVersionsKeep != 3 {
		t.Fatalf("DefaultBackupPastVersionsKeep = %d, want 3", DefaultBackupPastVersionsKeep)
	}
	if got := NewDefaultConfig().State.BackupPastVersionsKeep; got != DefaultBackupPastVersionsKeep {
		t.Fatalf("NewDefaultConfig().State.BackupPastVersionsKeep = %d, want %d", got, DefaultBackupPastVersionsKeep)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	templatePath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "state.yaml")
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read shipped state.yaml: %v", err)
	}
	var shipped struct {
		State struct {
			BackupPastVersionsKeep int `yaml:"backup_past_versions_keep"`
		} `yaml:"state"`
	}
	if err := yaml.Unmarshal(raw, &shipped); err != nil {
		t.Fatalf("parse shipped state.yaml: %v", err)
	}
	if shipped.State.BackupPastVersionsKeep != DefaultBackupPastVersionsKeep {
		t.Fatalf("shipped state.yaml backup_past_versions_keep = %d, want %d",
			shipped.State.BackupPastVersionsKeep, DefaultBackupPastVersionsKeep)
	}
}
