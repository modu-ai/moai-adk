package homestate

import (
	"context"
	"testing"
	"time"
)

// factory_run_rebind_t1973_test.go — card t1973: reproductions of the two
// out-of-scope findings the t1966 sync-audit recorded against the consolidation
// loop (codex leg rcpt-ee107d9287, never reproduced by an auditor).
//
// P1 — a TERMINAL row of the target holding the same card_id silently consumes
// the source run's LIVE work: held>0 abandons the source row without ever
// looking at the target row's state or evidence.
//
// P2 — a third run whose leader classified dead (matching stamp) is skipped
// even when it still holds a valid lease on the card and/or a living worker:
// the skip clause checks the leader's stamp and nothing else.

// TestRed_T1973_P1_StaleDoneTargetEatsLiveWork — the target already holds the
// SAME card as a terminal row with its own recorded evidence and approval. The
// consolidation must be REFUSED (the source's live row preserved, a fresh
// approval demanded), not silently abandoned under the target's past done row.
func TestRed_T1973_P1_StaleDoneTargetEatsLiveWork(t *testing.T) {
	t.Run("terminal done target row with its own evidence and approval", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		// The source's LIVE work on the card.
		rbCard(t, db, "run-old", "c1", CardMergeReady)
		rbBind(t, db, "c1", "run-old")
		// The target's PAST completion of the same card id: a done row with its
		// own evidence and a recorded approval.
		rbCard(t, db, "run-new", "c1", CardDone)
		if _, err := db.DB.Exec(`UPDATE cards SET evidence_sha='sha-stale-done-c1' WHERE run_id='run-new' AND card_id='c1'`); err != nil {
			t.Fatal(err)
		}
		rbApproval(t, db, "uuid-c1-target", "run-new", "c1")
		before := rbDump(t, db)

		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-new")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
		// The live work survived: the source row is untouched and its run stays
		// active for a fresh decision.
		c, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
		if loadErr != nil || c.State != CardMergeReady {
			t.Fatalf("run-old c1 = (%+v, %v), want the untouched merge-ready row", c, loadErr)
		}
		if got := runStatus(t, db, "run-old"); got != "active" {
			t.Fatalf("run-old status = %s, want active after the refusal", got)
		}
	})
	t.Run("terminal target row even with matching evidence", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbCard(t, db, "run-old", "c1", CardRun)
		rbBind(t, db, "c1", "run-old")
		// rbCard seeds evidence "sha-c1" on both rows: terminal alone refuses.
		rbCard(t, db, "run-new", "c1", CardDone)
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-new")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
	})
	t.Run("non-terminal target row with divergent evidence", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbCard(t, db, "run-old", "c1", CardRun)
		rbBind(t, db, "c1", "run-old")
		// The target re-picked the same card and its row carries DIFFERENT
		// evidence: the two rows are not the same work.
		rbCard(t, db, "run-new", "c1", CardPicked)
		if _, err := db.DB.Exec(`UPDATE cards SET evidence_sha='sha-other-work-c1' WHERE run_id='run-new' AND card_id='c1'`); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-new")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
	})
}

// TestRed_T1973_P2_LivingWorkerAndLeaseInThirdRun — a third run whose LEADER
// classified dead under the very stamp the skip clause matches still holds the
// card under a VALID lease and/or with a LIVING worker (a lane can outlive its
// leader). The consolidation must be REFUSED, not skipped past.
func TestRed_T1973_P2_LivingWorkerAndLeaseInThirdRun(t *testing.T) {
	t.Run("valid lease on the dead-leader third run's row", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbRun(t, db, "run-third", rbPIDDead) // leader dead, stamp matches the skip clause
		rbCard(t, db, "run-old", "c1", CardRun)
		rbBind(t, db, "c1", "run-old")
		// The third run holds the same card in a lease-holding state with an
		// unexpired lease — a live lane, whatever its leader's stamp says.
		rbCard(t, db, "run-third", "c1", CardRun)
		expiry := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339Nano)
		if _, err := db.DB.Exec(`UPDATE cards SET lease_holder='lane-third', lease_expires_at=? WHERE run_id='run-third' AND card_id='c1'`, expiry); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-third")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
		if got := rbBound(t, db, "c1"); got != "run-old" {
			t.Fatalf("card_dispatch[c1] = %s, want it unchanged at run-old", got)
		}
	})
	t.Run("living worker registered on the dead-leader third run", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbRun(t, db, "run-third", rbPIDDead)
		rbCard(t, db, "run-old", "c1", CardRun)
		rbBind(t, db, "c1", "run-old")
		// No lease on the third run's row, but a registered worker whose pid
		// the probe reports LIVE.
		rbCard(t, db, "run-third", "c1", CardAssigned)
		rbWorker(t, db, "run-third", "lane-third", rbPIDWorkerLive)
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-third")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
	})
}

