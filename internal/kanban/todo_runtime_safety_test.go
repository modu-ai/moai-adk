package kanban

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runtimeFixture(t *testing.T) (string, *BacklogStore, string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	s := NewBacklogStore(BacklogPathForRoot(root))
	card, _, err := s.Add("card")
	if err != nil {
		t.Fatal(err)
	}
	return root, s, card.ID
}

func TestTodoRuntimeSafetyAssignmentAbortRollsBackSeed(t *testing.T) {
	root, s, card := runtimeFixture(t)
	if err := RecordFactoryRunStart(root, "existing", BackendClaude, ""); err != nil {
		t.Fatal(err)
	}
	e, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.close() }()
	if _, err := e.db.Exec(`CREATE TRIGGER abort_assignment BEFORE INSERT ON todo_runtime_assignments BEGIN SELECT RAISE(ABORT,'fixture_assignment_abort'); END`); err != nil {
		t.Fatal(err)
	}
	err = RecordFactoryCardAssignment(root, "aborted", card, "direct", "")
	if err == nil || !strings.Contains(err.Error(), "fixture_assignment_abort") {
		t.Fatalf("abort trigger not reached: %v", err)
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Runtime.Runs) != 1 || r.Runtime.Runs[0].RunID != "existing" || len(r.Runtime.Assignments) != 0 {
		t.Fatalf("partial aborted request: %+v", r.Runtime)
	}
}

func TestTodoRuntimeSafetyErrorsPropagate(t *testing.T) {
	root, s, _ := runtimeFixture(t)
	if err := RecordFactoryRunStart(root, " ", BackendClaude, ""); err == nil {
		t.Error("empty run accepted")
	}
	if err := RecordFactoryRunStart(root, "seed", BackendClaude, ""); err != nil {
		t.Fatal(err)
	}
	e, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.close() }()
	if _, err := e.db.Exec(`CREATE TRIGGER abort_run BEFORE INSERT ON todo_runtime_runs BEGIN SELECT RAISE(ABORT,'fixture_run_abort'); END`); err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryRunStart(root, "abort", BackendClaude, ""); err == nil || !strings.Contains(err.Error(), "fixture_run_abort") {
		t.Fatalf("run error not propagated: %v", err)
	}
	if err := os.Mkdir(filepath.Join(filepath.Dir(s.path), backlogRetiredFileName), 0700); err != nil {
		t.Fatal(err)
	}
	if err := s.recordRuntime(TodoRuntimeRun{RunID: "bad-marker"}, nil); err == nil {
		t.Fatal("unreadable retired marker accepted")
	}
}

func TestTodoRuntimeSafetyCorruptRowsDoNotReadAsEmpty(t *testing.T) {
	for _, assignment := range []bool{false, true} {
		t.Run(fmt.Sprint(assignment), func(t *testing.T) {
			root, s, card := runtimeFixture(t)
			if err := RecordFactoryCardAssignment(root, "run", card, "owner", ""); err != nil {
				t.Fatal(err)
			}
			e, err := openBacklogEngine(s.EnginePath())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = e.close() }()
			ddl := `DROP TABLE todo_runtime_runs; CREATE TABLE todo_runtime_runs(run_id TEXT,backend TEXT,manifest_json TEXT); INSERT INTO todo_runtime_runs VALUES(NULL,'claude','{}')`
			if assignment {
				ddl = `DROP TABLE todo_runtime_assignments; CREATE TABLE todo_runtime_assignments(run_id TEXT,card_id TEXT,owner_label TEXT,reported_state TEXT,event_kind TEXT,provenance_json TEXT); INSERT INTO todo_runtime_assignments VALUES('run',NULL,'owner','picked','card.assigned','{}')`
			}
			if _, err := e.db.Exec(ddl); err != nil {
				t.Fatal(err)
			}
			if _, err := s.LoadPure(); err == nil {
				t.Fatal("corrupt runtime row became valid empty runtime")
			}
		})
	}
}

