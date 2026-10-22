package worktree

// migrate_sync_fix_test.go — card t1640 M8S: the sync-audit iter1 repairs
// (sync-audit-iter1.md FAIL 0.62; findings F1-F5, F7-F9). Each test pins the
// repaired behavior of one confirmed finding, RED-first against the tree the
// audit judged:
//
//	F1  runUpdateWorktreeMigration routes through the shared engine: the
//	    update consolidation selects from the M1 classifier (agent-* never
//	    moved), lands trees at the home-root destination, and the AC-WHR-019
//	    grep cell goes green (`Join(root, sessionWorktreeSubdir)` gone).
//	F2  --rollback refuses an identity mismatch with a refused line and a
//	    non-zero exit, and propagates reverse-step errors (no `restored` on
//	    a failed restore).
//	F3  the journal is written per-tree BEFORE the move (a pending entry
//	    precedes any move), and a journal write error fails the run.
//	F5  a failed-reverse pending line protects the tree at the path it
//	    actually occupies (MovedBy answers true at the destination).
//	F7  the manifest digest covers tracked-change content (a content
//	    mutation that preserves the status shape is detected).
//	F8  an unreadable session registry is a skip cause (fail-closed), not a
//	    silently-ignored probe.
//	F9  worktreeroot.Plan / RecordedKey / OwnerMarkerOf carry package-local
//	    tests (the Craft coverage finding).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/session"
)

// sfxWorld is the M8S rehearsal world: the m6 world plus the update
// fixtures the F1 route needs.
func sfxWorld(t *testing.T) *m6World {
	t.Helper()
	return m6NewWorld(t)
}

// TestUpdateConsolidation_MovesToHomeRootAndSkipsAgent — F1: the routine
// update's consolidation selects from the M1 classifier and lands trees at
// the home-root destination; agent-* trees are never selected; the move
// list records the update trigger.
func TestUpdateConsolidation_MovesToHomeRootAndSkipsAgent(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)

	buf := &strings.Builder{}
	if err := runUpdateConsolidation(w.repo, false, buf); err != nil {
		t.Fatalf("update consolidation: %v\n%s", err, buf.String())
	}
	// The eligible legacy tree landed at the home-root destination.
	dest := filepath.Join(w.homeRoot, "legacy6")
	if fi, statErr := os.Lstat(dest); statErr != nil || !fi.IsDir() {
		t.Errorf("the update did not move legacy6 to the home root %s (stat err=%v)", dest, statErr)
	}
	if m6Registered(t, w.repo, w.movedTree) {
		t.Errorf("legacy6 is still registered at the legacy path %s", w.movedTree)
	}
	// The agent-* tree was never selected (REQ-WHR-024's red premise).
	if !m6Registered(t, w.repo, w.agentTree) {
		t.Errorf("the update moved the agent-* tree %s; the classifier never selects it", w.agentTree)
	}
	// The move list carries the update trigger with the moved line.
	mlPath := filepath.Join(w.repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	raw, rerr := os.ReadFile(mlPath)
	if rerr != nil {
		t.Fatalf("read move list: %v", rerr)
	}
	sawMovedUpdate := false
	for _, line := range strings.Split(string(raw), "\n") {
		var rec struct {
			Result  string `json:"result"`
			Trigger string `json:"trigger"`
		}
		if jerr := json.Unmarshal([]byte(line), &rec); jerr == nil && rec.Result == "moved" {
			if rec.Trigger != "update" {
				t.Errorf("the moved line carries trigger %q, want update:\n%s", rec.Trigger, line)
			}
			sawMovedUpdate = true
		}
	}
	if !sawMovedUpdate {
		t.Errorf("no moved line with the update trigger in the move list:\n%s", raw)
	}
	// The card record bound to the moved tree is rebound (AC-WHR-019(g)).
	db, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("open factory store: %v", err)
	}
	oldPath := w.movedTree
	if _, perr := db.RecordPicked(context.Background(), "run-sfx", "sfx-card", homestate.CardFields{WorktreePath: &oldPath}, "sfx-test", time.Now().UTC()); perr != nil {
		t.Fatalf("record picked: %v", perr)
	}
	_ = db.Close()
	w2 := sfxWorld(t)
	m4InRepo(t, w2.repo)
	db2, err := homestate.OpenFactory(w2.repo)
	if err != nil {
		t.Fatalf("open factory store 2: %v", err)
	}
	p2 := w2.movedTree
	if _, perr := db2.RecordPicked(context.Background(), "run-sfx2", "sfx-card2", homestate.CardFields{WorktreePath: &p2}, "sfx-test", time.Now().UTC()); perr != nil {
		t.Fatalf("record picked 2: %v", perr)
	}
	_ = db2.Close()
	buf2 := &strings.Builder{}
	if err := runUpdateConsolidation(w2.repo, false, buf2); err != nil {
		t.Fatalf("second world consolidation: %v\n%s", err, buf2.String())
	}
	db3, err := homestate.OpenFactory(w2.repo)
	if err != nil {
		t.Fatalf("reopen factory store 2: %v", err)
	}
	defer func() { _ = db3.Close() }()
	card, lerr := db3.LoadCard(context.Background(), "run-sfx2", "sfx-card2")
	if lerr != nil {
		t.Fatalf("load card: %v", lerr)
	}
	if card.WorktreePath == p2 || card.WorktreePath == "" {
		t.Errorf("the update-time card rebind did not land: %q", card.WorktreePath)
	}
}

