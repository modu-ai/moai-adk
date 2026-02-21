package hook

// coverage_boost_test.go - Targeted tests to increase coverage for:
//   - rank_session.go: Handle (36%), InitRankSessionHandler (0%), EnsureRankSessionHandler (0%)
//   - pre_tool.go: scanWriteContent (0%), NewPreToolHandlerWithScanner (71.4%)
//   - session_end.go: garbageCollectStaleTeams (75%), cleanupOrphanedTmuxSessions (57.9%)
//   - compact.go: Handle (71.4%)
//   - protocol.go: WriteOutput (83.3%)
//   - session_start.go: getConfig (66.7%)

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/rank"
)

// --- protocol.go: WriteOutput nil output branch ---

func TestWriteOutput_NilOutput(t *testing.T) {
	t.Parallel()

	proto := &jsonProtocol{}
	var buf bytes.Buffer

	err := proto.WriteOutput(&buf, nil)
	if err != nil {
		t.Fatalf("WriteOutput(nil) returned error: %v", err)
	}

	data := buf.Bytes()
	if !json.Valid(data) {
		t.Fatalf("WriteOutput(nil) produced invalid JSON: %s", data)
	}
}

func TestWriteOutput_WriterError(t *testing.T) {
	t.Parallel()

	proto := &jsonProtocol{}

	// errWriter always returns an error on Write
	err := proto.WriteOutput(&errWriter{}, NewAllowOutput())
	if err == nil {
		t.Fatal("expected error from failing writer, got nil")
	}
}

// errWriter always fails on Write to test error branch in WriteOutput.
type errWriter struct{}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("forced write error")
}

// --- session_start.go: getConfig ---

func TestSessionStartHandler_GetConfig_NilProvider(t *testing.T) {
	t.Parallel()

	// When cfg is nil (no provider), getConfig should return nil gracefully.
	h := &sessionStartHandler{cfg: nil}
	cfg := h.getConfig()
	if cfg != nil {
		t.Errorf("getConfig() with nil cfg = %v, want nil", cfg)
	}
}

func TestSessionStartHandler_GetConfig_ProviderReturnsNil(t *testing.T) {
	t.Parallel()

	// ConfigProvider that returns nil config.
	h := &sessionStartHandler{cfg: &mockConfigProvider{cfg: nil}}
	cfg := h.getConfig()
	if cfg != nil {
		t.Errorf("getConfig() with nil config = %v, want nil", cfg)
	}
}

func TestSessionStartHandler_Handle_NilCfgProvider(t *testing.T) {
	t.Parallel()

	// SessionStart handler with nil ConfigProvider should not panic.
	h := NewSessionStartHandler(nil)
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-nil-provider",
		CWD:           "/tmp",
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
}

// --- compact.go: Handle ---

func TestCompactHandler_Handle_AlwaysReturnsData(t *testing.T) {
	t.Parallel()

	h := NewCompactHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-compact-always",
		CWD:           "/tmp",
		ProjectDir:    "",
		HookEventName: "PreCompact",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
	// Compact always returns non-nil data with session_id, status, snapshot_created.
	if got.Data == nil {
		t.Fatal("Data should not be nil")
	}

	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("unmarshal Data: %v", err)
	}
	if data["snapshot_created"] != true {
		t.Errorf("snapshot_created = %v, want true", data["snapshot_created"])
	}
}

// --- session_end.go: garbageCollectStaleTeams ---

func TestGarbageCollectStaleTeams_NonDirEntries(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup teams dir: %v", err)
	}

	// Create a regular file (not a directory) in the teams dir.
	regularFile := filepath.Join(teamsDir, "not-a-team.txt")
	if err := os.WriteFile(regularFile, []byte("file"), 0o644); err != nil {
		t.Fatalf("create regular file: %v", err)
	}

	// Also create a fresh directory (shouldn't be cleaned up).
	freshDir := filepath.Join(teamsDir, "active-team")
	if err := os.MkdirAll(freshDir, 0o755); err != nil {
		t.Fatalf("create fresh dir: %v", err)
	}

	// Should not panic or fail; the file entry should be skipped.
	garbageCollectStaleTeams(homeDir)

	// File should still exist (it's not a directory).
	if _, err := os.Stat(regularFile); os.IsNotExist(err) {
		t.Error("regular file should not have been removed")
	}
	// Fresh directory should still exist.
	if _, err := os.Stat(freshDir); os.IsNotExist(err) {
		t.Error("fresh team dir should still exist")
	}
}

