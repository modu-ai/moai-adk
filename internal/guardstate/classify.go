package guardstate

import (
	"context"
	"time"
)

// stateTableRowCount is the number of rows in the normative state table
// (spec.md §C.2, REQ-GSM-006). The tests assert it against the artifact rather
// than trusting it here — a table-driven classifier whose row count drifts from
// the table it claims to implement is a paraphrase again.
const stateTableRowCount = 8

// KindGitHubWorkflow is the one subject kind that ships a reader.
const KindGitHubWorkflow = "github-workflow"

// kindHasReader reports whether this deliverable can read the subject's history
// at all. Row 1 decides on it.
func kindHasReader(kind string) bool { return kind == KindGitHubWorkflow }

// kindIsDiskEnumerable reports whether a disk enumeration can judge the subject
// present or absent. Row 8's leading split is guarded on it.
//
// It is a SEPARATE predicate from kindHasReader deliberately, even though one
// kind ships and the two therefore agree today. They answer different
// questions — "can its history be read?" against "can its presence be seen?" —
// and a second kind could answer them differently. Collapsing them would make
// the difference unrecoverable at the moment it first mattered.
func kindIsDiskEnumerable(kind string) bool { return kind == KindGitHubWorkflow }

// Classification is one value of the closed classification vocabulary
// (REQ-GSM-007). The set is closed at seven and every value is reachable from
// at least one row of the state table.
type Classification string

const (
	// ClassOK is the CLEAN value — the only one that denotes nothing to
	// report. Rows 3 and 5 both produce it, which is why the vocabulary has
	// seven values over eight rows.
	ClassOK Classification = "OK"

	// ClassStale is row 4: the subject stopped firing.
	ClassStale Classification = "STALE"

	// ClassUnknown is row 6: no qualifying runs at all, and nothing excuses it.
	// Its implied action is "look again with a longer window" — retention may
	// have consumed the history — which is why an ERRORED query must never
	// land here.
	ClassUnknown Classification = "UNKNOWN"

	// ClassUndeclared is row 7: a workflow file on disk that no entry names.
	ClassUndeclared Classification = "UNDECLARED"

	// ClassUnreadable is row 1: the entry's kind has no reader, so the
	// comparison never applied. MEANINGLESS rather than incomplete, which is
	// why it alone among the non-clean values folds to `ok`.
	ClassUnreadable Classification = "UNREADABLE"

	// ClassUnresolved is row 2: the query could not be completed for a subject
	// that still exists and may therefore succeed later.
	ClassUnresolved Classification = "UNRESOLVED"

	// ClassOrphaned is row 8: a declared entry whose subject is gone, and no
	// query outcome can rehabilitate it.
	ClassOrphaned Classification = "ORPHANED"
)

// Surface is the three-value surface vocabulary the fold targets. It mirrors
// the consuming CheckStatus vocabulary (`ok` / `warn` / `fail`) and has no
// skipped state.
type Surface string

const (
	SurfaceOK   Surface = "ok"
	SurfaceWarn Surface = "warn"

	// SurfaceFail exists so the never-emit rule is expressible and testable.
	// NO classification folds to it (REQ-GSM-013): this SPEC's consumer renders
	// a non-blocking advisory, and a `fail` on a routine sweep would promote a
	// healthy installation to a failing exit status.
	SurfaceFail Surface = "fail"
)

// Classifications returns the closed vocabulary, in table order.
func Classifications() []Classification {
	return []Classification{
		ClassOK, ClassStale, ClassUnknown, ClassUndeclared,
		ClassUnreadable, ClassUnresolved, ClassOrphaned,
	}
}

// Fold is the classification's surface value, per the `Surface fold` column of
// the normative table.
//
// The boundary between the two arms is MEANINGLESS against INCOMPLETE, and it
// is where the leniency principle borrowed from the binary-freshness check
// stops. UNREADABLE folds to `ok` because the comparison never applied;
// UNKNOWN, UNRESOLVED and ORPHANED fold to `warn` because it applied and could
// not be completed. Folding the incomplete cases to `ok` would reproduce this
// SPEC's own subject inside its solution — a mechanism reporting green about a
// set it never learned.
func (c Classification) Fold() Surface {
	switch c {
	case ClassOK, ClassUnreadable:
		return SurfaceOK
	case ClassStale, ClassUnknown, ClassUndeclared, ClassUnresolved, ClassOrphaned:
		return SurfaceWarn
	default:
		// An unrecognised value is not green. It is also not `fail`, because
		// the never-emit rule binds the whole fold and not only the values
		// enumerated above.
		return SurfaceWarn
	}
}

