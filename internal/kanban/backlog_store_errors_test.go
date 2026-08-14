// backlog_store_errors_test.go — supplementary error-path coverage for the
// M1 backlog store (SPEC-KANBAN-TODO-CLI-001). The behavioral contract lives
// in backlog_store_test.go; these tests exercise the error branches that the
// happy-path contract cannot reach (unreadable file, lock/lock-dir failures,
// unwritable target dir, schema normalization of minimal files, and the id
// parser's refusals).
package kanban

import (
	"os"
	"path/filepath"
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

// TestBacklogWrite_UnwritableDirFailsTempCreation — the mutation callback
// itself revokes the directory's write bit, so the write path fails at temp
// file creation while the lock is already held: the error surfaces and no
// residue is left beyond the lock artifact.
func TestBacklogWrite_UnwritableDirFailsTempCreation(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory does not block the root user")
	}
	t.Parallel()
	store := newTestBacklogStore(t)
	dir := filepath.Dir(store.Path())
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := store.Mutate(func(rec *BacklogRecord) error {
		return os.Chmod(dir, 0o500)
	})
	if err == nil {
		t.Fatal("Mutate(read-only dir) err = nil, want temp-creation error")
	}
	for _, name := range readBacklogDirNames(t, store) {
		if name != "backlog.lock" {
			t.Fatalf("backlog dir holds leaked file %q after failed write", name)
		}
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
