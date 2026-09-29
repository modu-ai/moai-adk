package hook

// session_start_prune_test.go — card t1312. The SessionStart record writer is
// the prune-on-write trigger: after writing (or finding) its own record it
// sweeps session records older than the configured retention window. The
// config resolution is a targeted state.yaml read — the clean.go precedent —
// so the hook's 5s budget carries one small file read, and every failure in
// the resolution or the sweep fails OPEN: a launch is never gated on cleanup
// (the WriteBestEffort precedent this path already carries).

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedStaleRecord writes an expired session record (entered_at 40 days old)
// into the project's record directory and returns its path.
func seedStaleRecord(t *testing.T, root, sessionID string) string {
	t.Helper()
	rec := &kanban.Record{
		SessionID: sessionID,
		Backend:   kanban.BackendClaude,
		EnteredAt: time.Now().AddDate(0, 0, -40).UTC().Format(time.RFC3339),
	}
	if err := kanban.Write(root, rec); err != nil {
		t.Fatalf("seed stale record: %v", err)
	}
	return kanban.RecordPath(root, sessionID)
}

func writeStateYAML(t *testing.T, root, body string) {
	t.Helper()
	path := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(path, "state.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile state.yaml: %v", err)
	}
}

// The default path — no state.yaml at all — prunes with the 30-day default:
// a stale record goes, the just-written record and its fresh siblings stay.
func TestSessionStartPrunesExpiredRecordsByDefault(t *testing.T) {
	root := newMoaiProjectRoot(t)
	stale := seedStaleRecord(t, root, "stale-sess")

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	writeKanbanSessionRecord(&HookInput{SessionID: "new-sess", ProjectDir: root, CWD: root, Source: "startup"})

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale record survived a SessionStart prune (stat err = %v)", err)
	}
	if _, err := os.Stat(kanban.RecordPath(root, "new-sess")); err != nil {
		t.Fatalf("the session's own record is missing after the prune: %v", err)
	}
}

// The config override is honored: retention 0 disables the sweep and a stale
// record survives a launch.
func TestSessionStartPruneHonorsZeroRetentionOverride(t *testing.T) {
	root := newMoaiProjectRoot(t)
	stale := seedStaleRecord(t, root, "kept-sess")
	writeStateYAML(t, root, "state:\n  session_record_retention_days: 0\n")

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	writeKanbanSessionRecord(&HookInput{SessionID: "new-sess", ProjectDir: root, CWD: root, Source: "startup"})

	if _, err := os.Stat(stale); err != nil {
		t.Fatalf("retention 0 did not disable the prune: %v", err)
	}
}

// A non-default window is read from the same key: 7 days removes a 40-day-old
// record that the default would also remove, so the override is proven by a
// record INSIDE the default window but OUTSIDE the configured one.
func TestSessionStartPruneHonorsConfiguredWindow(t *testing.T) {
	root := newMoaiProjectRoot(t)
	rec := &kanban.Record{
		SessionID: "week-old-sess",
		Backend:   kanban.BackendClaude,
		EnteredAt: time.Now().AddDate(0, 0, -10).UTC().Format(time.RFC3339),
	}
	if err := kanban.Write(root, rec); err != nil {
		t.Fatalf("seed record: %v", err)
	}
	target := kanban.RecordPath(root, "week-old-sess")
	writeStateYAML(t, root, "state:\n  session_record_retention_days: 7\n")

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	writeKanbanSessionRecord(&HookInput{SessionID: "new-sess", ProjectDir: root, CWD: root, Source: "startup"})

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("a 10-day-old record survived a 7-day window (stat err = %v)", err)
	}
}

// A malformed state.yaml fails open: the launch is unaffected, the record is
// still written, and nothing prunes (the default cannot be read, so no sweep
// runs on an unreadable config).
func TestSessionStartPruneFailsOpenOnMalformedConfig(t *testing.T) {
	root := newMoaiProjectRoot(t)
	stale := seedStaleRecord(t, root, "stale-sess")
	writeStateYAML(t, root, "state: [not, a, mapping\n")

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	writeKanbanSessionRecord(&HookInput{SessionID: "new-sess", ProjectDir: root, CWD: root, Source: "startup"})

	if _, err := os.Stat(kanban.RecordPath(root, "new-sess")); err != nil {
		t.Fatalf("a malformed config blocked the record write: %v", err)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Fatalf("a malformed config must not enable a default-window sweep: %v", err)
	}
}

// An unreadable (present but unopenable) state.yaml takes the same
// fail-closed-for-deletion path as a malformed one: no sweep this launch, the
// record still written. A directory where the file should be is the portable
// way to produce a non-NotExist read error.
func TestSessionStartPruneFailsOpenOnUnreadableConfig(t *testing.T) {
	root := newMoaiProjectRoot(t)
	stale := seedStaleRecord(t, root, "stale-sess")
	if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, ".moai", "config", "sections", "state.yaml"), 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	writeKanbanSessionRecord(&HookInput{SessionID: "new-sess", ProjectDir: root, CWD: root, Source: "startup"})

	if _, err := os.Stat(kanban.RecordPath(root, "new-sess")); err != nil {
		t.Fatalf("an unreadable config blocked the record write: %v", err)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Fatalf("an unreadable config must not enable a sweep: %v", err)
	}
}

// The live-session guard: a record the same age as the window's edge belongs
// to a session that may still be running. Recency (not a process probe — the
// record carries no PID) is the cheap liveness signal, so a fresh record is
// never pruned however many launches happen around it.
func TestSessionStartPruneSparesRecentRecords(t *testing.T) {
	root := newMoaiProjectRoot(t)
	recent := &kanban.Record{
		SessionID: "recent-sess",
		Backend:   kanban.BackendClaude,
		EnteredAt: time.Now().AddDate(0, 0, -1).UTC().Format(time.RFC3339),
	}
	if err := kanban.Write(root, recent); err != nil {
		t.Fatalf("seed recent record: %v", err)
	}
	path := kanban.RecordPath(root, "recent-sess")

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	writeKanbanSessionRecord(&HookInput{SessionID: "new-sess", ProjectDir: root, CWD: root, Source: "startup"})

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("a one-day-old record was pruned — the live-session guard failed: %v", err)
	}
}
