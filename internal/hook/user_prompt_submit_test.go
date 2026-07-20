package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
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

// TestUserPromptSubmitHandler_WithoutSPEC_FirstPrompt verifies the new policy:
// with no SPEC and no transcript (the session's first prompt), the handler emits
// NO sessionTitle. This gives Claude Code's native ai-title generator a clear
// field. The previous "projectName / branchName" fallback churned an identical
// title on every prompt and shadowed the native title — that defect is removed.
func TestUserPromptSubmitHandler_WithoutSPEC_FirstPrompt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	handler := NewUserPromptSubmitHandler(cfg)

	input := &HookInput{
		SessionID: "test-session-456",
		Prompt:    "코드를 리뷰해줘",
		CWD:       tmpDir,
		// TranscriptPath intentionally empty: the session's first prompt.
	}

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if output == nil {
		t.Fatal("output is nil")
	}

	if got := sessionTitleOf(output); got != "" {
		t.Errorf("SessionTitle = %q, want empty (first prompt, no SPEC — let ai-title win)", got)
	}

	// The removed fallback must not resurface: no project name, no "/" separator.
	projectName := filepath.Base(tmpDir)
	if title := sessionTitleOf(output); strings.Contains(title, projectName) || strings.Contains(title, "/") {
		t.Errorf("SessionTitle still carries the removed project/branch fallback: %q", title)
	}
}

// sessionTitleOf safely extracts SessionTitle from a HookOutput, treating a nil
// output or nil HookSpecificOutput as an empty title.
func sessionTitleOf(out *HookOutput) string {
	if out == nil || out.HookSpecificOutput == nil {
		return ""
	}
	return out.HookSpecificOutput.SessionTitle
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

// newHookHandler builds a concrete *userPromptSubmitHandler with the given
// conversation_language (empty string keeps the default "en"). It returns the
// concrete type so unexported methods (buildSessionTitle) can be exercised.
func newHookHandler(lang string) *userPromptSubmitHandler {
	cfg := newTestConfig()
	if lang != "" {
		cfg.Language.ConversationLanguage = lang
	}
	return &userPromptSubmitHandler{cfg: &mockConfigProvider{cfg: cfg}}
}

// writeTranscript writes the given JSONL records to a transcript file inside a
// fresh temp dir and returns its absolute path.
func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}
	return path
}

// TestDeriveTitleFromPrompt verifies the pure title-derivation string function:
// slash-command stripping, directive-keyword stripping, first sentence/line
// selection, whitespace collapse, and the min-length reject rule.
func TestDeriveTitleFromPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{"plain korean prose", "로그인 기능을 구현해줘", "로그인 기능을 구현해줘"},
		{"slash command token stripped, rest kept", "/moai plan 인증 추가", "plan 인증 추가"},
		{"bare slash command with no space -> empty", "/clear", ""},
		{"bare slash namespace with no space -> empty", "/moai", ""},
		{"ultrathink. prefix stripped", "ultrathink. 대시보드를 개선해줘", "대시보드를 개선해줘"},
		{"ultrathink (no period) stripped", "ultrathink 코드 리뷰 요청", "코드 리뷰 요청"},
		{"ultracode stripped", "ultracode 성능 최적화 진행", "성능 최적화 진행"},
		{"uppercase directive stripped (case-insensitive)", "ULTRATHINK. 대문자 테스트 진행", "대문자 테스트 진행"},
		{"substring ultrathinking NOT stripped", "ultrathinking 은 무엇인가", "ultrathinking 은 무엇인가"},
		{"slash + directive + content", "/moai run ultrathink. 배포 준비 작업", "run 배포 준비 작업"},
		{"first sentence cut at period", "첫 문장입니다. 둘째 문장입니다.", "첫 문장입니다"},
		{"first sentence cut at fullwidth question mark", "이게 뭐죠？ 다음 질문", "이게 뭐죠"},
		{"first line cut at newline", "첫 줄 내용\n둘째 줄 내용", "첫 줄 내용"},
		{"internal whitespace collapsed", "여러    공백    정리해줘", "여러 공백 정리해줘"},
		{"empty prompt -> empty", "", ""},
		{"single rune -> empty (too short)", "가", ""},
		{"two runes kept", "가나", "가나"},
		{"directive-only prompt -> empty", "ultrathink.", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := deriveTitleFromPrompt(tt.prompt); got != tt.want {
				t.Errorf("deriveTitleFromPrompt(%q) = %q, want %q", tt.prompt, got, tt.want)
			}
		})
	}
}

// TestDeriveTitleFromPrompt_TruncationRuneSafe verifies that a title longer than
// titleMaxRunes is truncated at a rune boundary (no mid-character split on Korean
// text) with a single trailing ellipsis appended.
func TestDeriveTitleFromPrompt_TruncationRuneSafe(t *testing.T) {
	t.Parallel()

	// 80 Korean syllables, each a single 3-byte rune, no terminators/whitespace.
	long := strings.Repeat("가", 80)
	got := deriveTitleFromPrompt(long)

	if rc := utf8.RuneCountInString(got); rc != titleMaxRunes+1 {
		t.Fatalf("rune count = %d, want %d (60 content runes + 1 ellipsis)", rc, titleMaxRunes+1)
	}
	if !strings.HasSuffix(got, titleEllipsis) {
		t.Errorf("result %q missing trailing ellipsis %q", got, titleEllipsis)
	}
	if !utf8.ValidString(got) {
		t.Errorf("result %q is not valid UTF-8 (mid-rune split)", got)
	}
	// The content prefix (everything before the ellipsis) must be exactly 60 '가'.
	prefix := strings.TrimSuffix(got, titleEllipsis)
	if rc := utf8.RuneCountInString(prefix); rc != titleMaxRunes {
		t.Errorf("content prefix rune count = %d, want %d", rc, titleMaxRunes)
	}
	if prefix != strings.Repeat("가", titleMaxRunes) {
		t.Errorf("content prefix corrupted: %q", prefix)
	}
	// Byte-length proof of a clean boundary: 60*3 (가) + 3 (U+2026) = 183.
	if bl := len(got); bl != titleMaxRunes*3+len(titleEllipsis) {
		t.Errorf("byte length = %d, want %d", bl, titleMaxRunes*3+len(titleEllipsis))
	}
}

