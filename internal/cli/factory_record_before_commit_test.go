package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

// The queue's current-dispatch record is written INSIDE the factory
// transaction that moves the binding, right before its commit (turn-end gate,
// card t1538): a record that cannot be written stops the transition instead of
// trailing a committed one, and a transition that is refused moves no record.

// fcBindingRun reports the run the card's dispatch binding currently names.
func fcBindingRun(t *testing.T, root, cardID string) string {
	t.Helper()
	row, linked, err := fcOpen(t, root).RecordedCardRowReadonly(context.Background(), cardID)
	if err != nil || !linked {
		t.Fatalf("binding of %s: linked=%v err=%v", cardID, linked, err)
	}
	return row.RunID
}

// T2 re-points the binding onto the run in its own transaction, so a record
// write that trailed it left binding=run-new beside a queue record still
// naming run-old — where a retried older dispatch reads the record as current
// and drags the binding back.
func TestReviewAssignRefreshFailureLeavesBindingUnchanged(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	abortCurrentDispatchRefresh(t, store)

	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-2", "--run", "run-new"); err == nil {
		t.Fatal("assign with an unwritable current-dispatch record succeeded")
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-old" {
		t.Fatalf("binding after the failed assign = %q, want it unmoved at run-old", run)
	}
	if c, err := fcOpen(t, root).LoadCard(ctx, "run-new", "t1"); err == nil && c.State != homestate.CardPicked {
		t.Fatalf("run-new row after the failed assign = state %s owner %q, want it left picked", c.State, c.OwnerLabel)
	}
	fcWantCurrent(t, store, "t1", "run-old", "lane-1")
}

// The binding-only assign branches (the --to-less record and the
// state-preserving re-bind) write the record FIRST: their factory write has no
// guard that can refuse after it, so an unwritable record must stop them
// before the binding moves.
func TestReviewAssignBindingBranchesRefreshFailureLeavesBindingUnchanged(t *testing.T) {
	for name, args := range map[string][]string{
		"to_less":               {"assign", "t1", "--run", "run-new"},
		"state_preserving_bind": {"assign", "t1", "--to", "lane-1", "--run", "run-new"},
	} {
		t.Run(name, func(t *testing.T) {
			root, store := fcFixture(t)
			fcQueue(t, store, factory.BacklogStatePicked)
			if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
				t.Fatalf("assign run-old: %v", err)
			}
			seedOlderDispatch(t, root, store, "t1")
			if name == "state_preserving_bind" {
				// run-new already holds the card assigned: the re-bind keeps its state.
				fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-new", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 2})
			}
			abortCurrentDispatchRefresh(t, store)

			if _, _, err := runFactory(t, args...); err == nil {
				t.Fatal("assign with an unwritable current-dispatch record succeeded")
			}
			if run := fcBindingRun(t, root, "t1"); run != "run-old" {
				t.Fatalf("binding after the failed assign = %q, want it unmoved at run-old", run)
			}
			fcWantCurrent(t, store, "t1", "run-old", "lane-1")
		})
	}
}

// A refused T2 moves nothing — neither the binding nor the record: the record
// is written only for a transition that will land.
func TestReviewAssignRefusedTransitionLeavesTheRecordUntouched(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	// run-new holds the card picked behind a predecessor that never merged.
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-new", State: homestate.CardPicked, HintAfter: "t9", Version: 1})

	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-2", "--run", "run-new"); err == nil {
		t.Fatal("assign behind an unmerged predecessor succeeded")
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-old" {
		t.Fatalf("binding after the refused assign = %q, want it unmoved at run-old", run)
	}
	fcWantCurrent(t, store, "t1", "run-old", "lane-1")
}

