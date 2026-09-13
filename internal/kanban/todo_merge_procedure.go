// todo_merge_procedure.go — the queue-merge procedure shell
// (SPEC-TODO-QUEUE-HOME-MERGE-001 M2/M4/M5 rehearsal surface, plan.md §F).
//
// RunQueueMerge is the procedure the gated M4 execution will drive against
// the real stores, factored so every step rehearses on fixture stores first:
// ordering-evidence snapshots → verified backup → LoadPure ×2 → the pure
// merge core → ONE Mutate on the home store (or nothing, on dry-run) →
// programmatic verification. RetireQueueStore / UnretireQueueStore are the
// M5 retirement and its rollback branch.
//
// The real-data write path is exactly ONE Mutate transaction on the home
// store (REQ-TQM-009, AC-TQM-009); the mergeMutate seam lets the rehearsal
// test COUNT the invocations. No direct SQL mutation and no byte-level edits
// exist on this path.
package kanban

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// QueueMergePaths addresses the procedure by explicit absolute paths
// (REQ-TQM-018). The caller resolves them; the procedure never derives a
// store from the working directory.
type QueueMergePaths struct {
	HomeDir    string // canonical home store directory
	ProjectDir string // legacy project store directory
	BackupDir  string // backup destination (REQ-TQM-001)
}

// QueueMergeVerification is the programmatic post-merge check result
// (AC-TQM-002/003/004).
type QueueMergeVerification struct {
	ZeroLoss        bool     // every union id present under its id or a mapping row
	TotalPre        int      // |union of card ids across both stores, live+archived|
	TotalPost       int      // merged store's live+archived count
	Renumbered      int      // cards renumbered
	MappingRows     int      // mapping rows produced
	StaleReferences []string // old-id tokens still present in findings/assignments
}

// QueueMergeOutcome carries one procedure run's evidence.
type QueueMergeOutcome struct {
	DryRun            bool
	Backup            *QueueBackup
	Report            *MergeReport
	OrderingBefore    StoreDirSnapshot
	OrderingAfter     StoreDirSnapshot
	OrderingMutations []StoreDirMutation
	Verification      QueueMergeVerification
	CompletedAt       time.Time
}

// mergeMutate is the ONE-Mutate seam: the procedure's only home-store write.
// The rehearsal test counts invocations through it (AC-TQM-009).
var mergeMutate = func(s *BacklogStore, fn func(*BacklogRecord) error) error {
	return s.Mutate(fn)
}

