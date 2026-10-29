package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Issue #1948 (DELTA4): characterization of the six legacy launch forms. The
// role-flag redesign is undecided, so these pin what the entry parse does
// today at rc.38 + this tree; they change no behavior. Pure argv parsing: no
// claude binary is launched and the lane registry lives in a temp project dir.

func legacyEntry(t *testing.T, args ...string) (launcherEntryParse, factoryBranch) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	entry, err := parseLauncherEntry(args)
	if err != nil {
		t.Fatalf("parseLauncherEntry(%v): %v", args, err)
	}
	_, isLane := parseFactoryLaneLabel(entry.Rest)
	return entry, resolveFactoryBranch(entry.FactoryEnabled, isLane)
}

func TestLegacyForms_CcF_IsFactoryLeader(t *testing.T) {
	e, branch := legacyEntry(t, "-f")
	if !e.FactoryEnabled || branch != factoryBranchLeader {
		t.Fatalf("enabled=%v branch=%v, want leader", e.FactoryEnabled, branch)
	}
	if e.FactoryLanes != config.DefaultFactoryLeaderLanes {
		t.Errorf("lanes = %d, want %d", e.FactoryLanes, config.DefaultFactoryLeaderLanes)
	}
	if e.FactoryAutoNumber || e.FactoryLead != "" || len(e.Rest) != 0 {
		t.Errorf("unexpected lane fields on -f: %+v", e)
	}
}

func TestLegacyForms_CcL_IsLaneWithAutoNumber(t *testing.T) {
	for _, flag := range []string{"-l", "--lane"} {
		e, branch := legacyEntry(t, flag)
		if !e.FactoryEnabled || branch != factoryBranchLane {
			t.Fatalf("%s: enabled=%v branch=%v, want lane", flag, e.FactoryEnabled, branch)
		}
		if !e.FactoryAutoNumber {
			t.Errorf("%s: lane number not marked launcher-chosen", flag)
		}
		if label, ok := parseFactoryLaneLabel(e.Rest); !ok || !strings.HasPrefix(label, "lane-") {
			t.Errorf("%s: Rest %v lacks a canonical lane-<n> --name", flag, e.Rest)
		}
		if e.FactoryLead != "" {
			t.Errorf("%s: FactoryLead = %q, want empty (default leader)", flag, e.FactoryLead)
		}
	}
}

func TestLegacyForms_CcLWithLeader_CarriesRawLeaderName(t *testing.T) {
	for _, args := range [][]string{
		{"-l", "--leader", "leader-2"},
		{"-l", "--leader=leader-2"},
	} {
		e, branch := legacyEntry(t, args...)
		if branch != factoryBranchLane || e.FactoryLead != "leader-2" {
			t.Errorf("%v: branch=%v lead=%q, want lane + leader-2", args, branch, e.FactoryLead)
		}
	}
}

func TestLegacyForms_CcWorktree_IsNotAFactoryEntry(t *testing.T) {
	for _, args := range [][]string{{"-w", "feat-x"}, {"-w"}, {"--worktree", "feat-x"}} {
		e, branch := legacyEntry(t, args...)
		if e.FactoryEnabled || branch != factoryBranchNone {
			t.Errorf("%v: enabled=%v branch=%v, want no factory entry", args, e.FactoryEnabled, branch)
		}
		if !reflect.DeepEqual(e.Rest, args) {
			t.Errorf("%v: Rest = %v, want the -w tokens passed through untouched", args, e.Rest)
		}
	}
}

// `moai glm -l` shares the same entry parse as `moai cc -l`; only the backend
// differs. Pin both the shared parse and the documented usage lines.
func TestLegacyForms_GlmL_SharesTheCcEntryParse(t *testing.T) {
	e, branch := legacyEntry(t, "-l")
	if branch != factoryBranchLane || !e.FactoryEnabled {
		t.Fatalf("-l branch=%v enabled=%v", branch, e.FactoryEnabled)
	}
	for name, use := range map[string]string{"cc": ccCmd.Use, "glm": glmCmd.Use} {
		if !strings.Contains(use, "[-f | -l]") {
			t.Errorf("%s Use = %q lost its [-f | -l] entry", name, use)
		}
	}
}

// `--agent <x>` is not a launcher token today: it is neither a factory entry
// nor consumed, and it reaches the launch argv unchanged, with or without the
// `--` pass-through marker.
func TestLegacyForms_AgentFlag_PassesThroughUnchanged(t *testing.T) {
	for _, args := range [][]string{
		{"--agent", "x"},
		{"--", "--agent", "x"},
	} {
		e, branch := legacyEntry(t, args...)
		if e.FactoryEnabled || branch != factoryBranchNone {
			t.Errorf("%v: treated as a factory entry (enabled=%v branch=%v)", args, e.FactoryEnabled, branch)
		}
		if !reflect.DeepEqual(e.Rest, args) {
			t.Errorf("%v: Rest = %v, want unchanged", args, e.Rest)
		}
		got := resolveLaunchView(e.Rest).LaunchArgs
		if !reflect.DeepEqual(got, args) {
			t.Errorf("%v: launch argv = %v, want unchanged", args, got)
		}
	}
}

// One entry token per launch: -f beside -l is refused.
func TestLegacyForms_FAndLTogether_Refused(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	if _, err := parseLauncherEntry([]string{"-f", "-l"}); err == nil {
		t.Error("-f beside -l was accepted")
	}
}
