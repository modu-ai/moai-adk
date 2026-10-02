package hook

// factory_net_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1 (card t1399): the
// hook half of the factory safety net (AC-015) and the hook half of AC-017.
//
// TestFactoryNetSessionRecord and TestFactoryNetSessionStartNotices guard what
// M5b edits beside the factory code: the session-record role reader sits next
// to the kanban branches M5b deletes, and the factory notice block sits next
// to the kanban notice block. The notice assertions are deliberately limited
// to facts that survive the M4 rewrite of the leader notice (the run id, the
// leader socket, the lane label) — the lane-guidance wording is M4's to
// re-pin and is pinned by AC-010's tests, not here.
//
// The marker names of the surviving-session environment in the AC-017 test are
// string literals, not config constants (M5b deletes the constants; plan-audit
// finding D-A7).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// netScrubFactoryEnv removes every MOAI_FACTORY* and MOAI_KANBAN* variable
// (and the launch-provider fact the record writer reads) from this process for
// the test, by prefix: these tests run inside lane sessions whose own
// variables would otherwise read as the case's input. The t.Setenv call
// registers the restore, the Unsetenv makes the key truly absent.
func netScrubFactoryEnv(t *testing.T) {
	t.Helper()
	keys := map[string]bool{config.EnvMoaiLaunchProvider: true}
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

// TestFactoryNetSessionRecord pins the session-record role reader for a
// factory leader and a factory lane (mutant (e) removes its answer): a lane
// records its role, its number and its backend; a leader signalled by the
// factory fan-out variable alone — no kanban marker anywhere — records the
// leader role with lane 0; an ordinary session records nothing.
func TestFactoryNetSessionRecord(t *testing.T) {
	t.Run("lane", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, kanban.FactoryLaneLabel(3))
		t.Setenv(config.EnvMoaiFactoryWorkers, "0")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGLM)
		t.Setenv(config.EnvMoaiKanbanCard, "t1399")

		writeKanbanSessionRecord(&HookInput{SessionID: "net-lane-sess", ProjectDir: root, CWD: root})

		rec, err := kanban.Read(root, "net-lane-sess")
		if err != nil {
			t.Fatalf("the lane wrote no session record: %v", err)
		}
		if rec.Role != kanban.RoleLane || rec.Lane != 3 {
			t.Errorf("lane record = role %q lane %d, want role %q lane 3", rec.Role, rec.Lane, kanban.RoleLane)
		}
		if rec.Backend != kanban.BackendGLM {
			t.Errorf("lane record backend = %q, want %q", rec.Backend, kanban.BackendGLM)
		}
		if rec.CardID != "t1399" {
			t.Errorf("lane record card = %q, want the explicit card override t1399", rec.CardID)
		}
	})

	t.Run("leader signalled by the factory variable alone", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

		writeKanbanSessionRecord(&HookInput{SessionID: "net-leader-sess", ProjectDir: root, CWD: root})

		rec, err := kanban.Read(root, "net-leader-sess")
		if err != nil {
			t.Fatalf("the factory leader wrote no session record: %v", err)
		}
		if rec.Role != kanban.RoleLeader || rec.Lane != 0 {
			t.Errorf("leader record = role %q lane %d, want role %q lane 0", rec.Role, rec.Lane, kanban.RoleLeader)
		}
		if rec.Backend != kanban.BackendClaude {
			t.Errorf("leader record backend = %q, want %q", rec.Backend, kanban.BackendClaude)
		}
	})

	t.Run("ordinary session writes nothing", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)

		writeKanbanSessionRecord(&HookInput{SessionID: "net-plain-sess", ProjectDir: root, CWD: root})

		if _, err := os.Stat(kanban.RecordPath(root, "net-plain-sess")); err == nil {
			t.Error("a session outside every run wrote a session record")
		}
	})
}

// netSessionStart runs the real SessionStart handler for one startup and
// returns what it injects into the agent context and onto the operator
// channel. MOAI_HOME is sandboxed so the factory peer bind touches nothing
// outside the test.
func netSessionStart(t *testing.T, root, sessionID string) (additionalContext, systemMessage string) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{SessionID: sessionID, CWD: root, ProjectDir: root, Source: "startup"})
	if err != nil {
		t.Fatalf("SessionStart returned an error (it must stay non-blocking): %v", err)
	}
	if out == nil {
		t.Fatal("SessionStart returned no output")
	}
	if out.HookSpecificOutput != nil {
		additionalContext = out.HookSpecificOutput.AdditionalContext
	}
	return additionalContext, out.SystemMessage
}

