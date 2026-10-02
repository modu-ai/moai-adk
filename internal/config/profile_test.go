package config

// profile_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M1: tests for the re-ported
// llm.profile closed set, its effective-default resolution, and the
// llm.agent_overrides validation (SPEC-MODEL-PROFILE-MATRIX-001 machinery,
// REQ-AFR-002/006 surface). The legacy performance_tier read-time alias leg of
// the old EffectiveProfile is NOT re-ported (plan §D.4-3 — the alias key stays
// retired), so the resolution chain here is profile → default medium only.

import (
	"strings"
	"testing"
)

func TestEffectiveProfile(t *testing.T) {
	cases := []struct {
		name    string
		profile string
		want    string
	}{
		{"empty resolves the default", "", DefaultProfile},
		{"high passes through", "high", "high"},
		{"medium passes through", "medium", "medium"},
		{"low passes through", "low", "low"},
		{"legacy max alias folds to high", "max", "high"},
		{"surrounding space is trimmed", "  high  ", "high"},
		{"an out-of-set value is returned verbatim for the caller to reject", "bogus", "bogus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LLMConfig{Profile: tc.profile}.EffectiveProfile()
			if got != tc.want {
				t.Errorf("EffectiveProfile(%q) = %q, want %q", tc.profile, got, tc.want)
			}
		})
	}
}

func TestProfileClosedSet(t *testing.T) {
	if !IsValidProfile("high") || !IsValidProfile("medium") || !IsValidProfile("low") {
		t.Error("the canonical closed set {high, medium, low} must validate")
	}
	if !IsValidProfile("max") {
		t.Error("the legacy max alias must stay readable (never offered, never written by this SPEC's selector — but readable)")
	}
	if IsValidProfile("") {
		t.Error("empty is not a member of the closed set")
	}
	if IsValidProfile("bogus") {
		t.Error("an out-of-set value must not validate")
	}
	got := ValidProfiles()
	if len(got) != 3 || got[0] != ProfileHigh || got[1] != ProfileMedium || got[2] != ProfileLow {
		t.Errorf("ValidProfiles = %v, want [high medium low] (the legacy max alias is never offered)", got)
	}
}

func TestValidateProfileRule(t *testing.T) {
	if errs := validateProfile(&Config{LLM: LLMConfig{Profile: "high"}}); len(errs) != 0 {
		t.Errorf("a canonical profile must validate: %+v", errs)
	}
	if errs := validateProfile(&Config{LLM: LLMConfig{Profile: ""}}); len(errs) != 0 {
		t.Errorf("empty = the effective default, not an error: %+v", errs)
	}
	errs := validateProfile(&Config{LLM: LLMConfig{Profile: "bogus"}})
	if len(errs) != 1 {
		t.Fatalf("out-of-set profile errors = %d, want 1: %+v", len(errs), errs)
	}
	if errs[0].Field != "llm.profile" {
		t.Errorf("error field = %q, want llm.profile", errs[0].Field)
	}
}

func TestValidateAgentOverridesRule(t *testing.T) {
	valid := Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
		"manager-develop": {Model: "opus", Effort: "xhigh"},
		"manager-todo":    {Model: "haiku", Effort: "low"},
		"Explore":         {Model: "inherit"},
	}}}
	if errs := validateAgentOverrides(&valid); len(errs) != 0 {
		t.Errorf("valid overrides rejected: %+v", errs)
	}

	t.Run("unknown agent rejected", func(t *testing.T) {
		cfg := Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
			"mission-governor": {Model: "opus", Effort: "high"},
		}}}
		errs := validateAgentOverrides(&cfg)
		if len(errs) != 1 || !strings.Contains(errs[0].Field, "mission-governor") {
			t.Errorf("= %+v, want one error naming the agent", errs)
		}
	})
	t.Run("out-of-set model rejected", func(t *testing.T) {
		cfg := Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
			"manager-develop": {Model: "fancy-model", Effort: "high"},
		}}}
		errs := validateAgentOverrides(&cfg)
		if len(errs) != 1 || !strings.Contains(errs[0].Field, ".model") {
			t.Errorf("= %+v, want one .model error", errs)
		}
	})
	t.Run("out-of-set effort rejected", func(t *testing.T) {
		cfg := Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{
			"manager-develop": {Model: "opus", Effort: "absurd"},
		}}}
		errs := validateAgentOverrides(&cfg)
		if len(errs) != 1 || !strings.Contains(errs[0].Field, ".effort") {
			t.Errorf("= %+v, want one .effort error", errs)
		}
	})
	t.Run("nil and empty maps valid", func(t *testing.T) {
		if errs := validateAgentOverrides(&Config{}); len(errs) != 0 {
			t.Errorf("nil map rejected: %+v", errs)
		}
		if errs := validateAgentOverrides(&Config{LLM: LLMConfig{AgentOverrides: map[string]ModelEffort{}}}); len(errs) != 0 {
			t.Errorf("empty map rejected: %+v", errs)
		}
	})
}

// TestValidateWiresProfileRules is the integration half: the top-level
// Validate flow must route the two new rules, so an invalid llm.profile or a
// bad llm.agent_overrides entry reaches the console's atomic-reject set.
func TestValidateWiresProfileRules(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.LLM.Profile = "bogus"
	err := Validate(cfg, map[string]bool{"llm": true})
	if err == nil {
		t.Fatal("Validate accepted an out-of-set llm.profile")
	}
	verr, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("error type = %T, want *ValidationErrors", err)
	}
	found := false
	for _, e := range verr.Errors {
		if e.Field == "llm.profile" {
			found = true
		}
	}
	if !found {
		t.Errorf("validation errors lack the llm.profile row: %+v", verr.Errors)
	}
}
