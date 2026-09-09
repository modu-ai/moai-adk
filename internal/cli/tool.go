package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// newToolCmd creates the project harness management namespace.
func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tool",
		Short:   "Manage project tool harnesses",
		GroupID: "tools",
		Args:    cobra.NoArgs,
	}

	enable := &cobra.Command{
		Use:   "enable",
		Short: "Enable a tool harness in the current project",
		Args:  cobra.NoArgs,
	}
	enable.AddCommand(newToolEnableCodexCmd())
	cmd.AddCommand(enable)
	return cmd
}

// newToolEnableCodexCmd creates the additive Codex harness command.
func newToolEnableCodexCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "codex",
		Short: "Add or refresh the Codex harness without reinitializing the project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRoot(cmd)
			if err != nil {
				return fmt.Errorf("tool enable codex: %w", err)
			}
			if !isMoAIProject(root) {
				return fmt.Errorf("tool enable codex: not a MoAI project (missing .moai/config/sections/system.yaml); run `moai init` first")
			}
			return runToolEnableCodexAt(root, cmd.OutOrStdout(), cmd.ErrOrStderr(), dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show the Codex wiring plan without modifying the filesystem")
	cmd.Flags().String("project-root", "", "Project root path (default: current directory)")
	return cmd
}

func runToolEnableCodexAt(projectRoot string, out, errOut io.Writer, dryRun bool) error {
	if dryRun {
		emitCodexWiringDryRunPreview(out, "moai tool enable codex")
		return nil
	}
	return addCodexWiringAt(projectRoot, out, errOut)
}
