package factorylane

import (
	"path/filepath"
	"strings"
	"testing"
)

// The SPEC-less triple reads its verdict through the shared selector: a card
// that wrote only a sync-audit file (no verdict.md) is accepted, and
// verdict.md still wins when both exist.
func TestMergeTripleSpecLessUsesSharedSelector(t *testing.T) {
	r, head := newSpecLessRepo(t)
	r.write(t, filepath.Join(".moai", "reports", "t9", "sync-audit-iter2.md"), verdictBody("PASS", head))
	r.commit(t, "sync-audit only")
	if run := evalSpecLess(t, r, ""); !run.Checks[0].Passed {
		t.Fatalf("sync-audit-only card refused: %s", run.Checks[0].Detail)
	}

	r.write(t, r.verdictRel(), verdictBody("FAIL", head))
	r.commit(t, "verdict.md FAIL")
	run := evalSpecLess(t, r, "")
	if run.Checks[0].Passed || !strings.Contains(run.Checks[0].Detail, ".moai/reports/t9/verdict.md") {
		t.Fatalf("verdict.md must win over sync-audit files: %+v", run.Checks[0])
	}
}
