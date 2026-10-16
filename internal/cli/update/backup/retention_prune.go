package backup

// retention_prune.go — the pruner of SPEC-UPDATE-BACKUP-RETENTION-001 (design D1):
// containment check, run collection, census, selection, removal, report. The
// selection itself lives in retention_select.go.

// PruneItem is the report row of one excess run a prune decided on: the run's
// directories, what happened to it, and, for a retained run, why.
type PruneItem struct {
	// Run is the path of the run's first directory, the name the report uses.
	Run string
	// Dirs are the run's directories across both backup roots, ascending.
	Dirs []string
	// Action is RetentionActionRemoved or RetentionActionRetained.
	Action string
	// Reason is one of the RetentionReason constants for a retained run.
	Reason string
	// Protects counts the protected versions a run retained for them holds.
	Protects int
	// Detail names the unclassifiable entry, or the removal error.
	Detail string
}

// UnmanagedEntries summarizes the entries under the config backup root whose
// names no backup writer produces (REQ-UBR-012). They are reported, never deleted.
type UnmanagedEntries struct {
	Count int
	Bytes int64
}

// PruneReport is the outcome of one PruneBackups call.
type PruneReport struct {
	// Runs is the number of runs the pruner counted across both roots.
	Runs int
	// Window is the number of runs kept whole as the retention window.
	Window int
	// Items has one row per excess run, newest first.
	Items []PruneItem
	// Removed counts the backup directories the prune removed.
	Removed int
	// Protected is the number of distinct protected versions (REQ-UBR-010).
	Protected int
	// RootRefused is true when a backup root is a symbolic link or leads outside
	// the project; nothing was pruned (REQ-UBR-013).
	RootRefused bool
	// Unmanaged summarizes the entries no backup writer names.
	Unmanaged UnmanagedEntries
}

// PruneBackups prunes old backup runs with the budgeted retention of
// SPEC-UPDATE-BACKUP-RETENTION-001 and reports every decision.
func PruneBackups(projectRoot string, keepCount, pastKeep int) PruneReport {
	return PruneReport{}
}
