package cli

// @MX:NOTE: [AUTO] Telemetry reporting for skill usage effectiveness analysis
// @MX:NOTE: [AUTO] Data stored locally in .moai/evolution/telemetry/, never sent externally
// @MX:NOTE: [AUTO] All context hashed to prevent PII storage

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// telemetryCmd is the root command for telemetry subcommands.
var telemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Skill usage telemetry and reporting",
	Long: `Commands for inspecting skill usage telemetry data.

Telemetry is stored locally in .moai/evolution/telemetry/ and is never
sent to external services. All context is hashed to prevent PII storage.`,
	GroupID: "tools",
}

// telemetryReportCmd generates a skill effectiveness summary.
var telemetryReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate skill usage effectiveness report",
	Long: `Generate a skill usage effectiveness report from local telemetry data.

The report covers the specified time window and includes:
  - Per-skill usage counts grouped by outcome
  - Top co-occurring skills (loaded in the same session)
  - Underutilized skills (fewer than 3 uses in the window)

Telemetry data is stored in .moai/evolution/telemetry/ and is entirely local.`,
	RunE: runTelemetryReport,
}

func init() {
	telemetryReportCmd.Flags().Int("days", 30, "Time window in days for the report")
	telemetryCmd.AddCommand(telemetryReportCmd)
}

// runTelemetryReport executes the telemetry report subcommand.
func runTelemetryReport(cmd *cobra.Command, _ []string) error {
	days, err := cmd.Flags().GetInt("days")
	if err != nil {
		return fmt.Errorf("telemetry report: get days flag: %w", err)
	}
	if days <= 0 {
		return fmt.Errorf("telemetry report: --days must be a positive integer")
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("telemetry report: get working directory: %w", err)
	}

	report, err := telemetry.GenerateReport(projectRoot, days)
	if err != nil {
		return fmt.Errorf("telemetry report: %w", err)
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), report.String())
	return nil
}
