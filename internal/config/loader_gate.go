package config

import "log/slog"

// loader_gate.go — gate.yaml section loader.
//
// Completes the gate registry pair: GateConfig and the "gate" registry entry
// existed, but no loader path read gate.yaml, so the ast-grep pre-tool gate
// (AstGrepGate.Enabled) had no path to true. The ast-grep sub-gate remains
// OFF by default (opt-in enable): an absent gate.yaml, or one that omits
// ast_grep_gate.enabled, yields Enabled=false and unchanged gate behavior.

// loadGateSection loads the gate configuration section from gate.yaml.
//
// The wrapper is seeded with the populated default (cfg.Gate) so a gate.yaml
// that declares the section but omits keys retains the construction-time
// defaults instead of collapsing them to zero (partial-override contract,
// parallel to loadArchiveSection / loadHandoffSection).
func (l *Loader) loadGateSection(dir string, cfg *Config) {
	wrapper := &gateFileWrapper{Gate: cfg.Gate}
	loaded, err := loadYAMLFile(dir, "gate.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load gate config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Gate = wrapper.Gate
		l.loadedSections["gate"] = true
	}
}
