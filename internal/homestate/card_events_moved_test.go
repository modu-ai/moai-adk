package homestate

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"
)

// card_events_moved_test.go — SPEC-FACTORY-COMPLETION-RECOVERY-001 REQ-FCR-025
// (card t1963, ledger L11, AC-FCR-017 (11)): the event history of a card a
// consolidation moved into a run is read as that card's own history on the
// target run. Every fixture is an isolated sandbox factory; events are
// inserted directly (the shape card.transition / lease.expired rows carry) and
// the consolidation itself is the real RebindRun.

// lceBase is the anchor of the fixture timeline.
var lceBase = time.Date(2026, 10, 20, 8, 0, 0, 0, time.UTC)

// lceAt returns the fixture timestamp m minutes after lceBase.
func lceAt(m int) string { return lceBase.Add(time.Duration(m) * time.Minute).Format(time.RFC3339Nano) }

// lceEvent appends one event row and returns its seq.
func lceEvent(t *testing.T, db *FactoryDB, runID, kind string, payload any, at string) int64 {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return lceRawEvent(t, db, runID, kind, string(raw), at)
}

func lceRawEvent(t *testing.T, db *FactoryDB, runID, kind, payload, at string) int64 {
	t.Helper()
	res, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,?,?,?)`, runID, kind, payload, at)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	seq, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return seq
}

// lceTransition appends a card.transition row.
func lceTransition(t *testing.T, db *FactoryDB, runID, cardID, from, to string, m int) int64 {
	t.Helper()
	return lceEvent(t, db, runID, "card.transition", map[string]any{"card_id": cardID, "from": from, "to": to}, lceAt(m))
}

// lceOpts classifies a run dead when its pid is in dead, live otherwise; no
// worker is ever registered, so the worker probe is never reached.
func lceOpts(dead ...int) ReconcileOptions {
	return ReconcileOptions{
		Classify: func(pid int, _ string) OwnerClassification {
			if slices.Contains(dead, pid) {
				return OwnerDead
			}
			return OwnerLive
		},
		WorkerProbe: func(int) ProcessIdentityState { return ProcessIdentityDead },
	}
}

// lceRebind consolidates src into tgt under the classification opts.
func lceRebind(t *testing.T, db *FactoryDB, opts ReconcileOptions, src, tgt string) RebindResult {
	t.Helper()
	res, err := db.RebindRun(context.Background(), RebindRequest{
		Source: rbOwner(t, db, opts, src), Target: rbOwner(t, db, opts, tgt), RequireStampBasis: true,
	}, opts)
	if err != nil {
		t.Fatalf("rebind %s -> %s: %v", src, tgt, err)
	}
	return res
}

// lceSeqs lists the seqs of the events of one card.
func lceSeqs(events []CardEvent, cardID string) []int64 {
	var out []int64
	for _, e := range events {
		if e.CardID == cardID {
			out = append(out, e.Seq)
		}
	}
	return out
}

// lceLoad loads the card events of runID.
func lceLoad(t *testing.T, db *FactoryDB, runID string) []CardEvent {
	t.Helper()
	events, err := db.LoadCardEvents(context.Background(), runID)
	if err != nil {
		t.Fatalf("LoadCardEvents(%q): %v", runID, err)
	}
	return events
}

// lceSeedConsolidated builds the AC-FCR-017 (11) shape and consolidates it: a
// dead run-old holding c1 (run, three transitions, one lease expiry), c4
// (needs-decision, the target also holds a c4) and the terminal c5; a live
// run-new with its own c3 and c4 history. It returns the ordered seqs of the
// source events by card.
func lceSeedConsolidated(t *testing.T) (db *FactoryDB, c1, c4Src, c5, c3, c4Own []int64) {
	t.Helper()
	db = openSandboxFactory(t)
	rbRun(t, db, "run-old", rbPIDDead)
	rbRun(t, db, "run-new", rbPIDLive)
	rbCard(t, db, "run-old", "c1", CardRun)
	rbCard(t, db, "run-old", "c4", CardNeedsDecision)
	rbCard(t, db, "run-old", "c5", CardDone)
	rbCard(t, db, "run-new", "c3", CardAssigned)
	rbCard(t, db, "run-new", "c4", CardPicked)
	rbBind(t, db, "c1", "run-old")
	rbBind(t, db, "c4", "run-old")
	rbBind(t, db, "c5", "run-old")

	c1 = append(c1, lceTransition(t, db, "run-old", "c1", "", CardPicked, 0))
	c3 = append(c3, lceTransition(t, db, "run-new", "c3", "", CardPicked, 1))
	c1 = append(c1, lceTransition(t, db, "run-old", "c1", CardPicked, CardPlan, 5))
	c4Src = append(c4Src, lceTransition(t, db, "run-old", "c4", "", CardPicked, 6))
	c4Own = append(c4Own, lceTransition(t, db, "run-new", "c4", "", CardPicked, 7))
	c5 = append(c5, lceTransition(t, db, "run-old", "c5", "", CardPicked, 8))
	c1 = append(c1, lceTransition(t, db, "run-old", "c1", CardPlan, CardRun, 10))
	c1 = append(c1, lceEvent(t, db, "run-old", "lease.expired", map[string]any{"card_id": "c1", "from": CardRun, "to": CardAssigned}, lceAt(20)))
	c4Src = append(c4Src, lceTransition(t, db, "run-old", "c4", CardPicked, CardNeedsDecision, 21))
	c3 = append(c3, lceTransition(t, db, "run-new", "c3", CardPicked, CardAssigned, 22))

	res := lceRebind(t, db, rbOpts(), "run-old", "run-new")
	if !slices.Equal(res.Moved, []string{"c1"}) || !slices.Equal(res.Abandoned, []string{"c4"}) {
		t.Fatalf("fixture rebind = %+v, want moved [c1] abandoned [c4] (c5 terminal stays behind)", res)
	}
	return db, c1, c4Src, c5, c3, c4Own
}

// A moved card's card.transition and lease.expired history is visible on the
// target run, rewritten to the target, merged with the target's own events in
// Seq order. The stage-entry anchor reads known (the L11 witness).
func TestLoadCardEventsReadsMovedCardHistory(t *testing.T) {
	db, c1, _, _, c3, _ := lceSeedConsolidated(t)
	events := lceLoad(t, db, "run-new")

	if got := lceSeqs(events, "c1"); !slices.Equal(got, c1) {
		t.Fatalf("c1 event seqs on run-new = %v, want the source seqs %v (REQ-FCR-025)", got, c1)
	}
	if got := lceSeqs(events, "c3"); !slices.Equal(got, c3) {
		t.Fatalf("c3 (the target's own card) seqs = %v, want %v", got, c3)
	}
	for _, e := range events {
		if e.RunID != "run-new" {
			t.Fatalf("event seq %d carries RunID %q, want every returned event rewritten to run-new", e.Seq, e.RunID)
		}
	}
	for i := 1; i < len(events); i++ {
		if events[i-1].Seq >= events[i].Seq {
			t.Fatalf("events not in strictly ascending Seq order at %d: %d then %d", i, events[i-1].Seq, events[i].Seq)
		}
	}

	card, err := db.LoadCard(context.Background(), "run-new", "c1")
	if err != nil {
		t.Fatal(err)
	}
	now := lceBase.Add(3 * time.Hour)
	ages := ComputeCardAges(card, caaEventsOf(t, events, "run-new", "c1"), caaNoCommit, now)
	if !ages.Stage.Known {
		t.Fatal("stage-entry anchor of the moved card reads unknown on the target run (L11)")
	}
	if want := now.Sub(lceBase.Add(10 * time.Minute)); ages.Stage.Age != want {
		t.Fatalf("stage age = %s, want %s from the run transition in the source run", ages.Stage.Age, want)
	}
	if !ages.NoProgress.Known || ages.NoProgress.Age != now.Sub(lceBase.Add(10*time.Minute)) {
		t.Fatalf("no-progress age = %+v, want it anchored on the source-run stage entry", ages.NoProgress)
	}
}

// Loading is read-only: no event row is rewritten, copied, or deleted.
func TestLoadCardEventsLeavesSourceRowsByteIdentical(t *testing.T) {
	db, _, _, _, _, _ := lceSeedConsolidated(t)
	before := rbDump(t, db)
	for _, run := range []string{"run-new", "run-old", ""} {
		lceLoad(t, db, run)
	}
	if after := rbDump(t, db); after != before {
		t.Fatalf("LoadCardEvents changed stored rows:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM events WHERE run_id='run-new' AND kind IN ('card.transition','lease.expired') AND payload_json LIKE '%"c1"%'`); n != 0 {
		t.Fatalf("%d c1 event rows exist under run-new in storage; the history must be read through, never copied", n)
	}
}

