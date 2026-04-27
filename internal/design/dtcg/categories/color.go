// Package categories: DTCG 2025.10 §8 카테고리별 검증 함수 모음.
// 각 파일은 하나의 DTCG 카테고리를 담당한다.
package categories

import (
	"fmt"
	"regexp"
	"strings"
)

// hex 색상 패턴 - #rgb, #rgba, #rrggbb, #rrggbbaa
var (
	hexColor3  = regexp.MustCompile(`^#[0-9a-fA-F]{3}$`)
	hexColor4  = regexp.MustCompile(`^#[0-9a-fA-F]{4}$`)
	hexColor6  = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	hexColor8  = regexp.MustCompile(`^#[0-9a-fA-F]{8}$`)
	rgbPattern = regexp.MustCompile(`^rgba?\(`)
	hslPattern = regexp.MustCompile(`^hsla?\(`)
)

// namedColors: DTCG 2025.10에서 허용하는 CSS named 색상 (sRGB 기본 집합).
// CSS Level 1/2/3 named colors 전체 지원.
var namedColors = map[string]bool{
	"aliceblue": true, "antiquewhite": true, "aqua": true, "aquamarine": true,
	"azure": true, "beige": true, "bisque": true, "black": true,
	"blanchedalmond": true, "blue": true, "blueviolet": true, "brown": true,
	"burlywood": true, "cadetblue": true, "chartreuse": true, "chocolate": true,
	"coral": true, "cornflowerblue": true, "cornsilk": true, "crimson": true,
	"cyan": true, "darkblue": true, "darkcyan": true, "darkgoldenrod": true,
	"darkgray": true, "darkgreen": true, "darkgrey": true, "darkkhaki": true,
	"darkmagenta": true, "darkolivegreen": true, "darkorange": true, "darkorchid": true,
	"darkred": true, "darksalmon": true, "darkseagreen": true, "darkslateblue": true,
	"darkslategray": true, "darkslategrey": true, "darkturquoise": true, "darkviolet": true,
	"deeppink": true, "deepskyblue": true, "dimgray": true, "dimgrey": true,
	"dodgerblue": true, "firebrick": true, "floralwhite": true, "forestgreen": true,
	"fuchsia": true, "gainsboro": true, "ghostwhite": true, "gold": true,
	"goldenrod": true, "gray": true, "green": true, "greenyellow": true,
	"grey": true, "honeydew": true, "hotpink": true, "indianred": true,
	"indigo": true, "ivory": true, "khaki": true, "lavender": true,
	"lavenderblush": true, "lawngreen": true, "lemonchiffon": true, "lightblue": true,
	"lightcoral": true, "lightcyan": true, "lightgoldenrodyellow": true, "lightgray": true,
	"lightgreen": true, "lightgrey": true, "lightpink": true, "lightsalmon": true,
	"lightseagreen": true, "lightskyblue": true, "lightslategray": true, "lightslategrey": true,
	"lightsteelblue": true, "lightyellow": true, "lime": true, "limegreen": true,
	"linen": true, "magenta": true, "maroon": true, "mediumaquamarine": true,
	"mediumblue": true, "mediumorchid": true, "mediumpurple": true, "mediumseagreen": true,
	"mediumslateblue": true, "mediumspringgreen": true, "mediumturquoise": true, "mediumvioletred": true,
	"midnightblue": true, "mintcream": true, "mistyrose": true, "moccasin": true,
	"navajowhite": true, "navy": true, "oldlace": true, "olive": true,
	"olivedrab": true, "orange": true, "orangered": true, "orchid": true,
	"palegoldenrod": true, "palegreen": true, "paleturquoise": true, "palevioletred": true,
	"papayawhip": true, "peachpuff": true, "peru": true, "pink": true,
	"plum": true, "powderblue": true, "purple": true, "rebeccapurple": true,
	"red": true, "rosybrown": true, "royalblue": true, "saddlebrown": true,
	"salmon": true, "sandybrown": true, "seagreen": true, "seashell": true,
	"sienna": true, "silver": true, "skyblue": true, "slateblue": true,
	"slategray": true, "slategrey": true, "snow": true, "springgreen": true,
	"steelblue": true, "tan": true, "teal": true, "thistle": true,
	"tomato": true, "turquoise": true, "violet": true, "wheat": true,
	"white": true, "whitesmoke": true, "yellow": true, "yellowgreen": true,
	"transparent": true,
}

