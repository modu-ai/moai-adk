package session

import (
	"os"
	"testing"
	"time"
)

// TestRelocateSession_UpdatesCwd: the matching entry's CWD moves; other
// entries are untouched.
func TestRelocateSession_UpdatesCwd(t *testing.T) {
	reg := NewRegistry(t.TempDir()+"/.moai/state/active-sessions.json", nil)
	if err := reg.Register("sess-a", "SPEC-X", "run"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.Register("sess-b", "SPEC-Y", "plan"); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := reg.RelocateSession("sess-a", "/new/tree"); err != nil {
		t.Fatalf("RelocateSession: %v", err)
	}

	entries, err := reg.Query("")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	for _, e := range entries {
		switch e.SessionID {
		case "sess-a":
			if e.CWD != "/new/tree" {
				t.Errorf("sess-a CWD = %q, want /new/tree", e.CWD)
			}
		case "sess-b":
			if e.CWD == "/new/tree" {
				t.Errorf("sess-b CWD must be untouched, got %q", e.CWD)
			}
		}
	}
}

// TestRelocateSession_MissingEntryNoOp mirrors the Heartbeat contract
// (REQ-COORD-004): relocating an unregistered session is not an error.
func TestRelocateSession_MissingEntryNoOp(t *testing.T) {
	reg := NewRegistry(t.TempDir()+"/.moai/state/active-sessions.json", nil)
	if err := reg.Register("sess-a", "SPEC-X", "run"); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := reg.RelocateSession("sess-unknown", "/new/tree"); err != nil {
		t.Fatalf("RelocateSession on missing entry must be a no-op, got: %v", err)
	}

	entries, err := reg.Query("")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(entries) != 1 || entries[0].SessionID != "sess-a" {
		t.Fatalf("registry must be unchanged, got %+v", entries)
	}
}

// TestRelocateSession_EmptyArgsReject: empty identifiers are caller bugs,
// not registry misses.
func TestRelocateSession_EmptyArgsReject(t *testing.T) {
	reg := NewRegistry(t.TempDir()+"/.moai/state/active-sessions.json", nil)
	if err := reg.RelocateSession("", "/new/tree"); err == nil {
		t.Error("empty sessionID must be rejected")
	}
	if err := reg.RelocateSession("sess-a", ""); err == nil {
		t.Error("empty newCwd must be rejected")
	}
}

// TestLiveAnchoredSessions_CallerRegistrySource pins the cc-w lane shape at
// the session layer: the entry lives in the CALLER's project registry while
// the target tree has no local registry at all.
func TestLiveAnchoredSessions_CallerRegistrySource(t *testing.T) {
	tree := t.TempDir()
	launcher := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", launcher)
	host, _ := os.Hostname()
	writeAnchorRegistry(t, launcher, []Entry{
		anchorTestEntry(host, tree, os.Getpid(), 2*time.Hour),
	})

	got := LiveAnchoredSessions(tree, time.Now().UTC())
	if len(got) != 1 {
		t.Fatalf("caller-registry entry with cwd=tree must be detected, got %d entries", len(got))
	}
}
