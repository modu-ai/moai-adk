package kanban

import (
	"strconv"
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

func TestCompanionRolesAreTheThreeWorkers(t *testing.T) {
	t.Parallel()

	want := []string{"planner", "runner", "syncer"}
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
	if isCompanionRole(RoleLead) {
		t.Error(RoleLead + " must not be a companion role")
	}
	// The review role is retired (D1): a label carrying it no longer parses
	// as companion-shaped, in the bare form, the bump form, or the legacy
	// run-id form.
	for _, label := range []string{"review", "review-1", "review-tjlgt1"} {
		if isCompanionRole(label) {
			t.Errorf("review must not be a companion role: %q", label)
		}
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
		{"planner-tjlgt1", "planner", "tjlgt1", true},
		{"runner-tjlgt1", "runner", "tjlgt1", true},
		{"syncer-tjlgt1", "syncer", "tjlgt1", true},
		{"syncer-0", "syncer", "0", true},

		// The bare role is the launch form the lead announces (one machine,
		// one run — the run id no longer travels in companion names).
		{"planner", "planner", "", true},
		{"runner", "runner", "", true},
		{"syncer", "syncer", "", true},

		// A collision bump number parses back with the role intact.
		{"planner-1", "planner", "1", true},
		{"syncer-12", "syncer", "12", true},

		// Not companion-shaped: an unrelated named session must never be
		// mistaken for one, or its Stop-hook block cap changes silently.
		{"", "", "", false},
		{"planner-", "", "", false},
		{"-tjlgt1", "", "", false},
		{"leader-tjlgt1", "", "", false},
		{"planning-tjlgt1", "", "", false},
		{"oauth-migration", "", "", false},
		{"planner-TJLGT1", "", "", false},
		{"planner-tj_gt1", "", "", false},
		{"planner-tj gt1", "", "", false},

		// The retired review role (D1) is no longer companion-shaped in any
		// of its three historical forms — bare, bump, or legacy run id.
		{"review", "", "", false},
		{"review-1", "", "", false},
		{"review-tjlgt1", "", "", false},
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
// The announced form is the bare role; the suffixed forms below cover the
// collision bump (a number the launcher produces) and the legacy run-id shape
// (still parsed rather than rejected, so a stale muscle-memory launch joins as
// the right worker instead of being silently rerouted to the lead branch).
func TestCompanionLabelRoundTrips(t *testing.T) {
	t.Parallel()

	for _, want := range CompanionRoles {
		label := CompanionLabel(want)
		role, suffix, ok := SplitCompanionLabel(label)
		if !ok || role != want || suffix != "" {
			t.Errorf("bare round trip failed for %q: (%q, %q, %v)", label, role, suffix, ok)
		}
	}
}

// TestCompanionNumberLabelRoundTrips: the launcher's collision bump composes
// `<role>-<n>`, and every such label must parse back with its role intact —
// the bumped name is the address the lead dispatches to.
func TestCompanionNumberLabelRoundTrips(t *testing.T) {
	t.Parallel()

	for _, want := range CompanionRoles {
		for n := 1; n <= 3; n++ {
			label := CompanionNumberLabel(want, n)
			role, suffix, ok := SplitCompanionLabel(label)
			wantSuffix := strconv.Itoa(n)
			if !ok || role != want || suffix != wantSuffix {
				t.Errorf("bump round trip failed for %q: (%q, %q, %v), want (%q, %q, true)",
					label, role, suffix, ok, want, wantSuffix)
			}
		}
	}
}

// TestLeadLabelIsNotACompanion pins the property the launcher's name injection
// relies on: the lead label is hyphen-suffixed like a bumped companion label
// but is never recognized as a companion, because RoleLead is absent from
// CompanionRoles.
func TestLeadLabelIsNotACompanion(t *testing.T) {
	t.Parallel()

	runID := NewRunID()
	label := LeadLabel(runID)
	if want := RoleLead + "-" + runID; label != want {
		t.Errorf("LeadLabel(%q) = %q, want %q", runID, label, want)
	}
	if _, _, ok := SplitCompanionLabel(label); ok {
		t.Errorf("LeadLabel(%q) = %q reads as a companion label", runID, label)
	}
}

// TestSplitLeadLabelRoundTrip pins the property the launcher's run-id adoption
// relies on: whatever LeadLabel writes, SplitLeadLabel reads back unchanged.
func TestSplitLeadLabelRoundTrip(t *testing.T) {
	t.Parallel()

	runID := NewRunID()
	got, ok := SplitLeadLabel(LeadLabel(runID))
	if !ok || got != runID {
		t.Errorf("SplitLeadLabel(LeadLabel(%q)) = %q/%v, want %q/true", runID, got, ok, runID)
	}
}

// TestSplitLeadLabelRejectsNonLeadShapes covers what must NOT be adopted as a
// lead run id. The admitted shape is deliberately identical to the companion
// side's: `leader-notarunid` IS accepted, because `notarunid` is a well-formed
// run id as far as the grammar is concerned, exactly as `runner-notarunid` is
// accepted as a companion. Tightening one side alone is how the two branches
// drift apart.
func TestSplitLeadLabelRejectsNonLeadShapes(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		label string
		want  string // "" means reject
	}{
		{"leader-abc123", "abc123"},
		{"leader-notarunid", "notarunid"},
		{"", ""},
		{"leader", ""},
		{"leader-", ""},
		{"leader-ABC123", ""},
		{"leader-a-b", ""},
		{"leader-a_b", ""},
		// The retired `lead-` prefix (pre-t118 naming) is not the lead shape
		// — adopting it would fork the run id off a name no new launch prints.
		{"lead-abc123", ""},
		{"runner-abc123", ""},
		{"board-watch", ""},
	} {
		got, ok := SplitLeadLabel(c.label)
		if c.want == "" {
			if ok {
				t.Errorf("SplitLeadLabel(%q) = %q, want reject", c.label, got)
			}
			continue
		}
		if !ok || got != c.want {
			t.Errorf("SplitLeadLabel(%q) = %q/%v, want %q/true", c.label, got, ok, c.want)
		}
	}
}

// TestSplitLeadLabelAndCompanionAreDisjoint is the branch-safety property: no
// label can satisfy both discriminators, so recognizing a lead name can never
// reroute a lead session down the companion branch.
func TestSplitLeadLabelAndCompanionAreDisjoint(t *testing.T) {
	t.Parallel()

	runID := NewRunID()
	labels := []string{LeadLabel(runID)}
	for _, role := range CompanionRoles {
		labels = append(labels, CompanionLabel(role), CompanionNumberLabel(role, 1))
	}
	for _, label := range labels {
		_, isLead := SplitLeadLabel(label)
		_, _, isCompanion := SplitCompanionLabel(label)
		if isLead && isCompanion {
			t.Errorf("label %q satisfies both discriminators", label)
		}
	}
}
