package homestate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func frReadAbs(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// AC-FDA-009 (card-transition site): T7 applies the shared plan-phase
// predicate — a PASS label without the must-pass/blocking/hash fields, or
// with a failed must-pass, is refused — while T13 keeps the label-only check.
func TestFDA_T7AppliesThePlanPredicateAndT13StaysLabelOnly(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()

	place := func(id, state string) Card {
		c := frLeasedCard(repo, id, state)
		c.EvidenceSHA = repo.Commit
		frPlace(t, db, c)
		return c
	}
	write := func(id, name, body string) {
		frWrite(t, filepath.Join(repo.Dir, ".moai", "reports", id, name), body)
	}

	bare := place("bare-pass", CardPlanAudit)
	write("bare-pass", "plan-audit.md", "verdict: PASS\naudited_sha: "+repo.Commit+"\n")
	if _, err := db.Transition(ctx, frHolderRequest(bare, CardKickoff)); err == nil || !strings.Contains(err.Error(), "score") {
		t.Fatalf("label-only PASS at T7: err=%v, want a refusal naming the missing field", err)
	}

	mp := place("must-pass-failed", CardPlanAudit)
	frWriteVerdict(t, repo.Dir, "must-pass-failed", "plan-audit.md", "PASS", repo.Commit)
	raw := frReadAbs(t, filepath.Join(repo.Dir, ".moai", "reports", "must-pass-failed", "plan-audit.md"))
	write("must-pass-failed", "plan-audit.md", strings.Replace(raw, "must_pass_failed: 0", "must_pass_failed: 1", 1))
	if _, err := db.Transition(ctx, frHolderRequest(mp, CardKickoff)); err == nil || !strings.Contains(err.Error(), "must_pass_failed 1") {
		t.Fatalf("must-pass FAIL with a PASS label at T7: err=%v, want refusal", err)
	}

	ok := place("full-pass", CardPlanAudit)
	frWriteVerdict(t, repo.Dir, "full-pass", "plan-audit.md", "PASS", repo.Commit)
	if got, err := db.Transition(ctx, frHolderRequest(ok, CardKickoff)); err != nil || got.State != CardKickoff {
		t.Fatalf("full PASS at T7: state=%s err=%v, want kickoff", got.State, err)
	}

	sync := place("sync-label", CardSyncAudit)
	write("sync-label", "sync-audit.md", "verdict: PASS-WITH-DEBT\naudited_sha: "+repo.Commit+"\n")
	if got, err := db.Transition(ctx, frHolderRequest(sync, CardMergeReady)); err != nil || got.State != CardMergeReady {
		t.Fatalf("label-only PASS-WITH-DEBT at T13: state=%s err=%v, want merge-ready (unchanged)", got.State, err)
	}
}
