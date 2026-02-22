package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/rank"
)

// --- anonymizePath tests ---

func TestAnonymizePath_EmptyString(t *testing.T) {
	t.Parallel()

	got := anonymizePath("")
	if got != "" {
		t.Errorf("anonymizePath(\"\") = %q, want empty string", got)
	}
}

func TestAnonymizePath_ConsistentHash(t *testing.T) {
	t.Parallel()

	path := "/Users/test/project"
	h1 := anonymizePath(path)
	h2 := anonymizePath(path)
	if h1 != h2 {
		t.Errorf("anonymizePath not consistent: %q != %q", h1, h2)
	}
}

func TestAnonymizePath_TruncatedTo16(t *testing.T) {
	t.Parallel()

	got := anonymizePath("/some/path")
	if len(got) != 16 {
		t.Errorf("expected length 16, got %d (hash: %q)", len(got), got)
	}
}

func TestAnonymizePath_DifferentPaths(t *testing.T) {
	t.Parallel()

	h1 := anonymizePath("/project/a")
	h2 := anonymizePath("/project/b")
	if h1 == h2 {
		t.Errorf("different paths should produce different hashes: %q == %q", h1, h2)
	}
}

func TestAnonymizePath_HexEncoded(t *testing.T) {
	t.Parallel()

	got := anonymizePath("/some/path")
	for _, c := range got {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("hash contains non-hex character: %c in %q", c, got)
		}
	}
}

// --- EnvRankEnabled tests ---

func TestEnvRankEnabled_Default(t *testing.T) {
	t.Setenv("MOAI_RANK_ENABLED", "")
	if !EnvRankEnabled() {
		t.Error("expected true when env is empty (default enabled)")
	}
}

func TestEnvRankEnabled_ExplicitTrue(t *testing.T) {
	t.Setenv("MOAI_RANK_ENABLED", "true")
	if !EnvRankEnabled() {
		t.Error("expected true when MOAI_RANK_ENABLED=true")
	}
}

func TestEnvRankEnabled_ExplicitFalse(t *testing.T) {
	t.Setenv("MOAI_RANK_ENABLED", "false")
	if EnvRankEnabled() {
		t.Error("expected false when MOAI_RANK_ENABLED=false")
	}
}

func TestEnvRankEnabled_One(t *testing.T) {
	t.Setenv("MOAI_RANK_ENABLED", "1")
	if !EnvRankEnabled() {
		t.Error("expected true when MOAI_RANK_ENABLED=1")
	}
}

func TestEnvRankEnabled_Zero(t *testing.T) {
	t.Setenv("MOAI_RANK_ENABLED", "0")
	if EnvRankEnabled() {
		t.Error("expected false when MOAI_RANK_ENABLED=0")
	}
}

func TestEnvRankEnabled_InvalidValue(t *testing.T) {
	t.Setenv("MOAI_RANK_ENABLED", "not-a-bool")
	if !EnvRankEnabled() {
		t.Error("expected true (default) when env value is invalid")
	}
}

// --- EnvRankTimeout tests ---

func TestEnvRankTimeout_Default(t *testing.T) {
	t.Setenv("MOAI_RANK_TIMEOUT", "")
	got := EnvRankTimeout()
	if got != 10*time.Second {
		t.Errorf("expected 10s default, got %v", got)
	}
}

func TestEnvRankTimeout_CustomValue(t *testing.T) {
	t.Setenv("MOAI_RANK_TIMEOUT", "30")
	got := EnvRankTimeout()
	if got != 30*time.Second {
		t.Errorf("expected 30s, got %v", got)
	}
}

func TestEnvRankTimeout_InvalidValue(t *testing.T) {
	t.Setenv("MOAI_RANK_TIMEOUT", "not-a-number")
	got := EnvRankTimeout()
	if got != 10*time.Second {
		t.Errorf("expected 10s default for invalid value, got %v", got)
	}
}

func TestEnvRankTimeout_OneSecond(t *testing.T) {
	t.Setenv("MOAI_RANK_TIMEOUT", "1")
	got := EnvRankTimeout()
	if got != 1*time.Second {
		t.Errorf("expected 1s, got %v", got)
	}
}

// --- NewRankSessionHandler tests ---

func TestNewRankSessionHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewRankSessionHandler(nil, nil)
	if h.EventType() != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventSessionEnd)
	}
}

// --- mock CredentialStore ---

type mockCredStore struct {
	apiKey         string
	apiKeyErr      error
	hasCredentials bool
}

func (m *mockCredStore) Save(_ *rank.Credentials) error   { return nil }
func (m *mockCredStore) Load() (*rank.Credentials, error)  { return nil, nil }
func (m *mockCredStore) Delete() error                     { return nil }
func (m *mockCredStore) HasCredentials() bool              { return m.hasCredentials }
func (m *mockCredStore) GetAPIKey() (string, error)        { return m.apiKey, m.apiKeyErr }

// --- InitRankSessionHandlerWithStores tests ---

func TestInitRankSessionHandlerWithStores(t *testing.T) {
	t.Parallel()

	h := InitRankSessionHandlerWithStores(nil, nil)
	if h == nil {
		t.Fatal("InitRankSessionHandlerWithStores returned nil")
	}
	if h.EventType() != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventSessionEnd)
	}
}

// --- Handle tests ---

