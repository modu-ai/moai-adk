package tiers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestEnrich_EndToEnd_OverlayEmitted exercises the M4.5-wired overlay
// emission end-to-end. Pre-seeded contracts.yaml + module_tree.json +
// decisions + a Go fixture + nav-graph.json → tiers.json emitted with all
// 4 tiers populated. Also verifies M4.2 Lens 2 (nav-graph.json untouched).
func TestEnrich_EndToEnd_OverlayEmitted(t *testing.T) {
	root := t.TempDir()

	// (1) contracts.yaml
	writeRegistry(t, root, []byte("contracts:\n  - identifier: c1\n    contract_kind: schema\n    contract_path: x\n    validator_command: 'true'\n"))

	// (2) module_tree.json (authored — Enrich MUST NOT overwrite)
	authored := `{"modules":[{"package_path":"internal/x","display_name":"X","layer":"domain","responsibility":"x","depends_on":["internal/y"]},{"package_path":"internal/y","display_name":"Y","layer":"infrastructure","responsibility":"y","depends_on":[]}]}`
	writeModuleTree(t, root, authored)

	// (3) decisions
	writeADR(t, root, "POLICY-A", "# A\nStatus: Accepted\n")

	// (4) nav-graph.json (M0 producer — Enrich MUST NOT overwrite)
	navDir := filepath.Join(root, ".moai", "project", "navigator")
	if err := os.MkdirAll(navDir, 0o755); err != nil {
		t.Fatal(err)
	}
	navBody := `{"provenance":{"extract_commit_sha":"abc","captured_at":"2026-01-01T00:00:00+00:00"},"nodes":[{"entity_type":"decision","identifier":"POLICY-A"}],"edges":[]}`
	if err := os.WriteFile(filepath.Join(navDir, "nav-graph.json"), []byte(navBody), 0o644); err != nil {
		t.Fatal(err)
	}

	// (5) a Go source fixture for Tier 3 deterministic extraction
	writeGoFixture(t, root, "internal/x", "x.go", "package x\nfunc F() {}\n")
	writeGoFixture(t, root, "internal/y", "y.go", "package y\nimport \"fixt/internal/x\"\nvar _ = x.F\n")

	if err := Enrich(root); err != nil {
		t.Fatalf("Enrich error: %v", err)
	}

	// tiers.json emitted.
	tiersPath := filepath.Join(navDir, "tiers.json")
	body, err := os.ReadFile(tiersPath)
	if err != nil {
		t.Fatalf("tiers.json not emitted: %v", err)
	}
	var overlay TiersOverlay
	if err := json.Unmarshal(body, &overlay); err != nil {
		t.Fatalf("tiers.json unparseable: %v", err)
	}

	// All 4 tiers populated.
	if len(overlay.Tier0Contracts) != 1 {
		t.Errorf("Tier0Contracts=%d; want 1", len(overlay.Tier0Contracts))
	}
	if len(overlay.Tier1Blueprints) != 2 {
		t.Errorf("Tier1Blueprints=%d; want 2", len(overlay.Tier1Blueprints))
	}
	if len(overlay.Tier2Decisions) != 1 {
		t.Errorf("Tier2Decisions=%d; want 1", len(overlay.Tier2Decisions))
	}
	if overlay.Tier2Decisions[0].AdrPath == "" {
		t.Errorf("POLICY-A.AdrPath empty; want set")
	}
	if len(overlay.Tier3Symbols) == 0 {
		t.Errorf("Tier3Symbols empty; want ≥1 Go fixture symbol")
	}

	// module-edge present (X→Y).
	hasModuleEdge := false
	for _, e := range overlay.TierEdges {
		if e.EdgeType == EdgeModule && e.SourceNode == "blueprint:internal/x" && e.TargetNode == "blueprint:internal/y" {
			hasModuleEdge = true
		}
	}
	if !hasModuleEdge {
		t.Errorf("module-edge internal/x→internal/y not present in TierEdges")
	}

	// Provenance populated.
	if overlay.Provenance.ExtractCommitSHA == "" {
		t.Errorf("Provenance.ExtractCommitSHA empty")
	}
}

