package cli

import (
	"github.com/spf13/cobra"
)

// NewHookCommand creates the hook command
func NewHookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook [event]",
		Short: "Execute Claude Code hooks",
		Long: `Execute Claude Code hooks for specific events (session-start,
session-end, pre-tool-use, post-tool-use, etc.). This command is called
by Claude Code's hook system.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			event := args[0]
			// TODO: Implement hook command
			cmd.Printf("hook command for event '%s': not yet implemented\n", event)
			return nil
		},
	}

	return cmd
}