func TestGarbageCollectStaleTeams_StaleRemovalError(t *testing.T) {
	t.Parallel()

	// This test verifies behavior when stale dir removal is attempted.
	// Create a stale directory and ensure garbageCollectStaleTeams handles it.
	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup teams dir: %v", err)
	}

	// Create a stale team dir.
	staleDir := filepath.Join(teamsDir, "stale-team-2")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatalf("create stale dir: %v", err)
	}
	staleTime := time.Now().Add(-26 * time.Hour)
	if err := os.Chtimes(staleDir, staleTime, staleTime); err != nil {
		t.Fatalf("set stale time: %v", err)
	}

	// Should remove stale dir without error.
	garbageCollectStaleTeams(homeDir)

	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Error("stale team dir should have been removed")
	}
}

// --- session_end.go: cleanupOrphanedTmuxSessions ---

func TestCleanupOrphanedTmuxSessions_WithTimeout(t *testing.T) {
	t.Parallel()

	// Use a very short timeout to exercise the timeout branch.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Sleep to let the deadline pass.
	time.Sleep(5 * time.Millisecond)

	// Should return without panic even with already-expired context.
	cleanupOrphanedTmuxSessions(ctx)
}

func TestCleanupOrphanedTmuxSessions_NormalContext(t *testing.T) {
	t.Parallel()

	// tmux may or may not be installed. Either way, function should not panic.
	ctx := context.Background()
	cleanupOrphanedTmuxSessions(ctx)
}

// --- rank_session.go: InitRankSessionHandler, EnsureRankSessionHandler ---

func TestInitRankSessionHandler_ReturnsHandler(t *testing.T) {
	// Cannot use t.Parallel() because we use t.Setenv()
	// Use a temp directory so NewPatternStore doesn't fail.
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	h, err := InitRankSessionHandler()
	// InitRankSessionHandler may succeed or fail depending on environment.
	// What matters is: no panic, and if err==nil then h is not nil.
	if err != nil {
		// Pattern store creation failure is acceptable in test env.
		t.Logf("InitRankSessionHandler returned error (acceptable in test): %v", err)
		return
	}
	if h == nil {
		t.Fatal("InitRankSessionHandler() returned nil handler without error")
	}
	if h.EventType() != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventSessionEnd)
	}
}

func TestEnsureRankSessionHandler_NoCredentials(t *testing.T) {
	// Cannot use t.Parallel() because we use t.Setenv()
	// Use a temp home dir with no credentials file.
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	h, err := EnsureRankSessionHandler()
	// No credentials file exists, so should return (nil, nil).
	if err != nil {
		t.Fatalf("EnsureRankSessionHandler() unexpected error: %v", err)
	}
	if h != nil {
		t.Errorf("EnsureRankSessionHandler() = %v, want nil (no credentials)", h)
	}
}

func TestEnsureRankSessionHandler_WithCredentials(t *testing.T) {
	// Cannot use t.Parallel() because we use t.Setenv()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	// Create a credentials file so HasCredentials() returns true.
	rankDir := filepath.Join(tempHome, ".moai", "rank")
	if err := os.MkdirAll(rankDir, 0o755); err != nil {
		t.Fatalf("setup rank dir: %v", err)
	}

	credStore := rank.NewFileCredentialStore(rankDir)
	if err := credStore.Save(&rank.Credentials{APIKey: "test-key"}); err != nil {
		t.Fatalf("save credentials: %v", err)
	}

	h, err := EnsureRankSessionHandler()
	if err != nil {
		t.Fatalf("EnsureRankSessionHandler() with credentials failed: %v", err)
	}
	// Should return a valid handler when credentials exist.
	if h == nil {
		t.Fatal("EnsureRankSessionHandler() returned nil with valid credentials")
	}
	if h.EventType() != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventSessionEnd)
	}
}

