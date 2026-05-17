package cli

// @MX:NOTE: [AUTO] WorkflowLintIntent — moai workflow lint validates .moai/config/sections/workflow.yaml
// role_profiles to ensure write-heavy team roles (implementer/tester/designer) declare
// isolation:worktree. Static CI gate for REQ-ORC-004-008 (SPEC-V3R2-ORC-004).

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// WorkflowLintViolation represents a workflow.yaml lint rule violation.
type WorkflowLintViolation struct {
	Rule     string `json:"rule"`     // e.g. "ORC_WORKTREE_REQUIRED"
	Severity string `json:"severity"` // "error" | "warning"
	Path     string `json:"path"`     // YAML path, e.g. "workflow.team.role_profiles.implementer.isolation"
	Expected string `json:"expected"` // expected value
	Actual   string `json:"actual"`   // actual value
	Message  string `json:"message"`
}

// WorkflowLintOutput is the JSON output format for the workflow lint command.
type WorkflowLintOutput struct {
	Version    string                  `json:"version"`
	Summary    WorkflowLintSummary     `json:"summary"`
	Violations []WorkflowLintViolation `json:"violations"`
}

// WorkflowLintSummary contains summary statistics.
type WorkflowLintSummary struct {
	Total  int `json:"total"`
	Errors int `json:"errors"`
}

// workflowConfig is the internal type matching the relevant subset of workflow.yaml.
type workflowConfig struct {
	Workflow struct {
		Team struct {
			RoleProfiles map[string]workflowRoleProfile `yaml:"role_profiles"`
		} `yaml:"team"`
	} `yaml:"workflow"`
}

// workflowRoleProfile captures the fields relevant to isolation enforcement.
type workflowRoleProfile struct {
	Mode        string `yaml:"mode"`
	Model       string `yaml:"model"`
	Isolation   string `yaml:"isolation"`
	Description string `yaml:"description"`
}

// writeHeavyRoles enumerates the team-mode role profiles that MUST use isolation:worktree.
var writeHeavyRoles = []string{"implementer", "tester", "designer"}

// loadWorkflowYAML reads and parses workflow.yaml into a typed struct.
// Returns exit code 2 error on malformed YAML.
func loadWorkflowYAML(path string) (*workflowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow.yaml: %w", err)
	}

	var cfg workflowConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse workflow.yaml: %w", err)
	}

	return &cfg, nil
}

// validateRoleProfiles checks that role_profiles.{implementer,tester,designer}.isolation == "worktree".
// Returns a slice of WorkflowLintViolation (one per offending role).
func validateRoleProfiles(cfg *workflowConfig) []WorkflowLintViolation {
	var violations []WorkflowLintViolation

	if cfg == nil {
		return violations
	}

	profiles := cfg.Workflow.Team.RoleProfiles
	if profiles == nil {
		// No role_profiles defined — each write-heavy role is missing
		for _, role := range writeHeavyRoles {
			violations = append(violations, WorkflowLintViolation{
				Rule:     SentinelWorktreeRequired,
				Severity: string(SeverityError),
				Path:     fmt.Sprintf("workflow.team.role_profiles.%s.isolation", role),
				Expected: "worktree",
				Actual:   "(missing)",
				Message:  fmt.Sprintf("role_profiles.%s.isolation must be 'worktree' (got '(missing)') (SPEC-V3R2-ORC-004 %s)", role, SentinelWorktreeRequired),
			})
		}
		return violations
	}

	for _, role := range writeHeavyRoles {
		profile, exists := profiles[role]
		if !exists {
			// Role not defined — flag as missing
			violations = append(violations, WorkflowLintViolation{
				Rule:     SentinelWorktreeRequired,
				Severity: string(SeverityError),
				Path:     fmt.Sprintf("workflow.team.role_profiles.%s.isolation", role),
				Expected: "worktree",
				Actual:   "(not defined)",
				Message:  fmt.Sprintf("role_profiles.%s is not defined; write-heavy roles MUST declare isolation:worktree (SPEC-V3R2-ORC-004 %s)", role, SentinelWorktreeRequired),
			})
			continue
		}
		if profile.Isolation != "worktree" {
			actual := profile.Isolation
			if actual == "" {
				actual = "(empty)"
			}
			violations = append(violations, WorkflowLintViolation{
				Rule:     SentinelWorktreeRequired,
				Severity: string(SeverityError),
				Path:     fmt.Sprintf("workflow.team.role_profiles.%s.isolation", role),
				Expected: "worktree",
				Actual:   actual,
				Message:  fmt.Sprintf("role_profiles.%s.isolation must be 'worktree' (got '%s') (SPEC-V3R2-ORC-004 %s)", role, actual, SentinelWorktreeRequired),
			})
		}
	}

	return violations
}

// runWorkflowLint validates .moai/config/sections/workflow.yaml.
// Returns errLintViolations (cobra-friendly) when violations are found.
func runWorkflowLint(cmd *cobra.Command, _ []string) error {
	format := getStringFlag(cmd, "format")

	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", format)
	}

	// Locate workflow.yaml
	customPath := getStringFlag(cmd, "path")
	var workflowPath string
	if customPath != "" {
		workflowPath = customPath
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		workflowPath = filepath.Join(cwd, ".moai", "config", "sections", "workflow.yaml")
	}

	cfg, err := loadWorkflowYAML(workflowPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Exit 3: IO error (file not found)
			return fmt.Errorf("workflow.yaml not found at %s: %w", workflowPath, err)
		}
		// Exit 2: Malformed YAML
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error: malformed workflow.yaml: %v\n", err)
		os.Exit(2)
	}

	violations := validateRoleProfiles(cfg)

	errorCount := 0
	for _, v := range violations {
		if v.Severity == string(SeverityError) {
			errorCount++
		}
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		output := WorkflowLintOutput{
			Version: "1.0",
			Summary: WorkflowLintSummary{
				Total:  len(violations),
				Errors: errorCount,
			},
			Violations: violations,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		_, _ = fmt.Fprintln(out, string(data))
	} else {
		if len(violations) == 0 {
			_, _ = fmt.Fprintf(out, "%s No violations found\n", cliSuccess.Render("✓"))
		} else {
			for _, v := range violations {
				icon := cliError.Render("✗")
				_, _ = fmt.Fprintf(out, "%s [%s] %s: %s\n", icon, v.Rule, v.Path, v.Message)
			}
			_, _ = fmt.Fprintf(out, "\nSummary: %d total (%d errors)\n", len(violations), errorCount)
		}
	}

	if len(violations) > 0 {
		return errLintViolations
	}

	return nil
}

func init() {
	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow configuration commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		GroupID: "tools",
	}
	rootCmd.AddCommand(workflowCmd)

	workflowLintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint workflow configuration",
		Long: `Validate .moai/config/sections/workflow.yaml against SPEC-V3R2-ORC-004 rules.

  ORC_WORKTREE_REQUIRED: role_profiles.{implementer,tester,designer}.isolation must be 'worktree'

Exit Codes:
  0: No violations found
  1: Violations found
  2: Malformed workflow.yaml
  3: IO error`,
		RunE:          runWorkflowLint,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	workflowLintCmd.Flags().String("path", "", "Path to workflow.yaml (default: .moai/config/sections/workflow.yaml)")
	workflowLintCmd.Flags().String("format", "text", "Output format: text or json")

	workflowCmd.AddCommand(workflowLintCmd)
}
