package tiers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogNavigator_WritesDiagnosticLine covers logNavigator — the fail-open
// diagnostic sink for REQ-NS3-020.
func TestLogNavigator_WritesDiagnosticLine(t *testing.T) {
	root := t.TempDir()
	logNavigator(root, "tiers: diagnostic one")
	logNavigator(root, "tiers: diagnostic two")
	path := filepath.Join(root, navigatorLogRelPath)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("log not written: %v", err)
	}
	if !strings.Contains(string(body), "diagnostic one") || !strings.Contains(string(body), "diagnostic two") {
		t.Errorf("log missing lines:\n%s", body)
	}
	// Two lines.
	if got := len(strings.Split(strings.TrimSpace(string(body)), "\n")); got != 2 {
		t.Errorf("log line count=%d; want 2", got)
	}
}

// TestWriteOverlay_FailOpen_OnMkdirError covers writeOverlay's error path
// (project root whose .moai/project/navigator parent cannot be created).
func TestWriteOverlay_FailOpen_OnMkdirError(t *testing.T) {
	// Use a projectRoot path that's a file (not a dir) so MkdirAll fails.
	root := t.TempDir()
	filePath := filepath.Join(root, "blocker")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// projectRoot = filePath/a/b/c — MkdirAll will fail because filePath is a file.
	err := writeOverlay(filepath.Join(filePath, "a", "b"), TiersOverlay{})
	if err == nil {
		t.Errorf("writeOverlay returned nil; want MkdirAll error")
	}
}

// TestReadM0DecisionIDs_Unparseable verifies the unparseable-graph path
// logs and returns empty (REQ-NS3-020 fail-open).
func TestReadM0DecisionIDs_Unparseable(t *testing.T) {
	root := t.TempDir()
	navDir := filepath.Join(root, ".moai", "project", "navigator")
	if err := os.MkdirAll(navDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(navDir, "nav-graph.json"),
		[]byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	ids := readM0DecisionIDs(root)
	if len(ids) != 0 {
		t.Errorf("unparseable nav-graph yielded ids=%v; want empty", ids)
	}
	// Diagnostic logged.
	logPath := filepath.Join(root, navigatorLogRelPath)
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("diagnostic log not written under unparseable graph: %v", err)
	}
}

// TestBuildSupersededByEdges_FilterEmpty covers the empty-SupersededBy
// filter branch.
func TestBuildSupersededByEdges_FilterEmpty(t *testing.T) {
	in := []DecisionEnrichment{
		{Identifier: "a", SupersededBy: "b"},
		{Identifier: "c", SupersededBy: ""}, // filtered out
	}
	out := buildSupersededByEdges(in)
	if len(out) != 1 || out[0].SourceNode != "decision:a" || out[0].TargetNode != "decision:b" {
		t.Errorf("got %+v; want 1 edge a→b", out)
	}
}

// TestWriteNarrativeMetadata_RoundTripAlreadyDone covers the MkdirAll + rename
// path on a fresh root.
func TestWriteNarrativeMetadata_RoundTripAlreadyDone(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "deep", "dir", "x.metadata.json")
	if err := writeNarrativeMetadata(p, "sha", "hash"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("metadata not written: %v", err)
	}
}

// TestNarrativeMetadataPath_Shape verifies the sidecar path is filesystem-safe.
func TestNarrativeMetadataPath_Shape(t *testing.T) {
	p := narrativeMetadataPath("pkg.F")
	if !strings.HasSuffix(p, ".metadata.json") {
		t.Errorf("missing suffix: %q", p)
	}
}

// TestSymbolFilename_Replacements covers the special-char transforms.
func TestSymbolFilename_Replacements(t *testing.T) {
	cases := map[string]string{
		"a.B":            "a_B",
		"a/b/c":          "a_b_c",
		"recv.(*Type).F": "recv__ptrType__F",
	}
	for in, want := range cases {
		if got := symbolFilename(in); got != want {
			t.Errorf("symbolFilename(%q)=%q; want %q", in, got, want)
		}
	}
}

// TestShouldRedraft_MalformedMetadata covers the JSON-unparseable branch.
func TestShouldRedraft_MalformedMetadata(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "x.metadata.json")
	if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !shouldRedraft(p, "any") {
		t.Errorf("shouldRedraft=false on malformed metadata; want true (redraft)")
	}
}

// TestSplitTopLevel_AndSpecials covers splitTopLevel edge cases plus
// isNoiseToken and looksLikePath for completeness.
func TestSplitTopLevel_AndSpecials(t *testing.T) {
	// "and" separator.
	got := splitTopLevel("a and b, c")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Errorf("splitTopLevel len=%d; want %d (%v)", len(got), len(want), got)
	}
	// isNoiseToken: ASCII path is not noise, Korean is noise.
	if isNoiseToken("internal/cli") {
		t.Errorf("internal/cli flagged as noise")
	}
	if !isNoiseToken("한글") {
		t.Errorf("Korean text not flagged as noise")
	}
	if !isNoiseToken("subgraph") {
		t.Errorf("subgraph not flagged as noise")
	}
	// looksLikePath: needs slash or dot, no whitespace.
	if !looksLikePath("a/b") {
		t.Errorf("a/b should look like a path")
	}
	if looksLikePath("bareword") {
		t.Errorf("bareword should not look like a path")
	}
	if looksLikePath("has space/x") {
		t.Errorf("path with space should not look like a path")
	}
}

