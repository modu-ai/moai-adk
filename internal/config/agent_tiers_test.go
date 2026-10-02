package config

import (
	"os"
	"path/filepath"
	"testing"
)

// agent_tiers_test.go — the tier-surface validation and mapping tests
// (SPEC-AGENT-TIER-001 M2 / AC-TIER-006 / AC-TIER-007). Authored RED-first:
// at M2 start no tier surface exists in the package, so this file failed to
// COMPILE naming the undefined symbols (Evidence cell R4).

// loadTiersFixture writes a workflow.yaml carrying the given agent_tiers
// block body into a fresh config dir and runs the loader.
func loadTiersFixture(t *testing.T, tiersYAML string) error {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n    token_budget:\n        plan: 30000\n"
	if tiersYAML != "" {
		body += "    agent_tiers:\n" + tiersYAML
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	l := &Loader{}
	_, err := l.Load(root)
	return err
}

// TestAgentTiers_RejectsUnknownToken pins REQ-TIER-009 / AC-TIER-006: a tier
// assignment whose token is outside the closed set fails the load with an
// error naming the offending token and its file; each closed-set token loads
// cleanly.
func TestAgentTiers_RejectsUnknownToken(t *testing.T) {
	err := loadTiersFixture(t, "        classes:\n            manager-develop: extreme\n")
	if err == nil {
		t.Fatal("load with tier token `extreme` succeeded — want rejection naming the token and workflow.yaml")
	}
	if !containsAll(err.Error(), []string{"extreme", "workflow.yaml"}) {
		t.Errorf("rejection error = %q — want it to name the offending token and the file", err.Error())
	}

	for _, token := range ValidAgentTiers() {
		if err := loadTiersFixture(t, "        classes:\n            manager-develop: "+token+"\n"); err != nil {
			t.Errorf("load with tier token %q failed: %v — every closed-set token must load cleanly", token, err)
		}
	}
}

// TestDefaultAgentTierAssignment pins the default class→tier assignment
// (REQ-TIER-008 / AC-TIER-007) and the audit-surface exclusion (REQ-TIER-007):
// audit surfaces resolve to NO tier — they ride the workflow.audit pins.
func TestDefaultAgentTierAssignment(t *testing.T) {
	cfg := NewDefaultConfig()
	want := map[string]string{
		"super-advisor":   AgentTierMax,
		"manager-spec":    AgentTierMax,
		"manager-develop": AgentTierMedium,
		"manager-docs":    AgentTierMedium,
		"e2e-tester":      AgentTierMedium,
		"explore":         AgentTierLow,
		"lane":            AgentTierMedium, // the general implementation lane
	}
	for class, tier := range want {
		if got := ResolveAgentClassTier(cfg.Workflow.AgentTiers, class); got != tier {
			t.Errorf("class %q resolves to tier %q, want %q", class, got, tier)
		}
	}
	for _, class := range []string{"plan-auditor", "sync-auditor", "audit-claude", "audit-codex", "audit-glm"} {
		if got := ResolveAgentClassTier(cfg.Workflow.AgentTiers, class); got != "" {
			t.Errorf("audit surface %q resolves to tier %q — audit surfaces are excluded from the tier matrix (must resolve to NO tier)", class, got)
		}
	}
}

// TestAgentTiers_UserOverrideWins pins the project-override direction of
// REQ-TIER-008: a user-configured class entry replaces the built-in default
// for that class.
func TestAgentTiers_UserOverrideWins(t *testing.T) {
	tiers := AgentTiersConfig{Classes: map[string]string{"manager-develop": AgentTierLow}}
	if got := ResolveAgentClassTier(tiers, "manager-develop"); got != AgentTierLow {
		t.Errorf("user override: manager-develop resolves to %q, want %q", got, AgentTierLow)
	}
	// Untouched classes keep the built-in default.
	if got := ResolveAgentClassTier(tiers, "super-advisor"); got != AgentTierMax {
		t.Errorf("user override leaked: super-advisor resolves to %q, want the built-in %q", got, AgentTierMax)
	}
}

// TestAgentTierPair pins the tier→{model, effort} resolution onto the M1
// constants (REQ-TIER-002): each closed-set token resolves to its pair and an
// unknown token resolves nothing.
func TestAgentTierPair(t *testing.T) {
	cases := map[string]ModelEffort{
		AgentTierMax:    DefaultClaudeTierMax,
		AgentTierMedium: DefaultClaudeTierMedium,
		AgentTierLow:    DefaultClaudeTierLow,
	}
	for tier, want := range cases {
		got, ok := AgentTierPair(tier)
		if !ok || got != want {
			t.Errorf("AgentTierPair(%q) = {%s %s}, %v; want {%s %s}, true", tier, got.Model, got.Effort, ok, want.Model, want.Effort)
		}
	}
	if _, ok := AgentTierPair("extreme"); ok {
		t.Error("AgentTierPair(\"extreme\") resolved — an unknown token must resolve nothing")
	}
}

// containsAll reports whether s carries every substring.
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
