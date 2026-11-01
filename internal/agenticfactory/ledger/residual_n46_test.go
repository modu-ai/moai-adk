package ledger

// N46 of the final audit of a36ea8c368 (residual classification), the sibling of N24: a version 1 ledger in which c1 landed,
// the card was withdrawn from MERGING to REVIEWING at T, and a gate about ANOTHER commit c2 was recorded in the very millisecond
// of the withdrawal (after it). The migration counts evidence "at or before" a move, so c2's gate lands inside the withdrawal's
// mark and reads as work the holder had left behind; the card then finishes on the landing of c1 with newer work that the trunk
// does not hold. One millisecond later the same card is refused. The expectation: both cases are refused.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestResidualN46_AGateAboutAnotherCommitInTheMillisecondOfTheWithdrawalHoldsTheCardBack(t *testing.T) {
	for _, tc := range []struct{ name, at string }{
		{"in the millisecond of the withdrawal", "2026-10-20T10:00:09.000Z"},
		{"a millisecond after the withdrawal", "2026-10-20T10:00:09.001Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := v1Millisecond(t, false, nil)
			db, err := sql.Open("sqlite", "file:"+path)
			if err != nil {
				t.Fatal(err)
			}
			for _, q := range []string{
				`INSERT INTO transitions(card_id,from_stage,to_stage,epoch,actor,reason,at) VALUES('t1','MERGING','REVIEWING',1,'lane','t','2026-10-20T10:00:09.000Z')`,
				`UPDATE cards SET stage='REVIEWING' WHERE id='t1'`,
				`INSERT INTO evidence(card_id,kind,tree_sha,subject_sha,result,epoch,at) VALUES('t1','gate','tree-c2','c2','pass',1,'` + tc.at + `')`,
			} {
				if _, err := db.Exec(q); err != nil {
					t.Fatal(err)
				}
			}
			_ = db.Close()
			s, err := Open(bg, path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = s.Close() }()
			err = s.Move(bg, "t1", 1, core.Position{Stage: core.StageDone}, "lane", "done")
			got := cardOf(t, s, "t1").Position.Stage
			t.Logf("DONE from REVIEWING with c2's gate at %s: %v stage=%s", tc.at, err, got)
			if !errors.Is(err, core.ErrGuardMissing) || got == core.StageDone {
				t.Errorf("the card finished on the landing of c1 with a gate about c2 recorded at %s: err=%v stage=%s", tc.at, err, got)
			}
		})
	}
}
