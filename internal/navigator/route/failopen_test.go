package route

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// writeFixture is a test helper that writes content to a path under root,
// creating parent directories as needed.
func writeFixture(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// detectLine renders one detect-fixture JSONL row for a changed path. The path
// is marshaled rather than concatenated into the literal: on Windows a temp
// path contains backslashes (C:\Users\RUNNER~1\...), and "\U" inside a JSON
// string is an invalid escape, so a concatenated line fails to parse and the
// changed path silently reads back as empty.
func detectLine(t *testing.T, changedPath, changedAt string) string {
	t.Helper()
	row := map[string]any{
		"changed_path":   changedPath,
		"changed_at":     changedAt,
		"affected_nodes": []any{},
		"affected_edges": []any{},
	}
	b, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal detect line: %v", err)
	}
	return string(b)
}

// readWorkItemsJSON reads and parses the work-items.json output.
func readWorkItemsJSON(t *testing.T, root string) WorkItemsArtifact {
	t.Helper()
	path := filepath.Join(root, workItemsDir, "work-items.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read work-items.json: %v", err)
	}
	var art WorkItemsArtifact
	if err := json.Unmarshal(raw, &art); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return art
}

// --- 009a: audit absent → promote from detect + nav-graph only (partial) ---

func TestFailOpen_009a_AuditAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// No audit-report.json. Provide detect + graph.
	writeDetectFixture(t, root, "session-a",
		detectLine(t, filepath.Join(root, "internal/x.go"), "2026-01-01T00:00:00Z"))

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	art := readWorkItemsJSON(t, root)
	if len(art.WorkItems) == 0 {
		t.Error("expected work items from detect, got 0")
	}
	for _, item := range art.WorkItems {
		if item.SourceKind != SourceDetect {
			t.Errorf("expected all detect items, got %q", item.SourceKind)
		}
	}
}

// --- 009b: detect absent → promote from audit only (partial) ---

func TestFailOpen_009b_DetectAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// No detect state directory. Provide audit only.
	writeAuditFixture(t, root, &AuditReport{
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
		},
	})

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	art := readWorkItemsJSON(t, root)
	if len(art.WorkItems) == 0 {
		t.Error("expected work items from audit, got 0")
	}
	for _, item := range art.WorkItems {
		if item.SourceKind != SourceAuditOrphan {
			t.Errorf("expected all audit-orphan items, got %q", item.SourceKind)
		}
	}
}

// --- 009c: nav-graph absent → owner resolution degrades (all missing → low) ---

func TestFailOpen_009c_NavGraphAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// No nav-graph.json. Provide audit with a missing entry.
	writeAuditFixture(t, root, &AuditReport{
		Missing: []MissingEntry{
			{DesignName: "Feature", Source: AuditSource{File: "tech.md"}},
		},
	})

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	art := readWorkItemsJSON(t, root)
	for _, item := range art.WorkItems {
		if item.SourceKind == SourceAuditMissing && item.Confidence != ConfidenceLow {
			t.Errorf("missing without graph: confidence = %q, want low", item.Confidence)
		}
	}
}

// --- 009d: unparseable JSON → skip malformed input ---

func TestFailOpen_009d_UnparseableJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Malformed audit JSON.
	writeFixture(t, root, auditReportRelPath, `{not valid json`)
	// Valid detect.
	writeDetectFixture(t, root, "session-a",
		detectLine(t, filepath.Join(root, "internal/y.go"), "2026-01-01T00:00:00Z"))

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	// Audit was skipped; detect items should still be present.
	art := readWorkItemsJSON(t, root)
	hasDetect := false
	for _, item := range art.WorkItems {
		if item.SourceKind == SourceDetect {
			hasDetect = true
		}
	}
	if !hasDetect {
		t.Error("expected detect items after skipping malformed audit")
	}
}

// --- 009e: schema-invalid → degrade affected source ---

func TestFailOpen_009e_SchemaInvalid(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Schema-invalid audit: valid JSON but missing the orphan/missing arrays.
	writeFixture(t, root, auditReportRelPath, `{"audit_at":"2026-01-01","audit_commit":"abc"}`)
	// Valid detect.
	writeDetectFixture(t, root, "session-a",
		detectLine(t, filepath.Join(root, "internal/z.go"), "2026-01-01T00:00:00Z"))

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	art := readWorkItemsJSON(t, root)
	// The schema-invalid audit degrades: missing/orphan treated as empty.
	// Only detect items should be present.
	for _, item := range art.WorkItems {
		if item.SourceKind == SourceAuditOrphan || item.SourceKind == SourceAuditMissing {
			t.Errorf("expected no audit items from schema-invalid audit, got %q", item.SourceKind)
		}
	}
}

// --- 009f: owner-resolution error → mark confidence low, fallback owner ---

func TestFailOpen_009f_OwnerResolutionError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Graph with a symbol referencing a NON-EXISTENT code path.
	nonExistentPath := filepath.Join(root, "internal/nonexistent/deleted.go")
	graph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "deleted.Func"},
		},
		Edges: []navsync.Edge{
			{
				EdgeType:   navsync.EdgeSym,
				TargetNode: "symbol:deleted.Func",
				SourcePath: filepath.Join(root, ".moai/project/tech.md"),
			},
			{
				EdgeType:   navsync.EdgeSym,
				TargetNode: "symbol:deleted.Func",
				SourcePath: nonExistentPath, // does NOT exist on disk
			},
		},
	}
	writeGraphFixture(t, root, graph)
	writeAuditFixture(t, root, &AuditReport{
		Missing: []MissingEntry{
			{DesignName: "Deleted Feature", Source: AuditSource{File: ".moai/project/tech.md"}},
		},
	})

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	art := readWorkItemsJSON(t, root)
	for _, item := range art.WorkItems {
		if item.SourceKind == SourceAuditMissing {
			if item.Confidence != ConfidenceLow {
				t.Errorf("owner-resolution error: confidence = %q, want low (path does not exist on disk)", item.Confidence)
			}
		}
	}
}

// --- 009g: all inputs absent → write NO output ---

func TestFailOpen_009g_AllInputsAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	if err := RunDefault(root); err != nil {
		t.Fatalf("RunDefault returned non-nil: %v", err)
	}
	jsonPath := filepath.Join(root, workItemsDir, "work-items.json")
	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Errorf("work-items.json should NOT exist when all inputs absent, got err=%v", err)
	}
}

// --- 009h: timeout → exit 0, partial or empty set, no log ---

func TestFailOpen_009h_Timeout(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Provide inputs so there IS something to promote.
	writeAuditFixture(t, root, &AuditReport{
		Orphan: []OrphanEntry{
			{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
		},
	})

	// Cancel the context immediately — simulates a timeout before work starts.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := Run(ctx, root); err != nil {
		t.Fatalf("Run with canceled ctx returned non-nil: %v", err)
	}
	// With a canceled context, Run returns nil without writing output.
	// This is the expected behavior (partial or empty set, no log).
}

// --- Helper functions for fixtures ---

func writeAuditFixture(t *testing.T, root string, audit *AuditReport) {
	t.Helper()
	raw, err := json.Marshal(audit)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, auditReportRelPath, string(raw))
}

func writeGraphFixture(t *testing.T, root string, graph *navsync.Graph) {
	t.Helper()
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, navGraphRelPath, string(raw))
}

func writeDetectFixture(t *testing.T, root, sessionID, line string) {
	t.Helper()
	relPath := filepath.Join(detectStateDir, sessionID+".jsonl")
	writeFixture(t, root, relPath, line+"\n")
}
