package goal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSchemaFields (AC-GLE-005) asserts the schema carries every required
// field incl ceiling default 30.
func TestSchemaFields(t *testing.T) {
	g := NewGoal("sess-1", "all AC rows show PASS", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
		{Type: ConditionModel, Claim: "all AC rows show PASS in the transcript"},
	})

	if g.Goal == "" {
		t.Fatal("goal text missing")
	}
	if len(g.Conditions) != 2 {
		t.Fatalf("conditions: want 2, got %d", len(g.Conditions))
	}
	if g.Conditions[0].Type != ConditionMechanical {
		t.Errorf("cond[0] type: want mechanical, got %s", g.Conditions[0].Type)
	}
	if g.Conditions[1].Type != ConditionModel {
		t.Errorf("cond[1] type: want model, got %s", g.Conditions[1].Type)
	}
	if g.Ceiling.MaxTurns != DefaultMaxTurns {
		t.Errorf("ceiling default: want %d, got %d", DefaultMaxTurns, g.Ceiling.MaxTurns)
	}
	if g.SessionID == "" {
		t.Error("session_id missing")
	}
	if g.Status != StatusArmed {
		t.Errorf("status default: want armed, got %s", g.Status)
	}
	if g.ProgressionMode != DefaultProgressionMode {
		t.Errorf("progression_mode default: want autonomous, got %s", g.ProgressionMode)
	}

	// JSON round-trip must carry the literal field names the schema mandates.
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	js := string(data)
	for _, key := range []string{`"session_id"`, `"goal"`, `"conditions"`, `"ceiling"`, `"max_turns"`, `"turns_used"`, `"progress"`, `"progression_mode"`, `"created_at"`, `"status"`} {
		if !strings.Contains(js, key) {
			t.Errorf("JSON missing field %s", key)
		}
	}
}

// TestStatePathPerSession (AC-GLE-004) asserts the path is per-session and that
// no single shared filename is used.
func TestStatePathPerSession(t *testing.T) {
	a := StatePath("/proj", "sess-A")
	b := StatePath("/proj", "sess-B")
	if a == b {
		t.Fatalf("per-session paths must differ: %s", a)
	}
	if filepath.Base(a) != "sess-A.json" {
		t.Errorf("path A base: want sess-A.json, got %s", filepath.Base(a))
	}
	// StateDir is declared in slash form as a logical location; StatePath joins
	// it with filepath.Join, so the rendered path carries OS-native separators.
	wantDir := filepath.FromSlash(StateDir)
	if !strings.Contains(a, wantDir) {
		t.Errorf("path A must be under %s: %s", wantDir, a)
	}
}

// TestAtomicWrite (AC-GLE-006) asserts the write is temp+rename (no partial
// in-place write observable to a concurrent reader).
func TestAtomicWrite(t *testing.T) {
	root := t.TempDir()
	g := NewGoal("sess-atomic", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	if err := SaveGoal(root, g); err != nil {
		t.Fatal(err)
	}
	// The final file exists and parses.
	loaded, err := LoadGoal(root, "sess-atomic")
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || loaded.Goal != "converge" {
		t.Fatalf("load after atomic save: %+v", loaded)
	}
	// No partial temp files remain in the state dir.
	entries, err := os.ReadDir(filepath.Join(root, StateDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

// TestWriterPidFallback (AC-GLE-008) asserts the writer_pid fallback keys state
// when no session id is available.
func TestWriterPidFallback(t *testing.T) {
	p := StatePath("/proj", "")
	if !strings.HasPrefix(filepath.Base(p), "pid-") {
		t.Errorf("writer_pid fallback path: want pid-* prefix, got %s", filepath.Base(p))
	}
	if WriterPidKey() == "" {
		t.Error("WriterPidKey empty")
	}
}

// TestLoadGoalMissingIsNilNotError asserts a missing goal file returns (nil,nil)
// so the hook can exit 0 with no block.
func TestLoadGoalMissingIsNilNotError(t *testing.T) {
	root := t.TempDir()
	g, err := LoadGoal(root, "no-such-session")
	if err != nil {
		t.Fatalf("missing goal must not error: %v", err)
	}
	if g != nil {
		t.Fatalf("missing goal must return nil, got %+v", g)
	}
}

// TestClearGoalIdempotent asserts clearing an absent goal is not an error.
func TestClearGoalIdempotent(t *testing.T) {
	root := t.TempDir()
	if err := ClearGoal(root, "absent"); err != nil {
		t.Fatalf("clear absent: %v", err)
	}
}
