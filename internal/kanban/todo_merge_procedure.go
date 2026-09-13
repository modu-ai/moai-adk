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

// QueuePopulationCensus is the pre-merge population record (AC-TQM-010): the
// live/archived shape of both stores as read immediately before the merge
// decision. Every resolved duplicate must trace back to the project's
// ARCHIVED population — this census is what the population check is verified
// against.
type QueuePopulationCensus struct {
	HomeLive        int `json:"home_live"`
	HomeArchived    int `json:"home_archived"`
	ProjectLive     int `json:"project_live"`
	ProjectQueued   int `json:"project_queued"`
	ProjectPicked   int `json:"project_picked"`
	ProjectDropped  int `json:"project_dropped"`
	ProjectArchived int `json:"project_archived"`
}

// CensusRecords derives the pre-merge population census from the two pure
// reads.
func CensusRecords(home, project *BacklogRecord) QueuePopulationCensus {
	c := QueuePopulationCensus{
		HomeLive:     len(home.Items),
		HomeArchived: len(home.Archived),
		ProjectLive:  len(project.Items),
	}
	for _, it := range project.Items {
		switch it.State {
		case BacklogStateQueued:
			c.ProjectQueued++
		case BacklogStatePicked:
			c.ProjectPicked++
		case BacklogStateDropped:
			c.ProjectDropped++
		}
	}
	c.ProjectArchived = len(project.Archived)
	return c
}

