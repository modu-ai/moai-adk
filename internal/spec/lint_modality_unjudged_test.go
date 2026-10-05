package spec

import (
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-SPEC-LINT-BLIND-AXES-001 axis 2, BRANCH B (card t518, milestone M-A2).
//
// THE DEFECT. isModalityMalformed reads five ENGLISH prefixes and returns false
// for everything else. For a non-English requirement that false is
// indistinguishable from "well formed" — the tool cannot judge the text and
// says nothing, and saying nothing is exactly what it says when the text is
// fine. These tests are the RED half of announcing the difference.
//
// SCOPE. Branch A — actually judging Korean modality by accepting `해야 한다`
// / `해서는 안 된다` as SHALL equivalents — is OUT of scope by operator
// decision (spec.md §E, §I). Nothing here builds a Korean lexicon. What is in
// scope is the DISCRIMINATOR that decides what counts as unjudgeable, and the
// SHALL word-boundary repair (REQ-SLB-014) is part of that discriminator.
//
// Two disciplines from acceptance.md §A bind every test here:
//
//   - CONTROL PAIRS. English and non-English fixtures are compared AGAINST EACH
//     OTHER. A one-sided fixture cannot distinguish "both judged" from "both
//     ignored" — the exact shape in which this tool went silent.
//   - NO EXIT CODES. Assertions are on match counts and named finding codes.
//
// Mutant results (including the mutants NOT caught) are recorded in
// .moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/progress.md §E.2.

// --- fixtures -------------------------------------------------------------

// AC-SLB-006a control — ENGLISH, SHALL missing. Written in the NARROW list
// shape (`- REQ-XXX-NNN-NNN: text`) so the entry is NOT Widened: any advisory
// marking observed on it must then come from the emission site under test, not
// from the t385 widening path.
const modalityEnglishMalformedBody = `## Requirements

- REQ-FXE-001-001: When the tool runs, the system does report.
`

// AC-SLB-006b control — KOREAN, same defect (no SHALL-equivalent marker),
// same narrow list shape so Widened is false here too.
const modalityKoreanUnjudgedBody = `## Requirements

- REQ-FXK-001-001: 이 항목의 처리는 구현 재량에 맡긴다.
`

// AC-SLB-007 fixtures — VERBATIM from acceptance.md §C. Fixed there, before the
// implementation existed, precisely so the implementation cannot pick sentences
// that pass harmlessly.
const modalityConformingVerbatimBody = `## Requirements

- **REQ-FX-001** — 도구는 판정 결과를 출력에 남겨야 한다(SHALL).
`

const modalityUnjudgeableVerbatimBody = `## Requirements

- **REQ-FX-002** — 이 항목의 처리는 구현 재량에 맡긴다.
`

// AC-SLB-012 fixtures — VERBATIM from acceptance.md §C.
//
// [HARD] The verbatim REQ-FX-011 line CANNOT make the mutant guard, and that is
// measured rather than argued: its text already contains " SHALL" with a
// leading space (from "the system shall report"), so the OLD contact condition
// judges it well-formed too. Reverting the repair leaves it at zero findings.
// TestModality_ShallContactIsWordBoundary asserts that property on the fixture
// itself, so the record is mechanical and not a claim about it.
//
// It is kept — a verbatim fixture is evidence about the SPEC — and a second,
// CORRECTED fixture is added below whose only SHALL occurrence is parenthesized.
// That one is what makes the guard.
const modalityShallNoLeadingSpaceVerbatimBody = `## Requirements

- **REQ-FX-011** — When the tool runs, the system shall report.(SHALL)
`

const modalityShallLeadingSpaceControlBody = `## Requirements

- **REQ-FX-012** — When the tool runs, the system SHALL report.
`

// The corrected discriminating fixture: an English WHEN requirement whose ONLY
// SHALL occurrence is parenthesized, so the old leading-space condition misses
// it and the word-boundary condition sees it.
const modalityShallOnlyParenthesizedBody = `## Requirements

- **REQ-FXB-011** — When the tool runs, the tool reports.(SHALL)
`

// modalityMixedAdvisoryBody carries one table-form requirement (advisory via the
// existing Widened path) and one unjudgeable Korean list requirement (advisory
// via the new code's own emission site). AC-SLB-010 reads both halves at once.
const modalityMixedAdvisoryBody = `## Requirements

| 요구 | modality | 내용 |
|---|---|---|
| REQ-FXM-001 | Ubiquitous | 도구는 표 유래 항목을 자문으로 처리해야 한다 |

- REQ-FXM-002-001: 이 항목의 처리는 구현 재량에 맡긴다.
`

// --- helpers --------------------------------------------------------------

// modalityDoc parses a fixture written under t.TempDir() and returns the doc
// with the LIVE (provenance-carrying) collector applied.
//
// It fails the test when the fixture collected zero REQ entries: a fixture that
// matches nothing produces an empty finding set, and an empty finding set
// satisfies every absence assertion below for the wrong reason.
func modalityDoc(t *testing.T, id, body string) *SPECDoc {
	t.Helper()
	p := writeTableSpecFixture(t, id, body)
	doc := parseSPECDoc(p)
	if doc.ParseError != nil {
		t.Fatalf("fixture %s parse: %v", id, doc.ParseError)
	}
	doc.REQs = parseREQsWithProvenance(doc.Body)
	if len(doc.REQs) == 0 {
		t.Fatalf("fixture %s collected zero REQ entries — a control that matches nothing is not a control", id)
	}
	return doc
}

func modalityFindings(t *testing.T, id, body string) []Finding {
	t.Helper()
	return (&EARSModalityRule{}).Check(modalityDoc(t, id, body), nil)
}

func countCode(findings []Finding, code string) int {
	n := 0
	for _, f := range findings {
		if f.Code == code {
			n++
		}
	}
	return n
}

func codeSet(findings []Finding) map[string]int {
	m := map[string]int{}
	for _, f := range findings {
		m[f.Code]++
	}
	return m
}

// oldShallContact replicates the PRE-REPAIR contact condition verbatim
// (`strings.Contains(upper, " SHALL")`). It exists so a fixture's ability to
// make the word-boundary guard is asserted mechanically instead of asserted in
// prose. AC-SLB-012.
func oldShallContact(text string) bool {
	return strings.Contains(strings.ToUpper(text), " SHALL")
}

// --- AC-SLB-006a — control, English ---------------------------------------

// TestModality_EnglishControlStillJudged is the no-regression half of the axis-2
// control pair: an English requirement missing SHALL is still judged malformed,
// and the new unjudged code does NOT fire on it.
func TestModality_EnglishControlStillJudged(t *testing.T) {
	findings := modalityFindings(t, "FXE-001", modalityEnglishMalformedBody)

	if got := countCode(findings, "ModalityMalformed"); got != 1 {
		t.Errorf("English SHALL-less requirement: ModalityMalformed = %d, want 1", got)
	}
	if got := countCode(findings, "ModalityUnjudged"); got != 0 {
		t.Errorf("English SHALL-less requirement: ModalityUnjudged = %d, want 0 — it IS judged", got)
	}
}

// --- AC-SLB-006b — control, Korean ----------------------------------------

// TestModality_KoreanControlAnnounced asserts that a Korean requirement with the
// same defect produces EXACTLY ONE of {ModalityMalformed, ModalityUnjudged}.
// Both codes at zero is the FAIL condition — that is today's defect shape.
//
// It also asserts the pair DIVERGES: the English and Korean controls must not
// produce the same code set, or the pair is not a pair.
func TestModality_KoreanControlAnnounced(t *testing.T) {
	korean := modalityFindings(t, "FXK-001", modalityKoreanUnjudgedBody)

	malformed := countCode(korean, "ModalityMalformed")
	unjudged := countCode(korean, "ModalityUnjudged")

	if malformed+unjudged == 0 {
		t.Fatalf("Korean requirement produced NEITHER code — silence is indistinguishable from a pass, which is the defect")
	}
	if malformed != 0 || unjudged != 1 {
		t.Errorf("Korean requirement: ModalityMalformed=%d ModalityUnjudged=%d, want 0 and 1 (branch B announces, it does not judge)", malformed, unjudged)
	}

	english := modalityFindings(t, "FXE-001", modalityEnglishMalformedBody)
	if sameCodeSet(codeSet(english), codeSet(korean)) {
		t.Errorf("control pair did not diverge: english=%v korean=%v", codeSet(english), codeSet(korean))
	}
}

func sameCodeSet(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// --- AC-SLB-007 — silence vs conformance ----------------------------------

// TestModality_SilenceDiffersFromConformance uses the two acceptance.md §C
// fixtures verbatim. The conforming one must produce ZERO unjudged signals; the
// unjudgeable one at least one. "Both zero" is the FAIL condition.
//
// Note what makes the conforming fixture judgeable under branch B: it carries a
// literal `(SHALL)`. Recognising it requires the WORD-BOUNDARY contact condition
// of REQ-SLB-014 — the old leading-space form does not see `(SHALL)`. That is
// the coupling the SPEC asserted between AC-SLB-007 and AC-SLB-012, and it is
// why the repair is inside branch B rather than an extension of it.
func TestModality_SilenceDiffersFromConformance(t *testing.T) {
	conforming := modalityFindings(t, "FX-001", modalityConformingVerbatimBody)
	unjudgeable := modalityFindings(t, "FX-002", modalityUnjudgeableVerbatimBody)

	if got := countCode(conforming, "ModalityUnjudged"); got != 0 {
		t.Errorf("conforming fixture: ModalityUnjudged = %d, want 0", got)
	}
	if got := countCode(conforming, "ModalityMalformed"); got != 0 {
		t.Errorf("conforming fixture: ModalityMalformed = %d, want 0", got)
	}
	if got := countCode(unjudgeable, "ModalityUnjudged"); got < 1 {
		t.Errorf("unjudgeable fixture: ModalityUnjudged = %d, want >= 1", got)
	}
	if sameCodeSet(codeSet(conforming), codeSet(unjudgeable)) {
		t.Errorf("the two fixtures produced the SAME finding set (%v) — silence and conformance are still indistinguishable", codeSet(conforming))
	}
}

// --- AC-SLB-008 — the unjudged code is advisory ---------------------------

// TestModality_UnjudgedIsAdvisory asserts the new code reports without gating.
//
// The fixture is deliberately in the NARROW list shape, so REQEntry.Widened is
// false and the advisory marking CANNOT have come from the t385 widening path.
// The test asserts that precondition rather than assuming it.
func TestModality_UnjudgedIsAdvisory(t *testing.T) {
	doc := modalityDoc(t, "FXK-002", modalityKoreanUnjudgedBody)
	for _, r := range doc.REQs {
		if r.Widened {
			t.Fatalf("fixture precondition broken: REQ %s is Widened, so an advisory marking would be attributable to the widening path instead of to this code", r.ID)
		}
	}

	findings := (&EARSModalityRule{}).Check(doc, nil)
	seen := 0
	for _, f := range findings {
		if f.Code != "ModalityUnjudged" {
			continue
		}
		seen++
		if f.Severity != SeverityWarning {
			t.Errorf("ModalityUnjudged severity = %q, want %q", f.Severity, SeverityWarning)
		}
		if !f.Advisory {
			t.Errorf("ModalityUnjudged is not Advisory — --strict would escalate a signal that only says 'not judged'")
		}
	}
	if seen != 1 {
		t.Fatalf("ModalityUnjudged count = %d, want 1", seen)
	}

	for _, f := range findings {
		if f.Severity == SeverityError {
			t.Errorf("error-severity finding %s emitted on an unjudgeable fixture", f.Code)
		}
	}

	strict := &Report{Findings: findings, Strict: true}
	if strict.HasErrors() {
		t.Errorf("--strict escalated the unjudged finding to an error")
	}
}

// --- AC-SLB-008b — not judged is not reported as conforming ---------------

// TestModality_UnjudgedIsNotReportedConforming reads the verdict directly rather
// than inferring it from the absence of a finding. Absence is what the defect
// looks like, so absence cannot be the evidence.
//
// The mutant that routes the unjudged verdict into the conforming branch leaves
// the finding set of a well-formed fixture untouched and would pass any
// absence-only assertion; it turns this one RED.
func TestModality_UnjudgedIsNotReportedConforming(t *testing.T) {
	cases := []struct {
		name string
		text string
		want modalityVerdict
	}{
		{"unjudgeable korean (AC-SLB-007 verbatim)", "이 항목의 처리는 구현 재량에 맡긴다.", modalityUnjudged},
		{"conforming korean (AC-SLB-007 verbatim)", "도구는 판정 결과를 출력에 남겨야 한다(SHALL).", modalityJudgedConforming},
		{"english malformed", "When the tool runs, the system does report.", modalityJudgedMalformed},
	}
	for _, c := range cases {
		if got := judgeModality(c.text); got != c.want {
			t.Errorf("%s: judgeModality = %v, want %v", c.name, got, c.want)
		}
	}

	findings := modalityFindings(t, "FX-002b", modalityUnjudgeableVerbatimBody)
	if got := countCode(findings, "ModalityUnjudged"); got != 1 {
		t.Errorf("unjudgeable fixture: ModalityUnjudged = %d, want 1", got)
	}
	for _, f := range findings {
		low := strings.ToLower(f.Message)
		if strings.Contains(low, "conforms") || strings.Contains(low, "well-formed") || strings.Contains(low, " ok") {
			t.Errorf("a finding on the unjudgeable REQ reads as a conformance signal: %q", f.Message)
		}
	}
}

// --- AC-SLB-012 — the SHALL contact condition is a word boundary ----------

// TestModality_ShallContactIsWordBoundary carries BOTH acceptance.md fixtures
// verbatim plus the corrected one, and asserts for each whether it can make the
// mutant guard. The guard-capability assertions are what stop this test from
// silently degrading into three controls.
func TestModality_ShallContactIsWordBoundary(t *testing.T) {
	// The corrected discriminating fixture: word boundary sees the
	// parenthesized SHALL, the old leading-space condition does not.
	const discriminating = "When the tool runs, the tool reports.(SHALL)"
	if oldShallContact(discriminating) {
		t.Fatalf("corrected fixture is not discriminating: the OLD condition already matches %q", discriminating)
	}
	if judgeModality(discriminating) != modalityJudgedConforming {
		t.Errorf("word-boundary contact did not accept the parenthesized SHALL in %q", discriminating)
	}

	// The acceptance.md verbatim REQ-FX-011 line. Recorded, not argued: the old
	// condition already matches it, so reverting the repair leaves it at zero
	// findings and it cannot make the guard.
	const verbatim011 = "When the tool runs, the system shall report.(SHALL)"
	if !oldShallContact(verbatim011) {
		t.Errorf("acceptance.md's REQ-FX-011 was expected to be non-discriminating (old condition matches it); it no longer is — re-read the fixture")
	}

	for _, c := range []struct{ id, body string }{
		{"FXB-011", modalityShallOnlyParenthesizedBody},
		{"FX-011", modalityShallNoLeadingSpaceVerbatimBody},
		{"FX-012", modalityShallLeadingSpaceControlBody},
	} {
		findings := modalityFindings(t, c.id, c.body)
		if got := countCode(findings, "ModalityMalformed"); got != 0 {
			t.Errorf("%s: ModalityMalformed = %d, want 0", c.id, got)
		}
		if got := countCode(findings, "ModalityUnjudged"); got != 0 {
			t.Errorf("%s: ModalityUnjudged = %d, want 0 — an English prefix makes it judgeable", c.id, got)
		}
	}

	// A word boundary must not turn SHALL into a substring match: SHALLOW is
	// not SHALL. The old leading-space condition accepted " SHALLOW".
	const shallow = "When the tool runs, the pool is shallow."
	if !oldShallContact(shallow) {
		t.Fatalf("control broken: the old condition was expected to match %q", shallow)
	}
	if judgeModality(shallow) != modalityJudgedMalformed {
		t.Errorf("word-boundary contact accepted SHALLOW as SHALL in %q", shallow)
	}
}

// --- AC-SLB-010 (axis-2 half) — advisory does not leak into errors --------

// TestModality_AdvisoryDoesNotLeakIntoErrors runs a fixture that emits BOTH
// advisory paths at once — a table-collected entry (Widened) and an unjudgeable
// list entry (the new code) — and requires the error-severity count to stay at
// zero on both.
func TestModality_AdvisoryDoesNotLeakIntoErrors(t *testing.T) {
	doc := modalityDoc(t, "FXM-001", modalityMixedAdvisoryBody)

	var table, list int
	for _, r := range doc.REQs {
		switch r.Source {
		case REQSourceTable:
			table++
		default:
			list++
		}
	}
	if table == 0 || list == 0 {
		t.Fatalf("fixture precondition broken: table=%d list=%d — the mixed fixture must exercise BOTH advisory paths", table, list)
	}

	findings := (&EARSModalityRule{}).Check(doc, nil)
	if got := countCode(findings, "ModalityUnjudged"); got < 1 {
		t.Errorf("mixed fixture: ModalityUnjudged = %d, want >= 1", got)
	}
	for _, f := range findings {
		if f.Severity == SeverityError {
			t.Errorf("error-severity finding %s leaked from an advisory path", f.Code)
		}
	}
	if (&Report{Findings: findings, Strict: true}).HasErrors() {
		t.Errorf("--strict escalated an advisory finding")
	}
}

// --- corpus census — this milestone's deliverable figure ------------------

// TestModality_CorpusUnjudgedCensus counts, across the whole corpus, how many
// collected REQ entries this linter has NO OPINION about, split by the shape the
// entry was written in (REQEntry.Source). That count is what branch B exists to
// produce: it sizes the follow-up card that would actually judge Korean modality.
//
// It also counts the entries the WORD-BOUNDARY repair recovered — entries the
// old leading-space contact condition reported as ModalityMalformed and the
// repaired condition accepts.
//
// It asserts only NON-VACUITY. The numbers are the deliverable, not a threshold,
// and they are RE-DERIVED on every run rather than stored — a stored figure
// silently detaches from the corpus that moved underneath it.
//
// The corpus is read READ-ONLY. Nothing under .moai/specs is written.
func TestModality_CorpusUnjudgedCensus(t *testing.T) {
	root := findRepoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, ".moai/specs/SPEC-*/spec.md"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) < 100 {
		t.Fatalf("corpus glob matched %d files — too few for this census to mean anything", len(paths))
	}

	var entries, unjudged, unjudgedFromTable, malformed, conforming int
	var recoveredByWordBoundary int
	specsWithUnjudged := map[string]bool{}

	for _, p := range paths {
		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			continue
		}
		for _, req := range parseREQsWithProvenance(doc.Body) {
			entries++
			switch judgeModality(req.Text) {
			case modalityUnjudged:
				unjudged++
				specsWithUnjudged[filepath.Base(filepath.Dir(p))] = true
				if req.Source == REQSourceTable {
					unjudgedFromTable++
				}
			case modalityJudgedMalformed:
				malformed++
			default:
				conforming++
				// The old contact condition required a LEADING SPACE. An entry
				// that opens with a modality prefix, carries a word-bounded
				// SHALL, and has no " SHALL" is one the repair recovered.
				if !oldShallContact(req.Text) {
					upper := strings.ToUpper(req.Text)
					for _, prefix := range modalityPrefixes {
						if strings.HasPrefix(upper, prefix) {
							recoveredByWordBoundary++
							break
						}
					}
				}
			}
		}
	}

	if unjudged == 0 {
		t.Fatal("no corpus entry was unjudged — the census is vacuous, and branch B has nothing to announce")
	}
	if malformed == 0 || conforming == 0 {
		t.Fatalf("census is one-sided: malformed=%d conforming=%d — the verdict is not discriminating", malformed, conforming)
	}
	if recoveredByWordBoundary == 0 {
		t.Fatal("the word-boundary repair recovered nothing on this corpus — REQ-SLB-014's movement claim would be unattributable")
	}

	t.Logf("REQ entries=%d | unjudged=%d (across %d SPEC dirs; %d table-sourced, %d list-sourced) | malformed=%d | conforming=%d | recovered by word boundary=%d",
		entries, unjudged, len(specsWithUnjudged), unjudgedFromTable, unjudged-unjudgedFromTable,
		malformed, conforming, recoveredByWordBoundary)
}

