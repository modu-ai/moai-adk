package config

// loader_crosssession.go — the crosssession.yaml section loader.
//
// The section holds the user's cross-session messaging preferences
// (crossSessionInbound / isolatePeerMachines / dialogExpiry on the Claude Code
// side). It is a partial direct-read: consumed on demand by the moai
// launchers (internal/cli cc/glm/cg → transient --settings injection), not by
// the Loader.Load chain — the same pattern as lsp/mx/security, and registered
// in yamlToStructRegistry accordingly.
//
// Default posture (matches the template-shipped crosssession.yaml):
//   - inbound unset — Claude Code decides per message from the two sessions'
//     permission-mode classes (the accept < hold < refuse ladder stays live);
//   - isolate_machines false — cross-machine messages require NO approval.
//     A true from any Claude Code settings scope applies and cannot be turned
//     off from a lower scope, so the launcher only ever emits it on the
//     user's explicit opt-in here (guarded by
//     internal/cli TestLauncherNeverInjectsIsolatePeerMachinesByDefault and
//     TestTemplateNeverShipsIsolatePeerMachines);
//   - dialog_expiry unset — Claude Code's five-minute default applies.

import (
	"path/filepath"
)

// LoadCrossSessionConfig reads the crosssession.yaml section file from the
// given sections directory and returns the config with defaults pre-applied
// (partial-override contract, parallel to loadHandoffSection: a yaml
// specifying a subset of keys retains the default for the omitted keys).
// A missing file yields the defaults with no error — a project that never
// configured cross-session messaging keeps the neutral posture.
func LoadCrossSessionConfig(sectionsDir string) (CrossSessionConfig, error) {
	cfg := NewDefaultCrossSessionConfig()
	wrapper := &crossSessionFileWrapper{CrossSession: cfg}

	loaded, err := loadYAMLFile(filepath.Clean(sectionsDir), "crosssession.yaml", wrapper)
	if err != nil {
		return cfg, err
	}
	if loaded {
		cfg = wrapper.CrossSession
	}
	return cfg, nil
}

// crossSessionFileWrapper handles the crosssession.yaml section file.
type crossSessionFileWrapper struct {
	CrossSession CrossSessionConfig `yaml:"crosssession"`
}

// ValidCrossSessionInboundValues returns the closed value set of
// CrossSessionConfig.Inbound — the accept|hold|refuse ladder of Claude Code's
// crossSessionInbound. Both consumers (the launcher translation in
// internal/cli and the web console select in internal/settings) derive their
// membership checks and option lists from here — no second declaration.
func ValidCrossSessionInboundValues() []string {
	return []string{"accept", "hold", "refuse"}
}

// ValidCrossSessionDialogExpiryValues returns the closed value set of
// CrossSessionConfig.DialogExpiry — Claude Code's dialogExpiry durations
// ("" unset is the default 5m and is NOT a member of this set).
func ValidCrossSessionDialogExpiryValues() []string {
	return []string{"60s", "5m", "10m", "never"}
}

// IsValidCrossSessionInbound reports membership in the inbound closed set.
func IsValidCrossSessionInbound(v string) bool {
	for _, val := range ValidCrossSessionInboundValues() {
		if v == val {
			return true
		}
	}
	return false
}

// IsValidCrossSessionDialogExpiry reports membership in the dialog-expiry
// closed set.
func IsValidCrossSessionDialogExpiry(v string) bool {
	for _, val := range ValidCrossSessionDialogExpiryValues() {
		if v == val {
			return true
		}
	}
	return false
}
