// backlog_store_errors_test.go — supplementary error-path coverage for the
// M1 backlog store (SPEC-KANBAN-TODO-CLI-001). The behavioral contract lives
// in backlog_store_test.go; these tests exercise the error branches that the
// happy-path contract cannot reach (unreadable file, lock/lock-dir failures,
// unwritable target dir, schema normalization of minimal files, and the id
// parser's refusals).
package kanban

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestBacklogLoad_PathIsDirectorySurfacesReadError — a backlog path that is
// a directory is a read error, distinct from the missing-file empty queue.
func TestBacklogLoad_PathIsDirectorySurfacesReadError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store := NewBacklogStore(filepath.Join(dir, "backlog.json"))
	if err := os.Mkdir(store.Path(), 0o755); err != nil {
		t.Fatalf("seed directory-at-path: %v", err)
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("Load(directory path) err = nil, want read error")
	}
}

// TestBacklogMutate_CallbackRefusalLeavesFileUnchanged — a refusing mutate
// callback aborts with the file untouched.
func TestBacklogMutate_CallbackRefusalLeavesFileUnchanged(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	refused := os.ErrPermission
	if err := store.Mutate(func(rec *BacklogRecord) error { return refused }); err == nil {
		t.Fatal("Mutate(refusing callback) err = nil, want refusal")
	}
	if _, statErr := os.Stat(store.Path()); statErr == nil {
		t.Fatal("backlog file appeared despite refused mutation")
	}
}

// TestBacklogMutate_ParentIsFileFailsBeforeLock — a store path beneath a
// regular file cannot have its directory created, so the mutation fails
// before the lock is ever attempted.
func TestBacklogMutate_ParentIsFileFailsBeforeLock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed blocker file: %v", err)
	}
	store := NewBacklogStore(filepath.Join(blocker, "backlog.json"))

	if _, _, err := store.Add("unwritable"); err == nil {
		t.Fatal("Add(beneath a file) err = nil, want dir-creation error")
	}
}

// TestBacklogMutate_LockPathIsDirectorySurfacesNonLockError — a lock
// artifact path occupied by a directory is an acquisition failure that is
// NOT contention (no bounded retry), surfaced immediately.
func TestBacklogMutate_LockPathIsDirectorySurfacesNonLockError(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)
	if err := os.MkdirAll(store.LockPath(), 0o755); err != nil {
		t.Fatalf("seed directory-at-lock-path: %v", err)
	}

	if _, _, err := store.Add("blocked by dir"); err == nil {
		t.Fatal("Add(lock path is a directory) err = nil, want acquisition error")
	} else if IsBoardLockHeld(err) {
		t.Fatalf("err = %v, want a non-contention acquisition failure", err)
	}
}

// TestBacklogWrite_UnwritableDirFailsLoudly — a directory the process cannot
// write MUST surface an error rather than report a mutation it did not
// persist.
//
// The scenario moved with the storage swap and the move is the point. The
// pre-SQLite store created a temp file inside the directory on every write, so
// revoking the write bit from INSIDE the mutation callback failed the write.
// The engine has no temp-file step: by the time the callback runs, the
// database is already open, and an open descriptor keeps writing through a
// directory that turned read-only underneath it — the write lands, durably and
// correctly, and there is nothing to report. So this test revokes the bit
// BEFORE the store opens, which is the hazard that still exists.
//
// The failure observed on that path is the LOCK artifact's, not the database's:
// the mutation cannot create backlog.lock, so it never reaches the engine at
// all. That ordering is deliberate rather than incidental — the lock is what
// serializes factory lanes, and a mutation that proceeded without it would be
// the lost-update race the store exists to prevent. The read path's own
// failure (which does reach the engine) is asserted separately below.
func TestBacklogWrite_UnwritableDirFailsLoudly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only directory does not block file creation on Windows; temp creation succeeds and the error branch is never reached")
	}
	if os.Geteuid() == 0 {
		t.Skip("read-only directory does not block the root user")
	}
	t.Parallel()
	store := newTestBacklogStore(t)
	dir := filepath.Dir(store.Path())
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("revoke write bit: %v", err)
	}

	called := false
	err := store.Mutate(func(rec *BacklogRecord) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("Mutate(read-only dir) err = nil, want an open/write failure")
	}
	if called {
		t.Error("the mutation callback ran against a store that could not be opened")
	}
	if _, statErr := os.Stat(store.Path()); statErr == nil {
		t.Error("a queue file appeared in an unwritable directory")
	}
}

