package cli

import (
	"os"

	"github.com/anthropics/moai-adk-go/internal/hooks"
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
			// Run the hook dispatcher
			if err := hooks.Run(); err != nil {
				os.Exit(1)
			}
			return nil
		},
	}

	return cmd
}
