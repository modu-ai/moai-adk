package cli

// @MX:NOTE: [AUTO] V3R5 harness CLI unified factory — SPEC-V3R5-HARNESS-AUTONOMY-001 §6 + AC-HRA-009
// @MX:NOTE: [AUTO] newHarnessRouterCmd() integrates 8 additional lifecycle/proposal verbs in V3R5
// @MX:WARN: [AUTO] V3R5 supersedes the CLI retirement declared by SPEC-V3R4-HARNESS-001
// @MX:REASON: plan.md §6.4 + AC-HRA-009 (`./moai harness --help | grep ... ≥6 matches`) enforcement

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// defaultHarnessConfigPath is the default harness.yaml path.
// References the same path as the harnessConfigPath constant at internal/cli/harness.go:41.
const defaultHarnessConfigPath = ".moai/config/sections/harness.yaml"

// harnessRouteJSONOutput is the --json output schema.
// REQ-HRN-001-011, AC-HRN-001-06.
type harnessRouteJSONOutput struct {
	Level           string           `json:"level"`
	Rationale       router.Rationale `json:"rationale"`
	Effort          string           `json:"effort"`
	EvaluatorProfile string          `json:"evaluator_profile"`
	SprintContract  bool             `json:"sprint_contract"`
	PlanAudit       bool             `json:"plan_audit"`
}

// newHarnessRouterCmd is the `moai harness` parent command factory (V3R5 unified).
//
// ARCHITECTURE DECISION (Option A — merge into router):
// V3R5-HARNESS-AUTONOMY-001 §6.4 + AC-HRA-009 mandates that the `moai harness` tree
// expose all 10 of the following verbs:
//   - HRN-001 routing verbs: route, validate
//   - V3R5 lifecycle verbs (un-retired): status, apply, rollback, disable
//   - V3R5 proposal-management verbs (new in M4): mute, mute-list, unmute, verify
//
// V3R4-HARNESS-001 previously retired the lifecycle verbs, but V3R5 explicitly
// supersedes that retirement. This factory registers all 10 subcommands under a
// single parent command to satisfy AC-HRA-009
// (`./moai harness --help | grep -E '(status|apply|rollback|disable|mute|verify)'`
// must match at least 6 entries).
//
// A separate newHarnessCmd() (internal/cli/harness.go) is preserved per the
// deprecation marker contract in SPEC-V3R4-HARNESS-001 §2.1 but is no longer
// registered in the root tree (see TestHarnessFactoryStillCompiles).
// After the V3R5 supersedence, TestHarnessRetirement was updated to permit
// lifecycle verb registration.
//
// @MX:ANCHOR: [AUTO] V3R5 harness command factory (route/validate + 8 lifecycle/proposal verbs)
// @MX:REASON: fan_in >= 4: root.go registration, harness_route_test.go, harness_test.go, harness_mute_test.go, AC-HRA-009 verification
func newHarnessRouterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harness",
		Short: "Harness routing, validation, and learning subsystem management",
		Long: `Harness commands for SPEC complexity routing and learning subsystem management.

Routing verbs (SPEC-V3R2-HRN-001):
  route     Route a SPEC to minimal/standard/thorough harness level
  validate  Validate harness.yaml against schema and invariants

Lifecycle verbs (SPEC-V3R5-HARNESS-AUTONOMY-001 §6, un-retired):
  status    Show observation/tier/evolution summary
  apply     Manually trigger 5-Layer pipeline for queued proposal
  rollback  Revert applied evolution (snapshot restore)
  disable   Set learning.enabled: false

Proposal-management verbs (SPEC-V3R5-HARNESS-AUTONOMY-001 §6, new in M4):
  mute       Mute a proposal category (workflow.yaml)
  mute-list  Print current muted categories
  unmute     Remove a category from the mute list
  verify     Verify harness determinism (W4 placeholder)

Note: SPEC-V3R5-HARNESS-AUTONOMY-001 supersedes the lifecycle CLI retirement
that was previously declared by SPEC-V3R4-HARNESS-001. The unified Cobra tree
satisfies AC-HRA-009 (6+ verb surface).`,
	}

	// --project-root flag (shared by all lifecycle/proposal subcommands)
	cmd.PersistentFlags().String("project-root", "", "project root path (default: current directory)")

	// V3R2-HRN-001 routing verbs.
	cmd.AddCommand(newHarnessRouteCmd())
	cmd.AddCommand(newHarnessValidateCmd())

	// V3R5 lifecycle verbs (un-retired per plan.md §6.4).
	cmd.AddCommand(newHarnessStatusCmd())
	cmd.AddCommand(newHarnessApplyCmd())
	cmd.AddCommand(newHarnessRollbackCmd())
	cmd.AddCommand(newHarnessDisableCmd())

	// V3R5 M4 proposal-management verbs (REQ-HRA-033, REQ-HRA-036).
	cmd.AddCommand(newHarnessMuteCmd())
	cmd.AddCommand(newHarnessMuteListCmd())
	cmd.AddCommand(newHarnessUnmuteCmd())
	cmd.AddCommand(newHarnessVerifyCmd())

	return cmd
}

