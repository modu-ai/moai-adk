package permission

import (
	"testing"
)

// TestRejectIfStrict_RejectsBypass tests that bypassPermissions is rejected in strict mode
func TestRejectIfStrict_RejectsBypass(t *testing.T) {
	err := RejectIfStrict(ModeBypassPermissions, true)
	if err == nil {
		t.Error("RejectIfStrict() should reject bypassPermissions in strict mode")
	}
	// Check if error wraps ErrPermissionModeRejected
	if err == nil || !contains(err.Error(), "not allowed in strict mode") {
		t.Errorf("RejectIfStrict() error = %v, should contain 'not allowed in strict mode'", err)
	}
}

// TestRejectIfStrict_AllowsAcceptEdits tests that acceptEdits is allowed in strict mode
func TestRejectIfStrict_AllowsAcceptEdits(t *testing.T) {
	err := RejectIfStrict(ModeAcceptEdits, true)
	if err != nil {
		t.Errorf("RejectIfStrict() should allow acceptEdits in strict mode, got error: %v", err)
	}
}

// TestRejectIfStrict_StrictModeFalse tests that bypassPermissions is allowed when strictMode is false
func TestRejectIfStrict_StrictModeFalse(t *testing.T) {
	err := RejectIfStrict(ModeBypassPermissions, false)
	if err != nil {
		t.Errorf("RejectIfStrict() should allow bypassPermissions when strictMode is false, got error: %v", err)
	}
}

// TestRejectIfStrict_BubbleAlwaysAllowed tests that bubble mode is always allowed
func TestRejectIfStrict_BubbleAlwaysAllowed(t *testing.T) {
	err := RejectIfStrict(ModeBubble, true)
	if err != nil {
		t.Errorf("RejectIfStrict() should always allow bubble mode, got error: %v", err)
	}
}

// TestRejectIfStrict_DefaultAlwaysAllowed tests that default mode is always allowed
func TestRejectIfStrict_DefaultAlwaysAllowed(t *testing.T) {
	err := RejectIfStrict(ModeDefault, true)
	if err != nil {
		t.Errorf("RejectIfStrict() should always allow default mode, got error: %v", err)
	}
}

// TestRejectIfStrict_PlanAlwaysAllowed tests that plan mode is always allowed
func TestRejectIfStrict_PlanAlwaysAllowed(t *testing.T) {
	err := RejectIfStrict(ModePlan, true)
	if err != nil {
		t.Errorf("RejectIfStrict() should always allow plan mode, got error: %v", err)
	}
}