// --- boundary guard added AFTER a mutant escaped --------------------------

// TestModality_EveryEnglishPrefixIsStillJudged walks all five English modality
// openers with a SHALL-less body and requires each to be judged MALFORMED — not
// announced as unjudged.
//
// [HARD] This guard did NOT exist on the first mutant pass. Mutant M8 (drop
// "THE " from modalityPrefixes) passed the ENTIRE internal/spec package,
// including the pre-existing EARS tests: every fixture in the milestone happened
// to open with "When", so losing the Ubiquitous opener silently reclassified
// every `The system does X` requirement from malformed to unjudged and nothing
// noticed. That escape is recorded in progress.md §E.2 and is not erased by this
// guard existing now — "caught because a guard was added afterwards" is a
// different fact from "a guard was there".
func TestModality_EveryEnglishPrefixIsStillJudged(t *testing.T) {
	bodies := map[string]string{
		"WHEN ":  "When the tool runs, the system does report.",
		"WHILE ": "While the tool runs, the system does report.",
		"WHERE ": "Where the flag is set, the system does report.",
		"IF ":    "If the tool runs, the system does report.",
		"THE ":   "The system does report on every run.",
	}
	if len(bodies) != len(modalityPrefixes) {
		t.Fatalf("prefix table drifted: %d fixtures for %d prefixes — a dropped prefix would go unmeasured", len(bodies), len(modalityPrefixes))
	}
	for prefix, text := range bodies {
		if got := judgeModality(text); got != modalityJudgedMalformed {
			t.Errorf("prefix %q: judgeModality(%q) = %v, want malformed — the opener is no longer recognised, so the requirement is silently announced instead of judged", prefix, text, got)
		}
	}
	for _, prefix := range modalityPrefixes {
		if _, ok := bodies[prefix]; !ok {
			t.Errorf("prefix %q has no fixture — add one, or a drop of it goes unmeasured", prefix)
		}
	}
}
