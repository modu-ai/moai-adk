package guardstate

import (
	"context"
	"fmt"
	"time"
)

// Decision is one subject's outcome: which row of the normative table decided
// it, the classification that row carries, and that classification's fold.
//
// Row is not decoration. A classifier with a catch-all default satisfies "every
// case receives exactly one classification" while collapsing several rows into
// one value; carrying the row is what makes the value TRACEABLE to the table
// rather than merely plausible.
type Decision struct {
	Subject string
	Row     int
	Class   Classification
	Surface Surface

	// Expectation is what this subject was declared to do and did not, carried
	// only by the rows whose implied action needs it (axisRule.CarriesExpectation)
	// and nil everywhere else.
	//
	// nil is "nothing was missed", never "the expectation was blank". A value
	// struct could not tell those apart, and an expectation present on every
	// entry says nothing about any of them.
	Expectation *Expectation
}

// Expectation is a declared firing expectation, copied from the manifest entry
// that declared it so it compares equal field-against-field against the source.
type Expectation struct {
	Window  string
	Measure Measure
}

// Evaluation is one evaluation run: what was measured, how much of it was
// measured, and what each subject was decided to be.
type Evaluation struct {
	// MeasuredAt is the measurement's own timestamp.
	MeasuredAt time.Time

	// Enumerated is how many files the disk enumeration returned. It is
	// carried because `UNDECLARED: 0` is otherwise indistinguishable between
	// "the census is complete" and "the evaluator never looked".
	Enumerated int

	// Declared is how many entries the manifest declared.
	Declared int

	// Queried is how many per-subject queries SUCCEEDED. A query that returned
	// an error is not a successful query and is not counted here.
	Queried int

	// Decisions is one decision per declared entry plus one per undeclared
	// enumerated file.
	Decisions []Decision

	// Counts is the per-value count for each classification produced.
	Counts map[Classification]int

	// IntegrityOK is the enumeration-integrity verdict, computed BEFORE the
	// entry axis runs. A failed check SUPPRESSES row 8 for every entry.
	IntegrityOK bool

	// IntegrityFailure names why integrity failed, empty when it passed. An
	// unexplained refusal is not a report.
	IntegrityFailure string

	// AllClear is false whenever any subject is non-clean, the successfully
	// queried count is zero, or the enumeration failed its integrity check.
	AllClear bool
}

// Evaluate classifies every declared entry and every enumerated file that no
// entry declares. It reads and reports: nothing here writes to the working
// tree, commits, pushes, or mutates any forge state (REQ-GSM-011).
//
// The ORDER of the three stages below is normative, not stylistic:
//
//  1. Enumerate, then check the enumeration's INTEGRITY — before the entry axis
//     runs. The axis's leading split reads the enumeration, so it inherits
//     whatever that enumeration got wrong. With the gate placed after the axis,
//     a wrong glob returning 3 of 18 makes the 15 unenumerated entries satisfy
//     the leading split and classify ORPHANED: deletion advice for 15 correct
//     entries.
//
//  2. The entry axis, with row 8 suppressed on a failed check.
//
//  3. The disk axis, which is NOT suppressed. The asymmetry is deliberate,
//     because the two directions fail differently: row 8 emits DESTRUCTIVE
//     advice on a subject that is in fact present, while row 7 merely emits
//     FEWER findings and every finding it does emit is still sound. The
//     completeness row 7 loses is what the refused all-clear reports.
func Evaluate(ctx context.Context, m *Manifest, enum Enumerator, q RunQuerier, now time.Time) Evaluation {
	ev := Evaluation{MeasuredAt: now, Counts: map[Classification]int{}}

	// Seed a count for EVERY value of the closed vocabulary, including the ones
	// this run produces none of. A map that acquires a key only when the
	// classifier emits that value carries "no count" and "zero" identically —
	// the same two-states-one-representation defect this SPEC exists to close,
	// one layer down in the reporting.
	for _, c := range Classifications() {
		ev.Counts[c] = 0
	}

	var entries []Entry
	if m != nil {
		entries = m.Entries
	}
	ev.Declared = len(entries)

	files, enumErr := enum.Enumerate()
	enumerated := map[string]bool{}
	for _, f := range files {
		enumerated[f] = true
	}
	ev.Enumerated = len(enumerated)

	// ---- Stage 1: the enumeration-integrity gate -------------------------
	//
	// The tolerance is a SECOND, INDEPENDENT observation of the disk — a
	// per-subject point test of each named locator — and not a term computed
	// from the evaluator's own findings. Any term computed from the
	// classifications is circular: an entry missing from the enumeration is
	// exactly what produces an ORPHANED finding, so the miss count and the
	// finding count move together by construction and their difference is
	// identically zero for EVERY partial enumeration. Corroboration breaks the
	// circle because it cannot inherit the enumeration's own defect.
	pointTestAbsent := map[string]bool{}
	switch {
	case enumErr != nil:
		ev.IntegrityFailure = fmt.Sprintf("the disk enumeration failed: %v", enumErr)
	case ev.Enumerated == 0:
		// Stated as an INTEGRITY failure rather than only as an all-clear
		// refusal, so row 8 is suppressed here too: a wrong working directory
		// enumerates nothing, every locator corroborates as absent against the
		// same wrong root, and every declared entry would otherwise classify
		// ORPHANED on a result that was merely refused its all-clear.
		ev.IntegrityFailure = "the disk enumeration returned zero files"
	default:
		ev.IntegrityOK = true
		for _, e := range entries {
			if !kindIsDiskEnumerable(e.Kind) || enumerated[e.Locator] {
				continue
			}
			present, err := enum.Exists(e.Locator)
			if err != nil {
				ev.IntegrityOK = false
				ev.IntegrityFailure = fmt.Sprintf(
					"the corroborating existence test for %q could not be completed: %v", e.Locator, err)
				break
			}
			if present {
				ev.IntegrityOK = false
				ev.IntegrityFailure = fmt.Sprintf(
					"declared subject %q is absent from the enumeration and PRESENT under a direct existence test of its own locator: the enumeration is incomplete, not the subject deleted",
					e.Locator)
				break
			}
			pointTestAbsent[e.Locator] = true
		}
	}

	// ---- Stage 2: the entry axis ----------------------------------------
	for _, e := range entries {
		entry := e
		obs := &entryObservation{
			entry: entry,
			// The CARRIED timestamp, not the parameter. Every window
			// comparison is taken against it, which is what makes it an input
			// with a consumer rather than an annotation on the report.
			now:             ev.MeasuredAt,
			integrityOK:     ev.IntegrityOK,
			absentFromEnum:  !enumerated[entry.Locator],
			pointTestAbsent: pointTestAbsent[entry.Locator],
			query: func() (RunHistory, error) {
				return q.RunsForSubject(ctx, entry.Locator)
			},
		}
		d := classifyEntry(obs)
		if obs.queried && obs.queryErr == nil {
			ev.Queried++
		}
		ev.Decisions = append(ev.Decisions, d)
	}

	// ---- Stage 3: the disk axis (never suppressed) -----------------------
	declared := map[string]bool{}
	for _, e := range entries {
		declared[e.Locator] = true
	}
	for _, f := range files {
		if declared[f] {
			continue
		}
		ev.Decisions = append(ev.Decisions, Decision{
			Subject: f,
			Row:     7,
			Class:   ClassUndeclared,
			Surface: ClassUndeclared.Fold(),
		})
	}

	// ---- Counts and the all-clear ---------------------------------------
	for _, d := range ev.Decisions {
		ev.Counts[d.Class]++
	}
	ev.AllClear = len(ev.Refusals()) == 0

	return ev
}

