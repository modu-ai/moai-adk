package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/internal/manifest"
	"github.com/modu-ai/moai-adk-go/internal/shell"
	"github.com/modu-ai/moai-adk-go/internal/template"
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
	updateCmd.Flags().Bool("templates-only", false, "Sync templates without updating binary")
	updateCmd.Flags().Bool("shell-env", false, "Configure shell environment variables for Claude Code")
}

// runUpdate executes the self-update workflow by delegating to SPEC-UPDATE-001 modules.
func runUpdate(cmd *cobra.Command, _ []string) error {
	checkOnly := getBoolFlag(cmd, "check")
	templatesOnly := getBoolFlag(cmd, "templates-only")
	shellEnv := getBoolFlag(cmd, "shell-env")
	out := cmd.OutOrStdout()

	currentVersion := version.GetVersion()
	fmt.Fprintf(out, "Current version: moai-adk %s\n", currentVersion)

	// Detect development build and show appropriate message
	isDevBuild := strings.Contains(currentVersion, "dirty") ||
		strings.Contains(currentVersion, "dev") ||
		strings.Contains(currentVersion, "none")

	if isDevBuild && !templatesOnly && !shellEnv {
		fmt.Fprintln(out, "\nDevelopment build detected. To update:")
		fmt.Fprintln(out, "  cd ~/MoAI/moai-adk-go && git pull && make install")
		if checkOnly {
			return nil
		}
		// For dev builds, skip the actual update and return success
		return nil
	}

	// Handle shell-env mode
	if shellEnv {
		return runShellEnvConfig(cmd)
	}

	// Handle templates-only mode
	if templatesOnly {
		return runTemplateSync(cmd)
	}

	// Lazily initialize update dependencies
	if deps != nil {
		if err := deps.EnsureUpdate(); err != nil {
			deps.Logger.Debug("failed to initialize update system", "error", err)
		}
	}

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

// runTemplateSync synchronizes templates from the embedded filesystem.
func runTemplateSync(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	// Use current directory as project root
	projectRoot := "."

	fmt.Fprintln(out, "Syncing templates from embedded filesystem...")

	// Load embedded templates
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Initialize manifest manager
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	// Create deployer
	deployer := template.NewDeployer(embedded)

	// Deploy templates
	if err := deployer.Deploy(ctx, projectRoot, mgr, nil); err != nil {
		return fmt.Errorf("deploy templates: %w", err)
	}

	fmt.Fprintln(out, "Template sync complete.")
	return nil
}

// runShellEnvConfig configures shell environment variables for Claude Code.
func runShellEnvConfig(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "Configuring shell environment for Claude Code...")

	// Get recommendation first
	configurator := shell.NewEnvConfigurator(nil)
	rec := configurator.GetRecommendation()

	fmt.Fprintf(out, "  Shell: %s\n", rec.Shell)
	fmt.Fprintf(out, "  Config file: %s\n", rec.ConfigFile)
	fmt.Fprintf(out, "  Explanation: %s\n", rec.Explanation)
	fmt.Fprintln(out, "  Changes to add:")
	for _, change := range rec.Changes {
		fmt.Fprintf(out, "    - %s\n", change)
	}
	fmt.Fprintln(out)

	// Execute configuration
	result, err := configurator.Configure(shell.ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
		PreferLoginShell:        true,
	})
	if err != nil {
		return fmt.Errorf("configure shell environment: %w", err)
	}

	if result.Skipped {
		fmt.Fprintf(out, "Shell environment already configured in %s\n", result.ConfigFile)
	} else {
		fmt.Fprintf(out, "Shell environment configured in %s\n", result.ConfigFile)
		fmt.Fprintln(out, "Please restart your terminal or run:")
		fmt.Fprintf(out, "  source %s\n", result.ConfigFile)
	}

	return nil
}
