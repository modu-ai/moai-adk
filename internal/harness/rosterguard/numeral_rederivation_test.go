package rosterguard

import (
	"fmt"
	"testing"
)

// TestNumeralResidualArithmeticCloses re-derives the layer's population and
// residual IN-RUN, against the tree being changed, and asserts the arithmetic
// closes with nothing left implicit (AC-RNA-012, AC-RNA-013).
//
// # What "independent" means here, and what it does not
//
// The registry side is read from Registry() and NumeralExemptions() AT RUNTIME,
// never parsed out of registry.go's text. That distinction is load-bearing and
// is why this test exists in this shape: a text-level Path:/Claims: parse cannot
// see the four rows readmeSite() builds from a path PARAMETER, and two
// measurements of this tree disagreed 16-vs-20 for exactly that reason before
// the run began.
//
// The tree side calls ScanNumeralAxis — it reads the tree directly rather than
// consulting any total the layer reports about itself, so no figure below is
// taken on the layer's word. It is NOT a second, independently-written scanner:
// a duplicate noun class and adjacency window would drift from the real one and
// would then measure a population nobody ships. The residual risk that buys is
// stated plainly: a defect in the noun class would move both this arithmetic
// and the layer together.
func TestNumeralResidualArithmeticCloses(t *testing.T) {
	hits := scanLive(t)

	breadth := map[string]bool{}
	for _, h := range hits {
		breadth[h.Path] = true
	}

	countRows, countPaths := 0, map[string]bool{}
	unreachable := map[string]bool{}
	for _, s := range Registry() {
		if !s.Claims.Has(ClaimCount) {
			continue
		}
		countRows++
		countPaths[s.Path] = true
		if s.NumeralUnreachable != "" {
			unreachable[s.Path] = true
		}
	}

	exemptPaths := map[string]bool{}
	for _, e := range NumeralExemptions() {
		exemptPaths[e.Path] = true
	}

	discharged, exempted, residual := 0, 0, 0
	for p := range breadth {
		switch {
		case countPaths[p]:
			discharged++
		case exemptPaths[p]:
			exempted++
		default:
			residual++
		}
	}

	// Printed at column 0 for the same reason the breadth set is: a figure that
	// only exists inside an assertion cannot be read back as evidence.
	fmt.Println("numeral re-derivation (in-run, this tree)")
	fmt.Printf("  breadth-set hits                : %d\n", len(hits))
	fmt.Printf("  breadth-set paths               : %d\n", len(breadth))
	fmt.Printf("  registry rows carrying Count    : %d\n", countRows)
	fmt.Printf("  registry paths carrying Count   : %d\n", len(countPaths))
	fmt.Printf("  ...declared NumeralUnreachable  : %d\n", len(unreachable))
	fmt.Printf("  hits discharged by a Count row  : %d\n", discharged)
	fmt.Printf("  hits discharged by an exemption : %d\n", exempted)
	fmt.Printf("  residual (undeclared)           : %d\n", residual)
	fmt.Printf("  numeral exemptions declared     : %d\n", len(NumeralExemptions()))

	if len(breadth) == 0 {
		t.Fatal("an empty breadth set is a measurement failure, not a clean tree")
	}
	if discharged+exempted+residual != len(breadth) {
		t.Errorf("the arithmetic does not close: %d discharged + %d exempted + %d residual != %d breadth paths — some hit is unaccounted for",
			discharged, exempted, residual, len(breadth))
	}
	if residual != 0 {
		t.Errorf("%d breadth-set path(s) are neither registered nor exempted; the finding set must be empty", residual)
	}
	if len(countPaths)-len(unreachable) != discharged {
		t.Errorf("%d reachable Count paths but %d discharged hits — a registered count row is not being reached; re-measure it or declare NumeralUnreachable",
			len(countPaths)-len(unreachable), discharged)
	}
}
