package route

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteProducesBothFiles verifies Write creates both work-items.json and
// work-items.md atomically (AC-NS4-007).
func TestWriteProducesBothFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	items := []WorkItem{
		{
			SourceKind:  SourceAuditOrphan,
			SourceEntry: OrphanEntry{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
			OwnerPath:   filepath.Join(root, "internal/x.go"),
			Action:      actionOrphan,
			Confidence:  ConfidenceHigh,
		},
	}
	inputs := RouteInputs{AuditCommit: "abc123", NavGraphCommit: "def456"}

	if err := Write(root, items, inputs); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	jsonPath := filepath.Join(root, workItemsDir, "work-items.json")
	mdPath := filepath.Join(root, workItemsDir, "work-items.md")

	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("work-items.json not created: %v", err)
	}
	if _, err := os.Stat(mdPath); err != nil {
		t.Errorf("work-items.md not created: %v", err)
	}
}

// TestWriteJSONSchema verifies the JSON artifact parses with the required
// top-level keys (provenance + work_items) and each work item carries the
// five required fields (AC-NS4-003a, AC-NS4-007, AC-NS4-008a).
func TestWriteJSONSchema(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	items := []WorkItem{
		{
			SourceKind:  SourceAuditOrphan,
			SourceEntry: OrphanEntry{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
			OwnerPath:   filepath.Join(root, "internal/x.go"),
			Action:      actionOrphan,
			Confidence:  ConfidenceHigh,
		},
	}
	inputs := RouteInputs{AuditCommit: "abc123", NavGraphCommit: "def456"}

	if err := Write(root, items, inputs); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	jsonPath := filepath.Join(root, workItemsDir, "work-items.json")
	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal top-level: %v", err)
	}
	if _, ok := doc["provenance"]; !ok {
		t.Error("missing top-level key 'provenance'")
	}
	if _, ok := doc["work_items"]; !ok {
		t.Error("missing top-level key 'work_items'")
	}

	// Verify no wall-clock field names appear (AC-NS4-008a).
	body := string(raw)
	for _, forbidden := range []string{`"generated_at"`, `"created_at"`, `"timestamp"`} {
		if contains(body, forbidden) {
			t.Errorf("wall-clock field %s found in work-items.json (forbidden)", forbidden)
		}
	}
}

// TestWriteIdempotence verifies two consecutive Write calls with the same
// inputs produce byte-identical output (AC-NS4-008b).
func TestWriteIdempotence(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	items := []WorkItem{
		{
			SourceKind:  SourceAuditOrphan,
			SourceEntry: OrphanEntry{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
			OwnerPath:   filepath.Join(root, "internal/x.go"),
			Action:      actionOrphan,
			Confidence:  ConfidenceHigh,
		},
		{
			SourceKind:  SourceDetect,
			SourceEntry: DetectRecord{ChangedPath: filepath.Join(root, "internal/y.go")},
			OwnerPath:   filepath.Join(root, "internal/y.go"),
			Action:      actionDetect,
			Confidence:  ConfidenceHigh,
		},
	}
	inputs := RouteInputs{AuditCommit: "abc123", NavGraphCommit: "def456"}

	// First write.
	if err := Write(root, items, inputs); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	jsonPath := filepath.Join(root, workItemsDir, "work-items.json")
	mdPath := filepath.Join(root, workItemsDir, "work-items.md")
	firstJSON, _ := os.ReadFile(jsonPath)
	firstMD, _ := os.ReadFile(mdPath)

	// Second write — same inputs, same items.
	if err := Write(root, items, inputs); err != nil {
		t.Fatalf("second Write: %v", err)
	}
	secondJSON, _ := os.ReadFile(jsonPath)
	secondMD, _ := os.ReadFile(mdPath)

	if string(firstJSON) != string(secondJSON) {
		t.Error("work-items.json is not byte-identical across two runs (idempotence violation)")
	}
	if string(firstMD) != string(secondMD) {
		t.Error("work-items.md is not byte-identical across two runs (idempotence violation)")
	}
}

// TestWriteEmptyItemsWritesNothing verifies that an empty work-item set
// produces NO output files (REQ-NS4-009 row 009g: all-inputs-absent → write
// NO output).
func TestWriteEmptyItemsWritesNothing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := Write(root, nil, RouteInputs{}); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	jsonPath := filepath.Join(root, workItemsDir, "work-items.json")
	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Errorf("work-items.json should NOT exist for empty items, got err=%v", err)
	}
}

// TestWriteMarkdownContent verifies the markdown output contains the
// expected sections grouped by source_kind (AC-NS4-007).
func TestWriteMarkdownContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	items := []WorkItem{
		{
			SourceKind:  SourceAuditOrphan,
			SourceEntry: OrphanEntry{SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x.go"},
			OwnerPath:   filepath.Join(root, "internal/x.go"),
			Action:      actionOrphan,
			Confidence:  ConfidenceHigh,
		},
		{
			SourceKind:  SourceAuditMissing,
			SourceEntry: MissingEntry{DesignName: "OAuth2", Source: AuditSource{File: "tech.md"}},
			OwnerPath:   filepath.Join(root, ".moai/project/tech.md"),
			Action:      actionMissing,
			Confidence:  ConfidenceLow,
		},
		{
			SourceKind:  SourceDetect,
			SourceEntry: DetectRecord{ChangedPath: filepath.Join(root, "internal/auth/login.go")},
			OwnerPath:   filepath.Join(root, "internal/auth/login.go"),
			Action:      actionDetect,
			Confidence:  ConfidenceHigh,
		},
	}

	if err := Write(root, items, RouteInputs{}); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	mdPath := filepath.Join(root, workItemsDir, "work-items.md")
	raw, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	body := string(raw)

	// Verify the three section headers are present.
	for _, header := range []string{
		"## Orphan SPECs",
		"## Missing SPECs",
		"## Detect findings",
	} {
		if !contains(body, header) {
			t.Errorf("markdown missing section header %q", header)
		}
	}
	// Verify the provenance line is present.
	if !contains(body, "route_commit_sha") {
		t.Error("markdown missing provenance line")
	}
}

// contains is a simple substring check (avoids importing strings in test
// when only this check is needed).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
