package worktree

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// WorktreeCreator is wired by the parent cli package to the current
// materializeSessionWorktree implementation. Keeping the command behind this
// adapter avoids a second git worktree creation path in this subpackage.
var WorktreeCreator func(name string, out io.Writer) (string, error)

func newNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name>",
		Short: "Create a worktree through the shared MoAI materializer",
		Long: `Create one harness-neutral L1 worktree at .claude/worktrees/<name>.

The command delegates to MoAI's existing session-worktree materializer. It
does not enter the new tree and does not revive the retired base, path, tmux,
team, or BODP behavior. Enter with a launcher after creation.`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected argument <name>, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := validateNewWorktreeName(name); err != nil {
				return err
			}
			if WorktreeCreator == nil {
				return fmt.Errorf("worktree creator is not initialized")
			}
			path, err := WorktreeCreator(name, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("worktree %q could not be created: %w", name, err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created worktree %s at %s\n", name, path)
			return nil
		},
	}
}

func validateNewWorktreeName(name string) error {
	if name == "" || strings.TrimSpace(name) != name {
		return fmt.Errorf("worktree name must be a non-empty L1 leaf name")
	}
	if name == "." || filepath.IsAbs(name) || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid worktree name %q: expected one L1 leaf without traversal or path separators", name)
	}
	return nil
}
