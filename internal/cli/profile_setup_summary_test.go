package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// makeTestPrefs는 printProfileSummary 테스트용 완전히 채워진 ProfilePreferences를 생성한다.
func makeTestPrefs() profile.ProfilePreferences {
	return profile.ProfilePreferences{
		UserName:         "Alice",
		ConversationLang: "en",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "en",
		Model:            "opus",
		EffortLevel:      "high",
		PermissionMode:   "auto",
		StatuslineMode:   "default",
		StatuslineTheme:  "catppuccin-mocha",
	}
}

// TestPrintProfileSummary_Synced 동기화 성공 시 7개 설정 줄 +
// "Synced to project config:" 블록과 상대 경로 출력을 검증한다.
func TestPrintProfileSummary_Synced(t *testing.T) {
	var buf bytes.Buffer
	prefs := makeTestPrefs()
	txt := getProfileText("en")

	printProfileSummary(&buf, &txt, &prefs, "/some/project/root")

	output := buf.String()

	// SummaryHeader 확인
	if !strings.Contains(output, "Captured values:") {
		t.Errorf("expected SummaryHeader 'Captured values:', got:\n%s", output)
	}
	// 7개 설정 필드 확인
	for _, want := range []string{
		"User name",
		"Languages",
		"Model",
		"Effort level",
		"Permission mode",
		"Statusline mode",
		"Statusline theme",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected field label %q in output:\n%s", want, output)
		}
	}
	// 실제 값 확인
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected UserName 'Alice' in output:\n%s", output)
	}
	if !strings.Contains(output, "opus") {
		t.Errorf("expected model 'opus' in output:\n%s", output)
	}
	// 동기화 블록 확인
	if !strings.Contains(output, "Synced to project config:") {
		t.Errorf("expected SummarySyncedHeader in output:\n%s", output)
	}
	// S-1: 상대 경로 확인
	if !strings.Contains(output, ".moai/config/sections/statusline.yaml") {
		t.Errorf("expected relative statusline.yaml path in output:\n%s", output)
	}
	if !strings.Contains(output, ".moai/config/sections/language.yaml") {
		t.Errorf("expected relative language.yaml path in output:\n%s", output)
	}
	// 절대 경로가 포함되지 않는지 확인
	if strings.Contains(output, "/some/project/root") {
		t.Errorf("absolute project root should not appear in output:\n%s", output)
	}
}

// TestPrintProfileSummary_Skipped 동기화 미수행 시 중립적인 "No project-level sync" 메시지를 검증한다.
func TestPrintProfileSummary_Skipped(t *testing.T) {
	var buf bytes.Buffer
	prefs := makeTestPrefs()
	txt := getProfileText("en")

	printProfileSummary(&buf, &txt, &prefs, "")

	output := buf.String()
	// W-5: 중립적 표현 확인
	if !strings.Contains(output, "No project-level sync") {
		t.Errorf("expected neutral skip message, got:\n%s", output)
	}
	// 이전 오류성 표현이 없는지 확인
	if strings.Contains(output, "Sync skipped") {
		t.Errorf("old 'Sync skipped' message should not appear, got:\n%s", output)
	}
	// 동기화 블록이 없는지 확인
	if strings.Contains(output, "Synced to project config:") {
		t.Errorf("synced header should not appear when syncedProjectRoot is empty:\n%s", output)
	}
}

// TestPrintProfileSummary_EmptyFields 빈 UserName과 언어가 "-"로 렌더링되는지 확인한다.
func TestPrintProfileSummary_EmptyFields(t *testing.T) {
	var buf bytes.Buffer
	prefs := profile.ProfilePreferences{
		// UserName 비움
		ConversationLang: "en",
		// 나머지 언어 필드 비움
		StatuslineMode:  "default",
		StatuslineTheme: "catppuccin-mocha",
	}
	txt := getProfileText("en")

	printProfileSummary(&buf, &txt, &prefs, "")

	output := buf.String()
	// 빈 UserName은 "-"로 표시
	if !strings.Contains(output, "User name: -") {
		t.Errorf("empty UserName should render as '-', got:\n%s", output)
	}
	// 빈 언어 필드도 "-"로 표시
	if !strings.Contains(output, "-") {
		t.Errorf("expected dash for empty fields in output:\n%s", output)
	}
	// 빈 Model은 "(runtime default)"로 표시
	if !strings.Contains(output, "(runtime default)") {
		t.Errorf("empty Model should show runtime default, got:\n%s", output)
	}
}

// TestPrintProfileSummary_AllLanguages en/ko/ja/zh 각 언어에서 SummaryHeader가
// 출력에 포함되는지 확인한다.
func TestPrintProfileSummary_AllLanguages(t *testing.T) {
	prefs := makeTestPrefs()
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		t.Run(lang, func(t *testing.T) {
			var buf bytes.Buffer
			txt := getProfileText(lang)
			printProfileSummary(&buf, &txt, &prefs, "")
			output := buf.String()
			if !strings.Contains(output, txt.SummaryHeader) {
				t.Errorf("lang=%q: SummaryHeader %q not found in output:\n%s", lang, txt.SummaryHeader, output)
			}
		})
	}
}
