package permission

import "fmt"

// MigrateLegacyBypassRules migrates rules with the legacy v2 bypassPermissions
// action to acceptEdits (DecisionAllow).
//
// When a bypassPermissions action is detected:
//   - The Action is rerouted to DecisionAllow.
//   - A deprecation warning containing the origin file path is returned.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-040, AC-11.
func MigrateLegacyBypassRules(rules []PermissionRule) ([]PermissionRule, []string) {
	migrated := make([]PermissionRule, len(rules))
	var warnings []string

	for i, r := range rules {
		if r.Action == Decision("bypassPermissions") {
			// Legacy action → reroute to acceptEdits.
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
