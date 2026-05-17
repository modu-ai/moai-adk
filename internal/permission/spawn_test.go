package permission

import "testing"

// TestRejectIfStrict_RejectsBypass bypassPermissions + strictMode=true → 오류 반환 검증.
// T-RT002-16, AC-07 관련.
func TestRejectIfStrict_RejectsBypass(t *testing.T) {
	t.Parallel()

	err := RejectIfStrict(ModeBypassPermissions, true)
	if err == nil {
		t.Fatal("RejectIfStrict() should return error for bypassPermissions in strict mode")
	}
	if !containsMiddle(err.Error(), "not allowed in strict mode") {
		t.Errorf("RejectIfStrict() error = %q; want 'not allowed in strict mode'", err.Error())
	}
}

// TestRejectIfStrict_AllowsAcceptEdits acceptEdits + strictMode=true → nil 반환 검증.
// T-RT002-16 관련.
func TestRejectIfStrict_AllowsAcceptEdits(t *testing.T) {
	t.Parallel()

	err := RejectIfStrict(ModeAcceptEdits, true)
	if err != nil {
		t.Errorf("RejectIfStrict() should return nil for acceptEdits in strict mode, got: %v", err)
	}
}

// TestRejectIfStrict_StrictModeFalse strictMode=false → nil 반환 검증.
// T-RT002-16 관련.
func TestRejectIfStrict_StrictModeFalse(t *testing.T) {
	t.Parallel()

	err := RejectIfStrict(ModeBypassPermissions, false)
	if err != nil {
		t.Errorf("RejectIfStrict() should return nil when strictMode=false, got: %v", err)
	}
}