// IsClean reports whether the value denotes nothing to report. EXACTLY ONE
// value does.
//
// This is the axis that separates UNREADABLE from OK: both fold to `ok` and
// both imply no action, so a consumer partitioning on the FOLD would treat
// UNREADABLE as nothing-to-report. The cleanliness axis is what the result's
// designation carries to the consumer instead.
func (c Classification) IsClean() bool { return c == ClassOK }

// Run is one recorded run of a subject. An empty Conclusion means the run
// reached none — still in progress, or reported null — which is what separates
// fired-with-effect from verdict-rendered.
type Run struct {
	Conclusion string
	At         time.Time
}

// RunHistory is a subject's recorded runs.
type RunHistory struct{ Runs []Run }

// Qualifying returns the runs the measured quantity admits. The three values
// admit three different sets, which is what stops one number being asked to
// measure both whether a guard fired and whether a firing caught anything.
func (h RunHistory) Qualifying(m Measure) []Run {
	var out []Run
	for _, r := range h.Runs {
		if m.Admits(r.Conclusion) {
			out = append(out, r)
		}
	}
	return out
}

// RunQuerier reads run history.
//
// AllRuns — the repository-global listing — is on this interface DELIBERATELY
// and is never called by the evaluator. REQ-GSM-005 forbids deriving any
// subject's last-fired time from a global listing, because the listing is
// measurably incapable of answering for a low-frequency subject and its
// incapacity is invisible from inside it. Keeping the method CALLABLE is what
// lets AC-GSM-006 (b) be a measured call count rather than a source grep: a
// mutant that builds the same request from string fragments would defeat a
// grep, and cannot defeat a counter.
type RunQuerier interface {
	RunsForSubject(ctx context.Context, locator string) (RunHistory, error)
	AllRuns(ctx context.Context) (RunHistory, error)
}

// Enumerator is the disk side of the set comparison.
//
// Enumerate and Exists are two INDEPENDENT observations of the same fact, and
// the independence is the whole of the integrity check. Exists is a per-subject
// point test of one named locator; it must not be implemented by re-running or
// consulting the enumeration, or it inherits the enumeration's own defect and
// the check becomes circular.
type Enumerator interface {
	Enumerate() ([]string, error)
	Exists(locator string) (bool, error)
}

// entryObservation is what the entry axis reads about one entry.
//
// The query is LAZY and memoised. Rows 8 and 1 are decided ahead of it and must
// not issue one — row 8 because it is independent of what any query would
// return, row 1 because the kind has no reader to query with.
type entryObservation struct {
	entry Entry
	now   time.Time

	// integrityOK is the enumeration's verdict, computed BEFORE the axis runs.
	integrityOK bool
	// absentFromEnum is whether the enumeration returned this entry's locator.
	absentFromEnum bool
	// pointTestAbsent is the corroborating per-subject existence test.
	pointTestAbsent bool

	query func() (RunHistory, error)

	queried  bool
	history  RunHistory
	queryErr error
}

func (o *entryObservation) ensureQueried() {
	if o.queried {
		return
	}
	o.queried = true
	h, err := o.query()
	o.history = h
	o.queryErr = err
}

// hasQualifyingInsideWindow reports whether a qualifying run falls inside the
// declared expectation window.
//
// An uninterpretable window answers false. ParseManifest rejects such an entry
// by name, so one cannot arrive through the file; a hand-built entry that
// carries one is treated as having no qualifying run inside a window that could
// not be read, which routes it onward rather than reading clean.
func (o *entryObservation) hasQualifyingInsideWindow() bool {
	d, err := o.entry.WindowDuration()
	if err != nil {
		return false
	}
	cutoff := o.now.Add(-d)
	for _, r := range o.history.Qualifying(o.entry.Measure) {
		if r.At.After(cutoff) {
			return true
		}
	}
	return false
}

func (o *entryObservation) hasQualifyingRunsAtAll() bool {
	return len(o.history.Qualifying(o.entry.Measure)) > 0
}