func TestRankSessionHandler_Handle_NoCredentials(t *testing.T) {
	t.Parallel()

	cred := &mockCredStore{apiKey: "", hasCredentials: false}
	h := NewRankSessionHandler(nil, cred)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-no-creds",
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
	// Should return empty output when no credentials
	if got.Decision != "" {
		t.Errorf("Decision should be empty, got %q", got.Decision)
	}
}

func TestRankSessionHandler_Handle_ExcludedProject(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cred := &mockCredStore{apiKey: "test-key", hasCredentials: true}

	patternStore, err := rank.NewPatternStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := patternStore.AddExclude("/excluded/project"); err != nil {
		t.Fatal(err)
	}

	h := NewRankSessionHandler(patternStore, cred)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-excluded",
		CWD:           "/excluded/project",
		ProjectDir:    "/excluded/project",
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

func TestRankSessionHandler_Handle_ProjectDirFallbackToCWD(t *testing.T) {
	t.Parallel()

	cred := &mockCredStore{apiKey: "", hasCredentials: false}
	h := NewRankSessionHandler(nil, cred)
	ctx := context.Background()

	// ProjectDir empty, should fall back to CWD
	input := &HookInput{
		SessionID:     "sess-fallback",
		CWD:           "/my/project",
		ProjectDir:    "",
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

func TestRankSessionHandler_Handle_NilPatternStore(t *testing.T) {
	t.Parallel()

	cred := &mockCredStore{apiKey: "", hasCredentials: false}
	h := NewRankSessionHandler(nil, cred) // nil patternStore
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-nil-patterns",
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

// --- buildSessionSubmission tests ---

func TestBuildSessionSubmission_BasicFields(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tmpDir := t.TempDir()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cred := &mockCredStore{apiKey: "test-key", hasCredentials: true}
	handler := &rankSessionHandler{
		patternStore: nil,
		credStore:    cred,
	}

	input := &HookInput{
		SessionID:     "sess-build-sub",
		CWD:           tmpDir,
		ProjectDir:    tmpDir,
		HookEventName: "SessionEnd",
		Model:         "claude-sonnet-4",
	}

	submission, err := handler.buildSessionSubmission(input)
	if err != nil {
		t.Fatalf("buildSessionSubmission() error: %v", err)
	}
	if submission == nil {
		t.Fatal("expected non-nil submission")
	}
	if submission.EndedAt == "" {
		t.Error("EndedAt should not be empty")
	}
	if submission.SessionHash == "" {
		t.Error("SessionHash should not be empty")
	}
	if submission.AnonymousProjectID == "" {
		t.Error("AnonymousProjectID should not be empty")
	}
	if submission.DeviceID == "" {
		t.Error("DeviceID should not be empty")
	}
	if submission.ModelName != "claude-sonnet-4" {
		t.Errorf("ModelName = %q, want %q", submission.ModelName, "claude-sonnet-4")
	}
}

func TestBuildSessionSubmission_WithTranscript(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	// Create a transcript file for the session
	sessionID := "sess-with-transcript"
	claudeDir := filepath.Join(tempHome, ".claude", "projects", "test-proj")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	transcriptContent := `{"timestamp":"2026-01-01T10:00:00Z","type":"user","message":{}}
{"timestamp":"2026-01-01T10:00:10Z","type":"assistant","model":"claude-opus-4","message":{"usage":{"input_tokens":500,"output_tokens":200,"cache_creation_input_tokens":50,"cache_read_input_tokens":25}}}
`
	transcriptPath := filepath.Join(claudeDir, sessionID+".jsonl")
	if err := os.WriteFile(transcriptPath, []byte(transcriptContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cred := &mockCredStore{apiKey: "test-key", hasCredentials: true}
	handler := &rankSessionHandler{
		patternStore: nil,
		credStore:    cred,
	}

	input := &HookInput{
		SessionID:     sessionID,
		CWD:           "/tmp",
		ProjectDir:    "/tmp",
		HookEventName: "SessionEnd",
	}

	submission, err := handler.buildSessionSubmission(input)
	if err != nil {
		t.Fatalf("buildSessionSubmission() error: %v", err)
	}
	if submission.InputTokens != 500 {
		t.Errorf("InputTokens = %d, want 500", submission.InputTokens)
	}
	if submission.OutputTokens != 200 {
		t.Errorf("OutputTokens = %d, want 200", submission.OutputTokens)
	}
	if submission.ModelName != "claude-opus-4" {
		t.Errorf("ModelName = %q, want %q", submission.ModelName, "claude-opus-4")
	}
	if submission.TurnCount != 1 {
		t.Errorf("TurnCount = %d, want 1", submission.TurnCount)
	}
}

func TestBuildSessionSubmission_EmptyProjectDir(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cred := &mockCredStore{apiKey: "test-key", hasCredentials: true}
	handler := &rankSessionHandler{
		patternStore: nil,
		credStore:    cred,
	}

	input := &HookInput{
		SessionID:     "sess-empty-proj",
		CWD:           "/fallback/cwd",
		ProjectDir:    "",
		HookEventName: "SessionEnd",
	}

	submission, err := handler.buildSessionSubmission(input)
	if err != nil {
		t.Fatalf("buildSessionSubmission() error: %v", err)
	}
	// AnonymousProjectID should be based on CWD fallback, not empty
	if submission.AnonymousProjectID == "" {
		t.Error("AnonymousProjectID should not be empty when CWD is provided")
	}
}
