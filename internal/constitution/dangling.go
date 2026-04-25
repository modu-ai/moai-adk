package constitution

import "fmt"

// ValidateRuleReferences는 refs 슬라이스의 각 Rule ID가 registry에 존재하는지 검증한다.
// 존재하지 않는 ID에 대해 경고 문자열을 반환한다.
//
// REQ-CON-001-041 구현. SPEC-V3R2-SPC-003에서 CLI wiring과 SPEC frontmatter scanning이 추가될 예정.
//
// @MX:NOTE: [AUTO] SPEC-V3R2-SPC-003에서 CLI wiring 추가 예정
// @MX:SPEC: SPEC-V3R2-CON-001 REQ-CON-001-041
func ValidateRuleReferences(registry *Registry, refs []string) []string {
	var warnings []string
	for _, ref := range refs {
		if _, ok := registry.Get(ref); !ok {
			warnings = append(warnings,
				fmt.Sprintf("dangling reference: %s not found in registry", ref))
		}
	}
	return warnings
}