// TestEnumerateBlueprints_Unparseable_FailOpen covers the JSON-unparseable
// module_tree.json branch.
func TestEnumerateBlueprints_Unparseable_FailOpen(t *testing.T) {
	root := t.TempDir()
	writeModuleTree(t, root, "{not json")
	nodes, edges, err := enumerateBlueprints(root)
	if err != nil {
		t.Errorf("error on unparseable tree: %v (fail-open)", err)
	}
	if len(nodes) != 0 || len(edges) != 0 {
		t.Errorf("unparseable tree emitted records; want 0/0")
	}
}

// TestEnumerateBlueprints_SkipsEmptyPackagePath covers the package_path==""
// skip branch.
func TestEnumerateBlueprints_SkipsEmptyPackagePath(t *testing.T) {
	root := t.TempDir()
	writeModuleTree(t, root, `{"modules":[{"package_path":"","display_name":"x"}]}`)
	nodes, _, err := enumerateBlueprints(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 0 {
		t.Errorf("empty package_path entry emitted a node; want skip")
	}
}

// TestInstantiateOverview_DoesNotOverwriteAuthored covers the authored
// overview.md preservation branch.
func TestInstantiateOverview_DoesNotOverwriteAuthored(t *testing.T) {
	root := t.TempDir()
	node := BlueprintNode{
		Identifier:  "internal/x",
		DisplayName: "X",
	}
	overviewPath := filepath.Join(root, ".moai", "project", "blueprint", "internal", "x", "overview.md")
	if err := os.MkdirAll(filepath.Dir(overviewPath), 0o755); err != nil {
		t.Fatal(err)
	}
	authored := []byte("# hand-written overview\n")
	if err := os.WriteFile(overviewPath, authored, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := instantiateOverview(root, node, "abc"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(overviewPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(authored) {
		t.Errorf("authored overview.md was overwritten by instantiateOverview")
	}
}

// TestEnumerateContracts_SkipsIncompleteEntry covers the skip branch when
// identifier or contract_path is empty.
func TestEnumerateContracts_SkipsIncompleteEntry(t *testing.T) {
	root := t.TempDir()
	registry := "contracts:\n  - identifier: x\n    contract_kind: schema\n    contract_path: \"\"\n  - identifier: \"\"\n    contract_path: y\n    contract_kind: schema\n  - identifier: ok\n    contract_kind: schema\n    contract_path: ok-path\n    validator_command: 'true'\n"
	writeRegistry(t, root, []byte(registry))
	nodes, err := enumerateContracts(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Errorf("nodes=%d; want 1 (incomplete entries skipped)", len(nodes))
	}
}

// TestCheckContractDrift_ShellCommandError covers the case where sh -c
// fails to start (validator unparseable to shell). Already covered by the
// missing-binary path; this confirms the empty-validator shortcut.
func TestCheckContractDrift_WhitespaceOnlyValidator(t *testing.T) {
	root := t.TempDir()
	c := ContractNode{ValidatorCommand: "   "}
	if got := checkContractDrift(root, c); got != DriftUnknown {
		t.Errorf("whitespace validator status=%q; want unknown", got)
	}
}

// TestSortOverlay_AllTiersPopulated verifies the overlay-level sorter
// orders every slice.
func TestSortOverlay_AllTiersPopulated(t *testing.T) {
	o := TiersOverlay{
		Tier0Contracts: []ContractNode{
			{Identifier: "b"}, {Identifier: "a"},
		},
		Tier1Blueprints: []BlueprintNode{
			{Identifier: "z"}, {Identifier: "a"},
		},
		Tier2Decisions: []DecisionEnrichment{
			{Identifier: "z"}, {Identifier: "a"},
		},
		Tier3Symbols: []SymbolEnrichment{
			{Identifier: "z"}, {Identifier: "a"},
		},
		TierEdges: []TierEdge{
			{EdgeType: EdgeOwns, SourceNode: "b", TargetNode: "y"},
			{EdgeType: EdgeModule, SourceNode: "a", TargetNode: "b"},
		},
	}
	sortOverlay(&o)
	if o.Tier0Contracts[0].Identifier != "a" {
		t.Errorf("Tier0Contracts not sorted")
	}
	if o.Tier1Blueprints[0].Identifier != "a" {
		t.Errorf("Tier1Blueprints not sorted")
	}
	if o.Tier2Decisions[0].Identifier != "a" {
		t.Errorf("Tier2Decisions not sorted")
	}
	if o.Tier3Symbols[0].Identifier != "a" {
		t.Errorf("Tier3Symbols not sorted")
	}
	if o.TierEdges[0].EdgeType != EdgeModule {
		t.Errorf("TierEdges not sorted")
	}
}
