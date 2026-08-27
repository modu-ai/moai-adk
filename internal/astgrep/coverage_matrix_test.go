package astgrep

// RED-phase tests for the coverage-matrix checker (SPEC M2).
//
// Ordering discipline: these assertions were authored against known-failing
// synthetic inputs BEFORE the checker logic existed; the verbatim failing
// output for each class is captured as E8 evidence. The seeded-cells test
// loads rule ids from the REAL shipped ruleset so class 4's resolution source
// is exercised end-to-end, not only against a stub.

import (
	"os"
	"strings"
	"testing"
)

// fullMatrix builds one syntactically valid cell per Cartesian key. All are
// IMPLEMENTED with a per-key id present in stubRuleset().
func fullMatrix(t *testing.T) []MatrixCell {
	t.Helper()
	cells := make([]MatrixCell, 0, len(MatrixAxes.Families)*len(MatrixAxes.Languages))
	for _, fam := range MatrixAxes.Families {
		for _, lang := range MatrixAxes.Languages {
			cells = append(cells, MatrixCell{
				Family:   fam,
				Language: lang,
				State:    StateImplemented,
				RuleID:   "rule-" + strings.ToLower(fam) + "-" + lang,
			})
		}
	}
	return cells
}

func stubRuleset(cells []MatrixCell) map[string]bool {
	ids := map[string]bool{}
	for _, c := range cells {
		if c.RuleID != "" {
			ids[c.RuleID] = true
		}
	}
	return ids
}

func findingsOfClass(t *testing.T, got []CheckerFinding, want FailureClass) []CheckerFinding {
	t.Helper()
	var out []CheckerFinding
	for _, f := range got {
		if f.Class == want {
			out = append(out, f)
		}
	}
	return out
}

// Class 1a — a deleted cell must produce a key-set mismatch naming the key.
func TestCheckerKeySetMismatchDetectsDeletedCell(t *testing.T) {
	cells := fullMatrix(t)
	dropKey := "F3/ruby"
	filtered := cells[:0]
	for _, c := range cells { // safe reuse: cells rebuilt per test
		if c.Key() != dropKey {
			filtered = append(filtered, c)
		}
	}
	got := CheckCoverageMatrix(filtered, stubRuleset(append([]MatrixCell(nil), filtered...)))
	mismatches := findingsOfClass(t, got, ClassKeySetMismatch)
	if len(mismatches) == 0 {
		t.Fatalf("expected at least one key-set mismatch finding, got none; all findings: %+v", got)
	}
	found := false
	for _, f := range mismatches {
		if strings.Contains(f.Detail, dropKey) {
			found = true
		}
	}
	if !found {
		t.Fatalf("mismatch findings do not name dropped key %s: %+v", dropKey, mismatches)
	}
}

// Class 1b — THE load-bearing case (AC defect (b)): substitute one cell for
// another so the count stays exactly 112. A count-based checker reports
// nothing here; the set comparison must name both keys.
func TestCheckerKeySetMismatchDetectsSubstitutionAtConstantCount(t *testing.T) {
	cells := fullMatrix(t)
	const victimLang = "java" // F1/java is displaced by an extra F1/cpp row
	substituted := make([]MatrixCell, 0, len(cells))
	sawCppOnce := false
	for _, c := range cells {
		switch {
		case c.Family == "F1" && c.Language == victimLang:
			continue // deleted target
		case c.Family == "F1" && c.Language == "cpp" && !sawCppOnce:
			// duplicate: emit this row, then fall through to emit it again below
			substituted = append(substituted, c)
			substituted = append(substituted, MatrixCell{
				Family: "F1", Language: "cpp",
				State: StateImplemented, RuleID: c.RuleID,
			})
			sawCppOnce = true
		default:
			substituted = append(substituted, c)
		}
	}
	if len(substituted) != 112 {
		t.Fatalf("fixture error: substitution should preserve count 112, got %d", len(substituted))
	}
	got := CheckCoverageMatrix(substituted, stubRuleset(substituted))
	mismatches := findingsOfClass(t, got, ClassKeySetMismatch)
	if len(mismatches) == 0 {
		t.Fatalf("substituted matrix with correct cardinality passed silently; "+
			"a count-based result — expected key-set mismatch naming F1/%s and duplicated F1/cpp", victimLang)
	}
	var namesMissing, namesDup bool
	for _, f := range mismatches {
		if strings.Contains(f.Detail, "F1/"+victimLang) {
			namesMissing = true
		}
		if strings.Contains(f.Detail, "F1/cpp") {
			namesDup = true
		}
	}
	if !namesMissing || !namesDup {
		t.Fatalf("mismatch must name BOTH the missing F1/%s and the duplicated F1/cpp; got %+v", victimLang, mismatches)
	}
}

