package spec

import "testing"

// Card t2007 (H2 regression guard): every status named by the Status
// Transition Ownership Matrix, plus the legacy-optional `planned`, must stay a
// valid status. TestValidStatuses omits archived and rejected.
func TestIsValidStatus_MatrixAndLegacyValues(t *testing.T) {
	for _, s := range []string{
		"draft", "in-progress", "implemented", "completed",
		"superseded", "archived", "rejected", // matrix-named
		"planned", // legacy-optional
	} {
		if !IsValidStatus(s) {
			t.Errorf("IsValidStatus(%q) = false, want true", s)
		}
	}
}
