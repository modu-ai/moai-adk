package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestParseFactoryFlagLaneVocabulary pins the lane-axis entry tokens
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-002, re-pinned by SPEC-LAUNCHER-ENTRY-
// FLAGS-001): `-l` / `--lane` are the only lane entry and `-f` takes no value,
// so every `-f lane` spelling is refused. The legacy spellings are refused
// too — their refusal matrix lives in factory_role_refusal_m2_test.go.
func TestParseFactoryFlagLaneVocabulary(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"-l"}, {"--lane"}} {
		p, err := parseFactoryFlag(args)
		if err != nil {
			t.Fatalf("parseFactoryFlag(%v): %v", args, err)
		}
		if !p.Enabled || !p.LaneRole || p.LaneNumber != 0 || p.LaneLabel != "" {
			t.Errorf("parseFactoryFlag(%v) = %+v, want a lane entry carrying no number", args, p)
		}
	}
	for _, args := range [][]string{
		{"-f", "lane"}, {"-f=lane"}, {"--factory=lane"}, {"-f", "lane-3"}, {"--factory=lane-7"},
	} {
		if _, err := parseFactoryFlag(args); err == nil {
			t.Errorf("parseFactoryFlag(%v) = nil error, want the refusal naming -l", args)
		}
	}
}

// TestFactoryFlagUsageErrorAdvertisesLaneForms: the usage error names the lane
// entry and the bare leader form — the legacy spellings and the removed lane
// forms are refused, never taught (REQ-RNC-001).
func TestFactoryFlagUsageErrorAdvertisesLaneForms(t *testing.T) {
	t.Parallel()

	_, err := parseFactoryFlag([]string{"-f", "SPEC-X-001"})
	if err == nil {
		t.Fatal("want a usage error")
	}
	msg := err.Error()
	for _, want := range []string{"-l", "--lane", "bare -f"} {
		if !strings.Contains(msg, want) {
			t.Errorf("usage error %q missing %q", msg, want)
		}
	}
	for _, banned := range []string{"-f agent", "worker", "-f lane", "lane-2"} {
		if strings.Contains(msg, banned) {
			t.Errorf("usage error %q still advertises %q", msg, banned)
		}
	}
}

// TestParseLauncherEntryDesugarsLaneLabel: `-l` desugars into the --name
// channel the launch branches read (the next free lane-<n>), and an operator
// --name beside it is the naming conflict.
func TestParseLauncherEntryDesugarsLaneLabel(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())

	p, err := parseLauncherEntry([]string{"-l", "-b"})
	if err != nil {
		t.Fatalf("parseLauncherEntry(-l): %v", err)
	}
	label, ok := parseFactoryLaneLabel(p.Rest)
	if !ok {
		t.Fatalf("lane label not recognised in %v", p.Rest)
	}
	if len(p.Rest) != 3 || p.Rest[0] != "-b" || p.Rest[1] != "--name" || p.Rest[2] != label {
		t.Errorf("desugared rest = %v, want [-b --name %s]", p.Rest, label)
	}
	if _, err := parseLauncherEntry([]string{"-l", "--name", "lane-3"}); err == nil ||
		!strings.Contains(err.Error(), "already names the role") {
		t.Errorf("-l plus --name = %v, want the naming conflict", err)
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
func NextFactoryLaneNumberForTest(reg map[string]factory.FactoryLaneEntry, alive func(int) bool) int {
	return factory.NextFactoryLaneNumber(reg, alive)
}

// TestNextFactoryWorkerNumber: the lane join takes one past the highest LIVE
// canonical claim. Legacy claims hold no number — a live legacy record
// refuses the join instead (design §4).
func TestNextFactoryWorkerNumber(t *testing.T) {
	alive := func(int) bool { return true }
	if n := NextFactoryLaneNumberForTest(map[string]factory.FactoryLaneEntry{}, alive); n != 1 {
		t.Errorf("empty registry = %d, want 1", n)
	}
	reg := map[string]factory.FactoryLaneEntry{
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
