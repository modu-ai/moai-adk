package route

import (
	"fmt"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// TestRouteAccuracy mechanically measures the Route layer's promotion +
// owner-resolution quality against a fixture corpus of 30 input findings
// (AC-NS4-010, REQ-NS4-010). The observed accuracy percentage MUST be
// >= 70.0.
//
// The corpus is constructed in-memory (plan.md §E:184-187):
//   - 6 missing: 3 with @NAV:SYM (resolvable → medium), 3 without (→ low)
//   - 12 orphan: 10 with implementation_path (→ high), 2 empty (→ low)
//   - 12 detect: all unique changed_paths (→ high)
//
// Happy path: 3 medium + 10 high + 12 high = 25/30 = 83.3%.
// B5 fallback: 0 + 10 + 12 = 22/30 = 73.3%. Both ≥ 70%.
func TestRouteAccuracy(t *testing.T) {
	t.Parallel()

	const root = "/fixture-project"

	// --- nav-graph.json fixture: 3 symbol nodes declared in code ---
	graph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "auth.ParseBearer", DisplayName: "ParseBearer"},
			{EntityType: navsync.EntitySymbol, Identifier: "auth.RefreshToken", DisplayName: "RefreshToken"},
			{EntityType: navsync.EntitySymbol, Identifier: "session.Store", DisplayName: "Store"},
		},
		Edges: []navsync.Edge{
			// Design-doc → symbol (the @NAV:SYM references in tech.md).
			symEdge(root+"/.moai/project/tech.md", "symbol:auth.ParseBearer", 10),
			symEdge(root+"/.moai/project/tech.md", "symbol:auth.RefreshToken", 20),
			symEdge(root+"/.moai/project/tech.md", "symbol:session.Store", 30),
			// Code → symbol (the @NAV:SYM declarations in code files).
			symEdge(root+"/internal/auth/login.go", "symbol:auth.ParseBearer", 5),
			symEdge(root+"/internal/auth/refresh.go", "symbol:auth.RefreshToken", 8),
			symEdge(root+"/internal/session/store.go", "symbol:session.Store", 12),
		},
	}

	// --- audit-report.json fixture ---
	audit := &AuditReport{
		AuditCommit: "fixture-audit-sha",
		Missing: []MissingEntry{
			// 3 with @NAV:SYM (source.file = tech.md → symbols in graph → medium).
			{DesignName: "OAuth2 Auth", Source: AuditSource{File: ".moai/project/tech.md", HeadingPath: "## Auth"}},
			{DesignName: "Token Refresh", Source: AuditSource{File: ".moai/project/tech.md", HeadingPath: "## Token"}},
			{DesignName: "Session Store", Source: AuditSource{File: ".moai/project/tech.md", HeadingPath: "## Session"}},
			// 3 without @NAV:SYM (source.file = structure.md → no symbols → low).
			{DesignName: "Module Split", Source: AuditSource{File: ".moai/project/structure.md", HeadingPath: "## Modules"}},
			{DesignName: "Layer Boundaries", Source: AuditSource{File: ".moai/project/structure.md", HeadingPath: "## Layers"}},
			{DesignName: "Dependency Graph", Source: AuditSource{File: ".moai/project/structure.md", HeadingPath: "## Deps"}},
		},
		Orphan: buildOrphanCorpus(root),
	}

	// --- detect JSONL fixture: 12 unique changed_paths across 2 sessions ---
	detectRows := make([]DetectRecord, 12)
	for i := 0; i < 12; i++ {
		detectRows[i] = DetectRecord{
			ChangedPath: fmt.Sprintf("%s/internal/feature/feat%d.go", root, i+1),
			ChangedAt:   "2026-01-01T00:00:00Z",
		}
	}

	// --- Run the Route layer (pure Promote) ---
	items := Promote(audit, detectRows, graph, root)

	// --- Compute accuracy ---
	totalFindings := 30 // 6 missing + 12 orphan + 12 detect
	actionable := 0
	for _, item := range items {
		if item.OwnerPath != "" && (item.Confidence == ConfidenceHigh || item.Confidence == ConfidenceMedium) {
			actionable++
		}
	}
	accuracy := float64(actionable) / float64(totalFindings) * 100.0

	t.Logf("Route accuracy: %.1f%% (%d actionable / %d total)", accuracy, actionable, totalFindings)
	t.Logf("  high:   %d", countConf(items, ConfidenceHigh))
	t.Logf("  medium: %d", countConf(items, ConfidenceMedium))
	t.Logf("  low:    %d", countConf(items, ConfidenceLow))

	if accuracy < 70.0 {
		t.Errorf("Route accuracy %.1f%% < 70.0%% threshold (AC-NS4-010)", accuracy)
	}
}

// symEdge builds a sym-edge for the test graph fixture.
func symEdge(sourcePath, targetNode string, line int) navsync.Edge {
	return navsync.Edge{
		EdgeType:   navsync.EdgeSym,
		SourceNode: targetNode, // self-referencing for simplicity
		TargetNode: targetNode,
		SourcePath: sourcePath,
		LineNumber: line,
	}
}

// buildOrphanCorpus builds 12 orphan entries: 10 with implementation_path,
// 2 without.
func buildOrphanCorpus(root string) []OrphanEntry {
	orphans := make([]OrphanEntry, 12)
	for i := 0; i < 10; i++ {
		orphans[i] = OrphanEntry{
			SpecID:             fmt.Sprintf("SPEC-FEA-%03d", i+1),
			Title:              fmt.Sprintf("Feature %d", i+1),
			ImplementationPath: fmt.Sprintf("internal/feature/impl%d.go", i+1),
		}
	}
	// 2 with empty implementation_path.
	orphans[10] = OrphanEntry{SpecID: "SPEC-UNM-011", Title: "Unmapped A", ImplementationPath: ""}
	orphans[11] = OrphanEntry{SpecID: "SPEC-UNM-012", Title: "Unmapped B", ImplementationPath: ""}
	return orphans
}

// countConf counts work items with the given confidence level.
func countConf(items []WorkItem, conf Confidence) int {
	count := 0
	for _, item := range items {
		if item.Confidence == conf {
			count++
		}
	}
	return count
}
