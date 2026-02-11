package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config [key] [value]",
		Short: "Show or set worktree configuration",
		Long: `Show or set worktree configuration settings.

Get mode (0-1 args):
  moai worktree config          # Show all config
  moai worktree config root     # Show root directory only

Available keys:
  root        - Repository root directory
  all         - Show all configuration (default)

Examples:
  moai worktree config          # Show all config
  moai worktree config root     # Show root directory only`,
		Args: cobra.MaximumNArgs(2),
		RunE: runConfig,
	}
}

func runConfig(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	root := WorktreeProvider.Root()

	switch len(args) {
	case 0:
		// Show all config.
		_, _ = fmt.Fprintln(out, "Worktree Configuration:")
		_, _ = fmt.Fprintf(out, "  root: %s\n", root)
	case 1:
		// Show specific key.
		key := args[0]
		switch key {
		case "root":
			_, _ = fmt.Fprintf(out, "Worktree root: %s\n", root)
		case "all":
			_, _ = fmt.Fprintln(out, "Worktree Configuration:")
			_, _ = fmt.Fprintf(out, "  root: %s\n", root)
		default:
			return fmt.Errorf("unknown config key: %q (available: root, all)", key)
		}
	default:
		return fmt.Errorf("config set is not yet supported; configuration is derived from git repository")
	}

	return nil
}
