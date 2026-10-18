package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestInitialProfileResult_ShippedLaunchDefaults pins the CLAUDE-settings values
// the profile wizard pre-selects for a profile that has stored nothing (card
// t1866, operator decision 2026-10-18): model the 1M Sonnet picker value, and
// effort_level medium. The literals are deliberate: the constant the wizard
// reads must not be able to make this test pass by changing with it.
func TestInitialProfileResult_ShippedLaunchDefaults(t *testing.T) {
	got := initialProfileResult(profile.ProfilePreferences{})
	if got.Model != "sonnet[1m]" {
		t.Errorf("Model = %q, want sonnet[1m] (the picker value the web labels Sonnet 5.5)", got.Model)
	}
	if got.EffortLevel != "medium" {
		t.Errorf("EffortLevel = %q, want medium", got.EffortLevel)
	}
}

// TestInitialProfileResult_StoredLaunchChoicesWin pins the other side: a profile
// that already chose a value keeps it.
func TestInitialProfileResult_StoredLaunchChoicesWin(t *testing.T) {
	got := initialProfileResult(profile.ProfilePreferences{
		PermissionMode: "plan",
		Model:          "opus[1m]",
		EffortLevel:    "high",
	})
	if got.PermissionMode != "plan" {
		t.Errorf("PermissionMode = %q, want the stored plan", got.PermissionMode)
	}
	if got.Model != "opus[1m]" {
		t.Errorf("Model = %q, want the stored opus[1m]", got.Model)
	}
	if got.EffortLevel != "high" {
		t.Errorf("EffortLevel = %q, want the stored high", got.EffortLevel)
	}
}
