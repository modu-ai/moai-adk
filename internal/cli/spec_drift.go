package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/spf13/cobra"
)

func newSpecDriftCmd() *cobra.Command {
	var jsonOutput bool
	var exitCodeOnDrift bool
	var countOnly bool

	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Detect SPEC status drift between frontmatter and git log",
		Long: `Detect SPEC status drift by comparing frontmatter status field against
git log on main branch.

Examples:
  moai spec drift                    # Tabular report
  moai spec drift --json             # JSON output
  moai spec drift --exit-code-on-drift  # Exit 1 if drift detected
  moai spec drift --count            # Just print drift count`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("failed to find project root: %w", err)
			}

			if countOnly {
				count, err := spec.DriftCount(projectRoot)
				if err != nil {
					return fmt.Errorf("failed to count drift: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), count)
				return nil
			}

			report, err := spec.DetectDrift(projectRoot)
			if err != nil {
				return fmt.Errorf("failed to detect drift: %w", err)
			}

			if jsonOutput {
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			return printDriftReport(cmd.OutOrStdout(), report)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			if exitCodeOnDrift {
				projectRoot, err := findProjectRootFn()
				if err != nil {
					return nil
				}

				count, err := spec.DriftCount(projectRoot)
				if err != nil {
					return nil
				}

				if count > 0 {
					os.Exit(1)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&exitCodeOnDrift, "exit-code-on-drift", false, "Exit with code 1 if drift detected")
	cmd.Flags().BoolVar(&countOnly, "count", false, "Only print the drift count")

	return cmd
}

func printDriftReport(out io.Writer, report *spec.DriftReport) error {
	_, _ = fmt.Fprintf(out, "%-30s %-20s %-20s %-10s\n", "SPEC-ID", "Frontmatter", "Git-Implied", "Drift?")
	_, _ = fmt.Fprintln(out, strings.Repeat("-", 85))

	for _, record := range report.Records {
		driftMark := "aligned"
		if record.Drifted {
			driftMark = "DRIFT"
		}

		_, _ = fmt.Fprintf(out, "%-30s %-20s %-20s %-10s\n",
			record.SPECID,
			record.FrontmatterStatus,
			record.GitImpliedStatus,
			driftMark,
		)
	}

	_, _ = fmt.Fprintln(out, strings.Repeat("-", 85))
	_, _ = fmt.Fprintf(out, "Summary: %d/%d SPECs have status drift\n", report.Count, len(report.Records))

	return nil
}
