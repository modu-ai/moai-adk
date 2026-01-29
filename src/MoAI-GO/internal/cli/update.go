package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/anthropics/moai-adk-go/internal/update"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update MoAI templates",
		Long: `Update MoAI templates and configuration to the latest version.
This command syncs your local project with the latest template changes from
the MoAI-ADK distribution.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")

	return cmd
}

// runUpdate executes the update command logic
func runUpdate(dryRun bool) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Display header
	fmt.Println(output.HeaderStyle.Render("Updating MoAI-ADK templates"))
	fmt.Println()

	// Create backup manager
	backupMgr := update.NewBackupManager(cwd)

	// Create backup
	fmt.Println(output.InfoStyle.Render("Creating backup..."))
	backupDir, err := backupMgr.CreateBackup()
	if err != nil {
		return fmt.Errorf("error creating backup: %w", err)
	}
	fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Backed up to %s", backupDir)))
	fmt.Println()

	// Create sync manager
	syncMgr := update.NewSyncManager(cwd)
	syncMgr.SetDryRun(dryRun)

	// Perform sync
	fmt.Println(output.InfoStyle.Render("Synchronizing templates..."))
	result, err := syncMgr.Sync()
	if err != nil {
		// Rollback on error
		fmt.Println(output.ErrorStyle.Render("Error during sync, rolling back..."))
		_ = backupMgr.Restore(backupDir)
		return fmt.Errorf("sync failed: %w", err)
	}

	// Merge user configuration
	mergeMgr := update.NewMergeManager(cwd, backupDir)
	if err := mergeMgr.MergeUserConfig(); err != nil {
		fmt.Println(output.WarningStyle.Render("Warning: " + err.Error()))
	}

	// Print summary
	fmt.Println()
	fmt.Println(output.SuccessStyle.Render("✓ Update complete!"))
	fmt.Println()
	result.PrintSummary()
	fmt.Println()

	if dryRun {
		fmt.Println(output.MutedStyle.Render("Dry run complete. No changes were applied."))
		fmt.Println(output.MutedStyle.Render("Run without --dry-run to apply changes."))
		fmt.Println()
	}

	return nil
}
