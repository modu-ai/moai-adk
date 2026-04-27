package categories_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// TestValidateFontFamily_Positive: fontFamily 카테고리 양성 케이스
// DTCG 2025.10: string 또는 string 배열 (폴백 체인).
func TestValidateFontFamily_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "단일 폰트명", value: "Roboto"},
		{name: "공백 포함 폰트명", value: "Open Sans"},
		{name: "시스템 폰트", value: "sans-serif"},
		{name: "단일 원소 배열", value: []any{"Roboto"}},
		{name: "폴백 체인", value: []any{"Roboto", "Arial", "sans-serif"}},
		{name: "공백포함 배열", value: []any{"Open Sans", "Helvetica Neue", "sans-serif"}},
		// 에일리어스 참조
		{name: "에일리어스 참조", value: "{typography.font-family.primary}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateFontFamily("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateFontFamily(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateFontFamily_Negative: fontFamily 카테고리 음성 케이스
func TestValidateFontFamily_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "빈 문자열", value: ""},
		{name: "숫자 타입", value: 42},
		{name: "nil", value: nil},
		{name: "빈 배열", value: []any{}},
		{name: "배열에 비문자열", value: []any{"Roboto", 42}},
		{name: "배열에 빈 문자열", value: []any{"Roboto", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateFontFamily("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateFontFamily(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// TestValidateFontWeight_Positive: fontWeight 카테고리 양성 케이스
// DTCG 2025.10: 100-900 수치 또는 named ("thin","light" 등).
func TestValidateFontWeight_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		// 수치 (100 단위)
		{name: "100 Thin", value: float64(100)},
		{name: "200 ExtraLight", value: float64(200)},
		{name: "300 Light", value: float64(300)},
		{name: "400 Normal", value: float64(400)},
		{name: "500 Medium", value: float64(500)},
		{name: "600 SemiBold", value: float64(600)},
		{name: "700 Bold", value: float64(700)},
		{name: "800 ExtraBold", value: float64(800)},
		{name: "900 Black", value: float64(900)},
		// named
		{name: "thin", value: "thin"},
		{name: "light", value: "light"},
		{name: "normal", value: "normal"},
		{name: "regular", value: "regular"},
		{name: "medium", value: "medium"},
		{name: "bold", value: "bold"},
		{name: "heavy", value: "heavy"},
		// 에일리어스 참조
		{name: "에일리어스", value: "{typography.font-weight.bold}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateFontWeight("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateFontWeight(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateFontWeight_Negative: fontWeight 카테고리 음성 케이스
func TestValidateFontWeight_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "범위 초과 1000", value: float64(1000)},
		{name: "범위 미만 0", value: float64(0)},
		{name: "50 비표준", value: float64(50)},
		{name: "알 수 없는 named", value: "ultrablack"},
		{name: "빈 문자열", value: ""},
		{name: "nil", value: nil},
		{name: "bool", value: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateFontWeight("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateFontWeight(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// TestValidateFont_Positive: font 복합 카테고리 양성 케이스
// DTCG 2025.10: {family, size, weight, style?, lineHeight?} 복합 토큰.
func TestValidateFont_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name: "필수 필드만",
			value: map[string]any{
				"family": "Roboto",
				"size":   map[string]any{"value": 16.0, "unit": "px"},
				"weight": float64(400),
			},
		},
		{
			name: "전체 필드",
			value: map[string]any{
				"family":     []any{"Roboto", "sans-serif"},
				"size":       "1rem",
				"weight":     "bold",
				"style":      "italic",
				"lineHeight": "1.5",
			},
		},
		{
			name: "에일리어스 참조 family",
			value: map[string]any{
				"family": "{typography.font-family.primary}",
				"size":   map[string]any{"value": 16.0, "unit": "px"},
				"weight": "{typography.font-weight.body}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateFont("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateFont(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateFont_Negative: font 복합 카테고리 음성 케이스
func TestValidateFont_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "map이 아닌 타입", value: "Roboto"},
		{name: "nil", value: nil},
		{name: "family 누락", value: map[string]any{"size": "16px", "weight": float64(400)}},
		{name: "size 누락", value: map[string]any{"family": "Roboto", "weight": float64(400)}},
		{name: "weight 누락", value: map[string]any{"family": "Roboto", "size": "16px"}},
		{
			name: "잘못된 size 단위",
			value: map[string]any{
				"family": "Roboto",
				"size":   map[string]any{"value": 16.0, "unit": "vw"},
				"weight": float64(400),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateFont("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateFont(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}
