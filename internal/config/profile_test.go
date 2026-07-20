package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad_LegacyLLMYaml_Migration covers REQ-MPM-003/004 / AC-MPM-002: a legacy
// llm.yaml carrying plan_type + claude_models + performance_tier loads without
// error (unknown keys ignored), and the effective profile resolves via the
// performance_tier alias (max here).
func TestLoad_LegacyLLMYaml_Migration(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := `llm:
    plan_type: "subscription"
    performance_tier: "max"
    claude_models:
        high: "opus"
        medium: "sonnet"
        low: "sonnet"
`
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := NewLoader().Load(root)
	if err != nil {
		t.Fatalf("legacy llm.yaml must load without error, got: %v", err)
	}
	if got := cfg.LLM.EffectiveProfile(); got != "max" {
		t.Errorf("EffectiveProfile() = %q, want max (performance_tier alias)", got)
	}
	// profile: absent + performance_tier: max → profile resolves max; no error.
	if cfg.LLM.Profile != "" {
		t.Errorf("Profile should be empty (absent in legacy config), got %q", cfg.LLM.Profile)
	}
}

// TestLoad_NewSchemaLLMYaml covers AC-MPM-001: a new-schema llm.yaml with
// profile + profiles + agent_overrides loads and resolves correctly.
func TestLoad_NewSchemaLLMYaml(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := `llm:
    profile: low
    agent_overrides:
        manager-develop: { model: opus, effort: xhigh }
`
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := NewLoader().Load(root)
	if err != nil {
		t.Fatalf("new-schema llm.yaml must load, got: %v", err)
	}
	if cfg.LLM.EffectiveProfile() != "low" {
		t.Errorf("EffectiveProfile() = %q, want low", cfg.LLM.EffectiveProfile())
	}
	ov, ok := cfg.LLM.AgentOverrides["manager-develop"]
	if !ok || ov.Model != "opus" || ov.Effort != "xhigh" {
		t.Errorf("agent_overrides not parsed: %+v ok=%v", ov, ok)
	}
}

// TestEffectiveProfile covers REQ-MPM-002 / AC-MPM-001 / AC-MPM-002: profile
// pass-through, the performance_tier legacy alias (high→max), and the medium
// default.
func TestEffectiveProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		perfT   string
		want    string
	}{
		{"explicit low", "low", "", "low"},
		{"explicit max", "max", "", "max"},
		{"explicit medium", "medium", "", "medium"},
		{"profile wins over perf_tier", "low", "max", "low"},
		{"legacy perf_tier high -> max", "", "high", "max"},
		{"legacy perf_tier max pass-through", "", "max", "max"},
		{"legacy perf_tier low pass-through", "", "low", "low"},
		{"both absent -> medium default", "", "", "medium"},
		{"whitespace profile -> alias", "   ", "high", "max"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LLMConfig{Profile: tt.profile, PerformanceTier: tt.perfT}
			if got := l.EffectiveProfile(); got != tt.want {
				t.Fatalf("EffectiveProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestValidateProfile covers REQ-MPM-008 / AC-MPM-001: closed-set validation
// with the offending value + set named, and empty being valid.
func TestValidateProfile(t *testing.T) {
	if errs := validateProfile(&Config{LLM: LLMConfig{Profile: ""}}); len(errs) != 0 {
		t.Fatalf("empty profile should be valid, got %v", errs)
	}
	for _, v := range []string{"max", "medium", "low"} {
		if errs := validateProfile(&Config{LLM: LLMConfig{Profile: v}}); len(errs) != 0 {
			t.Fatalf("profile %q should be valid, got %v", v, errs)
		}
	}
	errs := validateProfile(&Config{LLM: LLMConfig{Profile: "bogus"}})
	if len(errs) != 1 {
		t.Fatalf("bogus profile should produce 1 error, got %d", len(errs))
	}
	if errs[0].Field != "llm.profile" || errs[0].Value != "bogus" {
		t.Fatalf("error should name field+value, got %+v", errs[0])
	}
	if !strings.Contains(errs[0].Message, "bogus") || !strings.Contains(errs[0].Message, "max, medium, low") {
		t.Fatalf("error message should name value + closed set, got %q", errs[0].Message)
	}
}

// TestValidateAgentOverrides covers REQ-MPM-007 / AC-MPM-004: valid entry
// passes; non-catalog agent, out-of-enum model, and out-of-enum effort each
// error with the offending agent/field named.
func TestValidateAgentOverrides(t *testing.T) {
	valid := &Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
		"manager-develop": {Model: "opus", Effort: "xhigh"},
	}}}
	if errs := validateAgentOverrides(valid); len(errs) != 0 {
		t.Fatalf("valid override should pass, got %v", errs)
	}

	badAgent := &Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
		"not-an-agent": {Model: "opus", Effort: "high"},
	}}}
	errs := validateAgentOverrides(badAgent)
	if len(errs) != 1 || errs[0].Value != "not-an-agent" {
		t.Fatalf("non-catalog agent should error naming the agent, got %v", errs)
	}

	badModel := &Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
		"manager-spec": {Model: "gpt4", Effort: "high"},
	}}}
	errs = validateAgentOverrides(badModel)
	if len(errs) != 1 || errs[0].Field != "llm.agent_overrides.manager-spec.model" {
		t.Fatalf("out-of-enum model should error on the model field, got %v", errs)
	}

	badEffort := &Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
		"manager-spec": {Model: "opus", Effort: "turbo"},
	}}}
	errs = validateAgentOverrides(badEffort)
	if len(errs) != 1 || errs[0].Field != "llm.agent_overrides.manager-spec.effort" {
		t.Fatalf("out-of-enum effort should error on the effort field, got %v", errs)
	}
}

