package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func waitTestDB(t *testing.T) *homestate.FactoryDB {
	t.Helper()
	db, err := homestate.OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func wakeLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "wake: ") {
			lines = append(lines, l)
		}
	}
	return lines
}

// TestFactoryWakeOnLanding drives the wait record end to end at the layer the
// complete verb calls: record a wait, land (or not) a card, count wake lines.
// No live session, no send: the verb only prints.
func TestFactoryWakeOnLanding(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Hour).Format(time.RFC3339Nano)

	cases := []struct {
		name      string
		wait      homestate.CardWait
		landing   homestate.Card
		wantWakes int
		wantNote  bool
		reruns    int // extra prints after the first; each must add 0 wake lines
	}{
		{
			name:      "predecessor landed local, lane waker: one wake, rerun silent",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "needs t2010 merged"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal},
			wantWakes: 1, reruns: 2,
		},
		{
			name:      "predecessor landed by pull request, leader waker",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: homestate.WaitWakerLeader, Reason: "r"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedPR},
			wantWakes: 1, reruns: 1,
		},
		{
			name:      "a different card landed: condition not met, nothing",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "r"},
			landing:   homestate.Card{CardID: "t2011", State: homestate.CardMergedLocal},
			wantWakes: 0,
		},
		{
			name:      "predecessor only has an expired lease, not landed: nothing",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "r"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardLeased, LeaseHolder: "lane-1", LeaseExpiresAt: expired},
			wantWakes: 0,
		},
		{
			name:      "operator-ruling wait: landing the ref card prints no lane wake",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondOperatorRuling, CondRef: "", Waker: homestate.WaitWakerOperator, Reason: "ruling on scope"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal},
			wantWakes: 0,
		},
		{
			name:      "operator waker on a predecessor wait: no direct lane wake, leader-facing note once",
			wait:      homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: homestate.WaitWakerOperator, Reason: "r"},
			landing:   homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal},
			wantWakes: 0, wantNote: true, reruns: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := waitTestDB(t)
			if err := db.RecordCardWait(ctx, tc.wait, now); err != nil {
				t.Fatalf("record wait: %v", err)
			}
			var first bytes.Buffer
			factoryPrintWakes(ctx, &first, db, tc.landing)
			if got := len(wakeLines(first.String())); got != tc.wantWakes {
				t.Fatalf("first print: want %d wake line(s), got %d:\n%s", tc.wantWakes, got, first.String())
			}
			if tc.wantWakes == 1 {
				line := wakeLines(first.String())[0]
				for _, want := range []string{"lane=" + tc.wait.Waker, "card=" + tc.wait.CardID, "reason="} {
					if !strings.Contains(line, want) {
						t.Errorf("wake line %q lacks %q", line, want)
					}
				}
				if !strings.Contains(first.String(), "moai factory status") || !strings.Contains(first.String(), "progress") {
					t.Errorf("wake output must point at lease-holder and progress record reads:\n%s", first.String())
				}
			}
			if tc.wantNote != strings.Contains(first.String(), "note: ") {
				t.Errorf("note presence: want %v, output:\n%s", tc.wantNote, first.String())
			}
			for i := 0; i < tc.reruns; i++ {
				var again bytes.Buffer
				factoryPrintWakes(ctx, &again, db, tc.landing)
				if strings.TrimSpace(again.String()) != "" {
					t.Fatalf("rerun %d must print nothing, got:\n%s", i+1, again.String())
				}
			}
		})
	}
}

func TestFactoryWakeTwoWaitsOnePredecessor(t *testing.T) {
	ctx := context.Background()
	db := waitTestDB(t)
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	for _, id := range []string{"t2020", "t2021"} {
		if err := db.RecordCardWait(ctx, homestate.CardWait{CardID: id, CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-" + id[len(id)-1:], Reason: "r"}, now); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	factoryPrintWakes(ctx, &out, db, homestate.Card{CardID: "t2010", State: homestate.CardMergedPR})
	if got := len(wakeLines(out.String())); got != 2 {
		t.Fatalf("want 2 wake lines, got %d:\n%s", got, out.String())
	}
}
