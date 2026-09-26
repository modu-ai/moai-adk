package cli

import (
	"slices"
	"strings"
	"testing"
)

func TestFactoryAgentVocabulary(t *testing.T) {
	for _, args := range [][]string{{"-f", "agent"}, {"-f=agent"}} {
		p, err := parseFactoryFlag(args)
		if err != nil || !p.Enabled || !p.WorkerRole {
			t.Errorf("parseFactoryFlag(%v) = (%+v, %v)", args, p, err)
		}
	}
	for _, args := range [][]string{{"-f", "agent-3"}, {"--factory=agent-3"}} {
		p, err := parseFactoryFlag(args)
		if err != nil || p.WorkerNumber != 3 || p.WorkerLabel != "agent-3" {
			t.Errorf("parseFactoryFlag(%v) = (%+v, %v)", args, p, err)
		}
	}
	for _, args := range [][]string{{"-f", "worker"}, {"-f", "worker-3"}, {"-f", "lane-3"}} {
		if _, err := parseFactoryFlag(args); err == nil {
			t.Errorf("parseFactoryFlag(%v) accepted removed spelling", args)
		}
		if _, _, _, _, err := stripCodexFactoryFlag(args); err == nil {
			t.Errorf("stripCodexFactoryFlag(%v) accepted removed spelling", args)
		}
	}
}

func TestFactoryAgentLaunchName(t *testing.T) {
	p, err := parseLauncherEntry([]string{"-f", "agent-2", "-b"})
	if err != nil || !slices.Equal(p.Rest, []string{"-b", "--name", "agent-2"}) {
		t.Fatalf("parseLauncherEntry = (%+v, %v)", p, err)
	}
	if label, ok := parseFactoryLaneLabel(p.Rest); !ok || label != "agent-2" {
		t.Fatalf("parseFactoryLaneLabel = (%q, %v)", label, ok)
	}
	if _, err := parseLauncherEntry([]string{"-f", "agent-2", "--name", "agent-3"}); err == nil {
		t.Fatal("expected duplicate-name error")
	}
	for _, args := range [][]string{{"-f", "agent"}, {"-f", "agent-3"}} {
		_, lead, role, lane, err := stripCodexFactoryFlag(args)
		if err != nil || lead || (role == "" && lane == "") {
			t.Errorf("stripCodexFactoryFlag(%v) = (%v, %q, %q, %v)", args, lead, role, lane, err)
		}
	}
}

func TestFactoryHelpAdvertisesAgent(t *testing.T) {
	for name, help := range map[string]string{"cc": ccCmd.Use + "\n" + ccCmd.Long, "glm": glmCmd.Use + "\n" + glmCmd.Long} {
		if !strings.Contains(help, "-f agent") {
			t.Errorf("%s help missing agent form", name)
		}
		if strings.Contains(help, "-f worker") {
			t.Errorf("%s help still advertises worker form", name)
		}
	}
}
