package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// seedWaitCard inserts one card row (only the NOT NULL columns) into the
// factory record that the wait verb opens for root.
func seedWaitCard(t *testing.T, root, runID, cardID, state string) {
	t.Helper()
	db := openWaitRoot(t, root)
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
		runID, cardID, state, "2026-11-02T00:00:00Z"); err != nil {
		t.Fatalf("seed card %s: %v", cardID, err)
	}
}

func onlyWait(t *testing.T, root, card string) homestate.CardWait {
	t.Helper()
	waits, err := openWaitRoot(t, root).LoadCardWaits(context.Background(), card)
	if err != nil || len(waits) != 1 {
		t.Fatalf("want exactly one wait for %s, got %+v err=%v", card, waits, err)
	}
	return waits[0]
}

func landingEdgeOutput(t *testing.T, root, pred, state string) string {
	t.Helper()
	var out bytes.Buffer
	factoryPrintWakes(context.Background(), &out, openWaitRoot(t, root), homestate.Card{CardID: pred, State: state})
	return out.String()
}

// (a) A wait recorded after the predecessor already landed is woken once, at
// record time, and the later landing edge cannot wake it a second time.
func TestFactoryWaitOnAlreadyLandedPredecessorWakesOnceAtRecord(t *testing.T) {
	for _, landed := range []string{homestate.CardMergedLocal, homestate.CardMergedPR} {
		t.Run(landed, func(t *testing.T) {
			root := t.TempDir()
			seedWaitCard(t, root, "run-1", "t2010", landed)
			out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010")
			if err != nil {
				t.Fatalf("wait: %v\n%s", err, out)
			}
			lines := wakeLines(out)
			if len(lines) != 1 {
				t.Fatalf("want exactly one wake line at record time, got %d:\n%s", len(lines), out)
			}
			for _, want := range []string{"lane=lane-3", "card=t2020", "reason=needs t2010", "(predecessor t2010 landed)", "already landed"} {
				if !strings.Contains(lines[0], want) {
					t.Errorf("wake line %q lacks %q", lines[0], want)
				}
			}
			if !strings.Contains(out, "wait recorded: card=t2020 condition=predecessor-landed t2010 waker=lane-3") {
				t.Errorf("record line missing:\n%s", out)
			}
			if w := onlyWait(t, root, "t2020"); w.WokenAt == "" {
				t.Errorf("the wait must carry woken_at after the record-time wake: %+v", w)
			}
			if later := landingEdgeOutput(t, root, "t2010", landed); strings.TrimSpace(later) != "" {
				t.Errorf("a later landing edge must print no second wake, got:\n%s", later)
			}
		})
	}
}

// (b) A predecessor that has not landed (still leased, or unknown to the
// record) keeps today's behavior: no wake at record time, the wait stays open,
// and the landing edge wakes it exactly once.
func TestFactoryWaitOnNotLandedPredecessorStaysOpenUntilLanding(t *testing.T) {
	cases := []struct {
		name  string
		state string // empty = predecessor absent from the record
	}{
		{"leased predecessor", homestate.CardLeased},
		{"merge-ready predecessor", homestate.CardMergeReady},
		{"unknown predecessor", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.state != "" {
				seedWaitCard(t, root, "run-1", "t2010", tc.state)
			}
			out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010")
			if err != nil {
				t.Fatalf("wait: %v\n%s", err, out)
			}
			if want := "wait recorded: card=t2020 condition=predecessor-landed t2010 waker=lane-3\n"; out != want {
				t.Errorf("output = %q, want %q", out, want)
			}
			if w := onlyWait(t, root, "t2020"); w.WokenAt != "" {
				t.Errorf("the wait must stay open: %+v", w)
			}
			edge := landingEdgeOutput(t, root, "t2010", homestate.CardMergedLocal)
			lines := wakeLines(edge)
			if len(lines) != 1 || strings.Contains(lines[0], "already landed") {
				t.Fatalf("landing edge must wake exactly once without the already-landed suffix, got %q", lines)
			}
			if again := landingEdgeOutput(t, root, "t2010", homestate.CardMergedLocal); strings.TrimSpace(again) != "" {
				t.Errorf("a second landing edge must print nothing, got:\n%s", again)
			}
		})
	}
}

// (c) An operator-ruling wait is never woken at record time, whatever landed.
func TestFactoryWaitOperatorRulingNeverWokenAtRecord(t *testing.T) {
	root := t.TempDir()
	seedWaitCard(t, root, "run-1", "t2010", homestate.CardMergedLocal)
	out, err := runFactoryWaitCmd(t, root, "t2020", "--operator", "--waker", "operator", "--reason", "ruling on scope")
	if err != nil {
		t.Fatalf("wait: %v\n%s", err, out)
	}
	if strings.Contains(out, "wake: ") || strings.Contains(out, "note: ") {
		t.Errorf("an operator-ruling wait must print no wake or note at record time:\n%s", out)
	}
	if w := onlyWait(t, root, "t2020"); w.WokenAt != "" {
		t.Errorf("an operator-ruling wait must stay open: %+v", w)
	}
}

// (d) Re-recording the identical wait after it was woken at record time writes
// nothing and wakes nobody a second time.
func TestFactoryWaitIdenticalReRecordAfterRecordTimeWakeIsSilent(t *testing.T) {
	root := t.TempDir()
	seedWaitCard(t, root, "run-1", "t2010", homestate.CardMergedPR)
	args := []string{"t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010"}
	if out, err := runFactoryWaitCmd(t, root, args...); err != nil || len(wakeLines(out)) != 1 {
		t.Fatalf("first record: err=%v out:\n%s", err, out)
	}
	before := onlyWait(t, root, "t2020")
	out, err := runFactoryWaitCmd(t, root, args...)
	if err != nil {
		t.Fatalf("re-record: %v\n%s", err, out)
	}
	if got := wakeLines(out); len(got) != 0 {
		t.Errorf("an identical re-record must print no second wake, got %q", got)
	}
	if after := onlyWait(t, root, "t2020"); after != before {
		t.Errorf("an identical re-record must write nothing: before=%+v after=%+v", before, after)
	}
}
