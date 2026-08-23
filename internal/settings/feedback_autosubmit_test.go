package settings

import "testing"

// feedback_autosubmit_test.go — SPEC-FEEDBACK-AUTO-SUBMIT-001 M7 (AC-F-023, web
// half): the feedback section is reopened as a seam-writable console section and
// carries the auto_submit consent toggle alongside the existing repository field.
//
// This reverses the SPEC-WEBCONF-SIMPLIFY-001 M3 reclassification for the
// feedback section only — the other six former seam sections stay RouteExcluded.

// TestFeedbackAutoSubmitFieldRegistered는 auto_submit 필드가 feedback 섹션의
// 편집 필드로 등록되고 seam 경로(feedback.yaml)로 영속화됨을 검증한다.
func TestFeedbackAutoSubmitFieldRegistered(t *testing.T) {
	t.Parallel()

	fields := SectionFields(SectionFeedback)
	var got *FieldDef
	for i := range fields {
		if fields[i].Name == "feedback.auto_submit" {
			got = &fields[i]
			break
		}
	}
	if got == nil {
		t.Fatal("SectionFields(SectionFeedback) has no feedback.auto_submit field")
	}
	if got.Type != TypeBool {
		t.Errorf("feedback.auto_submit type = %v, want TypeBool (consent toggle)", got.Type)
	}
	if got.Persist.Kind != PersistSeam {
		t.Errorf("feedback.auto_submit persist kind = %v, want PersistSeam", got.Persist.Kind)
	}
	if got.Persist.Section != "feedback" {
		t.Errorf("feedback.auto_submit persist section = %q, want %q", got.Persist.Section, "feedback")
	}
	if len(got.Persist.Path) != 2 || got.Persist.Path[0] != "feedback" || got.Persist.Path[1] != "auto_submit" {
		t.Errorf("feedback.auto_submit persist path = %v, want [feedback auto_submit]", got.Persist.Path)
	}
	if got.I18nKey != "f.feedback.auto_submit" {
		t.Errorf("feedback.auto_submit i18n key = %q, want %q", got.I18nKey, "f.feedback.auto_submit")
	}
}

// TestFeedbackSectionSeamWritable은 재개방된 feedback 섹션이 실제로 seam 쓰기를
// 받아들이고 값이 디스크에 남는지 확인한다 — 라우트만 바꾸고 쓰기 경로가 여전히
// 거부하면 콘솔 토글이 조용히 아무것도 하지 않는다.
func TestFeedbackSectionSeamWritable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedSectionFixture(t, root, "feedback")

	if err := ApplySchemaEdits(root, map[string]string{
		"feedback.auto_submit": "true",
	}); err != nil {
		t.Fatalf("ApplySchemaEdits(feedback.auto_submit): %v", err)
	}

	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues: %v", err)
	}
	if got := values["feedback.auto_submit"]; got != "true" {
		t.Errorf("feedback.auto_submit round-trip = %q, want %q", got, "true")
	}
}
