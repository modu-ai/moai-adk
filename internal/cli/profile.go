package cli

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage Claude configuration profiles",
	Long: `Manage Claude configuration profiles stored in ~/.moai/claude-profiles/.

Each profile is an isolated Claude configuration directory (CLAUDE_CONFIG_DIR).
Use -p/--profile with cc, cg, or glm to switch between profiles.

Run 'moai profile setup [name]' or 'moai profile --setup [name]' to configure
per-profile launch options (model, bypass, continue, Chrome MCP).`,
	RunE: runProfileCmd,
}

var profileListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available profiles",
	RunE:    runProfileList,
}

var profileCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current profile name",
	RunE:  runProfileCurrent,
}

var profileDeleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"rm"},
	Short:   "Delete a profile",
	Args:    cobra.ExactArgs(1),
	RunE:    runProfileDelete,
}

func init() {
	profileCmd.Flags().BoolP("setup", "s", false, "Run interactive setup wizard for launch options")
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCurrentCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	rootCmd.AddCommand(profileCmd)
}

// runProfileCmd handles 'moai profile' with optional --setup/-s flag.
func runProfileCmd(cmd *cobra.Command, args []string) error {
	setup, _ := cmd.Flags().GetBool("setup")
	if setup {
		return runProfileSetup(cmd, args)
	}
	return cmd.Help()
}

func runProfileList(cmd *cobra.Command, _ []string) error {
	entries := profile.List()
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no profiles found")
		return nil
	}
	for _, e := range entries {
		marker := "  "
		if e.Current {
			marker = "* "
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", marker, e.Name)
	}
	return nil
}

func runProfileCurrent(cmd *cobra.Command, _ []string) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), profile.GetCurrentName())
	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	if err := profile.Delete(args[0]); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}