// TestRed_T1973b_P1_SourceCardHoldingValidLeaseIsRekeyed — round 2, P1: the
// SOURCE run's own card holds a valid lease. The source leader's death does not
// kill a card lease, and refuseLiveWorker only probes workers registered UNDER
// the source run — a holder whose registration names another run (or none) is
// invisible to it. The leased row must not be re-keyed or abandoned.
func TestRed_T1973b_P1_SourceCardHoldingValidLeaseIsRekeyed(t *testing.T) {
	t.Run("valid lease on the source card, no source worker row", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbCard(t, db, "run-old", "c1", CardRun)
		rbBind(t, db, "c1", "run-old")
		// The lane leased the card but its worker row names ANOTHER run, so
		// refuseLiveWorker's run-scoped worker listing cannot see it.
		expiry := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339Nano)
		if _, err := db.DB.Exec(`UPDATE cards SET lease_holder='lane-x', lease_expires_at=? WHERE run_id='run-old' AND card_id='c1'`, expiry); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-old")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
		c, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
		if loadErr != nil || c.State != CardRun || c.LeaseHolder != "lane-x" {
			t.Fatalf("run-old c1 = (%+v, %v), want the leased run row untouched", c, loadErr)
		}
	})
	t.Run("holder registered as a worker of a different run", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbCard(t, db, "run-old", "c1", CardRun)
		rbBind(t, db, "c1", "run-old")
		expiry := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339Nano)
		if _, err := db.DB.Exec(`UPDATE cards SET lease_holder='lane-x', lease_expires_at=? WHERE run_id='run-old' AND card_id='c1'`, expiry); err != nil {
			t.Fatal(err)
		}
		// Even a LIVE worker row does not block when it is registered under the
		// target: the card's own lease must decide.
		rbWorker(t, db, "run-new", "lane-x", rbPIDWorkerLive)
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-old")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
	})
}

// TestRed_T1973b_P2_BothEmptyEvidenceTreatedAsMatch — round 2, P2: two EMPTY
// evidence values are not a match. The normal early state (the source row
// actively working, the target row freshly created/picked with no evidence yet)
// would abandon the source row, and its stage/worktree/spec_id never reach the
// target. The both-empty shape must refuse.
func TestRed_T1973b_P2_BothEmptyEvidenceTreatedAsMatch(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	rbRun(t, db, "run-old", rbPIDDead)
	rbRun(t, db, "run-new", rbPIDLive)
	rbCard(t, db, "run-old", "c1", CardRun)
	rbBind(t, db, "c1", "run-old")
	// The freshly created target row: non-terminal, no evidence recorded yet.
	rbCard(t, db, "run-new", "c1", CardPicked)
	if _, err := db.DB.Exec(`UPDATE cards SET evidence_sha='' WHERE card_id='c1'`); err != nil {
		t.Fatal(err)
	}
	// The source row carries progress the abandon pointer would strand.
	if _, err := db.DB.Exec(`UPDATE cards SET stage='run', worktree_path='/wt/c1', spec_id='SPEC-X' WHERE run_id='run-old' AND card_id='c1'`); err != nil {
		t.Fatal(err)
	}
	before := rbDump(t, db)
	_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
	rbRefusedWith(t, err, "c1", "run-new")
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
	}
	c, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
	if loadErr != nil || c.State != CardRun || c.Stage != "run" || c.WorktreePath != "/wt/c1" {
		t.Fatalf("run-old c1 = (%+v, %v), want the working row and its progress untouched", c, loadErr)
	}
}

