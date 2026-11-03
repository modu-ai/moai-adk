package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestCardWaitRecordIsIdempotentAndValidated(t *testing.T) {
	ctx := context.Background()
	db, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	w := CardWait{CardID: "t2020", CondKind: WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "first"}
	if err := db.RecordCardWait(ctx, w, now); err != nil {
		t.Fatal(err)
	}
	w.Reason = "second"
	if err := db.RecordCardWait(ctx, w, now); err != nil {
		t.Fatal(err)
	}
	got, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(got) != 1 || got[0].Reason != "second" || got[0].WokenAt != "" {
		t.Fatalf("want one updated unwoken wait, got %+v err=%v", got, err)
	}
	for _, bad := range []CardWait{
		{CardID: "t2020", CondKind: "lease-expired", Waker: "lane-3"},
		{CardID: "t2020", CondKind: WaitCondPredecessorLanded, CondRef: "", Waker: "lane-3"},
		{CardID: "t2020", CondKind: WaitCondOperatorRuling, Waker: "two words"},
		{CardID: "../x", CondKind: WaitCondOperatorRuling, Waker: "operator"},
	} {
		if err := db.RecordCardWait(ctx, bad, now); !errors.Is(err, ErrInvalidCardInput) {
			t.Errorf("want ErrInvalidCardInput for %+v, got %v", bad, err)
		}
	}
	settled, err := db.SettleLandedWaits(ctx, "t2010", now)
	if err != nil || len(settled) != 1 || settled[0].WokenAt == "" {
		t.Fatalf("first settle: %+v err=%v", settled, err)
	}
	again, err := db.SettleLandedWaits(ctx, "t2010", now)
	if err != nil || len(again) != 0 {
		t.Fatalf("second settle must return nothing: %+v err=%v", again, err)
	}
}