// TestBuildSessionTitle_NoSPEC_TranscriptPolicy verifies the transcript-driven
// title policy when no SPEC is active: ai-title / custom-title yield an empty
// title (never shadow, never churn), the first user prompt is derived otherwise,
// and missing/malformed/empty transcripts degrade to an empty title.
func TestBuildSessionTitle_NoSPEC_TranscriptPolicy(t *testing.T) {
	t.Parallel()

	const (
		userDashboard = `{"type":"user","message":{"role":"user","content":"대시보드 성능 개선"}}`
		userFirst     = `{"type":"user","message":{"role":"user","content":"첫 요청 처리해줘"}}`
		userSecond    = `{"type":"user","message":{"role":"user","content":"둘째 요청"}}`
		userLogin     = `{"type":"user","message":{"role":"user","content":"로그인 기능 구현"}}`
		userBlocks    = `{"type":"user","message":{"role":"user","content":[{"type":"text","text":"블록 콘텐츠 요청 처리"}]}}`
		assistantRec  = `{"type":"assistant","message":{"role":"assistant","content":"네, 진행하겠습니다"}}`
		aiTitleRec    = `{"type":"ai-title","aiTitle":"로그인 기능 구현","sessionId":"s1"}`
		customRec     = `{"type":"custom-title","customTitle":"이전 커스텀 타이틀","sessionId":"s1"}`
	)

	tests := []struct {
		name  string
		lines []string // nil -> use empty/absent path per usePath below
		// usePath overrides the transcript path directly (for empty/nonexistent).
		usePath string
		want    string
	}{
		{
			name:    "empty transcript path (first prompt) -> empty",
			usePath: "",
			want:    "",
		},
		{
			name:    "nonexistent transcript path -> empty",
			usePath: "/no/such/transcript-file.jsonl",
			want:    "",
		},
		{
			name:  "user prompt, no title records -> derived title",
			lines: []string{userDashboard},
			want:  "대시보드 성능 개선",
		},
		{
			name:  "first user prompt chosen, not the current/last one",
			lines: []string{userFirst, assistantRec, userSecond},
			want:  "첫 요청 처리해줘",
		},
		{
			name:  "ai-title present -> empty (never shadow native title)",
			lines: []string{userLogin, aiTitleRec},
			want:  "",
		},
		{
			name:  "custom-title present -> empty (stable, first-wins)",
			lines: []string{userLogin, customRec},
			want:  "",
		},
		{
			name:  "user content as content-block array -> derived title",
			lines: []string{userBlocks},
			want:  "블록 콘텐츠 요청 처리",
		},
		{
			name:  "malformed transcript -> empty (no panic, no error)",
			lines: []string{`this is not json at all`, `{"type":"user",`},
			want:  "",
		},
		{
			name:  "assistant-only transcript (no user, no title) -> empty",
			lines: []string{assistantRec},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.usePath
			if tt.lines != nil {
				path = writeTranscript(t, tt.lines...)
			}

			// A cwd with no SPEC ensures detectActiveSpec returns "" and the
			// transcript policy is exercised.
			cwd := t.TempDir()
			h := newHookHandler("ko")

			got := h.buildSessionTitle(context.Background(), cwd, path)
			if got != tt.want {
				t.Errorf("buildSessionTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBuildSessionTitle_SPECTakesPrecedence verifies that an active SPEC yields
// the SPEC title even when the transcript already contains an ai-title record —
// case 1 short-circuits before any transcript inspection.
func TestBuildSessionTitle_SPECTakesPrecedence(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	specDir := filepath.Join(cwd, ".moai", "specs", "SPEC-PREC-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# 우선순위 검증 기능\n"), 0o644); err != nil {
		t.Fatalf("failed to create spec.md: %v", err)
	}

	path := writeTranscript(t,
		`{"type":"user","message":{"role":"user","content":"로그인 구현"}}`,
		`{"type":"ai-title","aiTitle":"로그인 기능 구현","sessionId":"s1"}`,
	)

	h := newHookHandler("ko")
	got := h.buildSessionTitle(context.Background(), cwd, path)
	want := "SPEC-PREC-001: 우선순위 검증 기능"
	if got != want {
		t.Errorf("buildSessionTitle() = %q, want %q (SPEC must take precedence over transcript)", got, want)
	}
}

// TestBuildSessionTitle_NilConfig verifies buildSessionTitle degrades gracefully
// (no panic, empty title) when the config provider returns nil while deriving
// from a transcript.
func TestBuildSessionTitle_NilConfig(t *testing.T) {
	t.Parallel()

	path := writeTranscript(t, `{"type":"user","message":{"role":"user","content":"설정 없는 세션 테스트"}}`)
	h := &userPromptSubmitHandler{cfg: &mockConfigProvider{cfg: nil}}

	got := h.buildSessionTitle(context.Background(), t.TempDir(), path)
	if got != "설정 없는 세션 테스트" {
		t.Errorf("buildSessionTitle() with nil config = %q, want derived title", got)
	}
}
