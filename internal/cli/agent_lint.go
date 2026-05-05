package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// LintSeverity represents the severity level of a lint violation.
type LintSeverity string

const (
	// SeverityError is a critical issue that must be fixed.
	SeverityError LintSeverity = "error"
	// SeverityWarning is a non-critical issue that should be fixed.
	SeverityWarning LintSeverity = "warning"
)

// LintViolation represents a single lint rule violation.
type LintViolation struct {
	Rule    string       `json:"rule"`
	Severity LintSeverity `json:"severity"`
	File    string       `json:"file"`
	Line    int          `json:"line"`
	Message string       `json:"message"`
}

// LintOutput is the JSON output format for the lint command.
type LintOutput struct {
	Version  string         `json:"version"`
	Summary  LintSummary    `json:"summary"`
	Violations []LintViolation `json:"violations"`
}

// LintSummary contains summary statistics.
type LintSummary struct {
	Total    int `json:"total"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

// AgentFrontmatter represents the YAML frontmatter of an agent file.
type AgentFrontmatter struct {
	Name           string            `yaml:"name"`
	Tools          string            `yaml:"tools"`
	Skills         []string          `yaml:"skills"`
	Hooks          map[string][]Hook `yaml:"hooks"`
	Effort         string            `yaml:"effort"`
	Isolation      string            `yaml:"isolation"`
	PermissionMode string            `yaml:"permissionMode"`
}

// Hook represents a single hook configuration.
type Hook struct {
	Matcher string   `yaml:"matcher"`
	Hooks   []SubHook `yaml:"hooks"`
}

// SubHook represents a hook command.
type SubHook struct {
	Command string `yaml:"command"`
}

var agentLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint agent definition files",
	Long: `Validate agent definition files (.claude/agents/moai/*.md) against common issues.

Lint Rules:
  LR-01: Reject literal AskUserQuestion in body text (excluding code blocks)
  LR-02: Reject Agent token in tools: CSV list
  LR-03: Error on missing effort: field (promoted from warning per SPEC-V3R2-ORC-003)
  LR-04: Reject dead hook entries (matcher tool absent from tools:)
  LR-05: Error on missing isolation: worktree for write-heavy role profiles and standalone agents (SPEC-V3R2-ORC-004)
  LR-06: Warn on --deepthink boilerplate in description (error with --strict)
  LR-07: Reject duplicate Skeptical-Evaluator Mandate blocks
  LR-08: Warn on skill-preload drift within same category (error with --strict)
  LR-09: Reject isolation: worktree on read-only agents (permissionMode: plan) (SPEC-V3R2-ORC-004)
  LR-10: Reject static team-* agent files (dynamic generation only, SPEC-V3R2-ORC-005)

Exit Codes:
  0: No violations found
  1: Violations found
  2: Malformed frontmatter
  3: IO error`,
	RunE: runAgentLint,
}

func init() {
	// Create agent command group if it doesn't exist
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent management commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		GroupID: "tools",
	}
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentLintCmd)

	agentLintCmd.Flags().String("path", "", "Path to agent directory (default: .claude/agents/moai/ and internal/template/templates/.claude/agents/moai/)")
	agentLintCmd.Flags().String("format", "text", "Output format: text or json")
	agentLintCmd.Flags().Bool("strict", false, "Promote warnings to errors")
}

func runAgentLint(cmd *cobra.Command, _ []string) error {
	customPath := getStringFlag(cmd, "path")
	format := getStringFlag(cmd, "format")
	strict := getBoolFlag(cmd, "strict")

	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", format)
	}

	// Determine scan paths
	var scanPaths []string
	if customPath != "" {
		scanPaths = []string{customPath}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		scanPaths = []string{
			filepath.Join(cwd, defs.ClaudeDir, defs.AgentsMoaiSubdir),
			filepath.Join(cwd, "internal", "template", "templates", defs.ClaudeDir, defs.AgentsMoaiSubdir),
		}
	}

	// Collect all agent files
	var agentFiles []string
	for _, scanPath := range scanPaths {
		files, err := filepath.Glob(filepath.Join(scanPath, "*.md"))
		if err != nil {
			return fmt.Errorf("glob pattern failed: %w", err)
		}
		agentFiles = append(agentFiles, files...)
	}

	if len(agentFiles) == 0 {
		out := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(out, "No agent files found in %s\n", strings.Join(scanPaths, ", "))
		return nil
	}

	// Run all lint rules
	var allViolations []LintViolation
	for _, file := range agentFiles {
		violations, err := lintAgentFile(file, strict)
		if err != nil {
			// Return exit code 2 for malformed frontmatter
			return fmt.Errorf("lint %s: %w", file, err)
		}
		allViolations = append(allViolations, violations...)
	}

	// Run LR-07 (duplicate mandate blocks) across all files
	dupViolations := checkDuplicateMandateBlocks(agentFiles)
	allViolations = append(allViolations, dupViolations...)

	// Run LR-08 (skill preload drift) across all files
	driftViolations := checkSkillPreloadDrift(agentFiles)
	allViolations = append(allViolations, driftViolations...)

	// Count errors and warnings
	errors := 0
	warnings := 0
	for _, v := range allViolations {
		if v.Severity == SeverityError {
			errors++
		} else {
			warnings++
		}
	}

	// Output results
	out := cmd.OutOrStdout()
	if format == "json" {
		output := LintOutput{
			Version: "1.0.0",
			Summary: LintSummary{
				Total:    len(allViolations),
				Errors:   errors,
				Warnings: warnings,
			},
			Violations: allViolations,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		_, _ = fmt.Fprintln(out, string(data))
	} else {
		// Text format
		if len(allViolations) == 0 {
			_, _ = fmt.Fprintf(out, "%s No violations found\n", cliSuccess.Render("✓"))
		} else {
			for _, v := range allViolations {
				icon := "⚠"
				if v.Severity == SeverityError {
					icon = "✗"
					icon = cliError.Render(icon)
				} else {
					icon = cliWarn.Render(icon)
				}
				_, _ = fmt.Fprintf(out, "%s [%s] %s:%d: %s\n", icon, v.Rule, v.File, v.Line, v.Message)
			}
			_, _ = fmt.Fprintf(out, "\nSummary: %d total (%d errors, %d warnings)\n", len(allViolations), errors, warnings)
		}
	}

	// Set exit code
	if len(allViolations) > 0 {
		// Return exit code 1 for violations found
		return fmt.Errorf("") // Empty error triggers exit code 1
	}

	return nil
}

// lintAgentFile runs all lint rules on a single agent file.
func lintAgentFile(path string, strict bool) ([]LintViolation, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	// Split frontmatter and body
	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("missing frontmatter delimiters (need --- at start and end of frontmatter)")
	}

	frontmatterText := parts[1]
	bodyText := parts[2]

	// Parse frontmatter
	var frontmatter AgentFrontmatter
	if err := parseYAMLFrontmatter(frontmatterText, &frontmatter); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	var violations []LintViolation

	// LR-01: Literal AskUserQuestion in body (excluding code blocks)
	violations = append(violations, checkLiteralAskUserQuestion(path, bodyText)...)

	// LR-02: Agent token in tools: CSV
	violations = append(violations, checkAgentInTools(path, frontmatter)...)

	// LR-03: Missing effort: field
	violations = append(violations, checkMissingEffort(path, frontmatter)...)

	// LR-04: Dead hook entries
	violations = append(violations, checkDeadHooks(path, frontmatter)...)

	// LR-05: Missing isolation: worktree for write-heavy agents
	violations = append(violations, checkMissingIsolation(path, frontmatter)...)

	// LR-06: --deepthink boilerplate in description
	violations = append(violations, checkDeepthinkBoilerplate(path, frontmatterText, strict)...)

	// LR-09: isolation: worktree on read-only agents
	violations = append(violations, checkReadOnlyIsolation(path, frontmatter)...)

	// LR-10: Static team-* agent file detection
	violations = append(violations, checkStaticTeamAgent(path)...)

	return violations, nil
}

// checkLiteralAskUserQuestion checks for LR-01.
func checkLiteralAskUserQuestion(file string, body []byte) []LintViolation {
	var violations []LintViolation

	// Calculate frontmatter offset (count lines before body starts)
	content, _ := os.ReadFile(file)
	frontmatterEnd := bytes.Index(content, []byte("---\n"))
	if frontmatterEnd == -1 {
		frontmatterEnd = bytes.Index(content, []byte("---\r\n"))
	}
	if frontmatterEnd == -1 {
		return violations // Malformed, but let it slide
	}

	// Find the second --- delimiter
	secondDelimiter := bytes.Index(content[frontmatterEnd+3:], []byte("---"))
	if secondDelimiter == -1 {
		return violations
	}

	// Count lines up to the start of body
	linesBeforeBody := bytes.Count(content[:frontmatterEnd+3+secondDelimiter+3], []byte("\n")) + 1

	scanner := bufio.NewScanner(bytes.NewReader(body))
	lineNum := linesBeforeBody
	inCodeBlock := false

	for scanner.Scan() {
		line := scanner.Text()

		// Track code block state
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			lineNum++
			continue
		}

		// Skip checks inside code blocks
		if inCodeBlock {
			lineNum++
			continue
		}

		// Check for literal AskUserQuestion (case-sensitive)
		if strings.Contains(line, "AskUserQuestion") {
			violations = append(violations, LintViolation{
				Rule:     "LR-01",
				Severity: SeverityError,
				File:     file,
				Line:     lineNum,
				Message:  "Literal AskUserQuestion found in body text (use AskUserQuestion tool invocation only in orchestrator, never in agent body)",
			})
		}

		lineNum++
	}

	return violations
}

// checkAgentInTools checks for LR-02.
func checkAgentInTools(file string, fm AgentFrontmatter) []LintViolation {
	var violations []LintViolation

	if fm.Tools == "" {
		return violations // No tools list, nothing to check
	}

	// Parse CSV tool list
	tools := strings.Split(fm.Tools, ",")
	for _, tool := range tools {
		tool = strings.TrimSpace(tool)
		if tool == "Agent" {
			// Find line number by searching file
			lineNum := findFrontmatterLine(file, "tools:")
			violations = append(violations, LintViolation{
				Rule:     "LR-02",
				Severity: SeverityError,
				File:     file,
				Line:     lineNum,
				Message:  "Agent tool found in tools: CSV list (subagents cannot spawn sub-subagents - Agent tool is only for orchestrator)",
			})
			break
		}
	}

	return violations
}

// checkMissingEffort checks for LR-03.
func checkMissingEffort(file string, fm AgentFrontmatter) []LintViolation {
	var violations []LintViolation

	if fm.Effort == "" {
		severity := SeverityError

		lineNum := findFrontmatterLine(file, "name:")
		violations = append(violations, LintViolation{
			Rule:     "LR-03",
			Severity: severity,
			File:     file,
			Line:     lineNum,
			Message:  "Missing effort: field in frontmatter (add 'effort: low/medium/high/xhigh' for explicit session effort control)",
		})
	}

	return violations
}

// checkDeadHooks checks for LR-04.
func checkDeadHooks(file string, fm AgentFrontmatter) []LintViolation {
	var violations []LintViolation

	if len(fm.Hooks) == 0 || fm.Tools == "" {
		return violations
	}

	// Parse tools CSV into set
	toolsSet := make(map[string]bool)
	for _, tool := range strings.Split(fm.Tools, ",") {
		toolsSet[strings.TrimSpace(tool)] = true
	}

	// Check each hook's matcher
	for hookType, hookList := range fm.Hooks {
		for _, hook := range hookList {
			if hook.Matcher == "" {
				continue
			}

			// Parse matcher (e.g., "Write|Edit|MultiEdit")
			matcherTools := strings.Split(hook.Matcher, "|")
			for _, matcherTool := range matcherTools {
				matcherTool = strings.TrimSpace(matcherTool)
				if !toolsSet[matcherTool] {
					lineNum := findFrontmatterLine(file, "hooks:")
					violations = append(violations, LintViolation{
						Rule:     "LR-04",
						Severity: SeverityError,
						File:     file,
						Line:     lineNum,
						Message:  fmt.Sprintf("Dead hook entry: %s matcher references tool '%s' absent from tools: list", hookType, matcherTool),
					})
					break // Only report once per hook
				}
			}
		}
	}

	return violations
}

// checkMissingIsolation checks for LR-05.
// Error on missing isolation: worktree for write-heavy role profiles and standalone agents (SPEC-V3R2-ORC-004).
func checkMissingIsolation(file string, fm AgentFrontmatter) []LintViolation {
	var violations []LintViolation

	// Role profiles that require isolation: worktree (team mode)
	rolePatterns := []string{"implementer", "tester", "designer", "specialist"}

	nameLower := strings.ToLower(fm.Name)

	for _, pattern := range rolePatterns {
		if strings.Contains(nameLower, pattern) {
			if fm.Isolation != "worktree" {
				lineNum := findFrontmatterLine(file, "name:")
				violations = append(violations, LintViolation{
					Rule:     "LR-05",
					Severity: SeverityError,
					File:     file,
					Line:     lineNum,
					Message:  fmt.Sprintf("Agent name suggests role profile '%s' but missing 'isolation: worktree' (required for write teammates in team mode)", pattern),
				})
			}
			return violations
		}
	}

	// Write-heavy standalone agents that require isolation: worktree (SPEC-V3R2-ORC-004)
	writeHeavyAgents := []string{
		"manager-cycle", "expert-backend", "expert-frontend",
		"expert-refactoring", "researcher",
	}

	for _, agentName := range writeHeavyAgents {
		if nameLower == agentName {
			if fm.Isolation != "worktree" {
				lineNum := findFrontmatterLine(file, "name:")
				violations = append(violations, LintViolation{
					Rule:     "LR-05",
					Severity: SeverityError,
					File:     file,
					Line:     lineNum,
					Message:  fmt.Sprintf("Write-heavy agent '%s' must have 'isolation: worktree' (SPEC-V3R2-ORC-004)", fm.Name),
				})
			}
			return violations
		}
	}

	return violations
}

// checkDeepthinkBoilerplate checks for LR-06.
func checkDeepthinkBoilerplate(file string, frontmatter []byte, strict bool) []LintViolation {
	var violations []LintViolation

	// Check for --deepthink flag boilerplate in description
	frontmatterStr := string(frontmatter)
	if strings.Contains(frontmatterStr, "--deepthink flag:") {
		severity := SeverityWarning
		if strict {
			severity = SeverityError
		}

		lineNum := findFrontmatterLine(file, "description:")
		violations = append(violations, LintViolation{
			Rule:     "LR-06",
			Severity: severity,
			File:     file,
			Line:     lineNum,
			Message:  "Boilerplate '--deepthink flag:' text in description field (remove redundant activation instructions - sequential thinking is invoked via MCP tools, not description text)",
		})
	}

	return violations
}

// checkReadOnlyIsolation checks for LR-09.
// Rejects isolation: worktree on read-only agents (permissionMode: plan) per SPEC-V3R2-ORC-004.
func checkReadOnlyIsolation(file string, fm AgentFrontmatter) []LintViolation {
	var violations []LintViolation

	if fm.PermissionMode == "plan" && fm.Isolation == "worktree" {
		lineNum := findFrontmatterLine(file, "isolation:")
		violations = append(violations, LintViolation{
			Rule:     "LR-09",
			Severity: SeverityError,
			File:     file,
			Line:     lineNum,
			Message:  "Read-only agent (permissionMode: plan) MUST NOT have 'isolation: worktree' — plan mode already prevents writes (SPEC-V3R2-ORC-004)",
		})
	}

	return violations
}

// checkStaticTeamAgent checks for LR-10.
// Rejects files matching team-*.md pattern in agents directory (SPEC-V3R2-ORC-005).
// v3r2 uses exclusively dynamic team generation; static team-* agent files are prohibited.
func checkStaticTeamAgent(file string) []LintViolation {
	var violations []LintViolation

	base := filepath.Base(file)
	if strings.HasPrefix(base, "team-") && strings.HasSuffix(base, ".md") {
		violations = append(violations, LintViolation{
			Rule:     "LR-10",
			Severity: SeverityError,
			File:     file,
			Line:     1,
			Message:  "Static team-* agent file prohibited (v3r2 uses exclusively dynamic team generation via Agent(subagent_type: \"general-purpose\") with workflow.yaml role_profiles) — ORC_STATIC_TEAM_AGENT_PROHIBITED (SPEC-V3R2-ORC-005)",
		})
	}

	return violations
}

// checkDuplicateMandateBlocks checks for LR-07 across all agent files.
func checkDuplicateMandateBlocks(files []string) []LintViolation {
	var violations []LintViolation

	// Skeptical-Evaluator Mandate block fingerprint
	// Look for 3+ consecutive lines starting with "- " containing evaluation-related keywords
	mandatePattern := regexp.MustCompile(`(?i)^-\s+.*(evaluate|score|rubric|evidence|assess|criteria|performance|quality|security|robustness|scalability)`)

	type mandateLocation struct {
		file string
		line int
	}

	var mandateBlocks []mandateLocation

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(bytes.NewReader(content))
		lineNum := 0
		consecutiveCount := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if mandatePattern.MatchString(line) {
				consecutiveCount++
				if consecutiveCount == 1 {
					// Mark the start of a potential mandate block
					mandateBlocks = append(mandateBlocks, mandateLocation{
						file: file,
						line: lineNum,
					})
				}
			} else {
				consecutiveCount = 0
			}
		}
	}

	// If we found more than 1 mandate block, report duplicates
	if len(mandateBlocks) > 1 {
		for i, loc := range mandateBlocks {
			if i == 0 {
				continue // Skip first occurrence
			}
			violations = append(violations, LintViolation{
				Rule:     "LR-07",
				Severity: SeverityError,
				File:     loc.file,
				Line:     loc.line,
				Message:  fmt.Sprintf("Duplicate Skeptical-Evaluator Mandate block (first occurrence at %s:%d)", mandateBlocks[0].file, mandateBlocks[0].line),
			})
		}
	}

	return violations
}

// checkSkillPreloadDrift checks for LR-08 across all agent files.
func checkSkillPreloadDrift(files []string) []LintViolation {
	var violations []LintViolation

	// Group agents by category
	type agentInfo struct {
		file   string
		skills []string
	}

	categories := map[string][]agentInfo{
		"manager":  {},
		"expert":   {},
		"builder":  {},
		"evaluator": {},
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		parts := bytes.SplitN(content, []byte("---"), 3)
		if len(parts) < 3 {
			continue
		}

		// Parse skills list from YAML frontmatter
		skills := parseSkillsList(parts[1])
		name := parseFieldName(parts[1])

		// Determine category from agent name
		category := ""
		if strings.HasPrefix(name, "manager-") {
			category = "manager"
		} else if strings.HasPrefix(name, "expert-") {
			category = "expert"
		} else if strings.HasPrefix(name, "builder-") {
			category = "builder"
		} else if strings.HasPrefix(name, "evaluator-") {
			category = "evaluator"
		}

		if category != "" {
			categories[category] = append(categories[category], agentInfo{
				file:   file,
				skills: skills,
			})
		}
	}

	// Check for skill preload drift within each category
	for category, agents := range categories {
		if len(agents) < 2 {
			continue // Need at least 2 agents to compare
		}

		// Find baseline skill set (most common)
		skillCounts := make(map[string]int)
		for _, agent := range agents {
			for _, skill := range agent.skills {
				skillCounts[skill]++
			}
		}

		// Check for drift
		for _, agent := range agents {
			for _, skill := range agent.skills {
				if skillCounts[skill] < len(agents) {
					// This skill is not shared by all agents in category
					lineNum := findFrontmatterLine(agent.file, "skills:")
					violations = append(violations, LintViolation{
						Rule:     "LR-08",
						Severity: SeverityWarning,
						File:     agent.file,
						Line:     lineNum,
						Message:  fmt.Sprintf("Skill preload drift in category '%s': %s is not preloaded by all agents (may cause inconsistent context)", category, skill),
					})
				}
			}
		}
	}

	return violations
}

// parseSkillsList extracts the skills list from YAML frontmatter.
func parseSkillsList(frontmatter []byte) []string {
	var skills []string

	lines := strings.Split(string(frontmatter), "\n")
	inSkills := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "skills:") {
			inSkills = true
			continue
		}

		if inSkills {
			if strings.HasPrefix(line, "- ") {
				skill := strings.TrimSpace(strings.TrimPrefix(line, "- "))
				skills = append(skills, skill)
			} else if line != "" && !strings.HasPrefix(line, "#") {
				// End of skills list
				break
			}
		}
	}

	return skills
}

// parseFieldName extracts the name field from YAML frontmatter.
func parseFieldName(frontmatter []byte) string {
	lines := strings.Split(string(frontmatter), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

// parseYAMLFrontmatter parses YAML frontmatter into a struct.
func parseYAMLFrontmatter(data []byte, v interface{}) error {
	// Simple YAML parser for our specific use case
	// We'll parse key-value pairs line by line
	lines := strings.Split(string(data), "\n")

	type fieldSetter interface {
		setField(key, value string)
	}

	if fs, ok := v.(fieldSetter); ok {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			fs.setField(key, value)
		}
	}

	return nil
}

// setField implements fieldSetter for AgentFrontmatter.
func (fm *AgentFrontmatter) setField(key, value string) {
	switch key {
	case "name":
		fm.Name = value
	case "tools":
		fm.Tools = value
	case "effort":
		fm.Effort = value
	case "isolation":
		fm.Isolation = value
	case "permissionMode":
		fm.PermissionMode = value
	case "skills":
		// Skills is a list, but in our simple parser we just note its presence
		// The actual list parsing happens in the main parser
	case "hooks":
		// Hooks parsing is complex, we'll handle it separately
	}
}

// findFrontmatterLine finds the line number of a frontmatter key.
func findFrontmatterLine(file string, key string) int {
	content, err := os.ReadFile(file)
	if err != nil {
		return 1
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.HasPrefix(line, key) {
			return lineNum
		}
	}

	return 1
}