// TestBacklogRead_UnwritableDirSurfacesEngineFailure — the read path takes no
// lock, so it reaches the engine, and the engine cannot open a database it
// cannot create. That MUST surface as an error rather than as an empty queue:
// reporting "no cards" for a queue that could not be read would tell a lane
// there is no work when there may be eighty cards waiting.
func TestBacklogRead_UnwritableDirSurfacesEngineFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only directory does not block file creation on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("read-only directory does not block the root user")
	}
	t.Parallel()
	store := newTestBacklogStore(t)
	dir := filepath.Dir(store.Path())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("revoke write bit: %v", err)
	}

	if _, err := store.Load(); err == nil {
		t.Fatal("Load(unwritable dir) err = nil, want an engine open failure")
	}
}

// TestBacklogNormalize_MinimalFiles — a file carrying neither version nor
// items normalizes to a writable version-1 record with an empty item slice.
func TestBacklogNormalize_MinimalFiles(t *testing.T) {
	t.Parallel()

	noItems := newTestBacklogStore(t)
	seedBacklogFile(t, noItems, `{"version":1,"last_seq":4}`)
	if err := noItems.Mutate(func(rec *BacklogRecord) error { return nil }); err != nil {
		t.Fatalf("Mutate(no items key): %v", err)
	}
	rec, err := noItems.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rec.Items == nil || len(rec.Items) != 0 {
		t.Fatalf("Items = %v, want normalized empty non-nil slice", rec.Items)
	}
	if rec.LastSeq != 4 {
		t.Fatalf("LastSeq = %d, want preserved 4", rec.LastSeq)
	}

	noVersion := newTestBacklogStore(t)
	seedBacklogFile(t, noVersion, `{"items":[]}`)
	if err := noVersion.Mutate(func(rec *BacklogRecord) error { return nil }); err != nil {
		t.Fatalf("Mutate(no version key): %v", err)
	}
	rec, err = noVersion.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rec.Version != 1 {
		t.Fatalf("Version = %d, want normalized 1", rec.Version)
	}
}

// TestParseBacklogSeq_Refusals — ids that are not t<N> with N > 0 are not
// sequence numbers and never feed the high-water mark.
func TestParseBacklogSeq_Refusals(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"", "t", "x5", "tabc", "t0", "t-3", "T7"} {
		if n, ok := parseBacklogSeq(id); ok {
			t.Fatalf("parseBacklogSeq(%q) = %d, true; want refusal", id, n)
		}
	}
	if n, ok := parseBacklogSeq("t42"); !ok || n != 42 {
		t.Fatalf("parseBacklogSeq(t42) = %d, %v; want 42, true", n, ok)
	}
}

// TestJoinBacklogReleaseErr_BothFailuresSurvive — the regression this exists
// for: when the mutation fails AND the lock release fails, BOTH must reach the
// caller. The prior guard (`relErr != nil && err == nil`) dropped the release
// error in exactly that case, so a wedged lock — the artifact that blocks every
// later writer on Windows — surfaced as an ordinary mutation fault.
func TestJoinBacklogReleaseErr_BothFailuresSurvive(t *testing.T) {
	t.Parallel()
	mutErr := errors.New("mutation refused")
	relErr := errors.New("close: bad file descriptor")

	got := joinBacklogReleaseErr(mutErr, relErr, "/tmp/backlog.json")
	if got == nil {
		t.Fatal("joinBacklogReleaseErr(mut, rel) = nil; want both errors")
	}
	if !errors.Is(got, mutErr) {
		t.Errorf("mutation error did not survive the join: %v", got)
	}
	if !errors.Is(got, relErr) {
		t.Errorf("release error did not survive the join: %v", got)
	}
	if !strings.Contains(got.Error(), "lock release failed") {
		t.Errorf("joined error does not name the release failure: %v", got)
	}
}

// TestJoinBacklogReleaseErr_SingleAndCleanPaths — the other three combinations.
// A clean release must not manufacture an error, and either failure alone must
// pass through identifiably.
func TestJoinBacklogReleaseErr_SingleAndCleanPaths(t *testing.T) {
	t.Parallel()
	mutErr := errors.New("mutation refused")
	relErr := errors.New("close: bad file descriptor")

	if got := joinBacklogReleaseErr(nil, nil, "/tmp/backlog.json"); got != nil {
		t.Errorf("both-clean returned %v; want nil", got)
	}
	if got := joinBacklogReleaseErr(mutErr, nil, "/tmp/backlog.json"); !errors.Is(got, mutErr) {
		t.Errorf("mutation-only returned %v; want the mutation error", got)
	}
	got := joinBacklogReleaseErr(nil, relErr, "/tmp/backlog.json")
	if !errors.Is(got, relErr) {
		t.Errorf("release-only returned %v; want the release error", got)
	}
	if !strings.Contains(got.Error(), "lock release failed") {
		t.Errorf("release-only error does not name the failure: %v", got)
	}
}
