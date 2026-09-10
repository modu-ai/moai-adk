package guardstate

import (
	"context"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// AC-GSM-013 — every carried count has a consumer.
//
// Clause (a) is a shape assertion and is cheap. Clause (b) is the one with a
// history: the parent card's audit found no inert requirement but TWO inert
// fields — counts required, rendered, and read by nothing. The disjunct that
// made the clause vacuous ("consumed by ... the published result's own
// contract") was removed at plan time precisely because every carried count
// satisfies it by definition.
//
// So clause (b) is measured here the only way it can be: for each count, CHANGE
// IT and show a decision changes. A count whose value can be replaced without
// any outcome moving is inert, whatever the prose says about it.
// ---------------------------------------------------------------------------

// TestCarriedCountSetIsComplete is clause (a). The per-value counts are asserted
// over the whole closed vocabulary rather than over the values this fixture
// happened to produce: a map that only acquires a key when the classifier emits
// that value carries "no count" and "zero" identically, which is the same
// two-states-one-representation defect this SPEC exists to close.
func TestCarriedCountSetIsComplete(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	m := &Manifest{Entries: []Entry{wfEntry(".github/workflows/a.yml", "7d", MeasureFiredAtAll, "")}}
	enum := healthyEnum(".github/workflows/a.yml", ".github/workflows/b.yml")
	q := &fakeQuerier{histories: map[string]RunHistory{
		".github/workflows/a.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-time.Hour)}}},
	}}

	ev := Evaluate(context.Background(), m, enum, q, now)

	if !ev.MeasuredAt.Equal(now) {
		t.Fatalf("the measurement timestamp is %v, want %v", ev.MeasuredAt, now)
	}
	if ev.Enumerated != 2 {
		t.Fatalf("enumerated count %d, want 2", ev.Enumerated)
	}
	if ev.Declared != 1 {
		t.Fatalf("declared count %d, want 1", ev.Declared)
	}
	if ev.Queried != 1 {
		t.Fatalf("queried count %d, want 1", ev.Queried)
	}

	// A per-value count for EACH of the seven classifications, including the
	// six this fixture never produced.
	if len(ev.Counts) != len(Classifications()) {
		t.Fatalf("the count set holds %d values, want one per classification (%d)",
			len(ev.Counts), len(Classifications()))
	}
	for _, c := range Classifications() {
		if _, carried := ev.Counts[c]; !carried {
			t.Fatalf("no count is carried for %q", c)
		}
	}
}

// TestEveryCarriedCountFeedsAGuard is clause (b), measured per count.
//
// Each case takes a conforming evaluation, changes exactly one carried count,
// and asserts the all-clear decision moves. A count that can be changed with no
// outcome moving is inert — which is the finding this clause was written from.
func TestEveryCarriedCountFeedsAGuard(t *testing.T) {
	t.Parallel()

	// A conforming evaluation, built rather than evaluated so exactly one field
	// moves per case.
	conforming := func() Evaluation {
		counts := map[Classification]int{}
		for _, c := range Classifications() {
			counts[c] = 0
		}
		counts[ClassOK] = 2
		return Evaluation{
			MeasuredAt:  time.Now(),
			Enumerated:  2,
			Declared:    2,
			Queried:     2,
			IntegrityOK: true,
			Counts:      counts,
			Decisions: []Decision{
				{Subject: "a", Row: 3, Class: ClassOK, Surface: SurfaceOK},
				{Subject: "b", Row: 3, Class: ClassOK, Surface: SurfaceOK},
			},
		}
	}

	if refusals := conforming().Refusals(); len(refusals) != 0 {
		t.Fatalf("the conforming baseline is already refused, so no case below measures anything: %v", refusals)
	}

	cases := []struct {
		count  string
		change func(*Evaluation)
		reason string
	}{
		{
			count:  "the queried count",
			change: func(e *Evaluation) { e.Queried = 0 },
			reason: "queried",
		},
		{
			count:  "the enumerated-files count",
			change: func(e *Evaluation) { e.Enumerated = 0 },
			reason: "enumerat",
		},
		{
			count:  "the declared count",
			change: func(e *Evaluation) { e.Declared = 5 },
			reason: "declared",
		},
		{
			count: "a per-value count",
			change: func(e *Evaluation) {
				e.Counts[ClassOK] = 1
				e.Counts[ClassStale] = 1
			},
			reason: "not clean",
		},
	}

	for _, tc := range cases {
		t.Run(tc.count, func(t *testing.T) {
			t.Parallel()
			ev := conforming()
			tc.change(&ev)
			refusals := ev.Refusals()
			if len(refusals) == 0 {
				t.Fatalf("%s changed and no decision moved: the count is inert", tc.count)
			}
			joined := strings.ToLower(strings.Join(refusals, " | "))
			if !strings.Contains(joined, tc.reason) {
				t.Fatalf("%s changed and the refusal does not name it: %v", tc.count, refusals)
			}
		})
	}
}

// TestMeasurementTimestampFeedsTheWindowComparison is the timestamp's consumer.
// It is not a guard condition — it is the input the window test is taken
// against, so moving it moves which row an entry reaches. Two evaluations over
// the SAME history at two moments must classify differently.
func TestMeasurementTimestampFeedsTheWindowComparison(t *testing.T) {
	t.Parallel()

	fired := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	m := &Manifest{Entries: []Entry{wfEntry(".github/workflows/a.yml", "7d", MeasureFiredAtAll, "")}}
	enum := healthyEnum(".github/workflows/a.yml")
	history := map[string]RunHistory{
		".github/workflows/a.yml": {Runs: []Run{{Conclusion: "success", At: fired}}},
	}

	inside := Evaluate(context.Background(), m, enum, &fakeQuerier{histories: history}, fired.Add(24*time.Hour))
	outside := Evaluate(context.Background(), m, enum, &fakeQuerier{histories: history}, fired.Add(30*24*time.Hour))

	if inside.Decisions[0].Class != ClassOK {
		t.Fatalf("a run one day old classified %q, want %q", inside.Decisions[0].Class, ClassOK)
	}
	if outside.Decisions[0].Class != ClassStale {
		t.Fatalf("a run thirty days old classified %q, want %q", outside.Decisions[0].Class, ClassStale)
	}
}

// TestDroppedDeclaredEntryRefusesTheAllClear is the declared count's guard, and
// it closes a real hole rather than decorating the criterion: an evaluator that
// silently drops a declared entry — the "drop an unreadable-kind entry" mutant
// M2 probed — emits fewer findings, none of them wrong, and would otherwise
// report all-clear about a set smaller than the one it was given.
func TestDroppedDeclaredEntryRefusesTheAllClear(t *testing.T) {
	t.Parallel()

	counts := map[Classification]int{}
	for _, c := range Classifications() {
		counts[c] = 0
	}
	counts[ClassOK] = 1

	ev := Evaluation{
		MeasuredAt:  time.Now(),
		Enumerated:  2,
		Declared:    2, // two were declared
		Queried:     1,
		IntegrityOK: true,
		Counts:      counts,
		Decisions: []Decision{ // one was decided
			{Subject: "a", Row: 3, Class: ClassOK, Surface: SurfaceOK},
		},
	}

	refusals := ev.Refusals()
	if len(refusals) == 0 {
		t.Fatal("a declared entry was dropped and the all-clear stood")
	}
	if !strings.Contains(strings.ToLower(strings.Join(refusals, " | ")), "declared") {
		t.Fatalf("the refusal does not name the dropped declared entry: %v", refusals)
	}
}

// TestAllClearMatchesTheRefusalSet keeps the boolean and the reasons from
// drifting apart. An all-clear that disagrees with its own stated reasons is an
// unexplained verdict, which the evaluation's own IntegrityFailure field
// already refuses to be.
func TestAllClearMatchesTheRefusalSet(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	for name, build := range map[string]func() (*Manifest, *fakeEnum, *fakeQuerier){
		"all clean": func() (*Manifest, *fakeEnum, *fakeQuerier) {
			return &Manifest{Entries: []Entry{wfEntry(".github/workflows/a.yml", "7d", MeasureFiredAtAll, "")}},
				healthyEnum(".github/workflows/a.yml"),
				&fakeQuerier{histories: map[string]RunHistory{
					".github/workflows/a.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-time.Hour)}}},
				}}
		},
		"one stale": func() (*Manifest, *fakeEnum, *fakeQuerier) {
			return &Manifest{Entries: []Entry{wfEntry(".github/workflows/a.yml", "7d", MeasureFiredAtAll, "")}},
				healthyEnum(".github/workflows/a.yml"),
				&fakeQuerier{histories: map[string]RunHistory{
					".github/workflows/a.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-90 * 24 * time.Hour)}}},
				}}
		},
		"empty enumeration": func() (*Manifest, *fakeEnum, *fakeQuerier) {
			return &Manifest{Entries: []Entry{wfEntry(".github/workflows/a.yml", "7d", MeasureFiredAtAll, "")}},
				&fakeEnum{}, &fakeQuerier{}
		},
	} {
		m, enum, q := build()
		ev := Evaluate(context.Background(), m, enum, q, now)
		if ev.AllClear != (len(ev.Refusals()) == 0) {
			t.Fatalf("%s: AllClear=%v while the refusals are %v", name, ev.AllClear, ev.Refusals())
		}
	}
}
