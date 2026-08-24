package web

// factory_lanes_test.go — SPEC-WEB-CONSOLE-015 M1: the per-lane factory
// progress view model and the three-file join behind it.
//
// The join is non-unique on BOTH sides (spec.md §A.4, REQ-WC15-047): the
// factory registry can carry one pid on two lanes, and the session registry
// deduplicates by session id alone, so two of its entries may carry one pid.
// A completed-but-wrong lookup is the hazard these tests exist to forbid — an
// empty lookup renders as "unresolved", a wrong one renders a confident wrong
// row.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

// writeFactoryRegistry writes root's lane registry from a label→pid map.
func writeFactoryRegistry(t *testing.T, root string, lanes map[string]int) {
	t.Helper()
	reg := make(map[string]kanban.FactoryWorkerEntry, len(lanes))
	for label, pid := range lanes {
		reg[label] = kanban.FactoryWorkerEntry{PID: pid, RegisteredAt: time.Now().UTC().Format(time.RFC3339)}
	}
	path := kanban.FactoryRegistryPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir factory dir: %v", err)
	}
	body, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal registry: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

// writeActiveSessions writes root's session registry.
func writeActiveSessions(t *testing.T, root string, entries []session.Entry) {
	t.Helper()
	path := filepath.Join(root, ".moai", "state", "active-sessions.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	body, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write active-sessions: %v", err)
	}
}

// writeKanbanRecord writes one kanban record keyed by its session id.
func writeKanbanRecord(t *testing.T, root string, rec kanban.Record) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state", "kanban")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir kanban dir: %v", err)
	}
	body, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, rec.SessionID+".json"), body, 0o600); err != nil {
		t.Fatalf("write record: %v", err)
	}
}

// liveEntry is a registry entry whose heartbeat is fresh.
func liveEntry(sessionID string, pid int) session.Entry {
	return session.Entry{
		SessionID:     sessionID,
		SpecID:        "SPEC-EXAMPLE-001",
		PID:           pid,
		LastHeartbeat: time.Now().UTC(),
	}
}

// laneByNumber finds the row for lane n, failing when it is absent.
func laneByNumber(t *testing.T, lanes []LaneVM, n int) LaneVM {
	t.Helper()
	for _, l := range lanes {
		if l.Lane == n {
			return l
		}
	}
	t.Fatalf("lane %d absent from %d rows: %+v", n, len(lanes), lanes)
	return LaneVM{}
}

// TestFactoryLanesResolveCompleteJoin — AC-WC15-043a / AC-WC15-044: a lane
// whose pid resolves to exactly one session carrying a record shows that
// record's card id, spec id, state and stage.
func TestFactoryLanesResolveCompleteJoin(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-2": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-lane-2", pid)})
	writeKanbanRecord(t, root, kanban.Record{
		SessionID: "sess-lane-2",
		SpecID:    "SPEC-EXAMPLE-001",
		Role:      "lane",
		Backend:   kanban.BackendGLM,
		Lane:      2,
		CardID:    "t207",
	})

	_, byID := loadSessions(root, time.Now())
	lanes := loadFactoryLanes(root, byID, loadKanbanRecords(root))

	row := laneByNumber(t, lanes, 2)
	if row.Unresolved {
		t.Fatalf("complete join marked unresolved: %+v", row)
	}
	if row.CardID != "t207" {
		t.Errorf("CardID = %q, want t207", row.CardID)
	}
	if row.SpecID != "SPEC-EXAMPLE-001" {
		t.Errorf("SpecID = %q, want SPEC-EXAMPLE-001", row.SpecID)
	}
	if row.State != StateLive {
		t.Errorf("State = %q, want %q", row.State, StateLive)
	}
	if row.Stage != StageActive {
		t.Errorf("Stage = %q, want %q", row.Stage, StageActive)
	}
	if !row.StageEstimated {
		t.Error("StageEstimated = false, want true — heartbeat estimation is not a recorded transition")
	}
	if row.Backend != kanban.BackendGLM {
		t.Errorf("Backend = %q, want %q", row.Backend, kanban.BackendGLM)
	}
}

// TestFactoryLanesPresentsUnresolvedLanes — AC-WC15-043b: a lane whose pid is
// in no session entry, and a lane whose session has no record, are both
// PRESENT, carry their lane number, and carry the unresolved marker.
func TestFactoryLanesPresentsUnresolvedLanes(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-4": 999001, "lane-6": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-lane-6", pid)})
	// No record for sess-lane-6, and no session at all for lane-4's pid.

	_, byID := loadSessions(root, time.Now())
	lanes := loadFactoryLanes(root, byID, loadKanbanRecords(root))

	for _, n := range []int{4, 6} {
		row := laneByNumber(t, lanes, n)
		if !row.Unresolved {
			t.Errorf("lane %d Unresolved = false, want true: %+v", n, row)
		}
		if row.CardID != "" || row.SpecID != "" {
			t.Errorf("lane %d carries join values while unresolved: %+v", n, row)
		}
	}
}