// An abandoned card (the target already held it) and a terminal card that
// stayed behind contribute nothing from the source run.
func TestLoadCardEventsAbandonedAndLeftBehindCardsContributeNothing(t *testing.T) {
	db, _, c4Src, c5, _, c4Own := lceSeedConsolidated(t)
	events := lceLoad(t, db, "run-new")
	if got := lceSeqs(events, "c4"); !slices.Equal(got, c4Own) {
		t.Fatalf("c4 seqs on run-new = %v, want only the target's own %v (the abandoned source c4 %v must not merge)", got, c4Own, c4Src)
	}
	if got := lceSeqs(events, "c5"); len(got) != 0 {
		t.Fatalf("c5 (terminal, left behind in the source) seqs %v appeared on run-new; want none (source %v)", got, c5)
	}
}

// An empty runID reads every run's events once each: a moved card's history
// carries the run it now lives in (replacing the source copy, never added to
// it), and every other event — the abandoned card's, the left-behind terminal
// card's, the target's own — keeps its stored RunID.
func TestLoadCardEventsEmptyRunIDRewritesMovedCardsOnly(t *testing.T) {
	db, c1, c4Src, c5, c3, c4Own := lceSeedConsolidated(t)
	events := lceLoad(t, db, "")
	want := len(c1) + len(c4Src) + len(c5) + len(c3) + len(c4Own)
	if len(events) != want {
		t.Fatalf("empty runID returned %d events, want all %d stored card events (no duplicate under the source)", len(events), want)
	}
	for _, e := range events {
		wantRun := "run-new"
		if slices.Contains(c4Src, e.Seq) || slices.Contains(c5, e.Seq) {
			wantRun = "run-old" // abandoned and left-behind: unchanged
		}
		if e.RunID != wantRun {
			t.Fatalf("empty-runID event seq %d RunID = %q, want its stored run %q", e.Seq, e.RunID, wantRun)
		}
	}
	// The source run read alone also keeps its stored rows, moved card included.
	if got := lceSeqs(lceLoad(t, db, "run-old"), "c1"); !slices.Equal(got, c1) {
		t.Fatalf("run-old c1 seqs = %v, want %v", got, c1)
	}
}

