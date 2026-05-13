package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// SentinelWorktreeRequired is emitted by workflow lint when write-heavy role profiles lack isolation: worktree
	SentinelWorktreeRequired = "ORC_WORKTREE_REQUIRED"
)

// WorkflowLintViolation represents a workflow.yaml lint rule violation
type WorkflowLintViolation struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Message  string `json:"message"`
}

// workflowConfig represents the relevant subset of workflow.yaml for linting
type workflowConfig struct {
	Workflow struct {
		Team struct {
			RoleProfiles map[string]struct {
				Mode      string `yaml:"mode"`
				Model     string `yaml:"model"`
				Isolation string `yaml:"isolation"`
			} `yaml:"role_profiles"`
		} `yaml:"team"`
	} `yaml:"workflow"`
}

// runWorkflowLint validates .moai/config/sections/workflow.yaml
// Returns errLintViolations when violations are found
func runWorkflowLint(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	workflowPath := filepath.Join(cwd, defs.MoAIDir, "config", "sections", "workflow.yaml")

	cfg, err := loadWorkflowYAML(workflowPath)
	if err != nil {
		return fmt.Errorf("failed to load workflow.yaml: %w", err)
	}

	violations := validateRoleProfiles(cfg)

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Printf("error: %s: %s\n", v.Rule, v.Message)
		}
		return errLintViolations
	}

	fmt.Println("workflow lint: no violations found")
	return nil
}

// validateRoleProfiles checks role_profiles.{implementer,tester,designer}.isolation == "worktree"
func validateRoleProfiles(cfg *workflowConfig) []WorkflowLintViolation {
	var violations []WorkflowLintViolation

	writeHeavyRoles := []string{"implementer", "tester", "designer"}

	for _, role := range writeHeavyRoles {
		profile, exists := cfg.Workflow.Team.RoleProfiles[role]
		if !exists {
			continue
		}

		if profile.Isolation != "worktree" {
			violations = append(violations, WorkflowLintViolation{
				Rule:     SentinelWorktreeRequired,
				Severity: "error",
				Path:     fmt.Sprintf("workflow.team.role_profiles.%s.isolation", role),
				Expected: "worktree",
				Actual:   profile.Isolation,
				Message:  fmt.Sprintf("role_profiles.%s.isolation must be 'worktree' (got '%s') (SPEC-V3R2-ORC-004 %s)", role, profile.Isolation, SentinelWorktreeRequired),
			})
		}
	}

	return violations
}

// loadWorkflowYAML reads and parses .moai/config/sections/workflow.yaml into a typed struct
func loadWorkflowYAML(path string) (*workflowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow.yaml: %w", err)
	}

	var cfg workflowConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse workflow.yaml: %w", err)
	}

	return &cfg, nil
}

// workflowLintCmd is the cobra command for workflow linting
var workflowLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate workflow.yaml configuration",
	Long:  "Check workflow.yaml role_profiles for write-heavy agents missing isolation: worktree",
	RunE:  runWorkflowLint,
	GroupID: "tools",
}

func init() {
	// Create workflow command group
	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow configuration commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		GroupID: "tools",
	}
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowLintCmd)
}
