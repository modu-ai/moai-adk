// state_dir_test.go — the directory-rename matrix
// (SPEC-TODO-SQLITE-001 AC-TOSQ-006/007/008, REQ-TOSQ-015; M3, absorbing t309).
//
// The relocation moves a directory eight lanes and a lead are reading. The
// three outcomes it must produce — relocate, defer to the new name, serve the
// old one — are each exercised against a real filesystem here, including the
// refusal path, which is fired by making the rename impossible rather than by
// asserting the design says so.
package kanban

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// seedLegacyStateDir plants a queue file plus N session-registry files under
// the legacy directory and returns the registry file names.
func seedLegacyStateDir(t *testing.T, root string, registryCount int) []string {
	t.Helper()
	dir := LegacyStateDirForRoot(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("seed legacy dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, backlogFileName), []byte(migrationFixture), 0o644); err != nil {
		t.Fatalf("seed legacy queue: %v", err)
	}
	names := make([]string, 0, registryCount)
	for i := 0; i < registryCount; i++ {
		// Synthesized session ids — a uuid shape without production values.
		name := "0000000" + string(rune('a'+i)) + "-1111-2222-3333-444444444444.json"
		body := `{"session_id":"probe-` + string(rune('a'+i)) + `","spec_id":"SPEC-EXAMPLE-001"}`
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("seed registry file: %v", err)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// dirCensus lists a directory's entries, sorted.
func dirCensus(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("census %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func TestRelocateQueueArtifactsCopiesLogicalDataAndKeepsRollbackSource(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "global")
	if err := os.MkdirAll(from, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(from, backlogFileName)
	if err := os.WriteFile(source, []byte(migrationFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := filepath.Join(from, "session.json")
	if err := os.WriteFile(registry, []byte(`{"session_id":"session"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := relocateQueueArtifacts(from, to); err != nil {
		t.Fatal(err)
	}
	got, err := NewBacklogStore(filepath.Join(to, backlogFileName)).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 4 {
		t.Fatalf("migrated items=%d, want 4", len(got.Items))
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("rollback source missing: %v", err)
	}
	if _, err := os.Stat(registry); err != nil {
		t.Fatalf("unowned registry moved: %v", err)
	}
	if info, err := os.Stat(filepath.Join(to, "backlog.db")); err != nil {
		t.Fatalf("target database missing: %v", err)
	} else if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("target database mode=%#o, want 0600", got)
	}
}

func TestRelocateQueueArtifactsRemovesIncompleteTargetOnWriteFailure(t *testing.T) {
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "global")
	if err := os.MkdirAll(from, 0o700); err != nil {
		t.Fatal(err)
	}
	broken := `{"version":1,"last_seq":2,"items":[{"id":"t1","text":"a","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"},{"id":"t1","text":"b","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}`
	if err := os.WriteFile(filepath.Join(from, backlogFileName), []byte(broken), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := relocateQueueArtifacts(from, to); err == nil {
		t.Fatal("relocation unexpectedly succeeded")
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		path := filepath.Join(to, "backlog.db") + suffix
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("incomplete target survived at %s: %v", path, err)
		}
	}
}

func TestRelocateQueueArtifactsNeverOverwritesExistingHomeQueue(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "global")
	if err := os.MkdirAll(from, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, backlogFileName), []byte(migrationFixture), 0o600); err != nil {
		t.Fatal(err)
	}

	target := NewBacklogStore(filepath.Join(to, backlogFileName))
	if _, _, err := target.Add("new home card"); err != nil {
		t.Fatal(err)
	}
	if err := relocateQueueArtifacts(from, to); err != nil {
		t.Fatal(err)
	}
	rec, err := target.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 1 || rec.Items[0].Text != "new home card" {
		t.Fatalf("existing home queue was overwritten: %+v", rec.Items)
	}
}

func TestRelocateQueueArtifactsSerializesOnTargetQueueLock(t *testing.T) {
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "global")
	if err := os.MkdirAll(from, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, backlogFileName), []byte(migrationFixture), 0o600); err != nil {
		t.Fatal(err)
	}

	target := NewBacklogStore(filepath.Join(to, backlogFileName))
	targetLock, err := target.acquireLock()
	if err != nil {
		t.Fatalf("acquire target lock: %v", err)
	}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { _ = targetLock.Release() }) }
	t.Cleanup(release)

	done := make(chan error, 1)
	go func() { done <- relocateQueueArtifacts(from, to) }()
	select {
	case err := <-done:
		t.Fatalf("relocation bypassed target queue lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	release()
	if err := <-done; err != nil {
		t.Fatalf("relocation after target unlock: %v", err)
	}
}

func TestAdoptingPathMigratesLegacyHomeQueue(t *testing.T) {
	home := t.TempDir()
	root := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	legacyDir := filepath.Join(home, "todo", TodoQueueProjectKey(root))
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(legacyDir, backlogFileName)
	if err := os.WriteFile(legacy, []byte(migrationFixture), 0o600); err != nil {
		t.Fatal(err)
	}

	path := BacklogPathForRootAdopting(root)
	want := filepath.Join(home, "db", TodoQueueProjectKey(root), "todo", backlogFileName)
	if path != want {
		t.Fatalf("adopting path=%q, want %q", path, want)
	}
	rec, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 4 {
		t.Fatalf("migrated items=%d, want 4", len(rec.Items))
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("legacy rollback source missing: %v", err)
	}
}

func TestAdoptingPathFindsLegacyHomeQueueWithPreSanitizationKey(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(t.TempDir(), "한글 프로젝트")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(abs))
	legacyKey := fmt.Sprintf("%s-%x", filepath.Base(abs), sum[:4])
	if legacyKey == TodoQueueProjectKey(root) {
		t.Fatal("fixture did not produce a legacy/new key difference")
	}
	legacyDir := filepath.Join(home, "todo", legacyKey)
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(legacyDir, backlogFileName)
	if err := os.WriteFile(legacy, []byte(migrationFixture), 0o600); err != nil {
		t.Fatal(err)
	}

	path := BacklogPathForRootAdopting(root)
	rec, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 4 {
		t.Fatalf("migrated items=%d, want 4", len(rec.Items))
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("legacy rollback source missing: %v", err)
	}
}

func TestBacklogSQLiteArtifactsArePrivateWhileOpen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "backlog.db")
	eng, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = eng.close() }()
	if err := eng.writeRecord(t.Context(), &BacklogRecord{Version: 1, Items: []BacklogItem{}, Findings: []BacklogFinding{}, Archived: []BacklogArchiveEntry{}}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("%s mode=%#o, want 0600", filepath.Base(path), got)
		}
	}
}

// AC-TOSQ-006 / REQ-TOSQ-015: only the legacy directory exists. An adopting
// open relocates the WHOLE directory — the queue file AND every session
// registry file — and the queue reads through afterwards.
func TestStateDirRelocatesLegacyWithRegistryFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := seedLegacyStateDir(t, root, 3)

	path := BacklogPathForRootAdopting(root)
	if want := filepath.Join(StateDirForRoot(root), backlogFileName); path != want {
		t.Fatalf("adopting path = %q, want %q", path, want)
	}

	// Census: N registry files + the queue file, all under the new name.
	got := dirCensus(t, StateDirForRoot(root))
	want := append(append([]string{}, registry...), backlogFileName)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("relocated census = %v, want %v", got, want)
	}
	if dirExists(LegacyStateDirForRoot(root)) {
		t.Error("the legacy directory survived the relocation — a rename leaves nothing behind")
	}

	// One sampled record survives byte-for-byte: a census that counted files
	// while corrupting them would pass a count check.
	sample, err := os.ReadFile(filepath.Join(StateDirForRoot(root), registry[0]))
	if err != nil {
		t.Fatalf("read relocated registry file: %v", err)
	}
	if !strings.Contains(string(sample), `"session_id":"probe-a"`) {
		t.Errorf("relocated registry file content = %q, want it unchanged", sample)
	}

	// And the queue is readable through the store at its new home.
	rec, err := NewBacklogStore(path).Load()
	if err != nil {
		t.Fatalf("Load after relocation: %v", err)
	}
	if len(rec.Items) != 4 {
		t.Fatalf("items after relocation = %d, want 4", len(rec.Items))
	}

	// RecordPath follows the directory without being told: the registry and
	// the queue travel together or the relocation has split the channel.
	rp := RecordPath(root, "0000000a-1111-2222-3333-444444444444")
	if filepath.Dir(rp) != StateDirForRoot(root) {
		t.Errorf("RecordPath resolves to %q, want it under %q", filepath.Dir(rp), StateDirForRoot(root))
	}
}

// AC-TOSQ-007 / REQ-TOSQ-015 stale-copy: both directories exist. The new name
// wins on every path and the legacy directory is left STRICTLY untouched —
// visible, so an operator still writing to the dead path can see the
// divergence rather than have it silently absorbed.
func TestStateDirBothPresentLeavesLegacyUntouched(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLegacyStateDir(t, root, 2)

	// The new directory exists with its own distinct content.
	current := StateDirForRoot(root)
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatalf("create current dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(current, backlogFileName),
		[]byte(`{"version":1,"last_seq":99,"items":[],"findings":[]}`), 0o644); err != nil {
		t.Fatalf("seed current queue: %v", err)
	}

	legacyBefore := dirCensus(t, LegacyStateDirForRoot(root))

	// Exercise BOTH resolution forms plus a real read and a real write.
	if got := BacklogPathForRoot(root); filepath.Dir(got) != current {
		t.Errorf("pure resolution = %q, want it under the current dir", got)
	}
	path := BacklogPathForRootAdopting(root)
	if filepath.Dir(path) != current {
		t.Errorf("adopting resolution = %q, want it under the current dir", path)
	}
	store := NewBacklogStore(path)
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rec.LastSeq != 99 {
		t.Fatalf("last_seq = %d, want 99 — the current directory must win", rec.LastSeq)
	}
	if _, _, err := store.Add("written under the current name"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	legacyAfter := dirCensus(t, LegacyStateDirForRoot(root))
	if strings.Join(legacyBefore, ",") != strings.Join(legacyAfter, ",") {
		t.Fatalf("legacy directory changed %v -> %v — the stale copy is inviolable", legacyBefore, legacyAfter)
	}
	legacyQueue, err := os.ReadFile(filepath.Join(LegacyStateDirForRoot(root), backlogFileName))
	if err != nil {
		t.Fatalf("read legacy queue: %v", err)
	}
	if string(legacyQueue) != migrationFixture {
		t.Error("the legacy queue file's contents changed — it must be left strictly untouched")
	}
}

// AC-TOSQ-008 / REQ-TOSQ-015 fallback READ: the relocation is refused. No
// error reaches the verb; the queue is served from the old layout best-effort.
//
// The refusal is FIRED, not simulated: the parent directory's write bit is
// revoked, so the rename genuinely cannot happen. A filesystem that refuses
// for its own reasons (a cross-device mount) reaches the same branch.
func TestStateDirRelocationRefusedFallsBackToLegacyRead(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only directory does not block rename on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("read-only directory does not block the root user")
	}
	t.Parallel()
	root := t.TempDir()
	seedLegacyStateDir(t, root, 1)

	stateParent := filepath.Join(root, ".moai", "state")
	t.Cleanup(func() { _ = os.Chmod(stateParent, 0o755) })
	if err := os.Chmod(stateParent, 0o500); err != nil {
		t.Fatalf("revoke write bit on state parent: %v", err)
	}

	path := BacklogPathForRootAdopting(root)
	if filepath.Dir(path) != LegacyStateDirForRoot(root) {
		t.Fatalf("refused relocation resolved to %q, want the legacy directory served in place", filepath.Dir(path))
	}
	if dirExists(StateDirForRoot(root)) {
		t.Error("a partial new directory was created despite the refusal")
	}

	// The queue is still usable: this is the whole point of failing open.
	rec, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		t.Fatalf("LoadPure(legacy fallback) = %v, want the queue served best-effort", err)
	}
	if len(rec.Items) != 4 {
		t.Fatalf("items = %d, want 4 from the legacy layout", len(rec.Items))
	}
}

