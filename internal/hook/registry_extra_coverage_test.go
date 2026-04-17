package hook

// registry_extra_coverage_test.go: SetTraceWriter, EnableObservability,
// effectiveDecision, effectiveReason, writeTrace, ensureTraceWriter 커버리지

import (
	"context"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// --- SetTraceWriter ---

// TestSetTraceWriter_NilWriter sets nil without panic.
func TestSetTraceWriter_NilWriter(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigManager()
	reg := NewRegistry(cfg)
	reg.SetTraceWriter(nil)
	// no panic
}

// TestSetTraceWriter_ThenNil sets a writer then clears it.
func TestSetTraceWriter_ThenNil(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigManager()
	reg := NewRegistry(cfg)
	dir := t.TempDir()
	// Create a real TraceWriter using the exported NewRegistry path via EnableObservability.
	reg.EnableObservability(dir)
	// Reset to nil.
	reg.SetTraceWriter(nil)
	// no panic
}

// --- EnableObservability ---

// TestEnableObservability_SetsLogDir verifies logDir is stored.
func TestEnableObservability_SetsLogDir(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigManager()
	reg := NewRegistry(cfg)
	dir := t.TempDir()
	// After calling EnableObservability, Dispatch with sessionID should trigger
	// ensureTraceWriter; we only verify no panic here.
	reg.EnableObservability(dir)
}

// --- effectiveDecision / effectiveReason ---

// TestEffectiveDecision_NilHookSpecific returns top-level Decision.
func TestEffectiveDecision_NilHookSpecific(t *testing.T) {
	t.Parallel()

	out := &HookOutput{Decision: "allow"}
	got := effectiveDecision(out)
	if got != "allow" {
		t.Errorf("effectiveDecision() = %q, want %q", got, "allow")
	}
}

// TestEffectiveDecision_WithHookSpecific prefers PermissionDecision.
func TestEffectiveDecision_WithHookSpecific(t *testing.T) {
	t.Parallel()

	out := &HookOutput{
		Decision: "allow",
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision: "deny",
		},
	}
	got := effectiveDecision(out)
	if got != "deny" {
		t.Errorf("effectiveDecision() = %q, want %q", got, "deny")
	}
}

// TestEffectiveDecision_EmptyHookSpecific falls back to top-level.
func TestEffectiveDecision_EmptyHookSpecific(t *testing.T) {
	t.Parallel()

	out := &HookOutput{
		Decision: "block",
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision: "",
		},
	}
	got := effectiveDecision(out)
	if got != "block" {
		t.Errorf("effectiveDecision() = %q, want %q", got, "block")
	}
}

// TestEffectiveReason_NilHookSpecific returns top-level Reason.
func TestEffectiveReason_NilHookSpecific(t *testing.T) {
	t.Parallel()

	out := &HookOutput{Reason: "user denied"}
	got := effectiveReason(out)
	if got != "user denied" {
		t.Errorf("effectiveReason() = %q, want %q", got, "user denied")
	}
}

// TestEffectiveReason_WithHookSpecific prefers PermissionDecisionReason.
func TestEffectiveReason_WithHookSpecific(t *testing.T) {
	t.Parallel()

	out := &HookOutput{
		Reason: "fallback",
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecisionReason: "blocked by policy",
		},
	}
	got := effectiveReason(out)
	if got != "blocked by policy" {
		t.Errorf("effectiveReason() = %q, want %q", got, "blocked by policy")
	}
}

// TestEffectiveReason_EmptyHookSpecific falls back to top-level.
func TestEffectiveReason_EmptyHookSpecific(t *testing.T) {
	t.Parallel()

	out := &HookOutput{
		Reason: "top-level",
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecisionReason: "",
		},
	}
	got := effectiveReason(out)
	if got != "top-level" {
		t.Errorf("effectiveReason() = %q, want %q", got, "top-level")
	}
}

// --- writeTrace (via Dispatch with observability enabled) ---

// TestWriteTrace_NilTraceWriter exercises writeTrace with no writer set.
func TestWriteTrace_NilTraceWriter(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigManager()
	reg := NewRegistry(cfg)
	// writeTrace is called inside Dispatch; with no traceWriter it's a no-op.
	h := NewGenericHandler(EventSessionStart)
	reg.Register(h)

	input := &HookInput{
		SessionID: "test-write-trace-001",
		CWD:       t.TempDir(),
	}
	_, _ = reg.Dispatch(context.Background(), EventSessionStart, input)
	// no panic
}