// A claim whose T2 is refused moves nothing either: the lane's lease section
// writes the record inside the transition, so a refused edge never reaches it.
func TestReviewFactoryNextClaimRefusedTransitionLeavesTheRecordUntouched(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	// run-new holds the card picked behind a predecessor that never merged.
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-new", State: homestate.CardPicked, HintAfter: "t9", Version: 1})

	db := fcOpen(t, root)
	cur, err := db.LoadCard(ctx, "run-new", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _, err := factoryNextClaim(ctx, db, root, "run-new", cur, "lane-1"); err == nil || ok {
		t.Fatalf("claim behind an unmerged predecessor: ok=%v err=%v, want a refusal", ok, err)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-old" {
		t.Fatalf("binding after the refused claim = %q, want it unmoved at run-old", run)
	}
	fcWantCurrent(t, store, "t1", "run-old", "lane-1")
}

// The bundle loader's head assignment is a T2 inside the chain's transaction,
// so the queue's record follows inside it: a successful load leaves the record
// on the new run, and an unwritable record rolls the whole load back.
func TestReviewBundleHeadFollowsTheRecordInsideTheLoad(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")

	if _, _, err := runFactory(t, "bundle", "lane-1", "t1", "--run", "run-new"); err != nil {
		t.Fatalf("bundle load: %v", err)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-new" {
		t.Fatalf("binding after the bundle load = %q, want run-new", run)
	}
	fcWantCurrent(t, store, "t1", "run-new", "lane-1")
}

func TestReviewBundleRefreshFailureRollsTheLoadBack(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	abortCurrentDispatchRefresh(t, store)

	if _, _, err := runFactory(t, "bundle", "lane-1", "t1", "--run", "run-new"); err == nil {
		t.Fatal("bundle load with an unwritable current-dispatch record succeeded")
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-old" {
		t.Fatalf("binding after the failed load = %q, want it unmoved at run-old", run)
	}
	if _, err := fcOpen(t, root).LoadCard(ctx, "run-new", "t1"); !errors.Is(err, homestate.ErrCardNotFound) {
		t.Fatalf("run-new holds a row for the head after the failed load (err %v): the chain must land whole or not at all", err)
	}
	fcWantCurrent(t, store, "t1", "run-old", "lane-1")
}

// A factory verb leaves a queue that never carried the record exactly as it
// was (AC-FR-021): the refresh reads the queue without adopting it, so a
// legacy JSON queue is not migrated into SQLite just to find the table absent.
func TestReviewAssignDoesNotMigrateLegacyQueue(t *testing.T) {
	root, store := fcFixture(t)
	if err := os.MkdirAll(filepath.Dir(store.Path()), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"version":1,"last_seq":1,"items":[` +
		`{"id":"t1","text":"legacy factory card","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"picked"}],` +
		`"findings":[]}` + "\n"
	if err := os.WriteFile(store.Path(), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(store.EnginePath()); err == nil {
		t.Fatalf("fixture must start with no database at %s", store.EnginePath())
	}

	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-a"); err != nil {
		t.Fatalf("assign on a legacy queue: %v", err)
	}
	if _, err := os.Stat(store.EnginePath()); err == nil {
		t.Fatalf("assign migrated the legacy queue into %s", store.EnginePath())
	}
	if raw, err := os.ReadFile(store.Path()); err != nil || string(raw) != legacy {
		t.Fatalf("legacy queue document changed: err=%v\n%s", err, raw)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-a" {
		t.Fatalf("binding after the assign = %q, want run-a", run)
	}
}

// fcKickoffWorktree writes the plan artifacts T8a's guard reads — a verdict
// that passes the shared admission predicate and a progress record carrying
// the audit-ready signal — into a fresh directory, and returns it with the
// evidence SHA the verdict audited.
func fcKickoffWorktree(t *testing.T, cardID, specID string) (dir, sha string) {
	t.Helper()
	dir = t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", specID)
	write := func(path, body string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(specDir, "spec.md"), "fixture spec\n")
	write(filepath.Join(specDir, "progress.md"),
		"# progress\n\n## §E.1 Plan-phase Audit-Ready Signal\n\n```yaml\naudit_ready: true\n```\n\n## §E.2 Run-phase Evidence\n\n_<pending run-phase>_\n")
	hash, err := runtime.NewInMemoryCache().ComputeHash(specDir)
	if err != nil {
		t.Fatalf("fixture plan hash: %v", err)
	}
	sha = "0123456789abcdef0123456789abcdef01234567"
	write(filepath.Join(dir, ".moai", "reports", cardID, "plan-audit.md"),
		"# Audit report\n\nProse summary.\n\nverdict: PASS\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: "+hash+"\naudited_sha: "+sha+"\n")
	return dir, sha
}

// fcPlaceKickoff places the card at kickoff in runID under lane-1 with a
// plan-audit that satisfies every T8a condition.
func fcPlaceKickoff(t *testing.T, root, runID string) {
	t.Helper()
	const specID = "SPEC-FIXTURE-001"
	wt, sha := fcKickoffWorktree(t, "t1", specID)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", RunID: runID, State: homestate.CardKickoff, OwnerLabel: "lane-1", Version: 2,
		DecisionGate: homestate.DecisionGateKickoff, DecisionResume: homestate.CardRun,
		WorktreePath: wt, EvidenceSHA: sha,
	})
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET spec_id=? WHERE run_id=? AND card_id=?`, specID, runID, "t1"); err != nil {
		t.Fatal(err)
	}
}

// T8a leases from a run whose row may sit behind a stale binding and re-points
// the binding onto it. The queue's record has to follow, or a retried older
// dispatch reads the stale record as current and drags the binding back — and
// the older run's approval verifies again while the new run's lease is valid.
func TestReviewAuditLeaseCannotBeRegressedByOldRetry(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	fcPlaceKickoff(t, root, "run-new")

	out, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--decider", "audit", "--run", "run-new")
	if err != nil {
		t.Fatalf("audit approval: %v\n%s", err, out)
	}
	c, err := fcOpen(t, root).LoadCard(ctx, "run-new", "t1")
	if err != nil || c.State != homestate.CardRun || c.LeaseHolder != "lane-1" {
		t.Fatalf("run-new row after the audit approval: %+v err=%v, want run leased to lane-1", c, err)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-new" {
		t.Fatalf("binding after the audit approval = %q, want run-new", run)
	}

	// The older dispatch's mirror arrives late — a retry of run-old's operation.
	if err := writeFactoryAssignment(ctx, root, store, "run-old", "t1", "lane-1"); err != nil {
		t.Fatalf("the superseded retry must be a no-op, not an error: %v", err)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-new" {
		t.Fatalf("binding after the old retry = %q, want the new run's lease to stand at run-new", run)
	}
	fcWantCurrent(t, store, "t1", "run-new", "lane-1")
}

// An unwritable record stops T8a before its commit: no lease, no binding move.
func TestReviewAuditApprovalRefreshFailureCommitsNothing(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	fcPlaceKickoff(t, root, "run-new")
	abortCurrentDispatchRefresh(t, store)

	if _, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--decider", "audit", "--run", "run-new"); err == nil {
		t.Fatal("audit approval with an unwritable current-dispatch record succeeded")
	}
	c, err := fcOpen(t, root).LoadCard(ctx, "run-new", "t1")
	if err != nil || c.State != homestate.CardKickoff || c.LeaseHolder != "" {
		t.Fatalf("run-new row after the failed approval: %+v err=%v, want it left at kickoff with no lease", c, err)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-old" {
		t.Fatalf("binding after the failed approval = %q, want it unmoved at run-old", run)
	}
	fcWantCurrent(t, store, "t1", "run-old", "lane-1")
}

// A refused audit approval moves nothing: the hold arm is read inside the
// transition, so the record must not have been touched ahead of it.
func TestReviewAuditApprovalRefusedLeavesTheRecordUntouched(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	fcPlaceKickoff(t, root, "run-new")
	// The queue item goes on hold after the dispatch: the audit decider reads
	// the hold inside the transition and refuses.
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		r.Items[0].State = factory.BacklogStateHold
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	out, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--decider", "audit", "--run", "run-new")
	if err == nil || !strings.Contains(out, "on hold") {
		t.Fatalf("audit approval of a held card: out=%q err=%v, want a hold refusal", out, err)
	}
	if run := fcBindingRun(t, root, "t1"); run != "run-old" {
		t.Fatalf("binding after the refused approval = %q, want it unmoved at run-old", run)
	}
	fcWantCurrent(t, store, "t1", "run-old", "lane-1")
}
