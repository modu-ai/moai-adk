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
	"sort"
	"strings"

	"github.com/spf13/cobra"

	harnesscli "github.com/modu-ai/moai-adk/internal/cli/harness"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// defaultHarnessConfigPath is the default harness.yaml path.
// References the same path as the harnessConfigPath constant at internal/cli/harness.go:41.
const defaultHarnessConfigPath = ".moai/config/sections/harness.yaml"

// harnessRouteJSONOutput is the --json output schema.
// REQ-HRN-001-011, AC-HRN-001-06.
type harnessRouteJSONOutput struct {
	Level            string           `json:"level"`
	Rationale        router.Rationale `json:"rationale"`
	Effort           string           `json:"effort"`
	EvaluatorProfile string           `json:"evaluator_profile"`
	SprintContract   bool             `json:"sprint_contract"`
	PlanAudit        bool             `json:"plan_audit"`
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
// SPEC-HARNESS-LOOP-REPAIR-001 AC-HLR-012: the Long description is derived from
// the registered subcommands (cmd.Commands()) AFTER every AddCommand call, so
// the verb enumeration is a round trip from the command table — adding a verb
// via AddCommand automatically documents it, and a verb removed from AddCommand
// automatically disappears from --help. A hand-authored static list that drifts
// from the AddCommand calls is a token-presence violation (acceptance.md §A
// rule 2); the dynamic derivation is the load-bearing fix.
//
// @MX:ANCHOR: [AUTO] V3R5 harness command factory (route/validate + 8 lifecycle/proposal verbs)
// @MX:REASON: fan_in >= 4: root.go registration, harness_route_test.go, harness_test.go, harness_mute_test.go, AC-HRA-009 + AC-HLR-012 verification
func newHarnessRouterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harness",
		Short: "Harness routing, validation, and learning subsystem management",
		// Long is populated AFTER all AddCommand calls below via
		// buildHarnessRouterLong(cmd), so it enumerates every registered verb
		// by derivation (AC-HLR-012 round trip).
	}

	// --project-root flag (shared by all lifecycle/proposal subcommands)
	cmd.PersistentFlags().String("project-root", "", "project root path (default: current directory)")

	// V3R2-HRN-001 routing verbs.
	cmd.AddCommand(newHarnessRouteCmd())
	cmd.AddCommand(newHarnessValidateCmd())

	// V3R5 lifecycle verbs (un-retired per plan.md §6.4).
	cmd.AddCommand(newHarnessStatusCmd())
	// Failure-signature clustering read surface (SPEC-DIVECC-OBSERVABILITY-LOOP-001).
	cmd.AddCommand(newHarnessClustersCmd())
	cmd.AddCommand(newHarnessApplyCmd())
	cmd.AddCommand(newHarnessRollbackCmd())
	cmd.AddCommand(newHarnessDisableCmd())

	// V3R5 M4 proposal-management verbs (REQ-HRA-033, REQ-HRA-036).
	cmd.AddCommand(newHarnessMuteCmd())
	cmd.AddCommand(newHarnessMuteListCmd())
	cmd.AddCommand(newHarnessUnmuteCmd())
	cmd.AddCommand(newHarnessVerifyCmd())

	// SPEC-V3R6-HARNESS-PROPOSAL-GEN-001: V3R4 self-evolving harness loop
	// closure — `moai harness propose` consumes tier-promotions.jsonl and
	// emits draft SPEC proposals. The factory lives in package
	// internal/cli/harness/ so the C-HRA-008-class boundary guard
	// (TestPropose_NoAskUserQuestion) can scan a contained directory.
	cmd.AddCommand(harnesscli.NewProposeCmd())

	// SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001: `moai harness install` is the
	// live call path that wires the previously-orphaned InjectMarker (layer3)
	// + ScaffoldHarnessDir (layer5) installers so a generated harness actually
	// auto-triggers. The factory lives in the same internal/cli/harness/
	// package, sharing the TestPropose_NoAskUserQuestion boundary guard.
	cmd.AddCommand(harnesscli.NewInstallCmd())

	// SPEC-HARNESS-APPLY-EXECUTE-001: `moai harness execute` is the opt-in Go
	// apply path — the FIRST production caller of Applier.Apply(), activating the
	// dormant regression-gate + outcome-capture pipeline so the first apply-outcome
	// telemetry is generated. The factory lives in the same internal/cli/harness/
	// package, sharing the TestPropose_NoAskUserQuestion boundary guard. The `apply
	// --execute` UX delegates to this same RunExecute (see newHarnessApplyCmd).
	cmd.AddCommand(harnesscli.NewExecuteCmd())

	// SPEC-HARNESS-LOOP-REPAIR-001 M2-1: `moai harness promote` routes a
	// proposalgen discovery draft to its designed consumer — manager-spec SPEC
	// authoring — by materialising a SPEC skeleton carrying the draft ID as
	// provenance and moving the draft out of the pending queue. The factory lives
	// in the same boundary-guarded package.
	cmd.AddCommand(harnesscli.NewPromoteCmd())

	// SPEC-V3R6-HARNESS-V4-001 M4: v4 harness lifecycle verbs (list/edit/remove).
	// These enumerate / edit / atomically-remove harness-v4 entries under
	// .claude/commands/harness/. They share the same boundary-guarded package
	// (TestPropose_NoAskUserQuestion scans this directory).
	cmd.AddCommand(harnesscli.NewHarnessV4ListCmd())
	cmd.AddCommand(harnesscli.NewHarnessV4EditCmd())
	cmd.AddCommand(harnesscli.NewHarnessV4RemoveCmd())

	// SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-005: v4 reference-integrity smoke
	// gate (`moai harness doctor`). Distinct from the legacy `moai doctor`
	// learning-loop 5-layer check.
	cmd.AddCommand(harnesscli.NewHarnessDoctorCmd())

	// SPEC-HARNESS-EVOLVE-001: routing observation ledger (Loop 0) — the
	// `moai harness ledger record|evidence|list` recording surface. A separate
	// observation subject from the usage-log observer (REQ-HEV-009).
	cmd.AddCommand(newHarnessLedgerCmd())

	// AC-HLR-012: derive the Long description from the registered subcommands so
	// the verb enumeration is a round trip from the command table (no hand-
	// authored list that can drift from the AddCommand calls).
	cmd.Long = buildHarnessRouterLong(cmd)

	return cmd
}

