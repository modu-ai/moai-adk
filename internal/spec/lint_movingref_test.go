package spec_test

// lint_movingref_test.go — MovingRefUnpinnedRule acceptance suite
// (SPEC-MOVING-REF-GUARD-001 M3).
//
// Every test below names, in its doc comment, the MUTATION that must turn it
// red. Per acceptance.md §A this is not decoration: the deliverable is a guard,
// and a guard whose criterion cannot fail is indistinguishable from a guard that
// is switched off. Each mutation was planted, observed red, and reverted; the
// verbatim failing output is recorded in progress.md §E.2.
//
// Fixtures live under testdata/movingref/<case>/ and each is a schema-valid SPEC
// directory carrying `era: V3R6` in frontmatter. The era pin is load-bearing, not
// cosmetic (acceptance.md §A [HARD] fixture era precondition): a fixture with no
// progress.md classifies V2.x under heuristic H-1 and every warning on it is
// era-demoted to Advisory, which silently breaks the --strict half of AC-MRG-005
// for a reason having nothing to do with the rule under test.
//
// Fixtures are linted ONE AT A TIME (an explicit spec.md path per call), so they
// deliberately share a SPEC id; DuplicateSPECIDRule never sees two of them in the
// same run.
//
// NOT covered here: AC-MRG-013 and its CM-1 / CM-2 counter-mutations. REQ-MRG-010
// (the R4-form exclusion) is DEFERRED out of M3 by operator decision (spec.md §H
// Q0, option C) — spec.md §B.7 measured its reachable class as 0 of 42 candidate
// lines on two independent probes, so an exclusion for it could only over-exempt.
// No R4 branch exists in the rule, so there is nothing for those criteria to test.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// movingRefFixture returns the spec.md path of a fixture case.
func movingRefFixture(c string) string {
	return filepath.Join(testdataDir, "movingref", c, "spec.md")
}

// lintMovingRef lints one fixture case and returns the full report.
func lintMovingRef(t *testing.T, c string, strict bool) *spec.Report {
	t.Helper()
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      filepath.Join(testdataDir, "movingref", c),
		Strict:       strict,
	})
	report, err := linter.Lint([]string{movingRefFixture(c)})
	if err != nil {
		t.Fatalf("Lint(%s) returned unexpected error: %v", c, err)
	}
	return report
}

// movingRefFindings returns only the MovingRefUnpinned findings for a case.
func movingRefFindings(t *testing.T, c string) []spec.Finding {
	t.Helper()
	return findingsForCode(lintMovingRef(t, c, false).Findings, "MovingRefUnpinned")
}

// fixtureLine reads line n (1-indexed) of a fixture artifact, so a criterion can
// assert WHICH line was flagged without hardcoding a line number that any edit to
// the fixture's prose would invalidate.
func fixtureLine(t *testing.T, path string, n int) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(data), "\n")
	if n < 1 || n > len(lines) {
		t.Fatalf("line %d out of range in %s (%d lines)", n, path, len(lines))
	}
	return lines[n-1]
}

// TestMovingRef_FiresOnUnpinnedAnchor decides AC-MRG-001: the detector fires on
// the true-positive shape — a moving ref in a git-command context deciding an
// invariant claim, with no SHA pin and no frozen-baseline variable.
//
// Mutation that must turn it red: change the fixture's `origin/main` to
// `origin/mainx`; the count must drop to 0, proving the moving-ref token drives
// the finding rather than some incidental substring of the row.
func TestMovingRef_FiresOnUnpinnedAnchor(t *testing.T) {
	got := movingRefFindings(t, "unpinned-anchor")
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 MovingRefUnpinned finding, got %d: %v", len(got), got)
	}
	f := got[0]
	// The fixture row lives in spec.md — i.e. inside SPECDoc.Body — deliberately.
	// AC-MRG-009 asserts that its body-only mutant goes red "while AC-MRG-001
	// still passes"; a row placed in a sibling artifact takes both criteria down
	// together and that separation becomes unobservable.
	if filepath.Base(f.File) != "spec.md" {
		t.Errorf("expected the finding against spec.md, got %s", f.File)
	}
	line := fixtureLine(t, f.File, f.Line)
	if !strings.Contains(line, "git diff --name-only origin/main") {
		t.Errorf("finding points at %s:%d, whose content is not the flagged row: %q", f.File, f.Line, line)
	}
}

