package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// TestSessionStartPrunesGoalOrphans pins AC-GLE-038: the session-start path
// invokes goal.PruneOrphans, moving an orphan goal state file (a session id
// absent from active-sessions.json + older than OrphanTTL) into consumed/.
func TestSessionStartPrunesGoalOrphans(t *testing.T) {
	root := t.TempDir()

	// Arm an orphan goal for a session that will NOT be active.
	g := goal.NewGoal("dead-session", "converge", []goal.Condition{
		{Type: goal.ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatal(err)
	}
	orphan := filepath.Join(root, goal.StateDir, "dead-session.json")
	// Backdate its mtime beyond OrphanTTL so it is prune-eligible.
	old := time.Now().Add(-goal.OrphanTTL - time.Hour)
	if err := os.Chtimes(orphan, old, old); err != nil {
		t.Fatal(err)
	}

	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:  "live-session",
		ProjectDir: root,
	})
	if err != nil {
		t.Fatalf("Handle must be non-blocking, got error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	// The orphan must be gone from the state dir and present under consumed/.
	if _, statErr := os.Stat(orphan); !os.IsNotExist(statErr) {
		t.Errorf("orphan not pruned from state dir: err=%v", statErr)
	}
	moved := filepath.Join(root, goal.ConsumedDir, "dead-session.json")
	if _, statErr := os.Stat(moved); statErr != nil {
		t.Errorf("orphan not moved to consumed/: %v", statErr)
	}
}

// TestSessionStartGoalPruneFailOpen pins AC-GLE-038b: a PruneOrphans error MUST
// NOT block session start (fail-open). We force ReadDir to error by making the
// goal state directory a regular file.
func TestSessionStartGoalPruneFailOpen(t *testing.T) {
	root := t.TempDir()
	stateParent := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateParent, 0o755); err != nil {
		t.Fatal(err)
	}
	// .moai/state/goal is a FILE, so goal.PruneOrphans' os.ReadDir errors.
	if err := os.WriteFile(filepath.Join(stateParent, "goal"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:  "s",
		ProjectDir: root,
	})
	if err != nil {
		t.Fatalf("prune error must not block session start: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output on prune failure")
	}
}
