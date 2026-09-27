package cli

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestParseFactoryFlagWorkerVocabulary pins the worker-axis entry tokens and
// the retained legacy aliases: `-f worker` / `-f worker-<n>` are canonical,
// `-f agent` / `-f lane-<n>` still parse (keep-alias) and are marked legacy.
func TestParseFactoryFlagWorkerVocabulary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args       []string
		wantRole   bool
		wantLegacy bool
		wantNum    int
		wantLabel  string
	}{
		{args: []string{"-f", "worker"}, wantRole: true},
		{args: []string{"-f=worker"}, wantRole: true},
		{args: []string{"-f", "agent"}, wantRole: true, wantLegacy: true},
		{args: []string{"-f", "lane-3"}, wantNum: 3, wantLabel: "lane-3"},
		{args: []string{"--factory=lane-7"}, wantNum: 7, wantLabel: "lane-7"},
		{args: []string{"-f", "lane-3"}, wantNum: 3, wantLabel: "lane-3"},
		{args: []string{"-f", "lane"}, wantRole: true},
		// Legacy label shapes still parse at M1 (the claim refuses them);
		// the dedicated rejection wording is M2.
		{args: []string{"-f", "worker-3"}, wantNum: 3, wantLabel: "worker-3"},
	}
	for _, c := range cases {
		p, err := parseFactoryFlag(c.args)
		if err != nil {
			t.Fatalf("parseFactoryFlag(%v): %v", c.args, err)
		}
		if !p.Enabled || p.WorkerRole != c.wantRole || p.LegacyAgentToken != c.wantLegacy ||
			p.WorkerNumber != c.wantNum || p.WorkerLabel != c.wantLabel {
			t.Errorf("parseFactoryFlag(%v) = %+v, want role=%v legacy=%v num=%d label=%q",
				c.args, p, c.wantRole, c.wantLegacy, c.wantNum, c.wantLabel)
		}
	}
}

// TestFactoryFlagUsageErrorAdvertisesWorkerForms: the usage error names only
// the worker-axis forms — the legacy spellings still parse but are not taught.
func TestFactoryFlagUsageErrorAdvertisesWorkerForms(t *testing.T) {
	t.Parallel()

	_, err := parseFactoryFlag([]string{"-f", "SPEC-X-001"})
	if err == nil {
		t.Fatal("want a usage error")
	}
	msg := err.Error()
	for _, want := range []string{"-f worker", "-f worker-2"} {
		if !strings.Contains(msg, want) {
			t.Errorf("usage error %q missing %q", msg, want)
		}
	}
	for _, banned := range []string{"-f agent", "lane-"} {
		if strings.Contains(msg, banned) {
			t.Errorf("usage error %q still advertises %q", msg, banned)
		}
	}
}

// TestParseLauncherEntryDesugarsWorkerLabel: `-f worker-<n>` desugars into
// the same --name channel the legacy `-f lane-<n>` used.
func TestParseLauncherEntryDesugarsWorkerLabel(t *testing.T) {
	t.Parallel()

	p, err := parseLauncherEntry([]string{"-f", "worker-2", "-b"})
	if err != nil {
		t.Fatalf("parseLauncherEntry(-f worker-2): %v", err)
	}
	if !slices.Equal(p.Rest, []string{"-b", "--name", "worker-2"}) {
		t.Errorf("desugared rest = %v, want [-b --name worker-2]", p.Rest)
	}
	if label, ok := parseFactoryLaneLabel(p.Rest); !ok || label != "worker-2" {
		t.Errorf("worker label not recognised in %v: (%q, %v)", p.Rest, label, ok)
	}
	if _, err := parseLauncherEntry([]string{"-f", "worker-2", "--name", "worker-3"}); err == nil ||
		!strings.Contains(err.Error(), "-f worker-<n> already names the worker") {
		t.Errorf("worker label plus --name = %v, want the naming conflict", err)
	}
}

// TestResolveFactoryWorkerNameRefusesLegacyLabel: a legacy label on the
// input path is refused with an error naming the canonical lane-<n>
// (REQ-RNC-009; the dedicated rejection wording is M2).
func TestResolveFactoryWorkerNameRefusesLegacyLabel(t *testing.T) {
	cases := []struct{ label, want string }{
		{"worker-4", "lane-4"},
		{"agent-2", "lane-2"},
	}
	for _, c := range cases {
		var notes bytes.Buffer
		if got, err := resolveFactoryWorkerName(t.TempDir(), c.label, false, &notes); err == nil {
			t.Fatalf("resolve %s = %q, want an error naming %s", c.label, got, c.want)
		} else if !strings.Contains(err.Error(), c.want) {
			t.Errorf("resolve %s error %q lacks the canonical form %s", c.label, err.Error(), c.want)
		}
	}

	var notes bytes.Buffer
	if got, err := resolveFactoryWorkerName(t.TempDir(), "lane-1", false, &notes); err != nil || got != "lane-1" || notes.Len() != 0 {
		t.Errorf("canonical free label = (%q, %v, notes %q), want lane-1 with no note", got, err, notes.String())
	}
}

// NextFactoryWorkerNumberForTest delegates to the kanban SSOT.
func NextFactoryWorkerNumberForTest(reg map[string]kanban.FactoryWorkerEntry, alive func(int) bool) int {
	return kanban.NextFactoryWorkerNumber(reg, alive)
}

// TestNextFactoryWorkerNumber: the lane join takes one past the highest LIVE
// canonical claim. Legacy claims hold no number — a live legacy record
// refuses the join instead (design §4).
func TestNextFactoryWorkerNumber(t *testing.T) {
	alive := func(int) bool { return true }
	if n := NextFactoryWorkerNumberForTest(map[string]kanban.FactoryWorkerEntry{}, alive); n != 1 {
		t.Errorf("empty registry = %d, want 1", n)
	}
	reg := map[string]kanban.FactoryWorkerEntry{
		"lane-1":   {PID: 100},
		"agent-2":  {PID: 101}, // legacy row — holds no number
		"worker-5": {PID: 102}, // legacy row — holds no number
	}
	if n := NextFactoryWorkerNumberForTest(reg, alive); n != 2 {
		t.Errorf("registry up to lane-1 = %d, want 2", n)
	}
	dead := func(int) bool { return false }
	if n := NextFactoryWorkerNumberForTest(reg, dead); n != 1 {
		t.Errorf("dead claims pruned = %d, want 1", n)
	}
}

// TestLauncherHelpAdvertisesWorkerVocabulary: the cc and glm help surfaces
// teach the worker forms and no longer teach `-f agent` / `lane-<n>`.
func TestLauncherHelpAdvertisesWorkerVocabulary(t *testing.T) {
	t.Parallel()

	for name, text := range map[string]string{
		"cc":  ccCmd.Use + "\n" + ccCmd.Long,
		"glm": glmCmd.Use + "\n" + glmCmd.Long,
	} {
		for _, want := range []string{"-f worker-<n>", "worker-1"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s help missing %q", name, want)
			}
		}
		for _, banned := range []string{"-f agent", "lane-<", "lane-1", "lane-2", "lane-3"} {
			if strings.Contains(text, banned) {
				t.Errorf("%s help still advertises %q", name, banned)
			}
		}
	}
	if !strings.Contains(ccCmd.Long, "-f worker ") {
		t.Errorf("cc help missing the `-f worker` role-token entry")
	}
}