// TestFactoryNetSessionStartNotices pins the factory SessionStart notices
// through the real handler (mutant (f) removes the notice block): a leader
// receives its run announcement with the run id and the leader socket on BOTH
// channels (the agent context and the operator system message); a lane
// receives the join line naming its label; a session outside every run
// receives no factory notice.
func TestFactoryNetSessionStartNotices(t *testing.T) {
	t.Run("leader", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		t.Setenv(config.EnvMoaiKanbanID, "netrun01")
		t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-socket-factory/netrun01")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

		ctx, sys := netSessionStart(t, root, "net-notice-leader")
		for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
			for _, want := range []string{"netrun01", "/tmp/moai-socket-factory/netrun01"} {
				if !strings.Contains(text, want) {
					t.Errorf("leader notice on %s lacks %q:\n%s", channel, want, text)
				}
			}
			if !strings.Contains(strings.ToLower(text), "leader") {
				t.Errorf("leader notice on %s never names the leader role:\n%s", channel, text)
			}
		}
	})

	t.Run("lane", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, kanban.FactoryLaneLabel(2))
		t.Setenv(config.EnvMoaiFactoryWorkers, "0")
		t.Setenv(config.EnvMoaiKanbanID, "netrun01")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

		ctx, sys := netSessionStart(t, root, "net-notice-lane")
		for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
			if !strings.Contains(text, kanban.FactoryLaneLabel(2)) {
				t.Errorf("lane notice on %s does not name the lane label %q:\n%s", channel, kanban.FactoryLaneLabel(2), text)
			}
		}
	})

	t.Run("ordinary session gets no factory notice", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)

		ctx, sys := netSessionStart(t, root, "net-notice-plain")
		for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
			if strings.Contains(text, "Factory Mode") || strings.Contains(text, "moai-socket-factory") {
				t.Errorf("an ordinary session received a factory notice on %s:\n%s", channel, text)
			}
		}
	})
}

// TestPreexistingKanbanArtifactsTolerated is the hook half of AC-017: a
// project holding a session record with the retired chain role `plan`, a
// kanban-board directory with an unreadable role declaration, and a surviving
// session environment carrying the two kanban markers does not make the
// SessionStart hook fail, and the pre-existing record is left as it was.
//
// Measured at M1: the hook still emitted its kanban bootstrap notice for a
// surviving MOAI_KANBAN session, so the test then asserted tolerance only.
// M5b deleted the notice and the record reader's companion branch, and the
// test now also pins both: no notice on either channel, and no record for a
// session that carries only the retired markers (AC-014 pins the notice
// absence on its own in TestSessionStartEmitsNoKanbanNotice).
func TestPreexistingKanbanArtifactsTolerated(t *testing.T) {
	root := newMoaiProjectRoot(t)
	netScrubFactoryEnv(t)

	recPath := kanban.RecordPath(root, "old-plan")
	if err := os.MkdirAll(filepath.Dir(recPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// entered_at is now: a record older than the retention window is swept by the
	// hook's own prune, which is a different behavior from the one under test.
	original := `{"session_id":"old-plan","spec_id":"","role":"plan","backend":"claude","entered_at":"` + time.Now().UTC().Format(time.RFC3339) + `","deepscan_dir":"","verify_reentries":0}` + "\n"
	if err := os.WriteFile(recPath, []byte(original), 0o600); err != nil {
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

	// The same session id as the pre-existing record, and a different one.
	for _, sessionID := range []string{"old-plan", "new-session"} {
		ctx, sys := netSessionStart(t, root, sessionID)
		// M5b: the surviving markers announce nothing on either channel.
		for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
			if strings.Contains(strings.ToLower(text), "kanban") {
				t.Errorf("session %s received a kanban notice on %s:\n%s", sessionID, channel, text)
			}
		}
	}
	// M5b: the record role reader has no companion branch, so the surviving
	// markers give a session with a fresh id no record of its own.
	if _, err := os.Stat(kanban.RecordPath(root, "new-session")); err == nil {
		t.Error("a session carrying only the retired markers wrote a session record")
	}

	got, err := os.ReadFile(recPath)
	if err != nil {
		t.Fatalf("the pre-existing record vanished: %v", err)
	}
	if string(got) != original {
		t.Errorf("the pre-existing record was rewritten:\n got: %s\nwant: %s", got, original)
	}
	if _, err := kanban.Read(root, "old-plan"); err != nil {
		t.Errorf("the pre-existing role-plan record is unreadable through the reader: %v", err)
	}
}
