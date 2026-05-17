package permission

import "fmt"

// RejectIfStrict bypassPermissions 모드를 strict mode 에서 spawn 단계에서 거부한다.
//
// strictMode=true 이고 mode=ModeBypassPermissions 인 경우 오류를 반환한다.
// 다른 mode 이거나 strictMode=false 인 경우 nil 을 반환한다.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-022, AC-07.
func RejectIfStrict(mode PermissionMode, strictMode bool) error {
	if mode == ModeBypassPermissions && strictMode {
		return fmt.Errorf("permission mode rejected: bypassPermissions not allowed in strict mode")
	}
	return nil
}
