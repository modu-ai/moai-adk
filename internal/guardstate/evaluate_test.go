package guardstate

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// AC-GSM-006 — per-subject query only: exactly N per-subject queries, one per
// entry, and ZERO repository-global run listings.
//
// Mutant this kills: an evaluator that issues N targeted queries AND ALSO one
// global listing used as a fast path, which reintroduces the exact failure for
// whichever subject the global listing hides. Clause (b) is a measured call
// count rather than a source grep, because a grep for a call site is satisfied
// by a mutant building the same request from string fragments — so the global
// listing is a callable method on the interface and the fake counts it.
// ---------------------------------------------------------------------------

func TestEvaluate_PerSubjectQueriesOnly(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	const n = 5
	var entries []Entry
	locators := make([]string, 0, n)
	histories := map[string]RunHistory{}
	for i := 0; i < n; i++ {
		loc := fmt.Sprintf(".github/workflows/w%d.yml", i)
		locators = append(locators, loc)
		entries = append(entries, wfEntry(loc, "7d", MeasureFiredAtAll, ""))
		histories[loc] = RunHistory{Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}}
	}
	q := &fakeQuerier{histories: histories}
	ev := Evaluate(context.Background(), &Manifest{Entries: entries}, healthyEnum(locators...), q, now)

	if len(q.subjectCalls) != n {
		t.Errorf("(a) %d per-subject queries issued, want exactly %d — one per entry", len(q.subjectCalls), n)
	}
	seen := map[string]int{}
	for _, c := range q.subjectCalls {
		seen[c]++
	}
	for _, loc := range locators {
		if seen[loc] != 1 {
			t.Errorf("(a) subject %q was queried %d times, want exactly 1", loc, seen[loc])
		}
	}
	if q.globalCalls != 0 {
		t.Errorf("(b) %d repository-global run listings issued, want 0: a global listing is measurably incapable of answering for a low-frequency subject, and its incapacity is invisible from inside it", q.globalCalls)
	}
	if ev.Queried != n {
		t.Errorf("queried = %d, want %d", ev.Queried, n)
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-011 — the set comparison runs in BOTH directions at runtime.
//
// Mutant this kills: a criterion asserting only "appears in the output" is
// satisfied by an evaluator that lists the finding under a heading and still
// returns all-clear. The non-all-clear clause is what makes the classification
// consequential. Fixture (ii) is the direction that was missing: run history
// OUTLIVES a deleted workflow, so without the entry→disk direction a declared
// entry whose subject is gone is decided by stale history into STALE — or, on a
// recent enough run, into a FALSE `OK`.
// ---------------------------------------------------------------------------

func TestEvaluate_SetComparisonBothDirections(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	const (
		declaredLive = ".github/workflows/ci.yml"
		strayOnDisk  = ".github/workflows/stray.yml"
		deletedSubj  = ".github/workflows/deleted.yml"
	)
	fresh := RunHistory{Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}}

	t.Run("(i) disk file with no manifest entry is UNDECLARED", func(t *testing.T) {
		m := &Manifest{Entries: []Entry{wfEntry(declaredLive, "7d", MeasureFiredAtAll, "")}}
		q := &fakeQuerier{histories: map[string]RunHistory{declaredLive: fresh}}
		ev := Evaluate(context.Background(), m, healthyEnum(declaredLive, strayOnDisk), q, now)

		d := decisionFor(t, ev, strayOnDisk)
		if d.Class != ClassUndeclared {
			t.Errorf("classified %q, want %q", d.Class, ClassUndeclared)
		}
		if ev.Counts[ClassUndeclared] != 1 {
			t.Errorf("UNDECLARED count = %d, want 1", ev.Counts[ClassUndeclared])
		}
		if ev.AllClear {
			t.Error("the run reported an all-clear while naming an undeclared subject; the classification has to be consequential, not decorative")
		}
	})

	t.Run("(ii) manifest entry whose subject is absent from disk is ORPHANED", func(t *testing.T) {
		m := &Manifest{Entries: []Entry{
			wfEntry(declaredLive, "7d", MeasureFiredAtAll, ""),
			wfEntry(deletedSubj, "7d", MeasureFiredAtAll, ""),
		}}
		// Run history OUTLIVES the deletion — the fresh run is exactly what
		// would produce a false `OK` if the entry→disk direction were missing.
		q := &fakeQuerier{histories: map[string]RunHistory{declaredLive: fresh, deletedSubj: fresh}}
		ev := Evaluate(context.Background(), m, healthyEnum(declaredLive), q, now)

		d := decisionFor(t, ev, deletedSubj)
		if d.Class == ClassOK || d.Class == ClassStale {
			t.Fatalf("a declared entry whose subject is gone classified %q — decided by run history that outlives the deletion", d.Class)
		}
		if d.Class != ClassOrphaned {
			t.Errorf("classified %q, want %q", d.Class, ClassOrphaned)
		}
		if ev.Counts[ClassOrphaned] != 1 {
			t.Errorf("ORPHANED count = %d, want 1", ev.Counts[ClassOrphaned])
		}
		if ev.AllClear {
			t.Error("the run reported an all-clear while naming an orphaned entry")
		}
	})
}