// TestMovingRef_PinnedClaimNotFlagged decides AC-MRG-002 (REQ-MRG-008): a claim
// carrying its anchor's resolved SHA on the line is not flagged.
//
// The fixture RETAINS the moving-ref token and records the resolved SHA beside
// it — the shape filter 4 of spec.md §B.3 removes. A fixture that deleted the
// token instead would fail conjuncts 1+2 rather than being exempted, and the
// mutation below could not turn it red (the criterion would be vacuous).
//
// Mutation that must turn it red: delete the hex-SHA exclusion branch from the
// rule; the finding reappears.
func TestMovingRef_PinnedClaimNotFlagged(t *testing.T) {
	if got := movingRefFindings(t, "pinned-sha"); len(got) != 0 {
		t.Errorf("expected 0 findings on a SHA-pinned claim, got %d: %v", len(got), got)
	}
}

// TestMovingRef_MarkerSuppressesOnlyWithReason decides AC-MRG-003 (REQ-MRG-002 /
// REQ-MRG-003): (a) no marker fires, (b) a marker with a reason suppresses, and
// (c) a bare marker does NOT suppress — it reports the marker as incomplete.
//
// Case (c) is the incentive guard: were a reason-less marker honoured, silencing
// would be cheaper than fixing and the whole exemption would invert.
//
// Mutation that must turn it red: remove the non-empty-reason check; fixture (c)
// then reports zero and the criterion fails.
func TestMovingRef_MarkerSuppressesOnlyWithReason(t *testing.T) {
	if got := movingRefFindings(t, "unpinned-anchor"); len(got) != 1 {
		t.Errorf("(a) no marker: expected 1 finding, got %d: %v", len(got), got)
	}
	if got := movingRefFindings(t, "marker-reason"); len(got) != 0 {
		t.Errorf("(b) marker with reason: expected 0 findings, got %d: %v", len(got), got)
	}
	got := movingRefFindings(t, "marker-empty")
	if len(got) != 1 {
		t.Fatalf("(c) bare marker: expected 1 finding, got %d: %v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "incomplete") {
		t.Errorf("(c) bare marker: finding must name the marker as incomplete, got: %s", got[0].Message)
	}
}

// TestMovingRef_ThreeDotNotExempt decides AC-MRG-004 (REQ-MRG-007): `A...B` is
// treated exactly as the two-dot form.
//
// Merge-base is NOT stable under upstream advance — spec.md §B.2 measured the
// identical wrong answer from both forms in this tree — so exempting `...` is a
// tempting mitigation that does not work.
//
// Mutation that must turn it red: add a `strings.Contains(line, "...")`
// early-return to the rule.
func TestMovingRef_ThreeDotNotExempt(t *testing.T) {
	if got := movingRefFindings(t, "three-dot"); len(got) != 1 {
		t.Errorf("expected 1 finding on the three-dot form, got %d: %v", len(got), got)
	}
}

// TestMovingRef_SeverityWarningAndExitCode decides AC-MRG-005 (REQ-MRG-009 and
// spec.md §D.5): the finding is a non-advisory `warning`, so `moai spec lint`
// exits 0 while `--strict` exits non-zero.
//
// Severity above `warning` would red the 42 existing corpus candidates on the
// first run, and the rational response to that is a bulk suppression — the exact
// outcome this SPEC exists to prevent.
//
// Mutation that must turn it red: emit at SeverityError; the non-strict exit
// becomes non-zero.
func TestMovingRef_SeverityWarningAndExitCode(t *testing.T) {
	report := lintMovingRef(t, "unpinned-anchor", false)

	// The criterion is stated over a corpus "whose only findings are
	// MovingRefUnpinned". Info-severity findings do not enter HasErrors, so the
	// binding check is that no OTHER error- or warning-severity finding exists.
	for _, f := range report.Findings {
		if f.Code != "MovingRefUnpinned" && (f.Severity == spec.SeverityError || f.Severity == spec.SeverityWarning) {
			t.Fatalf("fixture is not clean: unexpected %s finding %s: %s", f.Severity, f.Code, f.Message)
		}
	}

	got := findingsForCode(report.Findings, "MovingRefUnpinned")
	if len(got) != 1 {
		t.Fatalf("expected 1 MovingRefUnpinned finding, got %d", len(got))
	}
	if got[0].Severity != spec.SeverityWarning {
		t.Errorf("expected severity %q, got %q", spec.SeverityWarning, got[0].Severity)
	}
	if got[0].Advisory {
		t.Error("finding is Advisory — --strict would not escalate it; check the fixture's era classification")
	}
	if report.HasErrors() {
		t.Error("non-strict lint must exit 0 when the only findings are MovingRefUnpinned warnings")
	}
	if strictReport := lintMovingRef(t, "unpinned-anchor", true); !strictReport.HasErrors() {
		t.Error("--strict lint must exit non-zero on a MovingRefUnpinned warning")
	}
}

// TestMovingRef_DivergenceFigureVariant decides AC-MRG-006 (REQ-MRG-006): a line
// citing a `rev-list --count --left-right` figure against a moving ref, with no
// SHA and no date, is the same defect on a different carrier — a measurement
// whose validity expired being re-served as current.
//
// SHOULD-tier (spec.md §H Q2). KEPT at M3: it did not over-fire on any fixture,
// and AC-MRG-014's negative control demonstrates it does not fire on a line that
// merely names the command.
//
// Mutation that must turn it red: append a resolved SHA to the fixture line; the
// finding must disappear, proving the rule keys on the missing pin rather than on
// the `rev-list` verb alone.
func TestMovingRef_DivergenceFigureVariant(t *testing.T) {
	got := movingRefFindings(t, "divergence")
	if len(got) != 1 {
		t.Fatalf("expected 1 finding on the divergence-figure form, got %d: %v", len(got), got)
	}
	if filepath.Base(got[0].File) != "progress.md" {
		t.Errorf("expected the finding against progress.md, got %s", got[0].File)
	}
}

// TestMovingRef_MessageNamesAllFourBranches decides AC-MRG-008 (REQ-MRG-004): the
// message names all four remediation branches of spec.md §D.2 and does not
// present pinning as the sole remedy.
//
// This is the mechanism by which the dominant failure mode reaches a reader: most
// people act on the message and never open the doctrine. With one remedy on offer
// the reader pins everything, destroying the subject-class claims that are correct
// as written.
//
// Mutation that must turn it red: shorten the message to name only pinning.
// Second mutation: drop ONLY R4, leaving three branches — the assertion must
// STILL fail, because R4 is the branch a reader cannot reach by intuition and the
// one the lead's own dispatch failure demonstrates is needed (spec.md §B.5).
func TestMovingRef_MessageNamesAllFourBranches(t *testing.T) {
	got := movingRefFindings(t, "unpinned-anchor")
	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(got))
	}
	msg := got[0].Message

	// One assertion per branch, so dropping any single branch fails on its own row.
	branches := map[string]string{
		"R1 (pin)":                   "pin the resolved SHA",
		"R2 (freeze at pre-flight)":  "freeze at pre-flight",
		"R3 (declare the exemption)": "declare the exemption",
		"R4 (state the command)":     "state the measuring command",
	}
	for name, want := range branches {
		if !strings.Contains(msg, want) {
			t.Errorf("message omits %s: expected substring %q in: %s", name, want, msg)
		}
	}
	// Pinning must not read as the sole or default remedy.
	if !strings.Contains(msg, "Pinning is one branch of four, not the default") {
		t.Errorf("message must not present pinning as the sole remedy; got: %s", msg)
	}
	// Limit L2: the finding is a question put to a human, never a verdict.
	if !strings.Contains(msg, "question, not a verdict") {
		t.Errorf("message must state that it is a question rather than a verdict; got: %s", msg)
	}
}

