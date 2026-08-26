// Package statusline tests for SPEC-SESSION-TELEMETRY-001: the per-session
// telemetry record — its path, its key, the model and effort it carries, and
// its refusal to accept a key that would escape the per-session directory.
package statusline

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// buildForSession renders one statusline for sessionID against projDir. Ambient
// GLM env is scrubbed so the synthetic window is not overridden by a dev
// machine running `moai glm`.
func buildForSession(t *testing.T, projDir, sessionID string) {
	t.Helper()
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	in := StdinData{
		SessionID: sessionID,
		Workspace: &WorkspaceInfo{CurrentDir: projDir},
		ContextWindow: &ContextWindowInfo{
			ContextWindowSize: 256000,
			UsedPercentage:    new(90.0),
		},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	b := &defaultBuilder{renderer: NewRenderer("default", true, nil), mode: ModeDefault}
	line, err := b.Build(context.Background(), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if line == "" {
		t.Fatal("Build() returned an empty statusline")
	}
}

// TestPerSessionRecordPath — AC-ST-001 (REQ-ST-001).
// The render persists to .moai/state/context-usage/<session-id>.json, and the
// single-slot .moai/state/context-usage.json does NOT exist afterwards (the
// hard cut: D-3 chose no dual-write window).
func TestPerSessionRecordPath(t *testing.T) {
	proj := t.TempDir()
	buildForSession(t, proj, "sess-per-session")

	stateDir := filepath.Join(proj, ".moai", "state")
	perSession := filepath.Join(stateDir, "context-usage", "sess-per-session.json")
	if _, err := os.Stat(perSession); err != nil {
		t.Fatalf("per-session record %q not written: %v", perSession, err)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "context-usage.json")); err == nil {
		t.Error("single-slot context-usage.json exists — the cut is not hard (D-3)")
	}

	rec, err := ReadSessionTelemetry(perSession)
	if err != nil {
		t.Fatalf("ReadSessionTelemetry: %v", err)
	}
	if rec.SessionID != "sess-per-session" {
		t.Errorf("session_id = %q, want sess-per-session", rec.SessionID)
	}
}

// TestRecordKeyIsThePayloadSessionID — AC-ST-002 (REQ-ST-002).
// The key is the identifier the render payload delivered (S), never the
// project-wide .moai/state/current-session-id.txt sidecar (T), whose
// last-writer-wins shape is the defect this SPEC removes.
func TestRecordKeyIsThePayloadSessionID(t *testing.T) {
	proj := t.TempDir()
	stateDir := filepath.Join(proj, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "current-session-id.txt"), []byte("sidecar-T"), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}

	buildForSession(t, proj, "payload-S")

	if _, err := os.Stat(filepath.Join(stateDir, "context-usage", "payload-S.json")); err != nil {
		t.Fatalf("record not keyed by the payload session id: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "context-usage", "sidecar-T.json")); err == nil {
		t.Error("a record named for the sidecar id exists — the key was sourced from the sidecar")
	}
}
