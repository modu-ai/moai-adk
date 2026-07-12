package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- SPEC-MODEL-TIER-PLANTYPE-001 M3 (REQ-MTP-018): moai update --plan-type ---

// writePlanTypeLLM writes a minimal llm.yaml under root's sections dir.
func writePlanTypeLLM(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile llm.yaml: %v", err)
	}
}

// writePlanTypeAgent writes a shipped-style agent .md with model:/effort:.
func writePlanTypeAgent(t *testing.T, root, name, model, effort string) {
	t.Helper()
	dir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll agents: %v", err)
	}
	content := "---\nname: " + name + "\nmodel: " + model + "\neffort: " + effort + "\n---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile agent %s: %v", name, err)
	}
}

func readPlanTypeAgent(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "moai", name+".md"))
	if err != nil {
		t.Fatalf("ReadFile agent %s: %v", name, err)
	}
	return string(b)
}

// TestUpdateCmd_HasPlanTypeFlag (REQ-MTP-018, AC-MTP-018b parse) — the
// --plan-type flag is registered on the update command.
func TestUpdateCmd_HasPlanTypeFlag(t *testing.T) {
	if updateCmd.Flags().Lookup("plan-type") == nil {
		t.Error("update command should have a --plan-type flag")
	}
}

// TestValidateUpdateFlags_ValidPlanType (REQ-MTP-018) — api/subscription pass.
func TestValidateUpdateFlags_ValidPlanType(t *testing.T) {
	for _, pt := range []string{"api", "subscription", ""} {
		t.Run(pt, func(t *testing.T) {
			if err := updateCmd.Flags().Set("plan-type", pt); err != nil {
				t.Fatal(err)
			}
			if err := validateUpdateFlags(updateCmd, []string{}); err != nil {
				t.Errorf("validateUpdateFlags with plan-type=%q should not error, got: %v", pt, err)
			}
		})
	}
	_ = updateCmd.Flags().Set("plan-type", "")
}

// TestValidateUpdateFlags_InvalidPlanType (REQ-MTP-018, AC-MTP-018b) — an
// out-of-set value errors, naming both api and subscription.
func TestValidateUpdateFlags_InvalidPlanType(t *testing.T) {
	for _, pt := range []string{"bogus", "enterprise"} {
		t.Run(pt, func(t *testing.T) {
			if err := updateCmd.Flags().Set("plan-type", pt); err != nil {
				t.Fatal(err)
			}
			err := validateUpdateFlags(updateCmd, []string{})
			if err == nil {
				t.Fatalf("validateUpdateFlags with plan-type=%q should error, got nil", pt)
			}
			msg := err.Error()
			if !strings.Contains(msg, "invalid --plan-type") {
				t.Errorf("error should mention 'invalid --plan-type', got: %v", err)
			}
			if !strings.Contains(msg, "api") || !strings.Contains(msg, "subscription") {
				t.Errorf("error should name both api and subscription, got: %v", err)
			}
		})
	}
	_ = updateCmd.Flags().Set("plan-type", "")
}

// TestApplyUpdateTierProfile_OverridePersistsAndApplies (REQ-MTP-018, AC-MTP-018b)
// — a project persisted plan_type: subscription, updated via --plan-type api,
// ends with llm.yaml plan_type: api AND agent frontmatter matching an api cell.
func TestApplyUpdateTierProfile_OverridePersistsAndApplies(t *testing.T) {
	root := t.TempDir()
	writePlanTypeLLM(t, root, "llm:\n  plan_type: \"subscription\"\n  performance_tier: \"medium\"\n")
	// Seed manager-develop with the subscription/medium cell (sonnet/high).
	writePlanTypeAgent(t, root, "manager-develop", "sonnet", "high")

	if err := applyUpdateTierProfile(root, "api"); err != nil {
		t.Fatalf("applyUpdateTierProfile: %v", err)
	}

	// llm.yaml now persists plan_type: api.
	llm, _ := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if !strings.Contains(string(llm), "plan_type: api") {
		t.Errorf("llm.yaml should persist plan_type: api, got:\n%s", llm)
	}
	// api/medium manager-develop = opus/high (spec §B.6), NOT the subscription cell.
	agent := readPlanTypeAgent(t, root, "manager-develop")
	if !strings.Contains(agent, "model: opus") || !strings.Contains(agent, "effort: high") {
		t.Errorf("manager-develop should be rewritten to api/medium (opus/high), got:\n%s", agent)
	}
	if strings.Contains(agent, "model: sonnet") {
		t.Errorf("manager-develop still carries the subscription model (sonnet):\n%s", agent)
	}
}

// TestApplyUpdateTierProfile_PersistedDefaultNoFlag (REQ-MTP-018, AC-MTP-018a) —
// with NO flag, the update path reads the persisted plan_type: api and re-applies
// the api-branch profile.
func TestApplyUpdateTierProfile_PersistedDefaultNoFlag(t *testing.T) {
	root := t.TempDir()
	writePlanTypeLLM(t, root, "llm:\n  plan_type: \"api\"\n  performance_tier: \"medium\"\n")
	writePlanTypeAgent(t, root, "manager-develop", "sonnet", "high") // seed with a subscription-looking value

	if err := applyUpdateTierProfile(root, ""); err != nil {
		t.Fatalf("applyUpdateTierProfile: %v", err)
	}

	agent := readPlanTypeAgent(t, root, "manager-develop")
	if !strings.Contains(agent, "model: opus") {
		t.Errorf("no-flag update should apply the persisted api profile (opus), got:\n%s", agent)
	}
	if strings.Contains(agent, "model: sonnet") {
		t.Errorf("no-flag update left the subscription model (sonnet):\n%s", agent)
	}
}

// TestApplyUpdateTierProfile_InvalidFlagError (REQ-MTP-018) — a defensive
// out-of-set flag value returns an error naming the closed set.
func TestApplyUpdateTierProfile_InvalidFlagError(t *testing.T) {
	root := t.TempDir()
	writePlanTypeLLM(t, root, "llm:\n  plan_type: \"subscription\"\n")
	err := applyUpdateTierProfile(root, "bogus")
	if err == nil {
		t.Fatal("applyUpdateTierProfile with bogus flag should error, got nil")
	}
	if !strings.Contains(err.Error(), "api") || !strings.Contains(err.Error(), "subscription") {
		t.Errorf("error should name the closed set, got: %v", err)
	}
}
