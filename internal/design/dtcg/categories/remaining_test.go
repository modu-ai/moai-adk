package categories_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// --- duration ---

// TestValidateDuration_Positive: duration 카테고리 양성 케이스
// DTCG 2025.10: {value: number, unit: "ms"|"s"} 구조화 또는 "300ms" 레거시.
func TestValidateDuration_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "ms 구조화", value: map[string]any{"value": 300.0, "unit": "ms"}},
		{name: "s 구조화", value: map[string]any{"value": 0.3, "unit": "s"}},
		{name: "ms 레거시", value: "300ms"},
		{name: "s 레거시", value: "0.3s"},
		{name: "ms 0", value: map[string]any{"value": 0.0, "unit": "ms"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateDuration("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateDuration(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateDuration_Negative: duration 카테고리 음성 케이스
func TestValidateDuration_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "지원안되는 단위 min", value: map[string]any{"value": 1.0, "unit": "min"}},
		{name: "unit 누락", value: map[string]any{"value": 300.0}},
		{name: "레거시 잘못된 단위", value: "300px"},
		{name: "숫자만", value: "300"},
		{name: "nil", value: nil},
		{name: "숫자 타입", value: 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateDuration("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateDuration(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// --- cubicBezier ---

// TestValidateCubicBezier_Positive: cubicBezier 카테고리 양성 케이스
// DTCG 2025.10: [x1,y1,x2,y2] x1,x2 ∈ [0,1], y1,y2 제한 없음.
func TestValidateCubicBezier_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "ease", value: []any{0.25, 0.1, 0.25, 1.0}},
		{name: "ease-in", value: []any{0.42, 0.0, 1.0, 1.0}},
		{name: "ease-out", value: []any{0.0, 0.0, 0.58, 1.0}},
		{name: "linear", value: []any{0.0, 0.0, 1.0, 1.0}},
		{name: "y 음수 허용", value: []any{0.5, -0.5, 0.5, 1.5}},
		{name: "x 경계 0", value: []any{0.0, 0.0, 0.0, 0.0}},
		{name: "x 경계 1", value: []any{1.0, 1.0, 1.0, 1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateCubicBezier("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateCubicBezier(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateCubicBezier_Negative: cubicBezier 카테고리 음성 케이스
func TestValidateCubicBezier_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "x1 > 1", value: []any{1.5, 0.0, 0.5, 1.0}},
		{name: "x2 > 1", value: []any{0.5, 0.0, 1.5, 1.0}},
		{name: "x1 < 0", value: []any{-0.5, 0.0, 0.5, 1.0}},
		{name: "x2 < 0", value: []any{0.5, 0.0, -0.5, 1.0}},
		{name: "원소 3개", value: []any{0.25, 0.1, 0.25}},
		{name: "원소 5개", value: []any{0.25, 0.1, 0.25, 1.0, 0.5}},
		{name: "빈 배열", value: []any{}},
		{name: "문자열 타입", value: "ease"},
		{name: "nil", value: nil},
		{name: "비숫자 원소", value: []any{"0.25", 0.1, 0.25, 1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateCubicBezier("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateCubicBezier(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// --- number ---

// TestValidateNumber_Positive: number 카테고리 양성 케이스
func TestValidateNumber_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "정수", value: float64(0)},
		{name: "양수", value: float64(42)},
		{name: "음수", value: float64(-1)},
		{name: "소수", value: 3.14},
		{name: "정수형 int", value: 100},
		{name: "float32", value: float32(1.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateNumber("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateNumber(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateNumber_Negative: number 카테고리 음성 케이스
func TestValidateNumber_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "문자열", value: "42"},
		{name: "bool", value: true},
		{name: "nil", value: nil},
		{name: "map", value: map[string]any{"value": 42.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateNumber("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateNumber(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// --- strokeStyle ---

// TestValidateStrokeStyle_Positive: strokeStyle 카테고리 양성 케이스
// DTCG 2025.10: enum string 또는 {dashArray, lineCap} 복합.
func TestValidateStrokeStyle_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "solid", value: "solid"},
		{name: "dashed", value: "dashed"},
		{name: "dotted", value: "dotted"},
		{name: "double", value: "double"},
		{name: "groove", value: "groove"},
		{name: "ridge", value: "ridge"},
		{name: "outset", value: "outset"},
		{name: "inset", value: "inset"},
		{
			name: "복합 dashArray",
			value: map[string]any{
				"dashArray": []any{"4px", "2px"},
				"lineCap":   "round",
			},
		},
		{
			name: "복합 lineCap butt",
			value: map[string]any{
				"dashArray": []any{"2px"},
				"lineCap":   "butt",
			},
		},
		{
			name: "복합 lineCap square",
			value: map[string]any{
				"dashArray": []any{"4px", "2px", "1px"},
				"lineCap":   "square",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateStrokeStyle("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateStrokeStyle(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateStrokeStyle_Negative: strokeStyle 카테고리 음성 케이스
func TestValidateStrokeStyle_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "알 수 없는 enum", value: "wavy"},
		{name: "알 수 없는 enum2", value: "none"},
		{name: "nil", value: nil},
		{name: "숫자", value: 1},
		{name: "dashArray 누락", value: map[string]any{"lineCap": "round"}},
		{name: "lineCap 누락", value: map[string]any{"dashArray": []any{"4px"}}},
		{name: "잘못된 lineCap", value: map[string]any{"dashArray": []any{"4px"}, "lineCap": "flat"}},
		{name: "빈 dashArray", value: map[string]any{"dashArray": []any{}, "lineCap": "round"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateStrokeStyle("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateStrokeStyle(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// --- transition ---

// TestValidateTransition_Positive: transition 복합 카테고리 양성 케이스
// DTCG 2025.10: {duration, delay?, timingFunction: cubicBezier}.
func TestValidateTransition_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name: "기본 transition",
			value: map[string]any{
				"duration":       map[string]any{"value": 300.0, "unit": "ms"},
				"timingFunction": []any{0.25, 0.1, 0.25, 1.0},
			},
		},
		{
			name: "delay 포함",
			value: map[string]any{
				"duration":       "300ms",
				"delay":          "100ms",
				"timingFunction": []any{0.42, 0.0, 1.0, 1.0},
			},
		},
		{
			name: "에일리어스 참조",
			value: map[string]any{
				"duration":       "{animation.duration.fast}",
				"timingFunction": "{animation.easing.ease}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateTransition("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateTransition(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateTransition_Negative: transition 카테고리 음성 케이스
func TestValidateTransition_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "문자열", value: "300ms ease"},
		{name: "duration 누락", value: map[string]any{"timingFunction": []any{0.25, 0.1, 0.25, 1.0}}},
		{name: "timingFunction 누락", value: map[string]any{"duration": "300ms"}},
		{
			name: "잘못된 cubicBezier",
			value: map[string]any{
				"duration":       "300ms",
				"timingFunction": []any{1.5, 0.0, 0.5, 1.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateTransition("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateTransition(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}

// --- gradient ---

// TestValidateGradient_Positive: gradient 카테고리 양성 케이스
// DTCG 2025.10: [{color, position: 0..1}, ...] 배열.
func TestValidateGradient_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name: "2-stop 그라디언트",
			value: []any{
				map[string]any{"color": "#0066ff", "position": 0.0},
				map[string]any{"color": "#00cc66", "position": 1.0},
			},
		},
		{
			name: "3-stop 그라디언트",
			value: []any{
				map[string]any{"color": "#ff0000", "position": 0.0},
				map[string]any{"color": "#ffffff", "position": 0.5},
				map[string]any{"color": "#0000ff", "position": 1.0},
			},
		},
		{
			name: "에일리어스 color 참조",
			value: []any{
				map[string]any{"color": "{color.brand.primary}", "position": 0.0},
				map[string]any{"color": "{color.brand.secondary}", "position": 1.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateGradient("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateGradient(%v) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateGradient_Negative: gradient 카테고리 음성 케이스
func TestValidateGradient_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "문자열", value: "linear-gradient(...)"},
		{name: "빈 배열", value: []any{}},
		{name: "단일 stop만", value: []any{map[string]any{"color": "#fff", "position": 0.0}}},
		{name: "color 누락", value: []any{map[string]any{"position": 0.0}, map[string]any{"color": "#fff", "position": 1.0}}},
		{name: "position 누락", value: []any{map[string]any{"color": "#fff"}, map[string]any{"color": "#000", "position": 1.0}}},
		{name: "position > 1", value: []any{map[string]any{"color": "#fff", "position": 0.0}, map[string]any{"color": "#000", "position": 1.5}}},
		{name: "position < 0", value: []any{map[string]any{"color": "#fff", "position": -0.1}, map[string]any{"color": "#000", "position": 1.0}}},
		{name: "잘못된 color", value: []any{map[string]any{"color": "notacolor", "position": 0.0}, map[string]any{"color": "#000", "position": 1.0}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateGradient("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateGradient(%v) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}
