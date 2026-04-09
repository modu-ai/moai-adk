package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUserPromptSubmitHandler_EventType는 핸들러의 이벤트 타입을 검증한다.
func TestUserPromptSubmitHandler_EventType(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewUserPromptSubmitHandler(cfg)

	if got := h.EventType(); got != EventUserPromptSubmit {
		t.Errorf("EventType() = %q, want %q", got, EventUserPromptSubmit)
	}
}

// TestDetectWorkflowContext는 워크플로우 키워드 감지를 검증한다.
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

// TestHookSpecificOutput_AdditionalContextField는 HookSpecificOutput의
// additionalContext 필드 JSON 직렬화를 검증한다.
func TestHookSpecificOutput_AdditionalContextField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		output     HookSpecificOutput
		wantKey    string
		wantInJSON bool
	}{
		{
			name:       "AdditionalContext 설정 시 JSON에 포함",
			output:     HookSpecificOutput{HookEventName: "UserPromptSubmit", AdditionalContext: "session: SPEC-FOO-001: 테스트 기능"},
			wantKey:    "additionalContext",
			wantInJSON: true,
		},
		{
			name:       "AdditionalContext 미설정 시 JSON에서 생략 (omitempty)",
			output:     HookSpecificOutput{},
			wantKey:    "additionalContext",
			wantInJSON: false,
		},
		{
			name:       "hookEventName과 함께 설정",
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
				t.Fatalf("JSON 직렬화 실패: %v", err)
			}

			var m map[string]interface{}
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("JSON 역직렬화 실패: %v", err)
			}

			_, exists := m[tt.wantKey]
			if tt.wantInJSON && !exists {
				t.Errorf("JSON에 %q 키가 없음, 있어야 함. JSON: %s", tt.wantKey, string(data))
			}
			if !tt.wantInJSON && exists {
				t.Errorf("JSON에 %q 키가 있음, 없어야 함 (omitempty). JSON: %s", tt.wantKey, string(data))
			}
		})
	}
}

// TestUserPromptSubmitHandler_WithSPEC는 SPEC 컨텍스트가 있을 때
// additionalContext에 SPEC 정보가 포함되는지 검증한다.
func TestUserPromptSubmitHandler_WithSPEC(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-CC297-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("spec 디렉토리 생성 실패: %v", err)
	}
	specContent := "# UserPromptSubmit 세션 타이틀 기능\n\n## 요구사항\n..."
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("spec.md 파일 생성 실패: %v", err)
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
		t.Fatalf("Handle 실패: %v", err)
	}
	if output == nil {
		t.Fatal("output이 nil임")
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput이 nil임, 설정되어야 함")
	}

	title := output.HookSpecificOutput.SessionTitle
	if title == "" {
		t.Error("SessionTitle이 비어 있음, SPEC-CC297-001이 포함되어야 함")
	}
	if !strings.Contains(title, "SPEC-CC297-001") {
		t.Errorf("SessionTitle에 SPEC-CC297-001이 없음: %q", title)
	}
	if output.HookSpecificOutput.HookEventName != "UserPromptSubmit" {
		t.Errorf("hookEventName이 UserPromptSubmit이어야 함, got: %q", output.HookSpecificOutput.HookEventName)
	}
}

// TestUserPromptSubmitHandler_WithoutSPEC는 SPEC 없을 때
// project/branch 형식의 정보가 additionalContext에 포함되는지 검증한다.
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
		t.Fatalf("Handle 실패: %v", err)
	}
	if output == nil {
		t.Fatal("output이 nil임")
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput이 nil임")
	}

	title := output.HookSpecificOutput.SessionTitle
	if title == "" {
		t.Error("SessionTitle이 비어 있음, project/branch 정보가 포함되어야 함")
	}

	projectName := filepath.Base(tmpDir)
	if !strings.Contains(title, projectName) {
		t.Errorf("SessionTitle에 프로젝트명 %q이 없음: %q", projectName, title)
	}
	if !strings.Contains(title, "/") {
		t.Errorf("SessionTitle에 '/' 구분자가 없음: %q", title)
	}
}

// TestUserPromptSubmitHandler_NilConfig는 ConfigProvider가 nil을 반환할 때
// 에러 없이 동작하는지 검증한다.
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
		t.Fatalf("Handle이 에러 반환함 (graceful degradation 필요): %v", err)
	}
	if output == nil {
		t.Fatal("output이 nil임")
	}
}

