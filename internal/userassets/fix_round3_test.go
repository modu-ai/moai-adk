// fix_round3_test.go — fix-round-3 reproducing tests (card t1509): items
// 1-8 from the gate's 8-finding dispatch.
package userassets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Items 3+4: concurrent acquireGuard calls from different HOMEs each get
// their own release closure — releasing one does not release the other.
func TestFR3_34_GuardHandlesAreIndependent(t *testing.T) {
	t.Parallel()
	homeA := t.TempDir()
	homeB := t.TempDir()
	_ = os.MkdirAll(filepath.Join(homeA, ".moai"), 0o755)
	_ = os.MkdirAll(filepath.Join(homeB, ".moai"), 0o755)
	relA, err := acquireGuard(filepath.Join(homeA, ".moai", "x.lock"), 100*time.Millisecond)
	if err != nil {
		t.Fatalf("guard A: %v", err)
	}
	relB, err := acquireGuard(filepath.Join(homeB, ".moai", "y.lock"), 100*time.Millisecond)
	if err != nil {
		t.Fatalf("guard B: %v", err)
	}
	relA() // release A — must NOT release B's flock
	// B's guard still held: a second acquire on B's path within the deadline
	// must fail (the flock is still held by relB's fd).
	if _, err := acquireGuard(filepath.Join(homeB, ".moai", "y.lock"), 30*time.Millisecond); err == nil {
		t.Error("Items 3+4: releasing A released B's flock (global handle race)")
	}
	relB()
	// After B's release the path is acquirable again.
	l, err := acquireGuard(filepath.Join(homeB, ".moai", "y.lock"), 100*time.Millisecond)
	if err != nil {
		t.Fatalf("re-acquire after B release: %v", err)
	}
	l()
}

// Item 4: the guard acquisition respects the SAME deadline — a held guard
// fails within the timeout instead of blocking indefinitely.
func TestFR3_4_GuardRespectsDeadline(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	path := filepath.Join(home, ".moai", "x.lock")
	_ = os.MkdirAll(filepath.Join(home, ".moai"), 0o755)
	rel, err := acquireGuard(path, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	started := time.Now()
	if _, err := acquireGuard(path, 80*time.Millisecond); err == nil {
		t.Error("second acquire while held succeeded")
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Errorf("guard waited %v — the deadline is not respected (blocking flock)", elapsed)
	}
	rel()
}

// Item 7: the flag-complete divergence arm KEEPS the manifest record with
// the journal's original hash + provenance — after journal cleanup the file
// is tracked as REQ-023 divergence, not flipped to a collision.
func TestFR3_7_DivergenceKeepsManifestRecord(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if err := os.MkdirAll(MoaiHome(f.home), 0o755); err != nil {
		t.Fatal(err)
	}
	divergeRel := "claude-skills/moai-alpha/workflows/a.md"
	if err := os.MkdirAll(filepath.Join(f.home, ".claude/skills/moai-alpha/workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.home, ".claude/skills/moai-alpha/workflows/a.md"), []byte("user edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	j := &PendingJournal{SchemaVersion: 1, Entries: []JournalEntry{{
		Path: divergeRel, ExpectedSHA256: sha256Hex([]byte("workflow a\n")),
		Bundle: "core", MoaiVersion: "vOld", InstalledAt: "t0", WriteCompleted: true,
	}}}
	if err := WriteJournal(JournalPath(f.home), j); err != nil {
		t.Fatal(err)
	}
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatal(err)
	}
	// The journal is cleared (manifest saved) — the file must STILL be
	// tracked (divergence record kept), so the NEXT run classifies it as
	// REQ-023 divergence, not a collision.
	if err := ClearJournal(JournalPath(f.home)); err != nil {
		t.Fatal(err)
	}
	in := f.installer(t)
	m, _ := Load(ManifestPath(f.home))
	res, err := in.PruneUnselected(m)
	_ = res
	if err != nil {
		t.Fatal(err)
	}
	if _, tracked := m.Files[divergeRel]; !tracked {
		t.Error("Item 7: divergence record dropped — the file flips to collision after journal cleanup")
	}
}
