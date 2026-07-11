package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/spf13/cobra"
)

func newSpecDriftCmd() *cobra.Command {
	var jsonOutput bool
	var exitCodeOnDrift bool
	var countOnly bool
	var noCache bool

	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Detect SPEC status drift between frontmatter and git log",
		Long: `Detect SPEC status drift by comparing frontmatter status field against
git log on main branch.

Results are cached against the current HEAD commit, so repeated runs at an
unchanged HEAD are served from cache. Because the cache key is the HEAD SHA,
an uncommitted frontmatter edit does not invalidate it — pass --no-cache to
force a fresh computation.

Examples:
  moai spec drift                    # Tabular report
  moai spec drift --json             # JSON output
  moai spec drift --exit-code-on-drift  # Exit 1 if drift detected
  moai spec drift --count            # Just print drift count
  moai spec drift --no-cache         # Bypass the HEAD-SHA cache (always fresh)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("failed to find project root: %w", err)
			}

			if countOnly {
				count, err := driftCountFn(projectRoot, noCache)
				if err != nil {
					return fmt.Errorf("failed to count drift: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), count)
				return nil
			}

			report, err := detectDriftFn(projectRoot, noCache)
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

				count, err := driftCountFn(projectRoot, noCache)
				if err != nil {
					return nil
				}

				if count > 0 {
					return &exitCodeError{code: 1, msg: "spec drift: drift detected (--exit-code-on-drift)"}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&exitCodeOnDrift, "exit-code-on-drift", false, "Exit with code 1 if drift detected")
	cmd.Flags().BoolVar(&countOnly, "count", false, "Only print the drift count")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass the HEAD-SHA result cache and recompute freshly")

	return cmd
}

// detectDriftFn routes to the cached or the fresh drift entry point.
// The --no-cache path is authoritative: it never reads the HEAD-SHA cache, so an
// operator can always obtain a fresh count regardless of cache state.
func detectDriftFn(projectRoot string, noCache bool) (*spec.DriftReport, error) {
	if noCache {
		return spec.DetectDriftFresh(projectRoot)
	}
	return spec.DetectDrift(projectRoot)
}

// driftCountFn is the count-only counterpart of detectDriftFn.
func driftCountFn(projectRoot string, noCache bool) (int, error) {
	if noCache {
		return spec.DriftCountFresh(projectRoot)
	}
	return spec.DriftCount(projectRoot)
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
