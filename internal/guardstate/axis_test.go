package guardstate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Fakes. NO test in this package reaches a forge: the per-subject query sits
// behind an interface precisely so the query count is MEASURED rather than
// grepped for, and so a test measures the classifier instead of the network.
// ---------------------------------------------------------------------------

// fakeQuerier records what was asked of it. The repository-global listing is a
// real method on the interface DELIBERATELY: AC-GSM-006 (b) requires a measured
// call count rather than a source grep, and a global listing that cannot be
// called is a clause that cannot fail.
type fakeQuerier struct {
	histories map[string]RunHistory
	errs      map[string]error

	subjectCalls []string
	globalCalls  int
}

func (f *fakeQuerier) RunsForSubject(_ context.Context, locator string) (RunHistory, error) {
	f.subjectCalls = append(f.subjectCalls, locator)
	if err, ok := f.errs[locator]; ok {
		return RunHistory{}, err
	}
	return f.histories[locator], nil
}

func (f *fakeQuerier) AllRuns(context.Context) (RunHistory, error) {
	f.globalCalls++
	return RunHistory{}, nil
}

// fakeEnum is the disk side. Enumerate and Exists are SEPARATE observations of
// the same fact — Exists is the per-subject point test, and it does not consult
// the enumeration, which is what stops it inheriting the enumeration's defect.
type fakeEnum struct {
	files []string
	// onDisk is the ground truth the point test reads. A locator in onDisk but
	// missing from files is exactly a degraded enumeration.
	onDisk map[string]bool
	err    error
}

func (f *fakeEnum) Enumerate() ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.files, nil
}

func (f *fakeEnum) Exists(locator string) (bool, error) {
	if f.onDisk == nil {
		// Default: the point test agrees with the enumeration.
		for _, p := range f.files {
			if p == locator {
				return true, nil
			}
		}
		return false, nil
	}
	return f.onDisk[locator], nil
}

func healthyEnum(locators ...string) *fakeEnum {
	on := map[string]bool{}
	for _, l := range locators {
		on[l] = true
	}
	return &fakeEnum{files: locators, onDisk: on}
}

// entry builds a workflow entry with the given locator and window.
func wfEntry(locator, window string, m Measure, when string) Entry {
	return Entry{
		Kind:         KindGitHubWorkflow,
		Locator:      locator,
		Events:       []string{"push"},
		Window:       window,
		Measure:      m,
		ExpectedWhen: when,
	}
}

func ago(now time.Time, d time.Duration) time.Time { return now.Add(-d) }

func decisionFor(t *testing.T, ev Evaluation, subject string) Decision {
	t.Helper()
	for _, d := range ev.Decisions {
		if d.Subject == subject {
			return d
		}
	}
	t.Fatalf("no decision for %q; the evaluator decided %d entries and none was this one", subject, len(ev.Decisions))
	return Decision{}
}

// ---------------------------------------------------------------------------
// AC-GSM-003 (c) — the three measured quantities yield different qualifying
// sets over ONE recorded fixture.
//
// Mutant this kills: a reader that accepts all three values and treats them
// identically. That mutant satisfies clauses (a) and (b) — which M1 already
// flipped — and makes REQ-GSM-003's one-number-two-events prohibition
// unenforceable: an author declares `verdict-rendered` and silently receives
// `fired-at-all` semantics.
// ---------------------------------------------------------------------------

func TestMeasure_ThreeValuesYieldDifferentQualifyingSets(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	// The fixture the criterion stipulates: a skipped, a cancelled, a success.
	stipulated := RunHistory{Runs: []Run{
		{Conclusion: "skipped", At: ago(now, time.Hour)},
		{Conclusion: "cancelled", At: ago(now, 2*time.Hour)},
		{Conclusion: "success", At: ago(now, 3*time.Hour)},
	}}

	if got := len(stipulated.Qualifying(MeasureFiredAtAll)); got != 3 {
		t.Errorf("fired-at-all admits %d of 3 runs, want all 3", got)
	}
	withEffect := stipulated.Qualifying(MeasureFiredWithEffect)
	if len(withEffect) != 1 || withEffect[0].Conclusion != "success" {
		t.Errorf("fired-with-effect admitted %+v, want the skipped and the cancelled excluded", withEffect)
	}
	verdict := stipulated.Qualifying(MeasureVerdictRendered)
	if len(verdict) != 1 || verdict[0].Conclusion != "success" {
		t.Errorf("verdict-rendered admitted %+v, want only the success", verdict)
	}
}

