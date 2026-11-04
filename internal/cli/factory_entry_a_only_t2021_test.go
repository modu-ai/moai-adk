package cli

import (
	"strings"
	"testing"
)

// Card t2021 (operator decision R385): the launcher keeps two entry flags, -f
// (leader) and -a (agent). -l / --lane are removed and must be refused with a
// message that points at -a, never mapped silently to another role.

func TestT2021_EntryFlags_FAndAAccepted(t *testing.T) {
	p, err := parseFactoryFlag([]string{"-f"})
	if err != nil || !p.Enabled || p.LaneRole {
		t.Fatalf("-f: got %+v, err %v; want leader entry", p, err)
	}
	p, err = parseFactoryFlag([]string{"-a"})
	if err != nil || !p.Enabled || !p.LaneRole {
		t.Fatalf("-a: got %+v, err %v; want agent (lane-role) entry", p, err)
	}
}

func TestT2021_RemovedLaneForms_RefusedPointingAtA(t *testing.T) {
	for _, args := range [][]string{{"-l"}, {"--lane"}, {"-l=x"}, {"--lane=x"}, {"-l", "--leader", "leader-2"}} {
		p, err := parseFactoryFlag(args)
		if err == nil {
			t.Errorf("%v: accepted (%+v), want a refusal", args, p)
			continue
		}
		if !strings.Contains(err.Error(), "-a") {
			t.Errorf("%v: error %q does not point to -a", args, err)
		}
		if p.Enabled || p.LaneRole {
			t.Errorf("%v: refused but still mapped to a role: %+v", args, p)
		}
	}
}

func TestT2021_AWithF_IsConflict(t *testing.T) {
	for _, args := range [][]string{{"-f", "-a"}, {"-a", "-f"}} {
		_, err := parseFactoryFlag(args)
		if err == nil || !strings.Contains(err.Error(), "two entry tokens") {
			t.Errorf("%v: error %v, want the two-entry-tokens conflict", args, err)
		}
	}
}

func TestT2021_AgentLongFlag_StillPassesThrough(t *testing.T) {
	p, err := parseFactoryFlag([]string{"--agent", "reviewer"})
	if err != nil || p.Enabled {
		t.Fatalf("--agent <name>: got %+v, err %v; want untouched pass-through", p, err)
	}
	if strings.Join(p.Rest, " ") != "--agent reviewer" {
		t.Errorf("Rest %v, want [--agent reviewer]", p.Rest)
	}
}
