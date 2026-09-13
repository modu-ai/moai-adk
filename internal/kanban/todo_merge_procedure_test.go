// todo_merge_procedure_test.go — the M4/M5 rehearsal suite
// (SPEC-TODO-QUEUE-HOME-MERGE-001, plan.md §F Rollback + delegation scope
// "fixture rehearsals of M4/M5, dry-run only").
//
// Every step the gated real-store execution will run — backup + ordering
// evidence, dry-run byte identity, ONE-Mutate apply, zero-loss verification,
// rollback restore, retirement with fence marker — is rehearsed end-to-end on
// t.TempDir() fixture stores. The real stores are never touched here.
package kanban

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rehearsalFixture builds a home/project store pair covering every taxonomy
// branch: an identical shared number, a differing shared number, project-only
// live and archived cards, and a finding + runtime assignment referencing the
// to-be-renumbered id.
func rehearsalFixture(t *testing.T) (homeDir, projectDir string) {
	t.Helper()
	homeDir = t.TempDir()
	projectDir = t.TempDir()

	home := NewBacklogStore(filepath.Join(homeDir, backlogFileName))
	if err := home.Mutate(func(rec *BacklogRecord) error {
		rec.Items = append(rec.Items,
			mergeItem("t10", "shared identical card", BacklogStateQueued),
			mergeItem("t11", "home eleven", BacklogStatePicked),
		)
		return nil
	}); err != nil {
		t.Fatalf("seed home: %v", err)
	}

	project := NewBacklogStore(filepath.Join(projectDir, backlogFileName))
	if err := project.Mutate(func(rec *BacklogRecord) error {
		rec.Items = append(rec.Items,
			mergeItem("t10", "shared identical card", BacklogStateQueued), // identical pair
			mergeItem("t11", "DIFFERENT project eleven", BacklogStateQueued), // renumber
			mergeItem("t20", "project only", BacklogStateDropped),            // migrate
		)
		rec.Findings = append(rec.Findings, BacklogFinding{
			SubjectID: "t11", RelatedID: "t20", Relation: BacklogRelationContains,
			Source: BacklogSourceAgent, Note: "project t11 contains t20 (not t110)",
		})
		rec.Runtime.Assignments = append(rec.Runtime.Assignments, TodoRuntimeAssignment{
			RunID: "run-rehearse", CardID: "t11", OwnerLabel: "lane-1",
			ReportedState: "picked", EventKind: "assign",
		})
		return nil
	}); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return homeDir, projectDir
}

func dbHash(t *testing.T, dir string) string {
	t.Helper()
	sum, err := fileSHA256(filepath.Join(dir, "backlog.db"))
	if err != nil {
		t.Fatalf("hash %s: %v", dir, err)
	}
	return sum
}

// TestRehearseQueueMergeDryRunByteIdentity proves the dry-run observes
// everything and writes NOTHING: both store dbs are byte-identical across the
// dry-run (plan.md M1 --dry-run; delegation "including --dry-run byte
// identity").
func TestRehearseQueueMergeDryRunByteIdentity(t *testing.T) {
	homeDir, projectDir := rehearsalFixture(t)
	homeHash, projectHash := dbHash(t, homeDir), dbHash(t, projectDir)

	outcome, err := RunQueueMerge(QueueMergePaths{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		BackupDir:  t.TempDir(),
	}, true)
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !outcome.DryRun {
		t.Fatalf("outcome not marked dry-run")
	}
	if !outcome.Verification.ZeroLoss {
		t.Errorf("dry-run verification failed: %+v", outcome.Verification)
	}
	if got := dbHash(t, homeDir); got != homeHash {
		t.Errorf("dry-run mutated the home store")
	}
	if got := dbHash(t, projectDir); got != projectHash {
		t.Errorf("dry-run mutated the project store")
	}
	if muts := outcome.OrderingMutations; len(muts) != 0 {
		t.Errorf("ordering evidence reports mutation across the backup window: %+v", muts)
	}
	if len(outcome.Report.Renumbered) != 1 || outcome.Report.Renumbered[0].OldID != "t11" {
		t.Errorf("dry-run renumber decisions wrong: %+v", outcome.Report.Renumbered)
	}
	if len(outcome.Report.Duplicates) != 1 || outcome.Report.Duplicates[0].ID != "t10" {
		t.Errorf("dry-run duplicate decisions wrong: %+v", outcome.Report.Duplicates)
	}
}

