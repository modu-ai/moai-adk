package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ccCmd = &cobra.Command{
	Use:   "cc",
	Short: "Switch to Claude backend",
	Long:  "Switch the active LLM backend to Claude Code. This is a top-level command (not under 'switch').",
	RunE:  runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude.
func runCC(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if deps == nil || deps.Config == nil {
		fmt.Fprintln(out, "Switched to Claude backend (config not loaded).")
		return nil
	}

	cfg := deps.Config.Get()
	if cfg != nil {
		cfg.LLM.DefaultModel = "claude"
		if err := deps.Config.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}

	fmt.Fprintln(out, "Switched to Claude backend.")
	return nil
}
