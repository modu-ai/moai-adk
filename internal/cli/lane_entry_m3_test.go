package cli

// lane_entry_m3_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M3 (card t1399): the
// Codex launcher entry. `moai codex -l` / `--lane` is the trigger of the
// supervising relaunch lane that `-f lane` started before; `-f` is no Codex
// entry in any shape; the Codex lane entry takes no argument and no other entry
// token; the leader selector composes with the lane entry only. The cc/glm half
// of AC-003 and AC-005 lives in lane_entry_m2_test.go, whose tests carry the
// codex rows beside theirs.
//
// Every test sets or clears every env axis it reads (netLaneFixture scrubs the
// marker families) and substitutes the binary lookup and the child launch, so
// no Codex session starts and no real home state is written.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// m3CodexRefusal drives one `moai codex` launch that must be refused and
// asserts the shared refusal contract: one diagnostic line on stderr, nothing
// on stdout, exit 1, no child, no lane claim, no new run row, and no transient
// file. It returns the one line.
func m3CodexRefusal(t *testing.T, args []string) string {
	t.Helper()
	root := netLaneFixture(t, kanban.BackendClaude)
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	cap := withCodexLaunchCapture(t)
	runsBefore := runRowCount(t, root)

	stdout, stderr, err := runCodexCmd(t, args...)
	if err == nil {
		t.Fatalf("codex %v was accepted, want a refusal", args)
	}
	if code, ok := ResolveExitCode(err); !ok || code != 1 {
		t.Errorf("codex %v: exit code = (%d, %v), want (1, true); err=%v", args, code, ok, err)
	}
	if stdout != "" {
		t.Errorf("codex %v: stdout = %q, want empty", args, stdout)
	}
	line := strings.TrimRight(stderr, "\n")
	if line == "" || strings.Contains(line, "\n") {
		t.Errorf("codex %v: stderr = %q, want exactly one non-empty line", args, stderr)
	}
	if loc := removedFormPattern.FindString(line); loc != "" {
		t.Errorf("codex %v: refusal %q names the removed form %q", args, line, loc)
	}
	codexWantLaunches(t, cap, 0, 0, 0)
	if reg := loadFactoryRegistry(factoryRegistryPath(root)); len(reg) != 0 {
		t.Errorf("codex %v wrote lane claims: %v", args, reg)
	}
	if got := runRowCount(t, root); got != runsBefore {
		t.Errorf("codex %v changed the run rows: %d -> %d", args, runsBefore, got)
	}
	if leftovers, _ := os.ReadDir(tmp); len(leftovers) != 0 {
		t.Errorf("codex %v wrote %d temp file(s) under TMPDIR", args, len(leftovers))
	}
	return line
}

// TestCodexLaneEntryStartsRelaunchLane — AC-002: `moai codex -l` and
// `moai codex --lane` start the supervising per-card relaunch loop exactly as
// `-f lane` did — the label is claimed automatically through the shared claim
// (the claim row exists afterwards, held by the launching process) and no Codex
// leader starts (the run rows stay the one leader run the fixture recorded, and
// the child carries no leader marker).
func TestCodexLaneEntryStartsRelaunchLane(t *testing.T) {
	for _, spelling := range []string{"-l", "--lane"} {
		lane := netCodexLaneChildFor(t, spelling)
		label := lane.env[config.EnvMoaiFactoryWorker]
		if lane.env[config.EnvFactoryRole] != config.FactoryRoleLane || label == "" {
			t.Errorf("codex %s: child is not a lane: role=%q worker=%q", spelling, lane.env[config.EnvFactoryRole], label)
		}
		if got := lane.env[config.EnvMoaiKanbanBackend]; got != kanban.BackendGPT {
			t.Errorf("codex %s: child %s = %q, want %q", spelling, config.EnvMoaiKanbanBackend, got, kanban.BackendGPT)
		}
		if card := lane.env[config.EnvMoaiKanbanCard]; card == "" {
			t.Errorf("codex %s: the relaunch loop handed the child no card id", spelling)
		}
		// No leader: the leader markers never reach a lane child, and the only
		// run row is the leader run the fixture recorded.
		for _, key := range []string{config.EnvMoaiFactoryWorkers, config.EnvMoaiKanbanLeadAddr} {
			if got := lane.env[key]; got != "" {
				t.Errorf("codex %s: child carries the leader marker %s=%q", spelling, key, got)
			}
		}
		if got := runRowCount(t, lane.root); got != 1 {
			t.Errorf("codex %s: run rows = %d, want the one fixture leader run (a codex leader must not start)", spelling, got)
		}
		// The claim row, not only the child's label.
		reg := loadFactoryRegistry(factoryRegistryPath(lane.root))
		if entry, ok := reg[label]; !ok || entry.PID != os.Getpid() {
			t.Errorf("codex %s: registry[%s] = (%+v, %v), want a claim held by pid %d", spelling, label, entry, ok, os.Getpid())
		}
		if len(reg) != 1 {
			t.Errorf("codex %s: registry = %v, want exactly the one claimed lane", spelling, reg)
		}
	}
}

// TestCodexFactoryFlagRefusals — AC-007: `moai codex -f`, in every shape, is
// refused with one line naming `moai codex -l` for a lane and `moai cc -f` /
// `moai glm -f` for a leader; nothing is leased, claimed, or written.
func TestCodexFactoryFlagRefusals(t *testing.T) {
	for _, args := range [][]string{
		{"-f"},
		{"-f", "lane"},
		{"-f", "lane-2"},
		{"-f", "3"},
		{"--factory", "lane"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			line := m3CodexRefusal(t, args)
			requireTokens(t, args, line, factoryUnsupportedBackendSentinel, "moai codex -l", "moai cc -f", "moai glm -f")
		})
	}
}
