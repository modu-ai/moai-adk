package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ModelPolicy represents the token consumption tier for agent models.
type ModelPolicy string

const (
	// ModelPolicyHigh uses inherit for all agents (maximum quality, highest token consumption).
	ModelPolicyHigh ModelPolicy = "high"
	// ModelPolicyMedium uses opus for critical agents, sonnet for implementation, haiku for mechanical.
	ModelPolicyMedium ModelPolicy = "medium"
	// ModelPolicyLow uses explicit opus/sonnet/haiku for maximum efficiency.
	ModelPolicyLow ModelPolicy = "low"
)

// DefaultModelPolicy is the default model policy for new projects.
const DefaultModelPolicy = ModelPolicyHigh

// ValidModelPolicies returns all valid model policy values.
func ValidModelPolicies() []string {
	return []string{string(ModelPolicyHigh), string(ModelPolicyMedium), string(ModelPolicyLow)}
}

// IsValidModelPolicy checks if the given string is a valid model policy.
func IsValidModelPolicy(s string) bool {
	switch ModelPolicy(s) {
	case ModelPolicyHigh, ModelPolicyMedium, ModelPolicyLow:
		return true
	}
	return false
}

// agentModelMap defines the model assignment for each agent under each policy.
// Key: agent name, Value: [medium_model, low_model]
// "high" policy always returns "inherit" for all agents.
var agentModelMap = map[string][2]string{
	// Manager Agents
	"manager-spec":     {"opus", "opus"},
	"manager-ddd":      {"sonnet", "sonnet"},
	"manager-tdd":      {"sonnet", "sonnet"},
	"manager-docs":     {"haiku", "haiku"},
	"manager-quality":  {"haiku", "haiku"},
	"manager-project":  {"sonnet", "opus"},
	"manager-strategy": {"opus", "opus"},
	"manager-git":      {"haiku", "haiku"},
	// Expert Agents
	"expert-backend":          {"sonnet", "sonnet"},
	"expert-frontend":         {"sonnet", "sonnet"},
	"expert-security":         {"opus", "opus"},
	"expert-devops":           {"sonnet", "sonnet"},
	"expert-performance":      {"sonnet", "sonnet"},
	"expert-debug":            {"sonnet", "opus"},
	"expert-testing":          {"sonnet", "sonnet"},
	"expert-refactoring":      {"sonnet", "sonnet"},
	"expert-chrome-extension": {"sonnet", "sonnet"},
	// Builder Agents
	"builder-agent":  {"sonnet", "haiku"},
	"builder-skill":  {"sonnet", "haiku"},
	"builder-plugin": {"sonnet", "haiku"},
	// Team Agents
	"team-researcher":   {"haiku", "haiku"},
	"team-analyst":      {"sonnet", "sonnet"},
	"team-architect":    {"opus", "opus"},
	"team-designer":     {"sonnet", "sonnet"},
	"team-backend-dev":  {"sonnet", "sonnet"},
	"team-frontend-dev": {"sonnet", "sonnet"},
	"team-tester":       {"sonnet", "haiku"},
	"team-quality":      {"haiku", "haiku"},
}

// GetAgentModel returns the model string for a given agent under the specified policy.
func GetAgentModel(policy ModelPolicy, agentName string) string {
	if policy == ModelPolicyHigh {
		return "inherit"
	}

	models, ok := agentModelMap[agentName]
	if !ok {
		return "inherit" // Unknown agent defaults to inherit
	}

	switch policy {
	case ModelPolicyMedium:
		return models[0]
	case ModelPolicyLow:
		return models[1]
	default:
		return "inherit"
	}
}

// modelLineRegex matches the "model:" line in YAML frontmatter.
var modelLineRegex = regexp.MustCompile(`(?m)^model:\s*\S+`)

// ApplyModelPolicy patches the model: field in all agent definition files
// under the given project root based on the specified model policy.
// It also updates the manifest hashes for patched files.
func ApplyModelPolicy(projectRoot string, policy ModelPolicy, mgr manifest.Manager) error {
	if policy == ModelPolicyHigh {
		return nil // No patching needed for "high" policy
	}

	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "moai")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No agents directory yet
		}
		return fmt.Errorf("read agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		agentName := strings.TrimSuffix(entry.Name(), ".md")
		targetModel := GetAgentModel(policy, agentName)
		if targetModel == "inherit" {
			continue // No change needed
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read agent file %q: %w", entry.Name(), err)
		}

		// Replace the model: line in YAML frontmatter
		newContent := modelLineRegex.ReplaceAll(content, []byte("model: "+targetModel))

		if string(newContent) == string(content) {
			continue // No change
		}

		if err := os.WriteFile(filePath, newContent, 0o644); err != nil {
			return fmt.Errorf("write agent file %q: %w", entry.Name(), err)
		}

		// Update manifest hash for the patched file
		relPath := filepath.Join(".claude", "agents", "moai", entry.Name())
		hash := manifest.HashBytes(newContent)
		if err := mgr.Track(relPath, manifest.TemplateManaged, hash); err != nil {
			return fmt.Errorf("track patched agent %q: %w", entry.Name(), err)
		}
	}

	return nil
}
