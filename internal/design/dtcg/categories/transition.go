package categories

import "fmt"

// ValidateTransition: DTCG 2025.10 transition 복합 카테고리 검증.
// 필수: duration, timingFunction(cubicBezier) / 선택: delay(duration).
func ValidateTransition(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': transition 값은 map이어야 함 (got %T)", tokenPath, value)
	}

	// duration 필수 필드 검증
	durationVal, ok := m["duration"]
	if !ok {
		return fmt.Errorf("토큰 '%s': transition 'duration' 필드 누락", tokenPath)
	}
	if s, isStr := durationVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateDuration(tokenPath+".duration", durationVal); err != nil {
			return err
		}
	}

	// timingFunction 필수 필드 검증 (cubicBezier)
	tfVal, ok := m["timingFunction"]
	if !ok {
		return fmt.Errorf("토큰 '%s': transition 'timingFunction' 필드 누락", tokenPath)
	}
	if s, isStr := tfVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateCubicBezier(tokenPath+".timingFunction", tfVal); err != nil {
			return err
		}
	}

	// delay 선택 필드 검증
	if delayVal, exists := m["delay"]; exists {
		if s, isStr := delayVal.(string); !isStr || !IsAlias(s) {
			if err := ValidateDuration(tokenPath+".delay", delayVal); err != nil {
				return err
			}
		}
	}

	return nil
}