// A→B→C: the history of a card moved twice is visible on C, from both hops.
// A card abandoned on the second hop does not carry the first hop's history.
func TestLoadCardEventsFollowsConsolidationChain(t *testing.T) {
	db, a1, a2, b1, b9, b2, c2Own := lceSeedChain(t)
	events := lceLoad(t, db, "run-c")
	if got, want := lceSeqs(events, "c1"), slices.Concat(a1, b1); !slices.Equal(got, want) {
		t.Fatalf("c1 on run-c = %v, want the history of both hops %v", got, want)
	}
	if got := lceSeqs(events, "c9"); !slices.Equal(got, b9) {
		t.Fatalf("c9 on run-c = %v, want %v", got, b9)
	}
	if got := lceSeqs(events, "c2"); !slices.Equal(got, c2Own) {
		t.Fatalf("c2 on run-c = %v, want only its own %v; abandoned on hop two, the history of hop one (%v) and the source (%v) stays out", got, c2Own, a2, b2)
	}
	for _, e := range events {
		if e.RunID != "run-c" {
			t.Fatalf("seq %d RunID %q, want run-c", e.Seq, e.RunID)
		}
	}
	// B is a target of the first hop: it still sees A's moved card.
	if got := lceSeqs(lceLoad(t, db, "run-b"), "c1"); !slices.Equal(got, slices.Concat(a1, b1)) {
		t.Fatalf("c1 on run-b = %v, want %v", got, slices.Concat(a1, b1))
	}
}