// rowUndeclared is the disk axis's row. Decisions carrying it are the files no
// entry declared, which is why they are excluded when the declared set's
// coverage is checked.
const rowUndeclared = 7

// Refusals names every reason this result is refused its all-clear, reading the
// CARRIED counts rather than recomputing from the decisions.
//
// Reading the carried values is the point, not an implementation detail. Every
// count this evaluation reports is required to have a consumer that READS it —
// a decision it changes or a guard condition it feeds — because the parent
// card's audit found two counts that were required, rendered, and read by
// nothing. A count recomputed from the decisions here would leave the carried
// field inert while looking consumed.
//
// It returns reasons rather than a boolean for the same reason IntegrityFailure
// carries text: an unexplained refusal is not a report.
func (e Evaluation) Refusals() []string {
	var out []string

	if !e.IntegrityOK {
		reason := e.IntegrityFailure
		if reason == "" {
			reason = "the disk enumeration failed its integrity check"
		}
		out = append(out, reason)
	}

	// The enumerated-files count. Zero files is not a complete census of an
	// empty repository — it is indistinguishable from never having looked.
	if e.Enumerated == 0 {
		out = append(out, "the enumeration returned zero files, so an all-clear would be about a set nothing read")
	}

	// The queried count. Counting only queries that SUCCEEDED is what makes
	// this a guard: a run in which every query errored asked nothing.
	if e.Queried == 0 {
		out = append(out, "zero subjects were queried successfully")
	}

	// The declared count, against the entries actually decided. An evaluator
	// that silently drops a declared entry emits fewer findings, none of them
	// wrong, and would otherwise report all-clear about a smaller set than it
	// was given.
	decided := 0
	for _, d := range e.Decisions {
		if d.Row != rowUndeclared {
			decided++
		}
	}
	if decided != e.Declared {
		out = append(out, fmt.Sprintf(
			"%d entries were declared and %d were decided: the evaluation does not cover the declared set",
			e.Declared, decided))
	}

	// The per-value counts. Summing the non-clean values is what gives every
	// one of the seven a reading consumer, rather than only the values this
	// run happened to produce.
	nonClean := 0
	for class, n := range e.Counts {
		if !class.IsClean() {
			nonClean += n
		}
	}
	if nonClean > 0 {
		out = append(out, fmt.Sprintf("%d subject(s) are not clean", nonClean))
	}

	return out
}
