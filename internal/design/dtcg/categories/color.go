// Package categories: DTCG 2025.10 §8 Category-specific validation function collection.
// Each file handles one DTCG category.
package categories

import (
	"fmt"
	"regexp"
	"strings"
)

// Hex color patterns - #rgb, #rgba, #rrggbb, #rrggbbaa
var (
	hexColor3  = regexp.MustCompile(`^#[0-9a-fA-F]{3}$`)
	hexColor4  = regexp.MustCompile(`^#[0-9a-fA-F]{4}$`)
	hexColor6  = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	hexColor8  = regexp.MustCompile(`^#[0-9a-fA-F]{8}$`)
	rgbPattern = regexp.MustCompile(`^rgba?\(`)
	hslPattern = regexp.MustCompile(`^hsla?\(`)
)

// namedColors: CSS named colors allowed in DTCG 2025.10 (sRGB base set).
// Full support for CSS Level 1/2/3 named colors.
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

// ValidateColor: DTCG 2025.10 §8.1 color category validation.
// Allowed formats: hex(#rgb, #rgba, #rrggbb, #rrggbbaa), rgb(), rgba(), hsl(), hsla(), named colors.
// Only sRGB color space supported (DTCG 2025.10 minimum requirement).
func ValidateColor(tokenPath string, value any) error {
	// Alias references pass validation
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	s, ok := value.(string)
	if !ok || s == "" {
		return fmt.Errorf("token '%s': color value must be a string (got %T)", tokenPath, value)
	}

	// Hex color validation
	if strings.HasPrefix(s, "#") {
		if hexColor3.MatchString(s) || hexColor4.MatchString(s) ||
			hexColor6.MatchString(s) || hexColor8.MatchString(s) {
			return nil
		}
		return fmt.Errorf("token '%s': invalid hex color format '%s' (allowed: #rgb, #rgba, #rrggbb, #rrggbbaa)", tokenPath, s)
	}

	// rgb()/rgba() validation
	if rgbPattern.MatchString(s) {
		return validateRGBColor(tokenPath, s)
	}

	// hsl()/hsla() validation
	if hslPattern.MatchString(s) {
		return validateHSLColor(tokenPath, s)
	}

	// Named color validation
	if namedColors[strings.ToLower(s)] {
		return nil
	}

	return fmt.Errorf("token '%s': unknown color value '%s'", tokenPath, s)
}

// validateRGBColor: rgb()/rgba() format validation.
// Parses regardless of whitespace inclusion.
func validateRGBColor(tokenPath, s string) error {
	// Extract bracket contents
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("token '%s': invalid rgb() format '%s'", tokenPath, s)
	}

	inner := s[start+1 : end]
	parts := strings.Split(inner, ",")

	// rgb() takes 3, rgba() takes 4 arguments
	isRGBA := strings.HasPrefix(strings.ToLower(s), "rgba")
	expected := 3
	if isRGBA {
		expected = 4
	}
	if len(parts) != expected {
		return fmt.Errorf("token '%s': rgb() requires %d arguments, got %d", tokenPath, expected, len(parts))
	}

	// r, g, b channel validation (0-255 integer or 0%-100%)
	for i, p := range parts[:3] {
		p = strings.TrimSpace(p)
		if strings.HasSuffix(p, "%") {
			// Allow percent format
			continue
		}
		var v float64
		if _, err := fmt.Sscanf(p, "%f", &v); err != nil {
			return fmt.Errorf("token '%s': rgb() channel %d failed to parse '%s'", tokenPath, i+1, p)
		}
		if v < 0 || v > 255 {
			return fmt.Errorf("token '%s': rgb() channel %d value %g out of range (0-255)", tokenPath, i+1, v)
		}
	}

	return nil
}

// validateHSLColor: hsl()/hsla() format validation.
func validateHSLColor(tokenPath, s string) error {
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("token '%s': invalid hsl() format '%s'", tokenPath, s)
	}

	inner := s[start+1 : end]
	parts := strings.Split(inner, ",")

	isHSLA := strings.HasPrefix(strings.ToLower(s), "hsla")
	expected := 3
	if isHSLA {
		expected = 4
	}
	if len(parts) != expected {
		return fmt.Errorf("token '%s': hsl() requires %d arguments, got %d", tokenPath, expected, len(parts))
	}

	return nil
}

// IsAlias: DTCG alias syntax ({group.token}) detection.
// Aliases have non-empty reference paths wrapped in curly braces.
func IsAlias(s string) bool {
	if len(s) < 3 {
		return false
	}
	return s[0] == '{' && s[len(s)-1] == '}' && len(s) > 2
}
