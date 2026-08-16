package quality

import (
	"testing"
)

func TestDefaultAstGrepGateConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultAstGrepGateConfig()

	if !cfg.Enabled {
		t.Error("Enabled: want true, got false")
	}
	// t50: RulesDir carries no code-level default. Where rules live is
	// per-project configuration (gate.yaml ast_grep_gate.rules_dir); the
	// template ships that key with an explicit value, so an empty RulesDir
	// here means "unconfigured", not "fall back to a guessed path".
	if cfg.RulesDir != "" {
		t.Errorf("RulesDir: want empty (gate.yaml is the SSOT), got %q", cfg.RulesDir)
	}
	if !cfg.BlockOnError {
		t.Error("BlockOnError: want true, got false")
	}
	if cfg.WarnOnlyMode {
		t.Error("WarnOnlyMode: want false, got true")
	}
}

func TestAstGrepGate_GateConfigIntegration(t *testing.T) {
	t.Parallel()

	// Verify that DefaultGateConfig sets the AstGrepGate field
	cfg := DefaultGateConfig()
	if cfg.AstGrepGate == nil {
		t.Fatal("DefaultGateConfig should set AstGrepGate")
	}
	if !cfg.AstGrepGate.Enabled {
		t.Error("AstGrepGate.Enabled should be true by default")
	}
	if cfg.AstGrepGate.RulesDir != "" {
		t.Errorf("AstGrepGate.RulesDir: want empty (gate.yaml is the SSOT), got %q", cfg.AstGrepGate.RulesDir)
	}
}