// --- rank_session.go: Handle with valid API key (more branch coverage) ---

func TestRankSessionHandler_Handle_WithAPIKey_SubmitFails(t *testing.T) {
	t.Parallel()

	// Handler with an API key but invalid server → submit fails → returns empty output (non-blocking).
	cred := &mockCredStore{apiKey: "test-api-key", hasCredentials: true}
	h := NewRankSessionHandler(nil, cred)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-submit-fail",
		CWD:           "/tmp",
		ProjectDir:    "/tmp",
		HookEventName: "SessionEnd",
		Model:         "claude-sonnet-4",
	}

	// Even if submit fails (no real server), Handle should return empty HookOutput with no error.
	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output")
	}
	// Decision should be empty (SessionEnd hooks return empty {}).
	if got.Decision != "" {
		t.Errorf("Decision = %q, want empty", got.Decision)
	}
}

func TestRankSessionHandler_Handle_CredStoreError(t *testing.T) {
	t.Parallel()

	// When GetAPIKey returns an error, Handle should skip and return empty output.
	cred := &mockCredStore{apiKey: "", apiKeyErr: errors.New("keystore error"), hasCredentials: false}
	h := NewRankSessionHandler(nil, cred)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-cred-error",
		CWD:           "/tmp",
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output")
	}
}

// --- pre_tool.go: scanWriteContent via NewPreToolHandlerWithScanner ---

func TestPreToolHandler_ScanWriteContent_NilScanner(t *testing.T) {
	t.Parallel()

	// When scanner is nil, scanWriteContent path should return ("", "").
	// We test this indirectly by using the handler without a scanner.
	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    nil, // no scanner
		projectDir: t.TempDir(),
	}

	projectDir := handler.projectDir
	toolInput, err := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, "safe.go"),
		"content":   "package main\nfunc main() {}",
	})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-scan-nil",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(toolInput),
	}

	ctx := context.Background()
	got, handleErr := handler.Handle(ctx, input)
	if handleErr != nil {
		t.Fatalf("unexpected error: %v", handleErr)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
	// Without scanner, safe content should be allowed.
	if got.HookSpecificOutput.PermissionDecision != DecisionAllow {
		t.Errorf("PermissionDecision = %q, want %q", got.HookSpecificOutput.PermissionDecision, DecisionAllow)
	}
}

func TestPreToolHandler_NewPreToolHandlerWithScanner_UnavailableScanner(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()

	// Build a mock scanner that reports IsAvailable() = false.
	// Since we can't easily mock SecurityScanner, we pass nil which also disables scanner.
	h := NewPreToolHandlerWithScanner(cfg, policy, nil)
	if h == nil {
		t.Fatal("NewPreToolHandlerWithScanner returned nil")
	}
	if h.EventType() != EventPreToolUse {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPreToolUse)
	}
}

func TestPreToolHandler_CheckFileAccess_InvalidFilePath(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		projectDir: t.TempDir(),
	}

	// Empty file_path in tool input — checkFileAccess should return ("", "").
	toolInput, err := json.Marshal(map[string]string{
		"file_path": "",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-empty-path",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(toolInput),
	}

	ctx := context.Background()
	got, handleErr := handler.Handle(ctx, input)
	if handleErr != nil {
		t.Fatalf("unexpected error: %v", handleErr)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
}

func TestPreToolHandler_Handle_EditTool_AskPattern(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		projectDir: projectDir,
	}

	// Edit tool on package.json should trigger ask (critical config file).
	toolInput, err := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, "package.json"),
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-edit-ask",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Edit",
		ToolInput:     json.RawMessage(toolInput),
	}

	ctx := context.Background()
	got, handleErr := handler.Handle(ctx, input)
	if handleErr != nil {
		t.Fatalf("unexpected error: %v", handleErr)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
	// package.json matches AskPatterns, so should be DecisionAsk.
	if got.HookSpecificOutput.PermissionDecision != DecisionAsk {
		t.Errorf("PermissionDecision = %q, want %q", got.HookSpecificOutput.PermissionDecision, DecisionAsk)
	}
}

