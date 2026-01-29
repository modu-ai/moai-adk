package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/anthropics/moai-adk-go/internal/status"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		Long: `Display the current status of your MoAI project, including
active SPEC documents, configuration state, and development progress.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}

	return cmd
}

// runStatus executes the status command logic
func runStatus() error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Display header
	fmt.Println(output.HeaderStyle.Render("Project Status"))
	fmt.Println()

	// Show git information
	showGitStatus()

	// Show active SPECs
	showSpecStatus(cwd)

	// Show quality gates
	showQualityStatus(cwd)

	// Show recent changes
	showRecentChanges()

	return nil
}

// showGitStatus displays git repository information
func showGitStatus() {
	gitInfo, err := status.GetGitInfo()
	if err != nil {
		fmt.Println(output.MutedStyle.Render("Git: Not a git repository"))
		fmt.Println()
		return
	}

	branchStatus := output.SuccessStyle.Render("✓")
	if gitInfo.HasChanges {
		branchStatus = output.WarningStyle.Render("✗")
	}

	fmt.Println(output.InfoStyle.Render("Git:"))
	fmt.Printf("  %s Branch: %s", branchStatus, gitInfo.Branch)
	if gitInfo.HasChanges {
		fmt.Printf(" (modified)")
	} else {
		fmt.Printf(" (clean)")
	}
	fmt.Println()
	fmt.Println()
}

// showSpecStatus displays active SPEC documents
func showSpecStatus(projectDir string) {
	specs, err := status.GetActiveSpecs(projectDir)
	if err != nil {
		fmt.Println(output.ErrorStyle.Render("Error reading SPECs"))
		fmt.Println()
		return
	}

	if len(specs) == 0 {
		fmt.Println(output.MutedStyle.Render("No active SPECs"))
		fmt.Println()
		return
	}

	fmt.Println(output.InfoStyle.Render("Active SPECs:"))
	for _, spec := range specs {
		icon := status.FormatStatusIcon(spec.Status)
		fmt.Printf("  %s %s (%s) [%s]\n", icon, spec.ID, spec.Title, spec.Status)
	}
	fmt.Println()
}

// showQualityStatus displays quality gate status
func showQualityStatus(projectDir string) {
	quality, err := status.GetQualityGateStatus(projectDir)
	if err != nil {
		fmt.Println(output.ErrorStyle.Render("Error checking quality gates"))
		fmt.Println()
		return
	}

	fmt.Println(output.InfoStyle.Render("Quality Gates:"))

	if quality.Tested.Passed {
		fmt.Println("  " + output.CheckmarkStyle.Render("✓") + " Tested: " + quality.Tested.Message)
	} else {
		fmt.Println("  " + output.CrossmarkStyle.Render("✗") + " Tested: " + quality.Tested.Message)
	}

	if quality.Readable.Passed {
		fmt.Println("  " + output.CheckmarkStyle.Render("✓") + " Readable: " + quality.Readable.Message)
	} else {
		fmt.Println("  " + output.CrossmarkStyle.Render("✗") + " Readable: " + quality.Readable.Message)
	}

	if quality.Unified.Passed {
		fmt.Println("  " + output.CheckmarkStyle.Render("✓") + " Unified: " + quality.Unified.Message)
	} else {
		fmt.Println("  " + output.CrossmarkStyle.Render("✗") + " Unified: " + quality.Unified.Message)
	}

	if quality.Secured.Passed {
		fmt.Println("  " + output.CheckmarkStyle.Render("✓") + " Secured: " + quality.Secured.Message)
	} else {
		fmt.Println("  " + output.CrossmarkStyle.Render("✗") + " Secured: " + quality.Secured.Message)
	}

	if quality.Trackable.Passed {
		fmt.Println("  " + output.CheckmarkStyle.Render("✓") + " Trackable: " + quality.Trackable.Message)
	} else {
		fmt.Println("  " + output.CrossmarkStyle.Render("✗") + " Trackable: " + quality.Trackable.Message)
	}

	fmt.Println()
}

// showRecentChanges displays recent git commits
func showRecentChanges() {
	commits, err := status.GetRecentCommits(3)
	if err != nil {
		// Don't show error if git fails
		return
	}

	if len(commits) == 0 {
		return
	}

	fmt.Println(output.InfoStyle.Render("Recent Changes:"))
	for _, commit := range commits {
		fmt.Printf("  • %s (%s)\n", commit.Message, commit.TimeAgo)
	}
	fmt.Println()
}
