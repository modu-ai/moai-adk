package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/llmswitch"
	"github.com/spf13/cobra"
)

// NewClaudeCommand creates the claude command for switching to Claude backend
func NewClaudeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claude",
		Aliases: []string{"cc"},
		Short:   "Switch to Claude backend",
		Long:    `Switch the LLM backend to Claude (Anthropic). Removes GLM environment variables from .claude/settings.local.json.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			return llmswitch.SwitchToClaude(cwd)
		},
	}
	return cmd
}

// NewGLMCommand creates the glm command for switching to GLM backend or updating API key
func NewGLMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "glm [api-key]",
		Short: "Switch to GLM backend or update API key",
		Long: `Switch the LLM backend to GLM. If an API key is provided, it updates the key without switching.
Use 'moai-go glm <key>' to set your API key first, then 'moai-go glm' to switch.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return llmswitch.UpdateGLMKey(args[0])
			}
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			return llmswitch.SwitchToGLM(cwd)
		},
	}
	return cmd
}
