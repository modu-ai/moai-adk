package categories_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// TestValidateDimension_Positive: 유효한 dimension 형식 검증 (양성 케이스)
// DTCG 2025.10 §8.2: 구조화된 형식 {value, unit} 및 레거시 string 형식 "16px" 모두 허용.
func TestValidateDimension_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		// 구조화된 형식 (DTCG 2025.10 권장)
		{name: "px 정수", value: map[string]any{"value": 16.0, "unit": "px"}},
		{name: "rem 소수", value: map[string]any{"value": 1.5, "unit": "rem"}},
		{name: "em 기본", value: map[string]any{"value": 1.0, "unit": "em"}},
		{name: "퍼센트", value: map[string]any{"value": 100.0, "unit": "%"}},
		{name: "px 영", value: map[string]any{"value": 0.0, "unit": "px"}},
		{name: "int value px", value: map[string]any{"value": float64(8), "unit": "px"}},
		// 레거시 string 형식 (하위 호환)
		{name: "레거시 px", value: "16px"},
		{name: "레거시 rem", value: "1.5rem"},
		{name: "레거시 em", value: "1em"},
		{name: "레거시 퍼센트", value: "100%"},
		{name: "레거시 영px", value: "0px"},
		{name: "레거시 소수px", value: "0.5px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateDimension("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateDimension(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateDimension_Negative: 유효하지 않은 dimension 형식 검증 (음성 케이스)
func TestValidateDimension_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		// 구조화된 형식 오류
		{name: "지원안되는 단위 cm", value: map[string]any{"value": 1.0, "unit": "cm"}},
		{name: "지원안되는 단위 pt", value: map[string]any{"value": 1.0, "unit": "pt"}},
		{name: "unit 필드 누락", value: map[string]any{"value": 1.0}},
		{name: "value 필드 누락", value: map[string]any{"unit": "px"}},
		{name: "value 문자열", value: map[string]any{"value": "16", "unit": "px"}},
		// 레거시 string 오류
		{name: "단위 없는 숫자 string", value: "16"},
		{name: "잘못된 단위", value: "16vw"},
		{name: "단위만", value: "px"},
		{name: "빈 문자열", value: ""},
		// 타입 오류
		{name: "숫자 타입", value: 16},
		{name: "bool 타입", value: true},
		{name: "nil", value: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateDimension("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateDimension(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}