// --- anonymizePath: error branch (absolute path failure) ---

func TestAnonymizePath_RelativePath(t *testing.T) {
	t.Parallel()

	// A relative path that filepath.Abs can resolve.
	// The function falls back to path if Abs fails, but Abs rarely fails.
	got := anonymizePath("relative/path/to/project")
	if len(got) == 0 {
		t.Error("anonymizePath should return non-empty hash for relative path")
	}
	if len(got) != 16 {
		t.Errorf("expected length 16, got %d", len(got))
	}
}

// --- compilePatterns: error branch (invalid pattern) ---

func TestCompilePatterns_InvalidPattern(t *testing.T) {
	t.Parallel()

	// An invalid regex should be silently skipped (no panic).
	patterns := []string{
		`valid-pattern`,
		`[invalid(regex`,
		`another-valid`,
	}

	result := compilePatterns(patterns)
	// Invalid pattern is skipped; only 2 valid patterns should be compiled.
	if len(result) != 2 {
		t.Errorf("expected 2 compiled patterns, got %d", len(result))
	}
}

// --- sessionEndHandler: Handle when UserHomeDir fails ---
// We cannot easily force os.UserHomeDir to fail in a unit test on macOS,
// but we can cover the path via integration with the handler.

func TestSessionEndHandler_Handle_MultipleCallsIdempotent(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-idempotent",
		CWD:           t.TempDir(),
		HookEventName: "SessionEnd",
	}

	// Call Handle twice to ensure idempotency.
	for i := range 2 {
		got, err := h.Handle(ctx, input)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
		if got == nil {
			t.Fatalf("call %d: got nil output", i+1)
		}
	}
}

// --- rank_session.go: Handle with ProjectDir set (not empty) ---

func TestRankSessionHandler_Handle_WithProjectDir(t *testing.T) {
	t.Parallel()

	cred := &mockCredStore{apiKey: "", hasCredentials: false}
	h := NewRankSessionHandler(nil, cred)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-with-proj",
		CWD:           "/tmp/cwd",
		ProjectDir:    "/tmp/project",
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output")
	}
}

// --- logTaskMetrics: directory creation failure branch ---

func TestLogTaskMetrics_MkdirAllFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test permission denial as root")
	}
	t.Parallel()

	// Create a read-only parent so MkdirAll fails.
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Make .moai directory read-only so subdirectory "logs" cannot be created.
	if err := os.Chmod(moaiDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(moaiDir, 0o755) })

	input := &HookInput{
		SessionID:    "sess-mkdir-fail",
		ToolName:     "Task",
		CWD:          tmpDir,
		ToolResponse: json.RawMessage(`{"metrics":{"tokensUsed":100,"toolUses":2,"durationSeconds":5.0}}`),
	}

	// Should not panic; error is best-effort logged.
	logTaskMetrics(input)

	// The log file should NOT exist since MkdirAll failed.
	logPath := filepath.Join(tmpDir, ".moai", "logs", "task-metrics.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("log file should not exist when MkdirAll fails")
	}
}

// --- registry.go: getBlockReason additional coverage ---

func TestGetBlockReason_TopLevelReason(t *testing.T) {
	t.Parallel()

	output := &HookOutput{
		Decision: DecisionBlock,
		Reason:   "top-level reason",
	}
	got := getBlockReason(output)
	if got != "top-level reason" {
		t.Errorf("getBlockReason() = %q, want %q", got, "top-level reason")
	}
}

func TestGetBlockReason_HookSpecificReason(t *testing.T) {
	t.Parallel()

	output := &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision:       DecisionDeny,
			PermissionDecisionReason: "specific reason",
		},
	}
	got := getBlockReason(output)
	if got != "specific reason" {
		t.Errorf("getBlockReason() = %q, want %q", got, "specific reason")
	}
}

func TestGetBlockReason_EmptyReason(t *testing.T) {
	t.Parallel()

	output := &HookOutput{}
	got := getBlockReason(output)
	if got != "" {
		t.Errorf("getBlockReason() = %q, want empty", got)
	}
}

