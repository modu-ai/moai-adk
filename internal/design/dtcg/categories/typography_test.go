package categories_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// TestValidateTypography_Positive: typography 복합 카테고리 양성 케이스
// DTCG 2025.10: font 복합 + letterSpacing, textDecoration, textTransform 확장.
func TestValidateTypography_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name: "기본 필드",
			value: map[string]any{
				"family": "Roboto",
				"size":   map[string]any{"value": 16.0, "unit": "px"},
				"weight": float64(400),
			},
		},
		{
			name: "모든 확장 필드 포함",
			value: map[string]any{
				"family":          []any{"Roboto", "sans-serif"},
				"size":            "1rem",
				"weight":          "bold",
				"style":           "normal",
				"lineHeight":      "1.5",
				"letterSpacing":   "0.02em",
				"textDecoration":  "none",
				"textTransform":   "uppercase",
			},
		},
		{
			name: "에일리어스로 모든 필드 참조",
			value: map[string]any{
				"family": "{font.family.primary}",
				"size":   "{font.size.body}",
				"weight": "{font.weight.normal}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateTypography("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateTypography(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateTypography_Negative: typography 복합 카테고리 음성 케이스
func TestValidateTypography_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "문자열 타입", value: "Roboto"},
		{name: "family 누락", value: map[string]any{"size": "16px", "weight": float64(400)}},
		{name: "size 누락", value: map[string]any{"family": "Roboto", "weight": float64(400)}},
		{name: "weight 누락", value: map[string]any{"family": "Roboto", "size": "16px"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateTypography("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateTypography(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// TestIsAlias: 에일리어스 문법 ({group.token}) 감지 테스트
func TestIsAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "기본 에일리어스", value: "{color.primary}", want: true},
		{name: "중첩 에일리어스", value: "{typography.font.family}", want: true},
		{name: "일반 문자열", value: "Roboto", want: false},
		{name: "중괄호 없음", value: "color.primary", want: false},
		{name: "여는 중괄호만", value: "{color.primary", want: false},
		{name: "닫는 중괄호만", value: "color.primary}", want: false},
		{name: "빈 중괄호", value: "{}", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := categories.IsAlias(tt.value)
			if got != tt.want {
				t.Errorf("IsAlias(%q) = %v; want %v", tt.value, got, tt.want)
			}
		})
	}
}
