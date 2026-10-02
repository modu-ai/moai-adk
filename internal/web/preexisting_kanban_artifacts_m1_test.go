package web

// preexisting_kanban_artifacts_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1
// (card t1399), AC-017 web half: a project holding session records with the
// retired chain roles, a kanban-board directory with an unreadable role
// declaration, and a surviving session environment carrying the two kanban
// markers does not make a console builder fail; an unreadable artifact
// degrades to absence.
//
// The marker names are string literals, not config constants (M5b deletes the
// constants; plan-audit finding D-A7).

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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
	// A record the loader cannot parse degrades to absence beside the good ones.
	if err := os.WriteFile(kanban.RecordPath(root, "old-garbled"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
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

	if got := len(loadKanbanRecords(root)); got < 3 {
		t.Fatalf("loadKanbanRecords returned %d record(s), want the three role records readable (positive control)", got)
	}

	a := newTestApp(t)
	a.cfg.ProjectRoot = root
	if _, err := a.buildOverview(time.Now()); err != nil {
		t.Errorf("buildOverview failed on pre-existing kanban artifacts: %v", err)
	}
	if _, err := a.buildKanban(time.Now()); err != nil {
		t.Errorf("buildKanban failed on pre-existing kanban artifacts: %v", err)
	}
}
