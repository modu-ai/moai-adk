// Package plan contains the analysis, classification, and namespace-protection
// predicate functions extracted from the root update command during M3d-A
// decomposition (SPEC-CLI-TUX-V3-003). Behavior is byte-identical to the
// pre-decomposition implementation; only the package location changed.
package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/merge"
	"gopkg.in/yaml.v3"
)

// MaxConfigSize is the maximum allowed size for a config YAML file to prevent DoS.
// Moved intact from root update.go during M3d-A decomposition; only GetProjectConfigVersion references it.
const MaxConfigSize = 10 * 1024 * 1024 // 10MB

func ClassifyFileRisk(filename string, exists bool) string {
	base := filepath.Base(filename)

	// High risk files
	highRiskFiles := []string{"CLAUDE.md", "settings.json"}
	if slices.Contains(highRiskFiles, base) {
		return "high"
	}

	// New files are low risk
	if !exists {
		return "low"
	}

	// Existing files are medium risk
	return "medium"
}
func DetermineStrategy(filename string) merge.MergeStrategy {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)

	switch {
	case base == "CLAUDE.md":
		return merge.SectionMerge
	case base == ".gitignore":
		return merge.EntryMerge
	case ext == ".json":
		return merge.JSONMerge
	case ext == ".yaml" || ext == ".yml":
		return merge.YAMLDeep
	default:
		return merge.LineMerge
	}
}
func DetermineChangeType(exists bool) string {
	if exists {
		return "update existing"
	}
	return "new file"
}
func AnalyzeFiles(templates []string, projectRoot string) []merge.FileAnalysis {
	var files []merge.FileAnalysis
	for _, tmpl := range templates {
		// Strip .tmpl suffix first - display and filter using rendered target path
		displayPath := tmpl
		if before, ok := strings.CutSuffix(tmpl, ".tmpl"); ok {
			displayPath = before
		}

		// Filter out MoAI-managed files - they are automatically installed
		if IsMoaiManaged(displayPath) {
			continue
		}

		// Use rendered target path for existence check
		targetPath := filepath.Join(projectRoot, displayPath)

		_, err := os.Stat(targetPath)
		exists := err == nil

		// Classify risk and determine strategy (use displayPath for classification)
		risk := ClassifyFileRisk(displayPath, exists)
		strategy := DetermineStrategy(displayPath)
		changeType := DetermineChangeType(exists)

		files = append(files, merge.FileAnalysis{
			Path:      displayPath,
			Changes:   changeType,
			Strategy:  strategy,
			RiskLevel: risk,
			Note:      "",
		})
	}
	return files
}
func IsUserAreaPath(rel string) bool {
	// Normalize to forward slashes for consistent matching on all platforms.
	norm := strings.ReplaceAll(rel, "\\", "/")

	// .claude/skills/hns-* (canonical user-owned, any sub-path inside)
	// @MX:NOTE: [AUTO] SPEC-HNS-PREFIX-RENAME-001 M2 — canonical hns-* recognition (REQ-HPR-005).
	if strings.HasPrefix(norm, ".claude/skills/hns-") {
		return true
	}

	// .claude/skills/harness-* (legacy gen-2 user-owned, any sub-path inside)
	// @MX:NOTE: [AUTO] SPEC-V3R6-HARNESS-NAMESPACE-V2-001 M1 — legacy harness-* recognition.
	if strings.HasPrefix(norm, ".claude/skills/harness-") {
		return true
	}

	// .claude/skills/my-harness-* (legacy user-owned, REQ-HNS-005 backward-compat)
	if strings.HasPrefix(norm, ".claude/skills/my-harness-") {
		return true
	}

	// .claude/agents/harness/ (canonical user-owned, any sub-path inside)
	if strings.HasPrefix(norm, ".claude/agents/harness/") || norm == ".claude/agents/harness" {
		return true
	}

	// .claude/agents/my-harness/ (legacy user-owned, REQ-HNS-005 backward-compat)
	if strings.HasPrefix(norm, ".claude/agents/my-harness/") || norm == ".claude/agents/my-harness" {
		return true
	}

	// .claude/commands/harness/ (SPEC-V3R6-HARNESS-V4-001 M1 — user-generated /harness:<name> commands, AC-HV4-010a)
	// @MX:NOTE: [AUTO] SPEC-V3R6-HARNESS-V4-001 M1 — harness command subdirectory user-owned.
	if strings.HasPrefix(norm, ".claude/commands/harness/") || norm == ".claude/commands/harness" {
		return true
	}

	// .claude/workflows/hns-*.js (SPEC-HNS-PREFIX-RENAME-001 M2 — canonical user Runner Workflows, REQ-HPR-005)
	// @MX:NOTE: [AUTO] SPEC-HNS-PREFIX-RENAME-001 M2 — hns- Runner Workflow user-owned.
	if strings.HasPrefix(norm, ".claude/workflows/hns-") {
		return true
	}

	// .claude/workflows/harness-*.js (SPEC-V3R6-HARNESS-V4-001 M1 — legacy user Runner Workflows, AC-HV4-010b)
	// @MX:NOTE: [AUTO] SPEC-V3R6-HARNESS-V4-001 M1 — harness Runner Workflow user-owned.
	if strings.HasPrefix(norm, ".claude/workflows/harness-") {
		return true
	}

	return false
}

