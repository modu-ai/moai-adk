package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/internal/cli/worktree"
	"github.com/modu-ai/moai-adk-go/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "moai",
	Short: "MoAI-ADK: Agentic Development Kit for Claude Code",
	Long: `MoAI-ADK (Go Edition) is a high-performance development toolkit
that serves as the runtime backbone for the MoAI framework within Claude Code.

It provides CLI tooling, configuration management, LSP integration,
Git operations, quality gates, and autonomous development loop capabilities.`,
	Version: version.GetVersion(),
}

// Execute initializes dependencies and runs the root command.
func Execute() error {
	InitDependencies()
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("moai-adk %s\n", version.GetVersion()))

	// Register worktree subcommand tree
	rootCmd.AddCommand(worktree.WorktreeCmd)
}
