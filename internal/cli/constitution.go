package cli

// @MX:NOTE: [AUTO] Constitution management commands for FROZEN/EVOLVABLE zone codification
// @MX:NOTE: [AUTO] SPEC-V3R2-CON-001 implements zone registry with safety gates
// @MX:NOTE: [AUTO] 5-layer safety gate: FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// constitutionRegistryEnvKey is the environment variable name for registry path.
const constitutionRegistryEnvKey = "MOAI_CONSTITUTION_REGISTRY"

// constitutionRegistryRelPath is the project-relative path to the default registry file.
const constitutionRegistryRelPath = ".claude/rules/moai/core/zone-registry.md"

// newConstitutionCmd creates the `moai constitution` root subcommand.
// Follows research.go pattern.
func newConstitutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "constitution",
		Short:   "Manage the zone registry (FROZEN/EVOLVABLE zone codification)",
		Long:    "Zone registry query and validation commands. SPEC-V3R2-CON-001 implementation.",
		GroupID: "tools",
	}
	cmd.AddCommand(newConstitutionListCmd())
	cmd.AddCommand(newConstitutionGuardCmd())
	cmd.AddCommand(newConstitutionAmendCmd())
	return cmd
}

// newConstitutionGuardCmd creates the `moai constitution guard` subcommand.
// Takes a list of changed rule IDs via --violations flag and returns FROZEN zone violation status.
// Implements SPEC-V3R2-CON-001 AC-CON-001-003.
func newConstitutionGuardCmd() *cobra.Command {
	var violationsFlag []string

	cmd := &cobra.Command{
		Use:   "guard",
		Short: "Check for FROZEN zone violations",
		Long:  "Takes a list of changed rule IDs and checks for Frozen zone violations. For CI integration.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory error: %w", err)
			}
			registryPath := resolveRegistryPath(cwd)
			return runConstitutionGuard(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, violationsFlag)
		},
	}

	cmd.Flags().StringSliceVar(&violationsFlag, "violations", nil, "List of changed rule IDs (comma-separated or repeated flag)")
	return cmd
}

// runConstitutionGuard detects Frozen zone violations from changed rule IDs.
// violations: list of changed rule IDs (empty means no violations).
// Returns: error if Frozen zone violation found, nil otherwise.
func runConstitutionGuard(w, wWarn io.Writer, projectDir, registryPath string, violations []string) error {
	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry load error %q: %w", registryPath, err)
	}

	// Print orphan warnings to stderr
	for _, warn := range reg.Warnings {
		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
	}

	// Detect Frozen zone violations from changed IDs
	var frozenViolations []string
	for _, id := range violations {
		rule, ok := reg.Get(id)
		if !ok {
			// ID not in registry is a dangling ref - print warning only
			_, _ = fmt.Fprintf(wWarn, "Warning: dangling reference %q - ID not in registry\n", id)
			continue
		}
		if rule.Zone == constitution.ZoneFrozen {
			frozenViolations = append(frozenViolations, id)
		}
	}

	if len(frozenViolations) > 0 {
		_, _ = fmt.Fprintf(w, "FROZEN zone violation detected (%d): %s\n",
			len(frozenViolations), strings.Join(frozenViolations, ", "))
		return fmt.Errorf("FROZEN zone violation: %s", strings.Join(frozenViolations, ", "))
	}

	_, _ = fmt.Fprintln(w, "constitution guard: OK - No Frozen zone violations")
	return nil
}

// newConstitutionListCmd creates the `moai constitution list` subcommand.
func newConstitutionListCmd() *cobra.Command {
	var zoneFlag string
	var fileFlag string
	var formatFlag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List zone registry entries",
		Long:  "Prints zone registry entries. Filterable via --zone, --file, --format flags.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory error: %w", err)
			}

			registryPath := resolveRegistryPath(cwd)

			var zoneFilter *constitution.Zone
			if zoneFlag != "" {
				z, parseErr := constitution.ParseZone(zoneFlag)
				if parseErr != nil {
					return fmt.Errorf("--zone parse error: %w", parseErr)
				}
				zoneFilter = &z
			}

			return runConstitutionList(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, zoneFilter, fileFlag, formatFlag)
		},
	}

	cmd.Flags().StringVar(&zoneFlag, "zone", "", "Zone filter (frozen|evolvable)")
	cmd.Flags().StringVar(&fileFlag, "file", "", "File path filter (partial match)")
	cmd.Flags().StringVar(&formatFlag, "format", "table", "Output format (table|json)")

	return cmd
}

