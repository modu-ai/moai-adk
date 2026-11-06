package cli

import (
	"strings"
	"testing"
)

// Card t2021 round 2 (operator decisions R437/R438): `-l` / `--lane` is the lane entry; `-a` and `--agent`
// are REMOVED. The history: issue #1948 S46 / DEC-AF-09 made `-a` an agent entry; card t2021 round 1 (R385)
// removed -l/--lane in its favour; round 2 flipped it back. The invariant that survives every round is that
// a removed spelling is never silently absorbed or passed to the backend: it is consumed as the lane entry
// or refused with a line pointing at the entry that exists.

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
		{[]string{"--agent", "x"}, "--agent"},
		{[]string{"--agent=x"}, "--agent=x"},
	} {
		if !consumedOrRefused(t, tc.args, tc.flag) {
			t.Errorf("%v: %s silently passed through to the backend", tc.args, tc.flag)
		}
	}
}

// -l / --lane is the only lane entry (card t2021 round 2): it selects the lane branch with an automatic
// lane-<n> name and carries no declared lane count.
func TestLaneFlag_IsTheLaneEntry(t *testing.T) {
	for _, flag := range []string{"-l", "--lane"} {
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
		e, err := parseLauncherEntry([]string{flag})
		if err != nil {
			t.Fatalf("%s: %v", flag, err)
		}
		if !e.FactoryEnabled || !e.FactoryAutoNumber || e.FactoryLanes != 0 {
			t.Errorf("%s: entry %+v is not the lane entry (enabled, automatic number, no declared count)", flag, e)
		}
		if label, ok := parseFactoryLaneLabel(e.Rest); !ok || !strings.HasPrefix(label, "lane-") {
			t.Errorf("%s: Rest %v lacks a canonical lane-<n> --name", flag, e.Rest)
		}
	}
}

// -a and --agent are removed: refused, pointing at -l (and --lane), before and after the `--` marker.
func TestAgentFlag_RefusalsPointAtTheLaneEntry(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-a"}, "the -a flag was removed; use -l (or --lane)"},
		{[]string{"-a=x"}, "the -a flag was removed; use -l (or --lane)"},
		{[]string{"-f", "-a"}, "the -a flag was removed; use -l (or --lane)"},
		{[]string{"--agent"}, "the --agent flag was removed; use -l (or --lane)"},
		{[]string{"--agent=x"}, "the --agent flag was removed; use -l (or --lane)"},
		{[]string{"--agent", "foo"}, "the --agent flag was removed; use -l (or --lane)"},
		{[]string{"--", "--agent", "foo"}, "the --agent flag was removed; use --lane"},
		{[]string{"--", "--agent=foo"}, "the --agent flag was removed; use --lane"},
		{[]string{"-l", "--", "--agent", "foo"}, "the --agent flag was removed; use --lane"},
		{[]string{"-f", "-l"}, "-f/--factory and -l/--lane are two entry tokens"},
		{[]string{"-l", "x"}, "takes no argument"},
		{[]string{"-l=x"}, "takes no argument"},
		{[]string{"--lane=x"}, "takes no argument"},
	} {
		e, err := parseLauncherEntry(tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: error %v, want one containing %q", tc.args, err, tc.want)
		}
		if e.FactoryEnabled {
			t.Errorf("%v: refused but still mapped to a role: %+v", tc.args, e)
		}
	}
}

// Everything after `--` is forwarded to the backend and never read as an entry token, except the removed
// --agent spelling (refused above).
func TestAgentFlag_AfterPassthroughMarkerIsForwarded(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for _, tok := range []string{"-a", "-l", "--lane", "-L", "-A"} {
		e, err := parseLauncherEntry([]string{"--", tok})
		if err != nil {
			t.Fatalf("%s: %v", tok, err)
		}
		if e.FactoryEnabled || !containsFlag(e.Rest, tok) {
			t.Errorf("entry %+v: %s after -- must reach the backend untouched", e, tok)
		}
	}
}
