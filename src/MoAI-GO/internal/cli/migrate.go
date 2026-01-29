// Package cli provides the migrate command for converting Python hooks to Go hooks
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/migration"
	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/spf13/cobra"
)

// SettingsJSON represents the Claude Code settings.json structure
type SettingsJSON struct {
	Hooks           map[string]HookEntry `json:"hooks,omitempty"`
	HookCommands    map[string]string    `json:"hookCommands,omitempty"`
	AutoApproval    *AutoApprovalConfig  `json:"autoApproval,omitempty"`
	OtherProperties map[string]any       `json:"-"`
}

// HookEntry represents a single hook configuration
type HookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// AutoApprovalConfig represents auto-approval settings
type AutoApprovalConfig struct {
	AllowedTools    []string `json:"allowedTools,omitempty"`
	DisallowedTools []string `json:"disallowedTools,omitempty"`
}

// NewMigrateCommand creates the migrate command
func NewMigrateCommand() *cobra.Command {
	var dryRun bool
	var forceGo bool
	var forcePython bool
	var goBinaryPath string
	var rollback bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate from Python to Go implementation",
		Long: `Migrate your MoAI-ADK installation from Python to Go implementation.

This command converts your .claude/settings.json to use Go hooks instead of
Python hooks. It preserves all non-hook settings and creates a backup before
making changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(dryRun, forceGo, forcePython, goBinaryPath, rollback)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	cmd.Flags().BoolVar(&forceGo, "force-go", false, "Force migration to Go implementation")
	cmd.Flags().BoolVar(&forcePython, "force-python", false, "Force migration to Python implementation")
	cmd.Flags().StringVar(&goBinaryPath, "go-binary-path", "", "Path to Go binary (for --force-go)")
	cmd.Flags().BoolVar(&rollback, "rollback", false, "Rollback to previous configuration")

	return cmd
}

// runMigrate executes the migration process
func runMigrate(dryRun, forceGo, forcePython bool, goBinaryPath string, rollback bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	settingsPath := filepath.Join(cwd, ".claude", "settings.json")

	// Handle rollback
	if rollback {
		return rollbackMigration(settingsPath)
	}

	// Detect implementation
	detector := migration.DetectImplementationWithOverride(forceGo, forcePython, goBinaryPath)

	fmt.Println(output.HeaderStyle.Render("MoAI-ADK Migration"))
	fmt.Println()

	if !detector.Found {
		return fmt.Errorf("no valid implementation found. Please install moai Go binary or ensure Python is available")
	}

	fmt.Printf("Target implementation: %s\n", detector.Type)

	if detector.Type == migration.ImplementationGo {
		fmt.Printf("Go binary path: %s\n", detector.BinaryPath)
		fmt.Printf("Go version: %s\n", detector.Version)
	} else {
		fmt.Printf("Python command: %s\n", detector.PythonCmd)
		if detector.UsingGoFallback {
			fmt.Println("Note: Using Python fallback (Go binary not found)")
		}
	}
	fmt.Println()

	// Read current settings
	settings, err := readSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("error reading settings: %w", err)
	}

	// Create backup
	if !dryRun {
		fmt.Println(output.InfoStyle.Render("Creating backup..."))
		backupPath := settingsPath + ".backup"
		if err := copyFile(settingsPath, backupPath); err != nil {
			return fmt.Errorf("error creating backup: %w", err)
		}
		fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Backed up to %s", backupPath)))
		fmt.Println()
	}

	// Convert hooks
	changes := convertHooks(settings, detector)

	// Print preview
	fmt.Println(output.InfoStyle.Render("Hook command changes:"))
	for _, change := range changes {
		if change.Old != change.New {
			fmt.Printf("  %s:\n", change.Hook)
			fmt.Printf("    Old: %s\n", output.MutedStyle.Render(change.Old))
			fmt.Printf("    New: %s\n", output.SuccessStyle.Render(change.New))
			fmt.Println()
		}
	}

	if len(changes) == 0 {
		fmt.Println(output.MutedStyle.Render("No changes needed. Hooks already in target format."))
		return nil
	}

	if dryRun {
		fmt.Println()
		fmt.Println(output.MutedStyle.Render("Dry run complete. No changes were applied."))
		fmt.Println(output.MutedStyle.Render("Run without --dry-run to apply changes."))
		return nil
	}

	// Write new settings
	if err := writeSettings(settingsPath, settings); err != nil {
		// Rollback on error
		_ = rollbackMigration(settingsPath)
		return fmt.Errorf("error writing settings: %w", err)
	}

	fmt.Println()
	fmt.Println(output.SuccessStyle.Render("✓ Migration complete!"))
	fmt.Println()
	fmt.Println("Test your hooks to ensure they work correctly.")
	fmt.Printf("To rollback, run: moai migrate --rollback\n")

	return nil
}

// HookChange represents a hook command change
type HookChange struct {
	Hook string
	Old  string
	New  string
}

// readSettings reads the settings.json file
func readSettings(path string) (*SettingsJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var settings SettingsJSON
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	// Extract other properties for preservation
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, err
	}
	settings.OtherProperties = rawMap

	return &settings, nil
}

// writeSettings writes the settings.json file
func writeSettings(path string, settings *SettingsJSON) error {
	// Merge preserved properties with new settings
	outputMap := make(map[string]any)

	// Copy all original properties
	for k, v := range settings.OtherProperties {
		outputMap[k] = v
	}

	// Update hook commands
	if settings.HookCommands != nil {
		outputMap["hookCommands"] = settings.HookCommands
	}

	data, err := json.MarshalIndent(outputMap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// convertHooks converts hook commands to target implementation
func convertHooks(settings *SettingsJSON, detector *migration.DetectionResult) []HookChange {
	// All MoAI hooks
	hooks := []string{
		"PreToolUse",
		"PostToolUse",
		"SessionStart",
		"SessionEnd",
		"PreCompact",
		"Notification",
	}

	var changes []HookChange

	if settings.HookCommands == nil {
		settings.HookCommands = make(map[string]string)
	}

	for _, hook := range hooks {
		// Get current command
		oldCmd := settings.HookCommands[hook]

		// Generate new command
		newCmd := generateHookCommand(hook, detector)

		// Track change
		if oldCmd != newCmd {
			changes = append(changes, HookChange{
				Hook: hook,
				Old:  oldCmd,
				New:  newCmd,
			})
		}

		// Update settings
		settings.HookCommands[hook] = newCmd
	}

	return changes
}

// generateHookCommand generates the appropriate hook command
func generateHookCommand(hook string, detector *migration.DetectionResult) string {
	if detector.Type == migration.ImplementationGo {
		// Convert hook name to kebab-case
		hookName := toKebabCase(hook)
		return fmt.Sprintf("%s hook %s --project-dir \"$CLAUDE_PROJECT_DIR\"", detector.BinaryPath, hookName)
	}

	// Python hook command
	return fmt.Sprintf(`${SHELL:-/bin/bash} -l -c 'uv run "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/%s.py"'`, toKebabCase(hook))
}

// toKebabCase converts PascalCase to kebab-case
func toKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// rollbackMigration rolls back to the previous configuration
func rollbackMigration(settingsPath string) error {
	backupPath := settingsPath + ".backup"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	// Rename current settings
	currentBackup := settingsPath + ".go"
	if err := os.Rename(settingsPath, currentBackup); err != nil {
		return fmt.Errorf("error backing up current settings: %w", err)
	}

	// Restore backup
	if err := os.Rename(backupPath, settingsPath); err != nil {
		// Rollback failed
		_ = os.Rename(currentBackup, settingsPath)
		return fmt.Errorf("error restoring backup: %w", err)
	}

	fmt.Println(output.SuccessStyle.Render("✓ Rollback complete!"))
	fmt.Printf("Previous configuration saved to: %s\n", currentBackup)

	return nil
}
