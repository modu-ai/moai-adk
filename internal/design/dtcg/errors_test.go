package dtcg_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// TestValidationError_Error: ValidationError 문자열 표현 검증.
func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	t.Run("값 포함", func(t *testing.T) {
		t.Parallel()
		e := &dtcg.ValidationError{
			TokenPath: "color.primary",
			Category:  "color",
			Rule:      "잘못된 hex 형식",
			Value:     "notacolor",
		}
		s := e.Error()
		if !strings.Contains(s, "color.primary") {
			t.Errorf("Error() = %q; 토큰 경로 포함 안됨", s)
		}
		if !strings.Contains(s, "color") {
			t.Errorf("Error() = %q; 카테고리 포함 안됨", s)
		}
		if !strings.Contains(s, "notacolor") {
			t.Errorf("Error() = %q; 값 포함 안됨", s)
		}
	})

	t.Run("값 없음", func(t *testing.T) {
		t.Parallel()
		e := &dtcg.ValidationError{
			TokenPath: "dimension.base",
			Category:  "dimension",
			Rule:      "unit 누락",
		}
		s := e.Error()
		if !strings.Contains(s, "dimension.base") {
			t.Errorf("Error() = %q; 토큰 경로 포함 안됨", s)
		}
	})
}

// TestValidationWarning_Warning: ValidationWarning 문자열 표현 검증.
func TestValidationWarning_Warning(t *testing.T) {
	t.Parallel()
	w := &dtcg.ValidationWarning{
		TokenPath: "color.brand",
		Category:  "color",
		Message:   "브랜드 색상과 충돌",
	}
	s := w.Warning()
	if !strings.Contains(s, "color.brand") {
		t.Errorf("Warning() = %q; 토큰 경로 포함 안됨", s)
	}
	if !strings.Contains(s, "브랜드 색상과 충돌") {
		t.Errorf("Warning() = %q; 메시지 포함 안됨", s)
	}
}

// TestReport_HasErrors: Report.HasErrors() 검증.
func TestReport_HasErrors(t *testing.T) {
	t.Parallel()

	t.Run("오류 없음", func(t *testing.T) {
		t.Parallel()
		r := &dtcg.Report{Valid: true}
		if r.HasErrors() {
			t.Error("HasErrors() = true; 오류 없어야 함")
		}
	})

	t.Run("오류 있음", func(t *testing.T) {
		t.Parallel()
		r := &dtcg.Report{Valid: true}
		r.AddError(&dtcg.ValidationError{TokenPath: "test", Category: "color", Rule: "오류"})
		if !r.HasErrors() {
			t.Error("HasErrors() = false; 오류 있어야 함")
		}
		if r.Valid {
			t.Error("Valid = true; AddError 후 false여야 함")
		}
	})
}

// TestReport_HasWarnings: Report.HasWarnings() 검증.
func TestReport_HasWarnings(t *testing.T) {
	t.Parallel()

	t.Run("경고 없음", func(t *testing.T) {
		t.Parallel()
		r := &dtcg.Report{Valid: true}
		if r.HasWarnings() {
			t.Error("HasWarnings() = true; 경고 없어야 함")
		}
	})

	t.Run("경고 있음", func(t *testing.T) {
		t.Parallel()
		r := &dtcg.Report{Valid: true}
		r.AddWarning(&dtcg.ValidationWarning{TokenPath: "test", Category: "color", Message: "경고"})
		if !r.HasWarnings() {
			t.Error("HasWarnings() = false; 경고 있어야 함")
		}
		// 경고는 Valid에 영향 없음
		if !r.Valid {
			t.Error("Valid = false; 경고는 Valid 변경 안해야 함")
		}
	})
}
