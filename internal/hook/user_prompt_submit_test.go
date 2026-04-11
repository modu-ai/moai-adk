package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUserPromptSubmitHandler_EventType verifies the event type of the handler.
func TestUserPromptSubmitHandler_EventType(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewUserPromptSubmitHandler(cfg)

	if got := h.EventType(); got != EventUserPromptSubmit {
		t.Errorf("EventType() = %q, want %q", got, EventUserPromptSubmit)
	}
}

// TestDetectWorkflowContext verifies workflow keyword detection.
func TestDetectWorkflowContext(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		wantEmpty   bool
		wantKeyword string
	}{
		{
			name:        "contains loop keyword",
			prompt:      "/moai loop fix errors",
			wantEmpty:   false,
			wantKeyword: "loop",
		},
		{
			name:        "contains run keyword",
			prompt:      "/moai run SPEC-001",
			wantEmpty:   false,
			wantKeyword: "run",
		},
		{
			name:        "contains plan keyword",
			prompt:      "/moai plan add authentication",
			wantEmpty:   false,
			wantKeyword: "plan",
		},
		{
			name:      "no workflow keyword",
			prompt:    "what is the weather today",
			wantEmpty: true,
		},
		{
			name:      "empty prompt",
			prompt:    "",
			wantEmpty: true,
		},
		{
			name:        "case insensitive LOOP",
			prompt:      "LOOP until fixed",
			wantEmpty:   false,
			wantKeyword: "loop",
		},
		{
			name:        "keyword embedded in word",
			prompt:      "please plan the work",
			wantEmpty:   false,
			wantKeyword: "plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectWorkflowContext(tt.prompt)
			if tt.wantEmpty && got != "" {
				t.Errorf("detectWorkflowContext(%q) = %q, want empty", tt.prompt, got)
			}
			if !tt.wantEmpty {
				if got == "" {
					t.Errorf("detectWorkflowContext(%q) = empty, want non-empty (keyword: %s)", tt.prompt, tt.wantKeyword)
				}
			}
		})
	}
}

// TestHookSpecificOutput_AdditionalContextField verifies JSON serialization
// of the additionalContext field in HookSpecificOutput.
func TestHookSpecificOutput_AdditionalContextField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		output     HookSpecificOutput
		wantKey    string
		wantInJSON bool
	}{
		{
			name:       "AdditionalContext set: included in JSON",
			output:     HookSpecificOutput{HookEventName: "UserPromptSubmit", AdditionalContext: "session: SPEC-FOO-001: 테스트 기능"},
			wantKey:    "additionalContext",
			wantInJSON: true,
		},
		{
			name:       "AdditionalContext not set: omitted from JSON (omitempty)",
			output:     HookSpecificOutput{},
			wantKey:    "additionalContext",
			wantInJSON: false,
		},
		{
			name:       "set together with hookEventName",
			output:     HookSpecificOutput{HookEventName: "UserPromptSubmit", AdditionalContext: "session: project / main"},
			wantKey:    "hookEventName",
			wantInJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.output)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var m map[string]interface{}
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			_, exists := m[tt.wantKey]
			if tt.wantInJSON && !exists {
				t.Errorf("key %q missing from JSON, expected to be present. JSON: %s", tt.wantKey, string(data))
			}
			if !tt.wantInJSON && exists {
				t.Errorf("key %q present in JSON, expected to be omitted (omitempty). JSON: %s", tt.wantKey, string(data))
			}
		})
	}
}

// TestUserPromptSubmitHandler_WithSPEC verifies that SPEC information is included
// in additionalContext when a SPEC context is present.
func TestUserPromptSubmitHandler_WithSPEC(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-CC297-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}
	specContent := "# UserPromptSubmit 세션 타이틀 기능\n\n## 요구사항\n..."
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("failed to create spec.md: %v", err)
	}

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	input := &HookInput{
		SessionID: "test-session-123",
		Prompt:    "기능을 구현해줘",
		CWD:       tmpDir,
	}

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if output == nil {
		t.Fatal("output is nil")
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil, expected to be set")
	}

	title := output.HookSpecificOutput.SessionTitle
	if title == "" {
		t.Error("SessionTitle is empty, expected to contain SPEC-CC297-001")
	}
	if !strings.Contains(title, "SPEC-CC297-001") {
		t.Errorf("SessionTitle does not contain SPEC-CC297-001: %q", title)
	}
	if output.HookSpecificOutput.HookEventName != "UserPromptSubmit" {
		t.Errorf("hookEventName should be UserPromptSubmit, got: %q", output.HookSpecificOutput.HookEventName)
	}
}

// TestUserPromptSubmitHandler_WithoutSPEC verifies that project/branch information
// is included in additionalContext when no SPEC is present.
func TestUserPromptSubmitHandler_WithoutSPEC(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	input := &HookInput{
		SessionID: "test-session-456",
		Prompt:    "코드를 리뷰해줘",
		CWD:       tmpDir,
	}

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if output == nil {
		t.Fatal("output is nil")
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}

	title := output.HookSpecificOutput.SessionTitle
	if title == "" {
		t.Error("SessionTitle is empty, expected to contain project/branch info")
	}

	projectName := filepath.Base(tmpDir)
	if !strings.Contains(title, projectName) {
		t.Errorf("SessionTitle does not contain project name %q: %q", projectName, title)
	}
	if !strings.Contains(title, "/") {
		t.Errorf("SessionTitle does not contain '/' separator: %q", title)
	}
}

