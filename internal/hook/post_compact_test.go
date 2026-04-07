package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/memo"
)

func TestPostCompactHandler_EventType(t *testing.T) {
	h := NewPostCompactHandler()
	if h.EventType() != EventPostCompact {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPostCompact)
	}
}

func TestPostCompactHandler_Handle_NoMemo(t *testing.T) {
	t.Parallel()

	h := NewPostCompactHandler()
	input := &HookInput{
		SessionID:     "test-session",
		HookEventName: "PostCompact",
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if len(out.Data) == 0 {
		t.Error("expected non-empty Data field")
	}
	// No memo file => no SystemMessage
	if out.SystemMessage != "" {
		t.Errorf("expected empty SystemMessage when no memo exists, got: %q", out.SystemMessage)
	}
}

func TestPostCompactHandler_Handle_WithMemo(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	// Write a memo via the memo package.
	sections := []memo.Section{
		{Priority: memo.P1Required, Title: "Context", Content: "SPEC-001 active\nphase: run"},
	}
	if err := memo.Write(projectDir, sections); err != nil {
		t.Fatalf("memo.Write() error: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-recover",
		CWD:           projectDir,
		HookEventName: "PostCompact",
	}

	h := NewPostCompactHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.SystemMessage == "" {
		t.Fatal("expected SystemMessage to be set when memo exists")
	}
	if !strings.Contains(out.SystemMessage, "Session Memo") {
		t.Errorf("SystemMessage should contain 'Session Memo', got: %q", out.SystemMessage)
	}
	if !strings.Contains(out.SystemMessage, "SPEC-001 active") {
		t.Errorf("SystemMessage should contain memo content, got: %q", out.SystemMessage)
	}
	if len(out.Data) == 0 {
		t.Error("Data field should be populated even when SystemMessage is set")
	}
}

func TestPostCompactHandler_Handle_EmptyMemoFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	stateDir := filepath.Join(projectDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Write an empty memo file.
	memoPath := filepath.Join(stateDir, "session-memo.md")
	if err := os.WriteFile(memoPath, []byte{}, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-empty-memo",
		CWD:           projectDir,
		HookEventName: "PostCompact",
	}

	h := NewPostCompactHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	// Empty memo => no SystemMessage
	if out.SystemMessage != "" {
		t.Errorf("expected empty SystemMessage for empty memo, got: %q", out.SystemMessage)
	}
}

func TestPostCompactHandler_Handle_SystemMessagePrefix(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	sections := []memo.Section{
		{Priority: memo.P1Required, Title: "Info", Content: "some context"},
	}
	if err := memo.Write(projectDir, sections); err != nil {
		t.Fatalf("memo.Write() error: %v", err)
	}

	input := &HookInput{
		CWD:           projectDir,
		SessionID:     "sess-prefix-check",
		HookEventName: "PostCompact",
	}

	h := NewPostCompactHandler()
	out, _ := h.Handle(context.Background(), input)

	expectedPrefix := "[Session Memo - Restored after context compaction]"
	if !strings.HasPrefix(out.SystemMessage, expectedPrefix) {
		t.Errorf("SystemMessage should start with %q, got: %q", expectedPrefix, out.SystemMessage)
	}
}
