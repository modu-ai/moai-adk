package cli

import (
	"strings"
	"testing"
)

// Issue #1948 S46 / DEC-AF-09: `-a` / `--agent` is the agent entry, the design name for the lane role. Before
// the implementation it was silently passed through to the backend (observed: err=nil, no factory entry, the
// token stayed in Rest; beside -f or -l it stayed in Rest too). The expected behavior is "consumed as the agent
// entry, or an explicit error", never a silent pass-through; this test was RED behind the afred tag until the
// entry landed.

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
		{[]string{"--agent"}, "--agent"},
		{[]string{"-f", "-a"}, "-a"},
		{[]string{"-l", "-a"}, "-a"},
	} {
		if !consumedOrRefused(t, tc.args, tc.flag) {
			t.Errorf("%v: %s silently passed through to the backend", tc.args, tc.flag)
		}
	}
}

// The agent entry is the lane entry under another spelling: it selects the lane branch and the same
// factory entry fields, and beside another entry token it is refused with a message naming both tokens.
func TestAgentFlag_IsTheLaneEntryUnderAnotherSpelling(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	lane, err := parseLauncherEntry([]string{"-l"})
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-a"}, {"--agent"}} {
		e, err := parseLauncherEntry(args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !e.FactoryEnabled || !e.FactoryAutoNumber || e.FactoryLanes != lane.FactoryLanes {
			t.Errorf("%v: entry %+v does not match the lane entry %+v", args, e, lane)
		}
		if strings.Join(e.Rest, " ") != strings.Join(lane.Rest, " ") {
			t.Errorf("%v: Rest %v, want the lane form %v", args, e.Rest, lane.Rest)
		}
	}
}

func TestAgentFlag_RefusalsNameBothTokens(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-f", "-a"}, "-a/--agent"},
		{[]string{"-l", "-a"}, "-l/--lane"},
		{[]string{"-a", "x"}, "takes no argument"},
		{[]string{"--agent=x"}, "takes no argument"},
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
