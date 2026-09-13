package kanban

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAuditPureReadMutatesSchema(t *testing.T) {
	q := filepath.Join(t.TempDir(), "backlog.json")
	s := NewBacklogStore(q)
	if _, _, err := s.Add("one"); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open(sqliteDriverName, backlogDSN(s.EnginePath()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("DROP TABLE archived_findings; DROP TABLE archived_items"); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	before := InspectBacklogArchiveVouch(q).HasArchive
	_, err = s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	after := InspectBacklogArchiveVouch(q).HasArchive
	t.Logf("LoadPure archive tables: before=%v after=%v", before, after)
	if before != after {
		t.Error("read-only call changed persistent schema and archive availability")
	}
}

func TestAuditSeparateGitDirsCollide(t *testing.T) {
	base := t.TempDir()
	var roots []string
	for _, name := range []string{"a", "b"} {
		work := filepath.Join(base, "work-"+name)
		cmd := exec.Command("git", "init", "-q", "--separate-git-dir", filepath.Join(base, "git-"+name), work)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init: %v %s", err, out)
		}
		root := ResolveTodoQueueRoot(work)
		roots = append(roots, root)
		t.Logf("work=%s resolvedRoot=%s", work, root)
	}
	if roots[0] == roots[1] {
		t.Error("two independent git repositories resolve to same todo root")
	}
}

func TestAuditInterruptedMigrationShadowsLegacy(t *testing.T) {
	q := filepath.Join(t.TempDir(), "backlog.json")
	if err := os.WriteFile(q, []byte(`{"version":1,"last_seq":1,"items":[{"id":"t1","text":"one","state":"queued"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	// Materialize the exact engine state reached after migrateLegacyBacklog opens
	// its destination and before writeRecord commits. Simulate process loss here.
	eng, err := openBacklogEngine(backlogSQLitePath(q))
	if err != nil {
		t.Fatal(err)
	}
	_ = eng.close()
	if _, err := NewBacklogStore(q).LoadPure(); !IsBacklogCorrupt(err) {
		t.Fatalf("pure interrupted read error=%v, want explicit refusal", err)
	}
	r, loadErr := NewBacklogStore(q).Load()
	raw, err := os.ReadFile(q)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("reopened error=%v; retained legacy=%s", loadErr, raw)
	if loadErr != nil && !IsBacklogCorrupt(loadErr) {
		t.Fatal(loadErr)
	}
	if loadErr == nil && len(r.Items) != 1 {
		t.Error("interrupted migration makes populated legacy queue disappear from reads")
	}
}

func TestAuditAddRewritesWholeArchive(t *testing.T) {
	q := filepath.Join(t.TempDir(), "backlog.json")
	s := NewBacklogStore(q)
	if err := s.Mutate(func(r *BacklogRecord) error {
		for i := 1; i <= 1000; i++ {
			r.Archived = append(r.Archived, BacklogArchiveEntry{Item: BacklogItem{ID: fmt.Sprintf("t%d", i), Text: "finished", State: BacklogStatePicked}})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	eng, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = eng.close() }()
	if _, err := eng.db.Exec("CREATE TABLE audit_count(n INTEGER); INSERT INTO audit_count VALUES(0); CREATE TRIGGER audit_archive_insert AFTER INSERT ON archived_items BEGIN UPDATE audit_count SET n=n+1; END; CREATE TRIGGER audit_archive_delete AFTER DELETE ON archived_items BEGIN UPDATE audit_count SET n=n+1; END;"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Add("one new live card"); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := eng.db.QueryRow("SELECT n FROM audit_count").Scan(&n); err != nil {
		t.Fatal(err)
	}
	t.Logf("one Add with 1000 unchanged archived cards executes %d archive row writes", n)
	if n != 0 {
		t.Error("append writes unrelated archive rows")
	}
}

func TestAuditResolvedWriterAfterRelocation(t *testing.T) {
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "canonical")
	stale := NewBacklogStore(filepath.Join(from, "backlog.json"))
	if _, _, err := stale.Add("before migration"); err != nil {
		t.Fatal(err)
	}
	if err := relocateQueueArtifacts(from, to); err != nil {
		t.Fatal(err)
	}
	// A command can resolve the old path, wait behind migration's source
	// lock, then continue here. Locking the copy does not redirect that writer.
	_, _, addErr := stale.Add("after migration")
	if !errors.Is(addErr, ErrBacklogRelocated) {
		t.Fatalf("late writer error=%v, want explicit relocation refusal", addErr)
	}
	canonical, err := NewBacklogStore(filepath.Join(to, "backlog.json")).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("late Add refused=%v; canonical count=%d", addErr, len(canonical.Items))
	if len(canonical.Items) != 1 {
		t.Error("successful queued writer lands only in rollback source")
	}
}

func TestAuditFutureSchemaMutatedBeforeRefusal(t *testing.T) {
	p := filepath.Join(t.TempDir(), "backlog.db")
	db, err := sql.Open(sqliteDriverName, backlogDSN(p))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err = db.Exec("CREATE TABLE meta(key TEXT PRIMARY KEY,value TEXT); INSERT INTO meta VALUES('schema_version','999')"); err != nil {
		t.Fatal(err)
	}
	var before, after int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table'").Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, openErr := openBacklogEngine(p)
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table'").Scan(&after); err != nil {
		t.Fatal(err)
	}
	t.Logf("open error=%v; tables before=%d after=%d", openErr, before, after)
	if openErr == nil {
		t.Fatal("expected refusal")
	}
	if before != after {
		t.Error("unknown-version store mutated before refusal")
	}
}

func TestAuditRestoreDanglingRelation(t *testing.T) {
	r := &BacklogRecord{Items: []BacklogItem{{ID: "t1", State: BacklogStateQueued}, {ID: "t2", State: BacklogStateQueued}}, Findings: []BacklogFinding{{SubjectID: "t1", RelatedID: "t2", Relation: BacklogRelationContains}}}
	for _, id := range []string{"t1", "t2"} {
		if err := r.ArchiveCard(id); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.RestoreCard("t1"); err != nil {
		t.Fatal(err)
	}
	t.Logf("live=%v archived=%v findings=%v", r.Items, r.Archived, r.Findings)
	for _, f := range r.Findings {
		if r.itemIndex(f.SubjectID) < 0 || r.itemIndex(f.RelatedID) < 0 {
			t.Error("restored live finding references archived card")
		}
	}
	if err := r.RestoreCard("t2"); err != nil {
		t.Fatal(err)
	}
	if len(r.Findings) != 1 || r.Findings[0].RelatedID != "t2" {
		t.Fatalf("suspended relation lost: %+v", r.Findings)
	}
}

func TestAuditConstructorPreservesLegacyLock(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, legacyBacklogLockFileName)
	if err := os.WriteFile(legacy, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	NewBacklogStore(filepath.Join(dir, backlogFileName))
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("constructor removed legacy lock: %v", err)
	}
}

func TestAuditRetirementBeforePublishResumes(t *testing.T) {
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "canonical")
	source := NewBacklogStore(filepath.Join(from, backlogFileName))
	if _, _, err := source.Add("retained"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, backlogRetiredFileName), []byte(filepath.Join(to, backlogFileName)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := relocateQueueArtifacts(from, to); err != nil {
		t.Fatal(err)
	}
	r, err := NewBacklogStore(filepath.Join(to, backlogFileName)).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Items) != 1 || r.Items[0].Text != "retained" {
		t.Fatalf("resumed migration lost source: %+v", r)
	}
}

func TestAuditReadSnapshot(t *testing.T) {
	q := filepath.Join(t.TempDir(), "backlog.json")
	s := NewBacklogStore(q)
	err := s.Mutate(func(r *BacklogRecord) error {
		for i := 1; i <= 300; i++ {
			r.Items = append(r.Items, BacklogItem{ID: fmt.Sprintf("t%d", i), Text: "audit", State: BacklogStateQueued})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	reader, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.close() }()
	writer, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = writer.close() }()
	stop := make(chan struct{})
	writerErrors := make(chan error, 1)
	var commits atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	defer func() {
		close(stop)
		wg.Wait()
		select {
		case err := <-writerErrors:
			t.Errorf("concurrent writer failed: %v", err)
		default:
		}
		if commits.Load() == 0 {
			t.Error("snapshot probe had no concurrent committed writes")
		}
		t.Logf("concurrent writer committed %d transactions", commits.Load())
	}()
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			tx, err := writer.db.Begin()
			if err != nil {
				writerErrors <- err
				return
			}
			_, err = tx.Exec("INSERT INTO archived_items(seq,id,text,added_at,spec_id,state,position,landing) SELECT seq,id,text,added_at,spec_id,state,seq-1,landing FROM items; DELETE FROM items")
			if err != nil {
				_ = tx.Rollback()
				writerErrors <- err
				return
			}
			if err := tx.Commit(); err != nil {
				writerErrors <- err
				return
			}
			commits.Add(1)
			tx, err = writer.db.Begin()
			if err != nil {
				writerErrors <- err
				return
			}
			_, err = tx.Exec("INSERT INTO items(seq,id,text,added_at,spec_id,state,landing) SELECT seq,id,text,added_at,spec_id,state,landing FROM archived_items; DELETE FROM archived_items")
			if err != nil {
				_ = tx.Rollback()
				writerErrors <- err
				return
			}
			if err := tx.Commit(); err != nil {
				writerErrors <- err
				return
			}
			commits.Add(1)
		}
	}()
	for i := 0; i < 1000; i++ {
		r, err := reader.readRecord(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Items)+len(r.Archived) != 300 {
			t.Fatalf("read %d mixed commits: live=%d archived=%d total=%d want=300", i, len(r.Items), len(r.Archived), len(r.Items)+len(r.Archived))
		}
	}
	t.Log("no inconsistent read observed")
}

func TestAuditQuarantineRetryMarker(t *testing.T) {
	q := filepath.Join(t.TempDir(), "backlog.json")
	if err := os.WriteFile(q, []byte(`{"version":1,"last_seq":1,"items":[{"id":"t1","text":"one","state":"queued"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	// Existing quarantine is protected and makes the first rename refuse.
	if err := os.WriteFile(q+backlogMigratedSuffix, []byte("protected prior snapshot"), 0600); err != nil {
		t.Fatal(err)
	}
	s := NewBacklogStore(q)
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(q+backlogMigratedSuffix, q+".prior"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}
	_, err := os.Stat(q)
	t.Logf("after obstruction removed and Load retried: legacy JSON still present=%v", err == nil)
	if err == nil {
		t.Error("failed quarantine marker cleared, later Load cannot retry")
	}
}
