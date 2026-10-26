package cli

// factory_card_rebind_stageless_test.go — card t1964: two residual defects in
// the completion chain for cards that landed through the PR path (so they never
// passed `factory stage`):
//
//	G1  `factory rebind-evidence` inferred the audit phase from the row's stage
//	    before it ever looked at the landing record, so a row with state
//	    leased/assigned and stage "" was refused with "cannot infer the audit
//	    phase" although the operator had recorded the landing.
//	G2  the done gate reads the receipt against the card's CURRENT run row (the
//	    dispatch binding); with G1 the only rebindable row was an older run's, and
//	    its receipt failed the run axis ("receipt binds run X, the card's run is Y").
//	    The receipt gate stays run-bound on purpose (REQ-FCR-001); the repair is
//	    that the current row is rebindable, and that approve says so when a
//	    receipt is issued for a run the done gate does not read.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// emptyStageRow turns card t1 into the PR-path shape: state leased, no stage.
func emptyStageRow(t *testing.T, root string) {
	t.Helper()
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET state=?, stage='' WHERE card_id='t1' AND run_id=?`,
		string(homestate.CardLeased), fcRun); err != nil {
		t.Fatalf("empty the stage: %v", err)
	}
	_ = db.Close()
}

func TestFactoryRebindEvidenceStageEmptyRow(t *testing.T) {
	t.Run("rebinds_from_the_recorded_landing_without_inferring_a_phase", func(t *testing.T) {
		root, _, _, squash, _ := squashLandedFixture(t, "squash")
		emptyStageRow(t, root)
		if out, err := rebindRun(t, "PR-path card: no stage, the landing is recorded"); err != nil {
			t.Fatalf("rebind-evidence on a stage-empty row: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want the landing commit %s", got, squash)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 1 || rows[0].Source != homestate.EvidenceSourceLanding {
			t.Fatalf("audit rows = %+v, want one row with source %q", rows, homestate.EvidenceSourceLanding)
		}
	})

	t.Run("still_refuses_without_a_landing_record", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "no stage and no landing")
		if err == nil || !strings.Contains(err.Error(), "cannot infer the audit phase") {
			t.Fatalf("err = %v, want the audit-phase refusal", err)
		}
		if after := fcCard(t, root, "t1"); after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("a refusal changed the record: %q v%d → %q v%d", before.EvidenceSHA, before.Version, after.EvidenceSHA, after.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("still_refuses_a_recorded_landing_that_is_not_on_the_landed_ref", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "unlanded")
		emptyStageRow(t, root)
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "recorded sha never reached the landed ref")
		if err == nil || !strings.Contains(err.Error(), "recorded landing commit refused") {
			t.Fatalf("err = %v, want the recorded-landing refusal", err)
		}
		if after := fcCard(t, root, "t1"); after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("a refusal changed the record: %q v%d → %q v%d", before.EvidenceSHA, before.Version, after.EvidenceSHA, after.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})
}

// twoRunRows adds the older run's merge-ready row beside the current run's
// stage-empty row, the shape the leader measured on t1632/t1649/t1807/….
func twoRunRows(t *testing.T, root string) {
	t.Helper()
	fcPlace(t, root, homestate.Card{
		RunID: "run-old", CardID: "t1", State: homestate.CardMergeReady, Stage: homestate.CardMergeReady,
		OwnerLabel: "lane-1", EvidenceSHA: "0ld0ld0ld0ld0ld0ld0ld0ld0ld0ld0ld0ld0ld",
	})
}

func TestCompletionChainAcrossRunRows(t *testing.T) {
	t.Run("rebind_and_approve_on_the_current_run_row_satisfy_the_done_gate", func(t *testing.T) {
		root, uuid, _, _, _ := squashLandedFixture(t, "squash")
		emptyStageRow(t, root)
		twoRunRows(t, root)
		if out, err := rebindRun(t, "current run row, PR path"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if _, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead"); err != nil {
			t.Fatalf("approve on the current run row: %v", err)
		}
		db := fcOpen(t, root)
		if err := db.VerifyApprovalReadonly(context.Background(), "t1", uuid); err != nil {
			t.Fatalf("the done gate refused the chain on the current run row: %v", err)
		}
	})

	t.Run("a_receipt_for_an_older_run_is_still_refused_and_approve_says_so", func(t *testing.T) {
		root, uuid, _, _, _ := squashLandedFixture(t, "squash")
		emptyStageRow(t, root)
		twoRunRows(t, root)
		_, stderr, err := runFactory(t, "approve", "t1", "--run", "run-old", "--issuer", "lead")
		if err != nil {
			t.Fatalf("approve on the older run row: %v", err)
		}
		if !strings.Contains(stderr, "will not satisfy") || !strings.Contains(stderr, fcRun) {
			t.Errorf("stderr = %q, want a note that the receipt for run-old will not satisfy the gate, naming the gate's run %s", stderr, fcRun)
		}
		db := fcOpen(t, root)
		if err := db.VerifyApprovalReadonly(context.Background(), "t1", uuid); !errors.Is(err, homestate.ErrApprovalRunMismatch) {
			t.Fatalf("err = %v, want the run-mismatch refusal: a previous run's receipt must not close the current run's work", err)
		}
	})

	t.Run("an_unlanded_stage_empty_card_still_cannot_reach_the_gate", func(t *testing.T) {
		root, uuid, _, _, _ := squashLandedFixture(t, "")
		emptyStageRow(t, root)
		twoRunRows(t, root)
		if _, err := rebindRun(t, "nothing landed"); err == nil {
			t.Fatal("rebind-evidence accepted a stage-empty row with no landing record")
		}
		if _, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead"); err == nil {
			t.Fatal("approve issued a receipt for a row with no evidence hash")
		}
		db := fcOpen(t, root)
		if err := db.VerifyApprovalReadonly(context.Background(), "t1", uuid); err == nil {
			t.Fatal("the done gate verified a card with no receipt")
		}
	})
}