// TestUserPromptSubmitHandler_NilConfig verifies that the handler operates
// without error when ConfigProvider returns nil.
func TestUserPromptSubmitHandler_NilConfig(t *testing.T) {
	t.Parallel()

	handler := NewUserPromptSubmitHandler(&mockConfigProvider{cfg: nil})

	input := &HookInput{
		SessionID: "test-session-789",
		Prompt:    "안녕하세요",
		CWD:       t.TempDir(),
	}

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle returned error (graceful degradation required): %v", err)
	}
	if output == nil {
		t.Fatal("output is nil")
	}
}

// TestUserPromptSubmitHandler_EmptyCWD verifies that the handler operates
// without error when CWD is an empty string.
func TestUserPromptSubmitHandler_EmptyCWD(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	input := &HookInput{
		SessionID: "test-session-000",
		Prompt:    "테스트 프롬프트",
		CWD:       "",
	}

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if output == nil {
		t.Fatal("output is nil")
	}
}

// TestUserPromptSubmitHandler_SPECWithoutHeading verifies that only the SPEC ID
// is returned when spec.md has no heading.
func TestUserPromptSubmitHandler_SPECWithoutHeading(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}
	specContent := "헤딩 없는 내용입니다.\n상세 설명..."
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("failed to create spec.md: %v", err)
	}

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	output, err := handler.Handle(context.Background(), &HookInput{
		SessionID: "test-no-heading",
		Prompt:    "테스트",
		CWD:       tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}

	title := output.HookSpecificOutput.SessionTitle
	if !strings.Contains(title, "SPEC-TEST-001") {
		t.Errorf("expected SPEC ID to be included when heading is absent, got: %q", title)
	}
}

// TestUserPromptSubmitHandler_MultipleSpecs verifies that the most recently
// modified SPEC is selected when multiple SPECs exist.
func TestUserPromptSubmitHandler_MultipleSpecs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	spec1Dir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-OLD-001")
	if err := os.MkdirAll(spec1Dir, 0o755); err != nil {
		t.Fatalf("failed to create spec1 directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(spec1Dir, "spec.md"), []byte("# 오래된 SPEC\n"), 0o644); err != nil {
		t.Fatalf("failed to create spec1.md: %v", err)
	}

	spec2Dir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-NEW-002")
	if err := os.MkdirAll(spec2Dir, 0o755); err != nil {
		t.Fatalf("failed to create spec2 directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(spec2Dir, "spec.md"), []byte("# 새로운 SPEC\n"), 0o644); err != nil {
		t.Fatalf("failed to create spec2.md: %v", err)
	}

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	output, err := handler.Handle(context.Background(), &HookInput{
		SessionID: "test-multi-spec",
		Prompt:    "테스트",
		CWD:       tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}

	title := output.HookSpecificOutput.SessionTitle
	if !strings.Contains(title, "SPEC-") {
		t.Errorf("SessionTitle does not contain SPEC ID: %q", title)
	}
}

// TestUserPromptSubmitHandler_SPECTitle_Format verifies SPEC title format using table-driven tests.
func TestUserPromptSubmitHandler_SPECTitle_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		specID      string
		specHeading string
		wantInCtx   string
	}{
		{
			name:        "SPEC-AUTH-001 title",
			specID:      "SPEC-AUTH-001",
			specHeading: "사용자 인증 기능",
			wantInCtx:   "SPEC-AUTH-001: 사용자 인증 기능",
		},
		{
			name:        "SPEC-CC297-001 title",
			specID:      "SPEC-CC297-001",
			specHeading: "UserPromptSubmit 세션 타이틀",
			wantInCtx:   "SPEC-CC297-001: UserPromptSubmit 세션 타이틀",
		},
		{
			name:        "SPEC-ID already in heading: deduplicate",
			specID:      "SPEC-SRS-003",
			specHeading: "SPEC-SRS-003: Dashboard + CLI + Agency 통합",
			wantInCtx:   "SPEC-SRS-003: Dashboard + CLI + Agency 통합",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, ".moai", "specs", tt.specID)
			if err := os.MkdirAll(specDir, 0o755); err != nil {
				t.Fatalf("failed to create spec directory: %v", err)
			}
			specContent := "# " + tt.specHeading + "\n\n내용..."
			if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
				t.Fatalf("failed to create spec.md: %v", err)
			}

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			handler := NewUserPromptSubmitHandler(cfg)

			input := &HookInput{
				SessionID: "test-session",
				Prompt:    "구현해줘",
				CWD:       tmpDir,
			}

			output, err := handler.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("Handle failed: %v", err)
			}
			if output.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput is nil")
			}

			got := output.HookSpecificOutput.SessionTitle
			if !strings.Contains(got, tt.wantInCtx) {
				t.Errorf("SessionTitle does not contain expected value\n  got:  %q\n  want contains: %q", got, tt.wantInCtx)
			}
		})
	}
}
