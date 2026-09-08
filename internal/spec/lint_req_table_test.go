package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-SPEC-LINT-BLIND-AXES-001 axis 1 (card t518, milestone M-A1).
//
// The collector was anchored to a markdown list bullet, so a SPEC that states
// its requirements as a TABLE produced an empty doc.REQs and the four REQ-based
// rules never visited it. These tests are the RED half of that repair.
//
// Two disciplines from acceptance.md §A bind every test here:
//
//   - CONTROL PAIRS. Every assertion is made on BOTH a list-form fixture and a
//     table-form fixture. A one-sided fixture cannot distinguish "both
//     collected" from "both ignored" — which is the exact shape in which this
//     tool went silent.
//   - NO EXIT CODES. Assertions are on match counts and named finding codes.
//
// Mutant results (including the mutants that were NOT caught) are recorded in
// .moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/progress.md §E.2.

// writeTableSpecFixture writes a fixture SPEC under t.TempDir() and returns its
// path. Fixtures never touch .moai/specs (acceptance.md §A rule 5).
func writeTableSpecFixture(t *testing.T, id, body string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "SPEC-"+id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	fm := strings.Join([]string{
		"---",
		"id: SPEC-" + id,
		`title: "fixture"`,
		`version: "0.1.0"`,
		"status: draft",
		"created: 2026-09-07",
		"updated: 2026-09-07",
		"author: t",
		"priority: P2",
		`phase: "v3.2.0 target"`,
		`module: "internal/spec"`,
		"lifecycle: spec-anchored",
		`tags: "fixture"`,
		"---",
		"",
	}, "\n")
	p := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(p, []byte(fm+body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

// listFixtureBody states three requirements in LIST form. It is the control
// half of every pair below.
//
// The lines are deliberately in the shape the NARROW reqLinePattern accepts —
// `- REQ-<DOM>-NNN-NNN: text`, four segments, colon separator, no bold markers.
// AC-SLB-001a compares the live path against parseREQs, and a line the narrow
// pattern cannot see makes that comparison iterate zero entries and assert
// nothing. A control that collects nothing is not a control. (The widened
// bold/em-dash list shapes are already covered by the t385 tests in
// lint_req_widen_test.go; re-testing them here would not make this pair a pair.)
const listFixtureBody = `## Requirements

- REQ-FXL-001-001: The system SHALL record every rejection.
- REQ-FXL-001-002: 도구는 목록 형식 수집을 보존해야 한다
- REQ-FXL-001-003: 도구는 판정 결과를 출력에 남겨야 한다(SHALL).
`

// tableFixtureBody states the SAME three requirements in TABLE form. Each row
// carries a DIFFERENT L1 marker — `SHALL`, `해야 한다`, `해서는 안 된다` — so all
// three lexicon entries are exercised and discriminator C-d admits every row.
//
// L1 matching is CASE-SENSITIVE: a lowercase "shall" does not contact it. That
// is a property of the chosen discriminator, not an oversight, so the fixture
// spells the marker the way the corpus does.
const tableFixtureBody = `## Requirements

| 요구 | modality | 내용 |
|---|---|---|
| REQ-FXT-001 | Ubiquitous | The system shall record every rejection (SHALL). |
| REQ-FXT-002 | Ubiquitous | 도구는 목록 형식 수집을 보존해야 한다 |
| REQ-FXT-003 | Unwanted | 도구는 판정하지 않은 요구사항을 적합하다고 보고해서는 안 된다 |
`

// dispositionFixtureBody carries the two NON-definition table shapes, both
// transcribed verbatim from .moai/specs/SPEC-BINLAG-INVOCATION-001/spec.md
// (:124 disposition, :140 tracking). acceptance.md AC-SLB-004 fixes them
// verbatim on purpose: an invented sentence can be chosen so the discriminator
// passes itself.
const dispositionFixtureBody = "## 처분\n" +
	"\n" +
	"| 요구 | B 아래에서의 성격 | 사유 |\n" +
	"|---|---|---|\n" +
	"| REQ-BLI-001 | **이 카드가 실제로 세우는 것** | 인용 규율이 곧 이 요구다. 거처: `verification-claim-integrity.md` §2.2 |\n" +
	"\n" +
	"## 추적\n" +
	"\n" +
	"| 요구 | 수락 |\n" +
	"|---|---|\n" +
	"| REQ-BLI-001 | AC-BLI-001 |\n"

// --- AC-SLB-001a (control: list form unchanged) — REQ-SLB-001, REQ-SLB-003 ---

func TestTableCollection_ListFormUnchanged(t *testing.T) {
	body := listFixtureBody

	narrow := parseREQs(body)
	live := parseREQsWithProvenance(body)

	// Every entry the narrow path collected must appear in the live path with
	// byte-identical ID, Text and Line, carrying no advisory marking.
	byLine := make(map[int]REQEntry, len(live))
	for _, r := range live {
		byLine[r.Line] = r
	}
	for _, n := range narrow {
		got, ok := byLine[n.Line]
		if !ok {
			t.Fatalf("narrow entry %s (line %d) vanished from the live path", n.ID, n.Line)
		}
		if got.ID != n.ID || got.Text != n.Text {
			t.Errorf("line %d: live (%q,%q) != narrow (%q,%q)", n.Line, got.ID, got.Text, n.ID, n.Text)
		}
		if got.Widened {
			t.Errorf("line %d (%s): narrow-collected entry must not be advisory", n.Line, n.ID)
		}
		if got.Source != REQSourceList {
			t.Errorf("line %d (%s): source = %v, want list", n.Line, n.ID, got.Source)
		}
	}
	if len(live) != len(narrow) {
		t.Errorf("list-only fixture: live collected %d entries, narrow %d — the table branch must add nothing here", len(live), len(narrow))
	}
}

// --- AC-SLB-001b (table form collected, all advisory) — REQ-SLB-001, REQ-SLB-002 ---

func TestTableCollection_TableFormCollected(t *testing.T) {
	got := parseREQsWithProvenance(tableFixtureBody)

	want := []string{"REQ-FXT-001", "REQ-FXT-002", "REQ-FXT-003"}
	if len(got) != len(want) {
		t.Fatalf("collected %d entries, want %d (one per definition row): %#v", len(got), len(want), got)
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Errorf("entry %d: ID = %q, want %q", i, got[i].ID, id)
		}
		if !got[i].Widened {
			t.Errorf("entry %d (%s): table-collected entry must carry the advisory marking", i, got[i].ID)
		}
		if got[i].Source != REQSourceTable {
			t.Errorf("entry %d (%s): source = %v, want table", i, got[i].ID, got[i].Source)
		}
	}
	// The body text must be the requirement statement, not the whole row.
	if got[0].Text != "The system shall record every rejection (SHALL)." {
		t.Errorf("entry 0 Text = %q, want the requirement body cell", got[0].Text)
	}
}

// --- AC-SLB-001c (the control pair actually diverges) — REQ-SLB-012 ---

func TestTableCollection_ControlPairDiverges(t *testing.T) {
	list := parseREQsWithProvenance(listFixtureBody)
	table := parseREQsWithProvenance(tableFixtureBody)

	if len(list) == 0 || len(table) == 0 {
		t.Fatalf("control pair is vacuous: list=%d table=%d — both halves must collect", len(list), len(table))
	}

	listAdvisory := 0
	for _, r := range list {
		if r.Widened {
			listAdvisory++
		}
	}
	tableAdvisory := 0
	for _, r := range table {
		if r.Widened {
			tableAdvisory++
		}
	}
	if listAdvisory != 0 {
		t.Errorf("list half: %d advisory entries, want 0", listAdvisory)
	}
	if tableAdvisory != len(table) {
		t.Errorf("table half: %d of %d advisory, want all", tableAdvisory, len(table))
	}
	if listAdvisory == tableAdvisory {
		t.Errorf("the two halves did not diverge (both %d advisory) — a pair that does not split proves nothing", listAdvisory)
	}
}

// --- AC-SLB-004 (discriminator rejects disposition/tracking tables) — REQ-SLB-004 ---
//
// This is an ABSENCE assertion: it is satisfied for free when table collection
// does not exist at all. It is only meaningful read together with
// TestTableCollection_TableFormCollected and with the recorded mutant M2
// (discriminator removed → this test goes RED).

// discriminatorBoundaryFixtureBody pins two properties of C-d that the pair
// fixtures above leave free. Both were found by mutation: mutants M6
// (L1 matching made case-insensitive) and M7 (row pattern loses its bold-marker
// tolerance) both passed the whole suite before this fixture existed.
//
// Row 1 must be REJECTED — L1 matching is case-sensitive, and `shall` in
// lowercase is not `SHALL`. That is a decided property of the chosen
// discriminator: widening it changes the corpus figures C-d was selected on.
//
// Row 2 must be ADMITTED — the corpus writes definition rows with bold IDs
// (`| **REQ-HCW-001** | Ubiquitous | … SHALL … |`, 4 rows in
// SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001), so the row pattern's `\**`
// tolerance is load-bearing rather than decorative.
const discriminatorBoundaryFixtureBody = `## Requirements

| 요구 | modality | 내용 |
|---|---|---|
| REQ-FXB-001 | Ubiquitous | The system shall be written in lowercase only. |
| **REQ-FXB-002** | Ubiquitous | The system SHALL be reachable through a bold ID cell. |
`

func TestTableCollection_DiscriminatorBoundaries(t *testing.T) {
	got := parseREQsWithProvenance(discriminatorBoundaryFixtureBody)

	ids := make([]string, 0, len(got))
	for _, r := range got {
		ids = append(ids, r.ID)
	}
	want := []string{"REQ-FXB-002"}
	if len(ids) != len(want) || (len(ids) == 1 && ids[0] != want[0]) {
		t.Fatalf("collected %v, want %v — row 1 must be rejected (lowercase 'shall' is not an L1 marker) and row 2 admitted (bold ID cell)", ids, want)
	}
}

func TestTableCollection_DiscriminatorRejectsDispositionTables(t *testing.T) {
	got := parseREQsWithProvenance(dispositionFixtureBody)
	if len(got) != 0 {
		t.Fatalf("disposition/tracking rows must be rejected, got %d collected: %#v", len(got), got)
	}
}

// --- AC-SLB-002 / AC-SLB-010 (advisory does not move the error tally) ---
//
// REQ-SLB-002, REQ-SLB-010. The table fixture below is built so a REQ-based
// rule actually FIRES on a table-collected entry: the L1 marker sits in the
// modality cell (so C-d admits the row) while the body cell is an English
// event-driven sentence with no SHALL (so ModalityMalformed fires on it).
// Placing the marker outside the body cell is the only way to hold both
// conditions at once under C-d, and it is a shape the corpus does use.

const tableMalformedFixtureBody = `## Requirements

| 요구 | modality | 내용 |
|---|---|---|
| REQ-FXM-001 | 해야 한다 | When the tool runs, the system emits a signal. |
`

const listMalformedFixtureBody = `## Requirements

- REQ-FXM-002-001: When the tool runs, the system emits a signal.
`

func lintPathFindings(t *testing.T, path string) []Finding {
	t.Helper()
	linter := NewLinter(LinterOptions{})
	rep, err := linter.Lint([]string{path})
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	return rep.Findings
}

func countBy(findings []Finding, pred func(Finding) bool) int {
	n := 0
	for _, f := range findings {
		if pred(f) {
			n++
		}
	}
	return n
}

func TestTableCollection_AdvisoryDoesNotGate(t *testing.T) {
	tablePath := writeTableSpecFixture(t, "FXM-001", tableMalformedFixtureBody)
	listPath := writeTableSpecFixture(t, "FXM-002", listMalformedFixtureBody)

	tableFindings := lintPathFindings(t, tablePath)
	listFindings := lintPathFindings(t, listPath)

	isModality := func(f Finding) bool { return f.Code == "ModalityMalformed" }
	isError := func(f Finding) bool { return f.Severity == SeverityError }

	// The finding EXISTS on the table-collected entry — otherwise this test is
	// an absence assertion dressed as a presence one.
	tableModality := countBy(tableFindings, isModality)
	if tableModality != 1 {
		t.Fatalf("table fixture: ModalityMalformed count = %d, want 1", tableModality)
	}
	// ...and it is advisory.
	advisoryModality := countBy(tableFindings, func(f Finding) bool { return isModality(f) && f.Advisory })
	if advisoryModality != 1 {
		t.Errorf("table fixture: advisory ModalityMalformed = %d, want 1", advisoryModality)
	}
	if n := countBy(tableFindings, func(f Finding) bool { return isModality(f) && isError(f) }); n != 0 {
		t.Errorf("table fixture: error-severity ModalityMalformed = %d, want 0", n)
	}

	// The control half proves the rule is not simply dead: the same sentence in
	// LIST form is an error-severity finding and is NOT advisory.
	listModality := countBy(listFindings, func(f Finding) bool { return isModality(f) && isError(f) && !f.Advisory })
	if listModality != 1 {
		t.Errorf("list fixture: error-severity non-advisory ModalityMalformed = %d, want 1", listModality)
	}
}

// --- AC-SLB-011 (source attributes the delta; it never decides severity) ---
//
// REQ-SLB-013, REQ-SLB-002.

func TestTableCollection_SourceRecordedPerOrigin(t *testing.T) {
	list := parseREQsWithProvenance(listFixtureBody)
	table := parseREQsWithProvenance(tableFixtureBody)

	for _, r := range list {
		if r.Source != REQSourceList {
			t.Errorf("list fixture entry %s: source = %v, want list", r.ID, r.Source)
		}
	}
	for _, r := range table {
		if r.Source != REQSourceTable {
			t.Errorf("table fixture entry %s: source = %v, want table", r.ID, r.Source)
		}
	}
}

// TestTableCollection_SourceDoesNotDecideSeverity flips the source value on
// every entry and asserts the severity distribution is unchanged. Without this
// branch, the source field could quietly become a second severity axis — which
// spec.md §B.2 forbids for a named reason.
func TestTableCollection_SourceDoesNotDecideSeverity(t *testing.T) {
	entries := append(parseREQsWithProvenance(listFixtureBody), parseREQsWithProvenance(tableFixtureBody)...)
	if len(entries) == 0 {
		t.Fatal("no entries collected — the flip test would be vacuous")
	}

	distribution := func(rs []REQEntry) map[Severity]int {
		d := map[Severity]int{}
		for _, r := range rs {
			sev, _ := reqFindingSeverity(r, SeverityError)
			d[sev]++
		}
		return d
	}

	before := distribution(entries)

	flipped := make([]REQEntry, len(entries))
	copy(flipped, entries)
	for i := range flipped {
		if flipped[i].Source == REQSourceTable {
			flipped[i].Source = REQSourceList
		} else {
			flipped[i].Source = REQSourceTable
		}
	}
	after := distribution(flipped)

	if len(before) != len(after) {
		t.Fatalf("severity distribution changed shape after flipping source: %v -> %v", before, after)
	}
	for sev, n := range before {
		if after[sev] != n {
			t.Errorf("severity %v: %d before flip, %d after — source must not reach reqFindingSeverity", sev, n, after[sev])
		}
	}
}

// --- AC-SLB-003 (list-form findings unchanged, corpus scale) — REQ-SLB-003 ---
//
// The claim is that adding table collection changed NOTHING about the findings
// that come from list-collected REQs. A fixture cannot carry that claim: the
// risk lives in the interaction between the two entry sets across 780-odd real
// documents, not in a hand-written pair.
//
// The comparison is between TWO DERIVATIONS ON ONE TREE, not two trees. The
// narrow path (parseREQs) is preserved by parseREQsWithProvenance, so:
//
//	(a) the REQ rules over parseREQs entries      — the pre-change behaviour
//	(b) the NON-ADVISORY findings of the same rules over the live path
//
// must agree code by code. There is no "before tree" to collect a golden from,
// and there never was — the acceptance criteria were rewritten around that fact.
//
// The known way for this to break is ordering-dependent: a table row and a list
// definition sharing one REQ ID, with the table row FIRST, would make
// REQIDUniquenessRule report the LIST entry as the duplicate — a non-advisory
// finding that did not exist before. No such document exists in the corpus
// today (measured: DuplicateREQID is 0 on both sides), which is why no
// machinery guards against it. This test is what would notice.
//
// The corpus is read READ-ONLY. Nothing under .moai/specs is written.
func TestTableCollection_CorpusListFindingsUnchanged(t *testing.T) {
	root := findRepoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, ".moai/specs/SPEC-*/spec.md"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) < 100 {
		t.Fatalf("corpus glob matched %d files — too few for this assertion to mean anything", len(paths))
	}

	rules := []Rule{&EARSModalityRule{}, &REQIDUniquenessRule{}, &CoverageRule{}}
	narrowCounts := map[string]int{}
	liveNonAdvisory := map[string]int{}
	scanned := 0

	for _, p := range paths {
		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			continue
		}
		scanned++
		body := doc.Body

		doc.REQs = parseREQs(body)
		for _, r := range rules {
			for _, f := range r.Check(doc, nil) {
				// The advisory filter is applied on BOTH sides, or the two
				// counters are not measuring the same quantity. CoverageRule
				// (see CoverageRule.Check) sets Advisory unconditionally, so
				// the live side can never count it, while an unfiltered narrow
				// side counts it the moment any finding appears — an
				// asymmetry that reads as a regression when the corpus grows.
				if !f.Advisory {
					narrowCounts[f.Code]++
				}
			}
		}

		doc.REQs = parseREQsWithProvenance(body)
		for _, r := range rules {
			for _, f := range r.Check(doc, nil) {
				if !f.Advisory {
					liveNonAdvisory[f.Code]++
				}
			}
		}
	}

	if scanned == 0 {
		t.Fatal("scanned 0 documents — the assertion would be vacuous")
	}
	codes := map[string]bool{}
	for c := range narrowCounts {
		codes[c] = true
	}
	for c := range liveNonAdvisory {
		codes[c] = true
	}
	if len(codes) == 0 {
		t.Fatalf("no findings of any code over %d documents — the assertion would be vacuous", scanned)
	}
	for c := range codes {
		if narrowCounts[c] != liveNonAdvisory[c] {
			t.Errorf("code %s: narrow-path findings = %d, live non-advisory findings = %d (scanned %d docs) — the table branch moved a list-form finding",
				c, narrowCounts[c], liveNonAdvisory[c], scanned)
		}
	}
	t.Logf("scanned %d documents; per-code narrow == live-non-advisory: %v", scanned, narrowCounts)
}