// Class 2 — a cell with neither rule id nor rationale (and not the
// plan-authorized PENDING marker) is unresolved.
func TestCheckerUnresolvedCellDetected(t *testing.T) {
	cells := fullMatrix(t)
	for i := range cells {
		if cells[i].Key() == "F6/elixir" {
			cells[i].State = "" // neither IMPLEMENTED nor EXEMPT nor PENDING
		}
	}
	got := CheckCoverageMatrix(cells, stubRuleset(cells))
	unresolved := findingsOfClass(t, got, ClassUnresolvedCell)
	if len(unresolved) != 1 {
		t.Fatalf("expected exactly one unresolved-cell finding, got %d; all findings: %+v", len(unresolved), got)
	}
	if unresolved[0].Key != "F6/elixir" {
		t.Fatalf("unresolved finding names %q, want F6/elixir", unresolved[0].Key)
	}
}

// Class 3 — evidence-rule enforcement: an EXEMPT whose evidence carries
// neither cite: nor probe: fails even when the rationale sounds plausible.
func TestCheckerUnevidencedExemptionDetected(t *testing.T) {
	cells := fullMatrix(t)
	for i := range cells {
		if cells[i].Key() == "F7/rust" {
			cells[i].State = StateExempt
			cells[i].RuleID = ""
			cells[i].Rationale = "no log-format sink exists in rust"
			cells[i].Evidence = "the language simply does not have this construct." // bare assertion
		}
	}
	got := CheckCoverageMatrix(cells, stubRuleset(cells))
	unevidenced := findingsOfClass(t, got, ClassUnevidencedExemption)
	if len(unevidenced) != 1 {
		t.Fatalf("expected exactly one unevidenced-exemption finding, got %d; all findings: %+v", len(unevidenced), got)
	}
	if unevidenced[0].Key != "F7/rust" {
		t.Fatalf("unevidenced finding names %q, want F7/rust", unevidenced[0].Key)
	}
}

// Class 3 positive controls — citation OR probe each satisfy the obligation.
func TestCheckerEvidencedExemptionsPass(t *testing.T) {
	base := fullMatrix(t)
	citeCell := base[0]
	citeCell.State, citeCell.RuleID, citeCell.Rationale, citeCell.Evidence =
		StateExempt, "", "no shell-string sink", "cite: cpp stdlib reference, process-spawn section"
	probeCell := base[1]
	probeCell.State, probeCell.RuleID, probeCell.Rationale, probeCell.Evidence =
		StateExempt, "", "Logger macros take no user format string",
		`probe: sg run -p '...' -l elixir --stdin -> no match`
	got := CheckCoverageMatrix(base, stubRuleset(base))
	if n := len(findingsOfClass(t, got, ClassUnevidencedExemption)); n != 0 {
		t.Fatalf("cite:/probe:-evidenced exemptions reported unevidenced (%d): %+v", n, got)
	}
}

// Class 4 — a cell naming a rule absent from the shipped ruleset dangles:
// sg test never reads the matrix, so nothing else catches the drift.
func TestCheckerDanglingRuleIDDetected(t *testing.T) {
	cells := fullMatrix(t)
	ruleset := stubRuleset(cells)
	delete(ruleset, "rule-f2-swift") // resolve every other id, leave one dangling
	got := CheckCoverageMatrix(cells, ruleset)
	dangling := findingsOfClass(t, got, ClassDanglingRuleID)
	if len(dangling) != 1 {
		t.Fatalf("expected exactly one dangling-rule-id finding, got %d; all findings: %+v", len(dangling), got)
	}
	if dangling[0].Key != "F2/swift" || !strings.Contains(dangling[0].Detail, "rule-f2-swift") {
		t.Fatalf("dangling finding = %+v, want key F2/swift naming rule-f2-swift", dangling[0])
	}
}