// ---------------------------------------------------------------------------
// AC-GSM-012 — a degraded enumeration cannot report all-clear.
//
// This is the criterion the IMPLEMENTATION ORDER decides. The fixture's every
// declared subject is PRESENT on disk and its run history would otherwise
// classify every entry `OK`; the only thing wrong is the enumeration.
//
// Mutant this kills, and it is the card's own defect one layer in: an evaluator
// whose enumeration silently returns a PARTIAL set has a non-zero queried
// count, so a zero-query guard does not bite; yields no UNDECLARED finding for
// the files it never saw; and reports all-clear about a set that had silently
// become the wrong set. And with the integrity gate placed AFTER the entry
// axis, the same partial enumeration makes the unseen-but-present entries
// satisfy row 8's leading split and classify ORPHANED — DELETION ADVICE for
// correct entries — which is why clause (c) is a per-value count of zero rather
// than an inspection of the entries the fixture happens to name.
// ---------------------------------------------------------------------------

func degradedFixture(now time.Time, enumerated int) (*Manifest, *fakeEnum, *fakeQuerier) {
	const total = 18
	var entries []Entry
	onDisk := map[string]bool{}
	histories := map[string]RunHistory{}
	var files []string
	for i := 0; i < total; i++ {
		loc := fmt.Sprintf(".github/workflows/w%02d.yml", i)
		entries = append(entries, wfEntry(loc, "7d", MeasureFiredAtAll, ""))
		// EVERY declared subject is present on disk. Only the enumeration is
		// wrong.
		onDisk[loc] = true
		histories[loc] = RunHistory{Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}}
		if i < enumerated {
			files = append(files, loc)
		}
	}
	return &Manifest{Entries: entries},
		&fakeEnum{files: files, onDisk: onDisk},
		&fakeQuerier{histories: histories}
}