// QueueMergeOutcome carries one procedure run's evidence.
type QueueMergeOutcome struct {
	DryRun            bool
	Backup            *QueueBackup
	Report            *MergeReport
	Census            QueuePopulationCensus
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
		Census:            CensusRecords(homeRec, projectRec),
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
// completeness (one row per renumbered card, every new id live carrying the
// project card's text); every project-provenance reference to a mapped old id
// rewritten to its new id. A GLOBAL "no old token" scan is deliberately NOT
// used: the old number legitimately stays occupied by the home card
// (REQ-TQM-007 keeps the number), so home-provenance references to it remain
// valid — the check is against the PROJECT's own references, whose meaning
// the renumber must preserve.
func VerifyMergedRecord(home, project, merged *BacklogRecord, report *MergeReport) QueueMergeVerification {
	v := QueueMergeVerification{
		Renumbered:  len(report.Renumbered),
		MappingRows: len(report.Mapping()),
	}

	// The union of card ids across both stores, live and archived.
	present := map[string]bool{}
	mergedTexts := map[string]bool{}
	countPresent := func(items []BacklogItem) {
		for _, it := range items {
			present[it.ID] = true
			mergedTexts[it.Text] = true
		}
	}
	countPresent(merged.Items)
	for _, e := range merged.Archived {
		present[e.Item.ID] = true
		mergedTexts[e.Item.Text] = true
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

	// Text-level cross-check (acceptance.md §D.3): every project card's text
	// survives in the merged store — under its id, its new id, or the
	// identical home copy of a resolved duplicate.
	projectText := map[string]string{}
	checkTexts := func(items []BacklogItem) {
		for _, it := range items {
			projectText[it.ID] = it.Text
			if !mergedTexts[it.Text] {
				v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("card text lost for %s: %q", it.ID, it.Text))
			}
		}
	}
	checkTexts(project.Items)
	for _, e := range project.Archived {
		projectText[e.Item.ID] = e.Item.Text
		if !mergedTexts[e.Item.Text] {
			v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("archived card text lost for %s", e.Item.ID))
		}
	}

	// AC-TQM-010 population check: every resolved duplicate's discarded copy
	// originated from the project store's ARCHIVED population. A duplicate row
	// naming a project LIVE card means the discriminator failed — a
	// live-work absorption, which is a zero-loss violation.
	duplicated := map[string]bool{}
	for _, d := range report.Duplicates {
		duplicated[d.ID] = true
	}
	projectArchivedIDs := map[string]bool{}
	for _, e := range project.Archived {
		projectArchivedIDs[e.Item.ID] = true
	}
	for _, d := range report.Duplicates {
		if !projectArchivedIDs[d.ID] {
			v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("duplicate row %s did not originate from the project ARCHIVED population (AC-TQM-010)", d.ID))
		}
	}

	// Mapping completeness: every new id live AND carrying the renumbered
	// card's own text.
	for _, row := range report.Mapping() {
		if !present[row.NewID] {
			v.StaleReferences = append(v.StaleReferences, "mapping target missing: "+row.NewID)
			continue
		}
		want := projectText[row.OldID]
		carried := false
		for _, it := range merged.Items {
			if it.ID == row.NewID && it.Text == want {
				carried = true
				break
			}
		}
		if !carried {
			for _, e := range merged.Archived {
				if e.Item.ID == row.NewID && e.Item.Text == want {
					carried = true
					break
				}
			}
		}
		if !carried {
			v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("mapping target %s does not carry the renumbered card's text", row.NewID))
		}
	}

	// Project-provenance reference rewrite (AC-TQM-004): every project
	// finding and assignment that names a mapped old id must appear in the
	// merged record naming the NEW id, with the note rewritten under the same
	// token boundaries.
	mapping := map[string]string{}
	for _, row := range report.Mapping() {
		mapping[row.OldID] = row.NewID
	}
	mergedHasFinding := func(want BacklogFinding) bool {
		for _, f := range merged.Findings {
			if f.SubjectID == want.SubjectID && f.RelatedID == want.RelatedID &&
				f.Relation == want.Relation && f.Source == want.Source && f.Note == want.Note {
				return true
			}
		}
		for _, e := range merged.Archived {
			for _, af := range e.Findings {
				if af.Finding.SubjectID == want.SubjectID && af.Finding.RelatedID == want.RelatedID &&
					af.Finding.Relation == want.Relation && af.Finding.Source == want.Source && af.Finding.Note == want.Note {
					return true
				}
			}
		}
		return false
	}
	touched := func(f BacklogFinding) bool {
		if _, ok := mapping[f.SubjectID]; ok {
			return true
		}
		if _, ok := mapping[f.RelatedID]; ok {
			return true
		}
		for old := range mapping {
			if cardTokenBoundaryContains(f.Note, old) {
				return true
			}
		}
		return false
	}
	for _, f := range project.Findings {
		if !touched(f) {
			continue
		}
		want := f
		want.SubjectID = rewriteCardTokens(f.SubjectID, mapping)
		want.RelatedID = rewriteCardTokens(f.RelatedID, mapping)
		want.Note = rewriteCardTokens(f.Note, mapping)
		if !mergedHasFinding(want) {
			v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("project finding %s->%s not rewritten into the merged record", f.SubjectID, f.RelatedID))
		}
	}
	for _, e := range project.Archived {
		for _, af := range e.Findings {
			f := af.Finding
			if !touched(f) {
				continue
			}
			want := f
			want.SubjectID = rewriteCardTokens(f.SubjectID, mapping)
			want.RelatedID = rewriteCardTokens(f.RelatedID, mapping)
			want.Note = rewriteCardTokens(f.Note, mapping)
			if !mergedHasFinding(want) {
				// REQ-TQM-006 v2: a duplicate-origin entry intentionally does
				// not carry into the merged record — the home store's same-id
				// copy is the surviving record. An identical home-original
				// finding keeps the reference valid (home keeps both
				// referenced ids), satisfying the rewrite the same way an
				// identical home card text satisfies the text check above. A
				// duplicate host whose home copy lacks the finding is still a
				// loss and stays flagged (t657 post-verify FAIL root cause).
				if !duplicated[e.Item.ID] || !mergedHasFinding(f) {
					v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("project archived finding %s->%s not rewritten", f.SubjectID, f.RelatedID))
				}
			}
		}
	}
	knownRuns := map[string]bool{}
	for _, run := range merged.Runtime.Runs {
		knownRuns[run.RunID] = true
	}
	for _, a := range project.Runtime.Assignments {
		if _, ok := mapping[a.CardID]; !ok || !knownRuns[a.RunID] {
			continue
		}
		found := false
		for _, m := range merged.Runtime.Assignments {
			if m.RunID == a.RunID && m.CardID == newID[a.CardID] {
				found = true
				break
			}
		}
		if !found {
			v.StaleReferences = append(v.StaleReferences, fmt.Sprintf("project assignment for %s (run %s) not rewritten", a.CardID, a.RunID))
		}
	}
	return v
}

// cardTokenBoundaryContains reports whether s carries token as a whole
// token — the same boundary rule the rewrite uses, applied as the verifier.
func cardTokenBoundaryContains(s, token string) bool {
	for _, tok := range mergeCardTokenPattern.FindAllString(s, -1) {
		if tok == token {
			return true
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
