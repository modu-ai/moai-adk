package spec

import (
	"strings"
	"testing"
)

// TestStatusTokenUnrecognizedDoesNotDuplicateStatusValueInvalid pins the
// REQ-STV-015 property on a document that lands in BOTH sets.
//
// Why a constructed fixture rather than the corpus measurement: the corpus
// intersection is 0 only because `StatusValueInvalid` has population 0 on this
// corpus, so the measurement reads the same against a rule that overlapped
// completely. The two rules are NOT structurally disjoint — a SPEC whose
// frontmatter carries an invalid status AND whose history carries an
// unrecognized token lands in both — so nothing but this fixture stops the
// guarantee from lapsing silently the first time such a document appears.
//
// What REQ-STV-015 actually prohibits is DUPLICATION: the same fact reported
// twice. It does not prohibit a document appearing in both sets, and the
// assertions below are what separate those two readings — on an overlapping
// document each finding must name its own subject, `StatusValueInvalid` the
// frontmatter's current value and `StatusTokenUnrecognized` a token seen only
// in git history.
//
// AC-STV-016's corpus-level clause is untouched by this test: the corpus
// intersection remains 0, and a corpus document that ever lands in both still
// fails that AC until its remedy path runs.
func TestStatusTokenUnrecognizedDoesNotDuplicateStatusValueInvalid(t *testing.T) {
	// "approved" appears ONLY in git history; "Completed" is the current
	// frontmatter value. That asymmetry is the discriminator the assertions
	// below use — a rule that reported the other one's subject would fail.
	const historyOnlyToken = "approved"
	const frontmatterToken = "Completed"

	b := buildTransitionFixture(t, transitionFixture{
		from:              historyOnlyToken,
		to:                frontmatterToken,
		trailer:           "manager-docs",
		outOfScopeHeading: true,
	})
	report := lintFixture(t, b, false)

	valueInvalid := findingsWithCode(report, "StatusValueInvalid")
	tokenUnrecognized := findingsWithCode(report, "StatusTokenUnrecognized")

	if len(valueInvalid) != 1 {
		t.Fatalf("StatusValueInvalid findings = %d, want 1 (the fixture's frontmatter status is outside the enum)", len(valueInvalid))
	}
	if len(tokenUnrecognized) != 1 {
		t.Fatalf("StatusTokenUnrecognized findings = %d, want 1 (the fixture's history names an unrecognized token)", len(tokenUnrecognized))
	}

	// The overlap is real. What must hold is that it is not duplication.
	vm := valueInvalid[0].Message
	tm := tokenUnrecognized[0].Message

	if !strings.Contains(vm, frontmatterToken) {
		t.Errorf("StatusValueInvalid message does not name the frontmatter value %q: %s", frontmatterToken, vm)
	}
	if strings.Contains(vm, historyOnlyToken) {
		t.Errorf("StatusValueInvalid message names the history-only token %q — it is judging git history, which is not its subject: %s", historyOnlyToken, vm)
	}
	if !strings.Contains(tm, historyOnlyToken) {
		t.Errorf("StatusTokenUnrecognized message does not name the history-only token %q — it is not reporting the fact that distinguishes it: %s", historyOnlyToken, tm)
	}

	// And the pair check must not also fire: an unrecognized token stops at its
	// own code (AC-STV-015's paired assertion, held here on the overlap case).
	if got := len(findingsWithCode(report, "StatusTransitionInvalid")); got != 0 {
		t.Errorf("StatusTransitionInvalid findings = %d, want 0 (the token check short-circuits the pair check)", got)
	}
}
