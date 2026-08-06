package astx

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestSpecIDsFromCapabilityMap_ReadsSpecIDColumn is the Bug-2 repro: a
// lightweight helper that returns the sorted-unique `spec-id` column values
// from a capability-map.md, WITHOUT calling astx.EnrichRows (which runs the
// full enrichment pipeline: per-row `filepath.WalkDir` of the
// implementation-path + tree-sitter `Extract()` on every file). The helper
// reuses the existing parseCapabilityMap parser and reads only the table.
//
// The fixture points implementation-path at a NON-existent directory; the
// helper still returns the spec-ids because it never walks the paths.
func TestSpecIDsFromCapabilityMap_ReadsSpecIDColumn(t *testing.T) {
	root := t.TempDir()
	capMap := filepath.Join(root, "capability-map.md")
	// implementation-path entries point at directories that do NOT exist on
	// disk; the helper must still return the spec-ids (it does not walk them).
	// Includes a duplicate spec-id and an empty spec-id row to exercise
	// de-duplication and the empty-skip.
	if err := os.WriteFile(capMap, []byte(
		"# Capability Map\n\n"+
			"| spec-id | title | implementation-path |\n"+
			"|---------|-------|---------------------|\n"+
			"| SPEC-A-001 | A | src/does-not-exist-a |\n"+
			"| SPEC-B-002 | B | src/does-not-exist-b |\n"+
			"| SPEC-A-001 | A dup | src/does-not-exist-a |\n"+
			"|   | empty | src/whatever |\n"+
			"| SPEC-C-003 | C | src/does-not-exist-c |\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := SpecIDsFromCapabilityMap(capMap)
	if err != nil {
		t.Fatalf("SpecIDsFromCapabilityMap error: %v", err)
	}
	// Sorted-unique, with the empty row skipped and the duplicate collapsed.
	want := []string{"SPEC-A-001", "SPEC-B-002", "SPEC-C-003"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SpecIDsFromCapabilityMap = %v, want %v", got, want)
	}
}

// TestSpecIDsFromCapabilityMap_AbsentFileReturnsEmpty covers the fail-open
// path: a missing capability-map file returns an empty slice + nil error
// (mirrors EnrichRows's absent-file behavior, REQ-NT-002 fail-open).
func TestSpecIDsFromCapabilityMap_AbsentFileReturnsEmpty(t *testing.T) {
	got, err := SpecIDsFromCapabilityMap(filepath.Join(t.TempDir(), "missing.md"))
	if err != nil {
		t.Errorf("absent file returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("absent file returned %d spec-ids, want 0: %v", len(got), got)
	}
}