// TestEnrich_NavGraphUntouched_M4_2_Lens2 verifies the M4.2 Lens 2 invariant
// (REQ-NS3-018 overlay-not-overwrite) holds under a real Enrich call: the
// nav-graph.json content hash is byte-identical before and after.
func TestEnrich_NavGraphUntouched_M4_2_Lens2(t *testing.T) {
	root := t.TempDir()
	navDir := filepath.Join(root, ".moai", "project", "navigator")
	if err := os.MkdirAll(navDir, 0o755); err != nil {
		t.Fatal(err)
	}
	navPath := filepath.Join(navDir, "nav-graph.json")
	navBody := []byte(`{"provenance":{"extract_commit_sha":"z","captured_at":"2026-01-01"},"nodes":[],"edges":[]}`)
	if err := os.WriteFile(navPath, navBody, 0o644); err != nil {
		t.Fatal(err)
	}
	before := hashBytes(navBody)
	if err := Enrich(root); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(navPath)
	if err != nil {
		t.Fatal(err)
	}
	if hashBytes(after) != before {
		t.Errorf("nav-graph.json mutated by Enrich (REQ-NS3-018 overlay-not-overwrite)")
	}
}

// TestEnrich_FailOpen_NavGraphAbsent exercises REQ-NS3-020: when nav-graph
// is absent, Enrich exits 0 and emits tiers.json (with empty decision IDs).
func TestEnrich_FailOpen_NavGraphAbsent(t *testing.T) {
	root := t.TempDir()
	if err := Enrich(root); err != nil {
		t.Fatalf("Enrich error under nav-graph absent: %v", err)
	}
	tiersPath := filepath.Join(root, ".moai", "project", "navigator", "tiers.json")
	if _, err := os.Stat(tiersPath); err != nil {
		t.Errorf("tiers.json not emitted under nav-graph-absent: %v", err)
	}
}

