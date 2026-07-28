package merge

import "testing"

// TestMergeAnalysis_Creation pins the field set of the two analysis types that
// survive outside this package. It was retargeted here from the deleted
// confirm_test.go: the test never exercised the removed confirmation program,
// only the struct shape that internal/cli/update.go, internal/cli/update_tux.go,
// internal/cli/update/merge, and internal/cli/update/plan depend on.
func TestMergeAnalysis_Creation(t *testing.T) {
	t.Parallel()

	analysis := MergeAnalysis{
		Files: []FileAnalysis{
			{
				Path:      "test.yaml",
				Changes:   "update",
				Strategy:  YAMLDeep,
				RiskLevel: "medium",
			},
		},
		HasConflicts: true,
		SafeToMerge:  false,
		Summary:      "Test merge",
		RiskLevel:    "high",
	}

	if len(analysis.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(analysis.Files))
	}

	if !analysis.HasConflicts {
		t.Error("Expected HasConflicts to be true")
	}

	if analysis.SafeToMerge {
		t.Error("Expected SafeToMerge to be false")
	}
}

// TestFileAnalysis_NamedFieldLiteral mirrors the composite literal built by
// internal/cli/update/plan, so a field rename fails here as well as at that
// call site.
func TestFileAnalysis_NamedFieldLiteral(t *testing.T) {
	t.Parallel()

	fa := FileAnalysis{
		Path:      ".claude/settings.json",
		Changes:   "modified",
		Strategy:  JSONMerge,
		RiskLevel: "low",
		Note:      "",
	}

	if fa.Path == "" || fa.Changes == "" || fa.Strategy == "" || fa.RiskLevel == "" {
		t.Errorf("FileAnalysis literal did not populate its field set: %+v", fa)
	}
}