// buildHarnessRouterLong derives the harness parent command's Long description
// from its registered subcommands (cmd.Commands()). Each registered verb appears
// as `  <Use>  <Short>` on its own line, so adding/removing a verb via
// AddCommand automatically updates `moai harness --help` — the enumeration
// cannot drift from the command table (AC-HLR-012 round trip).
//
// The verbs are sorted by Use for deterministic output (independent of cobra's
// EnableCommandSorting setting). Only the first whitespace-delimited token of
// each Use is shown, so a Use like "edit <name>" renders as "edit".
func buildHarnessRouterLong(cmd *cobra.Command) string {
	subs := cmd.Commands()
	// Defensive copy + sort by Use for deterministic output.
	verbs := make([]*cobra.Command, len(subs))
	copy(verbs, subs)
	sort.Slice(verbs, func(i, j int) bool {
		return useFirstToken(verbs[i].Use) < useFirstToken(verbs[j].Use)
	})

	// Compute the longest Use token for column alignment.
	maxUse := 0
	for _, sub := range verbs {
		if w := len(useFirstToken(sub.Use)); w > maxUse {
			maxUse = w
		}
	}

	var b strings.Builder
	b.WriteString("Harness commands for SPEC complexity routing and learning subsystem management.\n\n")
	b.WriteString("Registered verbs (auto-derived from the command tree — adding or removing a\n")
	b.WriteString("verb via AddCommand automatically updates this list; AC-HLR-012 round trip):\n")
	for _, sub := range verbs {
		use := useFirstToken(sub.Use)
		// Pad the Use column so the Short descriptions align.
		pad := maxUse - len(use)
		b.WriteString("  ")
		b.WriteString(use)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString("  ")
		b.WriteString(sub.Short)
		b.WriteString("\n")
	}
	b.WriteString("\nNote: SPEC-V3R5-HARNESS-AUTONOMY-001 supersedes the lifecycle CLI retirement\n")
	b.WriteString("previously declared by SPEC-V3R4-HARNESS-001.")
	return b.String()
}

// useFirstToken returns the first whitespace-delimited token of a cobra Use
// string (the verb name), dropping any argument placeholders (e.g. "edit <name>"
// → "edit").
func useFirstToken(use string) string {
	for i, r := range use {
		if r == ' ' || r == '\t' {
			return use[:i]
		}
	}
	return use
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
