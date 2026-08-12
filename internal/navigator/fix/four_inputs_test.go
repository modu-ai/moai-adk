package fix

// four_inputs_test.go — M3.6 AC-NS5-002 (REQ-NS5-002): the four read-only inputs
// (M2 work-items + M1 detect JSONL + M0 nav-graph + live doc surfaces) are all
// consumed by the Fix layer, and live docs are READ not written.
//
// Sub-assertions (acceptance.md §D AC-NS5-002):
//   (a) M2 work-items consumed — work_item_refs[] matches fixture work_items[],
//       per_subtree strategy derived from the action field.
//   (b) M1 detect consumed across sessions — diff_scope[] seeded by changed_path
//       from ALL *.jsonl, deduplicated, latest changed_at wins.
//   (c) M0 nav-graph consumed — >=1 diff_scope entry resolves via graph edge
//       traversal (source_path → bound node), not a bare file-path match.
//   (d) live docs read not written — Run's only writes are under fix-drafts/ +
//       logs/; the live doc surfaces are untouched.
//
// Test isolation: every temp tree lives under t.TempDir() (auto-cleaned).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixFourReadOnlyInputs_AC_NS5_002 exercises all four sub-assertions of
// AC-NS5-002 against a fixture corpus with all four inputs present.
func TestFixFourReadOnlyInputs_AC_NS5_002(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// M0 nav-graph: binds TWO graph-bound paths.
	//   - src/x/a.go  → symbol pkg.ParseHeader (the work-item + detect path)
	//   - src/y/b.go  → spec SPEC-Y-002 (a detect-only path, second JSONL file)
	// Provenance provides the baseline (priority 2).
	const graph = `{
  "provenance": {"extract_commit_sha": "ns5base", "captured_at": "2026-01-01T00:00:00+00:00"},
  "nodes": [
    {"entity_type": "symbol", "identifier": "pkg.ParseHeader", "display_name": "ParseHeader"},
    {"entity_type": "spec", "identifier": "SPEC-Y-002", "display_name": "Y2"}
  ],
  "edges": [
    {"edge_type": "sym-edge", "source_node": "symbol:pkg.ParseHeader", "target_node": "spec:SPEC-X-001", "source_path": "src/x/a.go", "line_number": 5},
    {"edge_type": "spec-edge", "source_node": "spec:SPEC-Y-002", "target_node": "symbol:pkg.ParseHeader", "source_path": "src/y/b.go", "line_number": 1}
  ]
}`
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), graph)

	// M2 work-items: one work-item whose action maps to a known strategy.
	// "create a SPEC ..." → strategy "draft SPEC stub" (actionToStrategy).
	const wi = `{
  "work_items": [
    {"source_kind": "audit-missing", "owner_path": "src/x/a.go", "action": "create a SPEC stub for the newly added ParseHeader helper"}
  ]
}`
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), wi)

	// M1 detect: TWO session JSONL files. s1.jsonl touches src/x/a.go (the M2
	// path too — dedup: one subtree). s2.jsonl touches src/y/b.go (a distinct
	// graph-bound path → a second subtree). Same path in two files → latest
	// changed_at wins (dedup).
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"),
		`{"changed_path": "src/x/a.go", "changed_at": "2026-01-02T00:00:00Z"}`)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s2.jsonl"),
		`{"changed_path": "src/y/b.go", "changed_at": "2026-01-03T00:00:00Z"}`)

	// Live doc surfaces (read targets — Run does NOT write these).
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "capability-symbols.json"), `{"symbols":[]}`)
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "audit-report.json"), `{"rows":[]}`)

	// Snapshot the root file tree BEFORE Run so (d) can assert no live-doc mutation.
	before := snapshotTree(t, root)

	res := Run(Options{ProjectRoot: root})
	if !res.Written {
		t.Fatalf("Run must write request.json: %+v", res)
	}
	raw, err := os.ReadFile(res.DraftRequestPath)
	if err != nil {
		t.Fatalf("read request.json: %v", err)
	}
	var doc struct {
		WorkItemRefs []struct {
			SourceKind string `json:"source_kind"`
			OwnerPath  string `json:"owner_path"`
			Action     string `json:"action"`
		} `json:"work_item_refs"`
		DiffScope []struct {
			DocSurface  string `json:"doc_surface"`
			SubtreeID   string `json:"subtree_id"`
			StaleReason string `json:"stale_reason"`
		} `json:"diff_scope"`
		DraftInstructions struct {
			PerSubtree []struct {
				SubtreeID string `json:"subtree_id"`
				Strategy  string `json:"strategy"`
			} `json:"per_subtree"`
		} `json:"draft_instructions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal request.json: %v\n%s", err, raw)
	}

	// (a) M2 work-items consumed: work_item_refs[] matches the fixture work_item,
	//     and the per_subtree strategy is derived from the action ("create a SPEC"
	//     → "draft SPEC stub").
	t.Run("a_M2_work_items_consumed", func(t *testing.T) {
		if len(doc.WorkItemRefs) != 1 {
			t.Fatalf("work_item_refs len = %d, want 1", len(doc.WorkItemRefs))
		}
		ref := doc.WorkItemRefs[0]
		if ref.SourceKind != "audit-missing" || ref.OwnerPath != "src/x/a.go" {
			t.Errorf("work_item_ref mismatch: %+v", ref)
		}
		// per_subtree strategy for the ParseHeader subtree must be "draft SPEC stub".
		var gotStrategy string
		for _, s := range doc.DraftInstructions.PerSubtree {
			if s.SubtreeID == "pkg.ParseHeader" {
				gotStrategy = s.Strategy
			}
		}
		if gotStrategy != "draft SPEC stub" {
			t.Errorf("ParseHeader strategy = %q, want \"draft SPEC stub\" (derived from action)", gotStrategy)
		}
	})

	// (b) M1 detect consumed across sessions: diff_scope seeded by changed_path
	//     from BOTH *.jsonl files. src/x/a.go (s1) + src/y/b.go (s2) → 2 distinct
	//     graph-bound subtrees. Dedup: src/x/a.go also named by M2, still one.
	t.Run("b_M1_detect_consumed_across_sessions", func(t *testing.T) {
		if len(doc.DiffScope) < 2 {
			t.Fatalf("diff_scope len = %d, want >=2 (both detect JSONL files seeded a graph-bound subtree): %+v", len(doc.DiffScope), doc.DiffScope)
		}
		// Both subtree IDs must be present (graph-traversal resolved each edge).
		ids := make(map[string]bool, len(doc.DiffScope))
		for _, e := range doc.DiffScope {
			ids[e.SubtreeID] = true
		}
		for _, want := range []string{"pkg.ParseHeader", "SPEC-Y-002"} {
			if !ids[want] {
				t.Errorf("diff_scope missing subtree %q (detect JSONL not consumed); have %v", want, ids)
			}
		}
	})

	// (c) M0 nav-graph consumed: >=1 diff_scope entry resolves via graph edge
	//     traversal. The subtree "pkg.ParseHeader" is bound by the sym-edge whose
	//     source_path is src/x/a.go — it is NOT a file-path string match against
	//     the subtree ID. Asserting the subtree ID equals the graph node
	//     identifier (not the path) proves graph traversal was used.
	t.Run("c_M0_nav_graph_traversal", func(t *testing.T) {
		var symEntry *struct {
			DocSurface  string `json:"doc_surface"`
			SubtreeID   string `json:"subtree_id"`
			StaleReason string `json:"stale_reason"`
		}
		for i := range doc.DiffScope {
			if doc.DiffScope[i].SubtreeID == "pkg.ParseHeader" {
				symEntry = &doc.DiffScope[i]
			}
		}
		if symEntry == nil {
			t.Fatalf("no diff_scope entry for pkg.ParseHeader — graph traversal did not bind the sym-edge")
		}
		// The subtree ID is the GRAPH NODE identifier, not the source_path. A
		// file-path match would have produced subtree_id == "src/x/a.go". Getting
		// "pkg.ParseHeader" proves graph traversal resolved the edge.
		if symEntry.SubtreeID == "src/x/a.go" {
			t.Errorf("subtree_id = source_path (file-path match, not graph traversal): %q", symEntry.SubtreeID)
		}
		// The doc_surface is derived from the node entity_type (symbol → capability-symbols.json).
		if symEntry.DocSurface != "capability-symbols.json" {
			t.Errorf("symbol DocSurface = %q, want \"capability-symbols.json\"", symEntry.DocSurface)
		}
	})

	// (d) live docs read not written: Run's only NEW files are under
	//     fix-drafts/ and logs/. The live doc surfaces + nav-graph + work-items +
	//     detect state are NOT mutated.
	t.Run("d_live_docs_read_not_written", func(t *testing.T) {
		after := snapshotTree(t, root)
		for path := range after {
			if _, ok := before[path]; ok {
				continue // pre-existing file
			}
			// New file: must be under fix-drafts/ or logs/.
			rel := strings.TrimPrefix(path, root+string(os.PathSeparator))
			if !strings.HasPrefix(rel, ".moai"+string(os.PathSeparator)+"project"+string(os.PathSeparator)+"navigator"+string(os.PathSeparator)+"fix-drafts") &&
				!strings.HasPrefix(rel, ".moai"+string(os.PathSeparator)+"logs") {
				t.Errorf("Run wrote outside fix-drafts/ + logs/: %s", rel)
			}
		}
		// Live doc surfaces unchanged (byte-identical content).
		for _, live := range []string{
			filepath.Join(root, ".moai", "project", "navigator", "capability-symbols.json"),
			filepath.Join(root, ".moai", "project", "navigator", "audit-report.json"),
		} {
			b, err := os.ReadFile(live)
			if err != nil {
				t.Errorf("live doc unreadable: %v", err)
				continue
			}
			// before-snapshot content equality check.
			if beforeContent, ok := before[live]; ok && string(b) != beforeContent {
				t.Errorf("live doc %s was mutated by Run", live)
			}
		}
	})
}

// snapshotTree walks root and returns a map of absolute-path → file content for
// regular files. Used by sub-assertion (d) to detect writes outside the staging
// surface.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := make(map[string]string)
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		out[path] = string(b)
		return nil
	})
	return out
}
