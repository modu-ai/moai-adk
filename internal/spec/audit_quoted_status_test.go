package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// Card t1119 — a YAML-quoted `status:` value (`status: "completed"`) is the same
// value as the bare form. The status line is read with a regexp, so the quotes
// must be normalized away before the value is compared; otherwise a correctly
// closed SPEC reads as `"completed"` != completed and SyncStatusDrift fires
// MUST-FIX on it.

// A closed SPEC whose status is quoted (double or single) with full close
// evidence must not be reported as SyncStatusDrift.
func TestAudit_SyncStatusDrift_QuotedCompletedIsClean(t *testing.T) {
	t.Parallel()
	for _, status := range []string{`"completed"`, `'completed'`} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			id := "SPEC-V3R6-QUOTED-DONE-001"
			if f := syncStatusDriftFor(t, id, makeAmendmentSpecMD(id, status, "", "")); f != nil {
				t.Errorf("SyncStatusDrift fired on a closed SPEC with status %s: details=%v", status, f.Details)
			}
		})
	}
}

// A quoted `status: "in-progress"` on an otherwise valid in-place amendment is
// the sanctioned amendment state and must be exempt, same as the bare form.
func TestAudit_SyncStatusDrift_QuotedInProgressAmendmentExempt(t *testing.T) {
	t.Parallel()
	for _, status := range []string{`"in-progress"`, `'in-progress'`} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			id := "SPEC-V3R6-QUOTED-AMEND-001"
			if f := syncStatusDriftFor(t, id, makeAmendmentSpecMD(id, status, id, "### Amendments")); f != nil {
				t.Errorf("SyncStatusDrift fired on a valid amendment with status %s: details=%v", status, f.Details)
			}
		})
	}
}

// Negative control: normalization must not turn a quoted non-completed status
// into a clean one. Full close evidence + quoted `implemented` is genuine drift,
// with or without an amendment record.
func TestAudit_SyncStatusDrift_QuotedNonCompletedStillFires(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, status, amendmentOf, heading string
	}{
		{"double-quoted implemented", `"implemented"`, "", ""},
		{"single-quoted implemented", `'implemented'`, "", ""},
		{"quoted implemented with amendment record", `"implemented"`, "SPEC-V3R6-QUOTED-DRIFT-001", "### Amendments"},
		{"quoted in-progress without amendment_of", `"in-progress"`, "", "### Amendments"},
		{"mismatched quotes are not stripped", `"completed'`, "", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id := "SPEC-V3R6-QUOTED-DRIFT-001"
			f := syncStatusDriftFor(t, id, makeAmendmentSpecMD(id, tc.status, tc.amendmentOf, tc.heading))
			requireMustFixSyncStatusDrift(t, f, tc.name)
		})
	}
}

// The close path reads the same status line; a quoted value must load as the
// bare value so close decisions (already-completed / implemented gate) hold.
func TestLoadSpecCloseState_QuotedStatusNormalized(t *testing.T) {
	t.Parallel()
	for status, want := range map[string]string{
		`"completed"`:   "completed",
		`'implemented'`: "implemented",
		`completed`:     "completed",
	} {
		status, want := status, want
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			id := "SPEC-V3R6-QUOTED-CLOSE-001"
			if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(makeAmendmentSpecMD(id, status, "", "")), 0o644); err != nil {
				t.Fatal(err)
			}
			state, err := loadSpecCloseState(dir, id)
			if err != nil {
				t.Fatalf("loadSpecCloseState: %v", err)
			}
			if state.SpecMDStatus != want {
				t.Errorf("SpecMDStatus = %q, want %q", state.SpecMDStatus, want)
			}
		})
	}
}
