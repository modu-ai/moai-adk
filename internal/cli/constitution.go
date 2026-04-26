package cli

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

// constitutionRegistryEnvKey is the environment variable name that specifies the registry path.
const constitutionRegistryEnvKey = "MOAI_CONSTITUTION_REGISTRY"

// constitutionRegistryRelPath is the project-relative path to the default registry file.
const constitutionRegistryRelPath = ".claude/rules/moai/core/zone-registry.md"

// newConstitutionCmd creates the `moai constitution` root subcommand.
// Follows the research.go pattern.
func newConstitutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "constitution",
		Short:   "Manage the zone registry (FROZEN/EVOLVABLE zone codification)",
		Long:    "Command for querying and validating the zone registry. Implements SPEC-V3R2-CON-001.",
		GroupID: "tools",
	}
	cmd.AddCommand(newConstitutionListCmd())
	cmd.AddCommand(newConstitutionGuardCmd())
	return cmd
}

// newConstitutionGuardCmd creates the `moai constitution guard` subcommand.
// Receives a list of changed rule IDs via --violations and reports FROZEN zone violations.
// Implements SPEC-V3R2-CON-001 AC-CON-001-003.
func newConstitutionGuardCmd() *cobra.Command {
	var violationsFlag []string

	cmd := &cobra.Command{
		Use:   "guard",
		Short: "Check for FROZEN zone violations",
		Long:  "Receives a list of changed rule IDs and checks for Frozen zone violations. Used for CI integration.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory check error: %w", err)
			}
			registryPath := resolveRegistryPath(cwd)
			return runConstitutionGuard(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, violationsFlag)
		},
	}

	cmd.Flags().StringSliceVar(&violationsFlag, "violations", nil, "List of changed rule IDs (comma-separated or repeated flag)")
	return cmd
}

// runConstitutionGuard detects Frozen zone violations among the changed rule IDs.
// violations: list of changed rule IDs (treated as no violations when empty).
// Returns an error when Frozen zone violations are found, nil otherwise.
func runConstitutionGuard(w, wWarn io.Writer, projectDir, registryPath string, violations []string) error {
	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry load error %q: %w", registryPath, err)
	}

	// Print orphan warnings to stderr.
	for _, warn := range reg.Warnings {
		_, _ = fmt.Fprintf(wWarn, "warning: %s\n", warn)
	}

	// Detect Frozen zone violations among the changed IDs.
	var frozenViolations []string
	for _, id := range violations {
		rule, ok := reg.Get(id)
		if !ok {
			// ID not in registry is a dangling ref — print warning only.
			_, _ = fmt.Fprintf(wWarn, "warning: dangling reference %q - ID not found in registry\n", id)
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

	_, _ = fmt.Fprintln(w, "constitution guard: OK - no Frozen zone violations")
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
		Long:  "Prints zone registry entries. Supports filtering with --zone, --file, and --format flags.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory check error: %w", err)
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

// resolveRegistryPath determines the registry file path according to priority.
// Priority: MOAI_CONSTITUTION_REGISTRY env var → CLAUDE_PROJECT_DIR-relative path → cwd-relative path.
func resolveRegistryPath(cwd string) string {
	if envPath := os.Getenv(constitutionRegistryEnvKey); envPath != "" {
		return envPath
	}

	if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
		return filepath.Join(projectDir, constitutionRegistryRelPath)
	}

	return filepath.Join(cwd, constitutionRegistryRelPath)
}

// runConstitutionList loads the registry and writes its entries to w.
// Warnings are written to wWarn (stderr) to avoid polluting stdout output.
// Pure function, test-friendly.
func runConstitutionList(w, wWarn io.Writer, projectDir, registryPath string, zoneFilter *constitution.Zone, fileFilter, format string) error {
	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry load error %q: %w", registryPath, err)
	}

	// Write warnings to stderr (wWarn).
	for _, warn := range reg.Warnings {
		_, _ = fmt.Fprintf(wWarn, "warning: %s\n", warn)
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

// constitutionJSONOutput is the output struct for JSON format.
type constitutionJSONOutput struct {
	Entries []constitutionJSONEntry `json:"entries"`
}

// constitutionJSONEntry is the entry struct for JSON serialization.
type constitutionJSONEntry struct {
	ID         string `json:"id"`
	Zone       string `json:"zone"`
	File       string `json:"file"`
	Anchor     string `json:"anchor"`
	Clause     string `json:"clause"`
	CanaryGate bool   `json:"canary_gate"`
}

// renderConstitutionJSON prints entries in JSON format.
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

// renderConstitutionTable prints entries in table format.
// Clause is truncated to 40 characters unless -v is specified.
func renderConstitutionTable(w io.Writer, entries []constitution.Rule) {
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(w, "no entries")
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

	_, _ = fmt.Fprintf(w, "\ntotal %d entries\n", len(entries))
}
