package fix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =====================================================================
// ConformDraftToScope — REQ-NS5-013 / AC-NS5-013
//
// Given a fixture draft with 3 patched subtrees (2 whose IDs ARE in
// diff_scope[] + 1 whose ID is NOT), scope-conformance returns exactly the
// 2 in-scope; the 1 out-of-scope is excluded AND a warning naming it + the
// diff_scope is logged to .moai/logs/navigator-sync.log.
// =====================================================================

func TestDraftScopeConformance(t *testing.T) {
	t.Parallel()

	// diff_scope[] carries 2 in-scope subtree IDs (REQ-NS5-013 fixture).
	diffScope := []DiffScopeEntry{
		{DocSurface: "capability-map.md", SubtreeID: "in-scope-A"},
		{DocSurface: "audit-report.json", SubtreeID: "in-scope-B"},
	}

	// The draft produced 3 subtree IDs: 2 in diff_scope + 1 over-produced
	// (simulating manager-develop drafting a subtree NOT in diff_scope[]).
	draftSubtreeIDs := []string{
		"in-scope-A",
		"in-scope-B",
		"out-of-scope-X",
	}

	inScope, excluded := ConformDraftToScope(draftSubtreeIDs, diffScope)

	if len(inScope) != 2 {
		t.Fatalf("inScope count = %d, want 2; got %v", len(inScope), inScope)
	}
	if !containsID(inScope, "in-scope-A") || !containsID(inScope, "in-scope-B") {
		t.Errorf("inScope = %v, want both in-scope-A and in-scope-B", inScope)
	}
	if len(excluded) != 1 {
		t.Fatalf("excluded count = %d, want 1; got %v", len(excluded), excluded)
	}
	if excluded[0] != "out-of-scope-X" {
		t.Errorf("excluded[0] = %q, want %q", excluded[0], "out-of-scope-X")
	}
	// Determinism: both outputs are sorted lexicographically.
	if !isSorted(inScope) {
		t.Errorf("inScope not sorted: %v", inScope)
	}
	if !isSorted(excluded) {
		t.Errorf("excluded not sorted: %v", excluded)
	}
}

// TestLogScopeExclusion_Warning verifies the excluded subtree ID + the
// diff_scope it is not in are written to .moai/logs/navigator-sync.log
// (AC-NS5-013 — "A warning naming the excluded subtree ID + the diff_scope[]
// it is not in is logged").
func TestLogScopeExclusion_Warning(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	diffScope := []DiffScopeEntry{
		{DocSurface: "capability-map.md", SubtreeID: "in-scope-A"},
		{DocSurface: "audit-report.json", SubtreeID: "in-scope-B"},
	}
	excluded := []string{"out-of-scope-X"}

	LogScopeExclusion(root, excluded, diffScope)

	logPath := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read navigator-sync.log: %v", err)
	}
	logText := string(data)

	// The warning names the excluded subtree ID.
	if !strings.Contains(logText, "out-of-scope-X") {
		t.Errorf("log missing excluded subtree ID %q; log:\n%s", "out-of-scope-X", logText)
	}
	// The warning names the diff_scope (at least one in-scope ID, so a reader
	// can see what the excluded ID is NOT in).
	if !strings.Contains(logText, "in-scope-A") {
		t.Errorf("log missing diff_scope representation (in-scope-A); log:\n%s", logText)
	}
	// The warning carries a scope-conformance marker so it is greppable.
	if !strings.Contains(logText, "scope-conformance") {
		t.Errorf("log missing scope-conformance marker; log:\n%s", logText)
	}
}

// TestConformDraftToScope_AllInScope verifies the no-over-production case:
// every draft subtree is in diff_scope[] → excluded is empty.
func TestConformDraftToScope_AllInScope(t *testing.T) {
	t.Parallel()

	diffScope := []DiffScopeEntry{
		{DocSurface: "capability-map.md", SubtreeID: "A"},
		{DocSurface: "capability-map.md", SubtreeID: "B"},
	}
	draftIDs := []string{"A", "B"}

	inScope, excluded := ConformDraftToScope(draftIDs, diffScope)

	if len(inScope) != 2 {
		t.Errorf("inScope = %v, want 2 entries", inScope)
	}
	if len(excluded) != 0 {
		t.Errorf("excluded = %v, want empty", excluded)
	}
}

// TestConformDraftToScope_EmptyInputs is the fail-open degenerate case.
func TestConformDraftToScope_EmptyInputs(t *testing.T) {
	t.Parallel()

	inScope, excluded := ConformDraftToScope(nil, nil)
	if len(inScope) != 0 || len(excluded) != 0 {
		t.Errorf("ConformDraftToScope(nil,nil) = %v, %v; want empty, empty", inScope, excluded)
	}
}

// --- helpers ---

func containsID(slice []string, id string) bool {
	for _, s := range slice {
		if s == id {
			return true
		}
	}
	return false
}

func isSorted(slice []string) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i-1] > slice[i] {
			return false
		}
	}
	return true
}