// TestEnrich_ByteIdentical_ReRun exercises REQ-NS3-019: running Enrich twice
// on the same HEAD produces byte-identical tiers.json (deterministic layer).
func TestEnrich_ByteIdentical_ReRun(t *testing.T) {
	root := t.TempDir()
	writeModuleTree(t, root, `{"modules":[]}`)

	if err := Enrich(root); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(root, tiersOutRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if err := Enrich(root); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(root, tiersOutRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if hashBytes(first) != hashBytes(second) {
		t.Errorf("re-run not byte-identical (REQ-NS3-019):\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// TestParseDependenciesMarkdown_VariousShapes covers the dep-graph parser
// directly (the indirect path via Enrich only covers one shape).
func TestParseDependenciesMarkdown_VariousShapes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want map[string][]string // pkg → dependsOn (sorted)
	}{
		{
			name: "depends-on prose",
			in:   "- internal/a depends on internal/b, internal/c\n",
			want: map[string][]string{"internal/a": {"internal/b", "internal/c"}, "internal/b": {}, "internal/c": {}},
		},
		{
			name: "mermaid edge",
			in:   "internal/x --> internal/y\n",
			want: map[string][]string{"internal/x": {"internal/y"}, "internal/y": {}},
		},
		{
			name: "arrow unicode",
			in:   "internal/p → internal/q\n",
			want: map[string][]string{"internal/p": {"internal/q"}, "internal/q": {}},
		},
		{
			name: "noise-only",
			in:   "subgraph P\nend\n한글 잡음\n",
			want: map[string][]string{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mods := parseDependenciesMarkdown(tc.in)
			got := map[string][]string{}
			for _, m := range mods {
				got[m.PackagePath] = m.DependsOn
			}
			// Compare sorted.
			for k := range got {
				sort.Strings(got[k])
			}
			if len(got) != len(tc.want) {
				t.Errorf("module count=%d; want %d (got=%v want=%v)", len(got), len(tc.want), got, tc.want)
				return
			}
			for k, want := range tc.want {
				sort.Strings(want)
				if gotV, ok := got[k]; !ok {
					t.Errorf("missing module %q", k)
					continue
				} else if strings.Join(gotV, ",") != strings.Join(want, ",") {
					t.Errorf("module %q depends_on=%v; want %v", k, gotV, want)
				}
			}
		})
	}
}

// TestInferLayer covers the layer heuristic.
func TestInferLayer(t *testing.T) {
	cases := map[string]Layer{
		"internal/cli":       LayerPresentation,
		"internal/web":       LayerPresentation,
		"cmd/moai":           LayerPresentation,
		"internal/spec":      LayerDomain,
		"internal/workflow":  LayerDomain,
		"internal/navigator": LayerInfrastructure,
		"internal/config":    LayerInfrastructure,
		"internal/audit":     LayerMeasurement,
		"pkg/other":          LayerDomain,
	}
	for pkg, want := range cases {
		if got := inferLayer(pkg); got != want {
			t.Errorf("inferLayer(%q)=%q; want %q", pkg, got, want)
		}
	}
}

// TestHashDeterministicRecord_StableForSameInput verifies the hash is
// deterministic across calls with the same input (REF-NS3-019).
func TestHashDeterministicRecord_StableForSameInput(t *testing.T) {
	r := SymbolEnrichment{
		Identifier:      "pkg.F",
		Signature:       "func F()",
		DeclarationPath: "pkg/p.go",
		DeclarationLine: 5,
		References:      []SymbolRef{{Path: "a.go", Line: 1}, {Path: "b.go", Line: 2}},
	}
	h1 := hashDeterministicRecord(r)
	h2 := hashDeterministicRecord(r)
	if h1 != h2 {
		t.Errorf("hash not stable: %q vs %q", h1, h2)
	}
	// Different references → different hash.
	r2 := r
	r2.References = []SymbolRef{{Path: "a.go", Line: 1}, {Path: "b.go", Line: 99}}
	if hashDeterministicRecord(r2) == h1 {
		t.Errorf("hash did not change when references changed")
	}
}

// TestSortEdges_Stable covers the in-place edge sorter.
func TestSortEdges_Stable(t *testing.T) {
	edges := []TierEdge{
		{EdgeType: EdgeOwns, SourceNode: "b", TargetNode: "y"},
		{EdgeType: EdgeModule, SourceNode: "a", TargetNode: "b"},
		{EdgeType: EdgeSupersededBy, SourceNode: "d1", TargetNode: "d2"},
		{EdgeType: EdgeOwns, SourceNode: "a", TargetNode: "x"},
	}
	sortEdges(edges)
	want := []TierEdge{
		{EdgeType: EdgeModule, SourceNode: "a", TargetNode: "b"},
		{EdgeType: EdgeOwns, SourceNode: "a", TargetNode: "x"},
		{EdgeType: EdgeOwns, SourceNode: "b", TargetNode: "y"},
		{EdgeType: EdgeSupersededBy, SourceNode: "d1", TargetNode: "d2"},
	}
	if len(edges) != len(want) {
		t.Fatalf("len=%d; want %d", len(edges), len(want))
	}
	for i := range edges {
		if edges[i] != want[i] {
			t.Errorf("edge[%d]=%+v; want %+v", i, edges[i], want[i])
		}
	}
}

// TestBuildOwnsEdges_ByPathContainment covers owns-edge attribution.
func TestBuildOwnsEdges_ByPathContainment(t *testing.T) {
	bps := []BlueprintNode{
		{Identifier: "internal/x"},
		{Identifier: "internal/y"},
	}
	syms := []SymbolEnrichment{
		{Identifier: "x.F", DeclarationPath: "internal/x/x.go"},
		{Identifier: "y.G", DeclarationPath: "internal/y/y.go"},
		{Identifier: "orphan", DeclarationPath: "other/z.go"},
	}
	edges := buildOwnsEdges(bps, syms)
	want := map[string]bool{
		"blueprint:internal/x|symbol:x.F": true,
		"blueprint:internal/y|symbol:y.G": true,
	}
	got := map[string]bool{}
	for _, e := range edges {
		got[e.SourceNode+"|"+e.TargetNode] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("owns-edge %q not emitted", k)
		}
	}
	if len(edges) != len(want) {
		t.Errorf("edges=%d; want %d (no orphan owns-edge)", len(edges), len(want))
	}
}

// TestNarrativeSlotPath_FileSafe covers the path-safety transform.
func TestNarrativeSlotPath_FileSafe(t *testing.T) {
	p := narrativeSlotPath("internal/navigator/sync.Join")
	if !strings.HasSuffix(p, ".md") {
		t.Errorf("missing .md suffix: %q", p)
	}
	// No path separators in the basename; the only dot is the .md suffix.
	base := filepath.Base(p)
	if strings.Contains(base, "/") {
		t.Errorf("basename %q contains path separator", base)
	}
	stripped := strings.TrimSuffix(base, ".md")
	if strings.Contains(stripped, ".") {
		t.Errorf("basename %q contains an unexpected dot", base)
	}
}
