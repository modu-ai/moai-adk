package config

// state_backup_past_keep_test.go — AC-UBR-10 (SPEC-UPDATE-BACKUP-RETENTION-001,
// REQ-UBR-007), struct half: the budget default is carried by StateConfig, an
// omitted key retains it through the loader, and an explicit value replaces it.
// The below-1 fallback is the pruner's rule and is pinned in the backup package.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultStateConfigCarriesBackupPastVersionsKeep(t *testing.T) {
	if DefaultBackupPastVersionsKeep != 3 {
		t.Fatalf("DefaultBackupPastVersionsKeep = %d, want the shipped default 3", DefaultBackupPastVersionsKeep)
	}
	cfg := NewDefaultStateConfig()
	if cfg.BackupPastVersionsKeep != DefaultBackupPastVersionsKeep {
		t.Fatalf("default backup_past_versions_keep = %d, want %d", cfg.BackupPastVersionsKeep, DefaultBackupPastVersionsKeep)
	}
}

func TestStateBackupPastVersionsKeepLoadsFromYAML(t *testing.T) {
	load := func(t *testing.T, body string) int {
		t.Helper()
		root := t.TempDir()
		sections := filepath.Join(root, "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sections, "state.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := NewLoader().Load(root)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		return cfg.State.BackupPastVersionsKeep
	}

	t.Run("omitted key retains default", func(t *testing.T) {
		if got := load(t, "state: {}\n"); got != DefaultBackupPastVersionsKeep {
			t.Fatalf("loaded backup_past_versions_keep = %d, want default %d", got, DefaultBackupPastVersionsKeep)
		}
	})
	t.Run("explicit value replaces default", func(t *testing.T) {
		if got := load(t, "state:\n  backup_past_versions_keep: 5\n"); got != 5 {
			t.Fatalf("loaded backup_past_versions_keep = %d, want 5", got)
		}
	})
}