// Stop condition half — the exact 14 already-implemented cells from the
// measured inventory pass against the REAL shipped ruleset.
func TestCheckerPassesSeededCellsAgainstRealRuleset(t *testing.T) {
	ids, err := LoadRulesetIDs("../../internal/template/templates/.moai/config/astgrep-rules")
	if err != nil {
		t.Fatalf("load shipped ruleset ids: %v", err)
	}
	if len(ids) < 21 { // 26 rules collapse to 21 distinct ids under case-id keying
		t.Fatalf("shipped ruleset yielded %d distinct ids, want >= 21", len(ids))
	}
	seeded := SeededImplementedCells()
	if len(seeded) != 14 {
		t.Fatalf("seeded implemented cells = %d, want exactly 14 (measured inventory)", len(seeded))
	}
	got := CheckCoverageMatrix(seeded, ids)
	for _, cls := range []FailureClass{
		ClassUnresolvedCell, ClassUnevidencedExemption, ClassDanglingRuleID,
	} {
		if f := findingsOfClass(t, got, cls); len(f) > 0 {
			t.Fatalf("seeded cells tripped %s: %+v", cls, f)
		}
	}
}

// Parser tests (GREEN stage) — pipe-table extraction, fenced-block skipping.

func TestParseCoverageMatrixExtractsRowsAndSkipsFences(t *testing.T) {
	doc := `intro prose | not a table

| family | language | state | rule id / rationale | evidence |
|---|---|---|---|---|
| F2 | go | IMPLEMENTED | sec-hardcoded-credential | internal/astgrep/testdata/rule-tests |
| F6 | ruby | EXEMPT | no csrf surface | cite: docs |

## Excluded languages

~~~text
sg run -l r --stdin
error: invalid value 'r' ... r is not supported!
~~~

| family | language | state | rule id / rationale | evidence |
|---|---|---|---|---|
| F3 | go | PENDING | pending fill | - |
`
	cells, err := ParseCoverageMatrix(doc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cells) != 3 {
		t.Fatalf("expected 3 parsed cells (fence rows excluded), got %d: %+v", len(cells), cells)
	}
	wantKeys := map[string]string{
		"F2/go":   StateImplemented,
		"F6/ruby": StateExempt,
		"F3/go":   StatePending,
	}
	for _, c := range cells {
		state, ok := wantKeys[c.Key()]
		if !ok {
			t.Errorf("unexpected cell %s (%+v)", c.Key(), c)
			continue
		}
		if c.State != state {
			t.Errorf("cell %s state = %q, want %q", c.Key(), c.State, state)
		}
	}
}

func TestParseCoverageMatrixRejectsMalformedRow(t *testing.T) {
	doc := `| family | language | state | rule id / rationale | evidence |
|---|---|---|---|---|
| F1 go only no pipes no trailing
`
	if _, err := ParseCoverageMatrix(doc); err == nil {
		t.Fatal("malformed table row must return an error, got nil")
	}
}

// --- Integration: the checker runs against the REAL artifacts in CI -------

const (
	// Paths are relative to this package directory (internal/astgrep).
	matrixDocPath   = "../../.moai/specs/SPEC-ASTGREP-LANG16-001/coverage-matrix.md"
	shippedRulesDir = "../../internal/template/templates/.moai/config/astgrep-rules"
)

func loadRealMatrix(t *testing.T) []MatrixCell {
	t.Helper()
	data, err := os.ReadFile(matrixDocPath)
	if err != nil {
		t.Fatalf("read matrix document: %v", err)
	}
	cells, err := ParseCoverageMatrix(string(data))
	if err != nil {
		t.Fatalf("parse matrix document: %v", err)
	}
	return cells
}

