package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// crosssession_test.go — the cross-session posture panel (t60): the
// EmptySubmits form semantics (the "off" direction of the toggle), the panel
// render, and the 4-locale i18n key coverage.

// postSchemaForm은 주어진 form 값으로 /save 요청을 만든다 (기존 widget_policy_test
// 패턴; integration_test.go의 postForm과는 별개 헬퍼).
func postSchemaForm(form url.Values) *http.Request {
	req := httptest.NewRequest("POST", "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// TestParseSchemaFormEmptySubmits는 EmptySubmits 옵트인의 폼 의미론을 검증한다:
// 옵트인 필드(crosssession.inbound)의 "" 제출은 중립 되돌림 edit로 수집되고,
// 비옵트인 select(report.format)의 ""는 여전히 empty=preserve로 버려진다.
func TestParseSchemaFormEmptySubmits(t *testing.T) {
	t.Run("EmptySubmits field: empty submission becomes an edit", func(t *testing.T) {
		edits, errs := parseSchemaForm(postSchemaForm(url.Values{
			"crosssession.inbound": {""},
		}), nil)
		if len(errs) != 0 {
			t.Fatalf("parseSchemaForm errors: %v", errs)
		}
		if got, ok := edits["crosssession.inbound"]; !ok || got != "" {
			t.Errorf("edits[crosssession.inbound] = (%q, %t), want (\"\", true) — the empty submission must revert the key", got, ok)
		}
	})
	t.Run("non-opted-in select: empty submission still preserves", func(t *testing.T) {
		edits, errs := parseSchemaForm(postSchemaForm(url.Values{
			"report.format": {""},
		}), nil)
		if len(errs) != 0 {
			t.Fatalf("parseSchemaForm errors: %v", errs)
		}
		if _, ok := edits["report.format"]; ok {
			t.Errorf("edits[report.format] present — empty=preserve (EC-1) must still hold for non-opted-in fields")
		}
	})
	t.Run("EmptySubmits field: absent key preserves (non-browser post)", func(t *testing.T) {
		// A form that carries no crosssession keys at all (an agentfm-only
		// submission, a test client) must not fabricate revert edits.
		edits, errs := parseSchemaForm(postSchemaForm(url.Values{
			"agentfm.dev-a.effort": {"high"},
		}), nil)
		if len(errs) != 0 {
			t.Fatalf("parseSchemaForm errors: %v", errs)
		}
		if _, ok := edits["crosssession.inbound"]; ok {
			t.Errorf("edits[crosssession.inbound] fabricated from an absent key — unsubmitted means preserve")
		}
	})
	t.Run("EmptySubmits field: out-of-set value rejected", func(t *testing.T) {
		edits, errs := parseSchemaForm(postSchemaForm(url.Values{
			"crosssession.inbound": {"maybe"},
		}), nil)
		if _, ok := errs["crosssession.inbound"]; !ok {
			t.Errorf("errs = %v, want a crosssession.inbound entry (out-of-set rejected)", errs)
		}
		if _, ok := edits["crosssession.inbound"]; ok {
			t.Errorf("out-of-set value must not become an edit")
		}
	})
}

// TestCrossSessionPanelRendered는 crosssession 탭과 3필드가 콘솔 페이지에
// 렌더되는지 확인한다 (탭 nav + 폼 위젯).
func TestCrossSessionPanelRendered(t *testing.T) {
	body := renderConsolePage(t)
	for _, probe := range []string{
		`data-panel="crosssession"`,
		`name="crosssession.inbound"`,
		`name="crosssession.isolate_machines"`,
		`name="crosssession.isolate_machines__present"`,
		`name="crosssession.dialog_expiry"`,
	} {
		if !strings.Contains(body, probe) {
			t.Errorf("rendered console missing %q", probe)
		}
	}
	// The empty option carries its i18n key — the revert affordance is visible.
	if !strings.Contains(body, `data-i18n="f.crosssession.inbound.opt.empty"`) {
		t.Errorf("inbound select does not render the labeled empty option")
	}
}

// TestCrossSessionI18nKeysInAllLocales는 t60 신규 키가 4로케일 전부에 존재함을
// 검증한다 (t45 미번역 재발 방지 — 카드 지시).
func TestCrossSessionI18nKeysInAllLocales(t *testing.T) {
	keys := []string{
		"sec.crosssession.title", "sec.crosssession.desc", "sec.crosssession.note",
		"f.crosssession.inbound.title", "f.crosssession.inbound.desc",
		"f.crosssession.inbound.opt.accept", "f.crosssession.inbound.opt.hold",
		"f.crosssession.inbound.opt.refuse", "f.crosssession.inbound.opt.empty",
		"f.crosssession.isolate_machines.title", "f.crosssession.isolate_machines.desc",
		"f.crosssession.dialog_expiry.title", "f.crosssession.dialog_expiry.desc",
		"f.crosssession.dialog_expiry.opt.60s", "f.crosssession.dialog_expiry.opt.5m",
		"f.crosssession.dialog_expiry.opt.10m", "f.crosssession.dialog_expiry.opt.never",
		"f.crosssession.dialog_expiry.opt.empty",
	}
	for _, k := range keys {
		if !i18nKeyInAllLocales(t, k) {
			t.Errorf("i18n.js missing crosssession key %q in all 4 locales", k)
		}
	}
}

// TestCrossSessionFieldsWired는 settings 스키마가 3필드를 crosssession 섹션으로
// 노출하는지 확인한다 (web ↔ settings 배선 누락 방지).
func TestCrossSessionFieldsWired(t *testing.T) {
	if n := len(settings.SectionFields(settings.SectionCrossSession)); n != 3 {
		t.Errorf("SectionFields(SectionCrossSession) = %d fields, want 3", n)
	}
}