// TestWriteTrace_WithObservability exercises writeTrace via Dispatch with real TraceWriter.
func TestWriteTrace_WithObservability(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigManager()
	reg := NewRegistry(cfg)
	logDir := t.TempDir()
	reg.EnableObservability(logDir)

	h := NewGenericHandler(EventSessionStart)
	reg.Register(h)

	input := &HookInput{
		SessionID: "obs-dispatch-001",
		CWD:       t.TempDir(),
	}
	_, _ = reg.Dispatch(context.Background(), EventSessionStart, input)
	// writeTrace should have been called (non-nil tw after ensureTraceWriter)
}

// TestEnsureTraceWriter_EmptySessionID is a no-op for empty sessionID.
func TestEnsureTraceWriter_EmptySessionID(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigManager()
	reg := NewRegistry(cfg)
	logDir := t.TempDir()
	reg.EnableObservability(logDir)

	h := NewGenericHandler(EventSessionStart)
	reg.Register(h)

	// sessionID="" → ensureTraceWriter returns early
	input := &HookInput{
		SessionID: "",
		CWD:       t.TempDir(),
	}
	_, _ = reg.Dispatch(context.Background(), EventSessionStart, input)
}

// --- NewSessionEndHandlerWithObservability ---

// TestNewSessionEndHandlerWithObservability_ReturnsHandler verifies non-nil handler.
func TestNewSessionEndHandlerWithObservability_ReturnsHandler(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	h := NewSessionEndHandlerWithObservability(dir, dir)
	if h == nil {
		t.Fatal("NewSessionEndHandlerWithObservability() returned nil")
	}
	if h.EventType() != EventSessionEnd {
		t.Errorf("EventType() = %v, want EventSessionEnd", h.EventType())
	}
}

// TestNewSessionEndHandlerWithObservability_Handle_NoTraceDir runs Handle with empty session ID.
func TestNewSessionEndHandlerWithObservability_Handle_NoTraceDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	h := NewSessionEndHandlerWithObservability(dir, dir)

	input := &HookInput{
		SessionID:  "", // empty → no generateSessionSummary call
		ProjectDir: t.TempDir(),
		CWD:        t.TempDir(),
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// TestNewSessionEndHandlerWithObservability_Handle_WithSessionID calls generateSessionSummary.
// The trace dir is empty, so GenerateSummary returns TotalHooks=0 → early return (no-op).
func TestNewSessionEndHandlerWithObservability_Handle_WithSessionID(t *testing.T) {
	t.Parallel()

	traceDir := t.TempDir()
	reportDir := t.TempDir()
	h := NewSessionEndHandlerWithObservability(traceDir, reportDir)

	input := &HookInput{
		SessionID:  "sess-obs-001",
		ProjectDir: t.TempDir(),
		CWD:        t.TempDir(),
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// --- generateSessionSummary (direct call) ---

// TestGenerateSessionSummary_EmptyTraceDir returns early (TotalHooks == 0, no file written).
func TestGenerateSessionSummary_EmptyTraceDir(t *testing.T) {
	t.Parallel()

	traceDir := t.TempDir()
	reportDir := t.TempDir()
	// No trace file → GenerateSummary returns empty summary → TotalHooks == 0 → no-op
	generateSessionSummary("test-session-001", traceDir, reportDir)

	// Verify no report file was written.
	entries, _ := os.ReadDir(reportDir)
	if len(entries) != 0 {
		t.Errorf("expected no report files, got %d", len(entries))
	}
}

// --- WorktreeCreate / WorktreeRemove with CWD + WorktreePath set ---

// TestWorktreeCreateHandler_Handle_WithWorktreePath exercises the registerWorktree branch.
func TestWorktreeCreateHandler_Handle_WithWorktreePath(t *testing.T) {
	t.Parallel()

	h := NewWorktreeCreateHandler()
	dir := t.TempDir()
	input := &HookInput{
		SessionID:      "sess-wt-create-002",
		CWD:            dir,
		WorktreePath:   dir + "/wt-path",
		WorktreeBranch: "feature/test",
		AgentName:      "test-agent",
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}
}

// TestWorktreeRemoveHandler_Handle_WithWorktreePath exercises the unregisterWorktree branch.
func TestWorktreeRemoveHandler_Handle_WithWorktreePath(t *testing.T) {
	t.Parallel()

	h := NewWorktreeRemoveHandler()
	dir := t.TempDir()
	input := &HookInput{
		SessionID:    "sess-wt-remove-002",
		CWD:          dir,
		WorktreePath: dir + "/wt-path",
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}
}
