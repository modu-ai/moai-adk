package rosterguard

import (
	"strings"
	"testing"
)

// agentSourceDir is the agent source directory SPEC-AGENT-CORE4-001 introduces.
// The sources are verbatim copies of the shipped definitions, so the one copy
// that carries a roster count claim (manager-design) needs the same registry
// rows its C1, C2 and C3 siblings have.
const agentSourceDir = "internal/template/agentemit/source/"

// TestNumeralSourceDirDischarged is AC-AC4-010: the numeral layer reaches the
// agent source directory and reports no undischarged hit there.
//
// The hit is asserted PRESENT before it is asserted discharged. Without that, a
// source directory the walk never reaches (or a source set that lost its
// manager-design copy) would satisfy "no undischarged hit" by finding nothing.
func TestNumeralSourceDirDischarged(t *testing.T) {
	hits := scanLive(t)

	var underSource []NumeralHit
	for _, h := range hits {
		if strings.HasPrefix(h.Path, agentSourceDir) {
			underSource = append(underSource, h)
		}
	}

	t.Run("source_manager_design_discharged", func(t *testing.T) {
		const want = agentSourceDir + "manager-design.md"
		found := false
		for _, h := range underSource {
			if h.Path == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("the numeral layer reports no hit in %s: the roster count claim of the copied definition is not being measured", want)
		}
		counted := false
		for _, s := range Registry() {
			if s.Path == want && s.Claims.Has(ClaimCount) {
				counted = true
			}
		}
		if !counted {
			t.Errorf("%s carries a roster count claim but the registry has no ClaimCount row for it", want)
		}
	})

	t.Run("no_undischarged_hit_under_source_dir", func(t *testing.T) {
		if len(underSource) == 0 {
			t.Fatal("empty breadth set under the agent source directory: the guard measured nothing")
		}
		for _, v := range DischargeNumeralHits(underSource, Registry(), NumeralExemptions()) {
			t.Error(v)
		}
	})
}
