// Package guardliveness surfaces guard firing-liveness to an operator who did
// not know to ask (SPEC-GUARD-LIVENESS-001, card t333).
//
// The package owns the SURFACING half of the event-history axis: how an
// evaluation result reaches a reader unbidden. It owns none of the state model
// — which classifications exist, what each means, how an entry is decided into
// one, and everything that produces a result belongs to
// SPEC-GUARD-STATE-MODEL-001 and is consumed here through the three-clause
// contract below (spec.md §B.1).
//
// The contract, in full:
//
//   - every entry in an evaluation result carries exactly one classification;
//   - exactly one value of that vocabulary means nothing to report — the clean
//     value;
//   - the result carries a machine-readable designation of WHICH value that is,
//     and the clean/non-clean partition is derived from that designation and by
//     no other means.
//
// Nothing here names a classification value, states how many exist, or decides
// any entry into any of them. A change to the vocabulary's size or to any of
// its value names requires no change to this package.
package guardliveness

import (
	"errors"
	"fmt"
	"strings"
)

// Entry is one subject's outcome inside an evaluation result.
//
// The JSON tags are load-bearing rather than decorative: a result travels from
// the refresh that produced it to the LATER activation that renders it through
// the store (store.go), and a shape that does not survive that round trip would
// arrive as a contract violation that never happened.
type Entry struct {
	// Subject names what was evaluated.
	Subject string `json:"subject"`

	// Classifications carries the entry's classification. The contract says
	// exactly one; the field is a slice so a result that violates that clause
	// is representable and can be refused (Partition) rather than silently
	// mis-partitioned. An entry carrying two classifications matches the clean
	// value on comparison and would be treated as nothing-to-report — it
	// under-fires rather than failing loudly, which is the failure this
	// representation exists to make visible.
	Classifications []string `json:"classifications"`

	// Surface is the producer's folded surface value, carried through and
	// DELIBERATELY never consulted. More than one classification folds to the
	// clean surface value while only one classification is clean, so a
	// partition taken from the fold silently treats every entry in that gap as
	// nothing-to-report. The field exists so a result carrying such an entry is
	// representable — that is the fixture AC-GDL-001 clause (c) requires — not
	// so the partition can read it.
	Surface string `json:"surface"`

	// Expectation is what this subject was declared to do and did not, carried
	// so an operator reading a non-clean entry is told what it missed rather
	// than only told to investigate.
	//
	// A POINTER, and nil means the entry carries no missed expectation. The
	// distinction is the same one Result.Clean is a pointer for: a value struct
	// would make "nothing was missed" indistinguishable from "the expectation
	// was blank", and a later reader would have no way to tell an entry that
	// was never a missed expectation from one whose producer filled the fields
	// with nothing. It is also what lets a result persisted before this field
	// existed decode as absent rather than as blank.
	//
	// Nothing in this package reads it. It is carried for the same reason
	// Surface is — the producer's obligation is to emit it (its own
	// AC-GSM-007 clause (d)); showing it is a render concern this package's
	// two render sites do not yet cover.
	Expectation *Expectation `json:"expectation"`
}

// Expectation is a declared firing expectation, copied from the producer's
// manifest entry rather than described.
//
// Both fields are plain strings and this package attaches no meaning to either
// value: it names no vocabulary, parses no window, and compares nothing. The
// producer owns what they mean, which is what keeps a change to either
// vocabulary off this package.
type Expectation struct {
	// Window is the declared expectation window, verbatim.
	Window string `json:"window"`

	// Measure is the declared measured quantity, verbatim.
	Measure string `json:"measure"`
}

// Designation carries which value of the producer's vocabulary is the clean
// one. Values holds exactly one entry on a conforming result; zero (null) and
// more than one (multi-valued) are both contract violations and are
// representable so they can be refused.
type Designation struct {
	Values []string `json:"values"`
}

// Result is one evaluation, as consumed by the advisory.
type Result struct {
	// Entries is the evaluated set.
	Entries []Entry `json:"entries"`

	// Clean designates the vocabulary's clean value. A nil pointer is an absent
	// designation — distinct from a present-but-null one.
	Clean *Designation `json:"clean"`
}

// Errors returned by Partition when the consumed result violates the contract.
// They are exactly the negations of the contract's clauses; widening the
// contract obliges widening this set.
var (
	// ErrDesignationAbsent — the result carries no clean-value designation.
	ErrDesignationAbsent = errors.New("guard liveness: evaluation result carries no clean-value designation")

	// ErrDesignationNull — a designation is present but designates nothing.
	ErrDesignationNull = errors.New("guard liveness: clean-value designation is null")

	// ErrDesignationMultiValued — the designation names more than one value, so
	// there is no single clean value to partition against.
	ErrDesignationMultiValued = errors.New("guard liveness: clean-value designation names more than one value")

	// ErrEntryClassificationCount — an entry carries a number of
	// classifications other than exactly one.
	ErrEntryClassificationCount = errors.New("guard liveness: entry does not carry exactly one classification")
)

// Partition returns the entries whose classification is not the clean value.
//
// The condition is the clean/non-clean partition and nothing else: it is never
// a list of particular classifications, and it is never the producer's surface
// fold. The clean value is read from the result's own carried designation, which
// is the only route that stays correct under a vocabulary change.
//
// A result that violates the contract has no partition to compute. Partition
// returns an error and a nil set rather than an empty non-clean set, because an
// empty set is indistinguishable from an all-clear and would report green about
// a set this package never managed to read.
func (r Result) Partition() ([]Entry, error) {
	clean, err := r.cleanValue()
	if err != nil {
		return nil, err
	}

	nonClean := make([]Entry, 0, len(r.Entries))
	for _, e := range r.Entries {
		if len(e.Classifications) != 1 {
			return nil, fmt.Errorf("%w: subject %q carries %d", ErrEntryClassificationCount, e.Subject, len(e.Classifications))
		}
		if e.Classifications[0] != clean {
			nonClean = append(nonClean, e)
		}
	}
	return nonClean, nil
}

// cleanValue reads the carried designation, refusing every shape that leaves
// the partition without a referent.
func (r Result) cleanValue() (string, error) {
	if r.Clean == nil {
		return "", ErrDesignationAbsent
	}
	switch len(r.Clean.Values) {
	case 0:
		return "", ErrDesignationNull
	case 1:
		return r.Clean.Values[0], nil
	default:
		return "", fmt.Errorf("%w: %d values designated", ErrDesignationMultiValued, len(r.Clean.Values))
	}
}

// Render returns the advisory text for a result, naming exactly the entries the
// partition fired on. An all-clean result renders as the empty string.
//
// M1 renders the fired set and nothing more. Arrival at the host surface, the
// measurement's age, and change-leading are M2 and M3 (REQ-GDL-005, 006, 007).
func Render(r Result) (string, error) {
	nonClean, err := r.Partition()
	if err != nil {
		return "", err
	}
	if len(nonClean) == 0 {
		return "", nil
	}

	var b strings.Builder
	b.WriteString("moai guard liveness — ")
	fmt.Fprintf(&b, "%d subject(s) not reporting clean:", len(nonClean))
	for _, e := range nonClean {
		fmt.Fprintf(&b, "\n  - %s", e.Subject)
	}
	return b.String(), nil
}