// TestMovingRef_ReadsSiblingArtifacts decides AC-MRG-009: the rule reads the
// SPEC's sibling artifacts, not just `SPECDoc.Body`.
//
// `Body` carries spec.md alone, while most real occurrences live in progress.md
// and acceptance.md (spec.md §B.3) — a body-only rule would miss the majority of
// the corpus while appearing to work.
//
// Mutation that must turn it red: restrict the rule to doc.Body. This criterion
// fails while AC-MRG-001 still passes, which is exactly why it is separate.
func TestMovingRef_ReadsSiblingArtifacts(t *testing.T) {
	got := movingRefFindings(t, "sibling-progress")
	if len(got) != 1 {
		t.Fatalf("expected 1 finding from the sibling progress.md, got %d: %v", len(got), got)
	}
	if filepath.Base(got[0].File) != "progress.md" {
		t.Errorf("expected the finding against progress.md, got %s", got[0].File)
	}
}

// TestMovingRef_NegativeControlOnClaimConjunct decides AC-MRG-014: a moving ref
// in a git-command context carrying NO invariant-claim marker is not a finding.
//
// This is the criterion the Definition of Done singles out. The mutant it catches
// — delete the claim-marker conjunct — is the SIMPLER rule, and it passes
// AC-MRG-001, AC-MRG-001's own mutation, and -002, -004, -005, -006, -008 and
// -009 unchanged. Measured over `.moai/specs/**` excluding this SPEC's own
// directory it yields 495 findings against the 42 that spec.md §D.5 sizes the
// `warning` severity on — a ~12x over-fire the whole suite would otherwise pass
// green, and one that defeats §D.5 directly, since bulk suppression is the
// rational response to 495 warnings.
//
// Mutation that must turn it red: delete the claim-marker conjunct from the rule.
func TestMovingRef_NegativeControlOnClaimConjunct(t *testing.T) {
	if got := movingRefFindings(t, "no-claim"); len(got) != 0 {
		t.Errorf("expected 0 findings on a claim-free instruction, got %d: %v", len(got), got)
	}
}