// TestMeasure_FiredWithEffectIsDistinctFromVerdictRendered is the assertion the
// stipulated fixture cannot carry, and it is recorded as a finding rather than
// smuggled in: over {skipped, cancelled, success} the SECOND and THIRD measures
// admit the SAME set — {success} — so the criterion's own words ("three
// different qualifying sets") are unsatisfiable on the fixture its own `Given`
// stipulates, and a mutant collapsing fired-with-effect into verdict-rendered
// survives it.
//
// The separating case is a run that reached NO conclusion (in progress, or a
// null conclusion): fired-with-effect admits it — it is neither skipped nor
// cancelled — and verdict-rendered does not, because no verdict was rendered.
// One extra run in the fixture is what makes the three sets genuinely three.
func TestMeasure_FiredWithEffectIsDistinctFromVerdictRendered(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	h := RunHistory{Runs: []Run{
		{Conclusion: "skipped", At: ago(now, time.Hour)},
		{Conclusion: "cancelled", At: ago(now, 2*time.Hour)},
		{Conclusion: "success", At: ago(now, 3*time.Hour)},
		{Conclusion: "", At: ago(now, 30*time.Minute)}, // in progress
	}}
	all := len(h.Qualifying(MeasureFiredAtAll))
	effect := len(h.Qualifying(MeasureFiredWithEffect))
	verdict := len(h.Qualifying(MeasureVerdictRendered))
	if all == effect || effect == verdict || all == verdict {
		t.Errorf("qualifying-set sizes %d/%d/%d are not three different sets; a reader treating two of the values identically would pass", all, effect, verdict)
	}
	if effect != 2 {
		t.Errorf("fired-with-effect admitted %d, want 2 (the success and the in-progress run)", effect)
	}
	if verdict != 1 {
		t.Errorf("verdict-rendered admitted %d, want 1 (the success alone)", verdict)
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-005 (a) — the manifest holds its watched set as data: an entry of a
// second, non-workflow kind is ACCEPTED, COUNTED in `declared`, and classified
// `UNREADABLE`.
//
// Mutant this kills: a classifier that drops an entry whose kind it cannot
// read. Dropping is the silent direction — the entry vanishes from `declared`
// and from every per-value count, and the result is green about a subject it
// never judged.
// ---------------------------------------------------------------------------

func TestAxis_SecondKindEntryIsCountedAndUnreadable(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	second := Entry{
		Kind: "policy-rule", Locator: "rules/no-sweep-staging",
		Events: []string{"pre-commit"}, Window: "30d", Measure: MeasureFiredAtAll,
	}
	m := &Manifest{Entries: []Entry{second}}
	q := &fakeQuerier{}
	ev := Evaluate(context.Background(), m, healthyEnum(".github/workflows/ci.yml"), q, now)

	if ev.Declared != 1 {
		t.Errorf("declared = %d, want 1: an entry of an unreadable kind is COUNTED, not dropped", ev.Declared)
	}
	d := decisionFor(t, ev, "rules/no-sweep-staging")
	if d.Class != ClassUnreadable {
		t.Errorf("second-kind entry classified %q, want %q (row 1)", d.Class, ClassUnreadable)
	}
	if len(q.subjectCalls) != 0 {
		t.Errorf("the evaluator queried %v for a kind that has no reader; row 1 precedes the query", q.subjectCalls)
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-007 (a)(b) — the state table is total: ONE CASE PER ROW, each case
// receives EXACTLY ONE classification, and the classifier is table-driven so
// each case's value traces to its row.
//
// Mutant this kills: a classifier with a catch-all default satisfies "every
// case receives exactly one classification" while collapsing rows 1, 2 and 6
// into one value. The per-row `wantRow`/`wantClass` pair is what separates
// them: a collapsed classifier returns the right VALUE from the wrong ROW for
// at least one case, or the wrong value.
// ---------------------------------------------------------------------------

// rowCase is one case per row of the normative table.
type rowCase struct {
	row     int
	name    string
	class   Classification
	build   func(now time.Time) (*Manifest, *fakeEnum, *fakeQuerier)
	subject string
}

func rowCases() []rowCase {
	const wf = ".github/workflows/"
	return []rowCase{
		{
			row: 1, name: "kind has no reader", class: ClassUnreadable, subject: "rules/second-kind",
			build: func(time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := Entry{Kind: "policy-rule", Locator: "rules/second-kind", Events: []string{"pre-commit"}, Window: "30d", Measure: MeasureFiredAtAll}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf + "ci.yml"), &fakeQuerier{}
			},
		},
		{
			row: 2, name: "subject on disk, query did not return", class: ClassUnresolved, subject: wf + "ci.yml",
			build: func(time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := wfEntry(wf+"ci.yml", "7d", MeasureFiredAtAll, "")
				q := &fakeQuerier{errs: map[string]error{wf + "ci.yml": errors.New("HTTP 401: Bad credentials")}}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf + "ci.yml"), q
			},
		},
		{
			row: 3, name: "qualifying run inside the window", class: ClassOK, subject: wf + "ci.yml",
			build: func(now time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := wfEntry(wf+"ci.yml", "7d", MeasureFiredAtAll, "")
				q := &fakeQuerier{histories: map[string]RunHistory{
					wf + "ci.yml": {Runs: []Run{{Conclusion: "success", At: ago(now, 24*time.Hour)}}},
				}}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf + "ci.yml"), q
			},
		},
		{
			row: 4, name: "aged run, condition does not excuse", class: ClassStale, subject: wf + "ci.yml",
			build: func(now time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := wfEntry(wf+"ci.yml", "7d", MeasureFiredAtAll, "")
				q := &fakeQuerier{histories: map[string]RunHistory{
					wf + "ci.yml": {Runs: []Run{{Conclusion: "success", At: ago(now, 40*24*time.Hour)}}},
				}}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf + "ci.yml"), q
			},
		},
		{
			row: 5, name: "declared condition excuses the absence", class: ClassOK, subject: wf + "release.yml",
			build: func(time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := wfEntry(wf+"release.yml", "90d", MeasureVerdictRendered, "release-cycle")
				q := &fakeQuerier{histories: map[string]RunHistory{wf + "release.yml": {}}}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf + "release.yml"), q
			},
		},
		{
			row: 6, name: "no qualifying runs at all, unexcused", class: ClassUnknown, subject: wf + "ci.yml",
			build: func(time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := wfEntry(wf+"ci.yml", "7d", MeasureFiredAtAll, "")
				q := &fakeQuerier{histories: map[string]RunHistory{wf + "ci.yml": {}}}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf + "ci.yml"), q
			},
		},
		{
			row: 7, name: "workflow file on disk with no entry", class: ClassUndeclared, subject: wf + "stray.yml",
			build: func(now time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				e := wfEntry(wf+"ci.yml", "7d", MeasureFiredAtAll, "")
				q := &fakeQuerier{histories: map[string]RunHistory{
					wf + "ci.yml": {Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}},
				}}
				return &Manifest{Entries: []Entry{e}}, healthyEnum(wf+"ci.yml", wf+"stray.yml"), q
			},
		},
		{
			row: 8, name: "declared entry whose subject is gone", class: ClassOrphaned, subject: wf + "deleted.yml",
			build: func(now time.Time) (*Manifest, *fakeEnum, *fakeQuerier) {
				live := wfEntry(wf+"ci.yml", "7d", MeasureFiredAtAll, "")
				gone := wfEntry(wf+"deleted.yml", "7d", MeasureFiredAtAll, "")
				// The enumeration is HEALTHY: it returns the one file that is
				// on disk, and the point test agrees the other is absent.
				enum := healthyEnum(wf + "ci.yml")
				q := &fakeQuerier{histories: map[string]RunHistory{
					wf + "ci.yml":      {Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}},
					wf + "deleted.yml": {Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}},
				}}
				return &Manifest{Entries: []Entry{live, gone}}, enum, q
			},
		},
	}
}

