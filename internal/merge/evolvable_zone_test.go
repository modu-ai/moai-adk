package merge

import (
	"strings"
	"testing"
)

// ---- ParseEvolvableZones tests ----

func TestParseEvolvableZones_Valid(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"# Skill",
		`<!-- moai:evolvable-start id="section-a" -->`,
		"User content A",
		"More A",
		"<!-- moai:evolvable-end -->",
		"Static line",
		`<!-- moai:evolvable-start id="section-b" -->`,
		"User content B",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	zones, err := ParseEvolvableZones(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(zones) != 2 {
		t.Fatalf("expected 2 zones, got %d", len(zones))
	}

	if zones[0].ID != "section-a" {
		t.Errorf("zone[0].ID = %q, want %q", zones[0].ID, "section-a")
	}
	wantContentA := "User content A\nMore A"
	if zones[0].Content != wantContentA {
		t.Errorf("zone[0].Content = %q, want %q", zones[0].Content, wantContentA)
	}

	if zones[1].ID != "section-b" {
		t.Errorf("zone[1].ID = %q, want %q", zones[1].ID, "section-b")
	}
	if zones[1].Content != "User content B" {
		t.Errorf("zone[1].Content = %q, want %q", zones[1].Content, "User content B")
	}
}

func TestParseEvolvableZones_NoZones(t *testing.T) {
	t.Parallel()

	zones, err := ParseEvolvableZones("# Just a plain file\nNo markers here.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(zones) != 0 {
		t.Errorf("expected 0 zones, got %d", len(zones))
	}
}

func TestParseEvolvableZones_NestedError(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		`<!-- moai:evolvable-start id="outer" -->`,
		"outer content",
		`<!-- moai:evolvable-start id="inner" -->`,
		"inner content",
		"<!-- moai:evolvable-end -->",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	_, err := ParseEvolvableZones(content)
	if err == nil {
		t.Fatal("expected error for nested markers, got nil")
	}
	if !strings.Contains(err.Error(), "nested") {
		t.Errorf("error should mention 'nested', got: %v", err)
	}
}

func TestParseEvolvableZones_DuplicateID(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		`<!-- moai:evolvable-start id="dup" -->`,
		"first",
		"<!-- moai:evolvable-end -->",
		`<!-- moai:evolvable-start id="dup" -->`,
		"second",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	_, err := ParseEvolvableZones(content)
	if err == nil {
		t.Fatal("expected error for duplicate id, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %v", err)
	}
}

func TestParseEvolvableZones_MissingEndMarker(t *testing.T) {
	t.Parallel()

	// Missing end marker should warn and discard the unclosed zone — no error returned.
	content := strings.Join([]string{
		`<!-- moai:evolvable-start id="section-a" -->`,
		"User content A",
		// no end marker
	}, "\n")

	zones, err := ParseEvolvableZones(content)
	if err != nil {
		t.Fatalf("expected no error for missing end marker, got: %v", err)
	}
	// Unclosed zone is discarded.
	if len(zones) != 0 {
		t.Errorf("expected 0 zones (unclosed discarded), got %d", len(zones))
	}
}

func TestParseEvolvableZones_OrphanEndMarker(t *testing.T) {
	t.Parallel()

	// End marker without matching start — should warn but not error.
	content := strings.Join([]string{
		"some content",
		"<!-- moai:evolvable-end -->",
		"more content",
	}, "\n")

	zones, err := ParseEvolvableZones(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(zones) != 0 {
		t.Errorf("expected 0 zones, got %d", len(zones))
	}
}

func TestParseEvolvableZones_EmptyZone(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		`<!-- moai:evolvable-start id="empty" -->`,
		"<!-- moai:evolvable-end -->",
	}, "\n")

	zones, err := ParseEvolvableZones(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
	if zones[0].Content != "" {
		t.Errorf("expected empty content, got %q", zones[0].Content)
	}
}

// ---- MergeEvolvableZones tests ----

func TestMergeEvolvableZones_TemplateUpdatesOutsideZone(t *testing.T) {
	t.Parallel()

	// Template adds a new static line outside zone; user changed something else outside.
	// Both changes should be preserved (normal line merge).
	base := strings.Join([]string{
		"# Header",
		`<!-- moai:evolvable-start id="zone1" -->`,
		"original zone content",
		"<!-- moai:evolvable-end -->",
		"static footer",
	}, "\n")

	current := strings.Join([]string{
		"# Header",
		`<!-- moai:evolvable-start id="zone1" -->`,
		"user modified zone content",
		"<!-- moai:evolvable-end -->",
		"static footer",
	}, "\n")

	updated := strings.Join([]string{
		"# Header (template updated)",
		`<!-- moai:evolvable-start id="zone1" -->`,
		"template changed zone content",
		"<!-- moai:evolvable-end -->",
		"static footer",
	}, "\n")

	merged, err := MergeEvolvableZones(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Header update from template should be applied.
	if !strings.Contains(merged, "# Header (template updated)") {
		t.Errorf("expected template header update in merged output\nmerged:\n%s", merged)
	}

	// User content inside zone should win.
	if !strings.Contains(merged, "user modified zone content") {
		t.Errorf("expected user zone content to win\nmerged:\n%s", merged)
	}

	// Template zone content should NOT appear.
	if strings.Contains(merged, "template changed zone content") {
		t.Errorf("template zone content should NOT appear (user wins)\nmerged:\n%s", merged)
	}
}

func TestMergeEvolvableZones_UserWinsInsideZone(t *testing.T) {
	t.Parallel()

	base := strings.Join([]string{
		`<!-- moai:evolvable-start id="anti-rationalization" -->`,
		"original template content",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	current := strings.Join([]string{
		`<!-- moai:evolvable-start id="anti-rationalization" -->`,
		"user added: Do not rationalize delays",
		"user added: Check actual output",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	updated := strings.Join([]string{
		`<!-- moai:evolvable-start id="anti-rationalization" -->`,
		"template updated section",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	merged, err := MergeEvolvableZones(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(merged, "user added: Do not rationalize delays") {
		t.Errorf("user zone content must survive update\nmerged:\n%s", merged)
	}
	if !strings.Contains(merged, "user added: Check actual output") {
		t.Errorf("user zone content must survive update\nmerged:\n%s", merged)
	}
	if strings.Contains(merged, "template updated section") {
		t.Errorf("template zone content must NOT appear when user has content\nmerged:\n%s", merged)
	}
}

func TestMergeEvolvableZones_TemplateAddsNewZone(t *testing.T) {
	t.Parallel()

	// Template adds a new zone that didn't exist in base or current.
	base := "# Skill\n\nStatic content."
	current := "# Skill\n\nStatic content."
	updated := strings.Join([]string{
		"# Skill",
		"",
		"Static content.",
		`<!-- moai:evolvable-start id="new-section" -->`,
		"New template content here",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	merged, err := MergeEvolvableZones(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(merged, "New template content here") {
		t.Errorf("new zone from template should be added\nmerged:\n%s", merged)
	}
	if !strings.Contains(merged, `moai:evolvable-start id="new-section"`) {
		t.Errorf("zone markers should be preserved\nmerged:\n%s", merged)
	}
}

func TestMergeEvolvableZones_TemplateRemovesZone(t *testing.T) {
	t.Parallel()

	// Template removes a zone that user had content in.
	// User content should be preserved with a warning.
	base := strings.Join([]string{
		"# Skill",
		`<!-- moai:evolvable-start id="removable" -->`,
		"original content",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	current := strings.Join([]string{
		"# Skill",
		`<!-- moai:evolvable-start id="removable" -->`,
		"user evolved content",
		"<!-- moai:evolvable-end -->",
	}, "\n")

	// Updated template has removed the zone entirely.
	updated := "# Skill\nStatic only."

	merged, err := MergeEvolvableZones(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// User content should be appended (preserved).
	if !strings.Contains(merged, "user evolved content") {
		t.Errorf("user content in removed zone should be preserved\nmerged:\n%s", merged)
	}
}

func TestMergeEvolvableZones_FallbackOnParseError(t *testing.T) {
	t.Parallel()

	// Malformed content in 'current' triggers fallback to LineMerge.
	base := "line one\nline two"
	// Nested markers cause parse error.
	current := strings.Join([]string{
		`<!-- moai:evolvable-start id="a" -->`,
		`<!-- moai:evolvable-start id="b" -->`,
		"nested",
		"<!-- moai:evolvable-end -->",
		"<!-- moai:evolvable-end -->",
	}, "\n")
	updated := "line one\nline two updated"

	// Should not return an error (falls back to line merge).
	_, err := MergeEvolvableZones(base, current, updated)
	if err != nil {
		t.Fatalf("fallback should not propagate error: %v", err)
	}
}

// ---- HasEvolvableZones tests ----

func TestHasEvolvableZones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "with marker",
			content: `<!-- moai:evolvable-start id="x" -->`,
			want:    true,
		},
		{
			name:    "without marker",
			content: "# Plain file\nno markers",
			want:    false,
		},
		{
			name:    "empty",
			content: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HasEvolvableZones(tt.content)
			if got != tt.want {
				t.Errorf("HasEvolvableZones(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}