// TestUpdateConsolidation_DryRunIsByteIdentical — F1's dry-run arm and
// AC-WHR-019(h): the update preview prints the plan and writes nothing.
func TestUpdateConsolidation_DryRunIsByteIdentical(t *testing.T) {
	repo, tree, home := rfxLegacyWorld(t)
	m4InRepo(t, repo)

	repoBefore := rfxWalkHash(t, repo)
	homeBefore := rfxWalkHash(t, home)
	buf := &strings.Builder{}
	if err := runUpdateConsolidation(repo, true, buf); err != nil {
		t.Fatalf("update dry run: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "rfx-legacy") {
		t.Errorf("the update preview does not name the eligible tree:\n%s", buf.String())
	}
	if after := rfxWalkHash(t, repo); after != repoBefore {
		t.Errorf("the update dry run wrote inside the repository (hash changed)")
	}
	if after := rfxWalkHash(t, home); after != homeBefore {
		t.Errorf("the update dry run wrote inside MOAI_HOME (hash changed)")
	}
	_ = tree
}

// TestRollback_RefusesIdentityMismatch — F2: a replaced destination is a
// refused step with the reason and a non-zero exit, never a silent skip.
func TestRollback_RefusesIdentityMismatch(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	journalPath := sfxLatestJournal(t, w.repo)
	if journalPath == "" {
		t.Fatalf("no journal written by the apply")
	}
	// Replace the destination with a same-shaped stranger.
	dest := filepath.Join(w.homeRoot, "legacy6")
	ident, ierr := TreeIdent(dest)
	if ierr != nil {
		t.Fatalf("ident of the moved tree: %v", ierr)
	}
	_ = ident
	if rerr := os.RemoveAll(dest); rerr != nil {
		t.Fatalf("remove the moved tree: %v", rerr)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("recreate the destination as a stranger: %v", err)
	}

	out, err := m4RunMigrate(t, "--rollback", journalPath)
	if err == nil {
		t.Errorf("rollback exits 0 with a replaced destination; output:\n%s", out)
	}
	if !strings.Contains(out, "refused") || !strings.Contains(out, "identity mismatch") {
		t.Errorf("rollback does not report the identity mismatch refusal:\n%s", out)
	}
}

// TestRollback_PropagatesReverseStepErrors — F2: a failed reverse re-key or
// rebind leaves the step pending (no `restored` line) and fails the run.
func TestRollback_PropagatesReverseStepErrors(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	journalPath := sfxLatestJournal(t, w.repo)
	if journalPath == "" {
		t.Fatalf("no journal written by the apply")
	}
	// Corrupt the state file the reverse re-key reads.
	statePath := filepath.Join(w.repo, ".moai", "state", "worktrees.json")
	if werr := os.WriteFile(statePath, []byte("not json at all\n"), 0o644); werr != nil {
		t.Fatalf("corrupt the state file: %v", werr)
	}

	out, err := m4RunMigrate(t, "--rollback", journalPath)
	if err == nil {
		t.Errorf("rollback exits 0 although the reverse step failed; output:\n%s", out)
	}
	if strings.Contains(out, "restored\t") {
		t.Errorf("rollback reports restored for a step whose reverse failed:\n%s", out)
	}
}

// sfxLatestJournal returns the newest migration journal path of root.
func sfxLatestJournal(t *testing.T, root string) string {
	t.Helper()
	mlDir := filepath.Join(root, ".moai", "state", "worktree-migrate")
	entries, err := os.ReadDir(mlDir)
	if err != nil {
		return ""
	}
	best, bestMod := "", time.Time{}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") || e.Name() == "move-list.jsonl" {
			continue
		}
		info, ierr := e.Info()
		if ierr == nil && info.ModTime().After(bestMod) {
			best, bestMod = filepath.Join(mlDir, e.Name()), info.ModTime()
		}
	}
	return best
}

