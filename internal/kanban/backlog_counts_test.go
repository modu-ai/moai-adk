// backlog_counts_test.go — the shared count surface and the small store
// accessors (SPEC-TODO-SQLITE-001 REQ-TOSQ-009, C-1; M6).
//
// BacklogCountsForRoot is the statusline's per-render path and the kanban
// notice's. It was reachable only from those packages' tests, which meant the
// contract lived one import away from the code that defines it.
package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

// The counts are read from whichever layout exists, and dropped cards are
// deliberately not counted — they are history, and a number that only ever
// grows is noise.
func TestBacklogCountsForRoot(t *testing.T) {
	t.Parallel()

	seed := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		store := NewBacklogStore(BacklogPathForRootAdopting(root))
		if err := store.Mutate(func(rec *BacklogRecord) error {
			spec := "SPEC-EXAMPLE-001"
			rec.LastSeq = 4
			rec.Items = []BacklogItem{
				{ID: "t1", Text: "a", AddedAt: "T", State: BacklogStateQueued},
				{ID: "t2", Text: "b", AddedAt: "T", State: BacklogStatePicked, SpecID: &spec},
				{ID: "t3", Text: "c", AddedAt: "T", State: BacklogStateDropped},
				{ID: "t4", Text: "d", AddedAt: "T", State: BacklogStateQueued},
			}
			return nil
		}); err != nil {
			t.Fatalf("seed: %v", err)
		}
		return root
	}

	t.Run("database layout", func(t *testing.T) {
		got := BacklogCountsForRoot(seed(t))
		if !got.Available || got.Queued != 2 || got.Picked != 1 {
			t.Fatalf("counts = %+v, want queued 2 / picked 1 / available", got)
		}
	})

	t.Run("legacy layout is counted without migrating", func(t *testing.T) {
		root := t.TempDir()
		dir := LegacyStateDirForRoot(root)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("seed dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, backlogFileName), []byte(migrationFixture), 0o644); err != nil {
			t.Fatalf("seed queue: %v", err)
		}
		got := BacklogCountsForRoot(root)
		if !got.Available || got.Queued != 2 || got.Picked != 1 {
			t.Fatalf("counts = %+v, want queued 2 / picked 1 / available", got)
		}
		if dirExists(StateDirForRoot(root)) {
			t.Error("counting relocated the state directory")
		}
	})

	t.Run("absent and empty-root read as unavailable", func(t *testing.T) {
		if got := BacklogCountsForRoot(t.TempDir()); got.Available {
			t.Errorf("absent queue = %+v, want Available=false", got)
		}
		if got := BacklogCountsForRoot(""); got.Available {
			t.Errorf("empty root = %+v, want Available=false", got)
		}
	})

	t.Run("corrupt database is unavailable, not zero", func(t *testing.T) {
		root := seed(t)
		db := filepath.Join(StateDirForRoot(root), "backlog.db")
		if err := os.WriteFile(db, []byte("not a database"), 0o600); err != nil {
			t.Fatalf("corrupt: %v", err)
		}
		got := BacklogCountsForRoot(root)
		if got.Available {
			t.Fatalf("corrupt queue = %+v, want Available=false — 'no cards' and 'could not read the cards' are different claims", got)
		}
	})

	t.Run("corrupt legacy json is unavailable, not zero", func(t *testing.T) {
		root := t.TempDir()
		dir := LegacyStateDirForRoot(root)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("seed dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, backlogFileName), []byte(`{ TRUNCATED`), 0o644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		if got := BacklogCountsForRoot(root); got.Available {
			t.Fatalf("corrupt legacy = %+v, want Available=false", got)
		}
	})
}

// QueuedBacklogCountForRoot is the one-call form the notice and the factory
// loop render from; it must agree with the fuller counts on the same queue.
func TestQueuedBacklogCountForRootAgreesWithCounts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	store := NewBacklogStore(BacklogPathForRootAdopting(root))
	for i := 0; i < 3; i++ {
		if _, _, err := store.Add("card"); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	if got, want := QueuedBacklogCountForRoot(root), BacklogCountsForRoot(root).Queued; got != want {
		t.Fatalf("QueuedBacklogCountForRoot = %d, BacklogCountsForRoot.Queued = %d — the two surfaces must not disagree about what 'waiting' means", got, want)
	}
	if QueuedBacklogCountForRoot("") != 0 {
		t.Error("QueuedBacklogCountForRoot(\"\") != 0; no project means no queue")
	}
}

// EnginePath and LockPath name the store's siblings. They are diagnostics, but
// tests stat what they return, so a wrong answer here misdirects every
// invariance assertion that uses them.
func TestBacklogStoreSiblingPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store := NewBacklogStore(filepath.Join(dir, "backlog.json"))

	if got, want := store.EnginePath(), filepath.Join(dir, "backlog.db"); got != want {
		t.Errorf("EnginePath = %q, want %q", got, want)
	}
	if got, want := store.LockPath(), filepath.Join(dir, backlogLockFileName); got != want {
		t.Errorf("LockPath = %q, want %q", got, want)
	}
	if store.Path() != filepath.Join(dir, "backlog.json") {
		t.Errorf("Path = %q, want the queue document", store.Path())
	}
}

// IsBacklogBusy completes the error-taxonomy predicate trio. The other two are
// exercised by the engine suite; this one is asserted here so all three have a
// direct test rather than two direct and one assumed.
func TestIsBacklogBusy(t *testing.T) {
	t.Parallel()
	if !IsBacklogBusy(ErrBacklogBusy) {
		t.Error("IsBacklogBusy(ErrBacklogBusy) = false, want true")
	}
	if IsBacklogBusy(ErrBacklogCorrupt) || IsBacklogBusy(nil) {
		t.Error("IsBacklogBusy matched a non-busy error")
	}
	wrapped := mapBacklogEngineError("op", ErrBacklogBusy)
	if !IsBacklogBusy(wrapped) {
		t.Error("IsBacklogBusy did not see through the op wrapper")
	}
}