// TestRed_T1973c_AssignedSourceRowCarryingStageIsAbandoned — round 4: the
// state-keyed narrowing is insufficient — applyLeaseExpiry returns a row to
// `assigned` KEEPING its stage/worktree, and T21 keeps the existing stage on a
// question receipt, so an ASSIGNED source row can carry real progress. The
// both-empty shape must refuse when the source row carries progress info
// (stage / worktree / spec_id) the target row does not already hold.
func TestRed_T1973c_AssignedSourceRowCarryingStageIsAbandoned(t *testing.T) {
	t.Run("assigned source with stage and worktree over a bare target row", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		// The source row: back to assigned by lease expiry, its stage and
		// worktree still on the row.
		rbCard(t, db, "run-old", "c1", CardAssigned)
		rbBind(t, db, "c1", "run-old")
		if _, err := db.DB.Exec(`UPDATE cards SET stage='run', worktree_path='/wt/c1', spec_id='SPEC-C1', evidence_sha='' WHERE run_id='run-old' AND card_id='c1'`); err != nil {
			t.Fatal(err)
		}
		// The target row: freshly picked, bare — no stage, no worktree, no
		// evidence.
		rbCard(t, db, "run-new", "c1", CardPicked)
		if _, err := db.DB.Exec(`UPDATE cards SET stage='', evidence_sha='' WHERE run_id='run-new' AND card_id='c1'`); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-new")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
		c, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
		if loadErr != nil || c.State != CardAssigned || c.Stage != "run" || c.WorktreePath != "/wt/c1" {
			t.Fatalf("run-old c1 = (%+v, %v), want the assigned row with its stage/worktree untouched", c, loadErr)
		}
	})
	t.Run("needs-decision source with stage over a bare target row", func(t *testing.T) {
		db := openSandboxFactory(t)
		opts := rbOpts()
		rbRun(t, db, "run-old", rbPIDDead)
		rbRun(t, db, "run-new", rbPIDLive)
		rbCard(t, db, "run-old", "c1", CardNeedsDecision)
		rbBind(t, db, "c1", "run-old")
		if _, err := db.DB.Exec(`UPDATE cards SET stage='run', worktree_path='/wt/c1', evidence_sha='' WHERE run_id='run-old' AND card_id='c1'`); err != nil {
			t.Fatal(err)
		}
		rbCard(t, db, "run-new", "c1", CardPicked)
		if _, err := db.DB.Exec(`UPDATE cards SET stage='', evidence_sha='' WHERE run_id='run-new' AND card_id='c1'`); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
		rbRefusedWith(t, err, "c1", "run-new")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
		}
	})
}

// TestT1973_R3_BothEmptyNonWorkingSourceKeepsAbandon — the restored t1963
// dedup, round-4 form: a both-empty pair where the source row is a BARE stub
// (empty stage, empty worktree, empty spec_id — nothing the target row lacks)
// keeps the abandon-with-pointer behavior the operator-move and leader-start-
// consolidation paths live on. (Round-3 form keyed on the source row's state;
// round 4 keys on the actual remaining progress info instead.)
func TestT1973_R3_BothEmptyNonWorkingSourceKeepsAbandon(t *testing.T) {
	for _, srcState := range []string{CardPicked, CardAssigned, CardNeedsDecision} {
		t.Run("bare source stub in "+srcState, func(t *testing.T) {
			db := openSandboxFactory(t)
			opts := rbOpts()
			rbRun(t, db, "run-old", rbPIDDead)
			rbRun(t, db, "run-new", rbPIDLive)
			rbCard(t, db, "run-old", "c1", srcState)
			rbBind(t, db, "c1", "run-old")
			rbCard(t, db, "run-new", "c1", CardAssigned)
			// Both rows bare: no stage, no evidence, no worktree, no spec.
			if _, err := db.DB.Exec(`UPDATE cards SET stage='', evidence_sha='' WHERE card_id='c1'`); err != nil {
				t.Fatal(err)
			}
			res, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
			if err != nil {
				t.Fatalf("the evidence-less bare-stub dedup refused: %v", err)
			}
			if len(res.Abandoned) != 1 || res.Abandoned[0] != "c1" {
				t.Fatalf("res = %+v, want c1 abandoned with a pointer", res)
			}
			old, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
			if loadErr != nil || old.State != CardAbandoned {
				t.Fatalf("run-old c1 = (%+v, %v), want abandoned", old, loadErr)
			}
			if got := runStatus(t, db, "run-old"); got != "retired" {
				t.Fatalf("run-old status = %s, want retired", got)
			}
		})
	}
}

// TestRed_T1973d_NeedsDecisionQuestionStranded — round 5: a needs-decision
// row entering via T21 carries decision_question — an UNRESOLVED operator
// decision — even when stage/worktree/spec_id/evidence are all empty. The
// round-4 gate does not compare decision fields, so the pending question is
// stranded and the target can be leased without the operator's answer.
func TestRed_T1973d_NeedsDecisionQuestionStranded(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	rbRun(t, db, "run-old", rbPIDDead)
	rbRun(t, db, "run-new", rbPIDLive)
	rbCard(t, db, "run-old", "c1", CardNeedsDecision)
	rbBind(t, db, "c1", "run-old")
	// The pending operator question — and nothing else on the row.
	if _, err := db.DB.Exec(`UPDATE cards SET stage='', evidence_sha='', decision_gate='question', decision_question='merge #1906 now or hold?', decision_resume='kickoff:...' WHERE run_id='run-old' AND card_id='c1'`); err != nil {
		t.Fatal(err)
	}
	rbCard(t, db, "run-new", "c1", CardPicked)
	if _, err := db.DB.Exec(`UPDATE cards SET stage='', evidence_sha='' WHERE run_id='run-new' AND card_id='c1'`); err != nil {
		t.Fatal(err)
	}
	before := rbDump(t, db)
	_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
	rbRefusedWith(t, err, "c1", "run-new")
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
	}
	c, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
	if loadErr != nil || c.DecisionQuestion != "merge #1906 now or hold?" {
		t.Fatalf("run-old c1 = (%+v, %v), want the pending question preserved", c, loadErr)
	}
}