// ValidateColor: DTCG 2025.10 §8.1 color 카테고리 검증.
// 허용 형식: hex(#rgb, #rgba, #rrggbb, #rrggbbaa), rgb(), rgba(), hsl(), hsla(), named colors.
// sRGB 색공간만 지원 (DTCG 2025.10 minimum requirement).
func ValidateColor(tokenPath string, value any) error {
	// 에일리어스 참조는 검증 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	s, ok := value.(string)
	if !ok || s == "" {
		return fmt.Errorf("토큰 '%s': color 값은 문자열이어야 함 (got %T)", tokenPath, value)
	}

	// hex 색상 검증
	if strings.HasPrefix(s, "#") {
		if hexColor3.MatchString(s) || hexColor4.MatchString(s) ||
			hexColor6.MatchString(s) || hexColor8.MatchString(s) {
			return nil
		}
		return fmt.Errorf("토큰 '%s': 잘못된 hex 색상 형식 '%s' (허용: #rgb, #rgba, #rrggbb, #rrggbbaa)", tokenPath, s)
	}

	// rgb()/rgba() 검증
	if rgbPattern.MatchString(s) {
		return validateRGBColor(tokenPath, s)
	}

	// hsl()/hsla() 검증
	if hslPattern.MatchString(s) {
		return validateHSLColor(tokenPath, s)
	}

	// named color 검증
	if namedColors[strings.ToLower(s)] {
		return nil
	}

	return fmt.Errorf("토큰 '%s': 알 수 없는 색상 값 '%s'", tokenPath, s)
}

// validateRGBColor: rgb()/rgba() 형식 검증.
// 공백 포함 여부에 관계없이 파싱한다.
func validateRGBColor(tokenPath, s string) error {
	// 괄호 내용 추출
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("토큰 '%s': 잘못된 rgb() 형식 '%s'", tokenPath, s)
	}

	inner := s[start+1 : end]
	parts := strings.Split(inner, ",")

	// rgb()는 3개, rgba()는 4개 인수
	isRGBA := strings.HasPrefix(strings.ToLower(s), "rgba")
	expected := 3
	if isRGBA {
		expected = 4
	}
	if len(parts) != expected {
		return fmt.Errorf("토큰 '%s': rgb() 인수 %d개 필요, %d개 제공됨", tokenPath, expected, len(parts))
	}

	// r, g, b 채널 검증 (0-255 정수 또는 0%-100%)
	for i, p := range parts[:3] {
		p = strings.TrimSpace(p)
		if strings.HasSuffix(p, "%") {
			// 퍼센트 형식 허용
			continue
		}
		var v float64
		if _, err := fmt.Sscanf(p, "%f", &v); err != nil {
			return fmt.Errorf("토큰 '%s': rgb() %d번째 채널 '%s' 파싱 실패", tokenPath, i+1, p)
		}
		if v < 0 || v > 255 {
			return fmt.Errorf("토큰 '%s': rgb() %d번째 채널 %g 범위 초과 (0-255)", tokenPath, i+1, v)
		}
	}

	return nil
}

// validateHSLColor: hsl()/hsla() 형식 검증.
func validateHSLColor(tokenPath, s string) error {
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("토큰 '%s': 잘못된 hsl() 형식 '%s'", tokenPath, s)
	}

	inner := s[start+1 : end]
	parts := strings.Split(inner, ",")

	isHSLA := strings.HasPrefix(strings.ToLower(s), "hsla")
	expected := 3
	if isHSLA {
		expected = 4
	}
	if len(parts) != expected {
		return fmt.Errorf("토큰 '%s': hsl() 인수 %d개 필요, %d개 제공됨", tokenPath, expected, len(parts))
	}

	return nil
}

// IsAlias: DTCG 에일리어스 문법 ({group.token}) 감지.
// 에일리어스는 중괄호로 감싸인 비어있지 않은 참조 경로를 가진다.
func IsAlias(s string) bool {
	if len(s) < 3 {
		return false
	}
	return s[0] == '{' && s[len(s)-1] == '}' && len(s) > 2
}
