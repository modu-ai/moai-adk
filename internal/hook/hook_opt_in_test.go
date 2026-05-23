// Package hook — hook_opt_in_test.go
// SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 — REQ-HOI-002 / REQ-HOI-003 / AC-HOI-007
//
// Tests the HOI master toggle gate (`hook.opt_in.enabled` in system.yaml) and
// the 4-quadrant cohabitation invariant with REQ-OBS-005 (observability.enabled,
// observability.yaml) and SPEC-V3R2-RT-006 REQ-040 (hook.observability_events).
package hook

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// newTestConfigWithHookOptIn returns a ConfigProvider with hook.opt_in.enabled
// set to the given value. Used by TestHookOptIn* tests to inject runtime state.
func newTestConfigWithHookOptIn(enabled bool) ConfigProvider {
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.OptIn.Enabled = enabled
	return &auditConfigProvider{cfg: cfg}
}

// newTestConfigWithHookOptInAndEvents returns a ConfigProvider with both the
// HOI master toggle and the RT-006 per-event whitelist set independently.
// Used by TestHookOptInCohabitation to exercise the 3-key independence contract.
func newTestConfigWithHookOptInAndEvents(optInEnabled bool, events []string) ConfigProvider {
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.OptIn.Enabled = optInEnabled
	cfg.System.Hook.ObservabilityEvents = events
	return &auditConfigProvider{cfg: cfg}
}

// TestHookOptInDisabled verifies AC-HOI-003: when hook.opt_in.enabled is false
// (default), the HOI gate returns false regardless of the RT-006 whitelist state.
// Maps to spec.md REQ-HOI-002.
func TestHookOptInDisabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  ConfigProvider
		want bool
	}{
		{
			name: "default zero-value config — opt-in disabled",
			cfg:  &auditConfigProvider{cfg: config.NewDefaultConfig()},
			want: false,
		},
		{
			name: "explicit opt_in.enabled: false",
			cfg:  newTestConfigWithHookOptIn(false),
			want: false,
		},
		{
			name: "nil ConfigProvider",
			cfg:  nil,
			want: false,
		},
		{
			name: "ConfigProvider returning nil",
			cfg:  &auditConfigProvider{cfg: nil},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hookOptInEnabled(tt.cfg)
			if got != tt.want {
				t.Errorf("hookOptInEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHookOptInEnabled verifies AC-HOI-004: when hook.opt_in.enabled is true,
// the HOI gate returns true. Maps to spec.md REQ-HOI-003.
func TestHookOptInEnabled(t *testing.T) {
	t.Parallel()

	cfg := newTestConfigWithHookOptIn(true)
	got := hookOptInEnabled(cfg)
	if !got {
		t.Errorf("hookOptInEnabled() = false, want true when opt_in.enabled=true")
	}
}

// TestHookOptInMissingKey_DefaultsDisabled verifies the edge case (acceptance.md
// Edge case 1) where a legacy project lacks the hook.opt_in sub-block entirely.
// The Go zero-value default (false) MUST apply. R3 mitigation per plan.md.
func TestHookOptInMissingKey_DefaultsDisabled(t *testing.T) {
	t.Parallel()

	// Simulate a legacy project: SystemHookConfig with ObservabilityEvents and
	// StrictMode set but OptIn never touched (zero-value).
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.ObservabilityEvents = []string{"notification"} // RT-006 enabled
	cfg.System.Hook.StrictMode = true
	// cfg.System.Hook.OptIn is zero-value HookOptInConfig{Enabled: false}

	provider := &auditConfigProvider{cfg: cfg}
	if hookOptInEnabled(provider) {
		t.Errorf("hookOptInEnabled() = true on legacy config without opt_in sub-block; want false (Go zero-value default)")
	}
}

// TestHookOptInCohabitation verifies AC-HOI-007: the 3-key independence
// contract. Exercises all 4 quadrants of (HOI ∈ {false, true}) × (RT-006
// whitelist ∈ {empty, non-empty}) and confirms the two gates are independent.
//
// Maps to spec.md §A.3 cohabitation contract and REQ-HOI-001 + cross-reference
// to SPEC-V3R2-RT-006 REQ-040. Observability.Enabled (REQ-OBS-005) lives in a
// separate file and is not exercised here — its config field is untouched.
func TestHookOptInCohabitation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		hookOptInEnabled  bool
		rt006Whitelist    []string
		wantHookOptIn     bool
		wantRT006Notif    bool // observabilityOptIn for "notification"
		wantRT006TaskCrtd bool // observabilityOptIn for "taskCreated"
	}{
		{
			name:              "Q1: HOI=false, RT-006 empty",
			hookOptInEnabled:  false,
			rt006Whitelist:    []string{},
			wantHookOptIn:     false,
			wantRT006Notif:    false,
			wantRT006TaskCrtd: false,
		},
		{
			name:              "Q2: HOI=false, RT-006 non-empty",
			hookOptInEnabled:  false,
			rt006Whitelist:    []string{"notification", "taskCreated"},
			wantHookOptIn:     false,
			wantRT006Notif:    true, // RT-006 gate independent of HOI master toggle
			wantRT006TaskCrtd: true,
		},
		{
			name:              "Q3: HOI=true, RT-006 empty",
			hookOptInEnabled:  true,
			rt006Whitelist:    []string{},
			wantHookOptIn:     true,
			wantRT006Notif:    false, // RT-006 still off when whitelist empty
			wantRT006TaskCrtd: false,
		},
		{
			name:              "Q4: HOI=true, RT-006 non-empty (full enable)",
			hookOptInEnabled:  true,
			rt006Whitelist:    []string{"notification", "taskCreated"},
			wantHookOptIn:     true,
			wantRT006Notif:    true,
			wantRT006TaskCrtd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfigWithHookOptInAndEvents(tt.hookOptInEnabled, tt.rt006Whitelist)

			if got := hookOptInEnabled(cfg); got != tt.wantHookOptIn {
				t.Errorf("hookOptInEnabled() = %v, want %v", got, tt.wantHookOptIn)
			}
			if got := observabilityOptIn(cfg, "notification"); got != tt.wantRT006Notif {
				t.Errorf("observabilityOptIn(notification) = %v, want %v", got, tt.wantRT006Notif)
			}
			if got := observabilityOptIn(cfg, "taskCreated"); got != tt.wantRT006TaskCrtd {
				t.Errorf("observabilityOptIn(taskCreated) = %v, want %v", got, tt.wantRT006TaskCrtd)
			}
		})
	}
}

// TestHookOptInIndependence_RT006Whitelist verifies the cohabitation contract's
// negative direction: enabling RT-006 whitelist MUST NOT enable the HOI master
// toggle, and disabling HOI MUST NOT disable the RT-006 whitelist.
// This is the regression guard against accidental gate unification (R5 mitigation).
func TestHookOptInIndependence_RT006Whitelist(t *testing.T) {
	t.Parallel()

	// RT-006 whitelist populated; HOI explicitly false.
	cfg := newTestConfigWithHookOptInAndEvents(false, []string{"notification", "elicitation"})

	if hookOptInEnabled(cfg) {
		t.Errorf("hookOptInEnabled() = true when only RT-006 whitelist is set; gates must be independent")
	}
	if !observabilityOptIn(cfg, "notification") {
		t.Errorf("observabilityOptIn(notification) = false when whitelist contains it; HOI=false MUST NOT mask RT-006")
	}
}