// TestJournal_PendingEntryPrecedesTheMove — F3: the journal records each
// tree as pending BEFORE its move, so a killed run leaves a rollback-able
// journal; and a journal write error fails the run instead of being
// discarded.
func TestJournal_PendingEntryPrecedesTheMove(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)

	var writes []string
	orig := journalWrite
	journalWrite = func(path string, data []byte) error {
		writes = append(writes, string(data))
		return orig(path, data)
	}
	t.Cleanup(func() { journalWrite = orig })

	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	sawPending := false
	for _, w := range writes {
		if strings.Contains(w, `"outcome": "pending"`) || strings.Contains(w, `"result": "pending"`) {
			sawPending = true
		}
	}
	if !sawPending {
		t.Errorf("no journal write ever carried a pending entry; the journal is written after the moves (%d writes)", len(writes))
	}
}

// TestJournal_WriteErrorFailsTheRun — F3: a journal write error is not
// discarded; the run exits non-zero.
func TestJournal_WriteErrorFailsTheRun(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)

	orig := journalWrite
	journalWrite = func(path string, data []byte) error {
		return os.ErrPermission
	}
	t.Cleanup(func() { journalWrite = orig })

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Errorf("the apply exits 0 although every journal write failed; output:\n%s", out)
	}
	if !strings.Contains(out, "journal") && !strings.Contains(err.Error(), "journal") {
		t.Errorf("the failure does not name the journal: %v", err)
	}
}

// TestMovedBy_ProtectsAFailedReversePendingTree — F5: after a failed
// reverse move the pending line protects the tree at the destination (the
// path it actually occupies), until a later line releases it.
func TestMovedBy_ProtectsAFailedReversePendingTree(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)

	dbPath, err := homestate.FactoryDBPath(w.repo)
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("mkdir factory dir: %v", err)
	}
	if werr := os.WriteFile(dbPath, []byte("not a sqlite database\n"), 0o644); werr != nil {
		t.Fatalf("write the corrupt factory db: %v", werr)
	}
	origReverse := reverseGitMove
	reverseGitMove = func(root, from, to string) error { return os.ErrPermission }
	t.Cleanup(func() { reverseGitMove = origReverse })

	if out, aerr := m4RunMigrate(t, "--apply"); aerr == nil {
		t.Fatalf("the apply unexpectedly succeeded:\n%s", out)
	}
	dest := filepath.Join(w.homeRoot, "legacy6")
	if !m6Registered(t, w.repo, dest) {
		t.Fatalf("the tree did not stay registered at its destination")
	}
	if !MovedBy(w.repo, dest) {
		t.Errorf("MovedBy = false at the destination after a failed reverse move; the cleanup protection window is open")
	}
}

// TestCaptureManifest_DetectsTrackedContentMutation — F7: the manifest
// digest covers tracked-change content, not just the status shape.
func TestCaptureManifest_DetectsTrackedContentMutation(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	_ = repo
	if err := os.WriteFile(filepath.Join(tree, "README.md"), []byte("tracked-a\n"), 0o644); err != nil {
		t.Fatalf("modify the tracked file: %v", err)
	}
	before := reviewCaptureManifest(t, tree)
	if err := os.WriteFile(filepath.Join(tree, "README.md"), []byte("tracked-b\n"), 0o644); err != nil {
		t.Fatalf("re-modify the tracked file: %v", err)
	}
	after := reviewCaptureManifest(t, tree)
	if before == after {
		t.Errorf("the manifest digest is blind to a tracked-content mutation that preserves the status shape")
	}
}

// TestMigrationSkipsOnUnreadableRegistry — F8: a session registry that
// cannot be read is a skip cause (fail-closed, I-1), not a voided probe.
func TestMigrationSkipsOnUnreadableRegistry(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	// Corrupt the registry under BOTH roots the readability check reads.
	for _, base := range []string{w.repo, w.movedTree} {
		regPath := filepath.Join(base, session.DefaultRegistryPath)
		if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
			t.Fatalf("mkdir registry dir: %v", err)
		}
		if werr := os.WriteFile(regPath, []byte("{broken json\n"), 0o644); werr != nil {
			t.Fatalf("corrupt the registry: %v", werr)
		}
	}
	out, err := m4RunMigrate(t, "--apply")
	if err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	if !strings.Contains(out, w.movedTree) || !strings.Contains(out, "registry") {
		t.Errorf("the tree with an unreadable registry was not skipped with a registry reason:\n%s", out)
	}
	if _, statErr := os.Lstat(filepath.Join(w.homeRoot, "legacy6")); statErr == nil {
		t.Errorf("legacy6 was moved despite the unreadable registry; the skip cause did not fire")
	}
}

// TestSessionRegistryProbeIsTheRealCheck — F8: the engine's default probe
// wires the registry readability check the old planner used (no stub).
func TestSessionRegistryProbeIsTheRealCheck(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if werr := os.WriteFile(filepath.Join(dir, session.DefaultRegistryPath), []byte("{broken\n"), 0o644); werr != nil {
		t.Fatalf("write broken registry: %v", werr)
	}
	if err := worktreeRegistryProbe(dir, dir); err == nil {
		t.Errorf("the default registry probe accepts an unreadable registry; the skip cause is dead code")
	}
}
