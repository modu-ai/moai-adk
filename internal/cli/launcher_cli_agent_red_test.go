package cli

import (
	"reflect"
	"strings"
	"testing"
)

// Issue #1948 S46 / DEC-AF-09: `-a` is the agent entry, the short spelling of the lane role. Before the
// implementation it was silently passed through to the backend (observed: err=nil, no factory entry, the token
// stayed in Rest; beside -f or -l it stayed in Rest too). The expected behavior is "consumed as the agent
// entry, or an explicit error", never a silent pass-through; this test was RED behind the afred tag until the
// entry landed. The long `--agent <name>` is NOT an entry token: Claude Code has its own --agent option and a
// user who passes it keeps it (leader decision), so it reaches the backend unchanged.

func consumedOrRefused(t *testing.T, args []string, flag string) bool {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	e, err := parseLauncherEntry(args)
	if err != nil {
		return true
	}
	return !containsFlag(e.Rest, flag)
}

func TestRED_AgentFlag_NotSilentlyAbsorbed(t *testing.T) {
	for _, tc := range []struct {
		args []string
		flag string
	}{
		{[]string{"-a"}, "-a"},
		{[]string{"-a", "x"}, "-a"},
		{[]string{"-f", "-a"}, "-a"},
		{[]string{"-l", "-a"}, "-a"},
	} {
		if !consumedOrRefused(t, tc.args, tc.flag) {
			t.Errorf("%v: %s silently passed through to the backend", tc.args, tc.flag)
		}
	}
}

// -a is the only lane entry (card t2021 removed -l/--lane): it selects the lane branch with an automatic
// lane-<n> name and carries no declared lane count.
func TestAgentFlag_ShortFormIsTheLaneEntry(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	e, err := parseLauncherEntry([]string{"-a"})
	if err != nil {
		t.Fatal(err)
	}
	if !e.FactoryEnabled || !e.FactoryAutoNumber || e.FactoryLanes != 0 {
		t.Errorf("entry %+v is not the lane entry (enabled, automatic number, no declared count)", e)
	}
	if label, ok := parseFactoryLaneLabel(e.Rest); !ok || !strings.HasPrefix(label, "lane-") {
		t.Errorf("Rest %v lacks a canonical lane-<n> --name", e.Rest)
	}
}

// The long `--agent <name>` reaches Claude Code untouched: not an entry, not consumed, not refused, with or
// without a value and with or without the `--` marker.
func TestAgentFlag_LongFormPassesThroughToTheBackend(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, args := range [][]string{
		{"--agent", "foo"},
		{"--agent=foo"},
		{"--", "--agent", "foo"},
	} {
		e, err := parseLauncherEntry(args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if e.FactoryEnabled {
			t.Errorf("%v: treated as a factory entry", args)
		}
		if !reflect.DeepEqual(e.Rest, args) {
			t.Errorf("%v: Rest = %v, want unchanged", args, e.Rest)
		}
	}
}

func TestAgentFlag_RefusalsNameBothTokens(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-f", "-a"}, "-a"},
		{[]string{"-l", "-a"}, "the -l flag was removed"},
		{[]string{"--lane"}, "the --lane flag was removed"},
		{[]string{"-a", "x"}, "takes no argument"},
		{[]string{"-a=x"}, "takes no argument"},
	} {
		_, err := parseLauncherEntry(tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: error %v, want one containing %q", tc.args, err, tc.want)
		}
	}
}

// Everything after `--` is forwarded to the backend and never read as an entry token.
func TestAgentFlag_AfterPassthroughMarkerIsForwarded(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	e, err := parseLauncherEntry([]string{"--", "-a"})
	if err != nil {
		t.Fatal(err)
	}
	if e.FactoryEnabled || !containsFlag(e.Rest, "-a") {
		t.Errorf("entry %+v: -a after -- must reach the backend untouched", e)
	}
}