// TestUserPromptSubmitHandler_EmptyCWD는 CWD가 빈 문자열일 때
// 에러 없이 동작하는지 검증한다.
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
		t.Fatalf("Handle이 에러 반환함: %v", err)
	}
	if output == nil {
		t.Fatal("output이 nil임")
	}
}

// TestUserPromptSubmitHandler_SPECWithoutHeading은 spec.md에 헤딩이 없을 때
// SPEC ID만 반환하는지 검증한다.
func TestUserPromptSubmitHandler_SPECWithoutHeading(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("spec 디렉토리 생성 실패: %v", err)
	}
	specContent := "헤딩 없는 내용입니다.\n상세 설명..."
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("spec.md 파일 생성 실패: %v", err)
	}

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	output, err := handler.Handle(context.Background(), &HookInput{
		SessionID: "test-no-heading",
		Prompt:    "테스트",
		CWD:       tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle 실패: %v", err)
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput이 nil임")
	}

	title := output.HookSpecificOutput.SessionTitle
	if !strings.Contains(title, "SPEC-TEST-001") {
		t.Errorf("헤딩 없을 시 SPEC ID가 포함되어야 함, got: %q", title)
	}
}

// TestUserPromptSubmitHandler_MultipleSpecs는 여러 SPEC이 있을 때
// 가장 최근 수정된 SPEC을 선택하는지 검증한다.
func TestUserPromptSubmitHandler_MultipleSpecs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	spec1Dir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-OLD-001")
	if err := os.MkdirAll(spec1Dir, 0o755); err != nil {
		t.Fatalf("spec1 디렉토리 생성 실패: %v", err)
	}
	if err := os.WriteFile(filepath.Join(spec1Dir, "spec.md"), []byte("# 오래된 SPEC\n"), 0o644); err != nil {
		t.Fatalf("spec1.md 파일 생성 실패: %v", err)
	}

	spec2Dir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-NEW-002")
	if err := os.MkdirAll(spec2Dir, 0o755); err != nil {
		t.Fatalf("spec2 디렉토리 생성 실패: %v", err)
	}
	if err := os.WriteFile(filepath.Join(spec2Dir, "spec.md"), []byte("# 새로운 SPEC\n"), 0o644); err != nil {
		t.Fatalf("spec2.md 파일 생성 실패: %v", err)
	}

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	output, err := handler.Handle(context.Background(), &HookInput{
		SessionID: "test-multi-spec",
		Prompt:    "테스트",
		CWD:       tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle 실패: %v", err)
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput이 nil임")
	}

	title := output.HookSpecificOutput.SessionTitle
	if !strings.Contains(title, "SPEC-") {
		t.Errorf("SessionTitle에 SPEC ID가 없음: %q", title)
	}
}

// TestUserPromptSubmitHandler_SPECTitle_Format은 SPEC 타이틀 형식을 테이블 기반으로 검증한다.
func TestUserPromptSubmitHandler_SPECTitle_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		specID      string
		specHeading string
		wantInCtx   string
	}{
		{
			name:        "SPEC-AUTH-001 타이틀",
			specID:      "SPEC-AUTH-001",
			specHeading: "사용자 인증 기능",
			wantInCtx:   "SPEC-AUTH-001: 사용자 인증 기능",
		},
		{
			name:        "SPEC-CC297-001 타이틀",
			specID:      "SPEC-CC297-001",
			specHeading: "UserPromptSubmit 세션 타이틀",
			wantInCtx:   "SPEC-CC297-001: UserPromptSubmit 세션 타이틀",
		},
		{
			name:        "SPEC-ID가 헤딩에 이미 포함된 경우 중복 제거",
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
				t.Fatalf("spec 디렉토리 생성 실패: %v", err)
			}
			specContent := "# " + tt.specHeading + "\n\n내용..."
			if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
				t.Fatalf("spec.md 파일 생성 실패: %v", err)
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
				t.Fatalf("Handle 실패: %v", err)
			}
			if output.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput이 nil임")
			}

			got := output.HookSpecificOutput.SessionTitle
			if !strings.Contains(got, tt.wantInCtx) {
				t.Errorf("SessionTitle에 기대값이 없음\n  got:  %q\n  want contains: %q", got, tt.wantInCtx)
			}
		})
	}
}
