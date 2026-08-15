package route

import (
	"strings"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// TestPromoteAuditOrphan verifies the promotion of audit orphan entries into
// work items (AC-NS4-002a, AC-NS4-003). Each orphan becomes a work item with
// source_kind=audit-orphan, the implementation_path as owner (high), and the
// non-generic orphan action directive.
func TestPromoteAuditOrphan(t *testing.T) {
	t.Parallel()

	audit := &AuditReport{
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-A-001", Title: "Feature A", ImplementationPath: "internal/a.go"},
			{SpecID: "SPEC-B-002", Title: "Feature B", ImplementationPath: ""},
		},
	}

	items := Promote(audit, nil, nil, testProjectRoot)

	if len(items) != 2 {
		t.Fatalf("expected 2 work items, got %d", len(items))
	}

	// Both must be audit-orphan with the non-generic action directive.
	for _, item := range items {
		if item.SourceKind != SourceAuditOrphan {
			t.Errorf("source_kind = %q, want audit-orphan", item.SourceKind)
		}
		if item.Action != actionOrphan {
			t.Errorf("action = %q, want %q", item.Action, actionOrphan)
		}
	}

	// Find the high-confidence orphan (with implementation_path) and the
	// low-confidence orphan (empty implementation_path) by their properties,
	// not by position — the sort is by owner_path, which may reorder them.
	var highItem, lowItem *WorkItem
	for i := range items {
		switch items[i].Confidence {
		case ConfidenceHigh:
			highItem = &items[i]
		case ConfidenceLow:
			lowItem = &items[i]
		}
	}

	if highItem == nil {
		t.Fatal("no high-confidence orphan found")
	}
	if highItem.OwnerPath != tp("internal/a.go") {
		t.Errorf("orphan with impl_path: owner = %q, want %q", highItem.OwnerPath, tp("internal/a.go"))
	}

	if lowItem == nil {
		t.Fatal("no low-confidence orphan found")
	}
}

// TestPromoteAuditMissing verifies the promotion of audit missing entries
// into work items (AC-NS4-002a/c, AC-NS4-003). Missing entries without a
// graph resolve to low; with a graph that has a symbol they resolve to
// medium.
func TestPromoteAuditMissing(t *testing.T) {
	t.Parallel()

	graph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "auth.ParseBearer"},
		},
		Edges: []navsync.Edge{
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:auth.ParseBearer",
				TargetNode: "symbol:auth.ParseBearer",
				SourcePath: tp(".moai/project/tech.md"),
				LineNumber: 10,
			},
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:auth.ParseBearer",
				TargetNode: "symbol:auth.ParseBearer",
				SourcePath: tp("internal/auth/login.go"),
				LineNumber: 5,
			},
		},
	}

	audit := &AuditReport{
		Missing: []MissingEntry{
			{
				DesignName: "OAuth2",
				Source:     AuditSource{File: ".moai/project/tech.md", HeadingPath: "## Auth"},
			},
			{
				DesignName: "Unknown feature",
				Source:     AuditSource{File: ".moai/project/structure.md", HeadingPath: "## Misc"},
			},
		},
	}

	items := Promote(audit, nil, graph, testProjectRoot)

	if len(items) != 2 {
		t.Fatalf("expected 2 work items, got %d", len(items))
	}

	// Both must be audit-missing.
	for _, item := range items {
		if item.SourceKind != SourceAuditMissing {
			t.Errorf("source_kind = %q, want audit-missing", item.SourceKind)
		}
		if item.Action != actionMissing {
			t.Errorf("action = %q, want %q", item.Action, actionMissing)
		}
	}

	// Find the medium (symbol resolved from tech.md) and low (structure.md,
	// no symbol in graph) items by their confidence, not by position.
	var medItem, lowItem *WorkItem
	for i := range items {
		switch items[i].Confidence {
		case ConfidenceMedium:
			medItem = &items[i]
		case ConfidenceLow:
			lowItem = &items[i]
		}
	}

	if medItem == nil {
		t.Fatal("no medium-confidence missing found")
	}
	if medItem.OwnerPath != tp("internal/auth/login.go") {
		t.Errorf("missing with symbol: owner = %q, want login.go path", medItem.OwnerPath)
	}

	if lowItem == nil {
		t.Fatal("no low-confidence missing found")
	}
	if !strings.HasSuffix(lowItem.OwnerPath, "structure.md") {
		t.Errorf("missing without symbol: owner = %q, want structure.md path", lowItem.OwnerPath)
	}
}

// TestPromoteDetect verifies the promotion of M1 detect records into work
// items (AC-NS4-002b, AC-NS4-003). Each record becomes a work item with
// source_kind=detect, changed_path as owner (high), and the non-generic
// detect action directive.
func TestPromoteDetect(t *testing.T) {
	t.Parallel()

	detectRows := []DetectRecord{
		{ChangedPath: tp("internal/a.go"), ChangedAt: "2026-01-01T00:00:00Z"},
		{ChangedPath: tp("internal/b.go"), ChangedAt: "2026-01-02T00:00:00Z"},
	}

	items := Promote(nil, detectRows, nil, testProjectRoot)

	if len(items) != 2 {
		t.Fatalf("expected 2 work items, got %d", len(items))
	}

	for _, item := range items {
		if item.SourceKind != SourceDetect {
			t.Errorf("source_kind = %q, want detect", item.SourceKind)
		}
		if item.Confidence != ConfidenceHigh {
			t.Errorf("detect confidence = %q, want high", item.Confidence)
		}
		if item.Action != actionDetect {
			t.Errorf("action = %q, want %q", item.Action, actionDetect)
		}
	}
}

