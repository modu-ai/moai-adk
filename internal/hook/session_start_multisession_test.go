package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// TestSessionStartMultiSessionProtocol verifies that the SessionStart
// handler invokes the 3-step coordination protocol (Register + Purge +
// Query) per SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-013..015.
//
// AC-COORD-007 verification.
func TestSessionStartMultiSessionProtocol(t *testing.T) {
	projectDir := t.TempDir()
	// Ensure .moai/state exists for the registry to write into.
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := NewSessionStartHandler(nil) // ConfigProvider may be nil — handler is resilient.
	input := &HookInput{
		SessionID:  "uuid-multisession-test-001",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// The handler MUST have populated the registry.
	registryPath := filepath.Join(projectDir, session.DefaultRegistryPath)
	data, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("registry file not created: %v", err)
	}
	var entries []session.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("registry invalid JSON: %v (data=%s)", err, string(data))
	}
	if len(entries) != 1 {
		t.Errorf("registry entry count: got %d, want 1", len(entries))
	}
	if entries[0].SessionID != input.SessionID {
		t.Errorf("registry session_id: got %q, want %q", entries[0].SessionID, input.SessionID)
	}
	if entries[0].SpecID != session.SpecIDNone {
		t.Errorf("registry spec_id at register: got %q, want %q", entries[0].SpecID, session.SpecIDNone)
	}
	if entries[0].Phase != session.PhaseNone {
		t.Errorf("registry phase at register: got %q, want %q", entries[0].Phase, session.PhaseNone)
	}

	// hook output Data should contain the multi_session_register marker.
	if out == nil || len(out.Data) == 0 {
		t.Fatal("hook output empty")
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Data, &payload); err != nil {
		t.Fatalf("hook output invalid JSON: %v", err)
	}
	if payload["multi_session_register"] != "ok" {
		t.Errorf("hook output missing multi_session_register=ok. Got: %v", payload)
	}
}

// TestSessionStartMultiSessionProtocolWithExistingEntry verifies that
// when another session is already registered, the new session detects it
// (the data map carries multi_session_other_active=N).
func TestSessionStartMultiSessionProtocolWithExistingEntry(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Pre-populate registry with an "other" session entry.
	registryPath := filepath.Join(projectDir, session.DefaultRegistryPath)
	otherReg := session.NewRegistry(registryPath, nil)
	if err := otherReg.Register("uuid-other-1234", "SPEC-X", "plan"); err != nil {
		t.Fatalf("pre-register other: %v", err)
	}

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "uuid-current-5678",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// Verify the data map carries the other-active count (1).
	var payload map[string]any
	if err := json.Unmarshal(out.Data, &payload); err != nil {
		t.Fatal(err)
	}
	other, ok := payload["multi_session_other_active"]
	if !ok {
		t.Errorf("hook output missing multi_session_other_active key. Got: %v", payload)
	}
	// json.Unmarshal decodes integers as float64.
	if f, ok := other.(float64); !ok || int(f) != 1 {
		t.Errorf("multi_session_other_active: got %v, want 1", other)
	}
}

// TestSessionStartMultiSessionProtocolEmptySessionID verifies that an
// empty SessionID short-circuits the protocol (defensive — empty
// session_id means Claude Code did not provide one, registry should not
// be polluted with empty entries).
func TestSessionStartMultiSessionProtocolEmptySessionID(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	registryPath := filepath.Join(projectDir, session.DefaultRegistryPath)
	if _, err := os.Stat(registryPath); err == nil {
		// File created — verify it's empty array.
		data, _ := os.ReadFile(registryPath)
		trimmed := strings.TrimSpace(string(data))
		if trimmed != "" && trimmed != "[]" && trimmed != "null" {
			t.Errorf("registry should not contain entries for empty session_id; got %q", trimmed)
		}
	}
}
