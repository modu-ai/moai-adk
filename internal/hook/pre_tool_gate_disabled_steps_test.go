package hook

import "testing"

// Issue #1265 (secondary finding) — gate.disabled_steps had no config path.
//
// quality.GateConfig.DisabledSteps existed and was honoured by the gate runner,
// but nothing ever populated it: config.GateConfig carried no matching field and
// loadGateConfig never mapped one. The knob was documented and dead. These tests
// pin the file->struct->runner wiring end to end.
//
// Note the runner's inverted convention (quality/gate.go): an entry whose value
// is FALSE skips that step. The mapping must carry values through verbatim
// rather than normalising them, or a user's `dotnet format: false` silently
// stops skipping.
func TestPreToolHandler_LoadGateConfig_MapsDisabledSteps(t *testing.T) {
	t.Parallel()

	appCfg := newTestConfig()
	appCfg.Gate.DisabledSteps = map[string]bool{"dotnet format": false, "mypy": true}

	h := &preToolHandler{
		cfg:    &mockConfigProvider{cfg: appCfg},
		policy: DefaultSecurityPolicy(),
	}
	cfg := h.loadGateConfig()
	if cfg == nil {
		t.Fatal("expected non-nil gate config")
	}
	if cfg.DisabledSteps == nil {
		t.Fatal("DisabledSteps was not mapped from config (issue #1265)")
	}
	if v, ok := cfg.DisabledSteps["dotnet format"]; !ok || v {
		t.Errorf(`DisabledSteps["dotnet format"] = (%v, %v), want (false, true)`, v, ok)
	}
	if v, ok := cfg.DisabledSteps["mypy"]; !ok || !v {
		t.Errorf(`DisabledSteps["mypy"] = (%v, %v), want (true, true)`, v, ok)
	}
}

// An absent disabled_steps section must not fabricate a map — the runner treats
// a missing key as "step enabled", so a nil map is the correct empty state.
func TestPreToolHandler_LoadGateConfig_DisabledStepsAbsent(t *testing.T) {
	t.Parallel()

	h := &preToolHandler{
		cfg:    &mockConfigProvider{cfg: newTestConfig()},
		policy: DefaultSecurityPolicy(),
	}
	cfg := h.loadGateConfig()
	if len(cfg.DisabledSteps) != 0 {
		t.Errorf("DisabledSteps = %v, want empty when unset", cfg.DisabledSteps)
	}
}