func TestEvaluate_DegradedEnumerationCannotReportAllClear(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name       string
		enumerated int
	}{
		{"(i) zero files — wrong working directory or a permissions failure", 0},
		{"(ii) non-zero but partial — 3 of 18, a wrong glob", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, enum, q := degradedFixture(now, tc.enumerated)
			ev := Evaluate(context.Background(), m, enum, q, now)

			// (a)
			if ev.AllClear {
				t.Error("(a) reported an all-clear on a degraded enumeration — honestly green about a set it never learned")
			}
			// (b)
			if ev.Enumerated != tc.enumerated {
				t.Errorf("(b) enumerated = %d, want %d: the count has to be visible in the output, or `UNDECLARED: 0` is indistinguishable between 'census complete' and 'the evaluator never looked'", ev.Enumerated, tc.enumerated)
			}
			// (c) — refused on the ENUMERATION-INTEGRITY check, not a zero test.
			if ev.IntegrityOK {
				t.Errorf("(c) the enumeration passed its integrity check; reason recorded: %q", ev.IntegrityFailure)
			}
			if ev.IntegrityFailure == "" {
				t.Error("(c) integrity failed with no recorded reason; an unexplained refusal is not a report")
			}
			// (c) — no entry in EITHER run classifies ORPHANED, verified as a
			// per-value count of zero.
			if n := ev.Counts[ClassOrphaned]; n != 0 {
				t.Errorf("(c) %d entries classified ORPHANED on a degraded enumeration: that is deletion advice for entries whose subjects are present on disk", n)
			}
		})
	}

	// The partial case is the one a zero-check cannot see. Stated separately so
	// a regression that reduces the guard to a zero test fails HERE with its
	// own message rather than inside the table above.
	t.Run("a zero-check alone is the card's own defect at a lower dose", func(t *testing.T) {
		m, enum, q := degradedFixture(now, 3)
		ev := Evaluate(context.Background(), m, enum, q, now)
		if ev.Enumerated == 0 {
			t.Fatal("the fixture is meant to be non-zero; a zero fixture would not test what this case exists for")
		}
		if ev.IntegrityOK {
			t.Error("3 of 18 files passed integrity: a non-zero enumeration is not a complete one, and the 15 unseen subjects yield no UNDECLARED finding")
		}
	})

	// The legitimate case must stay REACHABLE. Refusing every absent subject
	// would close the false ORPHANED by making the true one unreachable, which
	// is relocation rather than repair.
	t.Run("a healthy enumeration with a genuinely deleted subject still reaches row 8", func(t *testing.T) {
		const live = ".github/workflows/ci.yml"
		const gone = ".github/workflows/deleted.yml"
		m := &Manifest{Entries: []Entry{
			wfEntry(live, "7d", MeasureFiredAtAll, ""),
			wfEntry(gone, "7d", MeasureFiredAtAll, ""),
		}}
		q := &fakeQuerier{histories: map[string]RunHistory{
			live: {Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}},
		}}
		ev := Evaluate(context.Background(), m, healthyEnum(live), q, now)
		if !ev.IntegrityOK {
			t.Fatalf("a healthy enumeration failed integrity (%q); the point test agrees the subject is absent, which is the passing case", ev.IntegrityFailure)
		}
		if got := decisionFor(t, ev, gone).Class; got != ClassOrphaned {
			t.Errorf("the genuinely deleted subject classified %q, want %q — suppressing it here would be relocation rather than repair", got, ClassOrphaned)
		}
	})
}

// TestEvaluate_ZeroQueriedRefusesAllClear pins REQ-GSM-010's other half: the
// all-clear is refused while the successfully-queried count is zero, separately
// from the enumeration's integrity.
func TestEvaluate_ZeroQueriedRefusesAllClear(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	const loc = ".github/workflows/ci.yml"
	m := &Manifest{Entries: []Entry{wfEntry(loc, "7d", MeasureFiredAtAll, "")}}
	q := &fakeQuerier{errs: map[string]error{loc: errors.New("HTTP 401: Bad credentials")}}
	ev := Evaluate(context.Background(), m, healthyEnum(loc), q, now)

	if ev.Queried != 0 {
		t.Fatalf("queried = %d, want 0 for the fixture", ev.Queried)
	}
	if ev.AllClear {
		t.Error("reported an all-clear with a zero successfully-queried count")
	}
}

// TestEvaluate_HealthyRunReportsAllClear is the paired direction. Without it, an
// evaluator that NEVER reports all-clear passes every refusal criterion above
// and turns the consuming advisory into an always-red channel.
func TestEvaluate_HealthyRunReportsAllClear(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	const a = ".github/workflows/ci.yml"
	const b = ".github/workflows/release.yml"
	m := &Manifest{Entries: []Entry{
		wfEntry(a, "7d", MeasureFiredAtAll, ""),
		wfEntry(b, "90d", MeasureVerdictRendered, "release-cycle"),
	}}
	q := &fakeQuerier{histories: map[string]RunHistory{
		a: {Runs: []Run{{Conclusion: "success", At: ago(now, time.Hour)}}},
		b: {},
	}}
	ev := Evaluate(context.Background(), m, healthyEnum(a, b), q, now)

	if !ev.IntegrityOK {
		t.Fatalf("healthy fixture failed integrity: %q", ev.IntegrityFailure)
	}
	if !ev.AllClear {
		t.Errorf("a healthy repository did not report all-clear; decisions: %+v", ev.Decisions)
	}
	if ev.Declared != 2 || ev.Enumerated != 2 || ev.Queried != 2 {
		t.Errorf("declared/enumerated/queried = %d/%d/%d, want 2/2/2", ev.Declared, ev.Enumerated, ev.Queried)
	}
	if ev.MeasuredAt.IsZero() {
		t.Error("the result carries no measurement timestamp")
	}
}