// TestPromoteAllThreeInputs verifies the Route layer consumes all three
// read-only inputs and promotes findings from each (AC-NS4-002a/b/c).
func TestPromoteAllThreeInputs(t *testing.T) {
	t.Parallel()

	audit := &AuditReport{
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
		},
		Missing: []MissingEntry{
			{DesignName: "Feature Y", Source: AuditSource{File: ".moai/project/tech.md"}},
		},
	}
	detectRows := []DetectRecord{
		{ChangedPath: tp("internal/y.go"), ChangedAt: "2026-01-01T00:00:00Z"},
	}

	items := Promote(audit, detectRows, nil, testProjectRoot)

	if len(items) != 3 {
		t.Fatalf("expected 3 work items (1 orphan + 1 missing + 1 detect), got %d", len(items))
	}

	kinds := make(map[SourceKind]int)
	for _, item := range items {
		kinds[item.SourceKind]++
	}
	if kinds[SourceAuditOrphan] != 1 {
		t.Errorf("audit-orphan count = %d, want 1", kinds[SourceAuditOrphan])
	}
	if kinds[SourceAuditMissing] != 1 {
		t.Errorf("audit-missing count = %d, want 1", kinds[SourceAuditMissing])
	}
	if kinds[SourceDetect] != 1 {
		t.Errorf("detect count = %d, want 1", kinds[SourceDetect])
	}
}

// TestPromoteWorkItemFields verifies every work item carries exactly the five
// required fields with valid values (AC-NS4-003a).
func TestPromoteWorkItemFields(t *testing.T) {
	t.Parallel()

	audit := &AuditReport{
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
		},
	}
	detectRows := []DetectRecord{
		{ChangedPath: tp("internal/y.go"), ChangedAt: "2026-01-01T00:00:00Z"},
	}

	items := Promote(audit, detectRows, nil, testProjectRoot)

	validKinds := map[SourceKind]bool{
		SourceAuditMissing: true,
		SourceAuditOrphan:  true,
		SourceDetect:       true,
	}
	validConf := map[Confidence]bool{
		ConfidenceHigh:   true,
		ConfidenceMedium: true,
		ConfidenceLow:    true,
	}

	for i, item := range items {
		if !validKinds[item.SourceKind] {
			t.Errorf("item[%d]: invalid source_kind %q", i, item.SourceKind)
		}
		if item.SourceEntry == nil {
			t.Errorf("item[%d]: nil source_entry", i)
		}
		if item.OwnerPath == "" {
			t.Errorf("item[%d]: empty owner_path", i)
		}
		if !strings.HasPrefix(item.OwnerPath, "/") {
			t.Errorf("item[%d]: owner_path %q is not absolute", i, item.OwnerPath)
		}
		if item.Action == "" {
			t.Errorf("item[%d]: empty action", i)
		}
		if !validConf[item.Confidence] {
			t.Errorf("item[%d]: invalid confidence %q", i, item.Confidence)
		}
	}
}

// TestPromoteDedupAndSort verifies the work-item array is sorted by
// (source_kind, owner_path, identifier) and that duplicate findings within
// the same source_kind are collapsed (AC-NS4-003b).
func TestPromoteDedupAndSort(t *testing.T) {
	t.Parallel()

	audit := &AuditReport{
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-B-002", Title: "B", ImplementationPath: "internal/b.go"},
			{SpecID: "SPEC-A-001", Title: "A", ImplementationPath: "internal/a.go"},
			// Duplicate spec_id + path → should be deduplicated.
			{SpecID: "SPEC-A-001", Title: "A dup", ImplementationPath: "internal/a.go"},
		},
	}

	items := Promote(audit, nil, nil, testProjectRoot)

	// 2 unique orphans (SPEC-A-001 and SPEC-B-002; the third is a dup).
	if len(items) != 2 {
		t.Fatalf("expected 2 deduplicated items, got %d", len(items))
	}

	// Sorted: SPEC-A-001 (internal/a.go) before SPEC-B-002 (internal/b.go).
	if items[0].OwnerPath != tp("internal/a.go") {
		t.Errorf("first item owner = %q, want %q", items[0].OwnerPath, tp("internal/a.go"))
	}
	if items[1].OwnerPath != tp("internal/b.go") {
		t.Errorf("second item owner = %q, want %q", items[1].OwnerPath, tp("internal/b.go"))
	}
}

// TestPromoteEmptyInputs verifies the promotion engine returns an empty (not
// nil) slice when all inputs are absent — the fail-open baseline (partial
// surface of REQ-NS4-009, tested here at the pure-function level).
func TestPromoteEmptyInputs(t *testing.T) {
	t.Parallel()

	items := Promote(nil, nil, nil, testProjectRoot)

	if len(items) != 0 {
		t.Errorf("expected 0 items for empty inputs, got %d", len(items))
	}
}
