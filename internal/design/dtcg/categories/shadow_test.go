package categories_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// TestValidateShadow_Positive: shadow 복합 카테고리 양성 케이스
// DTCG 2025.10: 단일 또는 다층({color, offsetX, offsetY, blur, spread, inset?}) shadow.
func TestValidateShadow_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name: "단일 그림자 필수 필드",
			value: map[string]any{
				"color":   "#000000",
				"offsetX": map[string]any{"value": 0.0, "unit": "px"},
				"offsetY": map[string]any{"value": 4.0, "unit": "px"},
				"blur":    map[string]any{"value": 8.0, "unit": "px"},
				"spread":  map[string]any{"value": 0.0, "unit": "px"},
			},
		},
		{
			name: "단일 그림자 inset 포함",
			value: map[string]any{
				"color":   "rgba(0,0,0,0.2)",
				"offsetX": "0px",
				"offsetY": "2px",
				"blur":    "4px",
				"spread":  "0px",
				"inset":   true,
			},
		},
		{
			name: "다층 그림자 배열",
			value: []any{
				map[string]any{
					"color":   "#000",
					"offsetX": "0px",
					"offsetY": "2px",
					"blur":    "4px",
					"spread":  "0px",
				},
				map[string]any{
					"color":   "rgba(0,0,0,0.1)",
					"offsetX": "0px",
					"offsetY": "8px",
					"blur":    "16px",
					"spread":  "0px",
				},
			},
		},
		{
			name: "에일리어스 참조",
			value: map[string]any{
				"color":   "{color.shadow.default}",
				"offsetX": "{spacing.0}",
				"offsetY": map[string]any{"value": 4.0, "unit": "px"},
				"blur":    map[string]any{"value": 8.0, "unit": "px"},
				"spread":  map[string]any{"value": 0.0, "unit": "px"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateShadow("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateShadow(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateShadow_Negative: shadow 카테고리 음성 케이스
func TestValidateShadow_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "문자열 타입", value: "0px 4px 8px #000"},
		{name: "color 누락", value: map[string]any{"offsetX": "0px", "offsetY": "4px", "blur": "8px", "spread": "0px"}},
		{name: "offsetX 누락", value: map[string]any{"color": "#000", "offsetY": "4px", "blur": "8px", "spread": "0px"}},
		{name: "offsetY 누락", value: map[string]any{"color": "#000", "offsetX": "0px", "blur": "8px", "spread": "0px"}},
		{name: "blur 누락", value: map[string]any{"color": "#000", "offsetX": "0px", "offsetY": "4px", "spread": "0px"}},
		{name: "spread 누락", value: map[string]any{"color": "#000", "offsetX": "0px", "offsetY": "4px", "blur": "8px"}},
		{name: "잘못된 color", value: map[string]any{"color": "notacolor", "offsetX": "0px", "offsetY": "4px", "blur": "8px", "spread": "0px"}},
		{name: "빈 배열", value: []any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateShadow("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateShadow(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// TestValidateBorder_Positive: border 복합 카테고리 양성 케이스
// DTCG 2025.10: {color, width: dimension, style: strokeStyle} 복합 토큰.
func TestValidateBorder_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name: "기본 border",
			value: map[string]any{
				"color": "#000000",
				"width": map[string]any{"value": 1.0, "unit": "px"},
				"style": "solid",
			},
		},
		{
			name: "레거시 width 문자열",
			value: map[string]any{
				"color": "rgba(0,0,0,0.2)",
				"width": "2px",
				"style": "dashed",
			},
		},
		{
			name: "strokeStyle 복합",
			value: map[string]any{
				"color": "#333",
				"width": "1px",
				"style": map[string]any{
					"dashArray": []any{"4px", "2px"},
					"lineCap":   "round",
				},
			},
		},
		{
			name: "에일리어스 참조",
			value: map[string]any{
				"color": "{color.border.default}",
				"width": "{border.width.sm}",
				"style": "{border.style.default}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateBorder("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateBorder(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateBorder_Negative: border 카테고리 음성 케이스
func TestValidateBorder_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "문자열 타입", value: "1px solid #000"},
		{name: "color 누락", value: map[string]any{"width": "1px", "style": "solid"}},
		{name: "width 누락", value: map[string]any{"color": "#000", "style": "solid"}},
		{name: "style 누락", value: map[string]any{"color": "#000", "width": "1px"}},
		{name: "잘못된 color", value: map[string]any{"color": "notacolor", "width": "1px", "style": "solid"}},
		{name: "잘못된 width 단위", value: map[string]any{"color": "#000", "width": map[string]any{"value": 1.0, "unit": "vw"}, "style": "solid"}},
		{name: "잘못된 style", value: map[string]any{"color": "#000", "width": "1px", "style": "wavy"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateBorder("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateBorder(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}