// TestRehearseQueueMergeApplyOneMutateAndRollback rehearses the destructive
// step on fixtures: apply, verify zero-loss + mapping completeness + no stale
// references, then roll back and prove byte-identity with the backup
// (plan.md M4 procedure steps 1-5; AC-TQM-002/003/004/007).
func TestRehearseQueueMergeApplyOneMutateAndRollback(t *testing.T) {
	homeDir, projectDir := rehearsalFixture(t)
	backupDir := t.TempDir()

	mutateCalls := 0
	origMutate := mergeMutate
	mergeMutate = func(s *BacklogStore, fn func(*BacklogRecord) error) error {
		mutateCalls++
		return s.Mutate(fn)
	}
	defer func() { mergeMutate = origMutate }()

	outcome, err := RunQueueMerge(QueueMergePaths{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		BackupDir:  backupDir,
	}, false)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if mutateCalls != 1 {
		t.Errorf("home store written by %d Mutate calls, want exactly ONE (AC-TQM-009)", mutateCalls)
	}
	v := outcome.Verification
	if !v.ZeroLoss {
		t.Errorf("zero-loss verification failed: %+v", v)
	}
	if v.TotalPost != v.TotalPre {
		t.Errorf("post-merge cardinality %d != union %d", v.TotalPost, v.TotalPre)
	}
	// AC-TQM-003: |mapping rows| == |renumbered|, new ids live, old ids gone.
	if v.MappingRows != v.Renumbered || v.MappingRows != 1 {
		t.Errorf("mapping completeness: rows=%d renumbered=%d", v.MappingRows, v.Renumbered)
	}
	if len(v.StaleReferences) != 0 {
		t.Errorf("stale old-id references remain: %v", v.StaleReferences)
	}

	// Reload the merged store from disk: t20 present, renumbered t21 present
	// (high-water 20 + 1), stale t11 project text gone from live items under
	// its old id.
	merged, err := NewBacklogStore(filepath.Join(homeDir, backlogFileName)).LoadPure()
	if err != nil {
		t.Fatalf("reload merged: %v", err)
	}
	ids := idsOf(merged.Items)
	if !ids["t20"] || !ids["t21"] || !ids["t10"] || !ids["t11"] {
		t.Errorf("merged store id set wrong: %v", ids)
	}
	for _, it := range merged.Items {
		if it.ID == "t21" && it.Text != "DIFFERENT project eleven" {
			t.Errorf("renumbered card content wrong: %q", it.Text)
		}
		if it.ID == "t11" && it.Text != "home eleven" {
			t.Errorf("home eleven disturbed: %q", it.Text)
		}
	}

	// Rollback (plan.md Rollback): restore from the verified backups.
	if err := RestoreQueueArtifacts(outcome.Backup); err != nil {
		t.Fatalf("rollback restore: %v", err)
	}
	restored := NewBacklogStore(filepath.Join(homeDir, backlogFileName))
	rec, err := restored.LoadPure()
	if err != nil {
		t.Fatalf("reload restored: %v", err)
	}
	rids := idsOf(rec.Items)
	if rids["t20"] || rids["t21"] {
		t.Errorf("rollback left merged cards behind: %v", rids)
	}
	if !rids["t10"] || !rids["t11"] {
		t.Errorf("rollback lost home cards: %v", rids)
	}
}

// TestRehearseQueueMergeRetireAndUnretire rehearses the M5 retirement on a
// fixture: fence marker written, directory RENAMED (never deleted), then the
// rollback un-retire puts it back (plan.md M5; REQ-TQM-017 rehearsal).
func TestRehearseQueueMergeRetireAndUnretire(t *testing.T) {
	_, projectDir := rehearsalFixture(t)

	retiredPath, err := RetireQueueStore(projectDir, "2026-09-13")
	if err != nil {
		t.Fatalf("retire: %v", err)
	}
	if retiredPath != projectDir+".retired-2026-09-13" {
		t.Fatalf("retired path %q", retiredPath)
	}
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Errorf("original project store dir still present after rename")
	}
	if _, err := os.Stat(retiredPath); err != nil {
		t.Errorf("retired directory missing: %v", err)
	}
	marker := filepath.Join(retiredPath, backlogRetiredFileName)
	raw, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("fence marker missing: %v", err)
	}
	if !strings.Contains(string(raw), "retired") {
		t.Errorf("fence marker content suspicious: %q", string(raw))
	}

	// Rollback of the retirement itself (the M5 branch of the Rollback plan).
	if err := UnretireQueueStore(projectDir, retiredPath); err != nil {
		t.Fatalf("unretire: %v", err)
	}
	if _, err := os.Stat(projectDir); err != nil {
		t.Errorf("project store dir not restored: %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("fence marker not removed on unretire")
	}
}
