package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/shell"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
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
	_, _ = fmt.Fprintf(out, "Current version: moai-adk %s\n", currentVersion)

	// Check if using local update source
	updateSource := os.Getenv("MOAI_UPDATE_SOURCE")
	useLocalUpdate := updateSource == "local"

	// Detect development build and show appropriate message
	// Skip this check for local updates (dev builds can update from local releases)
	isDevBuild := strings.Contains(currentVersion, "dirty") ||
		strings.Contains(currentVersion, "dev") ||
		strings.Contains(currentVersion, "none")

	if isDevBuild && !templatesOnly && !shellEnv && !useLocalUpdate {
		_, _ = fmt.Fprintln(out, "\nDevelopment build detected.")
		_, _ = fmt.Fprintln(out, "Binary update skipped. To update binary:")
		_, _ = fmt.Fprintln(out, "  cd ~/MoAI/moai-adk-go && git pull && make install")
		if checkOnly {
			return nil
		}
		// For dev builds, skip binary update but still sync templates
		_, _ = fmt.Fprintln(out, "\nSyncing templates...")
		return runTemplateSync(cmd)
	}

	// Show update source info
	if useLocalUpdate {
		releasesDir := os.Getenv("MOAI_RELEASES_DIR")
		if releasesDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				homeDir = "."
			}
			releasesDir = filepath.Join(homeDir, ".moai", "releases")
		}
		_, _ = fmt.Fprintf(out, "Update source: local (%s)\n", releasesDir)
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
			_, _ = fmt.Fprintln(out, "Update checker not available. Using current version.")
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
		_, _ = fmt.Fprintf(out, "Latest version:  %s\n", info.Version)
		return nil
	}

	if deps.UpdateOrch == nil {
		return fmt.Errorf("update orchestrator not initialized")
	}

	// Check if Go binary is available before attempting update
	info, err := deps.UpdateChecker.CheckLatest(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(out, "Warning: could not check for updates: %v\n", err)
		_, _ = fmt.Fprintln(out, "Falling back to template sync only...")
		return runTemplateSync(cmd)
	}

	// If no Go binary URL available, skip binary update and sync templates only
	if info.URL == "" {
		_, _ = fmt.Fprintf(out, "Latest version:  %s\n", info.Version)
		_, _ = fmt.Fprintln(out, "\nNo Go binary available for this platform. Syncing templates only...")
		return runTemplateSync(cmd)
	}

	result, err := deps.UpdateOrch.Update(ctx)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Updated from %s to %s\n", result.PreviousVersion, result.NewVersion)
	_, _ = fmt.Fprintf(out, "  Files updated: %d, merged: %d\n", result.FilesUpdated, result.FilesMerged)

	// Sync templates after successful binary update
	_, _ = fmt.Fprintln(out, "Syncing templates...")

	// Save and restore current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	defer func() {
		if cwd != "" {
			_ = os.Chdir(cwd)
		}
	}()

	// Use project root (current directory) as deployment target
	projectRoot := "."

	// Load embedded templates
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		_, _ = fmt.Fprintf(out, "Warning: load templates failed: %v\n", err)
		_, _ = fmt.Fprintln(out, "You can run 'moai update --templates-only' to retry.")
		return nil
	}

	// Initialize manifest manager
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		_, _ = fmt.Fprintf(out, "Warning: load manifest failed: %v\n", err)
		_, _ = fmt.Fprintln(out, "You can run 'moai update --templates-only' to retry.")
		return nil
	}

	// Create deployer
	deployer := template.NewDeployer(embedded)

	// Deploy templates with independent context
	syncCtx := context.Background()
	if err := deployer.Deploy(syncCtx, projectRoot, mgr, nil); err != nil {
		_, _ = fmt.Fprintf(out, "Warning: template sync failed: %v\n", err)
		_, _ = fmt.Fprintln(out, "You can run 'moai update --templates-only' to retry.")
		return nil
	}

	_, _ = fmt.Fprintln(out, "Templates synced successfully.")
	return nil
}

// runTemplateSync synchronizes templates from the embedded filesystem.
func runTemplateSync(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	// Use current directory as project root
	projectRoot := "."

	_, _ = fmt.Fprintln(out, "Syncing templates from embedded filesystem...")

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

	_, _ = fmt.Fprintln(out, "Template sync complete.")
	return nil
}

// runShellEnvConfig configures shell environment variables for Claude Code.
func runShellEnvConfig(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out, "Configuring shell environment for Claude Code...")

	// Get recommendation first
	configurator := shell.NewEnvConfigurator(nil)
	rec := configurator.GetRecommendation()

	_, _ = fmt.Fprintf(out, "  Shell: %s\n", rec.Shell)
	_, _ = fmt.Fprintf(out, "  Config file: %s\n", rec.ConfigFile)
	_, _ = fmt.Fprintf(out, "  Explanation: %s\n", rec.Explanation)
	_, _ = fmt.Fprintln(out, "  Changes to add:")
	for _, change := range rec.Changes {
		_, _ = fmt.Fprintf(out, "    - %s\n", change)
	}
	_, _ = fmt.Fprintln(out)

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
		_, _ = fmt.Fprintf(out, "Shell environment already configured in %s\n", result.ConfigFile)
	} else {
		_, _ = fmt.Fprintf(out, "Shell environment configured in %s\n", result.ConfigFile)
		_, _ = fmt.Fprintln(out, "Please restart your terminal or run:")
		_, _ = fmt.Fprintf(out, "  source %s\n", result.ConfigFile)
	}

	return nil
}
