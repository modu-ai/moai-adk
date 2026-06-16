package hook

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// TestSessionStartEmptySessionIDEmitsWarning is the M1 characterization test
// for REQ-WPR-003 (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001).
//
// When the SessionStart hook handler receives input.SessionID == "", the
// multi-session protocol is bypassed (the gate at session_start.go:66). This
// is a leading hypothesis for K6 (empty registry): some Claude Code
// activation paths may emit an empty session_id. The handler MUST emit a
// non-blocking stderr warning so the orchestrator can observe that the write
// path was bypassed (AC-WPR-003).
//
// RED phase: before the fix, no warning is emitted; after GREEN the warning
// is present on stderr and the hook still exits 0 (non-blocking).
func TestSessionStartEmptySessionIDEmitsWarning(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Capture stderr — the warning is non-blocking and goes to stderr.
	oldStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = wErr

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "", // empty — the K3 gate bypass scenario
		CWD:        projectDir,
		ProjectDir: projectDir,
	}

	out, handleErr := h.Handle(context.Background(), input)
	// Close writer and restore stderr before reading so the pipe drains.
	_ = wErr.Close()
	os.Stderr = oldStderr

	if handleErr != nil {
		t.Fatalf("Handle returned error (must be non-blocking): %v", handleErr)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	// Read captured stderr.
	var stderrBuf bytes.Buffer
	if _, copyErr := stderrBuf.ReadFrom(rErr); copyErr != nil {
		t.Fatalf("read stderr pipe: %v", copyErr)
	}
	_ = rErr.Close()

	stderrText := stderrBuf.String()
	// AC-WPR-003: a warning MUST be emitted to stderr when SessionID is empty.
	if !strings.Contains(strings.ToLower(stderrText), "session_id") {
		t.Errorf("REQ-WPR-003: expected stderr warning mentioning session_id when input.SessionID is empty.\nstderr: %q", stderrText)
	}
	if !strings.Contains(strings.ToLower(stderrText), "empty") && !strings.Contains(strings.ToLower(stderrText), "missing") {
		t.Errorf("REQ-WPR-003: stderr warning should describe the empty/missing session_id condition.\nstderr: %q", stderrText)
	}

	// The registry MUST NOT be polluted with an empty-session_id entry.
	registryPath := filepath.Join(projectDir, session.DefaultRegistryPath)
	if data, statErr := os.ReadFile(registryPath); statErr == nil {
		trimmed := strings.TrimSpace(string(data))
		if trimmed != "" && trimmed != "[]" && trimmed != "null" {
			t.Errorf("registry should not contain entries for empty session_id; got %q", trimmed)
		}
	}
}
