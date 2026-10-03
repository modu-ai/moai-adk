package factory

import (
	"strings"
	"testing"
	"time"
)

// TestBase36 pins the encoding NewRunID depends on. The properties that matter
// downstream are lowercase-alphanumeric output (the leader-label shape check
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

// TestLeadLabelIsBareRole pins the t133 naming change: the lead label carries
// no run id.
func TestLeadLabelIsBareRole(t *testing.T) {
	t.Parallel()

	label := LeaderLabel()
	if label != RoleLeader {
		t.Errorf("LeaderLabel() = %q, want %q", label, RoleLeader)
	}
}

// TestSplitLeadLabelRoundTrip pins the migration property the launcher relies
// on: BOTH the bare form LeaderLabel now writes AND the legacy `lead-<run-id>`
// form an operator may still be pasting read back through SplitLeaderLabel.
// Rejecting the legacy form would drop such a launch down the branch that
// treats an unrecognized name as no lead name at all, which is the misroute
// the migration exists to prevent.
func TestSplitLeadLabelRoundTrip(t *testing.T) {
	t.Parallel()

	got, ok := SplitLeaderLabel(LeaderLabel())
	if !ok || got != "" {
		t.Errorf("SplitLeaderLabel(LeaderLabel()) = %q/%v, want \"\"/true", got, ok)
	}

	runID := NewRunID()
	got, ok = SplitLeaderLabel(RoleLeader + "-" + runID)
	if !ok || got != runID {
		t.Errorf("SplitLeaderLabel(legacy %q) = %q/%v, want %q/true", RoleLeader+"-"+runID, got, ok, runID)
	}

	got, ok = SplitLeaderLabel(LeaderNumberLabel(3))
	if !ok || got != "3" {
		t.Errorf("SplitLeaderLabel(LeaderNumberLabel(3)) = %q/%v, want \"3\"/true", got, ok)
	}
}

// TestSplitLeadLabelRejectsNonLeadShapes covers what must and must not carry
// the leader shape (SPEC-ROLE-NAMING-CODE-001: the role value is `leader`).
// The bare role parses (with an empty suffix), and `leader-notarunid` IS
// accepted because `notarunid` is a well-formed suffix as far as the grammar is
// concerned.
//
// The accepted-with-empty-suffix case is why this table carries an explicit ok
// column rather than treating an empty want as rejection: the bare `leader`
// parses AND yields "", so conflating the two states would let a regression
// that rejects the bare form pass unnoticed.
//
// The legacy `lead` / `lead-<suffix>` spellings are detection values now —
// they never parse as a leader label (REQ-RNC-009).
func TestSplitLeadLabelRejectsNonLeadShapes(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		label string
		want  string
		ok    bool
	}{
		{"leader", "", true}, // the bare form the notice announces
		{"leader-1", "1", true},
		{"leader-abc123", "abc123", true},
		{"leader-notarunid", "notarunid", true},
		{"", "", false},
		{"lead", "", false},        // legacy — detection only
		{"lead-1", "", false},      // legacy
		{"lead-abc123", "", false}, // legacy
		{"leader-", "", false},
		{"leader-ABC123", "", false},
		{"leader-a-b", "", false},
		{"leader-a_b", "", false},
		{"lead-", "", false},
		{"run-abc123", "", false},
		{"board-watch", "", false},
	} {
		got, ok := SplitLeaderLabel(c.label)
		if ok != c.ok || got != c.want {
			t.Errorf("SplitLeaderLabel(%q) = %q/%v, want %q/%v", c.label, got, ok, c.want, c.ok)
		}
	}
}