// REQ-TOSQ-015: the PURE resolution never relocates, on any branch. This is
// what keeps a console page render and a statusline tick out of a one-time
// irreversible directory move.
func TestStateDirPureResolutionNeverRelocates(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := seedLegacyStateDir(t, root, 2)

	path := BacklogPathForRoot(root)
	if filepath.Dir(path) != LegacyStateDirForRoot(root) {
		t.Fatalf("pure resolution = %q, want the legacy directory observed in place", filepath.Dir(path))
	}
	if dirExists(StateDirForRoot(root)) {
		t.Fatal("the pure resolution created the new directory — it must move nothing")
	}
	got := dirCensus(t, LegacyStateDirForRoot(root))
	want := append(append([]string{}, registry...), backlogFileName)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("legacy census changed to %v, want %v", got, want)
	}

	// RecordPath is a pure surface too — reading a session record must not
	// relocate the directory it lives in.
	if dir := filepath.Dir(RecordPath(root, "0000000a-1111-2222-3333-444444444444")); dir != LegacyStateDirForRoot(root) {
		t.Errorf("RecordPath resolved to %q, want the legacy directory", dir)
	}
	if dirExists(StateDirForRoot(root)) {
		t.Error("RecordPath relocated the directory")
	}
}

// First run: neither directory exists. The new name is what gets created, and
// nothing looks for a legacy directory that was never there.
func TestStateDirFirstRunUsesCurrentName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	if got := BacklogPathForRoot(root); filepath.Dir(got) != StateDirForRoot(root) {
		t.Errorf("pure first-run path = %q, want the current name", got)
	}
	path := BacklogPathForRootAdopting(root)
	if filepath.Dir(path) != StateDirForRoot(root) {
		t.Fatalf("adopting first-run path = %q, want the current name", path)
	}
	if _, _, err := NewBacklogStore(path).Add("first card"); err != nil {
		t.Fatalf("Add on a fresh root: %v", err)
	}
	if !dirExists(StateDirForRoot(root)) {
		t.Error("the current state directory was not created")
	}
	if dirExists(LegacyStateDirForRoot(root)) {
		t.Error("a legacy directory was created on a fresh root")
	}
}
