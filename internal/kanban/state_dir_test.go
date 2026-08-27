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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
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