// resolveRegistryPath determines registry file path by priority.
// Priority: MOAI_CONSTITUTION_REGISTRY env var → CLAUDE_PROJECT_DIR based path → cwd based path.
func resolveRegistryPath(cwd string) string {
	if envPath := os.Getenv(constitutionRegistryEnvKey); envPath != "" {
		return envPath
	}

	if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
		return filepath.Join(projectDir, constitutionRegistryRelPath)
	}

	return filepath.Join(cwd, constitutionRegistryRelPath)
}

// runConstitutionList loads registry and outputs to w.
// Prints warnings to wWarn (stderr) to avoid polluting stdout.
// Test-friendly pure function.
func runConstitutionList(w, wWarn io.Writer, projectDir, registryPath string, zoneFilter *constitution.Zone, fileFilter, format string) error {
	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry load error %q: %w", registryPath, err)
	}

	// Print warnings to stderr (wWarn)
	for _, warn := range reg.Warnings {
		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
	}

	// Apply filters
	entries := reg.Entries
	if zoneFilter != nil {
		entries = reg.FilterByZone(*zoneFilter)
	}
	if fileFilter != "" {
		var filtered []constitution.Rule
		for _, e := range entries {
			if strings.Contains(e.File, fileFilter) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	switch format {
	case "json":
		return renderConstitutionJSON(w, entries)
	default:
		renderConstitutionTable(w, entries)
		return nil
	}
}

// constitutionJSONOutput is the JSON output structure.
type constitutionJSONOutput struct {
	Entries []constitutionJSONEntry `json:"entries"`
}

// constitutionJSONEntry is the entry structure for JSON serialization.
type constitutionJSONEntry struct {
	ID         string `json:"id"`
	Zone       string `json:"zone"`
	File       string `json:"file"`
	Anchor     string `json:"anchor"`
	Clause     string `json:"clause"`
	CanaryGate bool   `json:"canary_gate"`
}

// renderConstitutionJSON outputs entries in JSON format.
func renderConstitutionJSON(w io.Writer, entries []constitution.Rule) error {
	jsonEntries := make([]constitutionJSONEntry, 0, len(entries))
	for _, e := range entries {
		jsonEntries = append(jsonEntries, constitutionJSONEntry{
			ID:         e.ID,
			Zone:       e.Zone.String(),
			File:       e.File,
			Anchor:     e.Anchor,
			Clause:     e.Clause,
			CanaryGate: e.CanaryGate,
		})
	}

	out := constitutionJSONOutput{Entries: jsonEntries}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON serialization error: %w", err)
	}

	_, _ = fmt.Fprintln(w, string(data))
	return nil
}

// renderConstitutionTable outputs entries in table format.
// Clause is truncated to 40 characters without -v option.
func renderConstitutionTable(w io.Writer, entries []constitution.Rule) {
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(w, "No entries.")
		return
	}

	const idWidth = 18
	const zoneWidth = 10
	const fileWidth = 50
	const clauseWidth = 40

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s",
		idWidth, "ID",
		zoneWidth, "Zone",
		fileWidth, "File",
		clauseWidth, "Clause",
	)
	separator := strings.Repeat("-", idWidth+2+zoneWidth+2+fileWidth+2+clauseWidth)

	_, _ = fmt.Fprintln(w, header)
	_, _ = fmt.Fprintln(w, separator)

	for _, e := range entries {
		clause := e.Clause
		if len(clause) > clauseWidth {
			clause = clause[:clauseWidth-3] + "..."
		}
		fileStr := e.File
		if len(fileStr) > fileWidth {
			fileStr = "..." + fileStr[len(fileStr)-(fileWidth-3):]
		}

		line := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s",
			idWidth, e.ID,
			zoneWidth, e.Zone.String(),
			fileWidth, fileStr,
			clauseWidth, clause,
		)
		_, _ = fmt.Fprintln(w, line)
	}

	_, _ = fmt.Fprintf(w, "\nTotal %d entries\n", len(entries))
}