// The CI gate: real document vs shipped ruleset, all four classes enforced.
func TestCoverageMatrixDocumentMatchesShippedRuleset(t *testing.T) {
	cells := loadRealMatrix(t)
	if len(cells) != len(MatrixAxes.Families)*len(MatrixAxes.Languages) {
		t.Fatalf("matrix parsed %d cells, want %d", len(cells),
			len(MatrixAxes.Families)*len(MatrixAxes.Languages))
	}
	ids, err := LoadRulesetIDs(shippedRulesDir)
	if err != nil {
		t.Fatalf("load shipped ruleset ids: %v", err)
	}
	findings := CheckCoverageMatrix(cells, ids)
	for _, cls := range []FailureClass{
		ClassKeySetMismatch, ClassUnresolvedCell,
		ClassUnevidencedExemption, ClassDanglingRuleID,
	} {
		if f := findingsOfClass(t, findings, cls); len(f) > 0 {
			t.Fatalf("%s findings on the real matrix: %+v", cls, f)
		}
	}

	// Seed-phase bookkeeping: exactly the 14 measured cells are resolved now;
	// the remaining keys stay pending until SPEC-ASTGREP-BREADTH-001 fills them.
	implemented, pending := 0, 0
	docByID := map[string]MatrixCell{}
	for _, c := range cells {
		switch c.State {
		case StateImplemented:
			implemented++
			docByID[c.Key()] = c
		case StatePending:
			pending++
		}
	}
	if implemented != 14 || pending != 98 {
		t.Fatalf("seed accounting drifted: %d IMPLEMENTED / %d PENDING, want 14 / 98", implemented, pending)
	}
	for _, seed := range SeededImplementedCells() {
		got, ok := docByID[seed.Key()]
		if !ok {
			t.Fatalf("seeded cell %s missing from document", seed.Key())
		}
		if got.RuleID != seed.RuleID {
			t.Fatalf("seeded cell %s rule id = %q, want %q", seed.Key(), got.RuleID, seed.RuleID)
		}
		if !ids[got.RuleID] {
			t.Fatalf("seeded cell %s rule id %q absent from shipped ruleset", seed.Key(), got.RuleID)
		}
	}
}

// AC-A16-014 support — the excluded-languages record carries both languages,
// the ast-grep version it was derived under, and the verbatim probe refusals.
func TestExcludedLanguagesRecordedWithVersionAndProbeOutput(t *testing.T) {
	data, err := os.ReadFile(matrixDocPath)
	if err != nil {
		t.Fatalf("read matrix document: %v", err)
	}
	doc := string(data)
	required := []string{
		MatrixAxes.ASTGrepVer,
		"| r |", "| flutter |",
		"sg run -l r --stdin", "r is not supported!",
		"sg run -l flutter --stdin", "flutter is not supported!",
		"equal-priority future additions",
	}
	for _, token := range required {
		if !strings.Contains(doc, token) {
			t.Errorf("excluded-languages record missing required token %q", token)
		}
	}
}

// --- AC-A16-013 through the DOCUMENT path --------------------------------
//
// The four class tests above drive CheckCoverageMatrix directly, which proves
// the classifier but not the gate: the wired CI check reads the markdown
// first, so a defect introduced in the document must still reach the class it
// belongs to. Case (c) is the one that separates the two paths — an EXEMPT row
// carries a rationale in the shared "rule id / rationale" column, and reading
// that column as a rule id turns an unevidenced exemption into a dangling id.

// mutateMatrixDoc returns the real document with one line replaced (or, when
// replacement is empty, deleted). It fails the test if the target is absent,
// so a matrix edit cannot silently void the defect.
func mutateMatrixDoc(t *testing.T, doc, target, replacement string) string {
	t.Helper()
	if !strings.Contains(doc, target) {
		t.Fatalf("fixture target row absent from matrix document: %q", target)
	}
	if replacement == "" {
		return strings.Replace(doc, target+"\n", "", 1)
	}
	return strings.Replace(doc, target, replacement, 1)
}

func checkMatrixDoc(t *testing.T, doc string) []CheckerFinding {
	t.Helper()
	cells, err := ParseCoverageMatrix(doc)
	if err != nil {
		t.Fatalf("parse mutated matrix: %v", err)
	}
	ids, err := LoadRulesetIDs(shippedRulesDir)
	if err != nil {
		t.Fatalf("load shipped ruleset ids: %v", err)
	}
	return CheckCoverageMatrix(cells, ids)
}

