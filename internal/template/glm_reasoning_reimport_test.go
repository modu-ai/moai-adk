package template

// glm_reasoning_reimport_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M1: tests for
// the re-ported per-agent GLM reasoning resolution (REQ-AFR-010's
// single-derivation anchor, plan §D.4-7). The original tests died with
// ResolveGLMReasoning under SPEC-AGENT-MODEL-INHERIT-001; these re-home the
// same contracts: the coding-max override set, the model-unaware resolution,
// and the model-aware flash pinning that the web console column must reuse
// verbatim (no second derivation).

import "testing"

func TestIsGLMCodingMaxOverrideAgent(t *testing.T) {
	if !IsGLMCodingMaxOverrideAgent("manager-develop") {
		t.Error("manager-develop must stay in the coding-max override set")
	}
	for _, name := range []string{"manager-spec", "builder-harness", "Explore", ""} {
		if IsGLMCodingMaxOverrideAgent(name) {
			t.Errorf("IsGLMCodingMaxOverrideAgent(%q) = true; the set is the singleton {manager-develop}", name)
		}
	}
	if got := GLMCodingMaxOverrideAgents(); len(got) != 1 || got[0] != "manager-develop" {
		t.Errorf("GLMCodingMaxOverrideAgents = %v, want the singleton [manager-develop]", got)
	}
}

func TestResolveGLMReasoning(t *testing.T) {
	cases := []struct {
		name   string
		agent  string
		effort string
		want   string
	}{
		{"coding-max override lifts regardless of effort", "manager-develop", "low", GLMStateMax},
		{"coding-max override lifts at medium", "manager-develop", "medium", GLMStateMax},
		{"low stays low", "manager-spec", "low", GLMStateLow},
		{"medium collapses to max", "manager-spec", "medium", GLMStateMax},
		{"high collapses to max", "manager-spec", "high", GLMStateMax},
		{"max collapses to max", "manager-spec", "max", GLMStateMax},
		{"unrecognized effort hits the totality clause (max)", "manager-spec", "bogus", GLMStateMax},
		{"empty effort hits the totality clause (max)", "manager-spec", "", GLMStateMax},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveGLMReasoning(tc.agent, tc.effort).Name; got != tc.want {
				t.Errorf("ResolveGLMReasoning(%q, %q) = %q, want %q", tc.agent, tc.effort, got, tc.want)
			}
		})
	}
}

func TestResolveGLMReasoningForModel(t *testing.T) {
	flash := "glm-5.3-flash"
	nonFlash := "glm-5.3"
	cases := []struct {
		name   string
		model  string
		agent  string
		effort string
		want   string
	}{
		{"flash pins every effort to max — including low", flash, "manager-spec", "low", GLMStateMax},
		{"flash pins the coding-max agent to max", flash, "manager-develop", "low", GLMStateMax},
		{"non-flash delegates to the collapse (low)", nonFlash, "manager-spec", "low", GLMStateLow},
		{"non-flash delegates to the collapse (medium)", nonFlash, "manager-spec", "medium", GLMStateMax},
		{"non-flash keeps the coding-max override", nonFlash, "manager-develop", "low", GLMStateMax},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveGLMReasoningForModel(tc.model, tc.agent, tc.effort).Name; got != tc.want {
				t.Errorf("ResolveGLMReasoningForModel(%q, %q, %q) = %q, want %q", tc.model, tc.agent, tc.effort, got, tc.want)
			}
		})
	}
}
