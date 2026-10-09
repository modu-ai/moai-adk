// m3_incomplete_refresh_test.go — SPEC-USERASSET-DEPLOY-GUARD-001, gate
// rounds 21(b)/22 [P1]: an interrupted pending refresh must COMPLETE on the
// retry, never wedge. The repro: v1 installed, a v2 entry recorded in the
// journal (flag-less — interrupted before the write), the retry must
// classify the untouched v1 file as an incomplete pending refresh and
// refresh it to v2 — not divergence (backing up a healthy file), not a
// collision, and never pinning the v2 hash onto the v1 bytes.
package userassets

import (
	"path/filepath"
	"testing"
	"testing/fstest"
)

// fstestMapFile adapts bytes to the fixture MapFS's file type.
func fstestMapFile(data []byte) fstest.MapFile {
	return fstest.MapFile{Data: data}
}

func TestJournalIncompletePendingRefreshCompletes(t *testing.T) {
	f := newFixture(t)
	const trackedKey = "claude-agents/manager-x.md"
	trackedAbs := filepath.Join(f.home, ".claude", "agents", "manager-x.md")

	// Run 1: v1 installed and manifest-tracked.
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("run 1: %v", err)
	}
	v1Bytes := readBytes(t, trackedAbs)

	// The interrupted v2 run: the source now ships v2, and a journal entry
	// recorded the v2 intent (flag-less — it never got to write).
	v2Bytes := []byte("---\nname: manager-x\n---\nagent body v2\n")
	mf := fstestMapFile(v2Bytes)
	f.src[".claude/agents/moai/manager-x.md"] = &mf
	seedRecoveredJournal(t, f.home, trackedKey, shaHex(v2Bytes), false)

	// Run 2: the untouched v1 file must be recognized as an incomplete
	// pending refresh and REFRESHED to v2.
	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if got := readBytes(t, trackedAbs); string(got) != string(v2Bytes) {
		t.Fatalf("the pending refresh did not complete: on-disk bytes = %q, want v2 %q", got, v2Bytes)
	}
	if res.DivergencePreserved != 0 || res.CollisionSkipped != 0 {
		t.Fatalf("the healthy v1 file was misclassified: divergences=%v collisions=%v — the hash discriminator is missing", res.Divergences, res.Collisions)
	}
	m, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatal(err)
	}
	fe, ok := m.Files[trackedKey]
	if !ok || fe.SHA256 != shaHex(v2Bytes) {
		t.Fatalf("the manifest did not record the completed refresh: %+v (want sha %s)", fe, shaHex(v2Bytes))
	}

	// Run 3: idempotence — the update must not be stuck.
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("run 3: %v", err)
	}
	res3, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("run 3b: %v", err)
	}
	if res3.DivergencePreserved != 0 || res3.Refreshed != 0 {
		t.Fatalf("the update is stuck: run 3b diverged=%d refreshed=%d — the v2 hash was pinned onto v1 bytes", res3.DivergencePreserved, res3.Refreshed)
	}
	_ = v1Bytes
}