// @MX:ANCHOR: [AUTO] IsUserOwnedNamespace is the authoritative user-owned namespace check for moai update
// @MX:REASON: [AUTO] called by collectUserOwnedFiles (strict preserve inventory), assertNoUserOwnedNamespaceTouch (sentinel), and isUserOwnedNamespaceConservative (REQ-SEC-005 backup superset); fan_in = 3
func IsUserOwnedNamespace(rel string) bool {
	// Normalize to forward slashes for consistent matching on all platforms (NFR-UNP-003).
	norm := strings.ReplaceAll(rel, "\\", "/")

	// REQ-HPR-005: user harness skills (canonical hns-* prefix)
	// @MX:NOTE: [AUTO] SPEC-HNS-PREFIX-RENAME-001 M2 — canonical hns-* recognition.
	if strings.HasPrefix(norm, ".claude/skills/hns-") {
		return true
	}

	// REQ-HNS-001: user harness skills (legacy gen-2 harness-* prefix, doctrine §24.1)
	// @MX:NOTE: [AUTO] SPEC-V3R6-HARNESS-NAMESPACE-V2-001 M1 — legacy harness-* recognition.
	if strings.HasPrefix(norm, ".claude/skills/harness-") {
		return true
	}

	// REQ-UNP-001 / REQ-HNS-005: legacy my-harness-* user harness skills (backward-compat)
	if strings.HasPrefix(norm, ".claude/skills/my-harness-") {
		return true
	}

	// REQ-UNP-002: harness agents directory (overrides IsMoaiManaged for this prefix)
	if norm == ".claude/agents/harness" || strings.HasPrefix(norm, ".claude/agents/harness/") {
		return true
	}

	// REQ-UNP-003: harness extension directory
	if norm == ".moai/harness" || strings.HasPrefix(norm, ".moai/harness/") {
		return true
	}

	// SPEC-V3R6-HARNESS-V4-001 M1 (AC-HV4-010a): .claude/commands/harness/ user-generated /harness:<name> commands.
	// @MX:NOTE: [AUTO] SPEC-V3R6-HARNESS-V4-001 M1 — harness command subdirectory user-owned.
	if norm == ".claude/commands/harness" || strings.HasPrefix(norm, ".claude/commands/harness/") {
		return true
	}

	// SPEC-HNS-PREFIX-RENAME-001 M2 (REQ-HPR-005): .claude/workflows/hns-*.js canonical user Runner Workflows.
	// @MX:NOTE: [AUTO] SPEC-HNS-PREFIX-RENAME-001 M2 — hns- Runner Workflow user-owned.
	if strings.HasPrefix(norm, ".claude/workflows/hns-") {
		return true
	}

	// SPEC-V3R6-HARNESS-V4-001 M1 (AC-HV4-010b): .claude/workflows/harness-*.js legacy user Runner Workflows.
	// @MX:NOTE: [AUTO] SPEC-V3R6-HARNESS-V4-001 M1 — harness Runner Workflow user-owned.
	if strings.HasPrefix(norm, ".claude/workflows/harness-") {
		return true
	}

	// REQ-UNP-009: user direct-added skills (any name except "moai" or "moai-*" prefix)
	if strings.HasPrefix(norm, ".claude/skills/") {
		rest := strings.TrimPrefix(norm, ".claude/skills/")
		seg := strings.SplitN(rest, "/", 2)[0]
		if seg != "" && seg != "moai" && !strings.HasPrefix(seg, "moai-") {
			return true
		}
	}

	// REQ-UNP-009: user direct-added agents (top-level files or directories
	// outside {core, expert, meta} and not "moai-" / "manager-" / "expert-" /
	// "builder-" / "evaluator-" prefixed system agents).
	if strings.HasPrefix(norm, ".claude/agents/") {
		rest := strings.TrimPrefix(norm, ".claude/agents/")
		seg := strings.SplitN(rest, "/", 2)[0]
		switch seg {
		case "core", "expert", "meta":
			return false // MoAI system agents
		case "harness":
			return true // REQ-UNP-002 (also covered above; defense-in-depth)
		}
		// Strip extension for filename-only segment (e.g., "custom-agent.md" → "custom-agent")
		base := seg
		if dot := strings.LastIndex(base, "."); dot > 0 {
			base = base[:dot]
		}
		if base != "" && !strings.HasPrefix(base, "moai-") && !strings.HasPrefix(base, "moai") &&
			!strings.HasPrefix(base, "manager-") && !strings.HasPrefix(base, "expert-") &&
			!strings.HasPrefix(base, "builder-") && !strings.HasPrefix(base, "evaluator-") {
			return true
		}
	}

	return false
}
func IsMoaiManaged(path string) bool {
	// Check .moai/config/ paths first
	if strings.HasPrefix(path, ".moai/config/") || strings.HasPrefix(path, ".moai\\config\\") {
		return true
	}

	// Protect .moai/evolution/ from template deployment (R1: Directory Protection).
	// All evolution data (learnings, new-skills, telemetry) must survive moai update.
	if strings.HasPrefix(path, ".moai/evolution/") || strings.HasPrefix(path, ".moai\\evolution\\") {
		return true
	}

	// Check if path is in .claude directory
	if !strings.Contains(path, ".claude") {
		return false
	}

	// Split by both '/' and '\' for cross-platform compatibility.
	// Template manifests always use '/' but filepath.Separator is '\' on Windows.
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	for i, part := range parts {
		switch part {
		case "skills", "rules", "commands", "output-styles", "hooks":
			// Check if the next directory starts with "moai-"
			if i+1 < len(parts) {
				itemName := parts[i+1]
				return strings.HasPrefix(itemName, "moai-") || strings.HasPrefix(itemName, "moai")
			}
		case "agents":
			// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: MoAI system agents live in
			// .claude/agents/{core,expert,meta}/ (3 domain subfolders).
			// .claude/agents/harness/ is user-owned per CLAUDE.local.md §24.4 and
			// SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-002 (preserved, not managed).
			// Legacy `agents/moai` directory was removed; only new domain subfolders
			// + `moai-` prefix names are MoAI-managed.
			if i+1 < len(parts) {
				itemName := parts[i+1]
				switch itemName {
				case "core", "expert", "meta":
					return true
				}
				return strings.HasPrefix(itemName, "moai-") || strings.HasPrefix(itemName, "moai")
			}
		}
	}

	return false
}
func GetProjectConfigVersion(projectRoot string) (string, error) {
	configPath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)

	// Check file size before reading to prevent DoS
	info, err := os.Stat(configPath)
	if err != nil {
		// If config file doesn't exist, return "0.0.0" to force update
		if os.IsNotExist(err) {
			return "0.0.0", nil
		}
		return "", fmt.Errorf("stat config file: %w", err)
	}

	// Reject files larger than 10MB
	if info.Size() > MaxConfigSize {
		return "", fmt.Errorf("config file too large: %d bytes (max: %d)", info.Size(), MaxConfigSize)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("read config file: %w", err)
	}

	// Parse YAML to extract moai.template_version
	var config struct {
		Moai struct {
			TemplateVersion string `yaml:"template_version"`
		} `yaml:"moai"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("parse config YAML: %w", err)
	}

	// If template_version is not set, return "0.0.0" to force update
	if config.Moai.TemplateVersion == "" {
		return "0.0.0", nil
	}

	return config.Moai.TemplateVersion, nil
}