// The all-runs read (empty runID, the default `factory status`) resolves a
// moved card to the FINAL target of its chain, replacing — never duplicating —
// the source-run copy; an abandoned card stays under the run that held it.
func TestLoadCardEventsAllRunsFollowsConsolidationChain(t *testing.T) {
	db, a1, a2, b1, b9, b2, c2Own := lceSeedChain(t)
	events := lceLoad(t, db, "")
	runOf := map[int64]string{}
	for _, e := range events {
		if _, dup := runOf[e.Seq]; dup {
			t.Fatalf("seq %d returned twice by the all-runs read", e.Seq)
		}
		runOf[e.Seq] = e.RunID
	}
	want := map[string][]int64{
		"run-c": slices.Concat(a1, b1, b9, c2Own), // moved through A→B→C, plus c2's own
		"run-b": slices.Concat(a2, b2),            // c2 moved A→B, then abandoned on B→C
	}
	total := 0
	for run, seqs := range want {
		for _, s := range seqs {
			total++
			if runOf[s] != run {
				t.Fatalf("seq %d RunID = %q, want %q", s, runOf[s], run)
			}
		}
	}
	if len(events) != total {
		t.Fatalf("all-runs read returned %d events, want %d (replace, never duplicate)", len(events), total)
	}
	for i := 1; i < len(events); i++ {
		if events[i-1].Seq >= events[i].Seq {
			t.Fatalf("not in ascending Seq order at %d", i)
		}
	}
}

// lceSeedChain builds A→B→C: c1 moves A→B→C; c9 moves B→C; c2 moves A→B and
// is abandoned on B→C because run-c already holds it.
func lceSeedChain(t *testing.T) (db *FactoryDB, a1, a2, b1, b9, b2, c2Own []int64) {
	t.Helper()
	db = openSandboxFactory(t)
	const pa, pb, pc = 40, 41, 42
	rbRun(t, db, "run-a", pa)
	rbRun(t, db, "run-b", pb)
	rbRun(t, db, "run-c", pc)
	rbCard(t, db, "run-a", "c1", CardRun)
	rbCard(t, db, "run-a", "c2", CardRun)
	rbCard(t, db, "run-b", "c9", CardPlan)

	a1 = []int64{lceTransition(t, db, "run-a", "c1", "", CardPicked, 0), lceTransition(t, db, "run-a", "c1", CardPicked, CardPlan, 3)}
	a2 = []int64{lceTransition(t, db, "run-a", "c2", "", CardPicked, 1)}
	lceRebind(t, db, lceOpts(pa), "run-a", "run-b")
	rbCard(t, db, "run-c", "c2", CardPicked) // run-c takes c2 later: abandoned on the B -> C hop
	b1 = []int64{lceTransition(t, db, "run-b", "c1", CardPlan, CardRun, 30)}
	b9 = []int64{lceTransition(t, db, "run-b", "c9", "", CardPicked, 31)}
	c2Own = []int64{lceTransition(t, db, "run-c", "c2", "", CardPicked, 32)}
	b2 = []int64{lceTransition(t, db, "run-b", "c2", CardPicked, CardPlan, 33)}

	// run-b is now dead, run-c live: the second consolidation.
	res := lceRebind(t, db, lceOpts(pb), "run-b", "run-c")
	if !slices.Equal(res.Moved, []string{"c1", "c9"}) || !slices.Equal(res.Abandoned, []string{"c2"}) {
		t.Fatalf("second hop = %+v, want moved [c1 c9] abandoned [c2]", res)
	}
	return db, a1, a2, b1, b9, b2, c2Own
}

