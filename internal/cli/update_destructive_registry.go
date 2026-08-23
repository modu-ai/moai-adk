package cli

// Destructive-target registry for the update subsystem
// (SPEC-UPDATE-DATA-SURVIVAL-001 M2; REQ-UDS-006, REQ-UDS-007, REQ-UDS-010).
//
// Every call site in the update subsystem that performs an irreversible
// filesystem operation — os.RemoveAll or os.Rename — is recorded here together
// with either the protection that stands between it and user data loss, or the
// reason it needs none.
//
// The registry is validated by TestDestructiveTargetRegistry_CoversAllSites,
// which enumerates the sites by statically parsing the Go source rather than by
// reading this table. That independence is the whole point (REQ-UDS-007): a
// guard that derived both sides of the comparison from this file would compare
// the registry against itself and pass forever, so a newly added destructive
// site would land unprotected and unnoticed.

// destructiveSite records one (file, enclosing function) pair that performs
// destructive filesystem operations, and how many such call sites it contains.
//
// The key is deliberately (File, Function, Sites) and NEVER a line number.
// Coordinates shift whenever anything above a site is edited — this SPEC's own
// M1 commit moved two of the rows below — so a line-keyed registry would report
// drift after every unrelated change while saying nothing about whether the set
// of destructive sites actually changed.
type destructiveSite struct {
	// File is the module-root-relative, slash-separated source path.
	File string
	// Function is the name of the enclosing function declaration.
	Function string
	// Sites is the number of os.RemoveAll / os.Rename call sites inside it.
	Sites int

	// Protection names what stands between this site and user data loss.
	// Exactly one of Protection and Exemption is non-empty.
	Protection string
	// Exemption records why this site destroys nothing that needs protecting.
	// Exactly one of Protection and Exemption is non-empty.
	Exemption string
}

// key identifies a registry row independently of its count and assignment.
func (s destructiveSite) key() string { return s.File + " " + s.Function }

