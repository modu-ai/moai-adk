package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// writeAnchorRegistry writes entries into the tree-local registry file that
// LiveAnchoredSessions reads. Direct JSON authorship keeps the frozen
// schema (REQ-COORD-024) as the only contract under test; Register() would
// pin CWD/PID to the test process.
func writeAnchorRegistry(t *testing.T, tree string, entries []Entry) {
	t.Helper()
	dir := filepath.Join(tree, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active-sessions.json"), data, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

func anchorTestEntry(host, cwd string, pid int, heartbeatAge time.Duration) Entry {
	ts := time.Now().UTC().Add(-heartbeatAge)
	return Entry{
		SessionID:     "aaaaaaaa-0000-0000-0000-000000000000",
		SpecID:        "SPEC-ANCHOR-001",
		Phase:         "run",
		StartedAt:     ts,
		LastHeartbeat: ts,
		PID:           pid,
		Host:          host,
		CWD:           cwd,
	}
}

// TestLiveAnchoredSessions_KeepsLivePIDWithStaleHeartbeat pins the real
// lane shape: alive process, heartbeat hours old (no per-turn heartbeat
// driver exists). The PID probe — not the timestamp — carries the verdict.
func TestLiveAnchoredSessions_KeepsLivePIDWithStaleHeartbeat(t *testing.T) {
	tree := t.TempDir()
	host, _ := os.Hostname()
	writeAnchorRegistry(t, tree, []Entry{
		anchorTestEntry(host, tree, os.Getpid(), 2*time.Hour),
	})

	got := LiveAnchoredSessions(tree, time.Now().UTC())
	if len(got) != 1 {
		t.Fatalf("live lane with stale heartbeat must be kept, got %d entries", len(got))
	}
	if got[0].PID != os.Getpid() {
		t.Errorf("returned entry pid = %d, want %d", got[0].PID, os.Getpid())
	}
}

// TestLiveAnchoredSessions_DropsDeadPIDStaleHeartbeat: a zombie whose
// heartbeat is also stale must not block disposal. Unix-only: the Windows
// probe is conservative-alive by design (see anchor_pid_windows.go).
func TestLiveAnchoredSessions_DropsDeadPIDStaleHeartbeat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows probe is conservative-alive; zombies age out via --force instead")
	}
	tree := t.TempDir()
	host, _ := os.Hostname()
	// PID 999999999 exceeds pid_max on linux (2^22) and macOS (99999), so
	// it can never be a live process.
	writeAnchorRegistry(t, tree, []Entry{
		anchorTestEntry(host, tree, 999999999, 2*time.Hour),
	})

	if got := LiveAnchoredSessions(tree, time.Now().UTC()); len(got) != 0 {
		t.Fatalf("dead pid with stale heartbeat must be dropped, got %d entries", len(got))
	}
}

// TestLiveAnchoredSessions_KeepsFreshHeartbeatDespiteDeadPID covers the
// conservative floor: a recently started session stays protected even when
// the probe cannot confirm it (same expectation on every platform — on
// Windows the probe claims alive, on Unix the freshness floor holds).
func TestLiveAnchoredSessions_KeepsFreshHeartbeatDespiteDeadPID(t *testing.T) {
	tree := t.TempDir()
	host, _ := os.Hostname()
	writeAnchorRegistry(t, tree, []Entry{
		anchorTestEntry(host, tree, 999999999, 1*time.Minute),
	})

	if got := LiveAnchoredSessions(tree, time.Now().UTC()); len(got) != 1 {
		t.Fatalf("fresh heartbeat must stay protected, got %d entries", len(got))
	}
}

// TestLiveAnchoredSessions_FiltersForeignHostAndCWD: entries from another
// host, or anchored outside the tree, are not this disposal's concern.
func TestLiveAnchoredSessions_FiltersForeignHostAndCWD(t *testing.T) {
	tree := t.TempDir()
	other := t.TempDir()
	writeAnchorRegistry(t, tree, []Entry{
		anchorTestEntry("other-host.example", tree, os.Getpid(), time.Minute),
		anchorTestEntry("", other, os.Getpid(), time.Minute),
	})

	if got := LiveAnchoredSessions(tree, time.Now().UTC()); len(got) != 0 {
		t.Fatalf("foreign-host and outside-tree entries must be dropped, got %d", len(got))
	}
}

// TestLiveAnchoredSessions_MissingRegistryFailsOpen: no registry file in
// the tree means nothing to protect — disposal proceeds.
func TestLiveAnchoredSessions_MissingRegistryFailsOpen(t *testing.T) {
	tree := t.TempDir()
	if got := LiveAnchoredSessions(tree, time.Now().UTC()); got != nil {
		t.Fatalf("missing registry must return nil, got %v", got)
	}
}