func TestCheckerClassesFireThroughDocumentPath(t *testing.T) {
	data, err := os.ReadFile(matrixDocPath)
	if err != nil {
		t.Fatalf("read matrix document: %v", err)
	}
	base := string(data)

	const (
		pendingRow     = "| F3 | ruby | PENDING | - | - |"
		otherPending   = "| F4 | scala | PENDING | - | - |"
		implementedRow = "| F3 | go | IMPLEMENTED | sec-weak-hash-md5 | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |"
	)

	cases := []struct {
		name     string
		doc      string
		want     FailureClass
		mustName string
		// forbid names classes that a correct classifier must NOT emit for
		// this defect; without it a misread column passes by landing in the
		// wrong bucket.
		forbid []FailureClass
	}{
		{
			name:     "(a) deleted cell",
			doc:      mutateMatrixDoc(t, base, pendingRow, ""),
			want:     ClassKeySetMismatch,
			mustName: "F3/ruby",
		},
		{
			name: "(b) substitution at constant count",
			doc: mutateMatrixDoc(t, mutateMatrixDoc(t, base, pendingRow, ""),
				otherPending, otherPending+"\n"+otherPending),
			want:     ClassKeySetMismatch,
			mustName: "F3/ruby",
		},
		{
			name: "(c) bare-assertion exemption",
			doc: mutateMatrixDoc(t, base, pendingRow,
				"| F3 | ruby | EXEMPT | ruby has no weak-hash sink worth flagging | it simply does not come up in practice |"),
			want:     ClassUnevidencedExemption,
			mustName: "F3/ruby",
			forbid:   []FailureClass{ClassDanglingRuleID, ClassUnresolvedCell},
		},
		{
			name: "(d) dangling rule id",
			doc: mutateMatrixDoc(t, base, implementedRow,
				"| F3 | go | IMPLEMENTED | sec-weak-hash-md4 | internal/astgrep/testdata/rule-tests/ (sg-test case pair) |"),
			want:     ClassDanglingRuleID,
			mustName: "F3/go",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkMatrixDoc(t, tc.doc)
			hits := findingsOfClass(t, got, tc.want)
			if len(hits) == 0 {
				t.Fatalf("defect %s produced no %s finding; all findings: %+v", tc.name, tc.want, got)
			}
			named := false
			for _, f := range hits {
				if f.Key == tc.mustName || strings.Contains(f.Detail, tc.mustName) {
					named = true
				}
			}
			if !named {
				t.Fatalf("%s findings do not name %s: %+v", tc.want, tc.mustName, hits)
			}
			for _, bad := range tc.forbid {
				if f := findingsOfClass(t, got, bad); len(f) > 0 {
					t.Fatalf("defect %s misclassified as %s: %+v", tc.name, bad, f)
				}
			}
		})
	}
}

// The EXEMPT row's shared column must land in Rationale, not RuleID — the
// parser-level statement of the same invariant.
func TestParseCoverageMatrixRoutesSharedColumnByState(t *testing.T) {
	doc := `| family | language | state | rule id / rationale | evidence |
|---|---|---|---|---|
| F2 | go | IMPLEMENTED | sec-hardcoded-credential | internal/astgrep/testdata/rule-tests |
| F6 | ruby | EXEMPT | no csrf token surface in this ecosystem | cite: rack docs |
`
	cells, err := ParseCoverageMatrix(doc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	byKey := map[string]MatrixCell{}
	for _, c := range cells {
		byKey[c.Key()] = c
	}
	impl := byKey["F2/go"]
	if impl.RuleID != "sec-hardcoded-credential" || impl.Rationale != "" {
		t.Errorf("IMPLEMENTED cell = %+v; want RuleID set, Rationale empty", impl)
	}
	ex := byKey["F6/ruby"]
	if ex.Rationale != "no csrf token surface in this ecosystem" || ex.RuleID != "" {
		t.Errorf("EXEMPT cell = %+v; want Rationale set, RuleID empty", ex)
	}
}
