package cli

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestParseFactoryFlagLaneVocabulary pins the lane-axis entry tokens
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-002): `-f lane` / `-f lane-<n>` are the
// only accepted role shapes. The legacy spellings are refused — their
// refusal matrix lives in factory_role_refusal_m2_test.go.
func TestParseFactoryFlagLaneVocabulary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args      []string
		wantRole  bool
		wantNum   int
		wantLabel string
	}{
		{args: []string{"-f", "lane"}, wantRole: true},
		{args: []string{"-f=lane"}, wantRole: true},
		{args: []string{"--factory=lane"}, wantRole: true},
		{args: []string{"-f", "lane-3"}, wantNum: 3, wantLabel: "lane-3"},
		{args: []string{"--factory=lane-7"}, wantNum: 7, wantLabel: "lane-7"},
	}
	for _, c := range cases {
		p, err := parseFactoryFlag(c.args)
		if err != nil {
			t.Fatalf("parseFactoryFlag(%v): %v", c.args, err)
		}
		if !p.Enabled || p.LaneRole != c.wantRole ||
			p.LaneNumber != c.wantNum || p.LaneLabel != c.wantLabel {
			t.Errorf("parseFactoryFlag(%v) = %+v, want role=%v num=%d label=%q",
				c.args, p, c.wantRole, c.wantNum, c.wantLabel)
		}
	}
}

// TestFactoryFlagUsageErrorAdvertisesLaneForms: the usage error names only
// the lane-axis forms — the legacy spellings are refused, never taught
// (REQ-RNC-001).
func TestFactoryFlagUsageErrorAdvertisesLaneForms(t *testing.T) {
	t.Parallel()

	_, err := parseFactoryFlag([]string{"-f", "SPEC-X-001"})
	if err == nil {
		t.Fatal("want a usage error")
	}
	msg := err.Error()
	for _, want := range []string{"-f lane", "-f lane-2"} {
		if !strings.Contains(msg, want) {
			t.Errorf("usage error %q missing %q", msg, want)
		}
	}
	for _, banned := range []string{"-f agent", "worker"} {
		if strings.Contains(msg, banned) {
			t.Errorf("usage error %q still advertises %q", msg, banned)
		}
	}
}

// TestParseLauncherEntryDesugarsLaneLabel: `-f lane-<n>` desugars into the
// --name channel the launch branches read, and a lane label plus an
// operator --name is the naming conflict.
func TestParseLauncherEntryDesugarsLaneLabel(t *testing.T) {
	t.Parallel()

	p, err := parseLauncherEntry([]string{"-f", "lane-2", "-b"})
	if err != nil {
		t.Fatalf("parseLauncherEntry(-f lane-2): %v", err)
	}
	if !slices.Equal(p.Rest, []string{"-b", "--name", "lane-2"}) {
		t.Errorf("desugared rest = %v, want [-b --name lane-2]", p.Rest)
	}
	if label, ok := parseFactoryLaneLabel(p.Rest); !ok || label != "lane-2" {
		t.Errorf("lane label not recognised in %v: (%q, %v)", p.Rest, label, ok)
	}
	if _, err := parseLauncherEntry([]string{"-f", "lane-2", "--name", "lane-3"}); err == nil ||
		!strings.Contains(err.Error(), "-f lane-<n> already names the lane") {
		t.Errorf("lane label plus --name = %v, want the naming conflict", err)
	}
}

// TestResolveFactoryWorkerNameRefusesLegacyLabel: a legacy label on the
// input path is refused with an error naming the canonical lane-<n>
// (REQ-RNC-009; the claim keeps the legacy refusal as the library-level
// defense behind the entry-parse refusal).
func TestResolveFactoryWorkerNameRefusesLegacyLabel(t *testing.T) {
	cases := []struct{ label, want string }{
		{"worker-4", "lane-4"},
		{"agent-2", "lane-2"},
	}
	for _, c := range cases {
		var notes bytes.Buffer
		if got, err := resolveFactoryLaneName(t.TempDir(), c.label, false, &notes); err == nil {
			t.Fatalf("resolve %s = %q, want an error naming %s", c.label, got, c.want)
		} else if !strings.Contains(err.Error(), c.want) {
			t.Errorf("resolve %s error %q lacks the canonical form %s", c.label, err.Error(), c.want)
		}
	}

	var notes bytes.Buffer
	if got, err := resolveFactoryLaneName(t.TempDir(), "lane-1", false, &notes); err != nil || got != "lane-1" || notes.Len() != 0 {
		t.Errorf("canonical free label = (%q, %v, notes %q), want lane-1 with no note", got, err, notes.String())
	}
}

// NextFactoryLaneNumberForTest delegates to the kanban SSOT.
func NextFactoryLaneNumberForTest(reg map[string]kanban.FactoryLaneEntry, alive func(int) bool) int {
	return kanban.NextFactoryLaneNumber(reg, alive)
}

// TestNextFactoryWorkerNumber: the lane join takes one past the highest LIVE
// canonical claim. Legacy claims hold no number — a live legacy record
// refuses the join instead (design §4).
func TestNextFactoryWorkerNumber(t *testing.T) {
	alive := func(int) bool { return true }
	if n := NextFactoryLaneNumberForTest(map[string]kanban.FactoryLaneEntry{}, alive); n != 1 {
		t.Errorf("empty registry = %d, want 1", n)
	}
	reg := map[string]kanban.FactoryLaneEntry{
		"lane-1":   {PID: 100},
		"agent-2":  {PID: 101}, // legacy row — holds no number
		"worker-5": {PID: 102}, // legacy row — holds no number
	}
	if n := NextFactoryLaneNumberForTest(reg, alive); n != 2 {
		t.Errorf("registry up to lane-1 = %d, want 2", n)
	}
	dead := func(int) bool { return false }
	if n := NextFactoryLaneNumberForTest(reg, dead); n != 1 {
		t.Errorf("dead claims pruned = %d, want 1", n)
	}
}
