package cli

// @MX:NOTE: [AUTO] Profile management for Claude configuration directories
// @MX:NOTE: [AUTO] Each profile is an isolated CLAUDE_CONFIG_DIR in ~/.moai/claude-profiles/
// @MX:NOTE: [AUTO] Used with -p/--profile flag on cc, cg, glm commands

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:     "profile",
	Short:   "Manage Claude configuration profiles",
	GroupID: "tools",
	Long: `Manage Claude configuration profiles stored in ~/.moai/claude-profiles/.

Each profile is an isolated Claude configuration directory (CLAUDE_CONFIG_DIR).
Use -p/--profile with cc, cg, or glm to switch between profiles.

Run 'moai profile setup [name]' to configure per-profile preferences
(identity, languages, model settings, display).`,
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
	profileCmd.Flags().BoolP("setup", "s", false, "Run interactive setup wizard for preferences")
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

// emitProfileScopeNotice emits the REQ-SW-017 honest scope notice for `moai
// profile` auto-entry. It names BOTH the project-scoped worktree path AND the
// global profile dir (profile.GetBaseDir()), and states EXPLICITLY that the
// profile dir is NOT isolated by this entry — so the user is not misled into
// thinking the per-profile launch-ledger race
// (~/.moai/claude-profiles/<name>/launch.yaml) is solved. The "NOT isolated"
// clause is load-bearing (AC-SW-017(c)); see
// TestEmitProfileScopeNotice_FalsificationRoundTrip.
func emitProfileScopeNotice(out io.Writer, wtPath string) {
	_, _ = fmt.Fprintf(out,
		"moai: profile auto-entry scoped to the PROJECT worktree %s; "+
			"the global profile dir %s is NOT isolated by this entry "+
			"(launch.yaml ledger race remains — out of scope, see spec.md §3)\n",
		wtPath, profile.GetBaseDir())
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
	name := profile.GetCurrentName()
	if root, err := findProjectRootFn(); err == nil {
		name = profile.GetCurrentNameForProject(root)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), name)
	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	if err := profile.Delete(args[0]); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}
