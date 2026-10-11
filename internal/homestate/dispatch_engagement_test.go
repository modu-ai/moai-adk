package homestate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// RecordDispatchEngagement (card t1538, SPEC-FACTORY-COMPLETION-RECOVERY-001,
// review rounds 23-24): the dispatch binding and the row's owner move land in
// ONE factory transaction, and a valid lease on the row refuses the whole
// write.

// deLegacyCanon is a stand-in for the caller's canonical owner vocabulary:
// the legacy `worker-<n>` spelling maps to `lane-<n>`.
func deLegacyCanon(label string) string { return strings.Replace(label, "worker-", "lane-", 1) }

func deBinding(t *testing.T, db *FactoryDB, cardID string) string {
	t.Helper()
	var run string
	err := db.DB.QueryRow(`SELECT run_id FROM card_dispatch WHERE card_id=?`, cardID).Scan(&run)
	if err != nil {
		return ""
	}
	return run
}

func deRow(runID, owner string) Card {
	return Card{RunID: runID, CardID: "t1", OwnerLabel: owner, State: CardAssigned}
}

func TestRecordDispatchEngagementMovesOwnerAndBindingTogether(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, deRow("run-1", "lane-1"))
	if err := db.RecordDispatchEngagement(context.Background(), "t1", "run-1", "lane-2", deLegacyCanon, frNow); err != nil {
		t.Fatal(err)
	}
	row, err := db.LoadCard(context.Background(), "run-1", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if row.OwnerLabel != "lane-2" || row.Version != 2 || row.State != CardAssigned {
		t.Fatalf("row = owner %q v%d state %s, want lane-2 v2 assigned (the version bump is what stales the prior owner's approval)", row.OwnerLabel, row.Version, row.State)
	}
	if got := deBinding(t, db, "t1"); got != "run-1" {
		t.Fatalf("binding = %q, want run-1", got)
	}
	if ev := frEvents(t, db, "card.dispatch-owner"); len(ev) != 1 || !strings.Contains(ev[0], `"from":"lane-1"`) || !strings.Contains(ev[0], `"to":"lane-2"`) {
		t.Fatalf("card.dispatch-owner events = %v", ev)
	}
}

func TestRecordDispatchEngagementEquivalentOwnerIsNotRestamped(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, deRow("run-1", "worker-1"))
	if err := db.RecordDispatchEngagement(context.Background(), "t1", "run-1", "lane-1", deLegacyCanon, frNow); err != nil {
		t.Fatal(err)
	}
	row, _ := db.LoadCard(context.Background(), "run-1", "t1")
	if row.OwnerLabel != "worker-1" || row.Version != 1 {
		t.Fatalf("row = owner %q v%d, want the legacy spelling untouched at v1", row.OwnerLabel, row.Version)
	}
	if got := deBinding(t, db, "t1"); got != "run-1" {
		t.Fatalf("binding = %q, want run-1", got)
	}
}

func TestRecordDispatchEngagementWithoutOwnerIsBindingOnly(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, deRow("run-1", "lane-1"))
	if err := db.RecordDispatchEngagement(context.Background(), "t1", "run-1", "", deLegacyCanon, frNow); err != nil {
		t.Fatal(err)
	}
	row, _ := db.LoadCard(context.Background(), "run-1", "t1")
	if row.OwnerLabel != "lane-1" || row.Version != 1 {
		t.Fatalf("row = owner %q v%d, want untouched", row.OwnerLabel, row.Version)
	}
	if got := deBinding(t, db, "t1"); got != "run-1" {
		t.Fatalf("binding = %q, want run-1", got)
	}
}

func TestRecordDispatchEngagementWithoutARowRecordsRunAndBinding(t *testing.T) {
	db := frOpen(t)
	if err := db.RecordDispatchEngagement(context.Background(), "t1", "run-9", "lane-1", nil, frNow); err != nil {
		t.Fatal(err)
	}
	if got := deBinding(t, db, "t1"); got != "run-9" {
		t.Fatalf("binding = %q, want run-9", got)
	}
	var runs int
	if err := db.DB.QueryRow(`SELECT count(*) FROM runs WHERE run_id='run-9'`).Scan(&runs); err != nil || runs != 1 {
		t.Fatalf("run row count = %d (err %v), want 1", runs, err)
	}
}

func TestRecordDispatchEngagementRefusesWholeWriteUnderValidLease(t *testing.T) {
	ctx := context.Background()
	for name, expires := range map[string]string{
		"future":     frLeaseUntil(time.Hour),
		"unparsable": "not-a-time",
	} {
		t.Run(name, func(t *testing.T) {
			db := frOpen(t)
			leased := deRow("run-old", "lane-1")
			leased.State, leased.Stage, leased.LeaseHolder, leased.LeaseExpiresAt = CardLeased, CardRun, "lane-1", expires
			frPlace(t, db, leased)
			if err := db.RecordDispatchBinding(ctx, "t1", "run-cur", frNow); err != nil {
				t.Fatal(err)
			}
			before := frRowDump(t, db, "run-old", "t1")
			err := db.RecordDispatchEngagement(ctx, "t1", "run-old", "lane-2", deLegacyCanon, frNow)
			if !errors.Is(err, ErrLeaseHolder) {
				t.Fatalf("err = %v, want ErrLeaseHolder", err)
			}
			if after := frRowDump(t, db, "run-old", "t1"); after != before {
				t.Fatalf("row/event log changed by the refused write:\nbefore %s\nafter  %s", before, after)
			}
			if got := deBinding(t, db, "t1"); got != "run-cur" {
				t.Fatalf("binding = %q after the refusal, want run-cur — a refusal commits nothing", got)
			}
		})
	}
}

func TestRecordDispatchEngagementExpiredLeaseAllowsTheMove(t *testing.T) {
	db := frOpen(t)
	leased := deRow("run-1", "lane-1")
	leased.State, leased.Stage, leased.LeaseHolder, leased.LeaseExpiresAt = CardLeased, CardRun, "lane-1", frLeaseUntil(-time.Hour)
	frPlace(t, db, leased)
	if err := db.RecordDispatchEngagement(context.Background(), "t1", "run-1", "lane-2", deLegacyCanon, frNow); err != nil {
		t.Fatal(err)
	}
	row, _ := db.LoadCard(context.Background(), "run-1", "t1")
	if row.OwnerLabel != "lane-2" || row.Version != 2 || row.State != CardLeased || row.LeaseHolder != "lane-1" {
		t.Fatalf("row = owner %q v%d state %s holder %q — only the owner moves; state and lease columns stay", row.OwnerLabel, row.Version, row.State, row.LeaseHolder)
	}
}

func TestRecordDispatchEngagementRejectsEmptyIdentifiers(t *testing.T) {
	db := frOpen(t)
	for _, c := range [][2]string{{"", "run-1"}, {"t1", ""}, {"  ", "run-1"}} {
		if err := db.RecordDispatchEngagement(context.Background(), c[0], c[1], "lane-1", nil, frNow); !errors.Is(err, ErrInvalidCardInput) {
			t.Fatalf("card %q run %q: err = %v, want ErrInvalidCardInput", c[0], c[1], err)
		}
	}
}
