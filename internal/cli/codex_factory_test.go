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
	restore, err := applyCodexFactoryEntry(&cobra.Command{}, "", "")
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
		role     string
		lane     string
		wantErr  string
	}{
		{name: "bare -f is the lead", head: []string{"-f"}, wantRest: []string{}, lead: true},
		{name: "--factory is the lead", head: []string{"--factory"}, wantRest: []string{}, lead: true},
		{name: "-f worker", head: []string{"-f", "worker"}, wantRest: []string{}, role: "worker"},
		{name: "--factory=worker", head: []string{"--factory=worker"}, wantRest: []string{}, role: "worker"},
		{name: "-f=worker-3", head: []string{"-f=worker-3"}, wantRest: []string{}, lane: "worker-3"},
		// Legacy spellings (keep-alias coverage): still parsed as typed; the
		// claim canonicalizes them and prints the deprecation hint.
		{name: "-f agent legacy alias", head: []string{"-f", "agent"}, wantRest: []string{}, role: "agent"},
		{name: "--factory=agent legacy alias", head: []string{"--factory=agent"}, wantRest: []string{}, role: "agent"},
		{name: "-f=lane-3 legacy alias", head: []string{"-f=lane-3"}, wantRest: []string{}, lane: "lane-3"},
		{name: "-f agent-2 legacy alias", head: []string{"-f", "agent-2"}, wantRest: []string{}, lane: "agent-2"},
		{name: "-f 4 retired", head: []string{"-f", "4"}, wantErr: "worker role token"},
		{name: "verb survives", head: []string{"-f", "cli"}, wantRest: []string{"cli"}, lead: true},
		{name: "two tokens with -f", head: []string{"-f", "cli", "--model"}, wantRest: []string{"cli", "--model"}, lead: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rest, lead, role, lane, err := stripCodexFactoryFlag(c.head)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("stripCodexFactoryFlag(%v) err = %v, want containing %q", c.head, err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("stripCodexFactoryFlag(%v): %v", c.head, err)
			}
			if lead != c.lead || role != c.role || lane != c.lane {
				t.Errorf("stripCodexFactoryFlag(%v) = (lead %v, role %q, lane %q), want (%v, %q, %q)", c.head, lead, role, lane, c.lead, c.role, c.lane)
			}
			if len(rest) != len(c.wantRest) {
				t.Errorf("rest = %v, want %v", rest, c.wantRest)
			}
		})
	}
}

// TestNextFactoryWorkerNumber: the worker join takes one past the highest
// live claim across the canonical and legacy shapes — the former separate
// agent-<n> and lane-<n> sequences are one worker numbering now.
func TestNextFactoryWorkerNumber(t *testing.T) {
	alive := func(int) bool { return true }
	if n := NextFactoryWorkerNumberForTest(map[string]kanban.FactoryWorkerEntry{}, alive); n != 1 {
		t.Errorf("empty registry = %d, want 1", n)
	}
	reg := map[string]kanban.FactoryWorkerEntry{
		"worker-1": {PID: 100},
		"agent-2":  {PID: 101}, // legacy row
		"lane-5":   {PID: 102}, // legacy row — shares the one numbering
	}
	if n := NextFactoryWorkerNumberForTest(reg, alive); n != 6 {
		t.Errorf("registry up to lane-5 = %d, want 6", n)
	}
	dead := func(int) bool { return false }
	if n := NextFactoryWorkerNumberForTest(reg, dead); n != 1 {
		t.Errorf("dead claims pruned = %d, want 1", n)
	}
}
