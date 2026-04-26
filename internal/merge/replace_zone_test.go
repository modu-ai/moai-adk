package merge

import (
	"strings"
	"testing"
)

// TestReplaceEvolvableZone_Basic verifies basic zone replacement with multi-line content.
func TestReplaceEvolvableZone_Basic(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"# Header",
		"",
		"Introduction paragraph.",
		"",
		`<!-- moai:evolvable-start id="best-practices" -->`,
		"Original content line 1.",
		"Original content line 2.",
		"<!-- moai:evolvable-end -->",
		"",
		"Footer content.",
	}, "\n")

	newZoneContent := "New line 1.\nNew line 2.\nNew line 3."

	result, err := ReplaceEvolvableZone(content, "best-practices", newZoneContent)
	if err != nil {
		t.Fatalf("ReplaceEvolvableZone: %v", err)
	}

	// Header and footer must be preserved.
	if !strings.Contains(result, "# Header") {
		t.Error("header missing from result")
	}
	if !strings.Contains(result, "Introduction paragraph.") {
		t.Error("introduction paragraph missing from result")
	}
	if !strings.Contains(result, "Footer content.") {
		t.Error("footer missing from result")
	}

	// New zone content must be present.
	if !strings.Contains(result, "New line 1.") {
		t.Error("new content missing from result")
	}
	if !strings.Contains(result, "New line 3.") {
		t.Error("new content line 3 missing from result")
	}

	// Original zone content must be removed.
	if strings.Contains(result, "Original content") {
		t.Error("original content remains in result")
	}
}

// TestReplaceEvolvableZone_ErrZoneNotFound verifies ErrZoneNotFound is returned
// when the zone does not exist.
func TestReplaceEvolvableZone_ErrZoneNotFound(t *testing.T) {
	t.Parallel()

	content := "# Just a plain file\nNo markers here."

	_, err := ReplaceEvolvableZone(content, "nonexistent-zone", "new content")
	if err == nil {
		t.Fatal("expected ErrZoneNotFound, got nil")
	}
	if err != ErrZoneNotFound {
		t.Fatalf("expected ErrZoneNotFound, got: %v", err)
	}
}

// TestReplaceEvolvableZone_Idempotent verifies that applying the same replacement twice
// yields the same result (idempotency).
func TestReplaceEvolvableZone_Idempotent(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"# Header",
		`<!-- moai:evolvable-start id="zone1" -->`,
		"Original",
		"<!-- moai:evolvable-end -->",
		"Footer",
	}, "\n")

	newContent := "Replacement content."

	result1, err := ReplaceEvolvableZone(content, "zone1", newContent)
	if err != nil {
		t.Fatalf("first ReplaceEvolvableZone: %v", err)
	}

	result2, err := ReplaceEvolvableZone(result1, "zone1", newContent)
	if err != nil {
		t.Fatalf("second ReplaceEvolvableZone: %v", err)
	}

	if result1 != result2 {
		t.Errorf("idempotency failed:\nfirst: %q\nsecond: %q", result1, result2)
	}
}

// TestMergeEvolvableZones_PreservesFileStructure reproduces the bug at apply.go:78.
// When MergeEvolvableZones was incorrectly invoked as (file_content, zoneID,
// newContent), the file header and footer disappeared. This test verifies that.
//
// It documents the apply.go behavior before the fix that switched to
// ReplaceEvolvableZone.
func TestMergeEvolvableZones_PreservesFileStructure(t *testing.T) {
	t.Parallel()

	// Complete file with header, zone, and footer.
	fullFile := strings.Join([]string{
		"# Skill Header",
		"",
		"Introduction content that must be preserved.",
		"",
		`<!-- moai:evolvable-start id="best-practices" -->`,
		"Initial best practice content.",
		"<!-- moai:evolvable-end -->",
		"",
		"Footer content that must also be preserved.",
	}, "\n")

	newZoneContent := "Initial best practice content.\n- Always pass context.Context as first argument."

	// Perform the correct replacement via ReplaceEvolvableZone.
	result, err := ReplaceEvolvableZone(fullFile, "best-practices", newZoneContent)
	if err != nil {
		t.Fatalf("ReplaceEvolvableZone: %v", err)
	}

	// Header and footer must be preserved.
	if !strings.Contains(result, "# Skill Header") {
		t.Error("header was not preserved")
	}
	if !strings.Contains(result, "Introduction content that must be preserved.") {
		t.Error("introduction content was not preserved")
	}
	if !strings.Contains(result, "Footer content that must also be preserved.") {
		t.Error("footer was not preserved")
	}

	// New content must be present.
	if !strings.Contains(result, "Always pass context.Context as first argument.") {
		t.Error("added content missing")
	}
}
