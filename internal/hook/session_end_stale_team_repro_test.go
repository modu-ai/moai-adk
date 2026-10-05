package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestGarbageCollectStaleTeams_KeepsLiveTeamWithOldDirMtime reproduces H02 (card t597).
//
// A directory's mtime advances when an entry is created, removed, or renamed
// inside it — NOT when an existing file is rewritten in place. A long-lived team
// whose config.json is updated repeatedly therefore keeps the mtime it was born
// with, and after 24 hours the collector reads that stale clock as "abandoned"
// and deletes the team plus its task list, while the owning session is still
// running and its tasks are still in_progress.
func TestGarbageCollectStaleTeams_KeepsLiveTeamWithOldDirMtime(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	for _, d := range []string{teamsDir, tasksDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("setup %s: %v", d, err)
		}
	}

	liveTeamDir := filepath.Join(teamsDir, "live-team")
	if err := os.MkdirAll(liveTeamDir, 0o755); err != nil {
		t.Fatalf("create live team dir: %v", err)
	}
	// Freshly written config owned by a DIFFERENT, still-running session.
	configPath := filepath.Join(liveTeamDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"leadSessionId":"other-live-session"}`), 0o644); err != nil {
		t.Fatalf("write live config: %v", err)
	}

	liveTaskDir := filepath.Join(tasksDir, "live-team")
	if err := os.MkdirAll(liveTaskDir, 0o755); err != nil {
		t.Fatalf("create live task dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(liveTaskDir, "task-1.json"), []byte(`{"status":"in_progress"}`), 0o644); err != nil {
		t.Fatalf("write live task: %v", err)
	}

	// The directory clock is old; everything inside it is minutes old.
	old := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(liveTeamDir, old, old); err != nil {
		t.Fatalf("age live team dir: %v", err)
	}

	garbageCollectStaleTeams(homeDir)

	if _, err := os.Stat(liveTeamDir); os.IsNotExist(err) {
		t.Error("live team directory was deleted: a stale directory mtime is not evidence the owning session ended")
	}
	if _, err := os.Stat(liveTaskDir); os.IsNotExist(err) {
		t.Error("live task directory was deleted alongside the live team directory")
	}
}

// TestGarbageCollectStaleTeams_StillCollectsFullyQuietTeam is the control: a team
// whose directory AND contents have all gone quiet past the window is still
// collected, so the repro above cannot be satisfied by disabling collection.
func TestGarbageCollectStaleTeams_StillCollectsFullyQuietTeam(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	for _, d := range []string{teamsDir, tasksDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("setup %s: %v", d, err)
		}
	}

	deadTeamDir := filepath.Join(teamsDir, "dead-team")
	if err := os.MkdirAll(deadTeamDir, 0o755); err != nil {
		t.Fatalf("create dead team dir: %v", err)
	}
	configPath := filepath.Join(deadTeamDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"leadSessionId":"long-gone-session"}`), 0o644); err != nil {
		t.Fatalf("write dead config: %v", err)
	}
	deadTaskDir := filepath.Join(tasksDir, "dead-team")
	if err := os.MkdirAll(deadTaskDir, 0o755); err != nil {
		t.Fatalf("create dead task dir: %v", err)
	}

	old := time.Now().Add(-25 * time.Hour)
	for _, path := range []string{configPath, deadTeamDir, deadTaskDir} {
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatalf("age %s: %v", path, err)
		}
	}

	garbageCollectStaleTeams(homeDir)

	if _, err := os.Stat(deadTeamDir); !os.IsNotExist(err) {
		t.Error("fully quiet team directory should still be collected")
	}
}

// TestGarbageCollectStaleTeams_KeepsTeamWhenTaskActivityUnmeasurable covers the
// preserve-on-doubt half of the rule: when the matching task directory exists
// but cannot be read, the team's real quiet time is unknown, and an unknown is
// never grounds for deleting data.
func TestGarbageCollectStaleTeams_KeepsTeamWhenTaskActivityUnmeasurable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: an unreadable directory is still traversable")
	}
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	for _, d := range []string{teamsDir, tasksDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("setup %s: %v", d, err)
		}
	}

	teamDir := filepath.Join(teamsDir, "sealed-tasks")
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("create team dir: %v", err)
	}
	taskDir := filepath.Join(tasksDir, "sealed-tasks")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("create task dir: %v", err)
	}

	old := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(teamDir, old, old); err != nil {
		t.Fatalf("age team dir: %v", err)
	}
	// Sealed, so walking it fails with EACCES rather than reporting a time.
	if err := os.Chmod(taskDir, 0o000); err != nil {
		t.Fatalf("seal task dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(taskDir, 0o755) })

	garbageCollectStaleTeams(homeDir)

	if _, err := os.Stat(teamDir); os.IsNotExist(err) {
		t.Error("team directory was deleted while its task activity could not be measured")
	}
}
