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
	// ModelPolicyHigh uses explicit opus for most agents (Max $200 plan, highest quality).
	ModelPolicyHigh ModelPolicy = "high"
	// ModelPolicyMedium uses opus for critical agents, sonnet for standard, haiku for mechanical (Max $100 plan).
	ModelPolicyMedium ModelPolicy = "medium"
	// ModelPolicyLow uses no opus (Plus $20 plan). Sonnet for core agents, Haiku for the rest.
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
// Key: agent name, Value: [high_model, medium_model, low_model]
var agentModelMap = map[string][3]string{
	// Manager Agents
	"manager-spec":     {"opus", "opus", "sonnet"},
	"manager-ddd":      {"opus", "sonnet", "sonnet"},
	"manager-tdd":      {"opus", "sonnet", "sonnet"},
	"manager-docs":     {"sonnet", "haiku", "haiku"},
	"manager-quality":  {"haiku", "haiku", "haiku"},
	"manager-project":  {"opus", "sonnet", "haiku"},
	"manager-strategy": {"opus", "opus", "sonnet"},
	"manager-git":      {"haiku", "haiku", "haiku"},
	// Expert Agents
	"expert-backend":          {"opus", "sonnet", "sonnet"},
	"expert-frontend":         {"opus", "sonnet", "sonnet"},
	"expert-security":         {"opus", "opus", "sonnet"},
	"expert-devops":           {"opus", "sonnet", "haiku"},
	"expert-performance":      {"opus", "sonnet", "haiku"},
	"expert-debug":            {"opus", "sonnet", "sonnet"},
	"expert-testing":          {"opus", "sonnet", "haiku"},
	"expert-refactoring":      {"opus", "sonnet", "sonnet"},
	"expert-chrome-extension": {"opus", "sonnet", "haiku"},
	// Builder Agents
	"builder-agent":  {"opus", "sonnet", "haiku"},
	"builder-skill":  {"opus", "sonnet", "haiku"},
	"builder-plugin": {"opus", "sonnet", "haiku"},
	// Team Agents
	"team-researcher":   {"haiku", "haiku", "haiku"},
	"team-analyst":      {"opus", "sonnet", "haiku"},
	"team-architect":    {"opus", "opus", "sonnet"},
	"team-designer":     {"opus", "sonnet", "haiku"},
	"team-backend-dev":  {"opus", "sonnet", "sonnet"},
	"team-frontend-dev": {"opus", "sonnet", "sonnet"},
	"team-tester":       {"opus", "sonnet", "haiku"},
	"team-quality":      {"haiku", "haiku", "haiku"},
}

// GetAgentModel returns the model string for a given agent under the specified policy.
func GetAgentModel(policy ModelPolicy, agentName string) string {
	models, ok := agentModelMap[agentName]
	if !ok {
		return "" // Unknown agent: caller should skip to preserve current model
	}

	switch policy {
	case ModelPolicyHigh:
		return models[0]
	case ModelPolicyMedium:
		return models[1]
	case ModelPolicyLow:
		return models[2]
	default:
		return "sonnet" // Unknown policy: safe fallback
	}
}

// modelLineRegex matches the "model:" line in YAML frontmatter.
var modelLineRegex = regexp.MustCompile(`(?m)^model:\s*\S+`)

// ApplyModelPolicy patches the model: field in all agent definition files
// under the given project root based on the specified model policy.
// It also updates the manifest hashes for patched files.
func ApplyModelPolicy(projectRoot string, policy ModelPolicy, mgr manifest.Manager) error {
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
		if targetModel == "" {
			continue // Unknown agent: preserve current model
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
