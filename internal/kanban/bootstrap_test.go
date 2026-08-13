package kanban

import (
	"strings"
	"testing"
	"time"
)

// TestBase36 pins the encoding NewRunID depends on. The properties that matter
// downstream are lowercase-alphanumeric output (the companion-label shape check
// rejects anything else) and monotonicity (a later run must sort after an
// earlier one).
func TestBase36(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{-1, "0"},
		{1, "1"},
		{9, "9"},
		{10, "a"},
		{35, "z"},
		{36, "10"},
		{1295, "zz"},  // 36^2 - 1
		{1296, "100"}, // 36^2
	}
	for _, c := range cases {
		if got := base36(c.in); got != c.want {
			t.Errorf("base36(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBase36IsMonotonicAndRunIDShaped(t *testing.T) {
	t.Parallel()

	prev := ""
	for _, sec := range []int64{1_000_000_000, 1_700_000_000, 1_900_000_000} {
		got := base36(sec)
		if !isRunIDShape(got) {
			t.Errorf("base36(%d) = %q is not run-id shaped", sec, got)
		}
		if got != strings.ToLower(got) {
			t.Errorf("base36(%d) = %q is not lowercase", sec, got)
		}
		// Longer wins; equal length compares lexically, which for a fixed-width
		// base36 string is the same order as the underlying number.
		ascending := len(got) > len(prev) || (len(got) == len(prev) && got > prev)
		if prev != "" && !ascending {
			t.Errorf("base36 not monotonic: %q then %q", prev, got)
		}
		prev = got
	}
}

// TestNewRunIDMatchesCurrentSecond asserts NewRunID is the base36 of the current
// Unix second, so an operator can correlate a run id with when it started.
func TestNewRunIDMatchesCurrentSecond(t *testing.T) {
	t.Parallel()

	before := time.Now().Unix()
	got := NewRunID()
	after := time.Now().Unix()

	if got != base36(before) && got != base36(after) {
		t.Errorf("NewRunID() = %q, want base36 of a second in [%d, %d]", got, before, after)
	}
	if !isRunIDShape(got) {
		t.Errorf("NewRunID() = %q is not run-id shaped", got)
	}
}

func TestCompanionRolesAreTheFourWorkers(t *testing.T) {
	t.Parallel()

	want := []string{"plan", "run", "review", "sync"}
	if len(CompanionRoles) != len(want) {
		t.Fatalf("CompanionRoles = %v, want %v", CompanionRoles, want)
	}
	for i := range want {
		if CompanionRoles[i] != want[i] {
			t.Errorf("CompanionRoles[%d] = %q, want %q", i, CompanionRoles[i], want[i])
		}
	}
	// The lead is not a companion: it is the only session carrying the kanban
	// token, and listing it here would invite a second chain driver.
	if isCompanionRole("lead") {
		t.Error("lead must not be a companion role")
	}
}

func TestSplitCompanionLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		label     string
		wantRole  string
		wantRunID string
		wantOK    bool
	}{
		{"plan-tjlgt1", "plan", "tjlgt1", true},
		{"run-tjlgt1", "run", "tjlgt1", true},
		{"review-tjlgt1", "review", "tjlgt1", true},
		{"sync-tjlgt1", "sync", "tjlgt1", true},
		{"sync-0", "sync", "0", true},

		// Not companion-shaped: an unrelated named session must never be
		// mistaken for one, or its Stop-hook block cap changes silently.
		{"", "", "", false},
		{"plan", "", "", false},
		{"plan-", "", "", false},
		{"-tjlgt1", "", "", false},
		{"lead-tjlgt1", "", "", false},
		{"planning-tjlgt1", "", "", false},
		{"oauth-migration", "", "", false},
		{"plan-TJLGT1", "", "", false},
		{"plan-tj_gt1", "", "", false},
		{"plan-tj gt1", "", "", false},
	}
	for _, c := range cases {
		role, runID, ok := SplitCompanionLabel(c.label)
		if ok != c.wantOK || role != c.wantRole || runID != c.wantRunID {
			t.Errorf("SplitCompanionLabel(%q) = (%q, %q, %v), want (%q, %q, %v)",
				c.label, role, runID, ok, c.wantRole, c.wantRunID, c.wantOK)
		}
	}
}

// TestCompanionLabelRoundTrips asserts the two halves of the label vocabulary
// agree — every label the lead announces must parse back on the companion side.
func TestCompanionLabelRoundTrips(t *testing.T) {
	t.Parallel()

	runID := NewRunID()
	for _, want := range CompanionRoles {
		label := CompanionLabel(want, runID)
		role, gotID, ok := SplitCompanionLabel(label)
		if !ok || role != want || gotID != runID {
			t.Errorf("round trip failed for %q: (%q, %q, %v)", label, role, gotID, ok)
		}
	}
}
