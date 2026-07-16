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
//
// M3 redesign: statusline schema fields (statusline_theme + statusline_segment.*)
// remain in the schema — consumed by the TUI / profile_setup CLI path — but are
// intentionally NOT rendered in the web console. They are skipped here; their
// schema-level preservation is covered by internal/settings tests.
func TestWebRendersSchemaFieldSet(t *testing.T) {
	body := renderConsolePage(t)

	names := settings.FieldNames()
	if len(names) == 0 {
		t.Fatal("schema field-name set is empty")
	}

	// Build the set of statusline-scoped schema field names to skip (web does not
	// render the statusline section, but the schema still carries the fields).
	statuslineFields := map[string]bool{}
	for _, f := range settings.SectionFields(settings.SectionStatusline) {
		statuslineFields[f.Name] = true
	}

	// SPEC-WEBCONF-SIMPLIFY-001 M3: skip fields belonging to removed sections
	// (their tabs/fieldsets are removed; fields are schema-preserved but
	// web-not-rendered — config keys persist in baked template YAML, REQ-WC-003).
	m3RemovedSections := map[settings.SectionID]bool{
		settings.SectionQualityExtras: true,
		settings.SectionWorkflow:      true,
		settings.SectionHarness:       true,
		settings.SectionRalph:         true,
		settings.SectionFeedback:      true,
		settings.SectionObservability: true,
		settings.SectionSecurity:      true,
		settings.SectionMx:            true,
		settings.SectionHandoff:       true,
		settings.SectionCache:         true,
		// git_strategy section removed from the web console (config stays in yaml).
		settings.SectionGitStrategy: true,
		// SPEC-DESIGN-MOAIWEBV2-001 M1: the orphan `project` panel (fieldsetProject)
		// was removed. It rendered the quality + git_convention sections
		// (development_mode / git_convention + nested); those fields are
		// schema-preserved but web-not-rendered (editable via yaml config / CLI).
		settings.SectionQuality:       true,
		settings.SectionGitConvention: true,
	}
	m3RemovedFields := map[string]bool{}
	for _, f := range settings.AllFields() {
		if m3RemovedSections[f.Section] {
			m3RemovedFields[f.Name] = true
		}
	}

	for _, f := range names {
		if statuslineFields[f] {
			continue // statusline: schema-preserved, web-not-rendered (M3 redesign)
		}
		if m3RemovedFields[f] {
			continue // M3: removed-section field — schema-preserved, web-not-rendered
		}
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

// TestWebStatuslineSectionNotRendered covers the M3 redesign (rendered side): the
// Statusline section is gone from the web console — no statusline_theme select and
// no seg_<key> segment toggles are rendered. The statusline schema fields +
// StatuslineSegmentKeys() accessor remain (consumed by the TUI path); their
// preservation is covered by internal/settings tests.
func TestWebStatuslineSectionNotRendered(t *testing.T) {
	body := renderConsolePage(t)

	if strings.Contains(body, `name="statusline_theme"`) {
		t.Error("rendered page must NOT contain a statusline_theme control (section removed)")
	}
	if strings.Contains(body, `id="statusline_theme"`) {
		t.Error("rendered page must NOT contain a statusline_theme element id (section removed)")
	}
	segs := settings.StatuslineSegmentKeys()
	if len(segs) != 16 {
		t.Fatalf("schema segment keys = %d, want 16 (accessor preserved)", len(segs))
	}
	for _, seg := range segs {
		if strings.Contains(body, `name="seg_`+seg+`"`) {
			t.Errorf("rendered page must NOT contain segment toggle for %q (section removed)", seg)
		}
		if strings.Contains(body, `name="seg_`+seg+`__present"`) {
			t.Errorf("rendered page must NOT contain __present companion for segment %q (section removed)", seg)
		}
	}
}
