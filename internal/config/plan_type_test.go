package config

// plan_type_test.go — SPEC-MODEL-TIER-PLANTYPE-001 M1 tests (AC-MTP-001..003).
// Covers: llm.plan_type loader parse, EffectivePlanType absent/empty → subscription
// resolution, and closed-set {api, subscription} validation whose error message
// names the offending value plus both closed-set tokens.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedLLMSection writes llm.yaml content to a temp project root and returns the root.
func seedLLMSection(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestLLMConfigPlanTypeParse verifies the LLM loader parses plan_type into
// LLMConfig.PlanType for both closed-set values (AC-MTP-001, REQ-MTP-001).
func TestLLMConfigPlanTypeParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{"api", "llm:\n    plan_type: api\n", "api"},
		{"subscription", "llm:\n    plan_type: subscription\n", "subscription"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := seedLLMSection(t, tt.yaml)
			cfg, err := NewConfigManager().LoadRaw(root)
			if err != nil {
				t.Fatalf("LoadRaw: %v", err)
			}
			if cfg.LLM.PlanType != tt.want {
				t.Errorf("PlanType = %q, want %q", cfg.LLM.PlanType, tt.want)
			}
		})
	}
}

// TestEffectivePlanTypeDefault verifies that an absent, empty, or whitespace-only
// plan_type resolves to subscription via the exported accessor, while an explicit
// value passes through (AC-MTP-002, REQ-MTP-002).
func TestEffectivePlanTypeDefault(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		planType string
		want     string
	}{
		{"absent/zero-value", "", PlanTypeSubscription},
		{"whitespace only", "   ", PlanTypeSubscription},
		{"api explicit", PlanTypeAPI, PlanTypeAPI},
		{"subscription explicit", PlanTypeSubscription, PlanTypeSubscription},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			llm := LLMConfig{PlanType: tt.planType}
			if got := llm.EffectivePlanType(); got != tt.want {
				t.Errorf("EffectivePlanType() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestIsValidPlanType verifies the closed-set membership predicate (REQ-MTP-001).
func TestIsValidPlanType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value string
		want  bool
	}{
		{PlanTypeAPI, true},
		{PlanTypeSubscription, true},
		{"", false},
		{"enterprise", false},
		{"API", false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			if got := IsValidPlanType(tt.value); got != tt.want {
				t.Errorf("IsValidPlanType(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestValidPlanTypes verifies the option-list helper returns exactly the closed set.
func TestValidPlanTypes(t *testing.T) {
	t.Parallel()
	got := ValidPlanTypes()
	if len(got) != 2 {
		t.Fatalf("ValidPlanTypes() len = %d, want 2 (got %v)", len(got), got)
	}
	seen := map[string]bool{}
	for _, v := range got {
		seen[v] = true
	}
	if !seen[PlanTypeAPI] || !seen[PlanTypeSubscription] {
		t.Errorf("ValidPlanTypes() = %v, want the set {api, subscription}", got)
	}
}

// TestValidatePlanTypeClosedSet verifies out-of-set plan_type values are rejected
// with an error naming the offending value and both closed-set tokens, while the
// closed-set members and the empty value (BC default) pass (AC-MTP-003, REQ-MTP-003).
func TestValidatePlanTypeClosedSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"api valid", "api", false},
		{"subscription valid", "subscription", false},
		{"empty valid (resolves to subscription)", "", false},
		{"enterprise invalid", "enterprise", true},
		{"uppercase API invalid (closed set is lowercase-exact)", "API", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := NewDefaultConfig()
			cfg.User.Name = "TestUser"
			cfg.LLM.PlanType = tt.value
			loaded := map[string]bool{"user": true}

			err := Validate(cfg, loaded)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Validate() expected no error for plan_type %q, got: %v", tt.value, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() expected error for plan_type %q, got nil", tt.value)
			}
			msg := err.Error()
			// AC-MTP-003: message must contain the offending value plus both closed-set tokens.
			for _, tok := range []string{tt.value, "api", "subscription"} {
				if !strings.Contains(msg, tok) {
					t.Errorf("error message %q missing token %q", msg, tok)
				}
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("expected ErrInvalidConfig for plan_type %q, got: %v", tt.value, err)
			}
		})
	}
}
