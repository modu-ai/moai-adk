package tiers

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// moduleTreeRel is the same constant the engine uses; mirrored here so tests
// do not depend on the engine's unexported constant shape.
const moduleTreeRel = ".moai/project/blueprint/module_tree.json"

// writeModuleTree writes a JSON body to the module_tree.json path under root.
func writeModuleTree(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Dir(filepath.Join(root, moduleTreeRel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, moduleTreeRel), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestBlueprint_ScaffoldAbsence_CreatesDraft exercises AC-NS3-004: given a
// codemaps dependencies.md and NO existing module_tree.json, a plain run
// scaffolds a draft module_tree.json (one BlueprintNode per detected module).
func TestBlueprint_ScaffoldAbsence_CreatesDraft(t *testing.T) {
	root := t.TempDir()
	seedCodemaps(t, root, `# 패키지 의존도 분석

## 의존도 그래프

- internal/navigator/sync depends on internal/navigator/astx
- internal/cli depends on internal/navigator/sync
`)
	if err := ensureModuleTreeScaffold(root, false); err != nil {
		t.Fatalf("ensureModuleTreeScaffold error: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, moduleTreeRel))
	if err != nil {
		t.Fatalf("scaffold not written: %v", err)
	}
	if !strings.Contains(string(body), "internal/navigator/sync") {
		t.Errorf("scaffold missing sync module:\n%s", body)
	}
}

// TestBlueprint_PlainRun_DoesNotOverwriteAuthoredTree is the LOAD-BEARING
// blueprint-first invariant test (AC-NS3-004 / REQ-NS3-006): given a
// human-edited module_tree.json, a plain run (no --rescaffold) MUST leave
// the file byte-identical. The deterministic layer never auto-replaces.
func TestBlueprint_PlainRun_DoesNotOverwriteAuthoredTree(t *testing.T) {
	root := t.TempDir()
	authored := `{
  "modules": [
    {
      "package_path": "internal/imagined",
      "display_name": "Authored Module",
      "layer": "domain",
      "responsibility": "hand-written content",
      "depends_on": []
    }
  ]
}
`
	writeModuleTree(t, root, authored)
	seedCodemaps(t, root, "- internal/different depends on nothing\n")

	wantHash := hashBytes([]byte(authored))

	if err := ensureModuleTreeScaffold(root, false); err != nil {
		t.Fatalf("ensureModuleTreeScaffold error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, moduleTreeRel))
	if err != nil {
		t.Fatal(err)
	}
	if hashBytes(got) != wantHash {
		t.Errorf("plain run overwrote authored module_tree.json (REQ-NS3-004):\nbefore:\n%s\nafter:\n%s",
			authored, got)
	}
}

// TestBlueprint_RescaffoldFlag_Overwrites verifies --rescaffold DOES
// overwrite (REQ-NS3-004 opt-in).
func TestBlueprint_RescaffoldFlag_Overwrites(t *testing.T) {
	root := t.TempDir()
	writeModuleTree(t, root, `{"modules":[]}`)
	seedCodemaps(t, root, "- internal/x depends on nothing\n")
	if err := ensureModuleTreeScaffold(root, true); err != nil {
		t.Fatalf("ensureModuleTreeScaffold(rescaffold) error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, moduleTreeRel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "internal/x") {
		t.Errorf("rescaffold did not overwrite with new seed:\n%s", got)
	}
}

// TestBlueprint_ScaffoldAbsence_CodemapsAbsent_FailOpen: no dependencies.md
// present → scaffold degrades to an empty (but valid) module_tree.json.
func TestBlueprint_ScaffoldAbsence_CodemapsAbsent_FailOpen(t *testing.T) {
	root := t.TempDir()
	if err := ensureModuleTreeScaffold(root, false); err != nil {
		t.Fatalf("codemaps absent returned error: %v (fail-open)", err)
	}
	got, err := os.ReadFile(filepath.Join(root, moduleTreeRel))
	if err != nil {
		t.Fatalf("scaffold not written under codemaps-absent: %v", err)
	}
	if !strings.Contains(string(got), `"modules"`) {
		t.Errorf("scaffold missing modules field:\n%s", got)
	}
}

// TestBlueprint_Enumerate_EmitsNodesAndModuleEdges exercises AC-NS3-007: an
// authored module_tree.json with two modules A→B yields two BlueprintNode
// records and one module-edge (A→B).
func TestBlueprint_Enumerate_EmitsNodesAndModuleEdges(t *testing.T) {
	root := t.TempDir()
	authored := `{
  "modules": [
    {"package_path":"internal/a","display_name":"A","layer":"domain","responsibility":"a","depends_on":["internal/b"]},
    {"package_path":"internal/b","display_name":"B","layer":"infrastructure","responsibility":"b","depends_on":[]}
  ]
}`
	writeModuleTree(t, root, authored)
	nodes, edges, err := enumerateBlueprints(root)
	if err != nil {
		t.Fatalf("enumerateBlueprints error: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("nodes=%d; want 2", len(nodes))
	}
	// Sort for stable assertion.
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Identifier < nodes[j].Identifier })
	if nodes[0].Identifier != "internal/a" {
		t.Errorf("nodes[0].Identifier=%q; want internal/a", nodes[0].Identifier)
	}
	// Exactly one module-edge A→B.
	var moduleEdges []TierEdge
	for _, e := range edges {
		if e.EdgeType == EdgeModule {
			moduleEdges = append(moduleEdges, e)
		}
	}
	if len(moduleEdges) != 1 {
		t.Fatalf("module-edges=%d; want 1 (%+v)", len(moduleEdges), moduleEdges)
	}
	want := TierEdge{EdgeType: EdgeModule, SourceNode: "blueprint:internal/a", TargetNode: "blueprint:internal/b"}
	if moduleEdges[0] != want {
		t.Errorf("module-edge=%+v; want %+v", moduleEdges[0], want)
	}
}

// TestBlueprint_Enumerate_AuthoredNodeShape verifies the loaded BlueprintNode
// preserves the authored fields (display_name, layer, responsibility,
// depends_on, overview_path).
func TestBlueprint_Enumerate_AuthoredNodeShape(t *testing.T) {
	root := t.TempDir()
	authored := `{
  "modules": [
    {"package_path":"pkg/x","display_name":"X Module","layer":"presentation","responsibility":"does X","depends_on":["pkg/y"],"overview_path":"x.md"}
  ]
}`
	writeModuleTree(t, root, authored)
	nodes, _, err := enumerateBlueprints(root)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("nodes=%d; want 1", len(nodes))
	}
	n := nodes[0]
	if n.DisplayName != "X Module" || n.Layer != LayerPresentation || n.Responsibility != "does X" {
		t.Errorf("node shape wrong: %+v", n)
	}
	if len(n.DependsOn) != 1 || n.DependsOn[0] != "pkg/y" {
		t.Errorf("DependsOn=%v; want [pkg/y]", n.DependsOn)
	}
	if n.OverviewPath != "x.md" {
		t.Errorf("OverviewPath=%q; want x.md", n.OverviewPath)
	}
}

// TestBlueprint_Overview_KiroSevenSections exercises AC-NS3-005: the
// overview.md template instantiation carries all seven Kiro sections as
// headings AND a provenance block with last_updated_commit.
func TestBlueprint_Overview_KiroSevenSections(t *testing.T) {
	root := t.TempDir()
	node := BlueprintNode{
		Identifier:     "internal/sample",
		DisplayName:    "Sample",
		Layer:          LayerDomain,
		Responsibility: "sample",
		DependsOn:      []string{},
	}
	if err := instantiateOverview(root, node, "abc1234"); err != nil {
		t.Fatalf("instantiateOverview error: %v", err)
	}
	overviewPath := filepath.Join(root, ".moai", "project", "blueprint", "internal", "sample", "overview.md")
	body, err := os.ReadFile(overviewPath)
	if err != nil {
		t.Fatalf("overview.md not written: %v", err)
	}
	want := []string{
		"Component Architecture",
		"Data Flow",
		"Data Model",
		"Error Handling",
		"Test Strategy",
		"Implementation Approach",
		"Migration",
	}
	for _, h := range want {
		if !strings.Contains(string(body), h) {
			t.Errorf("overview.md missing Kiro section %q", h)
		}
	}
	if !strings.Contains(string(body), "last_updated_commit") {
		t.Errorf("overview.md missing provenance block")
	}
	if !strings.Contains(string(body), "abc1234") {
		t.Errorf("overview.md missing last_updated_commit value abc1234")
	}
}

// TestBlueprint_NoDriftFailTest is REMOVED. The REQ-NS3-006 negative-grep AC
// (acceptance.md §D.AC-NS3-006) is canonically verified by the orchestrator's
// verification-batch grep over internal/ for the forbidden drift-failure code
// path. An in-test version of the same grep cannot exist without containing
// the forbidden patterns as string literals (which then self-trips the AC
// grep). The orchestrator-owned grep IS the negative test; no in-test mirror
// needed. See the AC matrix in progress.md §E.2 for the observed-zero output.

// hashBytes returns hex sha256 of b.
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// seedCodemaps writes a dependencies.md body under .moai/project/codemaps/.
func seedCodemaps(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dependencies.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
