package config

import (
	"strings"
	"testing"
)

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