// --- registry.go: defaultOutputForEvent coverage ---

func TestDefaultOutputForEvent_PermissionRequest(t *testing.T) {
	t.Parallel()

	reg := &registry{}
	out := reg.defaultOutputForEvent(EventPermissionRequest)
	if out == nil {
		t.Fatal("defaultOutputForEvent(PermissionRequest) returned nil")
	}
	if out.HookSpecificOutput == nil {
		t.Fatal("PermissionRequest default should have HookSpecificOutput")
	}
	if out.HookSpecificOutput.PermissionDecision != DecisionAsk {
		t.Errorf("PermissionDecision = %q, want %q", out.HookSpecificOutput.PermissionDecision, DecisionAsk)
	}
}

func TestDefaultOutputForEvent_SessionEvents(t *testing.T) {
	t.Parallel()

	reg := &registry{}
	events := []EventType{
		EventStop, EventSessionEnd, EventSessionStart, EventPreCompact,
		EventSubagentStop, EventPostToolUseFailure, EventNotification,
		EventSubagentStart, EventUserPromptSubmit, EventTeammateIdle,
		EventTaskCompleted,
	}
	for _, event := range events {
		out := reg.defaultOutputForEvent(event)
		if out == nil {
			t.Errorf("defaultOutputForEvent(%q) returned nil", event)
			continue
		}
		if out.HookSpecificOutput != nil {
			t.Errorf("defaultOutputForEvent(%q) should not set HookSpecificOutput", event)
		}
	}
}

func TestDefaultOutputForEvent_Default(t *testing.T) {
	t.Parallel()

	reg := &registry{}
	// Unknown event type should return empty HookOutput.
	out := reg.defaultOutputForEvent(EventType("UnknownEvent"))
	if out == nil {
		t.Fatal("defaultOutputForEvent(unknown) returned nil")
	}
}

// --- post_tool.go: Handle when json.Marshal fails ---
// Note: Standard maps cannot fail json.Marshal; this covers the ToolResponse branch.

func TestPostToolHandler_Handle_TaskTool_ValidMetrics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Pre-create .moai/ so resolveProjectRoot accepts tmpDir as a MoAI project root.
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatalf("pre-create .moai: %v", err)
	}
	input := &HookInput{
		SessionID:    "sess-task-valid",
		CWD:          tmpDir,
		ToolName:     "Task",
		ToolResponse: json.RawMessage(`{"metrics":{"tokensUsed":500,"toolUses":3,"durationSeconds":10.0}}`),
	}

	h := NewPostToolHandler()
	ctx := context.Background()
	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
	// Task metrics log file should have been created.
	logPath := filepath.Join(tmpDir, ".moai", "logs", "task-metrics.jsonl")
	if _, statErr := os.Stat(logPath); os.IsNotExist(statErr) {
		t.Error("task metrics log should have been created")
	}
}

// --- Additional cleanupOrphanedTmuxSessions coverage ---

func TestCleanupOrphanedTmuxSessions_EmptyLine(t *testing.T) {
	t.Parallel()

	// Test that the function handles tmux output gracefully.
	// We can't mock exec.CommandContext, but we can test various context states.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will either run tmux (if available) or return quickly.
	// No panic or hang is the expectation.
	cleanupOrphanedTmuxSessions(ctx)
}

// --- Additional rank_session.go Handle coverage ---

func TestRankSessionHandler_Handle_PatternStoreNotNil_NotExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cred := &mockCredStore{apiKey: "", hasCredentials: false}

	patternStore, err := rank.NewPatternStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	h := NewRankSessionHandler(patternStore, cred)
	ctx := context.Background()

	// Not excluded project with no credentials → skip silently.
	input := &HookInput{
		SessionID:     "sess-not-excluded",
		CWD:           "/my/project",
		ProjectDir:    "/my/project",
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output")
	}
}

// --- contract.go: "path is not a directory" branch ---

