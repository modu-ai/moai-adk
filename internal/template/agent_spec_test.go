package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAgentAntiTriggers verifies that agent descriptions contain anti-trigger boundaries
func TestAgentAntiTriggers(t *testing.T) {
	agentsDir := filepath.Join("templates", ".claude", "agents", "moai")

	agents, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("Failed to read agents directory: %v", err)
	}

	targetAgents := map[string]bool{
		"manager-quality.md":    true,
		"manager-strategy.md":   true,
		"manager-docs.md":       true,
		"manager-spec.md":       true,
		"manager-cycle.md":      true,
		"manager-project.md":    true,
		"manager-git.md":        true,
		"manager-develop.md":    true,
		"expert-backend.md":     true,
		"expert-frontend.md":    true,
		"expert-security.md":    true,
		"expert-devops.md":      true,
		"expert-performance.md": true,
		"expert-refactoring.md": true,
		"evaluator-active.md":   true,
		"evaluator-plan.md":     true,
		"evaluator-quality.md":  true,
	}

	antiTriggerCount := 0
	missingAntiTriggers := []string{}

	for _, agent := range agents {
		if !targetAgents[agent.Name()] {
			continue
		}

		agentPath := filepath.Join(agentsDir, agent.Name())
		content, err := os.ReadFile(agentPath)
		if err != nil {
			t.Fatalf("Failed to read agent file %s: %v", agent.Name(), err)
		}

		contentStr := string(content)
		hasAntiTrigger := strings.Contains(contentStr, "NOT for:") ||
			strings.Contains(contentStr, "Not for:") ||
			strings.Contains(contentStr, "not for:") ||
			strings.Contains(contentStr, "Scope exclusion")

		if !hasAntiTrigger {
			missingAntiTriggers = append(missingAntiTriggers, agent.Name())
		} else {
			antiTriggerCount++
		}
	}

	if len(missingAntiTriggers) > 0 {
		t.Errorf("Agents missing anti-trigger boundaries: %v", missingAntiTriggers)
	}

	if antiTriggerCount < 15 {
		t.Errorf("Expected at least 15 agents with anti-triggers, got %d", antiTriggerCount)
	}

	t.Logf("Anti-trigger check: %d/%d agents have anti-triggers", antiTriggerCount, len(targetAgents))
}

// TestReferenceSkillStructure verifies that reference skills follow the correct pattern
func TestReferenceSkillStructure(t *testing.T) {
	skillsDirs := []string{
		filepath.Join("templates", ".claude", "skills"),
		filepath.Join("..", "..", ".claude", "skills"),
	}

	targetSkills := []string{
		"moai-ref-api-patterns",
		"moai-ref-react-patterns",
		"moai-ref-git-workflow",
	}

	skillsFound := map[string]bool{}

	for _, skillsDir := range skillsDirs {
		for _, skillName := range targetSkills {
			skillPath := filepath.Join(skillsDir, skillName, "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				skillsFound[skillName] = true

				content, err := os.ReadFile(skillPath)
				if err != nil {
					t.Fatalf("Failed to read skill file %s: %v", skillName, err)
				}

				contentStr := string(content)

				if !strings.Contains(contentStr, "Target Agent") && !strings.Contains(contentStr, "target agent") {
					t.Errorf("Skill %s missing 'Target Agent' declaration", skillName)
				}

				hasReferenceContent := strings.Contains(contentStr, "|") ||
					strings.Contains(contentStr, "- [") ||
					strings.Contains(contentStr, "## ") ||
					strings.Contains(contentStr, "Pattern")

				if !hasReferenceContent {
					t.Errorf("Skill %s appears to be procedural instead of reference content", skillName)
				}
			}
		}
	}

	for _, skillName := range targetSkills {
		if !skillsFound[skillName] {
			t.Errorf("Required reference skill not found: %s", skillName)
		}
	}
}

// TestReferenceSkillTargetAgents verifies correct target agent declarations
func TestReferenceSkillTargetAgents(t *testing.T) {
	expectedTargets := map[string]string{
		"moai-ref-api-patterns":  "expert-backend",
		"moai-ref-react-patterns": "expert-frontend",
		"moai-ref-git-workflow":   "manager-git",
	}

	skillsDirs := []string{
		filepath.Join("templates", ".claude", "skills"),
		filepath.Join("..", "..", ".claude", "skills"),
	}

	for _, skillsDir := range skillsDirs {
		for skillName, expectedTarget := range expectedTargets {
			skillPath := filepath.Join(skillsDir, skillName, "SKILL.md")
			content, err := os.ReadFile(skillPath)
			if err != nil {
				continue
			}

			contentStr := string(content)

			if !strings.Contains(contentStr, expectedTarget) {
				t.Errorf("Skill %s should target %s, but declaration not found", skillName, expectedTarget)
			}
		}
	}
}
