package categories_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// TestValidateColor_Positive: 유효한 색상 형식 검증 (양성 케이스)
// DTCG 2025.10 §8.1 color 카테고리 규칙에 따른 hex, rgb, hsl, named 색상 포함.
func TestValidateColor_Positive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		// hex 3자리 (#rgb)
		{name: "hex 3자리 흰색", value: "#fff"},
		{name: "hex 3자리 검정", value: "#000"},
		{name: "hex 3자리 혼합", value: "#a1b"},
		// hex 4자리 (#rgba)
		{name: "hex 4자리 반투명", value: "#ffff"},
		{name: "hex 4자리 투명", value: "#0000"},
		// hex 6자리 (#rrggbb)
		{name: "hex 6자리 흰색", value: "#ffffff"},
		{name: "hex 6자리 검정", value: "#000000"},
		{name: "hex 6자리 파랑", value: "#0066ff"},
		{name: "hex 6자리 대문자", value: "#0066FF"},
		// hex 8자리 (#rrggbbaa)
		{name: "hex 8자리 반투명", value: "#0066ff80"},
		{name: "hex 8자리 완전불투명", value: "#ffffffff"},
		// rgb()
		{name: "rgb() 정수", value: "rgb(0, 128, 255)"},
		{name: "rgb() 공백없음", value: "rgb(0,128,255)"},
		{name: "rgb() 퍼센트", value: "rgb(0%, 50%, 100%)"},
		// rgba()
		{name: "rgba() 반투명", value: "rgba(0, 128, 255, 0.5)"},
		{name: "rgba() 완전투명", value: "rgba(0,0,0,0)"},
		// hsl()
		{name: "hsl() 기본", value: "hsl(240, 100%, 50%)"},
		{name: "hsl() 공백없음", value: "hsl(240,100%,50%)"},
		// hsla()
		{name: "hsla() 반투명", value: "hsla(240, 100%, 50%, 0.5)"},
		// named 색상 (sRGB 기본 집합)
		{name: "named blue", value: "blue"},
		{name: "named red", value: "red"},
		{name: "named white", value: "white"},
		{name: "named black", value: "black"},
		{name: "named transparent", value: "transparent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateColor("test.token", tt.value)
			if err != nil {
				t.Errorf("ValidateColor(%q) = %v; 오류 없어야 함", tt.value, err)
			}
		})
	}
}

// TestValidateColor_Negative: 유효하지 않은 색상 형식 검증 (음성 케이스)
func TestValidateColor_Negative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		// 잘못된 hex 형식
		{name: "hex 2자리 (짧음)", value: "#ff"},
		{name: "hex 5자리 (비표준)", value: "#fffff"},
		{name: "hex 7자리 (비표준)", value: "#fffffff"},
		{name: "hex 접두사 없음", value: "ffffff"},
		{name: "hex 잘못된 문자", value: "#xyzxyz"},
		// 잘못된 rgb
		{name: "rgb 범위초과", value: "rgb(300, 0, 0)"},
		{name: "rgb 음수", value: "rgb(-1, 0, 0)"},
		{name: "rgb 인수부족", value: "rgb(0, 0)"},
		// 잘못된 값 타입
		{name: "숫자 타입", value: 12345},
		{name: "bool 타입", value: true},
		{name: "nil 값", value: nil},
		// 빈 문자열
		{name: "빈 문자열", value: ""},
		// 알 수 없는 named
		{name: "알 수 없는 색상명", value: "notacolor"},
		{name: "공백 포함", value: "dark blue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := categories.ValidateColor("test.token", tt.value)
			if err == nil {
				t.Errorf("ValidateColor(%q) = nil; 오류 반환해야 함", tt.value)
			}
		})
	}
}
