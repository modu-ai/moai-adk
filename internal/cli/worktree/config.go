package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config [key]",
		Short: "Show worktree configuration",
		Long: `Show worktree configuration settings.

Available keys:
  root      - Repository root directory
  all       - Show all configuration (default)

Examples:
  moai worktree config        # Show all config
  moai worktree config root   # Show root directory only`,
		Args: cobra.MaximumNArgs(1),
		RunE: runConfig,
	}
}

func runConfig(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	key := "all"
	if len(args) > 0 {
		key = args[0]
	}

	root := WorktreeProvider.Root()

	switch key {
	case "root":
		_, _ = fmt.Fprintf(out, "Worktree root: %s\n", root)
	case "all":
		_, _ = fmt.Fprintln(out, "Worktree Configuration:")
		_, _ = fmt.Fprintf(out, "  root: %s\n", root)
	default:
		return fmt.Errorf("unknown config key: %q (available: root, all)", key)
	}

	return nil
}
