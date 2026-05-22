package permission

import "fmt"

// RejectIfStrict rejects the bypassPermissions mode at the spawn stage when
// running in strict mode.
//
// Returns an error when strictMode=true and mode=ModeBypassPermissions.
// Returns nil for any other mode or when strictMode=false.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-022, AC-07.
func RejectIfStrict(mode PermissionMode, strictMode bool) error {
	if mode == ModeBypassPermissions && strictMode {
		return fmt.Errorf("permission mode rejected: bypassPermissions not allowed in strict mode")
	}
	return nil
}
