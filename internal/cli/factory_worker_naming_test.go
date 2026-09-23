package cli

import (
	"bytes"
	"slices"
	"strings"
	"testing"
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
		{args: []string{"-f", "worker-3"}, wantNum: 3, wantLabel: "worker-3"},
		{args: []string{"--factory=worker-7"}, wantNum: 7, wantLabel: "worker-7"},
		{args: []string{"-f", "lane-3"}, wantNum: 3, wantLabel: "lane-3"},
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

// TestResolveFactoryWorkerNameCanonicalizesLegacyWithHint: a legacy label
// (typed `-f lane-<n>` / `--name lane-<n>`, or the `-f agent` desugar) is
// launched under the canonical `worker-<n>` and the operator is told the
// new spelling — the alias is kept, never silent.
func TestResolveFactoryWorkerNameCanonicalizesLegacyWithHint(t *testing.T) {
	cases := []struct {
		label, want, hint string
	}{
		{"lane-4", "worker-4", "-f worker-<n>"},
		{"agent-2", "worker-2", "-f worker"},
	}
	for _, c := range cases {
		var notes bytes.Buffer
		got, err := resolveFactoryWorkerName(t.TempDir(), c.label, &notes)
		if err != nil || got != c.want {
			t.Fatalf("resolve %s = (%q, %v), want %s", c.label, got, err, c.want)
		}
		out := notes.String()
		if !strings.Contains(out, "deprecated") || !strings.Contains(out, c.hint) || !strings.Contains(out, c.want) {
			t.Errorf("resolve %s notes = %q, want a deprecation hint naming %q and %q", c.label, out, c.hint, c.want)
		}
	}

	var notes bytes.Buffer
	if got, err := resolveFactoryWorkerName(t.TempDir(), "worker-1", &notes); err != nil || got != "worker-1" || notes.Len() != 0 {
		t.Errorf("canonical free label = (%q, %v, notes %q), want worker-1 with no note", got, err, notes.String())
	}
}

// TestStripCodexFactoryFlagWorkerVocabulary: the codex door accepts the same
// worker tokens and keeps the legacy role token as an alias.
func TestStripCodexFactoryFlagWorkerVocabulary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		head     []string
		wantRole string
		wantLane string
	}{
		{head: []string{"-f", "worker"}, wantRole: factoryWorkerRoleToken},
		{head: []string{"-f", "agent"}, wantRole: factoryLegacyAgentRoleToken},
		{head: []string{"-f", "worker-3"}, wantLane: "worker-3"},
		{head: []string{"-f", "lane-3"}, wantLane: "lane-3"},
	}
	for _, c := range cases {
		_, lead, role, lane, err := stripCodexFactoryFlag(c.head)
		if err != nil || lead || role != c.wantRole || lane != c.wantLane {
			t.Errorf("stripCodexFactoryFlag(%v) = (lead %v, role %q, lane %q, %v), want (false, %q, %q, nil)",
				c.head, lead, role, lane, err, c.wantRole, c.wantLane)
		}
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
