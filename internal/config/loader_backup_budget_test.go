package config

// loader_backup_budget_test.go — AC-UBR-10 reader half (SPEC-UPDATE-BACKUP-RETENTION-001,
// REQ-UBR-007): the update flow reads the budget through LoadBackupPastVersionsKeep,
// and every unreadable or below-1 value reads as the one declared default.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBackupPastVersionsKeep(t *testing.T) {
	read := func(t *testing.T, body string, write bool) int {
		t.Helper()
		root := t.TempDir()
		if write {
			sections := filepath.Join(root, ".moai", "config", "sections")
			if err := os.MkdirAll(sections, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(sections, "state.yaml"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return LoadBackupPastVersionsKeep(root)
	}

	cases := []struct {
		name  string
		body  string
		write bool
		want  int
	}{
		{"missing file reads the default", "", false, DefaultBackupPastVersionsKeep},
		{"omitted key reads the default", "state: {}\n", true, DefaultBackupPastVersionsKeep},
		{"explicit value is used", "state:\n  backup_past_versions_keep: 5\n", true, 5},
		{"zero reads the default", "state:\n  backup_past_versions_keep: 0\n", true, DefaultBackupPastVersionsKeep},
		{"negative reads the default", "state:\n  backup_past_versions_keep: -4\n", true, DefaultBackupPastVersionsKeep},
		{"unparseable file reads the default", "state: [unterminated\n", true, DefaultBackupPastVersionsKeep},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := read(t, tc.body, tc.write); got != tc.want {
				t.Fatalf("LoadBackupPastVersionsKeep = %d, want %d", got, tc.want)
			}
		})
	}
}