// RunQueueMerge executes the merge procedure against the addressed stores.
// With dryRun set, everything up to and including the merge DECISION runs and
// is verified in memory — no store byte is touched. Without it, the merged
// record lands through exactly ONE Mutate on the home store and the
// verification re-reads the committed record.
func RunQueueMerge(paths QueueMergePaths, dryRun bool) (*QueueMergeOutcome, error) {
	for _, dir := range []string{paths.HomeDir, paths.ProjectDir, paths.BackupDir} {
		if !filepath.IsAbs(dir) {
			return nil, fmt.Errorf("merge procedure: store paths must be absolute (REQ-TQM-018): %q", dir)
		}
	}
	homeStore := NewBacklogStore(filepath.Join(paths.HomeDir, backlogFileName))
	projectStore := NewBacklogStore(filepath.Join(paths.ProjectDir, backlogFileName))

	// Ordering evidence, BEFORE (AC-TQM-001: produced by the tooling).
	before := StoreDirSnapshot{}
	for label, dir := range map[string]string{"home": paths.HomeDir, "project": paths.ProjectDir} {
		snap, err := SnapshotStoreDirState(dir)
		if err != nil {
			return nil, fmt.Errorf("merge procedure: %s ordering snapshot: %w", label, err)
		}
		for rel, state := range snap {
			before[label+"/"+rel] = state
		}
	}

	// Backup + hash verification (REQ-TQM-001/002). Failure aborts with
	// neither store modified.
	backup, err := BackupQueueArtifacts([]string{paths.ProjectDir, paths.HomeDir}, paths.BackupDir)
	if err != nil {
		return nil, fmt.Errorf("merge procedure: backup: %w", err)
	}

	// Ordering evidence, AFTER the verified backup: zero mutation across the
	// window is the gate REQ-TQM-003 demands.
	after := StoreDirSnapshot{}
	for label, dir := range map[string]string{"home": paths.HomeDir, "project": paths.ProjectDir} {
		snap, err := SnapshotStoreDirState(dir)
		if err != nil {
			return nil, fmt.Errorf("merge procedure: %s ordering snapshot: %w", label, err)
		}
		for rel, state := range snap {
			after[label+"/"+rel] = state
		}
	}
	mutations := StoreDirMutations(before, after)
	if len(mutations) != 0 {
		return nil, fmt.Errorf("merge procedure: store mutated between audit start and backup verification (AC-TQM-001): %+v", mutations)
	}

	// Pure reads of both sources (REQ-TQM-009: LoadPure ×2).
	homeRec, err := homeStore.LoadPure()
	if err != nil {
		return nil, fmt.Errorf("merge procedure: load home: %w", err)
	}
	projectRec, err := projectStore.LoadPure()
	if err != nil {
		return nil, fmt.Errorf("merge procedure: load project: %w", err)
	}

	merged, report, err := MergeBacklogRecords(homeRec, projectRec, MergeOptions{})
	if err != nil {
		return nil, fmt.Errorf("merge procedure: merge core: %w", err)
	}

	outcome := &QueueMergeOutcome{
		DryRun:            dryRun,
		Backup:            backup,
		Report:            report,
		OrderingBefore:    before,
		OrderingAfter:     after,
		OrderingMutations: mutations,
	}

	if dryRun {
		outcome.Verification = VerifyMergedRecord(homeRec, projectRec, merged, report)
		return outcome, nil
	}

	// The destructive step: ONE locked whole-record transaction.
	if err := mergeMutate(homeStore, func(rec *BacklogRecord) error {
		*rec = *merged
		return nil
	}); err != nil {
		return nil, fmt.Errorf("merge procedure: home write (REQ-TQM-009): %w", err)
	}

	committed, err := homeStore.LoadPure()
	if err != nil {
		return nil, fmt.Errorf("merge procedure: reload committed home: %w", err)
	}
	outcome.Verification = VerifyMergedRecord(homeRec, projectRec, committed, report)
	outcome.CompletedAt = time.Now().UTC()
	if !outcome.Verification.ZeroLoss || len(outcome.Verification.StaleReferences) != 0 ||
		outcome.Verification.MappingRows != outcome.Verification.Renumbered {
		return outcome, fmt.Errorf("merge procedure: post-merge verification FAILED: %+v — restore from the verified backups before reporting (REQ-TQM-016)", outcome.Verification)
	}
	return outcome, nil
}

