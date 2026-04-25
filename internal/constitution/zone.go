// Package constitution은 MoAI-ADK 규칙 트리의 FROZEN/EVOLVABLE zone 모델을 구현한다.
// zone registry를 로드하고 조회하는 API를 제공하며,
// CLI(moai constitution list)와 doctor 체크에서 공통으로 사용된다.
package constitution

import (
	"fmt"
	"strings"
)

// Zone은 MoAI-ADK 규칙 조항의 zone을 나타내는 열거형이다.
// Zone은 uint8 기반으로 구현되어 향후 값 확장 여유 공간을 확보한다.
// SPEC-V3R2-CON-001 REQ-CON-001-003 직접 구현.
type Zone uint8

const (
	// ZoneFrozen은 변경 불가 조항 (constitutional invariant)을 나타낸다.
	// amendment 시 canary shadow evaluation 필수.
	ZoneFrozen Zone = iota // = 0
	// ZoneEvolvable은 graduation protocol을 통해 진화 가능한 조항을 나타낸다.
	ZoneEvolvable // = 1
)

// String은 Zone 값의 사람이 읽기 쉬운 문자열 표현을 반환한다.
func (z Zone) String() string {
	switch z {
	case ZoneFrozen:
		return "Frozen"
	case ZoneEvolvable:
		return "Evolvable"
	default:
		return fmt.Sprintf("Zone(%d)", uint8(z))
	}
}

// ParseZone은 문자열을 Zone 값으로 파싱한다.
// 대소문자를 무시하며, 알 수 없는 값은 오류를 반환한다.
func ParseZone(s string) (Zone, error) {
	switch strings.ToLower(s) {
	case "frozen":
		return ZoneFrozen, nil
	case "evolvable":
		return ZoneEvolvable, nil
	default:
		return 0, fmt.Errorf("알 수 없는 zone 값: %q (허용: Frozen, Evolvable)", s)
	}
}