// TestT1973_R5_StateTransitionMatrix — the full both-empty transition matrix
// (round 5, leader ask): source shapes × target shapes, eight cells.
//
//	| source shape                        | bare target | identical target |
//	|-------------------------------------|-------------|------------------|
//	| picked bare stub                    | abandon     | abandon          |
//	| assigned bare stub                  | abandon     | abandon          |
//	| needs-decision w/ pending question  | REFUSE      | abandon          |
//	| expired→assigned (stage+worktree)   | REFUSE      | abandon          |
//
// bare-stub sources with a bare target → abandon (the t1963 path); any
// stranded progress/decision field over a bare target → refuse; identical-
// hold pairs → abandon (the dedup). (An identical-hold target of a bare stub
// is itself bare, so the first two rows collapse in practice but stay pinned
// per the table.)
func TestT1973_R5_StateTransitionMatrix(t *testing.T) {
	const (
		qGate    = "question"
		qText    = "merge #1906 now or hold?"
		qResume  = "kickoff:..."
		expStage = "run"
		expWT    = "/wt/c1"
	)
	type shape struct {
		name  string
		state string
		set   string // SET fragment applied to the SOURCE row (beyond state)
	}
	sources := []shape{
		{"picked bare stub", CardPicked, ""},
		{"assigned bare stub", CardAssigned, ""},
		{"needs-decision w/ pending question", CardNeedsDecision, `decision_gate='` + qGate + `', decision_question='` + qText + `', decision_resume='` + qResume + `'`},
		{"expired-to-assigned keeping stage/worktree", CardAssigned, `stage='` + expStage + `', worktree_path='` + expWT + `'`},
	}
	for _, src := range sources {
		for _, identical := range []bool{false, true} {
			target := "bare picked target"
			if identical {
				target = "target identically holding the source's info"
			}
			t.Run(src.name+" over "+target, func(t *testing.T) {
				db := openSandboxFactory(t)
				opts := rbOpts()
				rbRun(t, db, "run-old", rbPIDDead)
				rbRun(t, db, "run-new", rbPIDLive)
				rbCard(t, db, "run-old", "c1", src.state)
				rbBind(t, db, "c1", "run-old")
				rbCard(t, db, "run-new", "c1", CardPicked)
				// Evidence is empty on both sides in every cell (the
				// both-empty shape this matrix pins); stage blanked on both.
				if _, err := db.DB.Exec(`UPDATE cards SET stage='', evidence_sha='' WHERE card_id='c1'`); err != nil {
					t.Fatal(err)
				}
				// Apply the source shape's info...
				if src.set != "" {
					if _, err := db.DB.Exec(`UPDATE cards SET ` + src.set + ` WHERE run_id='run-old' AND card_id='c1'`); err != nil {
						t.Fatal(err)
					}
				}
				// ...and mirror it onto the target when the cell says identical.
				if identical && src.set != "" {
					if _, err := db.DB.Exec(`UPDATE cards SET ` + src.set + ` WHERE run_id='run-new' AND card_id='c1'`); err != nil {
						t.Fatal(err)
					}
				}
				wantRefuse := src.set != "" && !identical
				res, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
				if wantRefuse {
					rbRefusedWith(t, err, "c1", "run-new")
					old, loadErr := db.LoadCard(context.Background(), "run-old", "c1")
					if loadErr != nil || old.State != src.state {
						t.Fatalf("run-old c1 = (%+v, %v), want the source row untouched in %s", old, loadErr, src.state)
					}
					return
				}
				if err != nil {
					t.Fatalf("the dedup refused: %v", err)
				}
				if len(res.Abandoned) != 1 || res.Abandoned[0] != "c1" {
					t.Fatalf("res = %+v, want c1 abandoned with a pointer", res)
				}
				if got := runStatus(t, db, "run-old"); got != "retired" {
					t.Fatalf("run-old status = %s, want retired", got)
				}
			})
		}
	}
}
