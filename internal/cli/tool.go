package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
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

	disable := &cobra.Command{
		Use:   "disable",
		Short: "Remove a tool harness's MoAI wiring from the current project",
		Args:  cobra.NoArgs,
	}
	disable.AddCommand(newToolDisableCodexCmd())
	cmd.AddCommand(disable)
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

// newToolDisableCodexCmd creates the Codex unwire command (REQ-DHR-005).
func newToolDisableCodexCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "codex",
		Short: "Remove the Codex wiring MoAI created, keeping everything else",
		Long: `Remove the Codex wiring MoAI created in this project.

Only parts MoAI recorded as its own and that are still exactly what it wrote
are removed: its hook handlers and hooks.json description, its
[mcp_servers.moai] and [tui] tables or the status_line line it inserted into
your [tui], and a wiring file MoAI created whole and nobody changed since.
Everything else stays byte-for-byte, and every part left in place is listed
with the reason. A part MoAI cannot prove it wrote is never removed; the
report tells you to remove it by hand if it is not yours.

"moai update" does not wire the project again afterwards; run
"moai tool enable codex" to wire it again.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRoot(cmd)
			if err != nil {
				return fmt.Errorf("tool disable codex: %w", err)
			}
			if !isMoAIProject(root) {
				return fmt.Errorf("tool disable codex: not a MoAI project (missing .moai/config/sections/system.yaml)")
			}
			return runToolDisableCodexAt(root, cmd.OutOrStdout(), cmd.ErrOrStderr(), dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show the Codex unwire plan without modifying the filesystem")
	cmd.Flags().String("project-root", "", "Project root path (default: current directory)")
	return cmd
}

func runToolDisableCodexAt(projectRoot string, out, errOut io.Writer, dryRun bool) error {
	if dryRun {
		emitCodexUnwireDryRunPreview(out)
		return nil
	}
	res, err := codexwiring.Unwire(projectRoot, errOut)
	reportCodexUnwire(out, res)
	if err != nil {
		if errors.Is(err, codexwiring.ErrWiringLockHeld) {
			return fmt.Errorf("tool disable codex: %w (%s); nothing was changed, retry once it is released", err, codexwiring.WiringLockRelPath)
		}
		return fmt.Errorf("tool disable codex: %w", err)
	}
	return nil
}

// reportCodexUnwire prints what the unwire removed and what it kept, one line
// per file or part, in the order the unwire met them.
func reportCodexUnwire(out io.Writer, res codexwiring.UnwireResult) {
	if n := len(res.Recovered); n > 0 {
		_, _ = fmt.Fprintf(out, "Recovered %d interrupted Codex wiring change(s) first.\n", n)
	}
	if len(res.Removed) == 0 {
		_, _ = fmt.Fprintln(out, "Codex wiring: nothing MoAI created was found to remove.")
	} else {
		_, _ = fmt.Fprintln(out, "Codex wiring disabled:")
	}
	for _, r := range res.Removed {
		_, _ = fmt.Fprintf(out, "  removed %s\n", codexPartLabel(r.Path, r.Part))
	}
	for _, k := range res.Kept {
		line := fmt.Sprintf("  kept %s (%s): %s", codexPartLabel(k.Path, k.Part), k.Reason, k.Detail)
		if k.Reason == codexwiring.ReasonUserOwned || k.Reason == codexwiring.ReasonNoProvenance {
			line += " — not MoAI's, left as is"
		}
		_, _ = fmt.Fprintln(out, line)
	}
	_, _ = fmt.Fprintf(out, "Run `%s` to wire Codex again.\n", codexwiring.RecoverCommand)
}

// codexPartLabel renders "<file>" or "<file> [<part>]".
func codexPartLabel(path, part string) string {
	if part == "" {
		return path
	}
	return path + " [" + part + "]"
}

// emitCodexUnwireDryRunPreview prints the unwire actions. It writes nothing.
func emitCodexUnwireDryRunPreview(out io.Writer) {
	_, _ = fmt.Fprintf(out, "Dry-run %s unwire plan (nothing written):\n", codexwiring.DisableCommand)
	_, _ = fmt.Fprintf(out, "  - remove the MoAI-created parts of %s (its hook handlers and description), or the whole file if MoAI created it and it is unchanged\n", codexwiring.HooksRelPath)
	_, _ = fmt.Fprintf(out, "  - cut the MoAI-created parts out of %s ([mcp_servers.moai], [tui] or its status_line line), keeping every other byte\n", codexwiring.ConfigRelPath)
	_, _ = fmt.Fprintln(out, "  - leave user-owned, unknown-origin, modified, and symlinked parts in place and list each with the reason")
	_, _ = fmt.Fprintf(out, "  - record the disable in %s so `moai update` does not wire the project again\n", codexwiring.SidecarPath)
	_, _ = fmt.Fprintln(out, "  - run without --dry-run to apply")
}
