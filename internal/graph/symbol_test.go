package graph

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// codeFixture builds a tree with an UNDOCUMENTED call A→B (no codemaps
// dependencies.md mentions it) and returns the root.
func codeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "wire"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package wire

func A() {
	B()
}

func B() {
	C()
}

func C() {}
`
	if err := os.WriteFile(filepath.Join(root, "internal", "wire", "wire.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func fmtInt(i int) string { return strconv.Itoa(i) }

// findCodeEdge scans edges for kind+source+target containment.
func hasCodeEdge(edges []Edge, kind, src, tgt string) bool {
	for _, e := range edges {
		if e.Kind == kind && strings.Contains(e.Source, src) && e.Target == tgt {
			return true
		}
	}
	return false
}

// AC-GF-017 — an UNDOCUMENTED A→B call appears as a code-call edge, and a
// blast-radius query at B reaches A through it.
func TestCodeEdges_UndocumentedCallAppears(t *testing.T) {
	requireCodeExtraction(t)
	root := codeFixture(t)
	edges, matrix, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges: %v", err)
	}

	if !hasCodeEdge(edges, KindCodeCall, "wire.go:A", "B") {
		t.Errorf("code-call edge A→B missing; edges=%v", edges)
	}
	if !hasCodeEdge(edges, KindCodeCall, "wire.go:B", "C") {
		t.Errorf("code-call edge B→C missing; edges=%v", edges)
	}
	// Every code-call edge carries the resolution grade used to derive it.
	for _, e := range edges {
		if e.Kind == KindCodeCall && e.Grade == "" {
			t.Errorf("code-call edge %+v lacks its resolution grade", e)
		}
	}

	// Blast radius at B reaches A (via reverse traversal of code-call edges).
	blast := BlastRadius(edges, "B")
	hit := false
	for _, n := range blast {
		if strings.HasSuffix(n, ":A") {
			hit = true
		}
	}
	if !hit {
		t.Errorf("A missing from B's blast radius: %v", blast)
	}

	// Matrix published with the go cell graded (name-based at minimum).
	if matrix["go"] != astx.GradeNameBased {
		t.Errorf("matrix[go] = %q, want %q", matrix["go"], astx.GradeNameBased)
	}
}

// docCodeFixture builds a tree carrying all three doc-derived inputs:
// codemaps dependencies.md (import edge), @MX:SPEC sub-line (mx-spec edge),
// and a spec.md depends_on (spec-depends edge) — plus a Go source whose call
// is NOT in the doc layer.
func docCodeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "project", "codemaps"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "dependencies.md"),
		[]byte("```mermaid\ngraph TD\n    a[\"internal/alpha\"] --> b[\"internal/beta\"]\n```\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".moai", "specs", "SPEC-X-001"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "specs", "SPEC-X-001", "spec.md"),
		[]byte("---\nid: SPEC-X-001\ntitle: \"t\"\ndepends_on: [SPEC-X-DEP-001]\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "demo", "demo.go"),
		[]byte("package demo\n\nimport (\n\t\"example.com/proj/internal/undoc\"\n\t\"fmt\"\n)\n\n// @MX:NOTE: [AUTO] demo\n// @MX:SPEC:SPEC-X-001\nfunc Demo() { Hidden(); undoc.Use(); fmt.Println(1) }\n\nfunc Hidden() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// go.mod makes the internal import LOCAL (module-path normalization) —
	// the import the doc layer never records, i.e. the suppressed direction.
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/proj\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "undoc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "undoc", "undoc.go"), []byte("package undoc\n\nfunc Use() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// AC-GF-018 — additivity: the doc-derived edge set is byte-preserved when
// the code layers join the build. E_doc ⊆ E_out unchanged.
func TestBuild_CodeLayersAreAdditive(t *testing.T) {
	requireCodeExtraction(t)
	root := docCodeFixture(t)

	docEdges, err := Build(root)
	if err != nil {
		t.Fatalf("Build (doc layers): %v", err)
	}
	if len(docEdges) == 0 {
		t.Fatal("fixture must produce doc edges (import/mx-spec/spec-depends)")
	}

	allEdges, _, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}

	// Every doc edge present with its RELATIONSHIP fields unchanged
	// (kind/source/target/line). The disagrees_with marker is the one
	// deliberate addition REQ-GF-015 mandates — asserted separately below.
	idx := map[string]int{}
	for _, e := range allEdges {
		idx[e.Kind+"\x00"+e.Source+"\x00"+e.Target+"\x00"+fmtInt(e.Line)]++
	}
	for _, want := range docEdges {
		k := want.Kind + "\x00" + want.Source + "\x00" + want.Target + "\x00" + fmtInt(want.Line)
		if idx[k] == 0 {
			t.Errorf("doc edge dropped or mutated by code layers: %+v", want)
		}
	}
	// Code layers add new kinds.
	kinds := map[string]bool{}
	for _, e := range allEdges {
		kinds[e.Kind] = true
	}
	if !kinds[KindCodeCall] && !kinds[KindCodeImport] {
		// The fixture's demo.go carries calls (mx fixture has none beyond
		// package decl), so code-import from go imports is the expected floor.
		t.Errorf("expected code-derived kinds in the build output; kinds=%v", kinds)
	}
}

// Revival path (lead ruling): the suppressed code-found/doc-silent direction
// stays RETRIEVABLE — DisagreementAll marks it with an explicit [revived]
// tag, and the default mode's genuine refutation markers are unaffected.
func TestBuild_DisagreementAllRevivesSuppressedDirection(t *testing.T) {
	requireCodeExtraction(t)
	root := docCodeFixture(t) // doc has internal/alpha→internal/beta; demo.go imports nothing internal

	refuteEdges, _, _, err := BuildWithCodeLayersMode(root, DisagreementRefuteOnly)
	if err != nil {
		t.Fatal(err)
	}
	allEdges, _, _, err := BuildWithCodeLayersMode(root, DisagreementAll)
	if err != nil {
		t.Fatal(err)
	}

	revived := 0
	for _, e := range allEdges {
		if e.DisagreesWith != "" && strings.Contains(e.DisagreesWith, "[revived]") {
			revived++
			if e.Kind != KindCodeImport {
				t.Errorf("[revived] marker on a non-code-import edge: %+v", e)
			}
		}
	}
	if revived == 0 {
		t.Error("DisagreementAll must revive at least one suppressed code-import marker")
	}
	for _, e := range refuteEdges {
		if strings.Contains(e.DisagreesWith, "[revived]") {
			t.Error("default mode must not emit the revived marker")
		}
	}

	// Doc-relationship preservation holds in BOTH modes.
	for _, want := range refuteEdges {
		if want.Kind != KindImport {
			continue
		}
		found := false
		for _, got := range allEdges {
			if got.Kind == want.Kind && got.Source == want.Source && got.Target == want.Target {
				found = true
			}
		}
		if !found {
			t.Errorf("doc edge lost under DisagreementAll: %+v", want)
		}
	}
}

// AC-GF-019 — a matrix with one grade removed is a DEFECT verdict naming the
// language; an omitted cell may not pass silently.
func TestGradeMatrixDefect_MissingCellIsReported(t *testing.T) {
	full := GradeMatrix()
	if len(full) != len(astx.SupportedLanguages()) {
		t.Fatalf("matrix size = %d, want %d", len(full), len(astx.SupportedLanguages()))
	}

	// Remove one cell and validate: the defect must name the language.
	broken := map[string]string{}
	for k, v := range full {
		broken[k] = v
	}
	delete(broken, "swift")
	defects := ValidateGradeMatrix(broken)
	if len(defects) == 0 {
		t.Fatal("a gradeless cell must produce a defect verdict")
	}
	if !strings.Contains(defects[0], "swift") {
		t.Errorf("defect must name the gradeless language, got: %v", defects)
	}
	if len(ValidateGradeMatrix(full)) != 0 {
		t.Errorf("full matrix must validate clean: %v", ValidateGradeMatrix(full))
	}
}

// code-import edges: an import in described sources yields a file→module edge.
func TestCodeEdges_ImportEdges(t *testing.T) {
	requireCodeExtraction(t)
	root := codeFixture(t)
	if err := os.WriteFile(filepath.Join(root, "internal", "wire", "imp.go"),
		[]byte("package wire\n\nimport \"fmt\"\n\nfunc D() { fmt.Println(1) }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edges, _, err := CodeEdges(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range edges {
		if e.Kind == KindCodeImport && e.Target == "fmt" && strings.Contains(e.Source, "imp.go") {
			found = true
		}
	}
	if !found {
		t.Errorf("code-import edge →fmt missing; edges=%v", edges)
	}
}

// Determinism: two runs on the same tree produce identical sorted output.
func TestCodeEdges_Deterministic(t *testing.T) {
	root := codeFixture(t)
	a, _, err := CodeEdges(root)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := CodeEdges(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("nondeterministic edge count: %d vs %d", len(a), len(b))
	}
	sort.Slice(a, func(i, j int) bool { return EdgeLess(a[i], a[j]) })
	sort.Slice(b, func(i, j int) bool { return EdgeLess(b[i], b[j]) })
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("edge %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}
