package permission

import "fmt"

// MigrateLegacyBypassRules legacy v2 bypassPermissions action 을 갖는 규칙을
// acceptEdits (DecisionAllow) 으로 마이그레이션한다.
//
// bypassPermissions action 이 발견되면:
//   - Action 을 DecisionAllow 로 reroute 한다.
//   - origin 파일 경로를 포함한 deprecation warning 을 반환한다.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-040, AC-11.
func MigrateLegacyBypassRules(rules []PermissionRule) ([]PermissionRule, []string) {
	migrated := make([]PermissionRule, len(rules))
	var warnings []string

	for i, r := range rules {
		if r.Action == Decision("bypassPermissions") {
			// legacy action → acceptEdits 로 reroute.
			migrated[i] = r
			migrated[i].Action = DecisionAllow
			warnings = append(warnings, fmt.Sprintf(
				"DEPRECATED: rule in %q uses legacy action 'bypassPermissions'; migrated to 'acceptEdits' for pattern %q",
				r.Origin, r.Pattern,
			))
		} else {
			migrated[i] = r
		}
	}

	return migrated, warnings
}
