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

// ModelIDOpus47 is the canonical model ID for Claude Opus 4.7.
// Used by launcher.go to route the new model and by profile translations.
const ModelIDOpus47 = "claude-opus-4-7"

// Effort level constants for the 5-tier effort system.
// These are separate from ModelPolicy (3-tier). ModelPolicy selects the model;
// effort levels control reasoning depth within a model session.
// Supported by Claude Code v2.1.68+ for Opus 4.6 and Opus 4.7.
const (
	// EffortLevelLow is the fastest, least thorough effort level.
	EffortLevelLow = "low"
	// EffortLevelMedium is the balanced default effort level.
	EffortLevelMedium = "medium"
	// EffortLevelHigh activates deep reasoning for complex tasks.
	EffortLevelHigh = "high"
	// EffortLevelXHigh is extended high reasoning for Opus 4.7+.
	// Not supported on Opus 4.6.
	EffortLevelXHigh = "xhigh"
	// EffortLevelMax is the maximum effort level.
	// On Opus 4.6, max is the highest supported level.
	// On Opus 4.7+, xhigh and max are both available.
	EffortLevelMax = "max"
)

// agentEffortMap specifies explicit effort overrides for reasoning-heavy agents.
// Only the 6 Opus 4.7 reasoning agents have entries.
// The remaining 22 agents return "" (empty string) so the Opus 4.7 runtime
// default (xhigh) applies without any explicit override injection.
//
// Key: agent name, Value: effort level string
var agentEffortMap = map[string]string{
	"manager-spec":       EffortLevelXHigh,
	"manager-strategy":   EffortLevelXHigh,
	"plan-auditor":       EffortLevelHigh,
	"evaluator-active":   EffortLevelHigh,
	"expert-security":    EffortLevelHigh,
	"expert-refactoring": EffortLevelHigh,
}

// GetAgentEffort returns the effort level override for the given agent.
// Returns "" (empty string) for agents not in agentEffortMap, which signals
// the caller to use the runtime default (Opus 4.7 defaults to xhigh).
//
// @MX:NOTE: [AUTO] Separate from GetAgentModel — ModelPolicy⊥Effort by design.
func GetAgentEffort(agentName string) string {
	return agentEffortMap[agentName]
}

// effortLineRegex matches an existing effort: field in YAML frontmatter.
var effortLineRegex = regexp.MustCompile(`(?m)^effort:\s*\S+`)

// frontmatterCloseRegex matches the closing --- delimiter of a YAML frontmatter block.
// It is intentionally written to match the second occurrence via string splitting logic
// in insertEffortInFrontmatter, not via regex alone.
var frontmatterOpenPrefix = "---\n"

// insertEffortInFrontmatter inserts "effort: <level>" into the YAML frontmatter block
// of content. Returns unchanged content when:
//   - the file does not start with "---\n" (no frontmatter)
//   - no closing "---" is found
//   - an effort: line already exists (caller is responsible for this guard)
func insertEffortInFrontmatter(content []byte, effortLevel string) []byte {
	s := string(content)
	if !strings.HasPrefix(s, frontmatterOpenPrefix) {
		return content // No YAML frontmatter
	}
	// Locate the closing --- after the opening one.
	// Content after the opening "---\n" (offset 4) is searched for "\n---".
	rest := s[len(frontmatterOpenPrefix):]
	closingIdx := strings.Index(rest, "\n---")
	if closingIdx == -1 {
		return content // Malformed frontmatter — leave untouched
	}
	// closingIdx points at the "\n" before "---". Insert before the newline.
	insertPos := len(frontmatterOpenPrefix) + closingIdx + 1 // position of the closing "---" line
	return []byte(s[:insertPos] + "effort: " + effortLevel + "\n" + s[insertPos:])
}

// ApplyEffortPolicy injects effort level overrides into agent definition files
// under the given project root. It mirrors ApplyModelPolicy in structure but
// operates on the effort: frontmatter field instead of model:.
//
// Injection rules:
//   - Agent already has effort: field → preserved (user customisation wins)
//   - Agent is in agentEffortMap and has no effort: field → value injected
//   - Agent is not in agentEffortMap → no-op (runtime default applies)
//   - File has no YAML frontmatter → skipped silently
//
// Manifest hashes are updated for every file that is written.
//
// @MX:ANCHOR: [AUTO] ApplyEffortPolicy — called from initializer and update paths; mirrors ApplyModelPolicy contract
// @MX:REASON: [AUTO] fan_in >= 2 (initializer.go + update.go); public API boundary for effort wiring
func ApplyEffortPolicy(projectRoot string, mgr manifest.Manager) error {
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
		targetEffort := GetAgentEffort(agentName)
		if targetEffort == "" {
			continue // Not in effort map — use runtime default, no injection
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read agent file %q: %w", entry.Name(), err)
		}

		// Preserve existing effort: value (user customisation wins)
		if effortLineRegex.Match(content) {
			continue
		}

		newContent := insertEffortInFrontmatter(content, targetEffort)
		if string(newContent) == string(content) {
			continue // No frontmatter or no change — skip
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
	"expert-backend":     {"opus", "sonnet", "sonnet"},
	"expert-frontend":    {"opus", "sonnet", "sonnet"},
	"expert-security":    {"opus", "opus", "sonnet"},
	"expert-devops":      {"opus", "sonnet", "haiku"},
	"expert-performance": {"opus", "sonnet", "haiku"},
	"expert-debug":       {"opus", "sonnet", "sonnet"},
	"expert-testing":     {"opus", "sonnet", "haiku"},
	"expert-refactoring": {"opus", "sonnet", "sonnet"},
	// Builder Agents
	"builder-agent":  {"opus", "sonnet", "haiku"},
	"builder-skill":  {"opus", "sonnet", "haiku"},
	"builder-plugin": {"opus", "sonnet", "haiku"},
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