// newConstitutionAmendCmd creates the `moai constitution amend` subcommand.
// SPEC-V3R2-CON-002 implementation. Constitutional amendment via 5-layer safety gate.
func newConstitutionAmendCmd() *cobra.Command {
	var (
		ruleIDFlag    string
		beforeFlag    string
		afterFlag     string
		evidenceFlag  string
		dryRunFlag    bool
		dryRunEnv     = os.Getenv("MOAI_CONSTITUTION_DRY_RUN") == "true"
	)

	cmd := &cobra.Command{
		Use:   "amend",
		Short: "Propose a constitutional amendment with 5-layer safety gate",
		Long: "Execute constitutional amendment proposal. Must pass 5-layer safety gate (FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight) to apply.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory error: %w", err)
			}

			// Validate required flags
			if ruleIDFlag == "" {
				return fmt.Errorf("--rule is required")
			}
			if beforeFlag == "" || afterFlag == "" {
				return fmt.Errorf("--before and --after are required")
			}

			// Environment variable dry-run takes precedence
			dryRun := dryRunFlag || dryRunEnv

			return runConstitutionAmend(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, ruleIDFlag, beforeFlag, afterFlag, evidenceFlag, dryRun)
		},
	}

	cmd.Flags().StringVar(&ruleIDFlag, "rule", "", "Rule ID (CONST-V3R2-NNN) [required]")
	cmd.Flags().StringVar(&beforeFlag, "before", "", "Current clause text [required]")
	cmd.Flags().StringVar(&afterFlag, "after", "", "New clause text [required]")
	cmd.Flags().StringVar(&evidenceFlag, "evidence", "", "Amendment justification (required for Frozen zone)")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Dry-run mode: simulation only without file modifications")

	return cmd
}

// runConstitutionAmend executes the constitutional amendment pipeline.
func runConstitutionAmend(w, wWarn io.Writer, projectDir, ruleID, before, after, evidence string, dryRun bool) error {
	// Load registry
	registryPath := resolveRegistryPath(projectDir)
	registry, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry load error: %w", err)
	}

	// Print warnings
	for _, warn := range registry.Warnings {
		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
	}

	// Verify rule exists
	rule, exists := registry.Get(ruleID)
	if !exists {
		return fmt.Errorf("rule %q not found", ruleID)
	}

	// Before verification (check if matches current clause)
	if rule.Clause != before {
		return fmt.Errorf("clause mismatch: --before differs from current clause\nCurrent: %s\nInput: %s", rule.Clause, before)
	}

	// Create proposal
	proposal := &constitution.AmendmentProposal{
		RuleID:   ruleID,
		Before:   before,
		After:    after,
		Evidence: evidence,
	}

	// Execute pipeline
	pipeline := constitution.NewPipeline()
	log, err := pipeline.Execute(proposal, projectDir, dryRun)
	if err != nil {
		return fmt.Errorf("amendment failed: %w", err)
	}

	// Print results
	if dryRun {
		_, _ = fmt.Fprintln(w, "=== Dry-run Results ===")
		_, _ = fmt.Fprintf(w, "Rule ID: %s\n", log.RuleID)
		_, _ = fmt.Fprintf(w, "Zone: %s\n", log.ZoneAfter)
		_, _ = fmt.Fprintf(w, "Clause Before: %s\n", log.ClauseBefore)
		_, _ = fmt.Fprintf(w, "Clause After: %s\n", log.ClauseAfter)
		_, _ = fmt.Fprintf(w, "Canary Verdict: %s\n", log.CanaryVerdict)
		if len(log.Contradictions) > 0 {
			_, _ = fmt.Fprintln(w, "Contradictions:")
			for _, c := range log.Contradictions {
				_, _ = fmt.Fprintf(w, "  - %s\n", c)
			}
		}
		_, _ = fmt.Fprintln(w, "\nDry-run success: files were not modified.")
	} else {
		_, _ = fmt.Fprintf(w, "Amendment success: %s\n", log.ID)
		_, _ = fmt.Fprintf(w, "Rule %s has been updated.\n", ruleID)
	}

	return nil
}