// VerifyMergedRecord is the programmatic verification (AC-TQM-002/003/004):
// every union id present under its original id or a mapping row; mapping
// completeness (one row per renumbered card, new ids live, old ids gone); no
// token-boundary old-id reference left in finding texts or assignment card
// references.
func VerifyMergedRecord(home, project, merged *BacklogRecord, report *MergeReport) QueueMergeVerification {
	v := QueueMergeVerification{
		Renumbered:  len(report.Renumbered),
		MappingRows: len(report.Mapping()),
	}

	// The union of card ids across both stores, live and archived.
	present := map[string]bool{}
	countPresent := func(items []BacklogItem) {
		for _, it := range items {
			present[it.ID] = true
		}
	}
	countPresent(merged.Items)
	for _, e := range merged.Archived {
		present[e.Item.ID] = true
	}
	newID := map[string]string{}
	for _, row := range report.Mapping() {
		newID[row.OldID] = row.NewID
	}
	countUnion := func(items []BacklogItem) {
		for _, it := range items {
			v.TotalPre++
			if present[it.ID] || present[newID[it.ID]] {
				v.TotalPost++
			}
		}
	}
	countUnion(home.Items)
	countUnion(project.Items)
	for _, e := range home.Archived {
		v.TotalPre++
		if present[e.Item.ID] || present[newID[e.Item.ID]] {
			v.TotalPost++
		}
	}
	for _, e := range project.Archived {
		v.TotalPre++
		if present[e.Item.ID] || present[newID[e.Item.ID]] {
			v.TotalPost++
		}
	}
	v.ZeroLoss = v.TotalPre == v.TotalPost

	// Mapping completeness beyond the row count: every new id live; and the
	// renumbered PROJECT card must not survive under its old id. The number
	// itself may legitimately stay occupied — by the home card, which
	// REQ-TQM-007 keeps in place — so the check is against the project
	// variant's own text, not against bare occupancy.
	projectText := map[string]string{}
	for _, it := range project.Items {
		projectText[it.ID] = it.Text
	}
	for _, e := range project.Archived {
		projectText[e.Item.ID] = e.Item.Text
	}
	for _, row := range report.Mapping() {
		if !present[row.NewID] {
			v.StaleReferences = append(v.StaleReferences, "mapping target missing: "+row.NewID)
		}
		for _, it := range merged.Items {
			if it.ID == row.OldID && it.Text == projectText[row.OldID] {
				v.StaleReferences = append(v.StaleReferences, "renumbered card still present under old id "+row.OldID)
			}
		}
	}

	// Stale old-id references in finding texts and assignment card refs
	// (token-boundary scan, AC-TQM-004).
	scan := func(s string) {
		for _, row := range report.Mapping() {
			if cardTokenBoundaryContains(s, row.OldID) {
				v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("stale %s in %q", row.OldID, s))
			}
		}
	}
	for _, f := range merged.Findings {
		scan(f.SubjectID + "\x00" + f.RelatedID + "\x00" + f.Note)
	}
	for _, e := range merged.Archived {
		for _, af := range e.Findings {
			scan(af.Finding.SubjectID + "\x00" + af.Finding.RelatedID + "\x00" + af.Finding.Note)
		}
	}
	for _, a := range merged.Runtime.Assignments {
		scan(a.CardID)
	}
	return v
}

// cardTokenBoundaryContains reports whether s carries token as a whole
// token — the same boundary rule the rewrite uses, applied as the verifier.
func cardTokenBoundaryContains(s, token string) bool {
	for _, field := range strings.Split(s, "\x00") {
		for _, tok := range mergeCardTokenPattern.FindAllString(field, -1) {
			if tok == token {
				return true
			}
		}
	}
	return false
}

// RetireQueueStore retires the project store (plan.md M5, REQ-TQM-017): the
// existing retired fence marker is written INSIDE the directory and the
// directory is RENAMED to <dir>.retired-<date> — preserved for inspection,
// never deleted.
func RetireQueueStore(projectDir string, dateStamp string) (string, error) {
	marker := filepath.Join(projectDir, backlogRetiredFileName)
	note := fmt.Sprintf("retired %s: merged into the canonical home store (SPEC-TODO-QUEUE-HOME-MERGE-001, card t657)\n", dateStamp)
	if err := os.WriteFile(marker, []byte(note), 0o600); err != nil {
		return "", fmt.Errorf("retire %s: fence marker: %w", projectDir, err)
	}
	retiredPath := projectDir + ".retired-" + dateStamp
	if err := os.Rename(projectDir, retiredPath); err != nil {
		_ = os.Remove(marker)
		return "", fmt.Errorf("retire %s: rename: %w", projectDir, err)
	}
	return retiredPath, nil
}

// UnretireQueueStore reverses a retirement (the M5 branch of the Rollback
// plan): rename back and remove the fence marker.
func UnretireQueueStore(projectDir, retiredPath string) error {
	if err := os.Rename(retiredPath, projectDir); err != nil {
		return fmt.Errorf("unretire %s: rename: %w", retiredPath, err)
	}
	if err := os.Remove(filepath.Join(projectDir, backlogRetiredFileName)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unretire %s: remove marker: %w", projectDir, err)
	}
	return nil
}
