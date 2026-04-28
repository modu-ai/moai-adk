package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// specIDPattern matches SPEC-XXX patterns in git commit messages
var specIDPattern = regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]+`)

// newSpecStatusCmd creates the 'moai spec status' subcommand
func newSpecStatusCmd() *cobra.Command {
	var dryRun bool
	var listAll bool
	var syncGit bool
	var syncConfirm bool
	var syncYes bool

	cmd := &cobra.Command{
		Use:   "status <SPEC-ID> <new-status>",
		Short: "Update or list SPEC document statuses",
		Long: `Update the status field in a SPEC document, or list all SPECs.

Examples:
  moai spec status SPEC-XXX completed        # Update status
  moai spec status SPEC-XXX completed --dry-run  # Preview change
  moai spec status --list                    # List all SPECs
  moai spec status --sync-git                # Sync from git log
  moai spec status --sync-git --yes          # Non-interactive sync`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if syncGit {
				return syncGitSpecStatuses(cmd, syncYes)
			}

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
	cmd.Flags().BoolVar(&syncGit, "sync-git", false, "Sync SPEC statuses from git log on main")
	cmd.Flags().BoolVar(&syncConfirm, "confirm", false, "Confirm sync-git changes interactively")
	cmd.Flags().BoolVar(&syncYes, "yes", false, "Non-interactive mode for sync-git (auto-confirm)")

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

// syncGitSpecStatuses scans git log on main for SPEC-XXX patterns and updates statuses.
// REQ-5 of SPEC-STATUS-AUTO-001.
func syncGitSpecStatuses(cmd *cobra.Command, autoConfirm bool) error {
	projectRoot, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	specIDsFromGit, err := getSPECIDsFromGitLog()
	if err != nil {
		return fmt.Errorf("failed to scan git log: %w", err)
	}

	if len(specIDsFromGit) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No SPEC-IDs found in git log.")
		return nil
	}

	specsDir := filepath.Join(projectRoot, ".moai", "specs")
	updated := 0
	skipped := 0
	notFound := 0

	for _, specID := range specIDsFromGit {
		specDir := filepath.Join(specsDir, specID)

		if _, err := os.Stat(filepath.Join(specDir, "spec.md")); os.IsNotExist(err) {
			fmt.Fprintf(cmd.OutOrStdout(), "  skipped %s: not found in .moai/specs/\n", specID)
			notFound++
			continue
		}

		currentStatus := "unknown"
		if s, err := spec.ParseStatus(specDir); err == nil {
			currentStatus = s
		}

		if currentStatus == "completed" || currentStatus == "implemented" {
			fmt.Fprintf(cmd.OutOrStdout(), "  skipped %s: already %s\n", specID, currentStatus)
			skipped++
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s → implemented\n", specID, currentStatus)

		if !autoConfirm {
			fmt.Fprintf(cmd.OutOrStdout(), "    Apply? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Fprintf(cmd.OutOrStdout(), "    skipped %s\n", specID)
				skipped++
				continue
			}
		}

		if err := spec.UpdateStatus(specDir, "implemented"); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  error updating %s: %v\n", specID, err)
			continue
		}
		updated++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nSummary: updated %d, skipped %d (already done), %d not found\n", updated, skipped, notFound)
	return nil
}

// getSPECIDsFromGitLog scans git log on main for SPEC-XXX patterns.
func getSPECIDsFromGitLog() ([]string, error) {
	branch := "main"
	if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
		branch = "master"
	}

	out, err := exec.Command("git", "log", branch, "--oneline", "--no-merges").Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	matches := specIDPattern.FindAllString(string(out), -1)

	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}

	return result, nil
}
