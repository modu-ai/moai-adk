package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// newHeartbeatProject builds a MoAI project root with one registered session
// and returns the root plus that session's id.
func newHeartbeatProject(t *testing.T) (root, sessionID string) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir .moai/state: %v", err)
	}
	sessionID = "sess-heartbeat"
	reg := session.NewRegistry(session.RegistryPathFor(root), nil)
	if err := reg.Register(sessionID, session.SpecIDNone, session.PhaseNone); err != nil {
		t.Fatalf("register: %v", err)
	}
	return root, sessionID
}

func readEntry(t *testing.T, root, sessionID string) session.Entry {
	t.Helper()
	entries, err := session.NewRegistry(session.RegistryPathFor(root), nil).Query("")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	for _, e := range entries {
		if e.SessionID == sessionID {
			return e
		}
	}
	t.Fatalf("session %q not in registry", sessionID)
	return session.Entry{}
}

// TestHeartbeatSeamRefreshesLastHeartbeat is the regression guard for GH #1711
// defect 1: session.Heartbeat had no production caller, so last_heartbeat froze
// at registration for a session's whole life.
//
// The mutant this catches: removing the HeartbeatSeamUserPromptSubmit call from
// the UserPromptSubmit handler, or gutting the seam, leaves last_heartbeat
// equal to started_at.
func TestHeartbeatSeamRefreshesLastHeartbeat(t *testing.T) {
	root, sessionID := newHeartbeatProject(t)
	t.Setenv(config.EnvClaudeProjectDir, root)

	before := readEntry(t, root, sessionID)
	// Precondition, stated rather than assumed: registration leaves the two
	// timestamps equal — this IS the frozen state the seam exists to end.
	if !before.LastHeartbeat.Equal(before.StartedAt) {
		t.Fatalf("precondition: expected last_heartbeat == started_at after Register, got %v vs %v",
			before.LastHeartbeat, before.StartedAt)
	}

	// The registry stamps real UTC time, so advance past its resolution.
	time.Sleep(2 * time.Millisecond)
	HeartbeatSeamUserPromptSubmit(&HookInput{SessionID: sessionID, CWD: root})

	after := readEntry(t, root, sessionID)
	if !after.LastHeartbeat.After(before.LastHeartbeat) {
		t.Errorf("last_heartbeat not advanced: before %v, after %v", before.LastHeartbeat, after.LastHeartbeat)
	}
	if !after.StartedAt.Equal(before.StartedAt) {
		t.Errorf("started_at must not move: before %v, after %v", before.StartedAt, after.StartedAt)
	}
}

// TestHeartbeatSeamIsFailOpen pins the three no-op paths. None of them may
// panic, error, or create a registry file.
func TestHeartbeatSeamIsFailOpen(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		HeartbeatSeamUserPromptSubmit(nil)
	})

	t.Run("empty session id", func(t *testing.T) {
		root, _ := newHeartbeatProject(t)
		t.Setenv(config.EnvClaudeProjectDir, root)
		HeartbeatSeamUserPromptSubmit(&HookInput{SessionID: "", CWD: root})
	})

	t.Run("no registry file leaves none behind", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		t.Setenv(config.EnvClaudeProjectDir, root)

		HeartbeatSeamUserPromptSubmit(&HookInput{SessionID: "sess-unregistered", CWD: root})

		if _, err := os.Stat(session.RegistryPathFor(root)); !os.IsNotExist(err) {
			t.Errorf("seam created a registry file where none existed (stat err = %v)", err)
		}
	})
}
