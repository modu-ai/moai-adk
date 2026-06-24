package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// TestSessionCurrentListedInHelp is the M2 test for AC-RDP-001
// (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001).
//
// `moai session --help` MUST list `current` alongside the 5 original verbs
// + `doctor`. 7 subcommands total (5 original + current + doctor).
func TestSessionCurrentListedInHelp(t *testing.T) {
	cmd := newSessionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, verb := range []string{"register", "heartbeat", "deregister", "list", "purge", "current", "doctor"} {
		if !strings.Contains(out, verb) {
			t.Errorf("AC-RDP-001: help output missing verb %q. Got: %s", verb, out)
		}
	}
}

// TestSessionCurrentFallbackWhenNoSideChannel is the M2 test for AC-RDP-006
// + AC-RDP-003 (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001).
//
// When the runtime does NOT expose session.id (no side-channel file), the
// `current` subcommand emits the canonical fallback string (REQ-RDP-006)
// and exits 0 (graceful degradation, NOT an error).
func TestSessionCurrentFallbackWhenNoSideChannel(t *testing.T) {
	dir := withTempRegistry(t)
	// Point CLAUDE_PROJECT_DIR at the temp dir so resolveProjectDir finds it
	// and the side-channel lookup hits the temp tree (where the file is absent).
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	out, err := runSession(t, "current")
	if err != nil {
		t.Fatalf("AC-RDP-006: current must exit 0 on fallback, got err: %v (out=%s)", err, out)
	}
	trimmed := strings.TrimSpace(out)
	if !strings.Contains(trimmed, CanonicalFallbackSessionID) {
		t.Errorf("AC-RDP-006: current should emit the canonical fallback string.\nwant: %q\ngot:  %q", CanonicalFallbackSessionID, trimmed)
	}
}

// TestSessionCurrentJSONFallback verifies the --json path carries the
// canonical fallback + availability flag.
func TestSessionCurrentJSONFallback(t *testing.T) {
	dir := withTempRegistry(t)
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	out, err := runSession(t, "current", "--json")
	if err != nil {
		t.Fatalf("current --json err: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("current --json not valid JSON: %v (out=%s)", err, out)
	}
	if payload["available"] != false {
		t.Errorf("available: got %v, want false (no side-channel)", payload["available"])
	}
	if payload["source"] != "fallback" {
		t.Errorf("source: got %v, want fallback", payload["source"])
	}
	if payload["session_id"] != CanonicalFallbackSessionID {
		t.Errorf("session_id should be canonical fallback. got: %v", payload["session_id"])
	}
	if payload["canonical_fallback"] != CanonicalFallbackSessionID {
		t.Errorf("canonical_fallback mismatch. got: %v", payload["canonical_fallback"])
	}
}

// TestSessionCurrentShowFallbackFlag verifies --show-fallback prints ONLY
// the canonical fallback string (for paste-ready resume emission).
func TestSessionCurrentShowFallbackFlag(t *testing.T) {
	withTempRegistry(t)

	out, err := runSession(t, "current", "--show-fallback")
	if err != nil {
		t.Fatalf("current --show-fallback err: %v", err)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != CanonicalFallbackSessionID {
		t.Errorf("--show-fallback should print ONLY the canonical fallback.\nwant: %q\ngot:  %q", CanonicalFallbackSessionID, trimmed)
	}
}

// TestSessionCurrentReadsSideChannel is the AC-RDP-002 happy-path test
// (post-M3). It writes a side-channel file (simulating the M3 SessionStart
// additionalContext injection) and verifies `current` resolves the UUID.
//
// Per plan.md §F M2 D4 note + research.md §D.0/D.1, AC-RDP-002 GREEN is
// satisfiable post-M3 because the side-channel write is the mechanism that
// makes the UUID available. This test simulates the side-channel file's
// presence (M3 will write it for real).
func TestSessionCurrentReadsSideChannel(t *testing.T) {
	dir := withTempRegistry(t)
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// Simulate the M3 side-channel write.
	sidecar := filepath.Join(dir, session.CurrentSideChannelFile)
	wantUUID := "11111111-2222-3333-4444-555555555555"
	if err := os.MkdirAll(filepath.Dir(sidecar), 0o755); err != nil {
		t.Fatalf("mkdir sidecar dir: %v", err)
	}
	if err := os.WriteFile(sidecar, []byte(wantUUID), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}

	out, err := runSession(t, "current", "--json")
	if err != nil {
		t.Fatalf("current --json err: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("current --json: %v (out=%s)", err, out)
	}
	if payload["available"] != true {
		t.Errorf("AC-RDP-002: available should be true when side-channel exists. got: %v", payload["available"])
	}
	if payload["session_id"] != wantUUID {
		t.Errorf("AC-RDP-002: session_id: got %v, want %q", payload["session_id"], wantUUID)
	}
	if !strings.HasPrefix(payload["source"].(string), "side-channel:") {
		t.Errorf("source should start with side-channel:. got: %v", payload["source"])
	}
}

// TestSessionCurrentCanonicalFallbackConstant pins the exact canonical
// string so doctrine edits (M4+M5) can reference this test as the
// byte-exact anchor (AC-FBC-001/002 byte-parity check).
func TestSessionCurrentCanonicalFallbackConstant(t *testing.T) {
	want := "source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>"
	if CanonicalFallbackSessionID != want {
		t.Errorf("CanonicalFallbackSessionID drifted.\nwant: %q\ngot:  %q", want, CanonicalFallbackSessionID)
	}
}

// Compile-time guard: reference session import so it is not dropped.
var _ = session.DefaultRegistryPath
