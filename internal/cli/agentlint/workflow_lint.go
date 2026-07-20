package agentlint

// @MX:NOTE: [AUTO] WorkflowLintIntent — moai workflow lint validates .moai/config/sections/workflow.yaml
// model_routing_profiles entries against the closed sets (perfTier / key structure /
// model / effort) by reusing config.ValidateModelRoutingProfiles. Repurposed from the
// retired Agent Teams role-profiles isolation check (SPEC-AGENT-TEAM-RETIRE-001 REQ-ATR-009).

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
	Rule     string `json:"rule"`     // e.g. "MODEL_ROUTING_INVALID"
	Severity string `json:"severity"` // "error" | "warning"
	Path     string `json:"path"`     // YAML path, e.g. "workflow.model_routing_profiles.max.S-run.model"
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
// seeding) so that a misconfigured model_routing_profiles entry surfaces as a
// violation rather than being masked by construction-time defaults.
type workflowLintWrapper struct {
	Workflow config.WorkflowConfig `yaml:"workflow"`
}

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

// validateModelRoutingProfiles checks every workflow.model_routing_profiles
// entry against the closed sets (perfTier {max,medium,low}; key <TIER>-<phase>;
// model {inherit,sonnet,opus,glm}; effort {low,medium,high,xhigh,max}) by
// reusing the canonical config.(*Config).ValidateModelRoutingProfiles
// (SPEC-AGENT-TEAM-RETIRE-001 REQ-ATR-009 — replaces the retired Agent Teams
// role-profiles isolation check). An absent or empty block is valid (every
// lookup falls back). Returns at most one violation: the canonical validator
// reports the first offending location.
func validateModelRoutingProfiles(cfg *config.WorkflowConfig) []WorkflowLintViolation {
	if cfg == nil {
		return nil
	}

	full := &config.Config{Workflow: *cfg}
	err := full.ValidateModelRoutingProfiles()
	if err == nil {
		return nil
	}

	violation := WorkflowLintViolation{
		Rule:     SentinelModelRoutingInvalid,
		Severity: string(SeverityError),
		Path:     "workflow.model_routing_profiles",
		Expected: "perfTier {max,medium,low}; key <TIER>-<phase>; model {inherit,sonnet,opus,glm}; effort {low,medium,high,xhigh,max}",
		Message:  fmt.Sprintf("model_routing_profiles validation failed: %v (SPEC-AGENT-TEAM-RETIRE-001 %s)", err, SentinelModelRoutingInvalid),
	}
	var ve *config.ValidationError
	if errors.As(err, &ve) {
		violation.Path = "workflow." + ve.Field
		violation.Actual = fmt.Sprintf("%v", ve.Value)
	}
	return []WorkflowLintViolation{violation}
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

	violations := validateModelRoutingProfiles(cfg)

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
		Long: `Validate .moai/config/sections/workflow.yaml model_routing_profiles entries.

  MODEL_ROUTING_INVALID: model_routing_profiles entries must satisfy the closed
  sets — perfTier {max,medium,low}; key <TIER>-<phase> with Tier in {S,M,L} and
  Phase in {plan,run,sync,mx}; model {inherit,sonnet,opus,glm}; effort
  {low,medium,high,xhigh,max} (SPEC-AGENT-TEAM-RETIRE-001 REQ-ATR-009)

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
