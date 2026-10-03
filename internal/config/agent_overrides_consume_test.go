package config

// agent_overrides_consume_test.go — REQ-AFR-015 (SPEC-WEB-AGENTFM-RESTORE-001
// v0.3.0 M7, card t1421): the llm.agent_overrides_consume opt-in key.
//
// Three axes the acceptance ledger EV-AFR-014 names:
//  1. the key defaults to false — a session without it keeps today's
//     storage-only behaviour byte-for-byte (REQ-AFR-002 succession);
//  2. an explicit true opts the session's spawns into consuming
//     llm.agent_overrides;
//  3. a non-boolean value joins the existing atomic-reject flow — via
//     ValidateLLMYAMLSection, the strict type check the console save path
//     runs at its write boundary (the lenient section loader silently falls
//     back to defaults on exactly this defect class, so the strict check is
//     the reject surface, not the load path).
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeConsumeFixture writes an llm.yaml body into the real sections layout
// under dir (dir acts as the project root).
func writeConsumeFixture(t *testing.T, dir, body string) {
	t.Helper()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAgentOverridesConsumeKey(t *testing.T) {
	t.Parallel()

	t.Run("default config is false", func(t *testing.T) {
		t.Parallel()
		if NewDefaultConfig().LLM.AgentOverridesConsume {
			t.Error("the shipped default must be false — the opt-in contract keeps key-less sessions on storage-only behaviour")
		}
	})

	t.Run("absent key resolves false", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeConsumeFixture(t, dir, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n  agent_overrides: {}\n")

		cfg, err := NewConfigManager().Load(dir)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.LLM.AgentOverridesConsume {
			t.Error("an llm.yaml without the key must resolve consume = false")
		}
	})

	t.Run("explicit true opts in", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeConsumeFixture(t, dir, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n  agent_overrides_consume: true\n")

		cfg, err := NewConfigManager().Load(dir)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if !cfg.LLM.AgentOverridesConsume {
			t.Error("agent_overrides_consume: true must resolve consume = on")
		}
		if err := ValidateLLMYAMLSection(dir); err != nil {
			t.Errorf("a real boolean must pass the strict check: %v", err)
		}
	})

	t.Run("non-boolean value joins the atomic-reject surface", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeConsumeFixture(t, dir, "llm:\n  mode: \"\"\n  glm_env_var: GLM_API_KEY\n  agent_overrides_consume: \"yes\"\n")

		// The lenient load path degrades to defaults (fail-open by the
		// section-loader contract); the strict check is the reject surface the
		// console save path joins. Both facts are pinned here: the first names
		// WHY the strict check exists, the second is the REQ-AFR-015 contract.
		if cfg, err := NewConfigManager().Load(dir); err != nil {
			t.Fatalf("the lenient loader stays fail-open: Load: %v", err)
		} else if cfg == nil {
			t.Fatal("the lenient loader returns defaults, not nil")
		}

		err := ValidateLLMYAMLSection(dir)
		if err == nil {
			t.Fatal("ValidateLLMYAMLSection must reject the non-boolean value (the atomic-reject join, REQ-AFR-015)")
		}
		if !strings.Contains(err.Error(), "agent_overrides_consume") {
			t.Errorf("the rejection must name the offending key, got: %v", err)
		}
	})

	t.Run("absent llm.yaml is not a rejection", func(t *testing.T) {
		t.Parallel()
		if err := ValidateLLMYAMLSection(t.TempDir()); err != nil {
			t.Errorf("a greenfield project has no llm.yaml to reject: %v", err)
		}
	})
}
