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
	if cfg.RulesDir != ".moai/config/astgrep-rules" {
		t.Errorf("RulesDir: want .moai/config/astgrep-rules, got %q", cfg.RulesDir)
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
	if cfg.AstGrepGate.RulesDir != ".moai/config/astgrep-rules" {
		t.Errorf("AstGrepGate.RulesDir: want .moai/config/astgrep-rules, got %q", cfg.AstGrepGate.RulesDir)
	}
}
