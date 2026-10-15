package backup

// retention_select.go — the selection step of backup retention
// (SPEC-UPDATE-BACKUP-RETENTION-001, design D2). It decides which excess runs a
// prune retains and which it removes, as a pure function of a census: it takes no
// filesystem handle, so the budget and cover rules are table-testable.

// Retention report vocabulary (REQ-UBR-011, design D8). A prune reports each run
// it removes and each excess run it retains; a retained run names its reason.
const (
	// RetentionActionRemoved and RetentionActionRetained are the two outcomes a
	// prune reports for an excess run.
	RetentionActionRemoved  = "removed"
	RetentionActionRetained = "retained"

	// RetentionReasonProtectsVersions: the run holds a version the per-file budget
	// protects and no other retained run holds it.
	RetentionReasonProtectsVersions = "protects-versions"
	// RetentionReasonUnclassifiable: the run holds an entry the pruner cannot read
	// or classify (REQ-UBR-008).
	RetentionReasonUnclassifiable = "unclassifiable-entry"
	// RetentionReasonRemovalFailed: the selection removed the run but the removal
	// failed; the pruner assigns this reason after the fact (design D2 step 5).
	RetentionReasonRemovalFailed = "removal-failed"
)

// versionKey identifies one backed-up file version: the file's path inside its
// backup directory together with the SHA-256 of its bytes (REQ-UBR-002).
type versionKey struct {
	path string
	sum  string
}

// censusRun is the census view of one backup run. id is opaque to the selection.
type censusRun struct {
	id string
	// keys are the regular-file versions the run holds, bookkeeping and the
	// .template-defaults render excluded.
	keys []versionKey
	// exempt marks a run holding an entry that cannot be read or classified,
	// including an entry that is not a regular file (REQ-UBR-008).
	exempt bool
	// exemptDetail names the first such entry, relative to its backup directory.
	exemptDetail string
}

// retainedRun is one excess run the selection keeps.
type retainedRun struct {
	id string
	// reason is one of the RetentionReason constants.
	reason string
	// protects counts the protected versions the run holds (reason
	// RetentionReasonProtectsVersions).
	protects int
	// detail names the unclassifiable entry (reason RetentionReasonUnclassifiable).
	detail string
}

// retentionPlan is the outcome of the selection for the excess runs, in the
// order they were given (newest first).
type retentionPlan struct {
	remove []string
	retain []retainedRun
	// protected is the number of distinct protected versions: the bound on the
	// number of non-exempt retained excess runs (REQ-UBR-010).
	protected int
}

// selectRetention returns the plan for runs, given newest first, with keepCount
// runs in the window and pastKeep past versions protected per file.
func selectRetention(runs []censusRun, keepCount, pastKeep int) retentionPlan {
	return retentionPlan{}
}
