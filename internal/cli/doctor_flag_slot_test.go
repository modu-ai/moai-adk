package cli

// doctor_flag_slot_test.go — card t702.
//
// Tests for the Shared Flag Slot doctor check: the machine-global
// cachedGrowthBookFeatures.tengu_harbor_kite slot in ~/.claude.json that gates
// the cross-session messaging channel. Only a first-party session
// (ANTHROPIC_BASE_URL = api.anthropic.com) writes the slot; a third-party
// backend session inherits whatever a first-party session last left, which is
// the state this check surfaces (cross-session-messaging-detail.md § The
// shared flag slot).

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

// overrideFlagSlotFixture replaces the ~/.claude.json read seam for the
// duration of the test.
func overrideFlagSlotFixture(t *testing.T, data []byte, err error) {
	t.Helper()
	orig := doctorFlagSlotClaudeJSON
	doctorFlagSlotClaudeJSON = func() ([]byte, error) { return data, err }
	t.Cleanup(func() { doctorFlagSlotClaudeJSON = orig })
}

func fixtureClaudeJSON(slotJSON string) []byte {
	return []byte(`{"cachedGrowthBookFeatures":{"tengu_harbor_kite":` + slotJSON + `}}`)
}

func TestCheckFlagSlot_FirstParty_NoBaseURL(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "")
	overrideFlagSlotFixture(t, nil, errors.New("must not be read in first-party mode"))

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %v, want OK (unset ANTHROPIC_BASE_URL is first-party)", check.Status)
	}
	if !strings.Contains(check.Message, "first-party") {
		t.Errorf("message %q should name the first-party session type", check.Message)
	}
}

func TestCheckFlagSlot_FirstParty_AnthropicHost(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://api.anthropic.com")
	overrideFlagSlotFixture(t, nil, errors.New("must not be read in first-party mode"))

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %v, want OK (api.anthropic.com is first-party)", check.Status)
	}
}

func TestCheckFlagSlot_FirstParty_AnthropicHostTrailingSlash(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://api.anthropic.com/")
	overrideFlagSlotFixture(t, nil, errors.New("must not be read in first-party mode"))

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %v, want OK (trailing slash keeps the host first-party)", check.Status)
	}
}

func TestCheckFlagSlot_ThirdParty_SlotTrue(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, fixtureClaudeJSON("true"), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %v, want OK (slot true means the channel is currently on)", check.Status)
	}
	if !strings.Contains(check.Message, "true") {
		t.Errorf("message %q should report the slot value", check.Message)
	}
	if !strings.Contains(check.Message, "last-writer") {
		t.Errorf("message %q should carry the last-writer-wins caveat", check.Message)
	}
}

func TestCheckFlagSlot_ThirdParty_SlotFalse(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, fixtureClaudeJSON("false"), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want Warn (slot false means the channel is off for this session)", check.Status)
	}
	if !strings.Contains(check.Message, "false") {
		t.Errorf("message %q should report the slot value", check.Message)
	}
	if !strings.Contains(check.Message, config.EnvClaudeCodeHarborKite) {
		t.Errorf("message %q should name the documented manual escape hatch", check.Message)
	}
}

func TestCheckFlagSlot_ThirdParty_SlotAbsent(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, []byte(`{"cachedGrowthBookFeatures":{}}`), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want Warn (slot never written is an unknown-off state)", check.Status)
	}
	if !strings.Contains(check.Message, "tengu_harbor_kite") {
		t.Errorf("message %q should name the slot key", check.Message)
	}
}

func TestCheckFlagSlot_ThirdParty_FeaturesObjectAbsent(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, []byte(`{"numStartups":4}`), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want Warn (no features object means no slot)", check.Status)
	}
}

func TestCheckFlagSlot_ThirdParty_ReadError(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, nil, errors.New("permission denied"))

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want Warn (unreadable file leaves the state unknown)", check.Status)
	}
	if !strings.Contains(check.Message, "~/.claude.json") {
		t.Errorf("message %q should name the file it could not read", check.Message)
	}
}

func TestCheckFlagSlot_ThirdParty_MalformedJSON(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, []byte("{not json"), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want Warn (malformed JSON leaves the state unknown)", check.Status)
	}
}

func TestCheckFlagSlot_ThirdParty_NonBoolSlot(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	overrideFlagSlotFixture(t, fixtureClaudeJSON(`"yes"`), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want Warn (non-boolean slot is an unrecognized state)", check.Status)
	}
}

func TestCheckFlagSlot_ThirdParty_OverrideEnvForcesOn(t *testing.T) {
	t.Setenv(config.EnvAnthropicBaseURL, "https://z.ai/api")
	t.Setenv(config.EnvClaudeCodeHarborKite, "1")
	overrideFlagSlotFixture(t, fixtureClaudeJSON("false"), nil)

	check := checkFlagSlot(false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %v, want OK (the gate honors the override before the slot)", check.Status)
	}
	if !strings.Contains(check.Message, config.EnvClaudeCodeHarborKite) {
		t.Errorf("message %q should report the active override", check.Message)
	}
}

func TestCheckFlagSlot_RegisteredInSystemChecks(t *testing.T) {
	results := runDiagnosticChecks(false, "Shared Flag Slot")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'Shared Flag Slot' filter, got %d", len(results))
	}
	if results[0].Name != "Shared Flag Slot" {
		t.Errorf("results[0].Name = %q, want 'Shared Flag Slot'", results[0].Name)
	}
}

func TestCheckFlagSlot_PresentInAllChecks(t *testing.T) {
	checks := runDiagnosticChecks(false, "")
	found := false
	for _, c := range checks {
		if c.Name == "Shared Flag Slot" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Shared Flag Slot check should be registered in runDiagnosticChecks")
	}
}