// A cycle in rebound_to terminates: a visited set bounds the walk, and an
// event is never returned twice.
func TestLoadCardEventsConsolidationCycleTerminates(t *testing.T) {
	db := openSandboxFactory(t)
	xs := lceTransition(t, db, "run-x", "c1", "", CardPicked, 0)
	ys := lceTransition(t, db, "run-y", "c2", "", CardPicked, 1)
	retired := func(run, to string, moved ...string) {
		lceEvent(t, db, run, "run.retired", map[string]any{"rebound_to": to, "moved_cards": moved, "abandoned_cards": []string{}}, lceAt(5))
	}
	retired("run-x", "run-y", "c1")
	retired("run-y", "run-x", "c2")
	retired("run-x", "run-x", "c1") // a self-reference

	load := func(run string) []CardEvent {
		done := make(chan []CardEvent, 1)
		go func() {
			events, err := db.LoadCardEvents(context.Background(), run)
			if err != nil {
				t.Errorf("LoadCardEvents(%q): %v", run, err)
			}
			done <- events
		}()
		select {
		case events := <-done:
			return events
		case <-time.After(10 * time.Second):
			t.Fatalf("LoadCardEvents(%q) did not terminate on a rebound_to cycle", run)
			return nil
		}
	}
	if all := load(""); len(all) != 2 {
		t.Fatalf("all-runs read over a cycle returned %d events, want the 2 stored ones exactly once", len(all))
	}
	events := load("run-x")
	seen := map[int64]bool{}
	for _, e := range events {
		if seen[e.Seq] {
			t.Fatalf("seq %d returned twice", e.Seq)
		}
		seen[e.Seq] = true
	}
	if !seen[ys] {
		t.Fatalf("run-y's moved card c2 (seq %d) missing from run-x; got %v", ys, events)
	}
	if !seen[xs] {
		t.Fatalf("run-x's own c1 (seq %d) missing; got %v", xs, events)
	}
}

// A run.retired row whose payload does not parse (or carries the wrong types)
// is skipped, never fatal; a well-formed sibling still works.
func TestLoadCardEventsMalformedRetiredPayloadSkipped(t *testing.T) {
	db := openSandboxFactory(t)
	good := lceTransition(t, db, "run-old", "c1", "", CardPicked, 0)
	own := lceTransition(t, db, "run-new", "c3", "", CardPicked, 1)
	lceRawEvent(t, db, "run-old", "run.retired", `not json at all`, lceAt(2))
	lceRawEvent(t, db, "run-old", "run.retired", `{"rebound_to": 7, "moved_cards": "c1"}`, lceAt(3))
	lceRawEvent(t, db, "run-old", "run.retired", `{"rebound_to": "run-new", "moved_cards": [1, 2]}`, lceAt(4))
	lceEvent(t, db, "run-old", "run.retired", map[string]any{"rebound_to": "run-new", "moved_cards": []string{"c1"}, "abandoned_cards": []string{}}, lceAt(5))
	// A card event with an unparseable payload is skipped too (the existing rule).
	lceRawEvent(t, db, "run-new", "card.transition", `{broken`, lceAt(6))

	events, err := db.LoadCardEvents(context.Background(), "run-new")
	if err != nil {
		t.Fatalf("a malformed run.retired payload failed the read: %v", err)
	}
	if got := lceSeqs(events, "c1"); !slices.Equal(got, []int64{good}) {
		t.Fatalf("c1 seqs = %v, want [%d] from the well-formed run.retired", got, good)
	}
	if got := lceSeqs(events, "c3"); !slices.Equal(got, []int64{own}) {
		t.Fatalf("c3 seqs = %v, want [%d]", got, own)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2: %v", len(events), fmt.Sprint(events))
	}
}
