package permission

import "testing"

// TestRejectIfStrict_RejectsBypass verifies that bypassPermissions + strictMode=true returns an error.
// Related to T-RT002-16 and AC-07.
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

// TestRejectIfStrict_AllowsAcceptEdits verifies that acceptEdits + strictMode=true returns nil.
// Related to T-RT002-16.
func TestRejectIfStrict_AllowsAcceptEdits(t *testing.T) {
	t.Parallel()

	err := RejectIfStrict(ModeAcceptEdits, true)
	if err != nil {
		t.Errorf("RejectIfStrict() should return nil for acceptEdits in strict mode, got: %v", err)
	}
}

// TestRejectIfStrict_StrictModeFalse verifies that strictMode=false returns nil.
// Related to T-RT002-16.
func TestRejectIfStrict_StrictModeFalse(t *testing.T) {
	t.Parallel()

	err := RejectIfStrict(ModeBypassPermissions, false)
	if err != nil {
		t.Errorf("RejectIfStrict() should return nil when strictMode=false, got: %v", err)
	}
}