func TestTodoRuntimeSafetyExtensionInstallRollsBack(t *testing.T) {
	root, s, _ := runtimeFixture(t)
	e, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.close() }()
	if _, err := e.db.Exec(`CREATE TRIGGER abort_runtime_stamp BEFORE INSERT ON meta WHEN NEW.key='runtime_schema_version' BEGIN SELECT RAISE(ABORT,'fixture_install_abort'); END`); err != nil {
		t.Fatal(err)
	}
	err = RecordFactoryRunStart(root, "aborted-install", BackendClaude, "")
	if err == nil || !strings.Contains(err.Error(), "fixture_install_abort") {
		t.Fatalf("install abort not reached: %v", err)
	}
	var count int
	if err := e.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name IN ('todo_runtime_runs','todo_runtime_assignments')`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("partial extension tables=%d", count)
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Items) != 1 || r.LastSeq != 1 || len(r.Runtime.Runs) != 0 {
		t.Fatalf("queue altered: %+v", r)
	}
}

func TestTodoRuntimeSafetyStampedPartialSchemaRefused(t *testing.T) {
	root, s, _ := runtimeFixture(t)
	if err := RecordFactoryRunStart(root, "existing", BackendClaude, ""); err != nil {
		t.Fatal(err)
	}
	e, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`DROP TABLE todo_runtime_assignments`); err != nil {
		t.Fatal(err)
	}
	if err := e.close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadPure(); !IsBacklogCorrupt(err) {
		t.Errorf("stamped partial read: %v", err)
	}
	if err := RecordFactoryRunStart(root, "partial", BackendClaude, ""); !IsBacklogCorrupt(err) {
		t.Errorf("stamped partial writer: %v", err)
	}
	after, err := os.ReadFile(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("stamped partial rejection changed bytes")
	}
}

func TestTodoRuntimeSafetyRetiredAndMissingCardRefuse(t *testing.T) {
	root, s, _ := runtimeFixture(t)
	if err := RecordFactoryCardAssignment(root, "missing", "t999", "direct", ""); err == nil {
		t.Fatal("missing card accepted")
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Runtime.Runs) != 0 {
		t.Fatal("missing card left run seed")
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(s.path), backlogRetiredFileName), []byte("new-root"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	// Use the resolved store to exercise an old writer retained before relocation.
	if err := s.recordRuntime(TodoRuntimeRun{RunID: "retired", ManifestJSON: "{}"}, nil); !errors.Is(err, ErrBacklogRelocated) {
		t.Fatalf("retired writer: %v", err)
	}
	after, err := os.ReadFile(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("retired writer changed DB")
	}
}

func TestTodoRuntimeSafetyArchiveAndStaleDTOArePreserved(t *testing.T) {
	root, s, card := runtimeFixture(t)
	if err := s.Mutate(func(r *BacklogRecord) error { return r.ArchiveCard(card) }); err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryCardAssignment(root, "archive", card, "first", ""); err != nil {
		t.Fatal(err)
	}
	stale, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryCardAssignment(root, "archive", card, "latest", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Mutate(func(r *BacklogRecord) error { r.Runtime = stale.Runtime; return nil }); err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Items) != 0 || len(r.Archived) != 1 || r.Archived[0].Item.ID != card || len(r.Runtime.Assignments) != 1 || r.Runtime.Assignments[0].OwnerLabel != "latest" {
		t.Fatalf("archive/stale overwrite: %+v", r)
	}
}

func TestTodoRuntimeSafetyExtensionPreservesLegacyContentAndSorts(t *testing.T) {
	root, s, card := runtimeFixture(t)
	other, _, err := s.Add("archived")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Mutate(func(r *BacklogRecord) error {
		r.LastSeq = 100
		r.Findings = append(r.Findings, BacklogFinding{SubjectID: card, RelatedID: other.ID, Relation: "contains", Source: "agent", Note: "retained", Score: 0.5, At: "2026-09-12T00:00:00Z"})
		return r.ArchiveCard(other.ID)
	}); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	for _, run := range []string{"z", "a"} {
		if err := RecordFactoryRunStart(root, run, "first", ""); err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{other.ID, card} {
			if err := RecordFactoryCardAssignment(root, run, id, "owner", ""); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := RecordFactoryRunStart(root, "a", "latest", ""); err != nil {
		t.Fatal(err)
	}
	after, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Runtime.Runs) != 2 || after.Runtime.Runs[0].RunID != "a" || after.Runtime.Runs[0].Backend != "latest" || len(after.Runtime.Assignments) != 4 || after.Runtime.Assignments[0].CardID != card {
		t.Fatalf("upsert/sort: %+v", after.Runtime)
	}
	after.Runtime = TodoRuntime{}
	got, err := json.Marshal(after)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got) {
		t.Fatalf("legacy content changed: before=%s after=%s", want, got)
	}
}

func TestTodoRuntimeSafetyRelocationPreservesRuntime(t *testing.T) {
	root, s, card := runtimeFixture(t)
	if err := RecordFactoryCardAssignment(root, "relocate", card, "owner", ""); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	to := filepath.Join(t.TempDir(), "adopted")
	if err := relocateQueueArtifacts(filepath.Dir(s.path), to); err != nil {
		t.Fatal(err)
	}
	after, err := NewBacklogStore(filepath.Join(to, backlogFileName)).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(before.Runtime)
	got, _ := json.Marshal(after.Runtime)
	if !bytes.Equal(want, got) {
		t.Fatalf("relocation dropped runtime: before=%s after=%s", want, got)
	}
}

func TestTodoRuntimeSafetyParityRejectsRuntimeLoss(t *testing.T) {
	a, b := &BacklogRecord{}, &BacklogRecord{}
	a.Runtime.Runs = []TodoRuntimeRun{{RunID: "retained", ManifestJSON: "{}"}}
	if err := assertBacklogParity(a, b); err == nil {
		t.Fatal("migration parity accepted lost runtime")
	}
}

func TestTodoRuntimeSafetyCardMutationAndAssignmentSerialize(t *testing.T) {
	root, s, card := runtimeFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	editDone := make(chan error, 1)
	go func() {
		editDone <- s.Mutate(func(r *BacklogRecord) error { close(entered); <-release; r.Items[0].Text = "edited"; return nil })
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("mutator did not acquire lock")
	}
	writeDone := make(chan error, 1)
	go func() { writeDone <- RecordFactoryCardAssignment(root, "concurrent", card, "direct", "") }()
	close(release)
	for _, done := range []<-chan error{editDone, writeDone} {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("writer did not finish")
		}
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if r.Items[0].Text != "edited" || len(r.Runtime.Assignments) != 1 {
		t.Fatalf("concurrent write lost: %+v", r)
	}
}

func TestTodoRuntimeSafetyPureSnapshotReadsActiveWAL(t *testing.T) {
	root, s, card := runtimeFixture(t)
	if err := RecordFactoryCardAssignment(root, "wal", card, "direct", ""); err != nil {
		t.Fatal(err)
	}
	e, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.close() }()
	if _, err := e.db.Exec(`PRAGMA wal_autocheckpoint=0`); err != nil {
		t.Fatal(err)
	}
	write := func(value string) error {
		tx, err := e.db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if _, err = tx.Exec(`UPDATE items SET text=? WHERE id=?`, value, card); err != nil {
			return err
		}
		if _, err = tx.Exec(`UPDATE todo_runtime_assignments SET reported_state=? WHERE card_id=?`, value, card); err != nil {
			return err
		}
		return tx.Commit()
	}
	if err := write("v0"); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(s.EnginePath() + "-wal")
	if err != nil || len(before) == 0 {
		t.Fatalf("no active WAL: %v", err)
	}
	for i := 0; i < 3; i++ {
		r, err := s.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if r.Items[0].Text != "v0" || r.Runtime.Assignments[0].ReportedState != "v0" {
			t.Fatalf("WAL values: %+v", r)
		}
	}
	after, err := os.ReadFile(s.EnginePath() + "-wal")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("pure read changed active WAL")
	}
	done := make(chan error, 1)
	go func() {
		for i := 1; i <= 30; i++ {
			if err := write(fmt.Sprintf("v%d", i)); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	// Wait for the bounded writer even when an assertion fails, before closing its DB.
	defer func() {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	for i := 0; i < 30; i++ {
		r, err := s.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if r.Items[0].Text != r.Runtime.Assignments[0].ReportedState {
			t.Fatalf("mixed snapshots: card=%s runtime=%s", r.Items[0].Text, r.Runtime.Assignments[0].ReportedState)
		}
	}
}