func TestStateTable_EveryRowClassifiesExactlyOnce(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	cases := rowCases()
	if len(cases) != stateTableRowCount {
		t.Fatalf("the fixture supplies %d cases for a %d-row table; a partial fixture proves a partial totality", len(cases), stateTableRowCount)
	}
	seenRows := map[int]bool{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, enum, q := c.build(now)
			ev := Evaluate(context.Background(), m, enum, q, now)

			d := decisionFor(t, ev, c.subject)
			if d.Class != c.class {
				t.Errorf("row %d case classified %q, want %q", c.row, d.Class, c.class)
			}
			// (b) — the value traces to its row. A catch-all default returns a
			// value without a row, or the wrong row's number.
			if d.Row != c.row {
				t.Errorf("row %d case traces to row %d; the classifier is not reading the table it claims to", c.row, d.Row)
			}
			// (a) — EXACTLY ONE. Not zero, not more than one.
			n := 0
			for _, other := range ev.Decisions {
				if other.Subject == c.subject {
					n++
				}
			}
			if n != 1 {
				t.Errorf("subject %q received %d classifications, want exactly 1", c.subject, n)
			}
		})
		seenRows[c.row] = true
	}
	for r := 1; r <= stateTableRowCount; r++ {
		if !seenRows[r] {
			t.Errorf("no case exercises row %d", r)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-008 — every classification value is REACHABLE across the same 8-row
// fixture.
//
// Mutant this kills: an evaluator that never emits `OK` passes every other
// criterion in the set — AC-GSM-009 wants UNRESOLVED, AC-GSM-010 wants UNKNOWN,
// AC-GSM-011 wants UNDECLARED — and nothing else requires a healthy subject to
// read clean. Such an evaluator marks every entry non-clean and turns the
// consuming advisory into an always-red channel.
// ---------------------------------------------------------------------------

func TestStateTable_EveryValueIsReachable(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	produced := map[Classification]bool{}
	for _, c := range rowCases() {
		m, enum, q := c.build(now)
		ev := Evaluate(context.Background(), m, enum, q, now)
		for _, d := range ev.Decisions {
			produced[d.Class] = true
		}
	}
	all := Classifications()
	if len(all) != 7 {
		t.Fatalf("the vocabulary has %d values, want 7; a shrunken vocabulary would make this check pass vacuously", len(all))
	}
	for _, v := range all {
		if !produced[v] {
			t.Errorf("value %q is produced by no case; it is an unused option, not a classification", v)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-009 — a failed query is `UNRESOLVED`, never `UNKNOWN` (inherited T2).
//
// Mutant this kills: a classifier that folds an errored query into UNKNOWN.
// UNKNOWN means retention-bounded absence and its implied action is "look again
// with a longer window", which is wrong advice for an auth failure. Two states
// with different implied actions under one label is the defect that produced
// this SPEC.
// ---------------------------------------------------------------------------

func TestAxis_FailedQueryIsUnresolvedNotUnknown(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	const loc = ".github/workflows/ci.yml"
	for _, shape := range []struct {
		name string
		err  error
	}{
		{"auth failure", errors.New("HTTP 401: Bad credentials")},
		{"rate limit", errors.New("HTTP 403: API rate limit exceeded")},
	} {
		t.Run(shape.name, func(t *testing.T) {
			m := &Manifest{Entries: []Entry{wfEntry(loc, "7d", MeasureFiredAtAll, "")}}
			q := &fakeQuerier{errs: map[string]error{loc: shape.err}}
			ev := Evaluate(context.Background(), m, healthyEnum(loc), q, now)

			d := decisionFor(t, ev, loc)
			if d.Class == ClassUnknown {
				t.Fatalf("an errored query classified UNKNOWN: the implied action becomes 'look again with a longer window', which is wrong advice for %v", shape.err)
			}
			if d.Class != ClassUnresolved {
				t.Errorf("classified %q, want %q", d.Class, ClassUnresolved)
			}
			if ev.Counts[ClassUnresolved] != 1 {
				t.Errorf("per-value count for UNRESOLVED = %d, want 1: the value needs a count of its own or it is invisible", ev.Counts[ClassUnresolved])
			}
			if ev.Queried != 0 {
				t.Errorf("queried = %d, want 0: a query that did not return is not a successful query", ev.Queried)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-010 — excused absence is `OK` whether the runs are ABSENT or merely
// AGED; unexcused absence is `UNKNOWN`.
//
// Fixture (iii) is the v0.8.0 crossed case and it is what the ORDER of the axis
// decides: a declared-quiet entry whose only qualifying run merely PREDATES the
// window. Consulting the declared condition only on the zero-run branch routes
// it to row 4 and reports STALE — a false alarm on a healthy subject, with
// RETENTION selecting the label, since the same real state reads STALE while
// the aged run is visible and OK once retention consumes it.
// ---------------------------------------------------------------------------

func TestAxis_ExcusedAbsenceIsOKAgedOrAbsent(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	const (
		quiet    = ".github/workflows/release.yml"
		unexcuse = ".github/workflows/ci.yml"
		aged     = ".github/workflows/spec-status-auto-sync.yml"
	)
	m := &Manifest{Entries: []Entry{
		wfEntry(quiet, "90d", MeasureVerdictRendered, "release-cycle"),
		wfEntry(unexcuse, "7d", MeasureFiredAtAll, ""),
		wfEntry(aged, "30d", MeasureFiredAtAll, "release-cycle"),
	}}
	q := &fakeQuerier{histories: map[string]RunHistory{
		quiet:    {},
		unexcuse: {},
		aged:     {Runs: []Run{{Conclusion: "success", At: ago(now, 120*24*time.Hour)}}},
	}}
	ev := Evaluate(context.Background(), m, healthyEnum(quiet, unexcuse, aged), q, now)

	if got := decisionFor(t, ev, quiet).Class; got != ClassOK {
		t.Errorf("(i) declared-quiet with zero runs classified %q, want %q", got, ClassOK)
	}
	if got := decisionFor(t, ev, unexcuse).Class; got != ClassUnknown {
		t.Errorf("(ii) unexcused with zero runs classified %q, want %q", got, ClassUnknown)
	}
	third := decisionFor(t, ev, aged)
	if third.Class == ClassStale {
		t.Fatalf("(iii) declared-quiet entry with an AGED run classified STALE: the declared condition is being consulted after the aged-versus-absent distinction, so retention selects the label")
	}
	if third.Class != ClassOK {
		t.Errorf("(iii) classified %q, want %q", third.Class, ClassOK)
	}
	if third.Row != 5 {
		t.Errorf("(iii) traces to row %d, want row 5: it must reach the declared-condition node, not the aged branch", third.Row)
	}
}