// destructiveTargetRegistry is the registry itself: 12 (file, function) rows
// covering 22 call sites, matching the source scan the drift guard performs.
//
// Rows carry either a Protection or an Exemption, never both and never neither.
// The exempt rows rest on two materially different grounds — same-call rewind
// (BackupMoaiConfig, backupUserOwnedNamespace) versus retention pruning of
// moai-authored backup directories (CleanupOldBackups) — and must not be
// collapsed into one reason: the pruning row destroys restore points from
// PREVIOUS runs, so it is exempt from the user-data protection set without
// being harmless to the recovery contract.
var destructiveTargetRegistry = []destructiveSite{
	{
		File: "internal/cli/update/deploy/deploy.go", Function: "backupThenRemove", Sites: 3,
		Protection: "Card t111 (REQ-UDS-008 generalization): every regular file the embedded " +
			"template FS does not carry at the same relative path is copied into the run's " +
			"pre-clean backup (.moai-backups/<timestamp>/pre-clean/) BEFORE os.RemoveAll runs, " +
			"and a backup failure aborts the removal. Template-managed files need no copy — " +
			"the deploy step that follows rewrites them. .moai/config is additionally backed " +
			"up wholesale by backup.BackupMoaiConfig beforehand. The former hazard — the " +
			".claude/skills/moai* glob matching a user-authored moai-prefixed skill — now " +
			"lands in the pre-clean backup too, so nothing under a managed root leaves " +
			"without either a redeploy or a copy behind.",
	},
	{
		File: "internal/cli/update/deploy/deploy.go", Function: "removeSymlink", Sites: 4,
		Protection: "SPEC-CLI-CLEAN-SYMLINK-001 (REQ-CSL-002..004): all four sites remove a " +
			"symbolic link ENTRY, and os.RemoveAll never follows a symlink, so the target " +
			"tree is structurally out of reach. The live-file-link branch additionally copies " +
			"the target bytes (read through the link) into the pre-clean backup BEFORE the " +
			"removal and aborts on copy failure — the same backup-first ordering as " +
			"backupThenRemove. The dangling branch has no target to lose; the live-directory " +
			"branch removes the link only and never reads, walks, or backs up through it " +
			"(REQ-CSL-003). The link itself is user state the run deliberately withdraws — " +
			"a lead-ratified disposition (plan.md D-5) surfaced by a progress line naming " +
			"the path and form, not by a backup.",
	},
	{
		File: "internal/cli/update/deploy/deploy.go", Function: "MigrateLegacyMemoryDir", Sites: 2,
		Protection: "REQ-UDS-008 (M2): the both-exist branch copies .moai/memory/ into a run-scoped " +
			"backup directory before removing it, and a failed copy aborts the removal. The rename " +
			"branch moves the directory rather than destroying it.",
	},
	{
		File: "internal/cli/update_archive.go", Function: "archiveSkill", Sites: 1,
		Protection: "Clears a stale archive destination before re-archiving into it. The source skill " +
			"is copied to that destination by the same call, so the archive is rebuilt rather than lost.",
	},
	{
		File: "internal/cli/update_archive.go", Function: "archiveLegacySkills", Sites: 1,
		Protection: "Renames a legacy skill directory into the archive location rather than deleting " +
			"it; the content survives the move.",
	},
	{
		File: "internal/cli/update/backup/backup.go", Function: "BackupMoaiConfig", Sites: 3,
		Exemption: "Same-call rewind. All three sites unwind the backup directory this very call " +
			"created, on its own error paths, so nothing that predates the run can be reached.",
	},
	{
		File: "internal/cli/update/backup/backup.go", Function: "CleanupOldBackups", Sites: 1,
		Exemption: "Retention pruning of moai-authored backup directories — NOT a same-call rewind. " +
			"It deletes the oldest excess backups of PREVIOUS runs, filtered to the YYYYMMDD_HHMMSS " +
			"name pattern under the backup root, so every directory it removes was authored by moai " +
			"itself under a declared retention policy. Exempt from the user-data protection set; not " +
			"harmless to the recovery contract, since a sufficiently old restore point can be rotated " +
			"out by a later run.",
	},
	{
		File: "internal/cli/update_clean_install.go", Function: "runCleanReinstall", Sites: 1,
		Protection: "Cross-SPEC: SPEC-UPDATE-REINSTALL-LOOP-002 REQ-RIL2-015 backs up deprecated " +
			"paths before deletion. Not re-specified here (plan.md §G, REQ-UDS-010).",
	},
	{
		File: "internal/cli/update_cleanup.go", Function: "removeDeprecatedFile", Sites: 1,
		Protection: "Removes a single registered defs.DeprecatedPaths entry, each of which is a path " +
			"the template no longer ships and whose removal is declared with a DeprecatedBy SPEC and " +
			"a RemovalSchedule.",
	},
	{
		File: "internal/cli/update.go", Function: "ensureGlobalSettingsEnv", Sites: 1,
		Protection: "Deletion-radius pinning is M4's REQ-UDS-011/014: this site removes a subdirectory " +
			"of the operator's HOME, outside the project root, so a widened radius — not a missing " +
			"backup — is the failure mode. A HOME-scoped backup is out of scope because it would " +
			"itself write outside the project (plan.md §C.1 row 13).",
	},
	{
		File: "internal/cli/update_namespace_protect.go", Function: "backupUserOwnedNamespace", Sites: 3,
		Exemption: "Same-call rewind. All three sites are defensive cleanup (EC-UNP-007) of the " +
			"staging directory this very call created after a failed namespace backup; no data " +
			"predating the run is reachable.",
	},
	{
		File: "internal/cli/update_residue_cleanup.go", Function: "runV3ResidueCleanup", Sites: 1,
		Protection: "Cross-SPEC: SPEC-UPDATE-REINSTALL-LOOP-002 REQ-RIL2-019 backs up the residue " +
			"sweep before deleting it, aborting on failure. Not a same-call rewind — the swept paths " +
			"predate the run — but not re-specified here either (plan.md §G, REQ-UDS-010).",
	},
}
