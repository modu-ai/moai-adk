package cli

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage Claude configuration profiles",
	Long: `Profile lists and deletes Claude configuration profiles
stored in ~/.claude-profiles/.

Profiles are shared with godo — both tools use the same storage.`,
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
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCurrentCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	rootCmd.AddCommand(profileCmd)
}

func runProfileList(cmd *cobra.Command, _ []string) error {
	entries := profile.List()
	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no profiles found")
		return nil
	}
	for _, e := range entries {
		marker := "  "
		if e.Current {
			marker = "* "
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", marker, e.Name)
	}
	return nil
}

func runProfileCurrent(cmd *cobra.Command, _ []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), profile.GetCurrentName())
	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	if err := profile.Delete(args[0]); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}