// TestTableCollection_CorpusTableEntryCensus is the instrument half: it counts
// what table collection actually adds across the corpus, so the recount in a
// later milestone has a Go-derived figure rather than only the python reader's.
// It asserts only NON-VACUITY — the numbers are the deliverable, not a
// threshold.
func TestTableCollection_CorpusTableEntryCensus(t *testing.T) {
	root := findRepoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, ".moai/specs/SPEC-*/spec.md"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	rowsSeen, admitted, rejected := 0, 0, 0
	blindAdmitted, specsWithTableEntries := 0, 0

	for _, p := range paths {
		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			continue
		}
		listEntries := parseREQsWide(doc.Body)
		tableEntries := parseREQsTable(doc.Body)
		for _, line := range strings.Split(doc.Body, "\n") {
			if reqTableRowPattern.MatchString(line) {
				rowsSeen++
				if isTableDefinitionRow(line) {
					admitted++
				} else {
					rejected++
				}
			}
		}
		if len(tableEntries) > 0 {
			specsWithTableEntries++
			if len(listEntries) == 0 {
				blindAdmitted += len(tableEntries)
			}
		}
	}

	if admitted == 0 {
		t.Fatal("no table row was admitted anywhere in the corpus — the census is vacuous")
	}
	if rejected == 0 {
		t.Fatal("no table row was rejected anywhere in the corpus — the discriminator is vacuous")
	}
	t.Logf("table rows seen=%d admitted=%d rejected=%d | SPECs gaining table entries=%d | admitted inside collector-blind SPECs=%d",
		rowsSeen, admitted, rejected, specsWithTableEntries, blindAdmitted)
}