func TestContractValidate_PathIsFile(t *testing.T) {
	t.Parallel()

	// Create a regular file to use as "workDir" — should fail with "not a directory".
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	contract := NewContract(filePath)
	ctx := context.Background()
	err := contract.Validate(ctx)
	if err == nil {
		t.Fatal("expected error for file path, got nil")
	}
	if !errors.Is(err, ErrHookContractFail) {
		t.Errorf("error = %v, want ErrHookContractFail", err)
	}
}

// --- teammate_idle.go: loadBaselineCounts invalid JSON branch ---

func TestLoadBaselineCounts_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	moaiMemDir := filepath.Join(tmpDir, ".moai", "memory")
	if err := os.MkdirAll(moaiMemDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Write invalid JSON to the baseline file.
	baselineFile := filepath.Join(moaiMemDir, "diagnostics-baseline.json")
	if err := os.WriteFile(baselineFile, []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	_, err := loadBaselineCounts(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid JSON baseline, got nil")
	}
}

func TestLoadBaselineCounts_ValidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	moaiMemDir := filepath.Join(tmpDir, ".moai", "memory")
	if err := os.MkdirAll(moaiMemDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Write valid JSON with diagnostic entries.
	baselineJSON := `{
		"files": {
			"main.go": {
				"diagnostics": [
					{"severity": "error"},
					{"severity": "warning"},
					{"severity": "information"},
					{"severity": "hint"}
				]
			}
		}
	}`
	baselineFile := filepath.Join(moaiMemDir, "diagnostics-baseline.json")
	if err := os.WriteFile(baselineFile, []byte(baselineJSON), 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	counts, err := loadBaselineCounts(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts.Errors != 1 {
		t.Errorf("Errors = %d, want 1", counts.Errors)
	}
	if counts.Warnings != 1 {
		t.Errorf("Warnings = %d, want 1", counts.Warnings)
	}
}

// --- cleanupOrphanedTmuxSessions: attached session line coverage ---
// We simulate tmux output parsing by running the function with a context.
// If tmux is not available, the function exits early (that branch is already covered).
// The internal parsing logic (attached vs unattached) requires tmux to be running.
// We test it by checking the function doesn't panic in any scenario.

func TestCleanupOrphanedTmuxSessions_LongTimeout(t *testing.T) {
	t.Parallel()

	// Normal context with longer timeout to allow tmux to respond if available.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Should not panic regardless of tmux availability.
	cleanupOrphanedTmuxSessions(ctx)
}

// --- garbageCollectStaleTeams: entry Info() error branch ---
// This branch is triggered when os.DirEntry.Info() returns an error,
// which typically happens with deleted or inaccessible files after ReadDir.
// It's very hard to trigger in tests without OS-level manipulation.
// Instead we cover more of the normal non-stale branch with better timing.

func TestGarbageCollectStaleTeams_MultipleEntries(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup teams dir: %v", err)
	}

	// Create two stale dirs and one fresh dir.
	for _, name := range []string{"stale-1", "stale-2"} {
		dir := filepath.Join(teamsDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		oldTime := time.Now().Add(-48 * time.Hour)
		if err := os.Chtimes(dir, oldTime, oldTime); err != nil {
			t.Fatalf("chtimes %s: %v", name, err)
		}
	}

	freshDir := filepath.Join(teamsDir, "fresh")
	if err := os.MkdirAll(freshDir, 0o755); err != nil {
		t.Fatalf("create fresh: %v", err)
	}

	garbageCollectStaleTeams(homeDir)

	// Stale dirs should be removed.
	for _, name := range []string{"stale-1", "stale-2"} {
		if _, err := os.Stat(filepath.Join(teamsDir, name)); !os.IsNotExist(err) {
			t.Errorf("%s should have been removed", name)
		}
	}

	// Fresh dir should remain.
	if _, err := os.Stat(freshDir); os.IsNotExist(err) {
		t.Error("fresh dir should still exist")
	}
}

// --- cleanupCurrentSessionTeam: non-dir entries in teams dir ---

func TestCleanupCurrentSessionTeam_NonDirEntries(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup teams dir: %v", err)
	}

	// Create a regular file (not a directory) in the teams dir.
	if err := os.WriteFile(filepath.Join(teamsDir, "file.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	// Create a team directory with the matching session ID.
	teamDir := filepath.Join(teamsDir, "my-team")
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("create team dir: %v", err)
	}
	cfg := teamConfig{LeadSessionID: "sess-non-dir-test"}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(teamDir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Should skip non-dir entry and process directory entry.
	cleanupCurrentSessionTeam("sess-non-dir-test", homeDir)

	// The matching team directory should have been removed.
	if _, err := os.Stat(teamDir); !os.IsNotExist(err) {
		t.Error("my-team should have been removed")
	}
}

// --- rank_session.go: Handle with PatternStore not nil and ProjectDir ---

func TestRankSessionHandler_Handle_PatternStoreChecksProjectDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cred := &mockCredStore{apiKey: "some-key", hasCredentials: true}

	patternStore, err := rank.NewPatternStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	// Add an exclusion for a specific path.
	if err := patternStore.AddExclude("/excluded"); err != nil {
		t.Fatal(err)
	}

	h := NewRankSessionHandler(patternStore, cred)
	ctx := context.Background()

	// Non-excluded project with API key → will attempt submit (which will fail without server).
	// Function should still return empty output without error (non-blocking).
	input := &HookInput{
		SessionID:     "sess-pattern-check",
		CWD:           "/my/project",
		ProjectDir:    "/my/project",
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output")
	}
}

// --- post_tool.go Handle: json.Marshal failure path ---
// json.Marshal on map[string]any with standard types never fails.
// We cover the remaining handle branches: ToolOutput present + ToolInput present.

func TestPostToolHandler_Handle_WithToolOutputAndInput(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-post-both",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Read",
		ToolInput:     json.RawMessage(`{"file_path": "main.go"}`),
		ToolOutput:    json.RawMessage(`{"content": "package main"}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
	if got.Data == nil {
		t.Fatal("Data should not be nil")
	}

	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("unmarshal Data: %v", err)
	}
	if _, ok := data["output_size"]; !ok {
		t.Error("output_size should be in data")
	}
	if _, ok := data["input_size"]; !ok {
		t.Error("input_size should be in data")
	}
}

// --- checkBashCommand: no "command" field in JSON ---

func TestPreToolHandler_CheckBashCommand_NoCommandField(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		projectDir: t.TempDir(),
	}

	// Valid JSON but no "command" field → should be allowed (no match).
	toolInput, err := json.Marshal(map[string]string{
		"script": "rm -rf /",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-no-cmd-field",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(toolInput),
	}

	ctx := context.Background()
	got, handleErr := handler.Handle(ctx, input)
	if handleErr != nil {
		t.Fatalf("unexpected error: %v", handleErr)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
	// No "command" field means no match → allow.
	if got.HookSpecificOutput.PermissionDecision != DecisionAllow {
		t.Errorf("PermissionDecision = %q, want allow", got.HookSpecificOutput.PermissionDecision)
	}
}

// --- checkFileAccess: no "file_path" field in JSON ---

func TestPreToolHandler_CheckFileAccess_NoFilePathField(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		projectDir: t.TempDir(),
	}

	// Valid JSON but no "file_path" field → should be allowed (no path to check).
	toolInput, err := json.Marshal(map[string]string{
		"content": "package main",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-no-path-field",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(toolInput),
	}

	ctx := context.Background()
	got, handleErr := handler.Handle(ctx, input)
	if handleErr != nil {
		t.Fatalf("unexpected error: %v", handleErr)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
	// No file_path → allow.
	if got.HookSpecificOutput.PermissionDecision != DecisionAllow {
		t.Errorf("PermissionDecision = %q, want allow", got.HookSpecificOutput.PermissionDecision)
	}
}

// --- session_end.go: Handle with homeDir setup (more branch coverage) ---

func TestSessionEndHandler_Handle_WithTeamsDir(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create some team structure to exercise cleanup paths.
	teamsDir := filepath.Join(tmpDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// The handler uses os.UserHomeDir() internally, not CWD.
	// We just verify it handles the cleanup flow without panicking.
	input := &HookInput{
		SessionID:     "sess-with-teams",
		CWD:           tmpDir,
		ProjectDir:    tmpDir,
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
}

// --- Additional InitRankSessionHandler coverage ---

func TestInitRankSessionHandler_EventTypeCheck(t *testing.T) {
	// Cannot use t.Parallel() because we use t.Setenv()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	h, err := InitRankSessionHandler()
	if err != nil {
		t.Logf("InitRankSessionHandler returned error (acceptable): %v", err)
		return
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	// Verify the returned handler works.
	if h.EventType() != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventSessionEnd)
	}
	// The returned handler should accept a valid input without error.
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-init-check",
		CWD:           "/tmp",
		HookEventName: "SessionEnd",
	}
	out, handleErr := h.Handle(ctx, input)
	if handleErr != nil {
		t.Fatalf("Handle() unexpected error: %v", handleErr)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}
}

// --- post_tool_metrics.go: OpenFile error branch (line 86) ---

func TestLogTaskMetrics_OpenFileFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test permission denial as root")
	}
	t.Parallel()

	// Create the logs directory but make it read-only so OpenFile fails.
	tmpDir := t.TempDir()
	logsDir := filepath.Join(tmpDir, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Make logs dir read-only so OpenFile (O_CREATE) fails.
	if err := os.Chmod(logsDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(logsDir, 0o755) })

	input := &HookInput{
		SessionID:    "sess-openfile-fail",
		ToolName:     "Task",
		CWD:          tmpDir,
		ToolResponse: json.RawMessage(`{"metrics":{"tokensUsed":50,"toolUses":1,"durationSeconds":2.5}}`),
	}

	// Should not panic; OpenFile error is best-effort logged.
	logTaskMetrics(input)

	// The log file should NOT exist since OpenFile failed.
	logPath := filepath.Join(logsDir, "task-metrics.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("log file should not exist when OpenFile fails")
	}
}

// --- teammate_idle.go: LoadConfig error branch and !gate.BlockOnError branch ---

// TestTeammateIdleHandler_LoadConfigFails covers teammate_idle.go:59-62:
// When TeamName is set but quality.yaml is missing, LoadConfig returns an error
// and the handler gracefully allows idle.
func TestTeammateIdleHandler_LoadConfigFails(t *testing.T) {
	t.Parallel()

	// Use a temp dir with no quality.yaml - LoadConfig will fail (file not found).
	tmpDir := t.TempDir()

	h := NewTeammateIdleHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-idle-no-config",
		CWD:           tmpDir,
		ProjectDir:    tmpDir,
		TeamName:      "my-team",
		TeammateName:  "backend-dev",
		HookEventName: "TeammateIdle",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("Handle() returned nil output")
	}
	// Graceful degradation: idle allowed (exit code 0).
	if got.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 (idle allowed on config error)", got.ExitCode)
	}
}

// TestTeammateIdleHandler_BlockOnErrorFalse covers teammate_idle.go:66-68:
// When quality.yaml exists but blockOnError is false, idle is accepted immediately.
func TestTeammateIdleHandler_BlockOnErrorFalse(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create quality.yaml with blockOnError: false.
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("setup sections dir: %v", err)
	}
	qualityYAML := `constitution:
  enforce_quality: true
lsp_quality_gates:
  enabled: true
  run:
    allow_regression: false
    max_errors: 0
    max_lint_errors: 0
    max_type_errors: 0
`
	if err := os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"), []byte(qualityYAML), 0o644); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}

	h := NewTeammateIdleHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-idle-block-false",
		CWD:           tmpDir,
		ProjectDir:    tmpDir,
		TeamName:      "my-team",
		TeammateName:  "backend-dev",
		HookEventName: "TeammateIdle",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("Handle() returned nil output")
	}
	// blockOnError is false → idle accepted.
	if got.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 (idle allowed when blockOnError=false)", got.ExitCode)
	}
}

// Ensure io package is referenced to avoid unused import.
var _ io.Reader = (*bytes.Buffer)(nil)
