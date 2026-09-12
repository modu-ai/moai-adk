package config

import (
	"errors"
	"strings"
	"testing"
)

const cgFixture = `llm:
  team_mode: cg # original mixed roles
  glm_env_var: MY_GLM_KEY
  glm:
    models: {high: glm-5.2, medium: glm-4.7, low: glm-4.5-air, fable: glm-4.7}
  unknown_keep: yes # user comment
`

func TestCGMigrationPlanPreservesNodesAndBecomesIdempotent(t *testing.T) {
	if err := GuardLegacyCG([]byte(cgFixture)); !errors.Is(err, ErrLegacyCG) {
		t.Fatalf("legacy guard %v", err)
	}
	plan, err := PlanCGMigration([]byte(cgFixture), "claude-only")
	if err != nil {
		t.Fatal(err)
	}
	for _, keep := range []string{"# original mixed roles", "# user comment", "glm_env_var: MY_GLM_KEY", "unknown_keep: yes", "high: glm-5.2", "teammate_mode: in-process", "teammate_provider: inherit"} {
		if !strings.Contains(string(plan.Bytes), keep) {
			t.Fatalf("lost %q: %s", keep, plan.Bytes)
		}
	}
	if err := GuardLegacyCG(plan.Bytes); err != nil {
		t.Fatal(err)
	}
	policy, err := ReadGatewayTeammatePolicy(plan.Bytes)
	if err != nil || !policy.Present || policy.Mode != "in-process" || policy.Provider != "inherit" {
		t.Fatalf("policy %+v %v", policy, err)
	}
	second, err := PlanCGMigration(plan.Bytes, "claude-only")
	if err != nil || !second.Unchanged || string(second.Bytes) != string(plan.Bytes) {
		t.Fatalf("not idempotent %v %+v", err, second)
	}
	if _, err := PlanCGMigration(plan.Bytes, "claude-glm"); err == nil {
		t.Fatal("opposite target accepted")
	}
	hybrid, err := PlanCGMigration([]byte(cgFixture), "claude-glm")
	if err != nil {
		t.Fatal(err)
	}
	policy, err = ReadGatewayTeammatePolicy(hybrid.Bytes)
	if err != nil || policy.Mode != "tmux" || policy.Provider != "glm" {
		t.Fatalf("hybrid %+v %v", policy, err)
	}
}

func TestCGRawReaderRejectsAmbiguityAndConflicts(t *testing.T) {
	cases := []string{
		"llm:\n  team_mode: cg\n  team_mode: claude\n",
		"base: &base {team_mode: cg}\nllm: *base\n",
		"llm:\n  team_mode: cg\n  mode: glm\n",
		"llm:\n  team_mode: cg\n  gateway: {teammate_mode: tmux, teammate_provider: glm}\n",
		"llm: []\n", "llm:\n  team_mode: [cg]\n", "llm: {team_mode: cg}\n---\nllm: {}\n",
	}
	for _, raw := range cases {
		if _, err := PlanCGMigration([]byte(raw), "claude-only"); err == nil {
			t.Fatalf("accepted ambiguous migration %q", raw)
		}
	}
	for _, raw := range []string{"llm: {team_mode: claude, gateway: {teammate_mode: tmux, teammate_provider: inherit}}", "llm: {team_mode: claude, gateway: {teammate_mode: in-process}}", "llm: {team_mode: glm, gateway: {teammate_mode: tmux, teammate_provider: glm}}"} {
		if _, err := ReadGatewayTeammatePolicy([]byte(raw)); err == nil {
			t.Fatalf("accepted conflicting policy %q", raw)
		}
	}
	for _, raw := range []string{"", "other: keep\n", "llm: {team_mode: claude}", "llm: {team_mode: glm}"} {
		if err := GuardLegacyCG([]byte(raw)); err != nil {
			t.Fatalf("normal guard: %v", err)
		}
	}
}

func TestCGRawMalformedPolicyAndMigrationEdges(t *testing.T) {
	for _, raw := range [][]byte{[]byte("["), []byte("[]"), {0xff}, []byte("llm: {gateway: []}"), []byte("llm: {gateway: {teammate_mode: [], teammate_provider: inherit}}"), []byte("llm: {gateway: {teammate_mode: in-process, teammate_provider: []}}"), []byte("llm: {team_mode: [], gateway: {teammate_mode: in-process, teammate_provider: inherit}}")} {
		if _, err := ReadGatewayTeammatePolicy(raw); err == nil {
			t.Fatalf("accepted policy %q", raw)
		}
	}
	for _, raw := range []string{"llm: {team_mode: []}", "llm: {team_mode: cg, team_mode: cg}"} {
		if err := GuardLegacyCG([]byte(raw)); err == nil {
			t.Fatal("guard accepted malformed")
		}
	}
	for _, raw := range []string{"", "llm: {team_mode: glm}", "llm: {team_mode: claude}", "llm: {team_mode: cg, mode: []}", "llm: {team_mode: cg, gateway: []}", "llm: {team_mode: claude, gateway: {teammate_mode: tmux}}"} {
		if _, err := PlanCGMigration([]byte(raw), "claude-only"); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	if _, err := PlanCGMigration([]byte(cgFixture), "unknown"); err == nil {
		t.Fatal("unknown target")
	}
	if p, err := ReadGatewayTeammatePolicy([]byte("llm: {gateway: {unrelated: keep}}")); err != nil || p.Present {
		t.Fatal("unrelated gateway setting")
	}
	matching := []byte("llm: {team_mode: cg, gateway: {teammate_mode: in-process, teammate_provider: inherit, other: keep}}")
	plan, err := PlanCGMigration(matching, "claude-only")
	if err != nil || !strings.Contains(string(plan.Bytes), "other: keep") {
		t.Fatalf("matching partial target %s %v", plan.Bytes, err)
	}
}
