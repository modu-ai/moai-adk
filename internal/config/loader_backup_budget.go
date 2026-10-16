package config

// loader_backup_budget.go — the single-key reader for state.backup_past_versions_keep
// (SPEC-UPDATE-BACKUP-RETENTION-001 REQ-UBR-007), modelled on LoadWorktreeBaseBranch:
// the update flow prunes backups outside the Loader's lifecycle and holds no
// loaded *Config at the point it needs the budget.

// LoadBackupPastVersionsKeep returns the per-file budget of past versions the
// update's backup pruning protects for the project rooted at projectRoot.
func LoadBackupPastVersionsKeep(projectRoot string) int {
	return 0
}
