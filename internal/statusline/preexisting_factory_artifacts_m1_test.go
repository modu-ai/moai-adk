package statusline

// preexisting_kanban_artifacts_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1
// (card t1399), AC-017 statusline half: the statusline's only reader of the
// state tree the kanban artifacts live in is the queue-count read. A project
// holding session records with the retired chain roles, a kanban-board
// directory with an unreadable role declaration, and a surviving session
// environment carrying the two kanban markers must not make that read fail:
// the unreadable artifacts degrade to absence (Available=false, no panic, no
// authoritative-looking zero).
//
// The marker names are string literals, not config constants (M5b deletes the
// constants; plan-audit finding D-A7).

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestPreexistingKanbanArtifactsTolerated(t *testing.T) {
	root := t.TempDir()
	for _, role := range []string{"plan", "run", "sync"} {
		path := kanban.RecordPath(root, "old-"+role)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		raw := `{"session_id":"old-` + role + `","spec_id":"","role":"` + role + `","backend":"claude","entered_at":"2026-09-01T00:00:00Z","deepscan_dir":"","verify_reentries":0}` + "\n"
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	roles := filepath.Join(root, ".moai", "state", "kanban-board", "roles")
	if err := os.MkdirAll(roles, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(roles, "plan.json"), []byte("\x00\x01 not a role declaration"), 0o600); err != nil {
		t.Fatal(err)
	}
	// The marker names are literals on purpose (see the file comment).
	t.Setenv("MOAI_KANBAN", "1")
	t.Setenv("MOAI_KANBAN_LABEL", "plan")

	// Positive control: the fixture really put files where the state tree is.
	if _, err := os.Stat(kanban.RecordPath(root, "old-plan")); err != nil {
		t.Fatalf("the fixture's session record is absent: %v", err)
	}

	got := resolveBacklogCounts(root)
	if got.Available {
		t.Errorf("counts = %+v, want Available=false: the artifacts hold no queue, and an unreadable artifact must degrade to absence", got)
	}
	if got.Picked != 0 || got.Queued != 0 {
		t.Errorf("counts = %+v, want zeroes alongside Available=false", got)
	}
}
