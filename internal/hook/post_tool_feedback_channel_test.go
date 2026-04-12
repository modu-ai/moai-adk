package hook

import (
	"context"
	"encoding/json"
	"testing"

	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"github.com/modu-ai/moai-adk/internal/loop"
)

// mockDiagnosticsCollectorWithLSP is a test double that returns lsphook.Diagnostic results.
type mockDiagnosticsCollectorWithLSP struct {
	diagnostics []lsphook.Diagnostic
	counts      lsphook.SeverityCounts
}

func (m *mockDiagnosticsCollectorWithLSP) GetDiagnostics(_ context.Context, _ string) ([]lsphook.Diagnostic, error) {
	return m.diagnostics, nil
}

func (m *mockDiagnosticsCollectorWithLSP) GetSeverityCounts(diags []lsphook.Diagnostic) lsphook.SeverityCounts {
	return m.counts
}

// TestPostToolHandler_WithFeedbackChannel_EmitsDiagnostics verifies REQ-LL-003:
// PostTool hook emits diagnostics to both systemMessage AND FeedbackChannel.
func TestPostToolHandler_WithFeedbackChannel_EmitsDiagnostics(t *testing.T) {
	t.Parallel()

	diagnostics := []lsphook.Diagnostic{
		{Severity: lsphook.SeverityError, Source: "compiler", Message: "undefined: foo"},
	}
	counts := lsphook.SeverityCounts{Errors: 1}

	coll := &mockDiagnosticsCollectorWithLSP{diagnostics: diagnostics, counts: counts}
	fbCh := loop.NewFeedbackChannel(16)

	h := NewPostToolHandlerWithFeedbackChannel(coll, nil, "", 0, nil, fbCh)

	input := &HookInput{
		SessionID:     "sess-1",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "main.go"}`),
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error = %v, want nil", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}

	// REQ-LL-003: FeedbackChannel must receive a Feedback event.
	fb, ok := fbCh.TryReceive()
	if !ok {
		t.Error("FeedbackChannel must receive a Feedback event after PostTool with Write+diagnostics")
	}
	if len(fb.LSPDiagnostics) == 0 {
		t.Error("Feedback.LSPDiagnostics must be populated from PostTool diagnostics")
	}
}

// TestPostToolHandler_WithFeedbackChannel_NonWriteNoEmit verifies that
// non-Write/Edit tools do not emit to the feedback channel.
func TestPostToolHandler_WithFeedbackChannel_NonWriteNoEmit(t *testing.T) {
	t.Parallel()

	fbCh := loop.NewFeedbackChannel(16)
	h := NewPostToolHandlerWithFeedbackChannel(nil, nil, "", 0, nil, fbCh)

	input := &HookInput{
		SessionID:     "sess-1",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(`{"command": "go test"}`),
	}

	_, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	_, ok := fbCh.TryReceive()
	if ok {
		t.Error("non-Write/Edit tools must not emit to FeedbackChannel")
	}
}

// TestPostToolHandler_WithNilFeedbackChannel_NoopOnDiagnostics verifies that
// nil feedbackCh is safe (no panic, REQ-LL-003 graceful degradation).
func TestPostToolHandler_WithNilFeedbackChannel_NoopOnDiagnostics(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandlerWithFeedbackChannel(nil, nil, "", 0, nil, nil)

	input := &HookInput{
		SessionID:     "sess-1",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "main.go"}`),
	}

	// Must not panic.
	_, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() with nil feedbackCh error = %v, want nil", err)
	}
}
