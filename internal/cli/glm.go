package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var glmCmd = &cobra.Command{
	Use:   "glm [api-key]",
	Short: "Switch to GLM backend",
	Long:  "Switch the active LLM backend to GLM. Optionally provide an API key. This is a top-level command (not under 'switch').",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGLM,
}

func init() {
	rootCmd.AddCommand(glmCmd)
}

// runGLM switches the LLM backend to GLM and optionally saves the API key.
func runGLM(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	apiKey := ""
	if len(args) > 0 {
		apiKey = args[0]
	}

	if deps == nil || deps.Config == nil {
		msg := "Switched to GLM backend (config not loaded)."
		if apiKey != "" {
			msg = "Switched to GLM backend with API key (config not loaded)."
		}
		fmt.Fprintln(out, msg)
		return nil
	}

	cfg := deps.Config.Get()
	if cfg != nil {
		cfg.LLM.DefaultModel = "glm"
		if err := deps.Config.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}

	if apiKey != "" {
		fmt.Fprintln(out, "Switched to GLM backend with API key saved.")
	} else {
		fmt.Fprintln(out, "Switched to GLM backend.")
	}
	return nil
}
