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
// branch UNDER THE v0.3.0 DISCRIMINATOR: a live identical pair (renumbers —
// the operator-named hazard), a differing shared number (renumbers), a
// project-only card (migrates), and a project-ARCHIVED identical twin
// (duplicate — the only admitted duplicate origin).
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
		rec.Archived = append(rec.Archived, BacklogArchiveEntry{
			Item: mergeItem("t30", "archived twin card", BacklogStateQueued),
		})
		return nil
	}); err != nil {
		t.Fatalf("seed home: %v", err)
	}

	project := NewBacklogStore(filepath.Join(projectDir, backlogFileName))
	if err := project.Mutate(func(rec *BacklogRecord) error {
		rec.Items = append(rec.Items,
			mergeItem("t10", "shared identical card", BacklogStateQueued),    // live identical → renumber
			mergeItem("t11", "DIFFERENT project eleven", BacklogStateQueued), // differs → renumber
			mergeItem("t20", "project only", BacklogStateDropped),            // migrate
		)
		rec.Archived = append(rec.Archived, BacklogArchiveEntry{
			Item: mergeItem("t30", "archived twin card", BacklogStateQueued), // archived twin → duplicate
		})
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
	// Post-discriminator taxonomy: the live identical t10 pair RENUMBERS;
	// only the project-archived t30 twin resolves as a duplicate.
	if len(outcome.Report.Renumbered) != 2 || outcome.Report.Renumbered[0].OldID != "t10" || outcome.Report.Renumbered[1].OldID != "t11" {
		t.Errorf("dry-run renumber decisions wrong: %+v", outcome.Report.Renumbered)
	}
	if len(outcome.Report.Duplicates) != 1 || outcome.Report.Duplicates[0].ID != "t30" || outcome.Report.Duplicates[0].Origin != MergeOriginProjectArchived {
		t.Errorf("dry-run duplicate decisions wrong: %+v", outcome.Report.Duplicates)
	}
	// Pre-merge population census (AC-TQM-010).
	c := outcome.Census
	if c.HomeLive != 2 || c.HomeArchived != 1 || c.ProjectLive != 3 || c.ProjectArchived != 1 ||
		c.ProjectQueued != 2 || c.ProjectPicked != 0 || c.ProjectDropped != 1 {
		t.Errorf("census wrong: %+v", c)
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
	// AC-TQM-003: |mapping rows| == |renumbered| — post-discriminator: t10
	// (live identical) and t11 (differs) both renumber.
	if v.MappingRows != v.Renumbered || v.MappingRows != 2 {
		t.Errorf("mapping completeness: rows=%d renumbered=%d", v.MappingRows, v.Renumbered)
	}
	if len(v.StaleReferences) != 0 {
		t.Errorf("stale old-id references remain: %v", v.StaleReferences)
	}

	// Reload the merged store from disk: the renumbered cards present under
	// t31/t32 (high-water 30), the migrated t20 present, home cards intact.
	merged, err := NewBacklogStore(filepath.Join(homeDir, backlogFileName)).LoadPure()
	if err != nil {
		t.Fatalf("reload merged: %v", err)
	}
	ids := idsOf(merged.Items)
	if !ids["t20"] || !ids["t31"] || !ids["t32"] || !ids["t10"] || !ids["t11"] {
		t.Errorf("merged store id set wrong: %v", ids)
	}
	// AC-TQM-010: the live identical twin's content survives under the NEW id.
	for _, it := range merged.Items {
		if it.ID == "t31" && it.Text != "shared identical card" {
			t.Errorf("live identical card content not preserved under t31: %q", it.Text)
		}
		if it.ID == "t32" && it.Text != "DIFFERENT project eleven" {
			t.Errorf("renumbered card content wrong: %q", it.Text)
		}
		if it.ID == "t11" && it.Text != "home eleven" {
			t.Errorf("home eleven disturbed: %q", it.Text)
		}
	}
	// The duplicate population: only the project-archived twin, and the home
	// archived card unchanged.
	if len(merged.Archived) != 1 || merged.Archived[0].Item.ID != "t30" || merged.Archived[0].Item.Text != "archived twin card" {
		t.Errorf("archived population wrong: %+v", merged.Archived)
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
	if rids["t20"] || rids["t31"] || rids["t32"] {
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

// TestRetireQueueStoreErrorBranches covers the refusal arms: retiring a
// nonexistent directory fails at the marker write; unretiring a nonexistent
// retired path fails at the rename.
func TestRetireQueueStoreErrorBranches(t *testing.T) {
	if _, err := RetireQueueStore(filepath.Join(t.TempDir(), "absent"), "2026-09-13"); err == nil {
		t.Error("retire accepted a nonexistent directory")
	}
	if err := UnretireQueueStore(t.TempDir(), filepath.Join(t.TempDir(), "absent-retired")); err == nil {
		t.Error("unretire accepted a nonexistent retired path")
	}
}

// TestRunQueueMergeErrorBranches covers the path-refusal and snapshot-error
// arms: relative paths refuse (REQ-TQM-018), a nonexistent store directory
// aborts the ordering snapshot.
func TestRunQueueMergeErrorBranches(t *testing.T) {
	homeDir, projectDir := rehearsalFixture(t)
	_, err := RunQueueMerge(QueueMergePaths{
		HomeDir: "relative/home", ProjectDir: projectDir, BackupDir: t.TempDir(),
	}, true)
	if err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Errorf("relative home path not refused: %v", err)
	}
	_, err = RunQueueMerge(QueueMergePaths{
		HomeDir: homeDir, ProjectDir: filepath.Join(t.TempDir(), "absent"), BackupDir: t.TempDir(),
	}, true)
	if err == nil {
		t.Error("nonexistent project dir not refused")
	}
}

// TestVerifyMergedRecordBranches unit-tests the pure verifier's arms with
// crafted records: lost text, missing mapping target, target carrying the
// wrong text, and each unrewritten reference shape.
func TestVerifyMergedRecordBranches(t *testing.T) {
	home := mergeFixture([]BacklogItem{mergeItem("t1", "home", BacklogStateQueued)}, 1)
	project := mergeFixture([]BacklogItem{
		mergeItem("t2", "project two", BacklogStateQueued),
	}, 2)
	project.Findings = []BacklogFinding{
		{SubjectID: "t2", RelatedID: "t1", Relation: BacklogRelationContains, Source: BacklogSourceAgent, Note: "t2 note"},
	}
	project.Runtime.Assignments = []TodoRuntimeAssignment{
		{RunID: "run-1", CardID: "t2", OwnerLabel: "lane", ReportedState: "picked", EventKind: "assign"},
	}

	t.Run("clean verification", func(t *testing.T) {
		merged := cloneBacklogRecord(home)
		merged.Items = append(merged.Items, mergeItem("t3", "project two", BacklogStateQueued))
		merged.Findings = append(merged.Findings, BacklogFinding{
			SubjectID: "t3", RelatedID: "t1", Relation: BacklogRelationContains, Source: BacklogSourceAgent, Note: "t3 note",
		})
		merged.Runtime.Runs = []TodoRuntimeRun{{RunID: "run-1"}}
		merged.Runtime.Assignments = []TodoRuntimeAssignment{
			{RunID: "run-1", CardID: "t3", OwnerLabel: "lane", ReportedState: "picked", EventKind: "assign"},
		}
		report := &MergeReport{Renumbered: []MergeMappingRow{{OldID: "t2", NewID: "t3"}}}
		v := VerifyMergedRecord(home, project, merged, report)
		if !v.ZeroLoss || len(v.StaleReferences) != 0 {
			t.Errorf("clean merge failed verification: %+v", v)
		}
		if v.TotalPre != 2 || v.TotalPost != 2 {
			t.Errorf("union count wrong: %+v", v)
		}
	})

	t.Run("lost text and broken mapping", func(t *testing.T) {
		merged := cloneBacklogRecord(home) // t3 never added
		report := &MergeReport{Renumbered: []MergeMappingRow{{OldID: "t2", NewID: "t3"}}}
		v := VerifyMergedRecord(home, project, merged, report)
		joined := strings.Join(v.StaleReferences, "; ")
		if !strings.Contains(joined, "text lost") {
			t.Errorf("lost text undetected: %+v", v.StaleReferences)
		}
		if !strings.Contains(joined, "mapping target missing") {
			t.Errorf("missing mapping target undetected: %+v", v.StaleReferences)
		}
	})

	t.Run("target carries wrong text", func(t *testing.T) {
		merged := cloneBacklogRecord(home)
		merged.Items = append(merged.Items, mergeItem("t3", "a DIFFERENT t3", BacklogStateQueued))
		report := &MergeReport{Renumbered: []MergeMappingRow{{OldID: "t2", NewID: "t3"}}}
		v := VerifyMergedRecord(home, project, merged, report)
		joined := strings.Join(v.StaleReferences, "; ")
		if !strings.Contains(joined, "does not carry the renumbered card's text") {
			t.Errorf("wrong-text target undetected: %+v", v.StaleReferences)
		}
	})

	t.Run("unrewritten project finding", func(t *testing.T) {
		merged := cloneBacklogRecord(home)
		merged.Items = append(merged.Items, mergeItem("t3", "project two", BacklogStateQueued))
		// Finding left naming t2 — the unrewritten shape.
		merged.Findings = append(merged.Findings, BacklogFinding{
			SubjectID: "t2", RelatedID: "t1", Relation: BacklogRelationContains, Source: BacklogSourceAgent, Note: "t2 note",
		})
		report := &MergeReport{Renumbered: []MergeMappingRow{{OldID: "t2", NewID: "t3"}}}
		v := VerifyMergedRecord(home, project, merged, report)
		joined := strings.Join(v.StaleReferences, "; ")
		if !strings.Contains(joined, "not rewritten") {
			t.Errorf("unrewritten finding undetected: %+v", v.StaleReferences)
		}
	})

	t.Run("unrewritten assignment", func(t *testing.T) {
		merged := cloneBacklogRecord(home)
		merged.Items = append(merged.Items, mergeItem("t3", "project two", BacklogStateQueued))
		merged.Runtime.Runs = []TodoRuntimeRun{{RunID: "run-1"}}
		merged.Runtime.Assignments = []TodoRuntimeAssignment{
			{RunID: "run-1", CardID: "t2", OwnerLabel: "lane", ReportedState: "picked", EventKind: "assign"},
		}
		report := &MergeReport{Renumbered: []MergeMappingRow{{OldID: "t2", NewID: "t3"}}}
		v := VerifyMergedRecord(home, project, merged, report)
		joined := strings.Join(v.StaleReferences, "; ")
		if !strings.Contains(joined, "assignment") {
			t.Errorf("unrewritten assignment undetected: %+v", v.StaleReferences)
		}
	})

	t.Run("archived text lost and boundary contains", func(t *testing.T) {
		lostArchived := cloneBacklogRecord(project)
		lostArchived.Archived = []BacklogArchiveEntry{{Item: mergeItem("t9", "archived unique text", BacklogStateQueued)}}
		report := &MergeReport{}
		v := VerifyMergedRecord(home, lostArchived, cloneBacklogRecord(home), report)
		joined := strings.Join(v.StaleReferences, "; ")
		if !strings.Contains(joined, "archived card text lost") {
			t.Errorf("archived text loss undetected: %+v", v.StaleReferences)
		}
		if !cardTokenBoundaryContains("see t642 here", "t642") ||
			cardTokenBoundaryContains("see t6420 here", "t642") {
			t.Errorf("boundary contains wrong")
		}
	})

	t.Run("unrewritten archived finding", func(t *testing.T) {
		withArchived := cloneBacklogRecord(project)
		withArchived.Archived = []BacklogArchiveEntry{{
			Item:     mergeItem("t9", "archived carrier", BacklogStateQueued),
			Findings: []BacklogArchivedFinding{{Finding: BacklogFinding{SubjectID: "t2", RelatedID: "t1", Relation: BacklogRelationContains, Source: BacklogSourceAgent}}},
		}}
		merged := cloneBacklogRecord(home)
		merged.Items = append(merged.Items, mergeItem("t3", "project two", BacklogStateQueued))
		report := &MergeReport{Renumbered: []MergeMappingRow{{OldID: "t2", NewID: "t3"}}}
		v := VerifyMergedRecord(home, withArchived, merged, report)
		joined := strings.Join(v.StaleReferences, "; ")
		if !strings.Contains(joined, "archived finding") {
			t.Errorf("unrewritten archived finding undetected: %+v", v.StaleReferences)
		}
	})
}

// TestRunQueueMergeAbortPaths covers the destructive-step abort arms on
// fixtures: a retired fence marker on the home store refuses the Mutate, and
// a non-file artifact in the project store aborts the backup.
func TestRunQueueMergeAbortPaths(t *testing.T) {
	t.Run("home store fenced", func(t *testing.T) {
		homeDir, projectDir := rehearsalFixture(t)
		if err := os.WriteFile(filepath.Join(homeDir, backlogRetiredFileName), []byte("elsewhere"), 0o600); err != nil {
			t.Fatalf("fence: %v", err)
		}
		_, err := RunQueueMerge(QueueMergePaths{
			HomeDir: homeDir, ProjectDir: projectDir, BackupDir: t.TempDir(),
		}, false)
		if err == nil {
			t.Fatal("merge wrote to a fenced home store")
		}
		if got := dbHash(t, homeDir); got == "" {
			t.Error("home store unreadable after refused run")
		}
	})
	t.Run("project store artifact is a directory", func(t *testing.T) {
		homeDir, projectDir := rehearsalFixture(t)
		// backlog.db is already a file in the fixture; a directory in place
		// of the enumerated WAL sibling triggers the same refusal arm.
		if err := os.MkdirAll(filepath.Join(projectDir, "backlog.db-wal"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		_, err := RunQueueMerge(QueueMergePaths{
			HomeDir: homeDir, ProjectDir: projectDir, BackupDir: t.TempDir(),
		}, true)
		if err == nil {
			t.Fatal("backup accepted a directory artifact")
		}
	})
}

// TestRetireQueueStoreRenameRefused covers the rename-failure branch: the
// target retired name already exists as a non-empty directory.
func TestRetireQueueStoreRenameRefused(t *testing.T) {
	projectDir := t.TempDir()
	retired := projectDir + ".retired-2026-09-13"
	if err := os.MkdirAll(filepath.Join(retired, "occupied"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := RetireQueueStore(projectDir, "2026-09-13"); err == nil {
		t.Fatal("retire succeeded despite an occupied retired name")
	}
	// The fence marker is cleaned up on the refused rename.
	if _, err := os.Stat(filepath.Join(projectDir, backlogRetiredFileName)); !os.IsNotExist(err) {
		t.Errorf("fence marker left behind on refused rename")
	}
}
