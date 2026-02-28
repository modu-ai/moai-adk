package cli

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/spf13/cobra"
)

var profileSetupCmd = &cobra.Command{
	Use:   "setup [name]",
	Short: "Interactive setup wizard for profile launch options",
	Long: `Configure per-profile claude launch preferences through an interactive wizard.

Settings are stored in ~/.moai/claude-profiles/<name>/.launch.yaml
and apply whenever you launch Claude with that profile.

Examples:
  moai profile setup          # Configure default profile
  moai profile setup work     # Configure 'work' profile
  moai profile --setup work   # Equivalent using flag form`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProfileSetup,
}

func init() {
	profileCmd.AddCommand(profileSetupCmd)
}

// runProfileSetup runs the interactive profile launch configuration wizard.
func runProfileSetup(cmd *cobra.Command, args []string) error {
	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}

	// Load existing config as defaults
	existing, err := profile.ReadLaunchConfig(profileName)
	if err != nil {
		return fmt.Errorf("read existing config: %w", err)
	}

	// Form values (pre-filled from existing config)
	model := existing.Model
	bypass := existing.Bypass
	cont := existing.Continue

	// Chrome: convert *bool → string for select field
	chromeVal := "default"
	if existing.Chrome != nil {
		if *existing.Chrome {
			chromeVal = "enabled"
		} else {
			chromeVal = "disabled"
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"Configuring launch options for profile '%s'\n\n", profileName)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Default model").
				Description("Override the model when launching with this profile.\nLeave as 'default' to use the system default.").
				Options(
					huh.NewOption("Default (no override)", ""),
					huh.NewOption("claude-opus-4-6 (most capable)", "claude-opus-4-6"),
					huh.NewOption("claude-sonnet-4-6 (balanced)", "claude-sonnet-4-6"),
					huh.NewOption("claude-haiku-4-5 (fastest)", "claude-haiku-4-5-20251001"),
				).
				Value(&model),

			huh.NewConfirm().
				Title("Bypass permission checks?").
				Description("Adds --dangerously-skip-permissions. Use for trusted environments only.").
				Value(&bypass),

			huh.NewConfirm().
				Title("Continue previous session by default?").
				Description("Adds --continue when launching. Falls back to new session if none exists.").
				Value(&cont),

			huh.NewSelect[string]().
				Title("Chrome MCP").
				Description("Controls Chrome browser integration for this profile.").
				Options(
					huh.NewOption("System default", "default"),
					huh.NewOption("Enabled (allow Chrome MCP)", "enabled"),
					huh.NewOption("Disabled (--no-chrome)", "disabled"),
				).
				Value(&chromeVal),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Setup cancelled.")
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	cfg := profile.LaunchConfig{
		Model:    model,
		Bypass:   bypass,
		Continue: cont,
	}
	switch chromeVal {
	case "enabled":
		t := true
		cfg.Chrome = &t
	case "disabled":
		f := false
		cfg.Chrome = &f
	// "default": leave Chrome nil (system default)
	}

	if err := profile.WriteLaunchConfig(profileName, cfg); err != nil {
		return fmt.Errorf("save launch config: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"Saved launch settings for profile '%s' → %s\n",
		profileName, profile.GetLaunchConfigPath(profileName))
	return nil
}
