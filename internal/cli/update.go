package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/pkg/version"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update MoAI-ADK to the latest version",
	Long:  "Check for and install the latest MoAI-ADK release. Supports check-only mode, forced updates, and template-only sync.",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("check", false, "Check for updates without installing")
	updateCmd.Flags().Bool("force", false, "Force update even if already up to date")
	updateCmd.Flags().Bool("templates-only", false, "Sync templates only without binary update")
	updateCmd.Flags().Bool("yes", false, "Auto-accept update without confirmation")
}

// runUpdate executes the self-update workflow by delegating to SPEC-UPDATE-001 modules.
func runUpdate(cmd *cobra.Command, _ []string) error {
	checkOnly := getBoolFlag(cmd, "check")
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Current version: moai-adk %s\n", version.GetVersion())

	if deps == nil || deps.UpdateChecker == nil {
		if checkOnly {
			fmt.Fprintln(out, "Update checker not available. Using current version.")
			return nil
		}
		return fmt.Errorf("update system not initialized (update module not available)")
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	if checkOnly {
		info, err := deps.UpdateChecker.CheckLatest(ctx)
		if err != nil {
			return fmt.Errorf("check latest version: %w", err)
		}
		fmt.Fprintf(out, "Latest version:  %s\n", info.Version)
		return nil
	}

	if deps.UpdateOrch == nil {
		return fmt.Errorf("update orchestrator not initialized")
	}

	result, err := deps.UpdateOrch.Update(ctx)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Fprintf(out, "Updated from %s to %s\n", result.PreviousVersion, result.NewVersion)
	fmt.Fprintf(out, "  Files updated: %d, merged: %d\n", result.FilesUpdated, result.FilesMerged)
	return nil
}
