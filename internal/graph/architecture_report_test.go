package graph

import (
	"strings"
	"testing"
)

// architecture_report_test.go — SPEC-GRAPH-REPORT-001 M2 unit layer
// (REQ-GR-005..007): the three deterministic report sections computed as
// pure functions over []Edge. Hand-written edge fixtures are the sanctioned
// fixture form here (plan §F M2 test spec): each fixture names the exact
// node universe it needs for the directory proxy.

// godNodesFixture builds the mixed-kind fan-in fixture:
//
//	import layer   internal/cli, internal/hook, internal/spec -> internal/graph
//	               internal/cli, internal/hook               -> internal/mx
//	code-call      two callers of Build (unambiguous callee in internal/graph)
//	               one caller of Scan (unambiguous callee in internal/cli)
//
// Expected: internal/graph fan-in 5 over kinds {import, code-call};
// internal/mx fan-in 2 over {import}; internal/cli fan-in 1 over {code-call}.
func godNodesFixture() []Edge {
	return []Edge{
		{Kind: KindImport, Source: "internal/cli", Target: "internal/graph"},
		{Kind: KindImport, Source: "internal/hook", Target: "internal/graph"},
		{Kind: KindImport, Source: "internal/spec", Target: "internal/graph"},
		{Kind: KindImport, Source: "internal/cli", Target: "internal/mx"},
		{Kind: KindImport, Source: "internal/hook", Target: "internal/mx"},
		// The callee node universe: Build lives only in internal/graph,
		// Scan lives only in internal/cli — both unambiguous.
		{Kind: KindCodeCall, Source: "internal/graph/graph.go:Build", Target: "Write", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/cli/scan.go:Scan", Target: "Write", Resolution: ResolutionInferred, Confidence: 0.85},
		// The scored edges.
		{Kind: KindCodeCall, Source: "internal/cli/main.go:Run", Target: "Build", Resolution: ResolutionExtracted, Confidence: 1.0},
		{Kind: KindCodeCall, Source: "internal/hook/start.go:Boot", Target: "Build", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/graph/check.go:Verify", Target: "Scan", Resolution: ResolutionIntraPackage, Confidence: 0.95},
	}
}

// TestGodNodes_MixedKindAggregationAndKindsLabel covers AC-GR-016's first
// clause: the ranking aggregates over the counted edge kinds and the result
// names exactly those kinds.
func TestGodNodes_MixedKindAggregationAndKindsLabel(t *testing.T) {
	res := GodNodes(godNodesFixture(), 0)

	if got, want := strings.Join(res.Kinds, ", "), "code-call, import"; got != want {
		t.Errorf("kinds label = %q, want %q", got, want)
	}
	if len(res.Nodes) != 3 {
		t.Fatalf("node count = %d, want 3: %+v", len(res.Nodes), res.Nodes)
	}
	// internal/graph: 3 distinct import sources + 2 distinct code-call
	// sources, normalized onto the internal/graph directory by the proxy.
	if n := res.Nodes[0]; n.Node != "internal/graph" || n.FanIn != 5 {
		t.Errorf("top node = {%s %d}, want {internal/graph 5}", n.Node, n.FanIn)
	}
	if n := res.Nodes[1]; n.Node != "internal/mx" || n.FanIn != 2 {
		t.Errorf("second node = {%s %d}, want {internal/mx 2}", n.Node, n.FanIn)
	}
	if n := res.Nodes[2]; n.Node != "internal/cli" || n.FanIn != 1 {
		t.Errorf("third node = {%s %d}, want {internal/cli 1}", n.Node, n.FanIn)
	}
}

// TestGodNodes_AmbiguousCalleeExcludedFromAggregation covers REQ-GR-005's
// directory-proxy disclosure: a bare callee matching 2+ nodes never joins the
// aggregation — it is excluded, never guessed onto one package.
func TestGodNodes_AmbiguousCalleeExcludedFromAggregation(t *testing.T) {
	edges := []Edge{
		// Node universe: "Dup" is declared by two nodes in different packages
		// (ambiguous); "Sibling" by exactly one node in internal/wire.
		{Kind: KindCodeCall, Source: "internal/wire/sib.go:Dup", Target: "Remote", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/far/far.go:Dup", Target: "Remote", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/wire/sib.go:Sibling", Target: "Remote", Resolution: ResolutionInferred, Confidence: 0.85},
		// Scored: the ambiguous target "Dup" is excluded, the unique target
		// "Sibling" joins internal/wire.
		{Kind: KindCodeCall, Source: "internal/wire/wire.go:A", Target: "Dup", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/wire/wire.go:A", Target: "Sibling", Resolution: ResolutionInferred, Confidence: 0.85},
	}
	res := GodNodes(edges, 0)
	if len(res.Nodes) != 1 {
		t.Fatalf("node count = %d, want 1 (only the unambiguous target): %+v", len(res.Nodes), res.Nodes)
	}
	if n := res.Nodes[0]; n.Node != "internal/wire" || n.FanIn != 1 {
		t.Errorf("node = {%s %d}, want {internal/wire 1}", n.Node, n.FanIn)
	}
}

// TestGodNodes_TieBrokenByNodeID covers AC-GR-007: equal fan-in ranks at the
// same tier and the tie resolves by node id ascending.
func TestGodNodes_TieBrokenByNodeID(t *testing.T) {
	edges := []Edge{
		{Kind: KindImport, Source: "internal/zzz", Target: "internal/yyy"},
		{Kind: KindImport, Source: "internal/xxx", Target: "internal/yyy"},
		{Kind: KindImport, Source: "internal/www", Target: "internal/yyy"},
		{Kind: KindImport, Source: "internal/zzz", Target: "internal/aaa"},
		{Kind: KindImport, Source: "internal/xxx", Target: "internal/aaa"},
		{Kind: KindImport, Source: "internal/www", Target: "internal/aaa"},
	}
	res := GodNodes(edges, 0)
	if len(res.Nodes) != 2 {
		t.Fatalf("node count = %d, want 2: %+v", len(res.Nodes), res.Nodes)
	}
	if res.Nodes[0].FanIn != 3 || res.Nodes[1].FanIn != 3 {
		t.Errorf("both tiers must be 3, got %d and %d", res.Nodes[0].FanIn, res.Nodes[1].FanIn)
	}
	if res.Nodes[0].Node != "internal/aaa" || res.Nodes[1].Node != "internal/yyy" {
		t.Errorf("tie order = [%s, %s], want [internal/aaa, internal/yyy] (id ascending)",
			res.Nodes[0].Node, res.Nodes[1].Node)
	}
}

// TestGodNodes_LimitTruncates preserves the --limit contract: limit > 0
// truncates, limit <= 0 returns the full ranking.
func TestGodNodes_LimitTruncates(t *testing.T) {
	edges := godNodesFixture()
	full := GodNodes(edges, 0)
	if got := len(GodNodes(edges, 2).Nodes); got != 2 {
		t.Errorf("limit 2 returned %d nodes, want 2", got)
	}
	if got := len(full.Nodes); got != 3 {
		t.Errorf("limit 0 returned %d nodes, want 3", got)
	}
}

// surprisingFixture builds the AC-GR-008 fixture:
//
//	boundary  internal/cli/main.go:Run -> Boot  (Boot only in internal/hook) INFERRED 0.85
//	boundary  internal/spec/audit.go:Check -> Boot (same callee, second caller) INFERRED 0.5
//	intra     internal/graph/a.go:Alpha -> Beta (Beta only in internal/graph) INFERRED 0.85
//	ambiguous internal/graph/a.go:Alpha -> Dup  (Dup in wire + far) INFERRED 0.85
//	non-inferred boundary edge (extracted) — excluded: only INFERRED edges score
func surprisingFixture() []Edge {
	return []Edge{
		// Node universe: Boot in internal/hook, Beta in internal/graph,
		// Dup in BOTH internal/wire and internal/far (ambiguous).
		{Kind: KindCodeCall, Source: "internal/hook/start.go:Boot", Target: "Scan", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/graph/b.go:Beta", Target: "Alpha", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/wire/sib.go:Dup", Target: "Sibling", Resolution: ResolutionInferred, Confidence: 0.85},
		{Kind: KindCodeCall, Source: "internal/far/far.go:Dup", Target: "Remote", Resolution: ResolutionInferred, Confidence: 0.85},

		// The scored edges.
		{Kind: KindCodeCall, Source: "internal/cli/main.go:Run", Target: "Boot", Resolution: ResolutionInferred, Confidence: 0.85, Line: 12, Grade: "full"},
		{Kind: KindCodeCall, Source: "internal/spec/audit.go:Check", Target: "Boot", Resolution: ResolutionInferred, Confidence: 0.5, Line: 30, Grade: "full"},
		{Kind: KindCodeCall, Source: "internal/graph/a.go:Alpha", Target: "Beta", Resolution: ResolutionInferred, Confidence: 0.85, Line: 7, Grade: "full"},
		{Kind: KindCodeCall, Source: "internal/graph/a.go:Alpha", Target: "Dup", Resolution: ResolutionInferred, Confidence: 0.85, Line: 9, Grade: "full"},
		// Extracted cross-package edge: not INFERRED, never scored (section (2)
		// scores INFERRED resolution — SPEC-GRAPH-EDGE-CONFIDENCE-001's
		// INFERRED tier is the surprising-connections signal).
		{Kind: KindCodeCall, Source: "internal/cli/main.go:Run", Target: "Scan", Resolution: ResolutionExtracted, Confidence: 1.0, Line: 13, Grade: "full"},
	}
}

// TestSurprisingConnections_BoundaryRankedFirstAmbiguousExcluded covers
// AC-GR-008: the boundary-crossing INFERRED edge ranks first (above the
// same-confidence intra-package edge, which the cross-package selector keeps
// out of the section), the ambiguous bare-callee edge is excluded, and the
// cross-package attribution comes from the endpoints' file directories.
func TestSurprisingConnections_BoundaryRankedFirstAmbiguousExcluded(t *testing.T) {
	conns := SurprisingConnections(surprisingFixture())

	if len(conns) != 2 {
		t.Fatalf("connection count = %d, want 2 (two boundary INFERRED edges): %+v", len(conns), conns)
	}
	first := conns[0]
	if first.From != "internal/cli/main.go:Run" || first.To != "internal/hook/start.go:Boot" {
		t.Errorf("first entry = %s -> %s, want internal/cli/main.go:Run -> internal/hook/start.go:Boot",
			first.From, first.To)
	}
	if first.FromPkg != "internal/cli" || first.ToPkg != "internal/hook" {
		t.Errorf("package attribution = %s -> %s, want internal/cli -> internal/hook (directory proxy)",
			first.FromPkg, first.ToPkg)
	}
	if first.Confidence != 0.85 {
		t.Errorf("first entry confidence = %v, want 0.85 (confidence-desc ranking)", first.Confidence)
	}
	second := conns[1]
	if second.From != "internal/spec/audit.go:Check" || second.Confidence != 0.5 {
		t.Errorf("second entry = %s (conf %v), want internal/spec/audit.go:Check (conf 0.5)",
			second.From, second.Confidence)
	}
	for _, c := range conns {
		if c.To == "" || strings.HasSuffix(c.To, "Dup") {
			t.Errorf("ambiguous callee reached the section: %+v", c)
		}
		if c.FromPkg == c.ToPkg {
			t.Errorf("intra-package edge reached the section: %+v", c)
		}
		if c.Confidence == 1.0 {
			t.Errorf("extracted (non-INFERRED) edge reached the section: %+v", c)
		}
	}
}

// importCycleFixture builds the AC-GR-009 import layer:
//
//	simple cycle  pkg/a -> pkg/c -> pkg/b -> pkg/a   (rotation from a: a,c,b)
//	branched SCC  b2/{A,B,C}: A<->C, B<->C           (no simple cycle)
//	self loop     pkg/self -> pkg/self
//	island        pkg/x -> pkg/leaf (acyclic, never reported)
func importCycleFixture() []Edge {
	return []Edge{
		{Kind: KindImport, Source: "pkg/a", Target: "pkg/c"},
		{Kind: KindImport, Source: "pkg/c", Target: "pkg/b"},
		{Kind: KindImport, Source: "pkg/b", Target: "pkg/a"},

		{Kind: KindImport, Source: "b2/A", Target: "b2/C"},
		{Kind: KindImport, Source: "b2/C", Target: "b2/A"},
		{Kind: KindImport, Source: "b2/B", Target: "b2/C"},
		{Kind: KindImport, Source: "b2/C", Target: "b2/B"},

		{Kind: KindImport, Source: "pkg/self", Target: "pkg/self"},

		{Kind: KindImport, Source: "pkg/x", Target: "pkg/leaf"},
	}
}

// TestImportCycles_SCCShapeAndCanonicalRotation covers AC-GR-009: cycles are
// reported as SCCs, the simple cycle renders its canonical rotation (smallest
// member first, following edge direction — a,c,b here, NOT the sorted list),
// the branched SCC renders its member list, and the SCC count is what the
// section counts.
func TestImportCycles_SCCShapeAndCanonicalRotation(t *testing.T) {
	sccs := ImportCycles(importCycleFixture())

	if len(sccs) != 3 {
		t.Fatalf("SCC count = %d, want 3 (simple cycle + branched SCC + self loop): %+v", len(sccs), sccs)
	}

	byFirst := map[string]ImportSCC{}
	for _, s := range sccs {
		byFirst[s.Members[0]] = s
	}

	simple, ok := byFirst["pkg/a"]
	if !ok {
		t.Fatalf("no SCC led by pkg/a: %+v", sccs)
	}
	if !simple.SimpleCycle() {
		t.Errorf("pkg/a SCC must be a simple cycle: %+v", simple)
	}
	if got, want := strings.Join(simple.Rotation, ","), "pkg/a,pkg/c,pkg/b"; got != want {
		t.Errorf("rotation = %s, want %s (smallest member first, following edges)", got, want)
	}
	if got, want := strings.Join(simple.Members, ","), "pkg/a,pkg/b,pkg/c"; got != want {
		t.Errorf("members = %s, want %s (sorted canonical member list)", got, want)
	}

	branched, ok := byFirst["b2/A"]
	if !ok {
		t.Fatalf("no SCC led by b2/A: %+v", sccs)
	}
	if branched.SimpleCycle() {
		t.Errorf("branched SCC must not claim a simple cycle: %+v", branched)
	}
	if len(branched.Rotation) != 0 {
		t.Errorf("branched SCC must render no rotation, got %v", branched.Rotation)
	}
	if got, want := strings.Join(branched.Members, ","), "b2/A,b2/B,b2/C"; got != want {
		t.Errorf("branched members = %s, want %s", got, want)
	}

	if _, ok := byFirst["pkg/self"]; !ok {
		t.Errorf("self-loop package must be reported as a one-member cycle: %+v", sccs)
	}
	for _, s := range sccs {
		for _, m := range s.Members {
			if m == "pkg/x" || m == "pkg/leaf" {
				t.Errorf("acyclic package reached the cycle section: %+v", s)
			}
		}
	}
}

// TestRenderArchitectureReport_SectionsAndKindsLine covers AC-GR-006's
// content clause at the pure-function layer: the three section headings and
// the fan-in kinds line (AC-GR-016).
func TestRenderArchitectureReport_SectionsAndKindsLine(t *testing.T) {
	edges := append(godNodesFixture(), surprisingFixture()...)
	body := RenderArchitectureReport(edges, 0)
	for _, want := range []string{
		"## God Nodes",
		"## Surprising Connections",
		"## Import Cycles",
		"fan-in over: code-call, import",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("report missing %q; body:\n%s", want, body)
		}
	}
	// The scored boundary edge renders; the ambiguous callee never does.
	if !strings.Contains(body, "internal/cli/main.go:Run -> internal/hook/start.go:Boot") {
		t.Errorf("report missing the boundary edge row; body:\n%s", body)
	}
	if strings.Contains(body, ":Dup") {
		t.Errorf("ambiguous callee rendered into the report; body:\n%s", body)
	}
}

// TestRenderArchitectureReport_Deterministic covers REQ-GR-005's determinism
// clause at the pure-function layer: two renders over the same edges are
// byte-identical (the CLI-level double-run cmp is AC-GR-011's binding form).
func TestRenderArchitectureReport_Deterministic(t *testing.T) {
	edges := append(importCycleFixture(), surprisingFixture()...)
	a := RenderArchitectureReport(edges, 10)
	b := RenderArchitectureReport(edges, 10)
	if a != b {
		t.Errorf("two renders over the same edges differ:\n--- A ---\n%s\n--- B ---\n%s", a, b)
	}
}

// TestRenderArchitectureReport_EmptySectionsStateReasons covers REQ-GR-006:
// empty sections stay present with a stated reason — the code-layer-absent
// shape for a doc-only artifact and the no-cycles shape for an acyclic one.
func TestRenderArchitectureReport_EmptySectionsStateReasons(t *testing.T) {
	docOnly := []Edge{
		{Kind: KindMXSpec, Source: "internal/demo/demo.go", Target: "SPEC-X-001", Line: 4},
		{Kind: KindImport, Source: "pkg/a", Target: "pkg/b"},
	}
	body := RenderArchitectureReport(docOnly, 0)
	for _, want := range []string{
		"## God Nodes",
		"## Surprising Connections",
		"## Import Cycles",
		"code layer absent: CGO disabled or no extraction",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("doc-only report missing %q; body:\n%s", want, body)
		}
	}
	// Import layer is present and acyclic: the reason must say so, and the
	// god-nodes ranking still answers from the import layer.
	if !strings.Contains(body, "no cycles") && !strings.Contains(body, "carries no cycles") {
		t.Errorf("import-cycles section must state the no-cycles reason; body:\n%s", body)
	}
	if !strings.Contains(body, "pkg/b") {
		t.Errorf("god nodes must rank the import-layer target; body:\n%s", body)
	}
}

// TestRenderArchitectureReport_NoEdgesAtAll is the fully-empty shape: all
// three sections present, each with its stated reason — never a missing file.
func TestRenderArchitectureReport_NoEdgesAtAll(t *testing.T) {
	body := RenderArchitectureReport(nil, 0)
	for _, want := range []string{
		"## God Nodes",
		"## Surprising Connections",
		"## Import Cycles",
		"code layer absent: CGO disabled or no extraction",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("empty-edges report missing %q; body:\n%s", want, body)
		}
	}
}
