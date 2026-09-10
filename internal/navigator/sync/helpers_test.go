package sync

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestItoa exercises the int→string helper including the zero and negative
// paths that the production caller (mx-bridge line numbers) rarely exercises.
func TestItoa(t *testing.T) {
	cases := map[int]string{
		0:   "0",
		1:   "1",
		-1:  "-1",
		42:  "42",
		-42: "-42",
	}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Errorf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}

// TestLastSegment covers the `.` / `/` and bare branches.
func TestLastSegment(t *testing.T) {
	cases := map[string]string{
		"pkg.ParseHeader":         "ParseHeader",
		"internal/x/Foo":          "Foo",
		"Bare":                    "Bare",
		"pkg.sub.More":            "More",
		"path/with/both.and.dots": "dots",
	}
	for in, want := range cases {
		if got := lastSegment(in); got != want {
			t.Errorf("lastSegment(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestDesignDocPaths_IncludesDocsTree exercises the `.moai/docs/**/*.md` walk.
func TestDesignDocPaths_IncludesDocsTree(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, ".moai", "project", "tech.md"), "# Tech\n")
	writeFixture(t, filepath.Join(root, ".moai", "docs", "sub", "a.md"), "# A\n")
	writeFixture(t, filepath.Join(root, ".moai", "docs", "b.md"), "# B\n")
	paths := designDocPaths(root)
	saw := map[string]bool{}
	for _, p := range paths {
		saw[filepath.Base(p)] = true
	}
	for _, want := range []string{"tech.md", "product.md", "structure.md", "a.md", "b.md"} {
		if !saw[want] {
			t.Errorf("designDocPaths missing %q; got %v", want, paths)
		}
	}
}

// TestResolveSymID exercises the D2 exact / suffix / new-node paths.
func TestResolveSymID(t *testing.T) {
	nodes := newNodeSet()
	nodes.add(EntitySymbol, "pkg.ParseHeader", "ParseHeader")
	nodes.add(EntitySymbol, "Foo", "Foo")
	cases := map[string]string{
		"pkg.ParseHeader": "pkg.ParseHeader", // exact
		"ParseHeader":     "ParseHeader",     // suffix
		"brand.New":       "brand.New",       // new node
		"Foo":             "Foo",             // exact
	}
	for in, want := range cases {
		if got := resolveSymID(nodes, in); got != want {
			t.Errorf("resolveSymID(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestDecideTarget covers both branches (decision matches a SPEC ID, or it
// falls back to a self-loop on the decision node).
func TestDecideTarget(t *testing.T) {
	specs := []string{"SPEC-X-001", "SPEC-Y-002"}
	nodes := newNodeSet()
	for _, s := range specs {
		nodes.add(EntitySpec, s, s)
	}
	if tKind, tID := decideTarget(nodes, "SPEC-X-001", specs); tKind != EntitySpec || tID != "SPEC-X-001" {
		t.Errorf("decideTarget(SPEC-X-001) = (%q,%q), want (spec,SPEC-X-001)", tKind, tID)
	}
	if tKind, tID := decideTarget(nodes, "AUTH-STRATEGY", specs); tKind != EntityDecision || tID != "AUTH-STRATEGY" {
		t.Errorf("decideTarget(AUTH-STRATEGY) = (%q,%q), want (decision,AUTH-STRATEGY)", tKind, tID)
	}
}

// TestBuildGraph_SymbolAndSpecNodes exercises buildGraph with empty 003
// output and only @NAV:SYM + @MX:SPEC records (degradation paths).
func TestBuildGraph_SymbolAndSpecNodes(t *testing.T) {
	prov := Provenance{ExtractCommitSHA: "abc", CapturedAt: "2026-01-01"}
	decRecs := []BindingRecord{
		{TokenFamily: FamilyNavDec, Identifier: "AUTH", SourcePath: "tech.md", LineNumber: 1, CommitSHA: "abc"},
	}
	symRecs := []BindingRecord{
		{TokenFamily: FamilyNavSym, Identifier: "pkg.X", SourcePath: "x.go", LineNumber: 2, CommitSHA: "abc"},
	}
	mx := []MxBridgeSpec{
		{SpecID: "SPEC-Z-001", SourcePath: "y.go", LineNumber: 3},
	}
	g := buildGraph(prov, []string{"SPEC-Z-001"}, capabilitySymbolsDoc{}, decRecs, symRecs, mx)
	if len(g.Nodes) == 0 {
		t.Fatalf("no nodes built")
	}
	sawDec, sawSym, sawSpec := false, false, false
	for _, n := range g.Nodes {
		switch n.EntityType {
		case EntityDecision:
			sawDec = true
		case EntitySymbol:
			sawSym = true
		case EntitySpec:
			sawSpec = true
		}
	}
	if !sawDec {
		t.Errorf("decision node missing")
	}
	if !sawSym {
		t.Errorf("symbol node missing")
	}
	if !sawSpec {
		t.Errorf("spec node missing")
	}
	sawDecEdge, sawSpecEdge, sawSymEdge := false, false, false
	for _, e := range g.Edges {
		switch e.EdgeType {
		case EdgeDec:
			sawDecEdge = true
		case EdgeSpec:
			sawSpecEdge = true
		case EdgeSym:
			sawSymEdge = true
		}
	}
	if !sawDecEdge || !sawSpecEdge || !sawSymEdge {
		t.Errorf("missing edge types; dec=%v spec=%v sym=%v", sawDecEdge, sawSpecEdge, sawSymEdge)
	}
}

// TestLoadCapabilitySymbols_Malformed covers the parse-error path that
// degrades to @NAV:SYM-only (REQ-NS-016 graceful degradation).
func TestLoadCapabilitySymbols_Malformed(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "bad.json")
	writeFixture(t, path, "{not-json")
	got := loadCapabilitySymbols(path, filepath.Join(root, "log"))
	if len(got.Rows) != 0 {
		t.Errorf("malformed JSON returned rows: %+v", got)
	}
}

// TestLoadAuditReport_Malformed covers the advisory-read parse-error path.
func TestLoadAuditReport_Malformed(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "bad.json")
	writeFixture(t, path, "{not-json")
	if got := loadAuditReport(path, filepath.Join(root, "log")); got != nil {
		t.Errorf("malformed JSON returned non-nil: %s", got)
	}
}

// TestCapabilityGateError exercises the (*CapabilityGateError).Error() path.
func TestCapabilityGateError(t *testing.T) {
	e := &CapabilityGateError{ProjectRoot: "/tmp/x"}
	if !strings.Contains(e.Error(), "/tmp/x") {
		t.Errorf("Error() missing project root: %q", e.Error())
	}
}

// TestSortMxBridge exercises the deterministic ordering of the mx-bridge slice.
func TestSortMxBridge(t *testing.T) {
	in := []MxBridgeSpec{
		{SpecID: "B", SourcePath: "p", LineNumber: 5},
		{SpecID: "A", SourcePath: "p", LineNumber: 2},
		{SpecID: "A", SourcePath: "p", LineNumber: 1},
	}
	sortMxBridge(in)
	if in[0].SpecID != "A" || in[0].LineNumber != 1 {
		t.Errorf("sortMxBridge order wrong: %+v", in)
	}
}
