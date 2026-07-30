package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/spf13/cobra"
)

// migrationCmd is the 'moai migration' cobra command group.
// REQ-V3R2-RT-007-015, REQ-V3R2-RT-007-040, REQ-V3R2-RT-007-041, REQ-V3R2-RT-007-024.
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Manage migrations (run, status, rollback)",
	Long: `Migration management tool.

'run': execute pending migrations.
'status': display the current migration status.
'rollback': roll back to a specific version (some migrations are not rollback-capable).

Note: do not confuse this with 'moai migrate agency'. 'migrate agency' is a
one-off agency migration command; this command is the version-controlled
migration framework.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// migrationRunCmd executes pending migrations.
// REQ-V3R2-RT-007-040: 'moai migration run' subcommand.
var migrationRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute pending migrations",
	Long: `Run, in order, every migration that has not yet been applied in the current project.

Migrations also run automatically via the session-start hook; this command lets you trigger them manually.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		runner := migration.NewRunner(cwd)
		ctx := context.Background()
		applied, err := runner.Apply(ctx)
		if err != nil {
			return fmt.Errorf("migration run failed: %w", err)
		}

		p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
		if len(applied) == 0 {
			p.Info("No pending migrations to apply.")
			return nil
		}

		p.Success("Applied %d migration(s) (versions: %v)", len(applied), applied)
		return nil
	},
}

// migrationStatusCmd displays the current migration status.
// REQ-V3R2-RT-007-015: 'moai migration status [--json]' subcommand.
var migrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the migration status",
	Long: `Display the current migration status for the project.

Output fields:
- Current version (most recently applied migration)
- List of pending migrations
- Details of the most recently applied migration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonFlag, _ := cmd.Flags().GetBool("json")
		p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		runner := migration.NewRunner(cwd)
		current, pending, lastApplied, err := runner.Status()
		if err != nil {
			return fmt.Errorf("failed to retrieve migration status: %w", err)
		}

		if jsonFlag {
			// JSON output (REQ-V3R2-RT-007-041)
			output := map[string]any{
				"current_version": current,
				"pending":         pending,
				"last_applied":    lastApplied,
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			_ = p.Data(string(data))
			return nil
		}

		// Human-readable output — composed into a single multi-line string so
		// p.Data writes the block to stdout byte-identical to the prior
		// sequential stdout writes (DECISION 2026-07-14).
		lines := []string{fmt.Sprintf("Current version: %d", current)}
		if len(pending) > 0 {
			lines = append(lines, fmt.Sprintf("Pending migrations (%d): %v", len(pending), pending))
		} else {
			lines = append(lines, "No pending migrations (up to date)")
		}
		if lastApplied != nil {
			lines = append(lines, fmt.Sprintf("Last applied: %s (version %d)", lastApplied.Name, lastApplied.Version))
		}
		_ = p.Data(strings.Join(lines, "\n"))

		return nil
	},
}

// migrationRollbackCmd rolls back to a specific version.
// REQ-V3R2-RT-007-024: 'moai migration rollback <version>' subcommand.
var migrationRollbackCmd = &cobra.Command{
	Use:   "rollback <version>",
	Short: "Roll back to a specific version",
	Long: `Roll back to the specified version.

Caution: some migrations (especially critical bug fixes) may be marked as
non-rollback-capable. Rollback reverts the version-file to the previous
version and executes the Rollback function of the affected migration.

Example:
  moai migration rollback 0  # Roll back every migration (initial state)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetVersion int
		_, err := fmt.Sscanf(args[0], "%d", &targetVersion)
		if err != nil {
			return fmt.Errorf("invalid version number: %s: %w", args[0], err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		runner := migration.NewRunner(cwd)
		if err := runner.Rollback(targetVersion); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
		p.Success("Rolled back to version %d", targetVersion)
		return nil
	},
}

func init() {
	// Register migration subcommands
	migrationCmd.AddCommand(migrationRunCmd)
	migrationCmd.AddCommand(migrationStatusCmd)
	migrationCmd.AddCommand(migrationRollbackCmd)

	// Add the --json flag to the status command
	migrationStatusCmd.Flags().Bool("json", false, "Output in JSON format")

	// Registration of the migration group on rootCmd is performed by root.go
}
