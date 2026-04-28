package constitution

import (
	"fmt"
	"regexp"
)

// ruleIDPattern은 CONST-V3R2-NNN 형식의 ID를 검증하는 정규표현식 상수이다.
// NNN은 3자리 이상의 숫자.
const ruleIDPattern = `^CONST-V3R2-\d{3,}$`

// ruleIDRegexp는 컴파일된 ruleIDPattern이다.
var ruleIDRegexp = regexp.MustCompile(ruleIDPattern)

// Rule은 zone registry의 단일 엔트리를 나타낸다.
// 정확히 6개의 exported 필드를 가지며, yaml.v3 태그로 registry YAML 스키마와 매핑된다.
// SPEC-V3R2-CON-001 REQ-CON-001-004 직접 구현.
//
// Orphan 필드(unexported)는 loader가 참조 파일 부재 시 내부적으로 설정한다.
// yaml 태그가 없으므로 registry 파일에는 나타나지 않는다.
type Rule struct {
	// ID는 CONST-V3R2-NNN 형식의 고유 식별자이다.
	ID string `yaml:"id"`
	// Zone은 조항의 zone (Frozen 또는 Evolvable).
	Zone Zone `yaml:"zone"`
	// File은 조항이 위치한 규칙 파일 경로이다.
	File string `yaml:"file"`
	// Anchor는 파일 내 위치 앵커 (#섹션명 또는 L라인번호).
	Anchor string `yaml:"anchor"`
	// Clause는 HARD 조항의 verbatim 텍스트이다.
	Clause string `yaml:"clause"`
	// CanaryGate는 amendment 시 shadow evaluation이 필요한지 여부이다.
	// Frozen → true, Evolvable → false (기본값).
	CanaryGate bool `yaml:"canary_gate"`

	// orphan은 참조 파일이 디스크에 존재하지 않을 때 loader가 true로 설정한다.
	// unexported이므로 YAML에 직렬화되지 않으며 Go reflect로도 exported로 보이지 않는다.
	orphan bool
}

// Orphan은 참조 파일이 디스크에 존재하지 않는지 여부를 반환한다.
// loader가 LoadRegistry 시 파일 존재 확인 후 설정한다.
func (r Rule) Orphan() bool {
	return r.orphan
}

// withOrphan은 orphan 플래그를 설정한 새 Rule을 반환하는 내부 헬퍼이다.
func (r Rule) withOrphan(orphan bool) Rule {
	r.orphan = orphan
	return r
}

// Validate는 Rule의 필수 필드를 검증한다.
// 빈 필드 또는 잘못된 ID 형식이 있으면 오류를 반환한다.
func (r Rule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("rule ID가 비어 있다")
	}
	if !ruleIDRegexp.MatchString(r.ID) {
		return fmt.Errorf("rule ID %q가 패턴 %q에 맞지 않는다", r.ID, ruleIDPattern)
	}
	if r.File == "" {
		return fmt.Errorf("rule %q: File 필드가 비어 있다", r.ID)
	}
	if r.Clause == "" {
		return fmt.Errorf("rule %q: Clause 필드가 비어 있다", r.ID)
	}
	return nil
}
