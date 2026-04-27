package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecStatusCmd creates the 'moai spec status' subcommand
func newSpecStatusCmd() *cobra.Command {
	var dryRun bool
	var listAll bool

	cmd := &cobra.Command{
		Use:   "status <SPEC-ID> <new-status>",
		Short: "Update or list SPEC document statuses",
		Long: `Update the status field in a SPEC document, or list all SPECs.

Examples:
  moai spec status SPEC-XXX completed        # Update status
  moai spec status SPEC-XXX completed --dry-run  # Preview change
  moai spec status --list                    # List all SPECs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle --list flag
			if listAll {
				return listAllSpecs(cmd)
			}

			// Require SPEC-ID and new-status arguments
			if len(args) < 2 {
				return cmd.Help()
			}

			specID := args[0]
			newStatus := args[1]

			return updateSpecStatus(cmd, specID, newStatus, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview change without writing")
	cmd.Flags().BoolVar(&listAll, "list", false, "List all SPECs and their status")

	return cmd
}

// updateSpecStatus updates the status of a SPEC document
func updateSpecStatus(cmd *cobra.Command, specID, newStatus string, dryRun bool) error {
	// Find project root
	projectRoot, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Locate SPEC directory
	specDir := filepath.Join(projectRoot, ".moai", "specs", specID)

	// Get current status for display
	oldStatus := "unknown"
	if content, err := spec.ParseStatus(specDir); err == nil {
		oldStatus = content
	}

	// Dry run: just show what would change
	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Would update: %s status %s → %s\n", specID, oldStatus, newStatus)
		return nil
	}

	// Perform update
	if err := spec.UpdateStatus(specDir, newStatus); err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s status updated: %s → %s\n", specID, oldStatus, newStatus)
	return nil
}

// listAllSpecs lists all SPECs and their current status
func listAllSpecs(cmd *cobra.Command) error {
	projectRoot, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	specsDir := filepath.Join(projectRoot, ".moai", "specs")

	// Check if directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		fmt.Fprintf(cmd.OutOrStdout(), "No SPECs directory found at %s\n", specsDir)
		return nil
	}

	// Read all subdirectories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Print header
	fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-15s %s\n", "SPEC-ID", "Status", "Modified")
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 80))

	// List each SPEC
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specID := entry.Name()
		specDir := filepath.Join(specsDir, specID)

		// Parse status
		status := "unknown"
		if s, err := spec.ParseStatus(specDir); err == nil {
			status = s
		}

		// Get modification time
		info, err := entry.Info()
		if err != nil {
			continue
		}
		modTime := info.ModTime().Format("2006-01-02 15:04")

		fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-15s %s\n", specID, status, modTime)
	}

	return nil
}
