package graph

import (
	"strings"
	"testing"
)

// M1 confidence tests (SPEC-GRAPH-EDGE-CONFIDENCE-001): the tier matrix over
// a t.TempDir() fixture tree, plus the ConfidenceFor single-definition map.

// codeCallBy returns the code-call edge whose source FILE (the part before
// the ":Caller" join) ends with fileSuffix and whose target is callee, or
// nil. A caller-less top-level edge's source is the bare file.
func codeCallBy(edges []Edge, fileSuffix, callee string) *Edge {
	for i := range edges {
		e := &edges[i]
		if e.Kind != KindCodeCall || e.Target != callee {
			continue
		}
		file := e.Source
		if idx := strings.LastIndex(file, ":"); idx > 0 {
			file = file[:idx]
		}
		if strings.HasSuffix(file, fileSuffix) {
			return e
		}
	}
	return nil
}

// ConfidenceFor is the SINGLE definition of the resolution→confidence map
// (REQ-GEC-001): total over the closed 3-value domain; anything outside it
// maps to 0 (absent — omitempty drops it on legacy/unknown labels).
func TestConfidenceFor(t *testing.T) {
	cases := []struct {
		res  string
		want float64
	}{
		{ResolutionExtracted, 1.0},
		{ResolutionIntraPackage, 0.95},
		{ResolutionInferred, 0.85},
		{"", 0},
		{"ambiguous", 0},
	}
	for _, tc := range cases {
		if got := ConfidenceFor(tc.res); got != tc.want {
			t.Errorf("ConfidenceFor(%q) = %v, want %v", tc.res, got, tc.want)
		}
	}
}

// AC-GEC-002 — every code-call edge carries resolution ∈ the closed 3-value
// domain and confidence ≡ ConfidenceFor(resolution); all three values occur
// on the fixture; doc-derived and code-import edges carry none.
func TestCodeCallConfidenceDomain(t *testing.T) {
	requireCodeExtraction(t)
	root := tierFixture(t)
	edges, _, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges: %v", err)
	}

	domain := map[string]bool{ResolutionExtracted: true, ResolutionIntraPackage: true, ResolutionInferred: true}
	seen := map[string]bool{}
	codeCalls := 0
	for _, e := range edges {
		if e.Kind != KindCodeCall {
			if e.Resolution != "" || e.Confidence != 0 {
				t.Errorf("non-code-call edge %s/%s carries confidence state: %q %v", e.Kind, e.Target, e.Resolution, e.Confidence)
			}
			continue
		}
		codeCalls++
		if !domain[e.Resolution] {
			t.Errorf("code-call edge %s→%s carries resolution %q outside the closed domain", e.Source, e.Target, e.Resolution)
			continue
		}
		if want := ConfidenceFor(e.Resolution); e.Confidence != want {
			t.Errorf("code-call edge %s→%s confidence %v ≠ ConfidenceFor(%q) = %v", e.Source, e.Target, e.Confidence, e.Resolution, want)
		}
		seen[e.Resolution] = true
	}
	if codeCalls == 0 {
		t.Fatal("fixture produced no code-call edges — nothing was swept")
	}
	for _, res := range []string{ResolutionExtracted, ResolutionIntraPackage, ResolutionInferred} {
		if !seen[res] {
			t.Errorf("fixture must exercise every tier; %q never observed", res)
		}
	}
}

// AC-GEC-003 — same-file declaration evidence promotes to extracted/1.0.
func TestSameFilePromotion(t *testing.T) {
	requireCodeExtraction(t)
	root := tierFixture(t)
	edges, _, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges: %v", err)
	}
	e := codeCallBy(edges, "wire.go", "B")
	if e == nil {
		t.Fatal("wire.go:A → B edge missing")
	}
	if e.Resolution != ResolutionExtracted || e.Confidence != 1.0 {
		t.Errorf("same-file edge A→B = %q/%v, want extracted/1.0", e.Resolution, e.Confidence)
	}
}

// AC-GEC-004 — Go-module import evidence promotes cross-file calls to
// extracted/1.0 (the ONLY T2 domain). Shared and the method Do are declared
// only in the imported internal/helper; Dup is declared in the imported
// module AND a same-dir sibling — T2 outranks T3 (declared dependency beats
// co-location, acceptance.md §D.2).
func TestImportEvidencePromotion(t *testing.T) {
	requireCodeExtraction(t)
	root := tierFixture(t)
	edges, _, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges: %v", err)
	}
	for _, callee := range []string{"Shared", "Dup", "Do"} {
		e := codeCallBy(edges, "wire.go", callee)
		if e == nil {
			t.Errorf("wire.go:A → %s edge missing", callee)
			continue
		}
		if e.Resolution != ResolutionExtracted || e.Confidence != 1.0 {
			t.Errorf("import-evidence edge A→%s = %q/%v, want extracted/1.0", callee, e.Resolution, e.Confidence)
		}
	}
}

// AC-GEC-005 — name-only fallback: a callee declared only in a module the
// caller does not import and outside its own directory resolves inferred/
// 0.85. Covers the Go case (Remote), the undeclared-anywhere case
// (Nowhere), and the NON-GO case (repair-round D1): a python caller's name
// declared in a different directory via a specifier T2 can never resolve —
// never promoted, always the fallback.
func TestNameOnlyFallback(t *testing.T) {
	requireCodeExtraction(t)
	root := tierFixture(t)
	edges, _, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges: %v", err)
	}
	cases := []struct {
		file, callee, why string
	}{
		{"wire.go", "Remote", "declared in a non-imported, non-sibling Go package"},
		{"wire.go", "Nowhere", "declared nowhere in the tree"},
		{"app.py", "helper_fn", "non-Go specifier: declaring dir not derivable — T2 never fires"},
	}
	for _, tc := range cases {
		e := codeCallBy(edges, tc.file, tc.callee)
		if e == nil {
			t.Errorf("%s → %s edge missing (%s)", tc.file, tc.callee, tc.why)
			continue
		}
		if e.Resolution != ResolutionInferred || e.Confidence != 0.85 {
			t.Errorf("fallback edge %s→%s = %q/%v, want inferred/0.85 (%s)", tc.file, tc.callee, e.Resolution, e.Confidence, tc.why)
		}
	}
}

// AC-GEC-012 — a cross-file callee declared in a sibling file of the same
// directory (Go: same package by construction) resolves intra-package/0.95.
// Both directions covered: caller imports a module but the callee lives in
// the sibling (wire.go:A → Sibling), and a zero-import caller reaches a
// sibling's declaration (sib.go:S2 → B).
func TestSamePackagePromotion(t *testing.T) {
	requireCodeExtraction(t)
	root := tierFixture(t)
	edges, _, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges: %v", err)
	}
	for _, tc := range []struct{ file, callee string }{{"wire.go", "Sibling"}, {"sib.go", "B"}} {
		e := codeCallBy(edges, tc.file, tc.callee)
		if e == nil {
			t.Errorf("%s → %s edge missing", tc.file, tc.callee)
			continue
		}
		if e.Resolution != ResolutionIntraPackage || e.Confidence != 0.95 {
			t.Errorf("same-package edge %s→%s = %q/%v, want intra-package/0.95", tc.file, tc.callee, e.Resolution, e.Confidence)
		}
	}
}
