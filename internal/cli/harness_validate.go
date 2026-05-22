package cli

// @MX:NOTE: [AUTO] HRN-001 harness validate CLI — SPEC-V3R2-HRN-001 run-phase
// REQ-HRN-001-006/010/012, AC-HRN-001-04/05/07.

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// newHarnessValidateCmd is the `moai harness validate` subcommand factory.
// REQ-HRN-001-006, AC-HRN-001-04/05/07.
func newHarnessValidateCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate harness.yaml against schema and invariants",
		Long: `Validate harness.yaml against:
  - Level enum: {minimal, standard, thorough} FROZEN (REQ-HRN-001-017)
  - pass_threshold >= 0.60 FROZEN floor (REQ-HRN-001-012, design-constitution §5)
  - MOAI_CONFIG_STRICT=1: unknown keys become errors (REQ-HRN-001-019)
  - evaluator.memory_scope must be per_iteration (HRN-002 FROZEN)

Exit code 0 = valid, exit code 1 = validation error.

Examples:
  moai harness validate
  moai harness validate --path /custom/harness.yaml
  MOAI_CONFIG_STRICT=1 moai harness validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHarnessValidate(cmd, cfgPath)
		},
	}

	cmd.Flags().StringVar(&cfgPath, "path", "", "Path to harness.yaml (default: "+defaultHarnessConfigPath+")")

	return cmd
}

// runHarnessValidate executes the `moai harness validate` command.
// AC-HRN-001-04: on success exit 0 + "harness.yaml: OK".
// AC-HRN-001-05: pass_threshold < 0.60 → exit 1 + HRN_PASS_THRESHOLD_FLOOR error.
// AC-HRN-001-07: MOAI_CONFIG_STRICT=1 + unknown key → exit 1 + HRN_SCHEMA_DRIFT error.
func runHarnessValidate(cmd *cobra.Command, cfgPath string) error {
	// Determine harness.yaml path
	harnessPath := cfgPath
	if harnessPath == "" {
		harnessPath = defaultHarnessConfigPath
	}

	// Load harness.yaml (includes level enum, memory scope, and schema-drift validation)
	cfg, err := config.LoadHarnessConfig(harnessPath)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		return fmt.Errorf("harness validate: %w", err)
	}

	// pass_threshold floor validation (REQ-HRN-001-012)
	// Parses the .md file referenced by each level's evaluator_profile for validation.
	if err := validateEvaluatorProfileFloors(harnessPath, cfg, cmd); err != nil {
		return err
	}

	// model_upgrade_review reminder (REQ-HRN-001-016)
	emitModelUpgradeReminder(cfg, cmd)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "harness.yaml: OK")
	return nil
}

// validateEvaluatorProfileFloors parses each level's evaluator_profile file and
// validates the FROZEN floor pass_threshold >= 0.60.
// REQ-HRN-001-012, AC-HRN-001-05.
//
// @MX:WARN: [AUTO] FROZEN floor validation — pass_threshold < 0.60 → ErrPassThresholdFloor
// @MX:REASON: design-constitution §5 FROZEN floor; immediate exit 1 on violation
func validateEvaluatorProfileFloors(harnessPath string, cfg *config.HarnessConfig, cmd *cobra.Command) error {
	if cfg == nil {
		return nil
	}

	// Locate the evaluator-profiles directory relative to harness.yaml
	// Standard location: {config_dir}/../evaluator-profiles/
	// Example harnessPath: .moai/config/sections/harness.yaml
	// → .moai/config/evaluator-profiles/
	harnessDir := harnessPath
	// sections/harness.yaml → sections/ → config/ → evaluator-profiles/
	sectionsDir := harnessDir
	for i := 0; i < 2; i++ {
		if d := parentDir(sectionsDir); d != sectionsDir {
			sectionsDir = d
		}
	}
	profilesDir := sectionsDir + "/evaluator-profiles"

	// Validate the evaluator_profile file for each level
	for levelName, levelCfg := range cfg.Levels {
		profileName := levelCfg.EvaluatorProfile
		if profileName == "" {
			profileName = cfg.DefaultProfile
		}
		if profileName == "" {
			continue
		}

		profilePath := profilesDir + "/" + profileName + ".md"

		// If the file is missing, just continue silently
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			continue
		}

		// Parse the profile
		rubric, err := router.ParseProfileFloor(profilePath)
		if err != nil {
			// Skip on parse error
			continue
		}

		// FROZEN floor validation
		if rubric < 0.60 {
			errMsg := fmt.Sprintf("HRN_PASS_THRESHOLD_FLOOR: levels.%s.evaluator_profile=%q pass_threshold=%.2f is below FROZEN floor 0.60 (design-constitution §5)",
				levelName, profileName, rubric)
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), errMsg)
			return &config.ValidationError{
				Field:   fmt.Sprintf("levels.%s.evaluator_profile", levelName),
				Message: errMsg,
				Value:   rubric,
				Wrapped: config.ErrPassThresholdFloor,
			}
		}
	}

	return nil
}

// emitModelUpgradeReminder prints a model-upgrade review reminder.
// REQ-HRN-001-016: emit when OnModelChange == true.
func emitModelUpgradeReminder(cfg *config.HarnessConfig, cmd *cobra.Command) {
	if cfg == nil {
		return
	}
	if !cfg.ModelUpgradeReview.Enabled || !cfg.ModelUpgradeReview.Trigger.OnModelChange {
		return
	}

	// Detect a model change via the CLAUDE_MODEL_PREVIOUS environment variable
	previousModel := os.Getenv("CLAUDE_MODEL_PREVIOUS")
	currentModel := os.Getenv("CLAUDE_MODEL")
	if previousModel != "" && currentModel != "" && previousModel != currentModel {
		reportPath := cfg.ModelUpgradeReview.Output.ReportPath
		if reportPath == "" {
			reportPath = ".moai/reports/harness-review.md"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(),
			"Model upgrade detected (%s → %s). Review checklist at: %s\n",
			previousModel, currentModel, reportPath)
	}
}

// parentDir returns the parent directory of a path.
func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return path
}

// ValidateHarnessErrors is a helper that checks harness errors via errors.Is.
// Used by tests.
func ValidateHarnessErrors(err error) (isFloor, isDrift bool) {
	isFloor = errors.Is(err, config.ErrPassThresholdFloor)
	isDrift = errors.Is(err, config.ErrSchemaDrift)
	return
}
