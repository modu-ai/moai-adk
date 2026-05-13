package permission

import (
	"fmt"
)

// RejectIfStrict checks if a permission mode should be rejected in strict mode.
// This is called at spawn time to prevent agents with disallowed modes from starting.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-022
func RejectIfStrict(mode PermissionMode, strictMode bool) error {
	if mode == ModeBypassPermissions && strictMode {
		return fmt.Errorf("%w: bypassPermissions not allowed in strict mode", ErrPermissionModeRejected)
	}
	return nil
}
