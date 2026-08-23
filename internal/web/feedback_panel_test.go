package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// feedback_panel_test.go — SPEC-FEEDBACK-AUTO-SUBMIT-001 M7 (AC-F-023, web half):
// the feedback tab, its panel, and the auto_submit toggle reach the rendered
// console, and the new field's i18n keys exist in all 4 locales.

// TestFeedbackPanelRendered는 feedback 탭·패널과 두 필드 위젯이 콘솔 페이지에
// 실제로 렌더되는지 확인한다 (라우트 재개방이 화면에 도달했다는 관측).
func TestFeedbackPanelRendered(t *testing.T) {
	body := renderConsolePage(t)
	for _, probe := range []string{
		`data-tab="feedback"`,
		`data-panel="feedback"`,
		`name="feedback.repository"`,
		`name="feedback.auto_submit"`,
		`name="feedback.auto_submit__present"`,
	} {
		if !strings.Contains(body, probe) {
			t.Errorf("rendered console missing %q", probe)
		}
	}
}

// TestFeedbackAutoSubmitI18nKeysInAllLocales는 M7 신규 키가 4로케일 전부에
// 존재함을 검증한다 (부분 로케일은 결함 — 카드 지시).
func TestFeedbackAutoSubmitI18nKeysInAllLocales(t *testing.T) {
	for _, k := range []string{
		"f.feedback.auto_submit.title",
		"f.feedback.auto_submit.desc",
	} {
		if !i18nKeyInAllLocales(t, k) {
			t.Errorf("i18n.js missing feedback key %q in all 4 locales", k)
		}
	}
}

// TestFeedbackPanelFieldsWired는 패널 메타가 settings 스키마의 feedback 필드를
// 그대로 싣는지 확인한다 (web ↔ settings 배선 누락 방지).
func TestFeedbackPanelFieldsWired(t *testing.T) {
	meta := schemaPanelMeta("feedback")
	if meta.ID != settings.SectionFeedback {
		t.Errorf("schemaPanelMeta(feedback).ID = %q, want %q", meta.ID, settings.SectionFeedback)
	}
	want := len(settings.SectionFields(settings.SectionFeedback))
	if want < 2 {
		t.Fatalf("SectionFields(SectionFeedback) = %d fields, want >= 2 (repository + auto_submit)", want)
	}
	if got := len(meta.Fields); got != want {
		t.Errorf("feedback panel renders %d fields, want %d (all section fields)", got, want)
	}
}
