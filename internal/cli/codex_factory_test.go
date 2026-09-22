package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

func TestApplyCodexFactoryEntryExportsBackendAndRestores(t *testing.T) {
	t.Setenv(config.EnvMoaiKanbanBackend, "outer")
	restore, err := applyCodexFactoryEntry(&cobra.Command{}, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(config.EnvMoaiKanbanBackend); got != codexFactoryBackend {
		t.Fatalf("factory backend = %q, want %q", got, codexFactoryBackend)
	}
	restore()
	if got := os.Getenv(config.EnvMoaiKanbanBackend); got != "outer" {
		t.Fatalf("restored factory backend = %q, want outer", got)
	}
}

// codex_factory_test.go — `moai codex -f` surface coverage (card t865):
// the factory token is intercepted before the verb lookup, resolved into
// the factory mode env, and never forwarded to the codex child.

func TestStripCodexFactoryFlag(t *testing.T) {
	cases := []struct {
		name     string
		head     []string
		wantRest []string
		lead     bool
		agent    bool
		lane     string
		wantErr  string
	}{
		{name: "bare -f is the lead", head: []string{"-f"}, wantRest: []string{}, lead: true},
		{name: "--factory is the lead", head: []string{"--factory"}, wantRest: []string{}, lead: true},
		{name: "-f agent", head: []string{"-f", "agent"}, wantRest: []string{}, agent: true},
		{name: "--factory=agent", head: []string{"--factory=agent"}, wantRest: []string{}, agent: true},
		{name: "-f=lane-3", head: []string{"-f=lane-3"}, wantRest: []string{}, lane: "lane-3"},
		{name: "-f agent-2 lane compat", head: []string{"-f", "agent-2"}, wantRest: []string{}, lane: "agent-2"},
		{name: "-f 4 retired", head: []string{"-f", "4"}, wantErr: "agent role token"},
		{name: "verb survives", head: []string{"-f", "cli"}, wantRest: []string{"cli"}, lead: true},
		{name: "two tokens with -f", head: []string{"-f", "cli", "--model"}, wantRest: []string{"cli", "--model"}, lead: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rest, lead, agent, lane, err := stripCodexFactoryFlag(c.head)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("stripCodexFactoryFlag(%v) err = %v, want containing %q", c.head, err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("stripCodexFactoryFlag(%v): %v", c.head, err)
			}
			if lead != c.lead || agent != c.agent || lane != c.lane {
				t.Errorf("stripCodexFactoryFlag(%v) = (lead %v, agent %v, lane %q), want (%v, %v, %q)", c.head, lead, agent, lane, c.lead, c.agent, c.lane)
			}
			if len(rest) != len(c.wantRest) {
				t.Errorf("rest = %v, want %v", rest, c.wantRest)
			}
		})
	}
}

func TestNextFactoryAgentNumber(t *testing.T) {
	alive := func(int) bool { return true }
	if n := NextFactoryAgentNumberForTest(map[string]kanban.FactoryWorkerEntry{}, alive); n != 1 {
		t.Errorf("empty registry = %d, want 1", n)
	}
	reg := map[string]kanban.FactoryWorkerEntry{
		"agent-1": {PID: 100},
		"agent-2": {PID: 101},
		"lane-5":  {PID: 102}, // lane slots do not bump the agent sequence
	}
	if n := NextFactoryAgentNumberForTest(reg, alive); n != 3 {
		t.Errorf("registry with agent-2 = %d, want 3", n)
	}
	dead := func(int) bool { return false }
	if n := NextFactoryAgentNumberForTest(reg, dead); n != 1 {
		t.Errorf("dead claims pruned = %d, want 1", n)
	}
}