// newHarnessRouteCmd is the `moai harness route` subcommand factory.
// REQ-HRN-001-006/011, AC-HRN-001-02/03/06/09.
func newHarnessRouteCmd() *cobra.Command {
	var (
		specID     string
		jsonOutput bool
		cfgPath    string
		baseDir    string
	)

	cmd := &cobra.Command{
		Use:   "route",
		Short: "Route a SPEC to a harness level",
		Long: `Route a SPEC to minimal, standard, or thorough harness level
based on Complexity Estimator signals (file_count, domain_count, keywords, priority).

Examples:
  moai harness route --spec SPEC-V3R2-ORC-001
  moai harness route --spec SPEC-V3R2-ORC-001 --json
  moai harness route --spec SPEC-V3R2-ORC-001 --path /custom/harness.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHarnessRoute(cmd, specID, jsonOutput, cfgPath, baseDir)
		},
	}

	cmd.Flags().StringVar(&specID, "spec", "", "SPEC ID to route (e.g., SPEC-V3R2-ORC-001)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output routing decision as JSON")
	cmd.Flags().StringVar(&cfgPath, "path", "", "Path to harness.yaml (default: "+defaultHarnessConfigPath+")")
	cmd.Flags().StringVar(&baseDir, "base-dir", "", "Base directory for .moai/specs/ lookup (default: current dir)")

	if err := cmd.MarkFlagRequired("spec"); err != nil {
		panic(fmt.Sprintf("harness route: MarkFlagRequired: %v", err))
	}

	return cmd
}

// runHarnessRoute executes the `moai harness route` command.
func runHarnessRoute(cmd *cobra.Command, specID string, jsonOutput bool, cfgPath string, baseDir string) error {
	// Determine harness.yaml path
	harnessPath := cfgPath
	if harnessPath == "" {
		harnessPath = defaultHarnessConfigPath
	}

	// Load harness.yaml
	cfg, err := config.LoadHarnessConfig(harnessPath)
	if err != nil {
		return fmt.Errorf("harness route: load config: %w", err)
	}

	// Resolve SPEC file path: SPEC-ID → .moai/specs/{SPEC-ID}/spec.md
	specPath, err := resolveSpecPath(specID, baseDir)
	if err != nil {
		return fmt.Errorf("harness route: resolve spec path: %w", err)
	}

	// Perform routing
	r := router.New(cfg)
	level, rationale, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		return fmt.Errorf("harness route: routing failed: %w", err)
	}

	// Determine effort level and evaluator profile
	effort := router.EffortForLevel(level, cfg)
	evaluatorProfile := cfg.DefaultProfile
	sprintContract := false
	planAudit := true

	if levelCfg, ok := cfg.Levels[string(level)]; ok {
		if levelCfg.EvaluatorProfile != "" {
			evaluatorProfile = levelCfg.EvaluatorProfile
		}
		sprintContract = levelCfg.SprintContract
		planAudit = levelCfg.PlanAudit.Enabled
	}

	// Output format
	if jsonOutput {
		output := harnessRouteJSONOutput{
			Level:            string(level),
			Rationale:        rationale,
			Effort:           effort,
			EvaluatorProfile: evaluatorProfile,
			SprintContract:   sprintContract,
			PlanAudit:        planAudit,
		}
		data, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("harness route: json marshal: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	} else {
		// plaintext output
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "SPEC: %s\n", specID)
		_, _ = fmt.Fprintf(w, "Level: %s\n", level)
		_, _ = fmt.Fprintf(w, "Matched Rule: %s\n", rationale.MatchedRule)
		_, _ = fmt.Fprintf(w, "Effort: %s\n", effort)
		_, _ = fmt.Fprintf(w, "Evaluator Profile: %s\n", evaluatorProfile)
		_, _ = fmt.Fprintf(w, "Sprint Contract: %v\n", sprintContract)
		_, _ = fmt.Fprintf(w, "Plan Audit: %v\n", planAudit)
		if len(rationale.Keywords) > 0 {
			_, _ = fmt.Fprintf(w, "Matched Keywords: %v\n", rationale.Keywords)
		}
	}

	return nil
}

// resolveSpecPath determines the spec.md file path from a SPEC-ID string.
// Uses baseDir as the root if provided; otherwise uses the current working directory.
func resolveSpecPath(specID string, baseDir string) (string, error) {
	// Determine base directory
	base := baseDir
	if base == "" {
		var err error
		base, err = os.Getwd()
		if err != nil {
			base = "."
		}
	}

	candidates := []string{
		filepath.Join(base, ".moai", "specs", specID, "spec.md"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("spec file not found for %q; tried: %v", specID, candidates)
}
