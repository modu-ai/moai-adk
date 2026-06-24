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

// TestSessionStartInjectsAdditionalContext is the M3 RED-GREEN test for
// REQ-RDP-004 / AC-RDP-004 (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001).
//
// When input.SessionID != "", the SessionStart hook response MUST include
// hookSpecificOutput.AdditionalContext carrying the orchestrator's own UUID
// so the Claude Code runtime surfaces it to the orchestrator at session
// start. The handler MUST also write a side-channel file so `moai session
// current` can read the UUID back later (the additionalContext is lost
// after compaction — the side-channel file persists).
func TestSessionStartInjectsAdditionalContext(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "uuid-m3-additional-context-001",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

	// AC-RDP-004: hookSpecificOutput.AdditionalContext MUST carry the UUID.
	if out.HookSpecificOutput == nil {
		t.Fatal("AC-RDP-004: HookSpecificOutput is nil — AdditionalContext not injected")
	}
	ac := out.HookSpecificOutput.AdditionalContext
	if ac == "" {
		t.Fatal("AC-RDP-004: AdditionalContext is empty — UUID not surfaced to orchestrator")
	}
	if !strings.Contains(ac, input.SessionID) {
		t.Errorf("AC-RDP-004: AdditionalContext should contain the session UUID %q.\nGot: %q", input.SessionID, ac)
	}
}

// TestSessionStartWritesSideChannel verifies the handler writes the
// side-channel file (.moai/state/current-session-id.txt) so `moai session
// current` can resolve the UUID post-compaction (REQ-RDP-002/004).
func TestSessionStartWritesSideChannel(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "uuid-m3-sidecar-002",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	sidecar := filepath.Join(projectDir, session.CurrentSideChannelFile)
	data, err := os.ReadFile(sidecar)
	if err != nil {
		t.Fatalf("REQ-RDP-004: side-channel file not written at %s: %v", sidecar, err)
	}
	got := strings.TrimSpace(string(data))
	if got != input.SessionID {
		t.Errorf("side-channel file content: got %q, want %q", got, input.SessionID)
	}
}

// TestSessionStartAdditionalContextStrictlyAdditive is the M3 test for
// AC-RDP-005 / REQ-RDP-005. The additionalContext injection is STRICTLY
// additive: existing behavior (Register → Purge → Query → stderr surface)
// MUST remain unchanged. This test re-runs the existing multisession
// scenario and additionally verifies the new fields land without
// disturbing the data map markers.
func TestSessionStartAdditionalContextStrictlyAdditive(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "uuid-m3-additive-003",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// Existing data map markers MUST still be present (Register → ok).
	var payload map[string]any
	if err := json.Unmarshal(out.Data, &payload); err != nil {
		t.Fatalf("data map invalid JSON: %v", err)
	}
	if payload["multi_session_register"] != "ok" {
		t.Errorf("AC-RDP-005: existing multi_session_register marker broken. Got: %v", payload)
	}

	// AND the new injection MUST be present.
	if out.HookSpecificOutput == nil || out.HookSpecificOutput.AdditionalContext == "" {
		t.Errorf("AC-RDP-004: AdditionalContext injection missing (additive check)")
	}
}

// TestSessionStartAdditionalContextSkippedOnEmptySessionID verifies the
// injection is gated on input.SessionID != "" (per research.md §D.0/D.1
// P1-outcome implication). When SessionID is empty, no UUID is injected
// and no side-channel file is written (would be an empty UUID).
func TestSessionStartAdditionalContextSkippedOnEmptySessionID(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Suppress the REQ-WPR-003 stderr warning for this test (it writes to
	// real os.Stderr; we only care about the injection-absence here).
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr
	t.Cleanup(func() {
		os.Stderr = oldStderr
		_ = wErr.Close()
		_, _ = rErr.Read(make([]byte, 4096))
		_ = rErr.Close()
	})

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "",
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

	// No AdditionalContext injection when SessionID is empty.
	if out.HookSpecificOutput != nil && out.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("AdditionalContext should NOT be injected when SessionID is empty. Got: %q",
			out.HookSpecificOutput.AdditionalContext)
	}

	// No side-channel file written (would carry an empty UUID).
	sidecar := filepath.Join(projectDir, session.CurrentSideChannelFile)
	if _, statErr := os.Stat(sidecar); statErr == nil {
		t.Errorf("side-channel file should NOT be written when SessionID is empty")
	}
}
