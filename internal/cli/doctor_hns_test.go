// SPEC-HNS-PREFIX-RENAME-001 M2 — RED-phase specification tests for the
// moai doctor skill layers (REQ-HPR-013):
//   - doctor_skills.go classifySkill: hns-* → "INFO" (user customization)
//   - doctor_harness.go checkLayer1Triggers: hns-* dirs included in the L1 scan
//   - doctor_harness.go checkLayer6AgentActivation: skills: refs carrying the
//     hns- prefix are resolved against disk (dangling refs FAIL)
//
// @MX:NOTE: [AUTO] AC-HPR-008 unit legs — hns- recognition in doctor layers.
// classifySkill("hns-*") already returned INFO via the non-moai catch-all; the
// explicit-branch test pins the classification so a future branch reorder
// cannot silently change it. The L1/L6 legs are behavioral (RED before M2).
// @MX:SPEC: SPEC-HNS-PREFIX-RENAME-001 acceptance.md AC-HPR-008
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClassifySkill_HNS pins the hns- classification as user customization.
func TestClassifySkill_HNS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		skillName string
		wantClass string
	}{
		{"hns-acme-verify", "INFO"},
		{"hns-moaiadk-patterns", "INFO"},
		// Template-managed builder namespace stays in moai- territory
		// (WARN = unknown moai- skill, pre-existing allowlist behavior).
		{"moai-harness-learner", "WARN"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.skillName, func(t *testing.T) {
			t.Parallel()
			if got := classifySkill(tt.skillName); got != tt.wantClass {
				t.Errorf("classifySkill(%q) = %q, want %q", tt.skillName, got, tt.wantClass)
			}
		})
	}
}

// TestCheckLayer1Triggers_HNSDirScanned asserts the L1 trigger scan includes
// hns- skill directories: an hns- dir missing its SKILL.md MUST surface as a
// problem (before M2 the hns- dir is skipped and the scan passes silently).
func TestCheckLayer1Triggers_HNSDirScanned(t *testing.T) {
	t.Parallel()

	skillsDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(skillsDir, "hns-acme-verify"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	status, detail := checkLayer1Triggers(skillsDir)
	if status != "FAIL" {
		t.Fatalf("checkLayer1Triggers = %q (%s), want FAIL for hns- dir without SKILL.md", status, detail)
	}
	if !strings.Contains(detail, "hns-acme-verify") {
		t.Errorf("detail %q does not name the hns- dir", detail)
	}
}

// TestCheckLayer6AgentActivation_HNSDanglingRef asserts the skills: reference
// resolver treats hns- prefixed refs as resolvable user-skill references: a
// dangling hns- ref MUST FAIL (before M2 the hns- ref is skipped as if it were
// a template skill and passes silently).
func TestCheckLayer6AgentActivation_HNSDanglingRef(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	agentsDir := filepath.Join(root, "agents")
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}

	agent := "---\nname: hns-acme-core-specialist\ndescription: acme core specialist\nskills:\n  - hns-acme-missing\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "hns-acme-core-specialist.md"), []byte(agent), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	status, detail := checkLayer6AgentActivation(agentsDir, skillsDir)
	if status != "FAIL" {
		t.Fatalf("checkLayer6AgentActivation = %q (%s), want FAIL for dangling hns- skills: ref", status, detail)
	}
	if !strings.Contains(detail, "hns-acme-missing") {
		t.Errorf("detail %q does not name the dangling hns- reference", detail)
	}
}

// TestCheckLayer6AgentActivation_HNSResolvedRef is the GREEN twin: when the
// referenced hns- skill directory exists, the agent passes.
func TestCheckLayer6AgentActivation_HNSResolvedRef(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	agentsDir := filepath.Join(root, "agents")
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(filepath.Join(skillsDir, "hns-acme-verify"), 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}

	agent := "---\nname: hns-acme-core-specialist\ndescription: acme core specialist\nskills:\n  - hns-acme-verify\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "hns-acme-core-specialist.md"), []byte(agent), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	status, detail := checkLayer6AgentActivation(agentsDir, skillsDir)
	if status != "PASS" {
		t.Errorf("checkLayer6AgentActivation = %q (%s), want PASS for resolved hns- ref", status, detail)
	}
}
