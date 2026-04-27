package categories

import "fmt"

// validStrokeStyleEnums: DTCG 2025.10 strokeStyle 허용 enum 값.
var validStrokeStyleEnums = map[string]bool{
	"solid":  true,
	"dashed": true,
	"dotted": true,
	"double": true,
	"groove": true,
	"ridge":  true,
	"outset": true,
	"inset":  true,
}

// validLineCaps: strokeStyle 복합 형식의 lineCap 허용 값.
var validLineCaps = map[string]bool{
	"butt":   true,
	"round":  true,
	"square": true,
}

// ValidateStrokeStyle: DTCG 2025.10 strokeStyle 카테고리 검증.
// 허용: enum string 또는 {dashArray: [<dimension>...], lineCap: "butt"|"round"|"square"}.
func ValidateStrokeStyle(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case string:
		if !validStrokeStyleEnums[v] {
			return fmt.Errorf("토큰 '%s': strokeStyle enum '%s' 미지원 (허용: solid, dashed, dotted, double, groove, ridge, outset, inset)", tokenPath, v)
		}
		return nil
	case map[string]any:
		return validateStrokeStyleMap(tokenPath, v)
	default:
		return fmt.Errorf("토큰 '%s': strokeStyle 값은 string enum 또는 map이어야 함 (got %T)", tokenPath, value)
	}
}

// validateStrokeStyleMap: 복합 strokeStyle {dashArray, lineCap} 검증.
func validateStrokeStyleMap(tokenPath string, m map[string]any) error {
	// dashArray 필드 검증
	rawDA, ok := m["dashArray"]
	if !ok {
		return fmt.Errorf("토큰 '%s': strokeStyle map에 'dashArray' 필드 누락", tokenPath)
	}
	dashArray, ok := rawDA.([]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': strokeStyle 'dashArray'는 배열이어야 함 (got %T)", tokenPath, rawDA)
	}
	if len(dashArray) == 0 {
		return fmt.Errorf("토큰 '%s': strokeStyle 'dashArray' 배열이 비어있음", tokenPath)
	}
	// dashArray 각 원소는 dimension 형식이어야 함
	for i, item := range dashArray {
		if err := ValidateDimension(fmt.Sprintf("%s.dashArray[%d]", tokenPath, i), item); err != nil {
			return err
		}
	}

	// lineCap 필드 검증
	rawLC, ok := m["lineCap"]
	if !ok {
		return fmt.Errorf("토큰 '%s': strokeStyle map에 'lineCap' 필드 누락", tokenPath)
	}
	lineCap, ok := rawLC.(string)
	if !ok {
		return fmt.Errorf("토큰 '%s': strokeStyle 'lineCap'은 string이어야 함 (got %T)", tokenPath, rawLC)
	}
	if !validLineCaps[lineCap] {
		return fmt.Errorf("토큰 '%s': strokeStyle 'lineCap' 값 '%s' 미지원 (허용: butt, round, square)", tokenPath, lineCap)
	}

	return nil
}
