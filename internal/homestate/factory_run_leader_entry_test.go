package homestate

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

// factory_run_leader_entry_test.go — the leader-entry classification behind
// REQ-FCR-017/018 (card t1963): which active runs block an adoption, which are
// adoption candidates, and in what order the dead-stamp ones are chosen.

func lePlanIDs(rs []RunOwner) []string {
	var ids []string
	for _, r := range rs {
		ids = append(ids, r.RunID)
	}
	return ids
}

func leSetCreated(t *testing.T, db *FactoryDB, runID, createdAt string) {
	t.Helper()
	if _, err := db.DB.Exec(`UPDATE runs SET created_at=? WHERE run_id=?`, createdAt, runID); err != nil {
		t.Fatal(err)
	}
}

// The newest dead-stamp candidate comes first (created_at descending, parsed —
// text order is unsafe for a variable-length fraction), and equal timestamps
// fall to run_id descending (AC-FCR-017 variant e).
func TestPlanLeaderEntryOrdersNewestFirstWithRunIDTieBreak(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	for _, id := range []string{"run-a", "run-b", "run-c", "run-d"} {
		rbRun(t, db, id, rbPIDDead)
		rbCard(t, db, id, "c-"+id, CardRun)
	}
	// run-b and run-c share a timestamp; run-a and run-d fall in the same second
	// but run-a carries a fraction, so a text comparison (where "Z" sorts above
	// ".") would put run-d ahead of run-a.
	leSetCreated(t, db, "run-a", "2026-10-01T00:00:00.5Z")
	leSetCreated(t, db, "run-b", "2026-10-02T00:00:00Z")
	leSetCreated(t, db, "run-c", "2026-10-02T00:00:00Z")
	leSetCreated(t, db, "run-d", "2026-10-01T00:00:00Z")

	plan, err := db.PlanLeaderEntry(context.Background(), opts)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.Blocked {
		t.Fatalf("plan blocked by %v, want free", lePlanIDs(plan.Blockers))
	}
	if got, want := lePlanIDs(plan.Stamped), []string{"run-c", "run-b", "run-a", "run-d"}; !slices.Equal(got, want) {
		t.Fatalf("stamped order = %v, want %v (newest first, run_id descending on a tie)", got, want)
	}
}

// A live or an indeterminate active run blocks the adoption: it may be a live
// leader, and adopting a dead run beside it would split the factory.
func TestPlanLeaderEntryIsBlockedByLiveOrIndeterminateRuns(t *testing.T) {
	for _, tc := range []struct {
		name string
		pid  int
	}{{"live", rbPIDLive}, {"indeterminate", rbPIDIndeterminate}} {
		t.Run(tc.name, func(t *testing.T) {
			db := openSandboxFactory(t)
			opts := rbOpts()
			rbRun(t, db, "run-dead", rbPIDDead)
			rbCard(t, db, "run-dead", "c1", CardRun)
			rbRun(t, db, "run-other", tc.pid) // no records at all: blocking does not depend on them
			plan, err := db.PlanLeaderEntry(context.Background(), opts)
			if err != nil {
				t.Fatalf("plan: %v", err)
			}
			if !plan.Blocked || !slices.Equal(lePlanIDs(plan.Blockers), []string{"run-other"}) {
				t.Fatalf("plan = %+v, want blocked by run-other", plan)
			}
			if !slices.Equal(lePlanIDs(plan.Candidates), []string{"run-dead"}) {
				t.Fatalf("candidates = %v, want run-dead listed for the warn path", lePlanIDs(plan.Candidates))
			}
		})
	}
}

// A candidate needs live records; only a dead-on-its-own-stamp one is Stamped.
// A run proved dead through the peer or the boot proof is a candidate for the
// warn path and never an adoption or consolidation source.
func TestPlanLeaderEntryStampedExcludesPeerAndBootProofs(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	rbRun(t, db, "run-stamp", rbPIDDead)
	rbCard(t, db, "run-stamp", "c1", CardRun)
	rbRun(t, db, "run-done", rbPIDDead)
	rbCard(t, db, "run-done", "c2", CardDone) // no live record: not a candidate at all
	rbRun(t, db, "run-peer", rbPIDDead)
	rbCard(t, db, "run-peer", "c3", CardRun)
	rbRun(t, db, "run-boot", rbPIDDead)
	rbCard(t, db, "run-boot", "c4", CardRun)
	// Rebuild the two proofs' identity shapes: a (0,'') stamp read through the
	// peer fallback, and an identity-less run proved dead by the boot proof.
	for _, id := range []string{"run-peer", "run-boot"} {
		if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=0, lead_process_start='' WHERE run_id=?`, id); err != nil {
			t.Fatal(err)
		}
	}
	opts.Fallback = func(runID string) (int, string, bool) {
		if runID == "run-peer" {
			return rbPIDDead, "peer-fp", true
		}
		return 0, "", false
	}
	// The host "booted" an hour from now, so every timestamp the run recorded
	// predates the boot — the boot proof holds for the run with no broker file.
	opts.BootTime = func() (time.Time, bool) { return time.Now().Add(time.Hour), true }
	opts.LegacyPeerRecordAbsent = func(runID string) bool { return runID == "run-boot" }

	plan, err := db.PlanLeaderEntry(context.Background(), opts)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.Blocked {
		t.Fatalf("plan blocked by %v, want free", lePlanIDs(plan.Blockers))
	}
	if got, want := lePlanIDs(plan.Stamped), []string{"run-stamp"}; !slices.Equal(got, want) {
		t.Fatalf("stamped = %v, want %v", got, want)
	}
	cands := lePlanIDs(plan.Candidates)
	slices.Sort(cands)
	if want := []string{"run-boot", "run-peer", "run-stamp"}; !slices.Equal(cands, want) {
		t.Fatalf("candidates = %v, want %v", cands, want)
	}
}

// AdoptRun reports a lost compare-and-set as a typed error — the launcher
// separates a lost race from a database failure — and keeps its message.
func TestAdoptRunStampLossIsTyped(t *testing.T) {
	db := openSandboxFactory(t)
	rbRun(t, db, "run-old", rbPIDDead)
	rbCard(t, db, "run-old", "c1", CardRun)
	owner := rbOwner(t, db, rbOpts(), "run-old")
	if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=?, lead_process_start='winner' WHERE run_id='run-old'`, rbPIDLive); err != nil {
		t.Fatal(err)
	}
	err := db.AdoptRun(context.Background(), owner, 4242, "loser-fp", "test-basis")
	if !errors.Is(err, ErrAdoptStampChanged) {
		t.Fatalf("err = %v, want ErrAdoptStampChanged", err)
	}
	if !strings.Contains(err.Error(), "no longer carries the classified owner stamp") {
		t.Fatalf("message %q lost its original text", err.Error())
	}
	// A database failure is not a lost race.
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	err = db.AdoptRun(context.Background(), owner, 4242, "loser-fp", "test-basis")
	if err == nil || errors.Is(err, ErrAdoptStampChanged) {
		t.Fatalf("closed-database err = %v, want a non-stamp error", err)
	}
}