// axisRule is one row of the normative entry axis. The SLICE ORDER is the split
// order, and the split order is normative rather than stylistic: more than one
// row's condition can hold at once, so totality over rows is not enough on its
// own — every entry must also reach exactly one row, and that is what the order
// decides.
type axisRule struct {
	Row   int
	Class Classification
	When  func(o *entryObservation) bool

	// CarriesExpectation marks the row whose implied action is not actionable
	// without the expectation the subject missed. Row 4 says "the subject
	// stopped firing"; without the declared window and measured quantity the
	// operator is told to investigate and is not told what it missed, which is
	// half a comparison.
	//
	// It is a column of the table rather than a row-number comparison in the
	// carrier, so a mutant that drops the obligation is visible HERE, next to
	// the row it belongs to, instead of hidden in a `== 4` somewhere else.
	CarriesExpectation bool
}

// entryAxis is the entry axis of spec.md §C.2, in its normative order.
//
// Two orderings in this slice are load-bearing and have each been wrong once:
//
//   - Row 8 LEADS. It is the only permanently non-retryable condition on the
//     axis: once the subject is gone, every later row's implied action is wrong
//     for it. Reading the axis in any other order routes such an entry to row 2
//     and hands the operator "retry; check credentials" for a subject no retry
//     can bring back. It is also gated on the enumeration's integrity, computed
//     before this axis runs — a leading split that reads an unchecked input
//     decides every entry on it.
//
//   - Row 5 precedes rows 4 and 6. The declared condition is consulted BEFORE
//     the aged-versus-absent distinction. Testing "are there qualifying runs?"
//     first means a correctly-quiet subject whose last run is merely AGED never
//     reaches the sentence that excuses it, and classifies STALE — a false
//     alarm on a healthy subject, with RETENTION selecting the label, since the
//     same real state reads STALE while the aged run is visible and OK once
//     retention consumes it.
func entryAxis() []axisRule {
	return []axisRule{
		{Row: 8, Class: ClassOrphaned, When: func(o *entryObservation) bool {
			return o.integrityOK &&
				kindIsDiskEnumerable(o.entry.Kind) &&
				o.absentFromEnum &&
				o.pointTestAbsent
		}},
		{Row: 1, Class: ClassUnreadable, When: func(o *entryObservation) bool {
			return !kindHasReader(o.entry.Kind)
		}},
		{Row: 2, Class: ClassUnresolved, When: func(o *entryObservation) bool {
			o.ensureQueried()
			return o.queryErr != nil
		}},
		{Row: 3, Class: ClassOK, When: func(o *entryObservation) bool {
			o.ensureQueried()
			return o.hasQualifyingInsideWindow()
		}},
		{Row: 5, Class: ClassOK, When: func(o *entryObservation) bool {
			return o.entry.IsConditional()
		}},
		{Row: 4, Class: ClassStale, CarriesExpectation: true, When: func(o *entryObservation) bool {
			return o.hasQualifyingRunsAtAll()
		}},
		// Row 6 is the terminal branch. The axis is total by construction
		// because this predicate is unconditionally true: no entry can fall off
		// the end of it.
		{Row: 6, Class: ClassUnknown, When: func(*entryObservation) bool { return true }},
	}
}

// classifyEntry walks the axis in order and returns the FIRST row whose
// condition holds.
func classifyEntry(o *entryObservation) Decision {
	for _, rule := range entryAxis() {
		if rule.When(o) {
			d := Decision{
				Subject: o.entry.Locator,
				Row:     rule.Row,
				Class:   rule.Class,
				Surface: rule.Class.Fold(),
			}
			if rule.CarriesExpectation {
				// COPIED from the entry, not described. A constant string or a
				// paraphrase satisfies a presence check while telling the
				// operator nothing they can compare against the manifest.
				d.Expectation = &Expectation{Window: o.entry.Window, Measure: o.entry.Measure}
			}
			return d
		}
	}
	// Unreachable while row 6's predicate is unconditional. Stated rather than
	// panicked: a totality claim that crashes when it is wrong is worse than
	// one that reports the gap.
	return Decision{Subject: o.entry.Locator, Row: 0, Class: ClassUnknown, Surface: ClassUnknown.Fold()}
}
