package cli

// lane_entry_golden_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1 (card
// t1399), AC-004 golden.
//
// The golden pins the lane environment markers (the MOAI_FACTORY* and
// MOAI_KANBAN* families) that today's `-f lane` launch publishes for the three
// backends: the cc and glm rows are the environment live at the engine launch
// seam, the codex row is the per-card child environment of the supervising
// loop. It is captured from today's code BEFORE the `-l` entry exists
// (testdata/lane_entry_env_golden.json, committed alone) and compared here
// against the same launches, so a later milestone that adds `-l` can prove the
// new entry publishes the identical set (label aside), and so a milestone that
// removes the kanban surface can prove it dropped no lane marker.
//
// The codex row carries MOAI_KANBAN_LABEL because today's lane child does; the
// stamp is removed at M5a and the parity comparison excludes exactly that key
// from M2 on (REQ-003's one exception), so the golden needs no re-pin.
//
// Capture: MOAI_LANE_GOLDEN_WRITE=<path> go test ./internal/cli -run
// '^TestLaneMarkerGoldenMatchesFLane$' -count=1 rewrites the golden at <path>
// instead of comparing. The marker names below are literals on purpose — the
// golden is a statement about the wire, not about the Go constants.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

const (
	laneGoldenPath     = "testdata/lane_entry_env_golden.json"
	laneGoldenWriteEnv = "MOAI_LANE_GOLDEN_WRITE"
)

// laneMarkerEnv keeps only the lane marker families of an environment. The
// match is by prefix with no trailing underscore so the bare MOAI_KANBAN
// variable would be captured too (it is absent on a lane today, and the golden
// records that absence).
func laneMarkerEnv(env map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range env {
		if strings.HasPrefix(k, "MOAI_FACTORY") || strings.HasPrefix(k, "MOAI_KANBAN") {
			out[k] = v
		}
	}
	return out
}

// netScrubLaneEnv removes every variable of the two marker families from this
// process for the duration of the test, by prefix rather than by a fixed list,
// so a lane session's ambient factory environment (this suite runs inside one)
// can neither leak into a capture nor be mistaken for a published marker. The
// t.Setenv call registers the restore; the Unsetenv makes the key truly absent
// (a key set to "" would read as a published empty marker).
func netScrubLaneEnv(t *testing.T) {
	t.Helper()
	keys := map[string]bool{
		config.EnvAutonomyTier:                     true,
		config.EnvClaudeCodeMaxConcurrentSubagents: true,
		config.EnvClaudeCodeStopHookBlockCap:       true,
	}
	for _, kv := range os.Environ() {
		name, _, found := strings.Cut(kv, "=")
		if found && (strings.HasPrefix(name, "MOAI_FACTORY") || strings.HasPrefix(name, "MOAI_KANBAN")) {
			keys[name] = true
		}
	}
	for key := range keys {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// netEnvMap flattens the live process environment into a map.
func netEnvMap() map[string]string {
	m := map[string]string{}
	for _, kv := range os.Environ() {
		if name, value, found := strings.Cut(kv, "="); found {
			m[name] = value
		}
	}
	return m
}

// captureLaneEnvToday launches a lane on each backend through `-l` and
// returns the marker environment each launch published, keyed by row name.
func captureLaneEnvToday(t *testing.T) map[string]map[string]string {
	t.Helper()
	rows := map[string]map[string]string{}

	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
	}{
		{"cc", kanban.BackendClaude, netCC},
		{"glm", kanban.BackendGLM, netGLM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := netLaneFixture(t, tc.backend)
			launch := netDriveLaunch(t, root, tc.entry, []string{"-l"})
			if !launch.launched || launch.err != nil {
				t.Fatalf("%s lane launch: launched=%v err=%v", tc.name, launch.launched, launch.err)
			}
			rows[tc.name] = laneMarkerEnv(launch.env)
		})
	}

	t.Run("codex", func(t *testing.T) {
		rows["codex"] = laneMarkerEnv(netCodexLaneChild(t).env)
	})
	return rows
}

// TestLaneMarkerGoldenMatchesFLane compares the lane launch marker
// environment, per backend, with the committed golden (AC-004 at M1). A key
// that appears, disappears, or changes value fails here with the full diff of
// the row.
func TestLaneMarkerGoldenMatchesFLane(t *testing.T) {
	got := captureLaneEnvToday(t)
	if t.Failed() {
		return
	}
	for _, name := range []string{"cc", "glm", "codex"} {
		if len(got[name]) == 0 {
			t.Fatalf("row %q captured no marker (positive control): got %v", name, got)
		}
	}

	if dest := os.Getenv(laneGoldenWriteEnv); dest != "" {
		raw, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, append(raw, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Logf("golden written to %s", dest)
		return
	}

	raw, err := os.ReadFile(laneGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var want map[string]map[string]string
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	for _, name := range []string{"cc", "glm", "codex"} {
		if reflect.DeepEqual(got[name], want[name]) {
			continue
		}
		t.Errorf("row %q differs from the golden\n got: %s\nwant: %s", name, formatMarkerRow(got[name]), formatMarkerRow(want[name]))
	}
	if len(want) != 3 {
		t.Errorf("golden holds %d rows, want exactly cc, glm, codex", len(want))
	}
}

func formatMarkerRow(row map[string]string) string {
	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+row[k])
	}
	return strings.Join(parts, " ")
}
