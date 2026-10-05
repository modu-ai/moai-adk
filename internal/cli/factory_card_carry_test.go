// factory_card_carry_test.go — card t1521: the card↔worktree binding carries
// over into a replacement run. When a factory run is replaced, the new run's
// card row is born with an empty worktree_path, and `factory next` refused
// the card's own surviving tree as foreign (the REQ-SD-011 refusal firing on
// the card's own tree, observed 2026-10-05 on t1453). The carry: a landing
// directory that a PREVIOUS run's record of the SAME card names is
// re-recorded on the new run's row at the lease boundary; every other shape —
// a directory no record of this card ever named, a previous binding naming a
// different path, a tree that did not survive — keeps the existing behavior.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	carryPrevRun = "run-old"
	carryNewRun  = "run-new"
)

// carryPlaceRunRow inserts a runs row with an explicit created_at, so the
// previous run reads older than the current one deterministically.
func carryPlaceRunRow(t *testing.T, root, runID, createdAt string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,lane_capacity,created_at,updated_at) VALUES(?,?,?,'retired','{}',0,'',1,?,?)`,
		runID, "", "", createdAt, createdAt); err != nil {
		t.Fatalf("place run %s: %v", runID, err)
	}
}

// carryCard loads one card row of one run.
func carryCard(t *testing.T, root, runID, cardID string) homestate.Card {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s/%s: %v", runID, cardID, err)
	}
	return c
}

// carryFixture builds the run-replacement fixture: a retired previous run
// whose record carries binding rows, and the current run the verb leases on.
// The root is canonicalized the way the verb's project-root resolution is, so
// a recorded binding shares the landing directory's spelling.
func carryFixture(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := fcFixture(t)
	if real, err := filepath.EvalSymlinks(root); err == nil {
		root = real
	}
	carryPlaceRunRow(t, root, carryPrevRun, "2026-10-04T00:00:00Z")
	carryPlaceRunRow(t, root, carryNewRun, "2026-10-05T00:00:00Z")
	return root, store
}

// carryLanding is the card's landing directory under the shared materializer
// root — the directory the REQ-SD-011 refusal checks.
func carryLanding(root, cardID string) string {
	return filepath.Join(sdWorktreesDir(root), cardID)
}

// TestFactoryNextCarriesWorktreeBindingAcrossRunReplacement is the t1521
// shape: the previous run leased t1 into the landing tree and recorded it; a
// replacement run leases t1 again — the tree must be re-entered, not refused
// and not re-created.
func TestFactoryNextCarriesWorktreeBindingAcrossRunReplacement(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	landing := carryLanding(root, "t1")
	if err := os.MkdirAll(sdWorktreesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "worktree", "add", "-q", landing, "-b", "WT-previous-tree")
	fcPlace(t, root, homestate.Card{RunID: carryPrevRun, CardID: "t1", State: homestate.CardDone, WorktreePath: landing})
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	out, stderr, err := runFactory(t, "next", "--run", carryNewRun)
	if err != nil {
		t.Fatalf("next over the card's own previous tree: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	if got := carryCard(t, root, carryNewRun, "t1").WorktreePath; !sameDirPath(got, landing) {
		t.Errorf("run-new binding = %q, want the carried previous tree %q", got, landing)
	}
	if entries, err := os.ReadDir(sdWorktreesDir(root)); err != nil {
		t.Fatalf("read the landing root: %v", err)
	} else if len(entries) != 1 {
		t.Errorf("landing root holds %d directories, want only the card's own tree", len(entries))
	}
}

// TestFactoryNextNominateCarriesWorktreeBindingAcrossRunReplacement pins the
// nominated lease on the same shape: the read-only validation must not refuse
// foreign-worktree for a tree the card's own previous run recorded.
func TestFactoryNextNominateCarriesWorktreeBindingAcrossRunReplacement(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	landing := carryLanding(root, "t1")
	if err := os.MkdirAll(sdWorktreesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "worktree", "add", "-q", landing, "-b", "WT-previous-tree")
	fcPlace(t, root, homestate.Card{RunID: carryPrevRun, CardID: "t1", State: homestate.CardDone, WorktreePath: landing})
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	out, stderr, err := runFactory(t, "next", "--run", carryNewRun, "--card", "t1")
	if err != nil {
		t.Fatalf("nominated next over the card's own previous tree: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	if got := carryCard(t, root, carryNewRun, "t1").WorktreePath; !sameDirPath(got, landing) {
		t.Errorf("run-new binding = %q, want the carried previous tree %q", got, landing)
	}
}

// TestFactoryNextRefusesAnotherCardsPreviousTree pins the invariant the
// carry-over must not weaken: an existing landing directory that only
// ANOTHER card's record names stays foreign — the card never adopts another
// card's tree, in this run or across runs.
func TestFactoryNextRefusesAnotherCardsPreviousTree(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued)
	foreign := carryLanding(root, "t1")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	// The previous run's records name only t2's own tree, never t1's landing.
	other := filepath.Join(sdWorktreesDir(root), "t2")
	if err := os.MkdirAll(sdWorktreesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "worktree", "add", "-q", other, "-b", "WT-other-tree")
	fcPlace(t, root, homestate.Card{RunID: carryPrevRun, CardID: "t2", State: homestate.CardDone, WorktreePath: other})
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	_, _, err := runFactory(t, "next", "--run", carryNewRun)
	if err == nil || !strings.Contains(err.Error(), foreign) {
		t.Fatalf("next over a foreign directory: err = %v, want a refusal naming %s", err, foreign)
	}
	if c := carryCard(t, root, carryNewRun, "t1"); c.State != homestate.CardPicked || c.WorktreePath != "" {
		t.Errorf("card row changed on the refusal: state %s worktree %q", c.State, c.WorktreePath)
	}
}

// TestFactoryNextRefusesWhenPreviousBindingNamesAnotherPath pins the tight
// carry shape: a surviving previous binding that names a DIFFERENT directory
// than the existing landing one explains nothing about that directory — the
// refusal stands and the binding is not carried.
func TestFactoryNextRefusesWhenPreviousBindingNamesAnotherPath(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	own := filepath.Join(root, "own-wt")
	fcGit(t, root, "worktree", "add", "-q", own, "-b", "WT-own-tree")
	unexplained := carryLanding(root, "t1")
	if err := os.MkdirAll(unexplained, 0o755); err != nil {
		t.Fatal(err)
	}
	fcPlace(t, root, homestate.Card{RunID: carryPrevRun, CardID: "t1", State: homestate.CardDone, WorktreePath: own})
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	_, _, err := runFactory(t, "next", "--run", carryNewRun)
	if err == nil || !strings.Contains(err.Error(), unexplained) {
		t.Fatalf("next over an unexplained landing directory: err = %v, want a refusal naming %s", err, unexplained)
	}
	if c := carryCard(t, root, carryNewRun, "t1"); c.WorktreePath != "" {
		t.Errorf("binding carried despite naming another path: %q", c.WorktreePath)
	}
}

// TestFactoryNextCreatesFreshTreeWhenPreviousTreeIsGone pins the survival
// gate: a previous binding whose tree did not survive carries nothing — the
// lease proceeds and the materializer creates a fresh tree at the landing.
func TestFactoryNextCreatesFreshTreeWhenPreviousTreeIsGone(t *testing.T) {
	root, store := carryFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	gone := carryLanding(root, "t1")
	fcPlace(t, root, homestate.Card{RunID: carryPrevRun, CardID: "t1", State: homestate.CardDone, WorktreePath: gone})
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	out, stderr, err := runFactory(t, "next", "--run", carryNewRun)
	if err != nil {
		t.Fatalf("next with a dead previous tree: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	info, statErr := os.Stat(gone)
	if statErr != nil || !info.IsDir() {
		t.Fatalf("fresh card worktree %s does not exist: %v", gone, statErr)
	}
	if got := carryCard(t, root, carryNewRun, "t1").WorktreePath; !sameDirPath(got, gone) {
		t.Errorf("run-new binding = %q, want the freshly created tree %q", got, gone)
	}
}