// TestFactoryLanesDuplicatePIDFactorySide — AC-WC15-047 first half: two lanes
// carrying one pid attribute the record to NEITHER, and both are unresolved.
func TestFactoryLanesDuplicatePIDFactorySide(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-1": pid, "lane-5": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-dup", pid)})
	writeKanbanRecord(t, root, kanban.Record{
		SessionID: "sess-dup", SpecID: "SPEC-EXAMPLE-001", Role: "lane",
		Backend: kanban.BackendClaude, Lane: 1, CardID: "t999",
	})

	_, byID := loadSessions(root, time.Now())
	lanes := loadFactoryLanes(root, byID, loadKanbanRecords(root))

	for _, n := range []int{1, 5} {
		row := laneByNumber(t, lanes, n)
		if !row.Unresolved {
			t.Errorf("lane %d Unresolved = false, want true: %+v", n, row)
		}
		if row.CardID == "t999" || row.SpecID == "SPEC-EXAMPLE-001" {
			t.Errorf("lane %d attributed an ambiguous record: %+v", n, row)
		}
	}
}

// TestFactoryLanesDuplicatePIDSessionSide — AC-WC15-047 second half: one lane,
// one pid, but TWO session entries bearing it. The registry deduplicates by
// session id alone, so this is reachable with no factory-side duplication.
func TestFactoryLanesDuplicatePIDSessionSide(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-1": pid})
	writeActiveSessions(t, root, []session.Entry{
		liveEntry("sess-stale", pid),
		liveEntry("sess-live", pid),
	})
	writeKanbanRecord(t, root, kanban.Record{
		SessionID: "sess-live", SpecID: "SPEC-EXAMPLE-001", Role: "lane",
		Backend: kanban.BackendClaude, Lane: 1, CardID: "t888",
	})

	_, byID := loadSessions(root, time.Now())
	lanes := loadFactoryLanes(root, byID, loadKanbanRecords(root))

	row := laneByNumber(t, lanes, 1)
	if !row.Unresolved {
		t.Fatalf("lane 1 Unresolved = false, want true — two session entries carry one pid: %+v", row)
	}
	if row.CardID != "" || row.SpecID != "" {
		t.Errorf("lane 1 attributed a record despite an ambiguous pid: %+v", row)
	}
}

// TestFactoryLanesMissingAndMalformedRegistry — AC-WC15-046: an absent or
// malformed registry yields zero lanes and no error.
func TestFactoryLanesMissingAndMalformedRegistry(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		_, byID := loadSessions(root, time.Now())
		if lanes := loadFactoryLanes(root, byID, nil); len(lanes) != 0 {
			t.Errorf("lanes = %d, want 0", len(lanes))
		}
	})
	t.Run("malformed", func(t *testing.T) {
		root := t.TempDir()
		path := kanban.FactoryRegistryPath(root)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		_, byID := loadSessions(root, time.Now())
		if lanes := loadFactoryLanes(root, byID, nil); len(lanes) != 0 {
			t.Errorf("lanes = %d, want 0", len(lanes))
		}
	})
}

// TestFactoryLanesJoinWritesNothing — AC-WC15-043a second half: the state
// directory listing is identical immediately before and after the join.
func TestFactoryLanesJoinWritesNothing(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-3": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-3", pid)})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-3", Role: "lane", Lane: 3, CardID: "t3"})

	before := listStateTree(t, root)
	_, byID := loadSessions(root, time.Now())
	_ = loadFactoryLanes(root, byID, loadKanbanRecords(root))
	after := listStateTree(t, root)

	if len(before) != len(after) {
		t.Fatalf("state tree changed: %d entries before, %d after", len(before), len(after))
	}
	for i := range before {
		if before[i] != after[i] {
			t.Errorf("state tree entry %d changed:\n before %q\n after  %q", i, before[i], after[i])
		}
	}
}

// listStateTree returns a path+size+mtime listing of root's state tree.
func listStateTree(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(filepath.Join(root, ".moai", "state"), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		out = append(out, p+"|"+itoa(int(info.Size()))+"|"+info.ModTime().Format(time.RFC3339Nano))
		return nil
	})
	if err != nil {
		t.Fatalf("walk state tree: %v", err)
	}
	return out
}

// TestFactoryLanesOrderedByNumber — lanes render in numeric order, not in the
// registry map's iteration order (which Go randomises).
func TestFactoryLanesOrderedByNumber(t *testing.T) {
	root := t.TempDir()
	writeFactoryRegistry(t, root, map[string]int{"lane-10": 999002, "lane-2": 999003, "lane-7": 999004})
	_, byID := loadSessions(root, time.Now())
	lanes := loadFactoryLanes(root, byID, nil)

	want := []int{2, 7, 10}
	if len(lanes) != len(want) {
		t.Fatalf("lanes = %d, want %d", len(lanes), len(want))
	}
	for i, n := range want {
		if lanes[i].Lane != n {
			t.Errorf("lanes[%d].Lane = %d, want %d", i, lanes[i].Lane, n)
		}
	}
}

// TestFactoryLanesIgnoresNonLaneLabels — a registry key that is not lane-shaped
// is not a lane and produces no row.
func TestFactoryLanesIgnoresNonLaneLabels(t *testing.T) {
	root := t.TempDir()
	writeFactoryRegistry(t, root, map[string]int{"lane-1": 999005, "lane-a": 999006, "lead": 999007})
	_, byID := loadSessions(root, time.Now())
	lanes := loadFactoryLanes(root, byID, nil)
	if len(lanes) != 1 || lanes[0].Lane != 1 {
		t.Fatalf("lanes = %+v, want exactly lane 1", lanes)
	}
}
