package agentlint

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

	"github.com/modu-ai/moai-adk/internal/config"
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

// workflowLintWrapper unmarshals only the workflow: section into the canonical
// config.WorkflowConfig type. The lint reads the file literally (no default
// seeding) so that a missing or misconfigured role_profiles entry surfaces as a
// violation rather than being masked by construction-time defaults.
type workflowLintWrapper struct {
	Workflow config.WorkflowConfig `yaml:"workflow"`
}

// writeHeavyRoles enumerates the team-mode role profiles that MUST use isolation:worktree.
var writeHeavyRoles = []string{"implementer", "tester", "designer"}

// loadWorkflowYAML reads and parses workflow.yaml into the canonical
// config.WorkflowConfig type. Returns exit code 2 error on malformed YAML.
func loadWorkflowYAML(path string) (*config.WorkflowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow.yaml: %w", err)
	}

	var wrapper workflowLintWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse workflow.yaml: %w", err)
	}

	return &wrapper.Workflow, nil
}

// validateRoleProfiles checks that role_profiles.{implementer,tester,designer}.isolation == "worktree".
// Returns a slice of WorkflowLintViolation (one per offending role).
func validateRoleProfiles(cfg *config.WorkflowConfig) []WorkflowLintViolation {
	var violations []WorkflowLintViolation

	if cfg == nil {
		return violations
	}

	profiles := cfg.Team.RoleProfiles
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
// exitCodeError carries a structured exit code so cmd/moai/main.go maps a
// non-default code via the ExitCoder boundary instead of cobra's default exit 1.
// SPEC-CLIFIX-CONTRACT-001 M1 (the agentlint package cannot import the cli root
// package without a cycle, so it owns a local type satisfying the same structural
// ExitCoder interface in cmd/moai/main.go).
type exitCodeError struct {
	code int
	msg  string
}

func (e *exitCodeError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return fmt.Sprintf("workflow lint: exit code %d", e.code)
}

// ExitCode satisfies the ExitCoder interface in cmd/moai/main.go.
func (e *exitCodeError) ExitCode() int { return e.code }

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
		// Exit 2: Malformed YAML. Returned as an ExitCoder so main.go maps
		// exit 2 and defers run (SPEC-CLIFIX-CONTRACT-001 M1).
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error: malformed workflow.yaml: %v\n", err)
		return &exitCodeError{code: 2, msg: fmt.Sprintf("malformed workflow.yaml: %v", err)}
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

// WorkflowCmd is the parent "workflow" command group. The parent cli package
// registers it onto rootCmd; the lint subcommand is wired onto it in init() below.
var WorkflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Workflow configuration commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	GroupID: "tools",
}

func init() {
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

	WorkflowCmd.AddCommand(workflowLintCmd)
}
