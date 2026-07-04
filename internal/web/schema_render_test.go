package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// renderConsolePage renders GET / and returns the HTML body.
func renderConsolePage(t *testing.T) string {
	t.Helper()
	a := newTestApp(t)
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

// schemaFieldToWebControlName maps a schema field name to the HTML control name=
// attribute the web console renders for it. Statusline segments render as
// seg_<key>; the theme renders as statusline_theme; nested project fields render
// as their dot-path names; the two flat scalars render as development_mode /
// git_convention.
func schemaFieldToWebControlName(fieldName string) string {
	switch {
	case strings.HasPrefix(fieldName, "statusline_segment."):
		return "seg_" + strings.TrimPrefix(fieldName, "statusline_segment.")
	default:
		return fieldName
	}
}

// TestWebRendersSchemaFieldSet covers AC-WC10-010 (web side): every schema field
// name must have a matching name= control attribute in the rendered page.
// SPEC-WEB-CONSOLE-011 M2b: 총계 하드코딩(구 34 pin)은 B11 파생 카운트 원칙에
// 따라 제거 — 확장 포함 전 필드가 렌더됨을 집합 기준으로 검증한다 (기존
// 34-필드 잔존은 settings 쪽 TestSchemaFieldNameSet이 별도 고정).
func TestWebRendersSchemaFieldSet(t *testing.T) {
	body := renderConsolePage(t)

	names := settings.FieldNames()
	if len(names) == 0 {
		t.Fatal("schema field-name set is empty")
	}

	for _, f := range names {
		control := schemaFieldToWebControlName(f)
		// The control appears as name="<control>" (selects, inputs) or
		// name="seg_<key>" (segment checkboxes).
		marker := `name="` + control + `"`
		if !strings.Contains(body, marker) {
			t.Errorf("rendered page missing control for schema field %q (expected %s)", f, marker)
		}
	}
}

// TestWebStatuslineNoPresetControl covers AC-WC10-011b (rendered side): the
// rendered page carries NO live preset form control / option (the retired selector
// must not reappear).
func TestWebStatuslineNoPresetControl(t *testing.T) {
	body := renderConsolePage(t)
	if strings.Contains(body, `name="statusline_preset"`) || strings.Contains(body, `name="preset"`) {
		t.Error("rendered page must NOT contain a preset form control (REQ-WC10-010)")
	}
	if strings.Contains(body, `id="statusline_preset"`) {
		t.Error("rendered page must NOT contain a preset element id (REQ-WC10-010)")
	}
}

// TestWebStatuslineRendersThemeAnd15Segments covers AC-WC10-011a (rendered side):
// the re-added Statusline section renders the theme select + 15 segment toggles.
func TestWebStatuslineRendersThemeAnd15Segments(t *testing.T) {
	body := renderConsolePage(t)

	if !strings.Contains(body, `name="statusline_theme"`) {
		t.Error("rendered page missing the statusline_theme control")
	}
	segs := settings.StatuslineSegmentKeys()
	if len(segs) != 15 {
		t.Fatalf("schema segment keys = %d, want 15", len(segs))
	}
	for _, seg := range segs {
		if !strings.Contains(body, `name="seg_`+seg+`"`) {
			t.Errorf("rendered page missing segment toggle for %q", seg)
		}
		// Each segment also has its hidden __present companion.
		if !strings.Contains(body, `name="seg_`+seg+`__present"`) {
			t.Errorf("rendered page missing __present companion for segment %q", seg)
		}
	}
}
