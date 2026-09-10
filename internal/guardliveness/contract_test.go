package guardliveness

import (
	"errors"
	"strings"
	"testing"
)

// The fixtures below invent their own classification vocabularies on purpose.
// SPEC-GUARD-LIVENESS-001 §B.1 resolves the seam to SPEC-GUARD-STATE-MODEL-001
// as a consumed contract rather than a dependency, so nothing in this package
// — tests included — may depend on the real vocabulary's size or value names.
// Two fixtures differing in BOTH is what makes that falsifiable.

func designate(values ...string) *Designation { return &Designation{Values: values} }

func entry(subject, classification, surface string) Entry {
	return Entry{Subject: subject, Classifications: []string{classification}, Surface: surface}
}

// resultA carries a three-value vocabulary whose clean value is "alpha".
// "beta" is non-clean yet folds to the SAME producer surface value as the clean
// entry, which is what makes AC-GDL-001 mutant (c) — partitioning by the
// surface fold — observable rather than merely forbidden in prose.
func resultA() Result {
	return Result{
		Clean: designate("alpha"),
		Entries: []Entry{
			entry("subject-1", "alpha", "settled"),
			entry("subject-2", "beta", "settled"),
			entry("subject-3", "gamma", "unsettled"),
		},
	}
}

// resultB carries a two-value vocabulary sharing no value name with resultA,
// and likewise contains a non-clean entry folding to the clean surface value.
func resultB() Result {
	return Result{
		Clean: designate("quiet"),
		Entries: []Entry{
			entry("subject-9", "quiet", "hushed"),
			entry("subject-8", "loud", "hushed"),
		},
	}
}

func subjectsOf(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Subject)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// AC-GDL-001 (a)+(c) — the partition comes from the carried clean-value
// designation, so it is correct under two unrelated vocabularies AND fires on
// an entry that folds to the clean surface value.
func TestPartitionFiresOnExactlyTheNonCleanEntries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result Result
		want   []string
	}{
		{"three-value vocabulary", resultA(), []string{"subject-2", "subject-3"}},
		{"two-value vocabulary, no shared value name", resultB(), []string{"subject-8"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nonClean, err := tc.result.Partition()
			if err != nil {
				t.Fatalf("Partition: %v", err)
			}
			got := subjectsOf(nonClean)
			if !equalStrings(got, tc.want) {
				t.Fatalf("fired on %v, want %v — an entry folding to the clean surface value is still non-clean", got, tc.want)
			}
		})
	}
}

// AC-GDL-001 (b) is the value-name grep specified in acceptance.md §D.2 and is
// run against this package's non-test source, not from inside a test. What IS
// checkable here is the other half of the same property: the partition must not
// consult the producer's folded surface value at all. Rewriting every surface
// value to one constant — which destroys all fold information — must not change
// the outcome.
func TestPartitionDoesNotConsultTheSurfaceFold(t *testing.T) {
	flattened := resultA()
	for i := range flattened.Entries {
		flattened.Entries[i].Surface = "collapsed"
	}
	nonClean, err := flattened.Partition()
	if err != nil {
		t.Fatalf("Partition: %v", err)
	}
	got := subjectsOf(nonClean)
	want := []string{"subject-2", "subject-3"}
	if !equalStrings(got, want) {
		t.Fatalf("fired on %v with every surface value collapsed, want %v — the partition read the surface fold", got, want)
	}
}

// AC-GDL-001 (a) — rendering names exactly the entries fired on, and says
// nothing when every entry is clean.
func TestRenderNamesExactlyTheFiredEntries(t *testing.T) {
	text, err := Render(resultA())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	for _, want := range []string{"subject-2", "subject-3"} {
		if !strings.Contains(text, want) {
			t.Errorf("render omits non-clean subject %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "subject-1") {
		t.Errorf("render names the clean subject subject-1:\n%s", text)
	}

	allClean := Result{Clean: designate("alpha"), Entries: []Entry{entry("subject-1", "alpha", "settled")}}
	quiet, err := Render(allClean)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if quiet != "" {
		t.Errorf("render spoke on an all-clean result: %q", quiet)
	}
}

// A result violating the consumed contract has no clean/non-clean partition to
// compute, and the path of least resistance is to identify nothing as non-clean
// and therefore report nothing — green about a set the mechanism never read
// (spec.md §A.0, at this consumer's layer). Partition refuses that by returning
// an error instead of an empty non-clean set.
//
// The full criterion for this shape is AC-GDL-013 (five fixtures, plus an
// advisory that NAMES the violation) and belongs to M2. What is asserted here is
// only what M1's own partition must not do.
func TestPartitionRefusesAContractViolatingResult(t *testing.T) {
	entries := []Entry{entry("subject-1", "alpha", "settled")}
	for _, tc := range []struct {
		name   string
		result Result
		want   error
	}{
		{"designation absent", Result{Clean: nil, Entries: entries}, ErrDesignationAbsent},
		{"designation null", Result{Clean: &Designation{}, Entries: entries}, ErrDesignationNull},
		{"designation multi-valued", Result{Clean: designate("alpha", "quiet"), Entries: entries}, ErrDesignationMultiValued},
		{"entry carries no classification", Result{
			Clean:   designate("alpha"),
			Entries: []Entry{{Subject: "subject-1", Surface: "settled"}},
		}, ErrEntryClassificationCount},
		{"entry carries two classifications", Result{
			Clean:   designate("alpha"),
			Entries: []Entry{{Subject: "subject-1", Classifications: []string{"alpha", "beta"}, Surface: "settled"}},
		}, ErrEntryClassificationCount},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nonClean, err := tc.result.Partition()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Partition error = %v, want %v", err, tc.want)
			}
			if nonClean != nil {
				t.Fatalf("Partition returned a partition (%v) for a result it could not read", subjectsOf(nonClean))
			}
		})
	}
}
