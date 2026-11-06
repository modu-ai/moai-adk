package config

// loader_backup_budget.go — the single-key reader for state.backup_past_versions_keep
// (SPEC-UPDATE-BACKUP-RETENTION-001 REQ-UBR-007), modelled on LoadWorktreeBaseBranch:
// the update flow prunes backups outside the Loader's lifecycle and holds no
// loaded *Config at the point it needs the budget.

import (
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadBackupPastVersionsKeep returns the per-file budget of past versions the
// update's backup pruning protects for the project rooted at projectRoot.
//
// Every failure path, and every value below 1, yields
// DefaultBackupPastVersionsKeep: a project that cannot be read behaves as one
// that never set the key, and a configured 0 or negative number cannot switch the
// protection of REQ-UBR-006 off. The file is read without blocking on a FIFO, as
// the sibling single-key readers do.
func LoadBackupPastVersionsKeep(projectRoot string) int {
	data, err := readNonBlockingRegularFile(filepath.Join(projectRoot, ".moai", "config", "sections", "state.yaml"))
	if err != nil {
		return DefaultBackupPastVersionsKeep
	}
	var wrapper stateFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return DefaultBackupPastVersionsKeep
	}
	if wrapper.State.BackupPastVersionsKeep < 1 {
		return DefaultBackupPastVersionsKeep
	}
	return wrapper.State.BackupPastVersionsKeep
}
