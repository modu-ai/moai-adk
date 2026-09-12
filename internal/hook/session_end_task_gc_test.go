package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestGarbageCollectOrphanedTasks_KeepsFreshStandaloneTask reproduces the
// hooks-audit H03 finding: a task directory belonging to a different session
// is deleted the instant it has no matching team directory, with no age or
// ownership check. The sibling collector garbageCollectStaleTeams already
// gates on a 24h age threshold; this one does not.
func TestGarbageCollectOrphanedTasks_KeepsFreshStandaloneTask(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	// Deliberately no ~/.claude/teams/<session> counterpart.
	taskDir := filepath.Join(tasksDir, "other-session")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("setup: create task dir: %v", err)
	}
	inProgress := filepath.Join(taskDir, "task-1.json")
	if err := os.WriteFile(inProgress, []byte(`{"status":"in_progress"}`), 0o644); err != nil {
		t.Fatalf("setup: write in_progress task: %v", err)
	}

	garbageCollectOrphanedTasks(homeDir)

	if _, err := os.Stat(inProgress); os.IsNotExist(err) {
		t.Fatal("freshly created standalone task list was deleted; it must be kept until it is provably finished")
	}
}

// TestGarbageCollectOrphanedTasks_CollectsStaleStandaloneTask is the control
// for the test above: the collector must still reclaim a standalone task
// directory once it is old enough to be abandoned.
func TestGarbageCollectOrphanedTasks_CollectsStaleStandaloneTask(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	taskDir := filepath.Join(tasksDir, "abandoned-session")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("setup: create task dir: %v", err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(taskDir, old, old); err != nil {
		t.Fatalf("setup: age the task dir: %v", err)
	}

	garbageCollectOrphanedTasks(homeDir)

	if _, err := os.Stat(taskDir); !os.IsNotExist(err) {
		t.Error("stale standalone task directory should have been collected")
	}
}

// TestGarbageCollectOrphanedTasks_KeepsTaskWhenTeamStatFails reproduces the
// second half of H03: os.Stat failing for a reason other than absence is read
// as absence, so a task directory whose team directory exists but is
// unreadable gets deleted.
func TestGarbageCollectOrphanedTasks_KeepsTaskWhenTeamStatFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: an unreadable parent directory is still traversable")
	}
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	sealed := filepath.Join(teamsDir, "sealed")
	if err := os.MkdirAll(sealed, 0o755); err != nil {
		t.Fatalf("setup: create team dir: %v", err)
	}
	taskDir := filepath.Join(tasksDir, "sealed")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("setup: create task dir: %v", err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(taskDir, old, old); err != nil {
		t.Fatalf("setup: age the task dir: %v", err)
	}
	// Make teamsDir untraversable so Stat on its child fails with EACCES,
	// not ENOENT. The team directory demonstrably still exists.
	if err := os.Chmod(teamsDir, 0o000); err != nil {
		t.Fatalf("setup: seal teams dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(teamsDir, 0o755) })

	garbageCollectOrphanedTasks(homeDir)

	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		t.Error("task directory was deleted on an unreadable team directory; a Stat error is not proof of absence")
	}
}
